package scheduler

import (
	"math"
	"path"
	"strings"
	"sync"
	"sync/atomic"

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

	// warmStartTasks configures the Hybrid-LinUCB warm-start window:
	// dispatches 1..N are routed by the least-loaded heuristic while
	// the bandit learns passively from their outcomes. 0 = pure LinUCB.
	warmStartTasks int64
	// loadPenalty is λ in the hybrid score p = mean + bonus − λ·loadRatio.
	// It acts as a hand-coded prior on the load feature: the model
	// cannot trust its learned load weight during early training, and
	// delayed rewards would otherwise let the scheduler pile tasks onto
	// one fast worker before any feedback arrives. 0 disables it.
	loadPenalty float64
	// totalDispatches counts every SelectWithDispatchInfo call,
	// including single-candidate fast-path dispatches — it measures
	// dispatch volume, not learning events. Accessed atomically; never
	// reset.
	totalDispatches int64

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
	// WarmStartTasks is the Hybrid-LinUCB warm-start window length.
	// Zero (the default) keeps pure-LinUCB behavior.
	WarmStartTasks int
	// LoadPenalty is the Hybrid-LinUCB λ coefficient. Zero (the
	// default) keeps pure-LinUCB behavior.
	LoadPenalty float64
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
	warmStart := cfg.WarmStartTasks
	if warmStart < 0 {
		warmStart = 0
	}
	loadPenalty := cfg.LoadPenalty
	if loadPenalty < 0 {
		loadPenalty = 0
	}
	return &LinUCBScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
		latencyTracker: lt,
		alpha:          alpha,
		dim:            featureDim(),
		warmStartTasks: int64(warmStart),
		loadPenalty:    loadPenalty,
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
//
// In hybrid mode (warmStartTasks/loadPenalty non-zero) the score is
// p = θ̂ᵀx + α√(xᵀA⁻¹x) − λ·loadRatio, and the first warmStartTasks
// dispatches are delegated to the least-loaded heuristic while the
// bandit observes their outcomes (docs/thesis/hybrid_linucb_proposal.md).
func (s *LinUCBScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := s.eligibleWorkers(buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}
	n := atomic.AddInt64(&s.totalDispatches, 1)
	if len(candidates) == 1 {
		// Fast path: no choice. Skip matrix ops entirely.
		return candidates[0], DispatchInfo{QValueAtDispatch: 0, WasExploration: false}, nil
	}
	if s.warmStartTasks > 0 && n <= s.warmStartTasks {
		return s.selectWarmStart(candidates, arch, ctx)
	}

	// Single pass: track argmax over (mean−λ·lr+bonus) for selection and
	// argmax over (mean−λ·lr) alone for the exploration flag. A dispatch
	// counts as exploration when the UCB winner is not the penalized-mean
	// argmax — i.e. the bonus, not the learned+heuristic value, drove the
	// choice.
	var best *registry.WorkerInfo
	var bestX *mat.VecDense
	var bestMeanID string
	bestP := math.Inf(-1)
	bestMean := math.Inf(-1)
	for _, w := range candidates {
		x := s.featureVector(w, arch, ctx)
		mean, bonus := s.score(w.ID, x)
		penalized := mean - s.loadPenalty*loadRatio(w)
		p := penalized + bonus
		if p > bestP {
			bestP = p
			best = w
			bestX = x
		}
		if penalized > bestMean {
			bestMean = penalized
			bestMeanID = w.ID
		}
	}
	wasExploration := bestMeanID != "" && bestMeanID != best.ID

	// Cache the chosen worker's feature vector so RecordOutcome updates
	// the bandit against the same x that drove selection.
	if ctx.TaskID != "" && bestX != nil {
		s.mu.Lock()
		s.pendingX[ctx.TaskID] = bestX
		s.mu.Unlock()
	}

	return best, DispatchInfo{QValueAtDispatch: bestP, WasExploration: wasExploration}, nil
}

// selectWarmStart routes a warm-start dispatch to the least-loaded
// candidate (min ActiveTasks, first-wins tie-break — mirroring
// LeastLoadedScheduler.Select) while the bandit learns passively: the
// chosen worker's feature vector is cached in pendingX so RecordOutcome
// applies the normal rank-1 update, exactly as if the bandit had made
// the choice itself.
//
// DispatchInfo semantics: QValueAtDispatch reports the learner's hybrid
// score for the delegated choice so offline analysis can watch Q
// converge during the warm-start window; WasExploration is true because
// the pick was not the learner's greedy argmax, which lets analysis
// split the warm-start window without relying on record ordering.
//
// Lock ordering: s.score locks s.mu internally, so it must complete
// before the pendingX store takes s.mu — never call it with the lock
// held.
func (s *LinUCBScheduler) selectWarmStart(candidates []*registry.WorkerInfo, arch pb.Architecture, ctx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	chosen := candidates[0]
	for _, w := range candidates[1:] {
		if w.ActiveTasks < chosen.ActiveTasks {
			chosen = w
		}
	}

	x := s.featureVector(chosen, arch, ctx)
	mean, bonus := s.score(chosen.ID, x)
	q := mean + bonus - s.loadPenalty*loadRatio(chosen)

	if ctx.TaskID != "" {
		s.mu.Lock()
		s.pendingX[ctx.TaskID] = x
		s.mu.Unlock()
	}

	return chosen, DispatchInfo{QValueAtDispatch: q, WasExploration: true}, nil
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
// Layout (12 dims) — derived from the paper-skeleton.md §3.3 design but
// pruned per code-review finding CRITICAL-2: build-type one-hot dims
// were always set to (1, 0, 0) under the current Compile() path,
// making them perfectly collinear with the bias. Removing them keeps
// the design space identifiable and frees Sherman-Morrison from a
// degenerate rank-2 subspace during warm-up. Build-type can be added
// back when Flutter/Unity reach the learning path.
//
// Dims 9-11 extend the design per the Hybrid-LinUCB proposal
// (docs/thesis/hybrid_linucb_proposal.md §3): task-language and raw
// task weight give the model a handle on the "goods" being compiled,
// and worker success rate exposes reliability the hardware dims miss.
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
//	[9]   task language is C++                                      (1.0 / 0.0)
//	[10]  log(1 + raw_source_size_bytes) / log(1 + 1 MiB)           (≈ [0, 1])
//	[11]  worker success rate, Laplace-smoothed (s+1)/(s+f+2)
func featureDim() int { return 12 }

// sizeNormDenom is log1p of a "typical big translation unit" — a 4 MiB
// preprocessed source. Using this denominator keeps the size feature
// in roughly [0, 1] for the workloads we observe (M1 P99 ≈ 2.3 MiB),
// avoiding the prior /16 divisor that compressed all real values into
// [0.6, 0.8] and made the feature near-constant for the benchmark
// (code-review finding MED-4).
var sizeNormDenom = math.Log1p(4 * 1024 * 1024)

// rawSizeNormDenom normalizes raw (unpreprocessed) source size. Raw
// sources lack expanded headers so they run 10-100x smaller than
// preprocessed TUs; a 1 MiB denominator covers the raw-file range of
// the observed workloads and keeps the feature spread across (0, 1]
// instead of clustering near zero under the 4 MiB preprocessed
// denominator.
var rawSizeNormDenom = math.Log1p(1024 * 1024)

// cppExtensions are the file extensions (lowercased) that identify a
// C++ translation unit. ".C" (uppercase) is handled separately since
// by Unix convention it means C++ while ".c" means C.
var cppExtensions = map[string]bool{
	".cpp": true, ".cc": true, ".cxx": true, ".c++": true,
	".hpp": true, ".hh": true,
}

// isCppSource reports whether the task compiles C++ rather than C.
// The file extension decides when recognized; otherwise a "++"
// compiler driver (g++, clang++) decides, matching GCC semantics
// where g++ compiles even .c inputs as C++.
func isCppSource(filename, compiler string) bool {
	ext := path.Ext(filename)
	if ext == ".C" {
		return true
	}
	if cppExtensions[strings.ToLower(ext)] {
		return true
	}
	return strings.Contains(compiler, "++")
}

// loadRatio returns ActiveTasks/MaxParallel clamped to [0, 1], using
// the same MaxParallel<=0 → 4 default as eligibleWorkers. Feature [7]
// and the hybrid load penalty must observe the identical value, so
// both go through this helper.
func loadRatio(w *registry.WorkerInfo) float64 {
	maxP := w.MaxParallel
	if maxP <= 0 {
		maxP = 4
	}
	lr := float64(w.ActiveTasks) / float64(maxP)
	if lr > 1.0 {
		lr = 1.0
	}
	return lr
}

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

	x.SetVec(7, loadRatio(w))

	rttNorm := s.latencyTracker.Get(w.ID) / 100.0
	if rttNorm > 1.0 {
		rttNorm = 1.0
	}
	x.SetVec(8, rttNorm)

	if isCppSource(ctx.SourceFilename, ctx.Compiler) {
		x.SetVec(9, 1.0)
	}

	rawSize := math.Log1p(float64(ctx.RawSourceSizeBytes)) / rawSizeNormDenom
	if rawSize > 1.0 {
		rawSize = 1.0
	}
	x.SetVec(10, rawSize)

	// Laplace-smoothed success rate (s+1)/(s+f+2). The raw rate with an
	// optimistic prior of 1.0 is exactly 1.0 for every worker on a
	// fully-successful workload — perfectly collinear with the bias dim,
	// the same degeneracy class removed in CRITICAL-2. Smoothing keeps
	// the value experience-dependent (0/0 → 0.5, rising toward the true
	// rate) so the column is never constant.
	total := w.SuccessfulTasks + w.FailedTasks
	x.SetVec(11, float64(w.SuccessfulTasks+1)/float64(total+2))

	return x
}

// mulMatVec returns A · v as a new VecDense.
func mulMatVec(A *mat.Dense, v *mat.VecDense) *mat.VecDense {
	r, _ := A.Dims()
	out := mat.NewVecDense(r, nil)
	out.MulVec(A, v)
	return out
}
