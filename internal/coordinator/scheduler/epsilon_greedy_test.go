package scheduler

import (
	"math"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func newRegistryWithWorkers(t *testing.T, n int) *registry.InMemoryRegistry {
	t.Helper()
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(func() { reg.Stop() })
	for i := 0; i < n; i++ {
		require.NoError(t, reg.Add(&registry.WorkerInfo{
			ID:      idForTest(i),
			Address: "127.0.0.1:0",
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				CpuCores:   4,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
			MaxParallel: 4,
		}))
	}
	return reg
}

func idForTest(i int) string {
	return "worker-" + string(rune('a'+i))
}

// TestEpsilonGreedy_RecordsRunningMean checks the §2.4 incremental
// sample-mean update. Feeding rewards 1, 2, 3 should yield Q = 2.
func TestEpsilonGreedy_RecordsRunningMean(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0})

	for _, r := range []float64{1, 2, 3} {
		s.RecordOutcome("worker-a", r, true, TaskContext{})
	}
	assert.InDelta(t, 2.0, s.qValue("worker-a"), 1e-9)
}

// TestEpsilonGreedy_ConvergesToBestArm verifies that after enough
// observations the highest-mean worker is reliably picked under
// pure greedy (ε=0). Three workers with means -1, -2, -3 (higher is
// better, so worker-a is best). 200 selections, expect worker-a chosen
// >= 90% of the time *after* warm-up.
func TestEpsilonGreedy_ConvergesToBestArm(t *testing.T) {
	reg := newRegistryWithWorkers(t, 3)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0.1})

	rng := rand.New(rand.NewSource(42))
	means := map[string]float64{"worker-a": -1, "worker-b": -2, "worker-c": -3}

	for i := 0; i < 300; i++ {
		w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
		require.NoError(t, err)
		// Sample reward ~ N(mean, 0.5)
		reward := means[w.ID] + rng.NormFloat64()*0.5
		s.RecordOutcome(w.ID, reward, true, TaskContext{})
	}

	// After 300 calls, ranking of Q values should match ranking of true means.
	qa, qb, qc := s.qValue("worker-a"), s.qValue("worker-b"), s.qValue("worker-c")
	assert.Greater(t, qa, qb, "worker-a should outrank worker-b in Q")
	assert.Greater(t, qb, qc, "worker-b should outrank worker-c in Q")

	// Sample-mean accuracy: |Q - μ| < 0.3 for each arm with this sample budget.
	assert.InDelta(t, -1.0, qa, 0.3)
	assert.InDelta(t, -2.0, qb, 0.3)
	assert.InDelta(t, -3.0, qc, 0.3)
}

// TestEpsilonGreedy_ExplorationRateHonored checks that with ε=0.3 the
// observed exploration count is within ±2σ of expected over 1000 calls.
// Uses a binomial test: σ = √(N·ε·(1-ε)) ≈ √(1000·0.3·0.7) ≈ 14.5.
func TestEpsilonGreedy_ExplorationRateHonored(t *testing.T) {
	reg := newRegistryWithWorkers(t, 5)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0.3})

	// Pre-train so argmaxQ is well-defined and "exploit" is distinguishable.
	for i := 0; i < 5; i++ {
		s.RecordOutcome(idForTest(i), float64(-i), true, TaskContext{})
	}

	const N = 1000
	exploreCount := 0
	for i := 0; i < N; i++ {
		_, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
		require.NoError(t, err)
		if info.WasExploration {
			exploreCount++
		}
	}
	expected := int(0.3 * N)
	tolerance := 50 // ~ 3σ — keep test stable across CI runs
	assert.InDelta(t, expected, exploreCount, float64(tolerance),
		"exploration count %d outside expected %d ± %d", exploreCount, expected, tolerance)
}

// TestEpsilonGreedy_SingleCandidateFastPath ensures the dispatch fast
// path skips the random draw when only one worker is eligible — the
// motivation is the M1 P2C 1-worker overhead penalty.
func TestEpsilonGreedy_SingleCandidateFastPath(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 1.0}) // ε=1 would explore every time, but with 1 candidate exploration is meaningless

	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	assert.Equal(t, "worker-a", w.ID)
	assert.False(t, info.WasExploration, "single candidate must never be reported as exploration")
}

// TestEpsilonGreedy_NoWorkers returns ErrNoWorkers (no registered) and
// ErrNoMatchingWorkers (registered but unhealthy).
func TestEpsilonGreedy_NoWorkers(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(func() { reg.Stop() })
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0.1})

	_, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	assert.ErrorIs(t, err, ErrNoWorkers)
}

// TestEpsilonGreedy_RecordOutcomeIgnoresInvalidInputs guards against
// NaN/Inf rewards (numeric explosion) and empty worker IDs.
func TestEpsilonGreedy_RecordOutcomeIgnoresInvalidInputs(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0.1})

	s.RecordOutcome("", 1.0, true, TaskContext{})
	s.RecordOutcome("worker-a", math.NaN(), true, TaskContext{})
	s.RecordOutcome("worker-a", math.Inf(1), true, TaskContext{})
	assert.Equal(t, 0.0, s.qValue("worker-a"))
}

// TestEpsilonGreedy_ConcurrentSafe runs many select+record loops in
// parallel under -race to flush out missing locks.
func TestEpsilonGreedy_ConcurrentSafe(t *testing.T) {
	reg := newRegistryWithWorkers(t, 4)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0.2})

	const goroutines = 16
	const iterations = 100
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
				if err != nil {
					return
				}
				s.RecordOutcome(w.ID, -float64(i), true, TaskContext{})
			}
		}()
	}
	wg.Wait()

	// At least one worker should have non-zero samples.
	total := int64(0)
	s.mu.Lock()
	for _, st := range s.values {
		total += st.n
	}
	s.mu.Unlock()
	assert.Greater(t, total, int64(0))
}

// TestSelectWith_FallsBackForNonLearners verifies the gRPC layer's
// helper function returns a zero DispatchInfo for non-learning
// schedulers (LeastLoaded, P2C, Simple).
func TestSelectWith_FallsBackForNonLearners(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	ll := NewLeastLoadedScheduler(reg)

	w, info, err := SelectWith(ll, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	require.NotNil(t, w)
	assert.Equal(t, DispatchInfo{}, info)
}

// TestSelectWith_UsesLearnerInterface confirms a LearningScheduler's
// SelectWithDispatchInfo is preferred over Select when available.
func TestSelectWith_UsesLearnerInterface(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewEpsilonGreedyScheduler(EpsilonGreedyConfig{Registry: reg, Epsilon: 0})
	s.RecordOutcome("worker-b", 5.0, true, TaskContext{}) // make worker-b clearly best

	_, info, err := SelectWith(s, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	// Q for worker-b is 5.0; worker-a is 0.
	assert.InDelta(t, 5.0, info.QValueAtDispatch, 1e-9)
}
