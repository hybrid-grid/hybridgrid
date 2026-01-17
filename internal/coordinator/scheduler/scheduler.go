package scheduler

import (
	"crypto/rand"
	"errors"
	"math/big"
	"sync/atomic"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

var (
	// ErrNoWorkers is returned when no workers are available.
	ErrNoWorkers = errors.New("no workers available")

	// ErrNoMatchingWorkers is returned when no workers match the requirements.
	ErrNoMatchingWorkers = errors.New("no workers match requirements")
)

// Scheduler selects workers for build tasks.
type Scheduler interface {
	// Select chooses a worker for the given build type and architecture.
	Select(buildType pb.BuildType, arch pb.Architecture) (*registry.WorkerInfo, error)
}

// SimpleScheduler implements round-robin worker selection.
type SimpleScheduler struct {
	registry registry.Registry
	counter  uint64
}

// NewSimpleScheduler creates a new simple scheduler.
func NewSimpleScheduler(reg registry.Registry) *SimpleScheduler {
	return &SimpleScheduler{
		registry: reg,
	}
}

// Select chooses the next available worker using round-robin.
func (s *SimpleScheduler) Select(buildType pb.BuildType, arch pb.Architecture) (*registry.WorkerInfo, error) {
	workers := s.registry.ListByCapability(buildType, arch)
	if len(workers) == 0 {
		// Check if there are any workers at all
		if s.registry.Count() == 0 {
			return nil, ErrNoWorkers
		}
		return nil, ErrNoMatchingWorkers
	}

	// Filter out busy workers with too many tasks (simple load awareness)
	available := make([]*registry.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		// Allow up to 2 concurrent tasks per worker
		if w.ActiveTasks < 2 && w.State == registry.WorkerStateIdle {
			available = append(available, w)
		}
	}

	// If all workers are busy, use any healthy worker
	if len(available) == 0 {
		for _, w := range workers {
			if w.State != registry.WorkerStateUnhealthy {
				available = append(available, w)
			}
		}
	}

	if len(available) == 0 {
		return nil, ErrNoMatchingWorkers
	}

	// Round-robin selection
	idx := atomic.AddUint64(&s.counter, 1)
	selected := available[int(idx)%len(available)]

	return selected, nil
}

// LeastLoadedScheduler selects the worker with the fewest active tasks.
type LeastLoadedScheduler struct {
	registry registry.Registry
}

// NewLeastLoadedScheduler creates a new least-loaded scheduler.
func NewLeastLoadedScheduler(reg registry.Registry) *LeastLoadedScheduler {
	return &LeastLoadedScheduler{
		registry: reg,
	}
}

// Select chooses the worker with the least load.
func (s *LeastLoadedScheduler) Select(buildType pb.BuildType, arch pb.Architecture) (*registry.WorkerInfo, error) {
	workers := s.registry.ListByCapability(buildType, arch)
	if len(workers) == 0 {
		if s.registry.Count() == 0 {
			return nil, ErrNoWorkers
		}
		return nil, ErrNoMatchingWorkers
	}

	var best *registry.WorkerInfo
	for _, w := range workers {
		if w.State == registry.WorkerStateUnhealthy {
			continue
		}

		if best == nil || w.ActiveTasks < best.ActiveTasks {
			best = w
		}
	}

	if best == nil {
		return nil, ErrNoMatchingWorkers
	}

	return best, nil
}

// Scoring weights per DOCUMENTATION.md 4.2.4
const (
	ScoreNativeArchMatch = 50.0  // +50 if native arch matches target
	ScoreCrossCompile    = 25.0  // +25 if can cross-compile
	ScorePerCPUCore      = 10.0  // +10 per CPU core
	ScorePerGBMemory     = 5.0   // +5 per GB RAM
	ScorePerActiveTask   = -15.0 // -15 per active task
	ScorePerMsLatency    = -0.5  // -0.5 per ms latency
	ScoreLANSource       = 20.0  // +20 if LAN discovery
	ScoreMaxActiveTasks  = 8     // Workers above this are deprioritized
)

// P2CScheduler implements Power of Two Choices scheduling with weighted scoring.
type P2CScheduler struct {
	registry       registry.Registry
	latencyTracker *metrics.LatencyTracker
	circuitChecker CircuitChecker
}

// CircuitChecker checks if a worker's circuit breaker is open.
type CircuitChecker interface {
	IsOpen(workerID string) bool
}

// P2CConfig holds P2C scheduler configuration.
type P2CConfig struct {
	Registry       registry.Registry
	LatencyTracker *metrics.LatencyTracker
	CircuitChecker CircuitChecker
}

// NewP2CScheduler creates a new P2C scheduler.
func NewP2CScheduler(cfg P2CConfig) *P2CScheduler {
	lt := cfg.LatencyTracker
	if lt == nil {
		lt = metrics.NewLatencyTracker()
	}
	return &P2CScheduler{
		registry:       cfg.Registry,
		latencyTracker: lt,
		circuitChecker: cfg.CircuitChecker,
	}
}

// Select implements P2C: pick 2 random workers, select the one with higher score.
func (s *P2CScheduler) Select(buildType pb.BuildType, arch pb.Architecture) (*registry.WorkerInfo, error) {
	workers := s.registry.ListByCapability(buildType, arch)
	if len(workers) == 0 {
		if s.registry.Count() == 0 {
			return nil, ErrNoWorkers
		}
		return nil, ErrNoMatchingWorkers
	}

	// Filter out unhealthy workers and those with open circuits
	candidates := make([]*registry.WorkerInfo, 0, len(workers))
	for _, w := range workers {
		if w.State == registry.WorkerStateUnhealthy {
			continue
		}
		if s.circuitChecker != nil && s.circuitChecker.IsOpen(w.ID) {
			continue
		}
		// Allow up to ScoreMaxActiveTasks concurrent tasks
		if w.ActiveTasks >= ScoreMaxActiveTasks {
			continue
		}
		candidates = append(candidates, w)
	}

	// If no candidates after filtering, relax criteria
	if len(candidates) == 0 {
		for _, w := range workers {
			if w.State != registry.WorkerStateUnhealthy {
				candidates = append(candidates, w)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, ErrNoMatchingWorkers
	}

	// If only 1 candidate, return it
	if len(candidates) == 1 {
		return candidates[0], nil
	}

	// Pick 2 random workers
	idx1, idx2 := pickTwo(len(candidates))
	w1, w2 := candidates[idx1], candidates[idx2]

	// Calculate scores
	score1 := s.scoreWorker(w1, arch)
	score2 := s.scoreWorker(w2, arch)

	// Return the worker with higher score
	if score1 >= score2 {
		return w1, nil
	}
	return w2, nil
}

// scoreWorker calculates a worker's score for task assignment.
func (s *P2CScheduler) scoreWorker(w *registry.WorkerInfo, targetArch pb.Architecture) float64 {
	var score float64

	if w.Capabilities == nil {
		return 0
	}

	caps := w.Capabilities

	// Architecture matching
	if targetArch == pb.Architecture_ARCH_UNSPECIFIED ||
		caps.NativeArch == targetArch {
		score += ScoreNativeArchMatch
	} else if caps.DockerAvailable {
		// Can cross-compile via Docker
		score += ScoreCrossCompile
	}

	// CPU cores (normalized, max 16 cores contribute)
	cpuContrib := float64(caps.CpuCores)
	if cpuContrib > 16 {
		cpuContrib = 16
	}
	score += cpuContrib * ScorePerCPUCore

	// Memory (in GB, max 64GB contributes)
	memGB := float64(caps.MemoryBytes) / (1024 * 1024 * 1024)
	if memGB > 64 {
		memGB = 64
	}
	score += memGB * ScorePerGBMemory

	// Active tasks penalty
	score += float64(w.ActiveTasks) * ScorePerActiveTask

	// Latency penalty
	latencyMs := s.latencyTracker.Get(w.ID)
	score += latencyMs * ScorePerMsLatency

	// LAN source bonus (check discovery source if available)
	if w.DiscoverySource == "mdns" || w.DiscoverySource == "LAN" {
		score += ScoreLANSource
	}

	return score
}

// ReportSuccess records successful task completion latency.
func (s *P2CScheduler) ReportSuccess(workerID string, latencyMs float64) {
	s.latencyTracker.Record(workerID, latencyMs)
}

// ReportFailure can be used to track failures (for future use).
func (s *P2CScheduler) ReportFailure(workerID string, err error) {
	// Circuit breaker handles failures, but we could track metrics here
}

// pickTwo returns two different random indices from [0, n).
func pickTwo(n int) (int, int) {
	if n < 2 {
		return 0, 0
	}

	idx1 := cryptoRandInt(n)
	idx2 := cryptoRandInt(n - 1)
	if idx2 >= idx1 {
		idx2++ // Ensure different indices
	}
	return idx1, idx2
}

// cryptoRandInt returns a random int in [0, n) using crypto/rand.
func cryptoRandInt(n int) int {
	if n <= 0 {
		return 0
	}
	big, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(big.Int64())
}
