package resilience

import (
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sony/gobreaker"
)

// CircuitState represents the circuit breaker state.
type CircuitState string

const (
	CircuitClosed   CircuitState = "CLOSED"
	CircuitHalfOpen CircuitState = "HALF_OPEN"
	CircuitOpen     CircuitState = "OPEN"
)

// CircuitConfig holds circuit breaker configuration.
type CircuitConfig struct {
	// MaxRequests is the number of requests allowed in half-open state.
	MaxRequests uint32
	// Interval is the cyclic period of the closed state.
	Interval time.Duration
	// Timeout is the period of the open state.
	Timeout time.Duration
	// FailureRatio is the failure rate threshold (0.0 to 1.0).
	FailureRatio float64
	// MinRequests is the minimum requests before failure ratio is checked.
	MinRequests uint32
}

// DefaultCircuitConfig returns sensible defaults per DOCUMENTATION.md.
func DefaultCircuitConfig() CircuitConfig {
	return CircuitConfig{
		MaxRequests:  3,                // Allow 3 requests in half-open
		Interval:     10 * time.Second, // 10s sliding window
		Timeout:      60 * time.Second, // 60s open duration
		FailureRatio: 0.6,              // 60% failure rate triggers open
		MinRequests:  3,                // Need at least 3 requests
	}
}

// CircuitManager manages per-worker circuit breakers.
type CircuitManager struct {
	mu       sync.RWMutex
	breakers map[string]*gobreaker.CircuitBreaker
	config   CircuitConfig
	onChange func(workerID string, from, to CircuitState)
}

// NewCircuitManager creates a new circuit manager.
func NewCircuitManager(cfg CircuitConfig) *CircuitManager {
	return &CircuitManager{
		breakers: make(map[string]*gobreaker.CircuitBreaker),
		config:   cfg,
	}
}

// OnStateChange sets a callback for circuit state changes.
func (m *CircuitManager) OnStateChange(fn func(workerID string, from, to CircuitState)) {
	m.onChange = fn
}

// getOrCreate gets or creates a circuit breaker for a worker.
func (m *CircuitManager) getOrCreate(workerID string) *gobreaker.CircuitBreaker {
	m.mu.RLock()
	cb, exists := m.breakers[workerID]
	m.mu.RUnlock()

	if exists {
		return cb
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = m.breakers[workerID]; exists {
		return cb
	}

	settings := gobreaker.Settings{
		Name:        workerID,
		MaxRequests: m.config.MaxRequests,
		Interval:    m.config.Interval,
		Timeout:     m.config.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < m.config.MinRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= m.config.FailureRatio
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			fromState := gobreakerStateToCircuitState(from)
			toState := gobreakerStateToCircuitState(to)

			log.Info().
				Str("worker_id", name).
				Str("from", string(fromState)).
				Str("to", string(toState)).
				Msg("Circuit breaker state change")

			if m.onChange != nil {
				m.onChange(name, fromState, toState)
			}
		},
	}

	cb = gobreaker.NewCircuitBreaker(settings)
	m.breakers[workerID] = cb

	return cb
}

// Execute wraps a function call with circuit breaker protection.
func (m *CircuitManager) Execute(workerID string, fn func() (interface{}, error)) (interface{}, error) {
	cb := m.getOrCreate(workerID)
	return cb.Execute(fn)
}

// IsOpen returns true if the circuit breaker is open for a worker.
func (m *CircuitManager) IsOpen(workerID string) bool {
	m.mu.RLock()
	cb, exists := m.breakers[workerID]
	m.mu.RUnlock()

	if !exists {
		return false
	}

	return cb.State() == gobreaker.StateOpen
}

// GetState returns the current state of a worker's circuit breaker.
func (m *CircuitManager) GetState(workerID string) CircuitState {
	m.mu.RLock()
	cb, exists := m.breakers[workerID]
	m.mu.RUnlock()

	if !exists {
		return CircuitClosed
	}

	return gobreakerStateToCircuitState(cb.State())
}

// GetAllStates returns the states of all tracked circuit breakers.
func (m *CircuitManager) GetAllStates() map[string]CircuitState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]CircuitState, len(m.breakers))
	for id, cb := range m.breakers {
		result[id] = gobreakerStateToCircuitState(cb.State())
	}
	return result
}

// Remove removes a circuit breaker for a worker.
func (m *CircuitManager) Remove(workerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.breakers, workerID)
}

// gobreakerStateToCircuitState converts gobreaker state to our CircuitState.
func gobreakerStateToCircuitState(state gobreaker.State) CircuitState {
	switch state {
	case gobreaker.StateClosed:
		return CircuitClosed
	case gobreaker.StateHalfOpen:
		return CircuitHalfOpen
	case gobreaker.StateOpen:
		return CircuitOpen
	default:
		return CircuitClosed
	}
}
