package scheduler

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// iceccReward inverts the shaped reward the coordinator feeds back, so tests
// can express observations in the milliseconds a worker would report.
func iceccReward(compileMs float64) float64 { return -math.Log(1 + compileMs) }

func iceccRegistry(t *testing.T, ids ...string) registry.Registry {
	t.Helper()
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(reg.Stop)
	for _, id := range ids {
		require.NoError(t, reg.Add(&registry.WorkerInfo{
			ID: id,
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

func iceccSelect(t *testing.T, s *IceccFastestScheduler) (*registry.WorkerInfo, DispatchInfo) {
	t.Helper()
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	return w, info
}

// Reward inversion is the load-bearing assumption of the port: the interface
// hands us -log(1+t) but icecream's estimator needs t itself.
func TestIcecc_RecoversCompileTimeFromShapedReward(t *testing.T) {
	reg := iceccRegistry(t, "a")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	s.RecordOutcome("a", iceccReward(250), true, TaskContext{SourceSizeBytes: 1000})

	st := s.stats["a"]
	require.NotNil(t, st)
	assert.InDelta(t, 250.0, st.cumTime, 1e-6, "compile time must round-trip through the reward transform")
}

// pick_server_new: an unmeasured idle node is dispatched to before any
// measured node, regardless of how good the measured node looks.
func TestIcecc_DispatchesToUnmeasuredNodeFirst(t *testing.T) {
	reg := iceccRegistry(t, "measured", "fresh")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	s.RecordOutcome("measured", iceccReward(10), true, TaskContext{SourceSizeBytes: 1_000_000})

	w, info := iceccSelect(t, s)
	assert.Equal(t, "fresh", w.ID, "zero-stat idle node must win before argmax runs")
	assert.True(t, info.WasExploration, "cold-arm dispatch is an exploration step")
}

// Once every node is measured, selection is argmax of work-per-millisecond.
func TestIcecc_PicksHigherThroughputNode(t *testing.T) {
	reg := iceccRegistry(t, "fast", "slow")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	// Equal work, but "fast" needs a quarter of the time. Push both past
	// the under-sampling window so the optimism factor cancels out.
	for i := 0; i < iceccPessimismSamples; i++ {
		s.RecordOutcome("fast", iceccReward(100), true, TaskContext{SourceSizeBytes: 100_000})
		s.RecordOutcome("slow", iceccReward(400), true, TaskContext{SourceSizeBytes: 100_000})
	}

	w, info := iceccSelect(t, s)
	assert.Equal(t, "fast", w.ID)
	assert.False(t, info.WasExploration)
	assert.InDelta(t, 1000.0, info.QValueAtDispatch, 1e-6, "100000 bytes / 100 ms = 1000 bytes per ms")
}

// The under-sampling factor is optimism, not pessimism: a node with a single
// observation is boosted 4x and beats an equal-throughput, well-sampled node.
func TestIcecc_UnderSampledNodeIsBoosted(t *testing.T) {
	reg := iceccRegistry(t, "seasoned", "novice")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	for i := 0; i < iceccPessimismSamples+3; i++ {
		s.RecordOutcome("seasoned", iceccReward(100), true, TaskContext{SourceSizeBytes: 100_000})
	}
	// Identical throughput, one sample. Busy so pick_server_new skips it.
	s.RecordOutcome("novice", iceccReward(100), true, TaskContext{SourceSizeBytes: 100_000})
	require.NoError(t, reg.IncrementTasks("novice"))

	w, _ := iceccSelect(t, s)
	assert.Equal(t, "novice", w.ID, "one-sample node gets 4x, minus 12.5% saturation, and still wins")
}

// The queue-saturation factor must throttle a node as work piles onto it,
// which is the term structurally equivalent to Hybrid-LinUCB's load penalty.
func TestIcecc_SaturationThrottlesBusyNode(t *testing.T) {
	reg := iceccRegistry(t, "busy", "idle")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	// "busy" is genuinely twice as fast, but is running 3 of its 4 slots:
	// 2000 * (1 - 0.5*3/4) = 1250 versus idle's 1000.
	for i := 0; i < iceccPessimismSamples; i++ {
		s.RecordOutcome("busy", iceccReward(50), true, TaskContext{SourceSizeBytes: 100_000})
		s.RecordOutcome("idle", iceccReward(100), true, TaskContext{SourceSizeBytes: 100_000})
	}
	for i := 0; i < 3; i++ {
		require.NoError(t, reg.IncrementTasks("busy"))
	}

	w, info := iceccSelect(t, s)
	assert.Equal(t, "busy", w.ID)
	assert.InDelta(t, 1250.0, info.QValueAtDispatch, 1e-6)

	// One more concurrent job drops it to 1000, tying idle; the fourth
	// exhausts MaxParallel and removes it from the candidate set entirely.
	require.NoError(t, reg.IncrementTasks("busy"))
	w, _ = iceccSelect(t, s)
	assert.Equal(t, "idle", w.ID, "saturated node must yield once throttled below its rival")
}

// Failed tasks are dropped rather than fed back, matching add_job_stats.
// This is a deliberate divergence from our own learners.
func TestIcecc_IgnoresFailedTasks(t *testing.T) {
	reg := iceccRegistry(t, "a")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	s.RecordOutcome("a", iceccReward(9999), false, TaskContext{SourceSizeBytes: 1000})

	assert.Nil(t, s.stats["a"], "a failed job must not enter the throughput window")
	assert.Zero(t, s.recorded)
}

// The rolling window is capped, and the cumulative sums must describe exactly
// the retained samples or the estimator drifts permanently.
func TestIcecc_WindowEvictsOldestAndKeepsSumsConsistent(t *testing.T) {
	reg := iceccRegistry(t, "a")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	// Fill the window with slow jobs, then overwrite it entirely with fast
	// ones so no trace of the slow era may survive.
	for i := 0; i < iceccWindowSize; i++ {
		s.RecordOutcome("a", iceccReward(500), true, TaskContext{SourceSizeBytes: 1000})
	}
	for i := 0; i < iceccWindowSize; i++ {
		s.RecordOutcome("a", iceccReward(100), true, TaskContext{SourceSizeBytes: 1000})
	}

	st := s.stats["a"]
	require.Len(t, st.window, iceccWindowSize, "window must stay capped")

	var wantWork, wantTime float64
	for _, j := range st.window {
		wantWork += j.work
		wantTime += j.timeMs
	}
	assert.InDelta(t, wantWork, st.cumWork, 1e-6)
	assert.InDelta(t, wantTime, st.cumTime, 1e-6)
	assert.InDelta(t, 100.0*float64(iceccWindowSize), st.cumTime, 1e-3,
		"only the fast era should remain after a full turnover")
}

// While the cluster has recorded nothing at all, selection is uniform-random
// but must still respect admission rules and report itself as exploration.
func TestIcecc_RandomWhileClusterUnmeasured(t *testing.T) {
	reg := iceccRegistry(t, "a", "b")
	s := NewIceccFastestScheduler(IceccConfig{Registry: reg})

	w, info := iceccSelect(t, s)
	assert.Contains(t, []string{"a", "b"}, w.ID)
	assert.True(t, info.WasExploration)
}

// The scheduler must satisfy the learning interface so Compile() feeds it.
func TestIcecc_ImplementsLearningScheduler(t *testing.T) {
	var _ LearningScheduler = NewIceccFastestScheduler(IceccConfig{
		Registry: iceccRegistry(t, "a"),
	})
}
