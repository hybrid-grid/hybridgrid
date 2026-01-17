package metrics

import (
	"sync"
)

const (
	// DefaultAlpha is the default EWMA smoothing factor.
	DefaultAlpha = 0.5
	// DefaultLatencyMs is the default latency for unknown workers.
	DefaultLatencyMs = 100.0
)

// LatencyTracker tracks per-worker latency using EWMA.
type LatencyTracker struct {
	mu       sync.RWMutex
	workers  map[string]*EWMA
	alpha    float64
	defValue float64
}

// NewLatencyTracker creates a new latency tracker.
func NewLatencyTracker() *LatencyTracker {
	return &LatencyTracker{
		workers:  make(map[string]*EWMA),
		alpha:    DefaultAlpha,
		defValue: DefaultLatencyMs,
	}
}

// NewLatencyTrackerWithConfig creates a tracker with custom configuration.
func NewLatencyTrackerWithConfig(alpha, defaultLatency float64) *LatencyTracker {
	return &LatencyTracker{
		workers:  make(map[string]*EWMA),
		alpha:    alpha,
		defValue: defaultLatency,
	}
}

// Record records a latency measurement for a worker.
func (t *LatencyTracker) Record(workerID string, latencyMs float64) {
	t.mu.Lock()
	defer t.mu.Unlock()

	ewma, ok := t.workers[workerID]
	if !ok {
		ewma = NewEWMA(t.alpha)
		t.workers[workerID] = ewma
	}
	ewma.Update(latencyMs)
}

// Get returns the EWMA latency for a worker.
// Returns the default value if worker hasn't been recorded yet.
func (t *LatencyTracker) Get(workerID string) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if ewma, ok := t.workers[workerID]; ok && ewma.IsInitialized() {
		return ewma.Value()
	}
	return t.defValue
}

// Remove removes a worker from tracking.
func (t *LatencyTracker) Remove(workerID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.workers, workerID)
}

// Reset clears all tracked latencies.
func (t *LatencyTracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.workers = make(map[string]*EWMA)
}

// All returns latencies for all tracked workers.
func (t *LatencyTracker) All() map[string]float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()

	result := make(map[string]float64, len(t.workers))
	for id, ewma := range t.workers {
		result[id] = ewma.Value()
	}
	return result
}
