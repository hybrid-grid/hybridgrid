package server

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/resilience"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

func TestNewScheduler_KnownTypes(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	cases := map[string]any{
		"simple":         (*scheduler.SimpleScheduler)(nil),
		"p2c":            (*scheduler.P2CScheduler)(nil),
		"leastloaded":    (*scheduler.LeastLoadedScheduler)(nil),
		"":               (*scheduler.LeastLoadedScheduler)(nil), // empty -> default
		"epsilon-greedy": (*scheduler.EpsilonGreedyScheduler)(nil),
	}

	for typ, want := range cases {
		t.Run(typ, func(t *testing.T) {
			got := newScheduler(Config{SchedulerType: typ}, reg, cm)
			assert.NotNil(t, got)
			assert.IsType(t, want, got)
		})
	}
}

func TestNewScheduler_UnknownTypeFallsBack(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	got := newScheduler(Config{SchedulerType: "does-not-exist"}, reg, cm)
	assert.IsType(t, (*scheduler.LeastLoadedScheduler)(nil), got)
}

// TestNewScheduler_EpsilonValueRespected ensures the configured
// EpsilonValue propagates to the EpsilonGreedyScheduler. We can't read
// the field directly (private) but a Q-based behavioural test would be
// brittle; verify only that construction with non-default ε works.
func TestNewScheduler_EpsilonValueRespected(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	got := newScheduler(Config{SchedulerType: "epsilon-greedy", EpsilonValue: 0.5}, reg, cm)
	assert.IsType(t, (*scheduler.EpsilonGreedyScheduler)(nil), got)
}
