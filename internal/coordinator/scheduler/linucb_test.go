package scheduler

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// TestLinUCB_ArmInitialization checks Algorithm 1 lines 5–6: every new
// arm must start with A_a = I_d and b_a = 0. Uses 2 workers so the
// scoring path runs (single-worker case takes the fast path which
// intentionally avoids creating arm state).
func TestLinUCB_ArmInitialization(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})

	_, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
	require.NoError(t, err)

	s.mu.Lock()
	arm := s.arms["worker-a"]
	s.mu.Unlock()
	require.NotNil(t, arm)

	d := s.dim
	for i := 0; i < d; i++ {
		assert.InDelta(t, 1.0, arm.A.At(i, i), 1e-12, "A_a diagonal[%d]", i)
		assert.InDelta(t, 1.0, arm.Ainv.At(i, i), 1e-12, "Ainv diagonal[%d]", i)
		assert.InDelta(t, 0.0, arm.b.AtVec(i), 1e-12, "b_a[%d]", i)
		for j := 0; j < d; j++ {
			if i != j {
				assert.InDelta(t, 0.0, arm.A.At(i, j), 1e-12, "A_a off-diag (%d,%d)", i, j)
				assert.InDelta(t, 0.0, arm.Ainv.At(i, j), 1e-12, "Ainv off-diag (%d,%d)", i, j)
			}
		}
	}
}

// TestLinUCB_ShermanMorrisonMatchesBruteForce verifies that the
// incremental A^{-1} update equals a fresh inversion of the new A. This
// is the single most important correctness guard — if Sherman–Morrison
// drifts, the bandit's exploration bonus is wrong.
func TestLinUCB_ShermanMorrisonMatchesBruteForce(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})

	// Drive 50 full Select→RecordOutcome cycles so the production
	// pendingX caching path is exercised. Distinct task IDs guarantee
	// each outcome lands against the matching feature vector.
	rng := rand.New(rand.NewSource(7))
	for i := 0; i < 50; i++ {
		ctx := TaskContext{SourceSizeBytes: rng.Intn(1 << 20), TaskID: "task-" + string(rune('A'+i%26)) + string(rune('a'+(i/26)%26))}
		w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		require.NoError(t, err)
		s.RecordOutcome(w.ID, -rng.Float64()*5, true, ctx)
	}

	s.mu.Lock()
	arm := s.arms["worker-a"]
	if arm == nil {
		s.mu.Unlock()
		// 50 trials is enough to expect at least one dispatch to
		// worker-a; if not, fall back to the other worker so the test
		// still exercises Sherman-Morrison on a populated arm.
		s.mu.Lock()
		arm = s.arms["worker-b"]
	}
	require.NotNil(t, arm, "no arm received any update")
	A := mat.DenseCopyOf(arm.A)
	cachedInv := mat.DenseCopyOf(arm.Ainv)
	s.mu.Unlock()

	// Fresh inverse via gonum.
	d := s.dim
	freshInv := mat.NewDense(d, d, nil)
	require.NoError(t, freshInv.Inverse(A))

	// Compare element-wise with a tolerance that allows for floating
	// point drift over 50 rank-1 updates.
	for i := 0; i < d; i++ {
		for j := 0; j < d; j++ {
			assert.InDelta(t, freshInv.At(i, j), cachedInv.At(i, j), 1e-6, "Ainv (%d,%d)", i, j)
		}
	}
}

// TestLinUCB_LearnsBestArm runs the scheduler against a fixed reward
// model where one worker is consistently better than another. After a
// modest number of trials, the scheduler should pick the better worker
// at least 70% of the time. This validates the end-to-end learning
// loop, not just individual primitives.
func TestLinUCB_LearnsBestArm(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60_000_000_000) // 1 minute TTL in ns
	t.Cleanup(reg.Stop)

	// Big strong worker (ARCH match, 8 cores)
	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID:      "fast",
		Address: "127.0.0.1:0",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:  pb.Architecture_ARCH_X86_64,
			CpuCores:    8,
			MemoryBytes: 8 * 1024 * 1024 * 1024,
			Cpp:         &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))
	// Small weak worker (matched arch, 1 core)
	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID:      "slow",
		Address: "127.0.0.1:0",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:  pb.Architecture_ARCH_X86_64,
			CpuCores:    1,
			MemoryBytes: 1 * 1024 * 1024 * 1024,
			Cpp:         &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))

	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 0.5})

	rng := rand.New(rand.NewSource(11))
	rewards := map[string]float64{"fast": -3, "slow": -6}

	const trials = 200
	const warmup = 50
	fastPicks := 0
	for i := 0; i < trials; i++ {
		// TaskID binds the Select-time feature vector to the outcome via
		// pendingX; without it RecordOutcome drops every update and the
		// bandit never learns (the test then measures only the bonus).
		ctx := TaskContext{SourceSizeBytes: 100_000, TaskID: fmt.Sprintf("t-%d", i)}
		w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		require.NoError(t, err)
		if i >= warmup && w.ID == "fast" {
			fastPicks++
		}
		r := rewards[w.ID] + rng.NormFloat64()*0.5
		s.RecordOutcome(w.ID, r, true, ctx)
	}
	frac := float64(fastPicks) / float64(trials-warmup)
	assert.Greater(t, frac, 0.65, "expected >65%% picks on the fast worker after warm-up; got %.3f", frac)
}

// TestLinUCB_FastPathSingleCandidate ensures the fast path runs and
// reports zero exploration when only one worker is eligible.
func TestLinUCB_FastPathSingleCandidate(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 1.0})

	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
	require.NoError(t, err)
	assert.Equal(t, "worker-a", w.ID)
	assert.False(t, info.WasExploration)
	assert.Equal(t, 0.0, info.QValueAtDispatch)
}

// TestLinUCB_FeatureVectorDimensions sanity-checks that featureVector
// produces a vector of the expected length and that key features are
// populated correctly.
func TestLinUCB_FeatureVectorDimensions(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})

	w := &registry.WorkerInfo{
		ID: "x",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:  pb.Architecture_ARCH_ARM64,
			CpuCores:    8,
			MemoryBytes: 32 * 1024 * 1024 * 1024,
		},
		MaxParallel:     4,
		ActiveTasks:     2,
		SuccessfulTasks: 3,
		FailedTasks:     1,
	}
	ctx := TaskContext{
		SourceSizeBytes:    1 << 16,
		RawSourceSizeBytes: 32 * 1024,
		SourceFilename:     "obj.cpp",
	}
	x := s.featureVector(w, pb.Architecture_ARCH_ARM64, ctx)
	assert.Equal(t, s.dim, x.Len(), "feature dim must be 12 per the Hybrid-LinUCB layout")
	assert.Equal(t, 1.0, x.AtVec(0)) // bias
	// dim 1: log size feature; just check it is in the expected band.
	assert.Greater(t, x.AtVec(1), 0.5)
	assert.LessOrEqual(t, x.AtVec(1), 1.0)
	assert.Equal(t, 1.0, x.AtVec(3)) // ARM64 target
	assert.Equal(t, 1.0, x.AtVec(6)) // native arch matches target
	// dim 4: log-scale cpu_millis feature. No CpuMillis set on this test
	// worker (cgroup-less), so effectiveCPUMillis falls back to
	// CpuCores*1000 = 8000; log1p(8000)/log1p(16000).
	assert.InDelta(t, math.Log1p(8000)/math.Log1p(16000), x.AtVec(4), 1e-9)
	// dim 5: log-scale mem_bytes feature; log1p(32GiB)/log1p(64GiB).
	assert.InDelta(t, math.Log1p(32*1024*1024*1024)/math.Log1p(64*1024*1024*1024), x.AtVec(5), 1e-9)
	assert.InDelta(t, 0.5, x.AtVec(7), 1e-9) // 2/4 active tasks
	assert.Equal(t, 1.0, x.AtVec(9))         // .cpp is C++
	assert.InDelta(t, math.Log1p(32*1024)/math.Log1p(1024*1024), x.AtVec(10), 1e-9)
	assert.InDelta(t, 4.0/6.0, x.AtVec(11), 1e-9) // Laplace (3+1)/(3+1+2)
}

// TestLinUCB_FeatureVector_DistinguishesFractionalCPUQuotas is a
// regression test for the degenerate-feature bug documented in
// docs/thesis/bao-cao-tien-do-260809.md §1.5: workers under Docker --cpus quotas of
// 0.5/0.6/0.8/1.0/1.1 (see test/stress/docker-compose-hetero.yml) all
// reported cpu_cores=<host core count> before capability.Detect() read
// cgroup limits, so dim [4] was bit-for-bit identical (variance 0) across
// all five workers — the bandit had no CPU signal to learn from. This
// test asserts dim [4] now has strictly increasing, pairwise-distinct
// values across the same five real quotas, expressed as CpuMillis (what
// a cgroup-aware worker now reports).
func TestLinUCB_FeatureVector_DistinguishesFractionalCPUQuotas(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})
	ctx := TaskContext{SourceSizeBytes: 1 << 16, SourceFilename: "obj.c"}

	quotasMillis := []int32{500, 600, 800, 1000, 1100}
	var dims []float64
	for _, millis := range quotasMillis {
		w := &registry.WorkerInfo{
			ID: "w",
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_ARM64,
				CpuCores:   10, // host-reported core count — identical across all 5, as observed
				CpuMillis:  millis,
			},
			MaxParallel: 2,
		}
		x := s.featureVector(w, pb.Architecture_ARCH_ARM64, ctx)
		dims = append(dims, x.AtVec(4))
	}

	for i := 1; i < len(dims); i++ {
		assert.Greater(t, dims[i], dims[i-1],
			"dim[4] must strictly increase with CPU quota (%d millis -> %f, %d millis -> %f)",
			quotasMillis[i-1], dims[i-1], quotasMillis[i], dims[i])
	}

	// The old bug: with CpuMillis unset, all 5 collapse to whatever
	// cpu_cores/16 (or now, the fallback cpu_cores*1000 log-scaled)
	// gives — identical for identical cpu_cores. Assert that path is
	// still distinct from the cgroup-aware path for at least one quota,
	// so a future change that silently ignores CpuMillis again would
	// fail this test.
	fallback := &registry.WorkerInfo{
		ID:           "w",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_ARM64, CpuCores: 10},
		MaxParallel:  2,
	}
	fallbackDim := s.featureVector(fallback, pb.Architecture_ARCH_ARM64, ctx).AtVec(4)
	assert.NotEqual(t, dims[0], fallbackDim,
		"a worker with real CpuMillis=500 must score differently from one reporting no cgroup limit at all")
}

// TestLinUCB_RewardMonotonicity verifies that giving better rewards to
// one worker increases its θ̂ᵀx faster than another worker's. Uses the
// pendingX cache via Select→RecordOutcome so the test exercises the
// production path (post-CRITICAL-1 fix).
func TestLinUCB_RewardMonotonicity(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 0})

	// Manually insert a feature vector for each worker so we can fix
	// the rewards we feed back. Bypass Select to isolate the update
	// behaviour from the selection policy under α = 0.
	xa := s.featureVector(&registry.WorkerInfo{ID: "worker-a"}, pb.Architecture_ARCH_X86_64, TaskContext{SourceSizeBytes: 100_000})
	xb := s.featureVector(&registry.WorkerInfo{ID: "worker-b"}, pb.Architecture_ARCH_X86_64, TaskContext{SourceSizeBytes: 100_000})

	for i := 0; i < 30; i++ {
		idA := "ta-" + string(rune('A'+i%26))
		idB := "tb-" + string(rune('A'+i%26))
		s.mu.Lock()
		s.pendingX[idA] = xa
		s.pendingX[idB] = xb
		s.mu.Unlock()
		s.RecordOutcome("worker-a", -1.0, true, TaskContext{TaskID: idA})
		s.RecordOutcome("worker-b", -10.0, true, TaskContext{TaskID: idB})
	}

	meanA, _ := s.score("worker-a", xa)
	meanB, _ := s.score("worker-b", xb)
	assert.Greater(t, meanA, meanB, "worker-a should have higher θ̂ᵀx than worker-b")
}

// TestLinUCB_ConcurrentSafe runs many select+record loops in parallel
// under -race.
func TestLinUCB_ConcurrentSafe(t *testing.T) {
	reg := newRegistryWithWorkers(t, 4)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 0.5})

	const goroutines = 16
	const iterations = 50
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1000 + i})
				if err != nil {
					return
				}
				s.RecordOutcome(w.ID, -float64(i), true, TaskContext{SourceSizeBytes: 1000 + i})
			}
		}()
	}
	wg.Wait()
}

// TestLinUCB_RejectsInvalidInputs guards against NaN/Inf rewards and
// empty worker IDs poisoning the arm state.
func TestLinUCB_RejectsInvalidInputs(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 1})

	s.RecordOutcome("", 1.0, true, TaskContext{})
	s.RecordOutcome("worker-a", math.NaN(), true, TaskContext{SourceSizeBytes: 100})
	s.RecordOutcome("worker-a", math.Inf(1), true, TaskContext{SourceSizeBytes: 100})

	s.mu.Lock()
	_, exists := s.arms["worker-a"]
	s.mu.Unlock()
	// Arm may be lazily created on first valid call but should not have
	// absorbed the malformed observations.
	if exists {
		s.mu.Lock()
		arm := s.arms["worker-a"]
		var bSum float64
		for i := 0; i < s.dim; i++ {
			bSum += arm.b.AtVec(i)
		}
		s.mu.Unlock()
		assert.InDelta(t, 0.0, bSum, 1e-12, "b vector should be untouched after invalid inputs")
	}
}

// newHybridTestWorker builds a CPP-capable worker for hybrid tests.
func newHybridTestWorker(id string, cores int32, memGiB int64, active int32) *registry.WorkerInfo {
	return &registry.WorkerInfo{
		ID:      id,
		Address: "127.0.0.1:0",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:  pb.Architecture_ARCH_X86_64,
			CpuCores:    cores,
			MemoryBytes: memGiB * 1024 * 1024 * 1024,
			Cpp:         &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
		ActiveTasks: active,
	}
}

// TestLinUCB_WarmStartDelegatesToLeastLoadedAndLearns checks the two
// warm-start invariants: selection follows the least-loaded heuristic,
// and the bandit still learns passively from the delegated dispatch
// (pendingX cached, Sherman-Morrison update applied on outcome).
func TestLinUCB_WarmStartDelegatesToLeastLoadedAndLearns(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60_000_000_000)
	t.Cleanup(reg.Stop)
	require.NoError(t, reg.Add(newHybridTestWorker("busy", 4, 4, 3)))
	require.NoError(t, reg.Add(newHybridTestWorker("idle", 4, 4, 1)))
	require.NoError(t, reg.Add(newHybridTestWorker("half", 4, 4, 2)))

	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, WarmStartTasks: 10})

	ctx := TaskContext{SourceSizeBytes: 1024, TaskID: "w1"}
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
	require.NoError(t, err)
	assert.Equal(t, "idle", w.ID, "warm-start must pick the least-loaded worker")
	assert.True(t, info.WasExploration, "warm-start dispatches are flagged as exploration")

	s.mu.Lock()
	_, cached := s.pendingX["w1"]
	s.mu.Unlock()
	require.True(t, cached, "warm-start must cache the feature vector for passive learning")

	s.RecordOutcome(w.ID, -1.0, true, ctx)
	s.mu.Lock()
	arm := s.arms[w.ID]
	var bSum float64
	for i := 0; i < s.dim; i++ {
		bSum += math.Abs(arm.b.AtVec(i))
	}
	count := arm.count
	s.mu.Unlock()
	assert.Equal(t, int64(1), count, "passive learning must apply the rank-1 update")
	assert.Greater(t, bSum, 0.0, "b must absorb the delegated dispatch's reward")
}

// TestLinUCB_WarmStartHandsOverAfterN checks the warm-start boundary:
// dispatches 1..N go least-loaded, dispatch N+1 switches to the UCB
// argmax (with θ = 0 the bonus ∝ ‖x‖ favors the bigger worker).
func TestLinUCB_WarmStartHandsOverAfterN(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60_000_000_000)
	t.Cleanup(reg.Stop)
	require.NoError(t, reg.Add(newHybridTestWorker("big", 8, 8, 1)))
	require.NoError(t, reg.Add(newHybridTestWorker("small", 1, 1, 0)))

	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: 1, WarmStartTasks: 2})

	for i := 1; i <= 2; i++ {
		w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
		require.NoError(t, err)
		assert.Equal(t, "small", w.ID, "dispatch %d is inside the warm-start window", i)
	}

	w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
	require.NoError(t, err)
	assert.Equal(t, "big", w.ID, "dispatch 3 must come from the UCB argmax (bonus ∝ ‖x‖)")
}

// TestLinUCB_LoadPenaltySteersAwayFromLoadedWorker isolates the λ
// term: with exploration disabled (Alpha < 0 — note the constructor
// maps Alpha == 0 to the 0.5 default) and θ = 0, the only score
// difference between two identical workers is −λ·loadRatio.
func TestLinUCB_LoadPenaltySteersAwayFromLoadedWorker(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60_000_000_000)
	t.Cleanup(reg.Stop)
	require.NoError(t, reg.Add(newHybridTestWorker("loaded", 4, 4, 3)))
	require.NoError(t, reg.Add(newHybridTestWorker("free", 4, 4, 0)))

	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg, Alpha: -1, LoadPenalty: 1.0})

	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
	require.NoError(t, err)
	assert.Equal(t, "free", w.ID, "penalty must steer away from the loaded worker")
	assert.False(t, info.WasExploration)
}

// TestIsCppSource covers the extension table and the compiler-driver
// fallback (a "++" driver compiles even .c inputs as C++).
func TestIsCppSource(t *testing.T) {
	cases := []struct {
		filename string
		compiler string
		want     bool
	}{
		{"main.cpp", "g++", true},
		{"main.c", "gcc", false},
		{"a.cc", "", true},
		{"a.cxx", "", true},
		{"Foo.C", "", true},
		{"foo.hpp", "", true},
		{"main.c", "clang++", true},
		{"", "gcc", false},
		{"", "g++", true},
	}
	for _, tc := range cases {
		t.Run(tc.filename+"/"+tc.compiler, func(t *testing.T) {
			assert.Equal(t, tc.want, isCppSource(tc.filename, tc.compiler))
		})
	}
}

// TestLinUCB_SuccessRatePrior checks the Laplace-smoothed success-rate
// feature: (s+1)/(s+f+2), so an unobserved worker sits at 0.5 and the
// column is never constant even on all-success workloads.
func TestLinUCB_SuccessRatePrior(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})

	cases := []struct {
		succ, fail int64
		want       float64
	}{
		{0, 0, 0.5},
		{3, 1, 4.0 / 6.0},
		{0, 2, 1.0 / 4.0},
	}
	for _, tc := range cases {
		w := &registry.WorkerInfo{ID: "x", SuccessfulTasks: tc.succ, FailedTasks: tc.fail}
		x := s.featureVector(w, pb.Architecture_ARCH_X86_64, TaskContext{})
		assert.InDelta(t, tc.want, x.AtVec(11), 1e-9, "success rate for %d/%d", tc.succ, tc.fail)
	}
}

// TestLinUCB_TotalDispatchesCountsFastPath pins the counter semantics:
// totalDispatches measures dispatch volume, so single-candidate
// fast-path dispatches count too.
func TestLinUCB_TotalDispatchesCountsFastPath(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: reg})

	for i := 0; i < 3; i++ {
		_, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 1024})
		require.NoError(t, err)
	}
	assert.Equal(t, int64(3), atomic.LoadInt64(&s.totalDispatches))
}
