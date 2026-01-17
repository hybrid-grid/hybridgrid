package metrics

import (
	"sync"
)

// EWMA calculates Exponentially Weighted Moving Average.
type EWMA struct {
	mu    sync.RWMutex
	alpha float64
	value float64
	init  bool
}

// NewEWMA creates a new EWMA with the given alpha (smoothing factor).
// Alpha should be between 0 and 1. Higher alpha gives more weight to recent values.
// Common values: 0.5 for balanced, 0.9 for recent bias, 0.1 for historical bias.
func NewEWMA(alpha float64) *EWMA {
	if alpha <= 0 || alpha > 1 {
		alpha = 0.5
	}
	return &EWMA{
		alpha: alpha,
	}
}

// Update adds a new value to the EWMA.
func (e *EWMA) Update(value float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.init {
		e.value = value
		e.init = true
		return
	}

	// EWMA formula: new_avg = alpha * new_value + (1-alpha) * old_avg
	e.value = e.alpha*value + (1-e.alpha)*e.value
}

// Value returns the current EWMA value.
func (e *EWMA) Value() float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.value
}

// Reset resets the EWMA to uninitialized state.
func (e *EWMA) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.value = 0
	e.init = false
}

// IsInitialized returns true if at least one value has been added.
func (e *EWMA) IsInitialized() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.init
}
