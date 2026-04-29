package scheduler

import (
	"math"
	"sync"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// HEFTScheduler is the online adaptation of the Heterogeneous Earliest
// Finish Time (HEFT) algorithm by Topcuoglu, Hariri, Wu (2002),
// "Performance-effective and low-complexity task scheduling for
// heterogeneous computing", IEEE TPDS 13(3):260–274
// (DOI: 10.1109/71.993206).
//
// **Adaptation note (cited in docs/thesis/theory-notes.md §6.2)**:
// the original HEFT operates on a fully-known DAG offline. For
// independent compilation tasks the upward-rank function reduces to
// rank_u(n) = w̄(n) (Topcuoglu 2002 Eq. 1, with succ(n) = ∅), and the
// processor-selection phase becomes a pure earliest-finish-time
// (EFT, Eq. 3) greedy assignment. We therefore implement only the
// EFT phase, computing
//
//	EST(task, p_j) = avail[j]                      (no predecessors → no comm. delay)
//	EFT(task, p_j) = w_{ij} + EST(task, p_j)
//
// where w_{ij} is a per-worker EWMA of observed compilation times and
// avail[j] = ActiveTasks(j) × w̄_j approximates the queue-clear time.
//
// Selection: argmin_{p_j} EFT(task, p_j). This is the canonical
// HEFT-EFT rule with the streaming relaxation made explicit. Any
// deviation from Topcuoglu 2002 is documented inline above the rule it
// affects so it survives code review.
//
// HEFTScheduler implements LearningScheduler so Compile() can feed it
// observed compile times. Reward sign is preserved (higher is better)
// but the relevant signal is the latency itself, which is also what
// drives EFT.
type HEFTScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker

	mu     sync.Mutex
	wbar   map[string]*metrics.EWMA // per-worker EWMA of compile times in ms
	alpha  float64                  // EWMA smoothing factor
	count  map[string]int64         // total observations per worker
}

// HEFTConfig holds construction parameters.
type HEFTConfig struct {
	Registry       registry.Registry
	CircuitChecker CircuitChecker
	// EWMAAlpha is the smoothing factor for per-worker compile time.
	// Default 0.3 matches the cost-scheduler design rationale: compile
	// time is more stable than RTT (which uses 0.5) but should still
	// adapt to gradual drift like thermal throttling.
	EWMAAlpha float64
}

// NewHEFTScheduler constructs the scheduler.
func NewHEFTScheduler(cfg HEFTConfig) *HEFTScheduler {
	alpha := cfg.EWMAAlpha
	if alpha <= 0 || alpha > 1 {
		alpha = 0.3
	}
	return &HEFTScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
		wbar:           make(map[string]*metrics.EWMA),
		alpha:          alpha,
		count:          make(map[string]int64),
	}
}

// Select implements the base Scheduler interface.
func (s *HEFTScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS, TaskContext{})
	return w, err
}

// SelectWithDispatchInfo implements LearningScheduler. The QValueAtDispatch
// reports the negative of the chosen worker's EFT (so larger Q = better,
// matching the convention used elsewhere). WasExploration is true only
// when no historical data is available for any candidate (cold start).
func (s *HEFTScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, _ TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := s.eligibleWorkers(buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}
	if len(candidates) == 1 {
		return candidates[0], DispatchInfo{}, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	bestEFT := -1.0
	var best *registry.WorkerInfo
	allCold := true
	for _, w := range candidates {
		wbar := s.wbarLocked(w.ID)
		if !wbar.IsInitialized() {
			// Cold worker: prior estimate from worker capability heuristic.
			// We adopt the cost-scheduler plan's hardware prior:
			// base 800 ms (median compile time observed in M1) scaled by
			// 1/sqrt(cores/4). This is a one-shot guess; the EWMA replaces
			// it after the first real observation.
			cores := 4.0
			if w.Capabilities != nil && w.Capabilities.CpuCores > 0 {
				cores = float64(w.Capabilities.CpuCores)
			}
			prior := 800.0 / math.Sqrt(cores/4.0)
			wbar.Update(prior)
		} else {
			allCold = false
		}
		w_ij := wbar.Value()
		avail := float64(w.ActiveTasks) * w_ij
		eft := avail + w_ij
		if best == nil || eft < bestEFT {
			best = w
			bestEFT = eft
		}
	}
	return best, DispatchInfo{QValueAtDispatch: -bestEFT, WasExploration: allCold}, nil
}

// RecordOutcome implements LearningScheduler. Failed tasks still update
// the EWMA: a worker that times out is genuinely slow for our purposes.
// The TaskContext is unused — HEFT does not condition on per-task features.
func (s *HEFTScheduler) RecordOutcome(workerID string, _ float64, _ bool, _ TaskContext) {
	if workerID == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	// We're given reward (negative log latency), not raw latency. The
	// scheduler instead uses the registry's AvgCompileTime which is
	// updated on DecrementTasks. Refresh our EWMA from registry state.
	if w, ok := s.registry.Get(workerID); ok && w.AvgCompileTime > 0 {
		ms := float64(w.AvgCompileTime.Milliseconds())
		s.wbarLocked(workerID).Update(ms)
		s.count[workerID]++
	}
}

func (s *HEFTScheduler) wbarLocked(workerID string) *metrics.EWMA {
	if e, ok := s.wbar[workerID]; ok {
		return e
	}
	e := metrics.NewEWMA(s.alpha)
	s.wbar[workerID] = e
	return e
}

// eligibleWorkers mirrors the admission rules of the other schedulers
// for apples-to-apples comparison.
func (s *HEFTScheduler) eligibleWorkers(buildType pb.BuildType, arch pb.Architecture, clientOS string) ([]*registry.WorkerInfo, error) {
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
	cands := make([]*registry.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		if w.State == registry.WorkerStateUnhealthy {
			continue
		}
		if s.circuitChecker != nil && s.circuitChecker.IsOpen(w.ID) {
			continue
		}
		maxP := w.MaxParallel
		if maxP <= 0 {
			maxP = 4
		}
		if w.ActiveTasks >= maxP {
			continue
		}
		cands = append(cands, w)
	}
	if len(cands) == 0 {
		for _, w := range workers {
			if w.State == registry.WorkerStateUnhealthy {
				continue
			}
			maxP := w.MaxParallel
			if maxP <= 0 {
				maxP = 4
			}
			if w.ActiveTasks < maxP {
				cands = append(cands, w)
			}
		}
	}
	if len(cands) == 0 {
		return nil, ErrNoMatchingWorkers
	}
	return cands, nil
}

