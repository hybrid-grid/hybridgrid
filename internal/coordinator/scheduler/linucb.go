package scheduler

import (
	"math"
	"sync"

	"gonum.org/v1/gonum/mat"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// LinUCBScheduler implements the disjoint linear contextual bandit
// algorithm of Li, Chu, Langford & Schapire (2010), "A Contextual-Bandit
// Approach to Personalized News Article Recommendation," WWW '10
// (DOI: 10.1145/1772690.1772758, arXiv:1003.0146 v2).
//
// Algorithm 1 (verbatim from §3.1) maintains for each arm a:
//
//	A_a ∈ ℝ^{d×d}, initialized to I_d                        (line 5)
//	b_a ∈ ℝ^d,    initialized to 0                          (line 6)
//	θ̂_a = A_a^{-1} b_a                                     (line 8, ridge LS)
//	p_{t,a} = θ̂_a^T x + α √(x^T A_a^{-1} x)                (line 9, UCB score)
//
// Selection: a_t = argmax p_{t,a}.
// Update on observed reward r:
//
//	A_{a_t} ← A_{a_t} + x x^T                                (line 12, rank-1)
//	b_{a_t} ← b_{a_t} + r x                                  (line 13)
//
// We maintain A_a^{-1} incrementally with the Sherman–Morrison formula
// (Sherman & Morrison 1950; Golub & Van Loan, Matrix Computations 4th
// ed., §2.1.4) so each update costs O(d²):
//
//	A_new^{-1} = A_old^{-1} - (A_old^{-1} x x^T A_old^{-1})
//	             ───────────────────────────────────────────
//	                       1 + x^T A_old^{-1} x
//
// The exploration parameter α has the theoretical form
//
//	α = 1 + sqrt(ln(2/δ) / 2)               (Li 2010, Eq. 4)
//
// for confidence 1−δ. The paper notes this is "conservatively large in
// some applications" and that practical tuning is common. We expose α
// as a knob and default to 1.0 (matches the conservative δ → 0 limit's
// rough magnitude). The thesis must report the value chosen and label
// it as empirically tuned.
//
// Regret guarantees from Chu, Li, Reyzin & Schapire (2011), "Contextual
// Bandits with Linear Payoff Functions," AISTATS, are stated for
// SupLinUCB (a variant that decouples confidence updates across
// elimination phases). Plain LinUCB Algorithm 1 is what we implement;
// its tight regret proof remains open. We cite Chu 2011's Theorem 1
// (regret O(√(Td log³(KT log T / δ)))) as the theoretical context but
// not as a direct guarantee for this code.
//
// LinUCB is sensitive to non-linear reward structure and to drift in
// θ_a^*. See docs/thesis/theory-notes.md §3.4 and §5 for caveats.
type LinUCBScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker
	latencyTracker *metrics.LatencyTracker
	alpha          float64
	dim            int

	mu   sync.Mutex
	arms map[string]*linUCBArm
	// pendingX caches the feature vector observed at Select time for
	// each in-flight task. RecordOutcome consumes the cached value so
	// the bandit update sees the same x that drove selection — the
	// alternative (rebuilding x from registry state at outcome time)
	// suffers from target leakage because ActiveTasks and RTT have
	// already been mutated by the time the outcome arrives. Keyed by
	// TaskContext.TaskID; entries are deleted on consumption.
	pendingX map[string]*mat.VecDense
}

// linUCBArm holds the per-worker bandit state. We keep both A and its
// inverse so we can reconstruct from disk in the future and sanity-check
// the Sherman-Morrison update against a fresh inversion in tests.
type linUCBArm struct {
	A     *mat.Dense // d×d
	Ainv  *mat.Dense // d×d cached inverse
	b     *mat.VecDense
	theta *mat.VecDense // A^{-1} b, recomputed lazily after updates
	dirty bool          // theta needs recomputing
	count int64
}

// LinUCBConfig holds construction parameters.
type LinUCBConfig struct {
	Registry       registry.Registry
	CircuitChecker CircuitChecker
	LatencyTracker *metrics.LatencyTracker
	// Alpha is the UCB exploration coefficient α. Default 1.0. Per
	// Li 2010 Eq. (4) the theoretical form is 1 + sqrt(ln(2/δ)/2);
	// practical tuning typically lies in [0.1, 2.0]. Values ≤ 0
	// disable exploration (pure greedy on θ̂_a^T x).
	Alpha float64
}

// NewLinUCBScheduler constructs the scheduler. The feature dimension
// is fixed at the value returned by featureDim() to keep the state
// shape stable across the cluster's lifetime.
//
// Default α is 0.5. The Li 2010 theoretical form (Eq. 4) is
// "conservatively large in some applications" per the paper itself;
// 0.5 is in the practical range Chu et al. 2011 §5 reported, and
// keeps the UCB bonus from dominating the mean estimate during the
// short warm-up our build workloads expose (≤ 900 dispatches per run).
func NewLinUCBScheduler(cfg LinUCBConfig) *LinUCBScheduler {
	alpha := cfg.Alpha
	if alpha == 0 {
		alpha = 0.5
	}
	if alpha < 0 {
		alpha = 0
	}
	lt := cfg.LatencyTracker
	if lt == nil {
		lt = metrics.NewLatencyTracker()
	}
	return &LinUCBScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
		latencyTracker: lt,
		alpha:          alpha,
		dim:            featureDim(),
		arms:           make(map[string]*linUCBArm),
		pendingX:       make(map[string]*mat.VecDense),
	}
}

// Select implements the base Scheduler interface; it delegates to the
// learner path with a zero-valued TaskContext.
func (s *LinUCBScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS, TaskContext{})
	return w, err
}

// SelectWithDispatchInfo implements LearningScheduler. It computes
// p_{t,a} for every eligible worker and returns argmax along with the
// score (Q value) and an exploration flag. We mark a dispatch as
// "exploration" when the chosen arm's UCB bonus exceeds its mean term —
// i.e. selection was driven by uncertainty rather than learned value.
func (s *LinUCBScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := s.eligibleWorkers(buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}
	if len(candidates) == 1 {
		// Fast path: no choice. Skip matrix ops entirely.
		return candidates[0], DispatchInfo{QValueAtDispatch: 0, WasExploration: false}, nil
	}

	var best *registry.WorkerInfo
	var bestX *mat.VecDense
	bestP := math.Inf(-1)
	for _, w := range candidates {
		x := s.featureVector(w, arch, ctx)
		mean, bonus := s.score(w.ID, x)
		p := mean + bonus
		if p > bestP {
			bestP = p
			best = w
			bestX = x
		}
	}

	// Compute exploration flag honestly: the dispatch counts as
	// exploration if the winner's selection was not the argmax of the
	// pure mean θ̂ᵀx. We re-evaluate means alone and compare to the UCB
	// winner — if a different arm has a higher mean, the bonus drove
	// the choice (true exploration).
	wasExploration := false
	{
		bestMeanID := ""
		bestMean := math.Inf(-1)
		for _, w := range candidates {
			x := s.featureVector(w, arch, ctx)
			m, _ := s.score(w.ID, x)
			if m > bestMean {
				bestMean = m
				bestMeanID = w.ID
			}
		}
		if bestMeanID != "" && bestMeanID != best.ID {
			wasExploration = true
		}
	}

	// Cache the chosen worker's feature vector so RecordOutcome updates
	// the bandit against the same x that drove selection.
	if ctx.TaskID != "" && bestX != nil {
		s.mu.Lock()
		s.pendingX[ctx.TaskID] = bestX
		s.mu.Unlock()
	}

	return best, DispatchInfo{QValueAtDispatch: bestP, WasExploration: wasExploration}, nil
}

// score returns the mean estimate (θ̂^T x) and the UCB exploration bonus
// (α √(x^T A^{-1} x)) for the given arm and context.
func (s *LinUCBScheduler) score(workerID string, x *mat.VecDense) (mean, bonus float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	arm := s.armForLocked(workerID)
	if arm.dirty {
		arm.theta = mulMatVec(arm.Ainv, arm.b)
		arm.dirty = false
	}
	mean = mat.Dot(arm.theta, x)
	// x^T A^{-1} x — symmetric quadratic form
	tmp := mulMatVec(arm.Ainv, x)
	q := mat.Dot(x, tmp)
	if q < 0 {
		// numerical: A^{-1} should be PSD, but rounding can produce
		// tiny negatives. Clamp to zero.
		q = 0
	}
	bonus = s.alpha * math.Sqrt(q)
	return mean, bonus
}

// RecordOutcome implements LearningScheduler. It applies the rank-1
// updates from Algorithm 1 lines 12–13 and refreshes the cached A^{-1}
// via Sherman–Morrison.
//
// We tolerate NaN/Inf rewards by skipping the update — the learner
// must not be poisoned by malformed observations. The feature vector
// used for the update is the one cached at Select time keyed by
// ctx.TaskID; if no cached vector is available (legacy callers, or
// the scheduler restarted between Select and RecordOutcome) we skip
// the update rather than fall back to a stale reconstruction that
// would inject target-leaked features into θ̂.
func (s *LinUCBScheduler) RecordOutcome(workerID string, reward float64, _ bool, ctx TaskContext) {
	if workerID == "" || math.IsNaN(reward) || math.IsInf(reward, 0) {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	x, ok := s.pendingX[ctx.TaskID]
	if !ok {
		// No cached x — either the caller did not set TaskID or the
		// dispatch happened before this scheduler was constructed.
		// Drop the update to avoid biasing θ̂ with a reconstructed,
		// post-completion feature vector.
		return
	}
	delete(s.pendingX, ctx.TaskID)

	arm := s.armForLocked(workerID)

	// b ← b + r x  (Algorithm 1 line 13)
	for i := 0; i < s.dim; i++ {
		arm.b.SetVec(i, arm.b.AtVec(i)+reward*x.AtVec(i))
	}

	// Sherman–Morrison incremental update for A^{-1}:
	// let u = A^{-1} x, denom = 1 + x^T u
	// A_new^{-1} = A^{-1} - (u u^T) / denom
	u := mulMatVec(arm.Ainv, x)
	denom := 1.0 + mat.Dot(x, u)
	if denom <= 0 {
		// Should never happen if A is PSD and grows by xxᵀ; refuse the
		// update rather than corrupting state.
		return
	}
	// outer = u u^T / denom
	outer := mat.NewDense(s.dim, s.dim, nil)
	for i := 0; i < s.dim; i++ {
		ui := u.AtVec(i)
		for j := 0; j < s.dim; j++ {
			outer.Set(i, j, ui*u.AtVec(j)/denom)
		}
	}
	// A^{-1} ← A^{-1} - outer
	arm.Ainv.Sub(arm.Ainv, outer)

	// A ← A + x x^T (kept for diagnostics and persistence)
	for i := 0; i < s.dim; i++ {
		xi := x.AtVec(i)
		for j := 0; j < s.dim; j++ {
			arm.A.Set(i, j, arm.A.At(i, j)+xi*x.AtVec(j))
		}
	}

	arm.count++
	arm.dirty = true
}

// armForLocked returns (creating if needed) the bandit state for a
// worker. Caller must hold s.mu.
func (s *LinUCBScheduler) armForLocked(workerID string) *linUCBArm {
	if a, ok := s.arms[workerID]; ok {
		return a
	}
	d := s.dim
	A := mat.NewDense(d, d, nil)
	Ainv := mat.NewDense(d, d, nil)
	for i := 0; i < d; i++ {
		A.Set(i, i, 1)    // A_a = I_d (Li 2010 Algorithm 1 line 5)
		Ainv.Set(i, i, 1) // I^{-1} = I
	}
	arm := &linUCBArm{
		A:     A,
		Ainv:  Ainv,
		b:     mat.NewVecDense(d, nil),
		theta: mat.NewVecDense(d, nil),
		dirty: false,
	}
	s.arms[workerID] = arm
	return arm
}

// eligibleWorkers applies the same admission rules as P2C and ε-greedy.
func (s *LinUCBScheduler) eligibleWorkers(buildType pb.BuildType, arch pb.Architecture, clientOS string) ([]*registry.WorkerInfo, error) {
	workers := s.registry.ListByCapability(buildType, arch)
	if len(workers) == 0 {
		if s.registry.Count() == 0 {
			return nil, ErrNoWorkers
		}
		return nil, ErrNoMatchingWorkers
	}
	if clientOS != "" {
		workers = filterByOS(workers, clientOS)
		if len(workers) == 0 {
			return nil, ErrNoMatchingWorkers
		}
	}
	candidates := make([]*registry.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		if w.State == registry.WorkerStateUnhealthy {
			continue
		}
		if s.circuitChecker != nil && s.circuitChecker.IsOpen(w.ID) {
			continue
		}
		maxParallel := w.MaxParallel
		if maxParallel <= 0 {
			maxParallel = 4
		}
		if w.ActiveTasks >= maxParallel {
			continue
		}
		candidates = append(candidates, w)
	}
	if len(candidates) == 0 {
		for _, w := range workers {
			if w.State == registry.WorkerStateUnhealthy {
				continue
			}
			maxParallel := w.MaxParallel
			if maxParallel <= 0 {
				maxParallel = 4
			}
			if w.ActiveTasks < maxParallel {
				candidates = append(candidates, w)
			}
		}
	}
	if len(candidates) == 0 {
		return nil, ErrNoMatchingWorkers
	}
	return candidates, nil
}

// featureDim is the fixed feature-vector dimension. Increasing this
// requires a one-time rebuild of all arm states.
//
// Layout (9 dims) — derived from the paper-skeleton.md §3.3 design but
// pruned per code-review finding CRITICAL-2: build-type one-hot dims
// were always set to (1, 0, 0) under the current Compile() path,
// making them perfectly collinear with the bias. Removing them keeps
// the design space identifiable and frees Sherman-Morrison from a
// degenerate rank-2 subspace during warm-up. Build-type can be added
// back when Flutter/Unity reach the learning path.
//
//	[0]   bias                                                      = 1.0
//	[1]   log(1 + source_size_bytes) / log(1 + 4 MiB)               (≈ [0, 1])
//	[2]   target_arch == X86_64
//	[3]   target_arch == ARM64
//	[4]   worker.cpu_cores / 16                                     (capped at 1.0)
//	[5]   worker.mem_bytes / (64 * 2^30)                            (capped at 1.0)
//	[6]   worker.native_arch == target_arch                         (1.0 / 0.0)
//	[7]   worker.active_tasks / max_parallel
//	[8]   worker.recent_rpc_latency_ms / 100                        (capped at 1.0)
func featureDim() int { return 9 }

// sizeNormDenom is log1p of a "typical big translation unit" — a 4 MiB
// preprocessed source. Using this denominator keeps the size feature
// in roughly [0, 1] for the workloads we observe (M1 P99 ≈ 2.3 MiB),
// avoiding the prior /16 divisor that compressed all real values into
// [0.6, 0.8] and made the feature near-constant for the benchmark
// (code-review finding MED-4).
var sizeNormDenom = math.Log1p(4 * 1024 * 1024)

// featureVector builds x_{t,a} for a given (worker, target_arch, ctx).
// All features are normalized roughly to [0, 1] so ‖x‖ stays bounded —
// matching the Chu 2011 convention.
func (s *LinUCBScheduler) featureVector(w *registry.WorkerInfo, targetArch pb.Architecture, ctx TaskContext) *mat.VecDense {
	d := s.dim
	x := mat.NewVecDense(d, nil)
	x.SetVec(0, 1.0) // bias

	logSize := math.Log1p(float64(ctx.SourceSizeBytes)) / sizeNormDenom
	if logSize > 1.0 {
		logSize = 1.0
	}
	x.SetVec(1, logSize)

	// target arch one-hot (build type omitted — see featureDim doc).
	switch targetArch {
	case pb.Architecture_ARCH_X86_64:
		x.SetVec(2, 1.0)
	case pb.Architecture_ARCH_ARM64:
		x.SetVec(3, 1.0)
	}

	caps := w.Capabilities
	var cpuCores, memBytes float64
	if caps != nil {
		cpuCores = float64(caps.CpuCores) / 16.0
		if cpuCores > 1.0 {
			cpuCores = 1.0
		}
		memBytes = float64(caps.MemoryBytes) / (64.0 * 1024 * 1024 * 1024)
		if memBytes > 1.0 {
			memBytes = 1.0
		}
	}
	x.SetVec(4, cpuCores)
	x.SetVec(5, memBytes)

	if caps != nil && caps.NativeArch == targetArch {
		x.SetVec(6, 1.0)
	}

	maxP := w.MaxParallel
	if maxP <= 0 {
		maxP = 4
	}
	loadRatio := float64(w.ActiveTasks) / float64(maxP)
	if loadRatio > 1.0 {
		loadRatio = 1.0
	}
	x.SetVec(7, loadRatio)

	rttNorm := s.latencyTracker.Get(w.ID) / 100.0
	if rttNorm > 1.0 {
		rttNorm = 1.0
	}
	x.SetVec(8, rttNorm)

	return x
}

// mulMatVec returns A · v as a new VecDense.
func mulMatVec(A *mat.Dense, v *mat.VecDense) *mat.VecDense {
	r, _ := A.Dims()
	out := mat.NewVecDense(r, nil)
	out.MulVec(A, v)
	return out
}
