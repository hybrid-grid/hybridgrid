package scheduler

import (
	"math"
	"strconv"
	"sync"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// ewmaAlpha is the registered default; other values are used only for the fairness sweep.
const ewmaAlpha = 0.2

// The coordinator normalizes rewards against its fixed two-minute scale.
const ewmaRewardNormalizationMs = 120000.0

type ewmaObservation struct {
	workerID string
	size     int
}

// EWMAScheduler keeps a log-domain EWMA of milliseconds per source byte for
// each worker and selects the smallest log(active tasks + 1) plus estimate.
type EWMAScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker
	warmStartTasks int
	alpha          float64

	mu         sync.Mutex
	dispatches int
	est        map[string]float64
	n          map[string]int
	pending    map[string]ewmaObservation
}

type EWMAConfig struct {
	Registry       registry.Registry
	CircuitChecker CircuitChecker
	WarmStartTasks int
	Alpha          float64
}

func NewEWMAScheduler(cfg EWMAConfig) *EWMAScheduler {
	alpha := cfg.Alpha
	if !(alpha > 0 && alpha <= 1) {
		alpha = ewmaAlpha
	}
	return &EWMAScheduler{
		registry: cfg.Registry, circuitChecker: cfg.CircuitChecker,
		warmStartTasks: max(cfg.WarmStartTasks, 0),
		alpha:          alpha,
		est:            make(map[string]float64), n: make(map[string]int),
		pending: make(map[string]ewmaObservation),
	}
}

// Params reports the effective scheduler parameters used for task logs.
func (s *EWMAScheduler) Params() string {
	return "alpha=" + strconv.FormatFloat(s.alpha, 'g', -1, 64) +
		" warm_start=" + strconv.Itoa(s.warmStartTasks)
}

func (s *EWMAScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS, TaskContext{})
	return w, err
}

func (s *EWMAScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	candidates, err := eligibleCandidates(s.registry, s.circuitChecker, buildType, arch, clientOS)
	if err != nil {
		return nil, DispatchInfo{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.dispatches++
	warm := s.dispatches <= s.warmStartTasks
	chosen := candidates[0]
	exploration := warm
	if warm {
		for _, w := range candidates[1:] {
			if w.ActiveTasks < chosen.ActiveTasks {
				chosen = w
			}
		}
	} else {
		// First unobserved eligible worker wins; candidate order breaks ties.
		unobserved := false
		for _, w := range candidates {
			if s.n[w.ID] == 0 {
				chosen, unobserved = w, true
				break
			}
		}
		if unobserved {
			exploration = true
		} else {
			for _, w := range candidates[1:] {
				if s.score(w) < s.score(chosen) {
					chosen = w
				}
			}
		}
	}
	if ctx.TaskID != "" {
		// Reselection overwrites the previous dispatch's size and worker.
		s.pending[ctx.TaskID] = ewmaObservation{workerID: chosen.ID, size: max(ctx.SourceSizeBytes, 1)}
	}
	return chosen, DispatchInfo{QValueAtDispatch: -s.score(chosen), WasExploration: exploration}, nil
}

func (s *EWMAScheduler) score(w *registry.WorkerInfo) float64 {
	return math.Log(float64(sedActive(w)+1)) + s.est[w.ID]
}

// ewmaRewardToDurationMs exactly inverts the coordinator's fixed reward
// transform for both successful and failed compilations.
func ewmaRewardToDurationMs(reward float64) float64 {
	return math.Expm1(-reward * math.Log1p(ewmaRewardNormalizationMs))
}

func (s *EWMAScheduler) RecordOutcome(workerID string, reward float64, _ bool, ctx TaskContext) {
	s.mu.Lock()
	defer s.mu.Unlock()
	obs, ok := s.pending[ctx.TaskID]
	delete(s.pending, ctx.TaskID)
	if !ok || obs.workerID != workerID || math.IsNaN(reward) || math.IsInf(reward, 0) {
		return
	}
	durationMs := ewmaRewardToDurationMs(reward)
	if math.IsNaN(durationMs) || math.IsInf(durationMs, 0) || durationMs < 0 {
		return
	}
	y := math.Log(math.Max(durationMs, 1) / float64(obs.size))
	if s.n[workerID] == 0 {
		s.est[workerID] = y
	} else {
		s.est[workerID] = (1-s.alpha)*s.est[workerID] + s.alpha*y
	}
	s.n[workerID]++
}

var _ LearningScheduler = (*EWMAScheduler)(nil)
