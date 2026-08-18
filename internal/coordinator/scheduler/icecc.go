package scheduler

import (
	"math"
	"sync"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// IceccFastestScheduler is a faithful port of icecream's default
// scheduling algorithm ("fastest"), read from scheduler/scheduler.cpp of
// github.com/icecc/icecream at master (2608 lines, retrieved 2026-08-02).
//
// Why this exists: the thesis needs a like-for-like comparison against the
// scheduler that production distributed-compilation clusters actually run.
// Benchmarking `icecc` end-to-end would confound scheduler quality with
// transport, compression and toolchain-shipping differences; reimplementing
// its selection rule inside Hybrid-Grid isolates the algorithm, so a
// difference in makespan is attributable to the decision rule alone.
//
// Contrary to how the distributed-build literature usually characterises
// these systems, icecream's rule is NOT a static heuristic: it estimates
// each node's throughput online from completed jobs, boosts under-sampled
// nodes, and dispatches to never-measured nodes first. In bandit terms it
// is a hand-tuned, context-free value estimator with ad-hoc exploration.
//
// The upstream scoring function, verbatim:
//
//	float f = (float)cs->cumCompiled().outputSize()
//	          / (float) cs->cumCompiled().compileTimeUser();   // line 263
//	f *= float(1000 - cs->load()) / 1000;                      // line 311
//	f *= (1.0f - (0.5f * cs->currentJobCount() / cs->maxJobs()));  // line 319
//	if (cs->lastCompiledJobs().size() < 7)
//	    f *= (-0.5 * cs->lastCompiledJobs().size() + 4.5);     // line 323
//
// selected by argmax f, but preceded by pick_server_new() which returns any
// node holding zero statistics, and by a uniform-random pick while the
// cluster as a whole has recorded nothing.
//
// # Two documented deviations
//
// Both are forced by what the Hybrid-Grid protocol carries, and both are
// recorded here so the paper can state them rather than have a reviewer
// find them.
//
//  1. Work proxy. icecream measures work as the *object file* size the job
//     produced, normalising it for -g/-O flags (add_job_stats, line 148).
//     Our worker never reports object size, so we substitute the
//     preprocessed source size already present in TaskContext. Both are
//     stand-ins for "how much work was this translation unit"; the
//     substitution changes the constant of proportionality, not the shape
//     of the estimator, because the ratio is taken per node against its own
//     accumulated time.
//
//  2. System load. icecream's daemons report kernel load average on a
//     0..1000 scale, which we do not collect. The `(1000-load)/1000` factor
//     is therefore omitted. The queue-saturation factor on the following
//     line survives intact, and it carries the same signal from a different
//     source (assigned jobs rather than OS load), so the omission weakens
//     but does not remove icecream's load awareness.
type IceccFastestScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker

	mu sync.Mutex
	// stats holds the rolling per-node job history keyed by worker ID.
	stats map[string]*iceccNodeStats
	// recorded counts observations across the whole cluster, mirroring
	// icecream's all_job_stats: while it is zero, pick_server_fastest
	// falls through to a uniform-random choice.
	recorded int64
}

// iceccWindowSize mirrors icecream's cap on CompileServer::lastCompiledJobs,
// which retains at most the 200 most recent jobs per node so a node that has
// changed speed is not judged forever by stale samples.
const iceccWindowSize = 200

// iceccPessimismSamples is the sample count below which icecream inflates a
// node's speed (scheduler.cpp line 322). Despite the upstream comment naming
// it a "pessimism factor", multiplying a *speed* by (-0.5n + 4.5) — 4.0x at
// one sample, decaying to 1.5x at six — makes an under-sampled node MORE
// likely to win argmax. It is optimism under uncertainty: a hand-rolled UCB
// bonus that decays with n, without the variance derivation.
const iceccPessimismSamples = 7

type iceccJobStat struct {
	work   float64 // work proxy (preprocessed source bytes)
	timeMs float64 // observed compile time
}

type iceccNodeStats struct {
	window  []iceccJobStat
	cumWork float64
	cumTime float64
}

// add appends an observation, evicting the oldest once the window is full so
// that cumWork/cumTime always describe exactly the retained samples.
func (n *iceccNodeStats) add(st iceccJobStat) {
	n.window = append(n.window, st)
	n.cumWork += st.work
	n.cumTime += st.timeMs
	if len(n.window) > iceccWindowSize {
		old := n.window[0]
		n.window = n.window[1:]
		n.cumWork -= old.work
		n.cumTime -= old.timeMs
	}
}

// IceccConfig holds construction parameters.
type IceccConfig struct {
	Registry       registry.Registry
	CircuitChecker CircuitChecker
}

// NewIceccFastestScheduler constructs the scheduler.
func NewIceccFastestScheduler(cfg IceccConfig) *IceccFastestScheduler {
	return &IceccFastestScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
		stats:          make(map[string]*iceccNodeStats),
	}
}

// Select implements the base Scheduler interface.
func (s *IceccFastestScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS, TaskContext{})
	return w, err
}

// SelectWithDispatchInfo implements LearningScheduler, reproducing
// pick_server_fastest in order: random while the cluster is unmeasured,
// then any unmeasured node, then argmax of the adjusted speed.
//
// QValueAtDispatch reports the winning adjusted speed (higher is better,
// matching the convention of the other learners). WasExploration is true for
// both of icecream's exploration paths, so evaluation can split explore from
// exploit traffic exactly as it does for the bandits.
func (s *IceccFastestScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, _ TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := eligibleCandidates(s.registry, s.circuitChecker, buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}
	// Single-candidate fast path, matching the other learners so that
	// cluster-size effects do not differ between schedulers.
	if len(candidates) == 1 {
		return candidates[0], DispatchInfo{}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// pick_server_fastest, line 769: "If we have no statistics simply use
	// any server which is usable."
	if s.recorded == 0 {
		return candidates[cryptoRandInt(len(candidates))], DispatchInfo{WasExploration: true}, nil
	}

	// pick_server_new, line 753: nodes that have compiled nothing and are
	// currently idle are dispatched to first, "so we can get the stats we
	// need". Upstream additionally prefers such a node that already holds
	// the toolchain; Hybrid-Grid ships no toolchain, so that tie-break has
	// no analogue and is omitted.
	for _, w := range candidates {
		st := s.stats[w.ID]
		if (st == nil || len(st.window) == 0) && w.ActiveTasks == 0 {
			return w, DispatchInfo{WasExploration: true}, nil
		}
	}

	var best *registry.WorkerInfo
	bestSpeed := math.Inf(-1)
	for _, w := range candidates {
		if f := s.speedLocked(w); f > bestSpeed {
			best, bestSpeed = w, f
		}
	}
	if best == nil {
		return nil, DispatchInfo{}, ErrNoMatchingWorkers
	}
	return best, DispatchInfo{QValueAtDispatch: bestSpeed}, nil
}

// speedLocked is server_speed() with a job in hand: throughput learned from
// the rolling window, throttled by queue saturation, then inflated while the
// sample is small. Callers must hold s.mu.
func (s *IceccFastestScheduler) speedLocked(w *registry.WorkerInfo) float64 {
	st := s.stats[w.ID]
	// Upstream returns 0 for a node with no samples, which loses argmax
	// against any measured node. Reached only when the node is busy, since
	// an idle unmeasured node was already claimed by pick_server_new above.
	if st == nil || len(st.window) == 0 || st.cumTime == 0 {
		return 0
	}
	f := st.cumWork / st.cumTime

	maxParallel := w.MaxParallel
	if maxParallel <= 0 {
		maxParallel = 4
	}
	// "Gradually throttle with the number of assigned jobs."
	f *= 1.0 - 0.5*float64(w.ActiveTasks)/float64(maxParallel)

	if n := len(st.window); n < iceccPessimismSamples {
		f *= -0.5*float64(n) + 4.5
	}
	return f
}

// RecordOutcome implements LearningScheduler.
//
// The interface hands us the shaped reward r = -log(1 + t_ms) rather than the
// latency, but that transform is invertible, so t_ms = e^(-r) - 1 recovers the
// worker-reported compile time exactly and icecream's estimator can be fed the
// units it was designed for.
//
// Failed tasks are dropped rather than penalised, matching add_job_stats
// (line 152): "We don't want to base our timings on failed or too small
// jobs." This is a real behavioural difference from our own learners, which
// feed failures back as worst-case rewards — icecream leaves fault handling
// to a separate mechanism, as we leave it to circuit breakers.
func (s *IceccFastestScheduler) RecordOutcome(workerID string, reward float64, success bool, ctx TaskContext) {
	if workerID == "" || !success {
		return
	}
	work := float64(ctx.SourceSizeBytes)
	if work <= 0 {
		return
	}
	timeMs := math.Exp(-reward) - 1
	if timeMs <= 0 || math.IsInf(timeMs, 0) || math.IsNaN(timeMs) {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.stats[workerID]
	if !ok {
		st = &iceccNodeStats{}
		s.stats[workerID] = st
	}
	st.add(iceccJobStat{work: work, timeMs: timeMs})
	s.recorded++
}
