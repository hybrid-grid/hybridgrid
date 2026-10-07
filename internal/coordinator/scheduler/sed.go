package scheduler

import (
	"math/bits"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// defaultSEDCapacityMillis is the capacity assumed for a worker that reports
// neither a cgroup quota nor a core count (one core), so it still gets a
// finite, comparable score instead of dividing by zero.
const defaultSEDCapacityMillis = 1000

// SEDScheduler implements Shortest Expected Delay: it picks the worker with
// the smallest (ActiveTasks+1)/capacity, where capacity is the worker's CPU
// quota in milli-cores (cores*1000 when no quota is reported, one core when
// nothing is) and a negative ActiveTasks counts as zero. Compared with LeastLoaded it knows that a worker with
// twice the capacity drains its queue twice as fast, but it still never
// learns from observed outcomes — it is the strongest non-learning heuristic
// that uses the same capability information as the bandit.
//
// Capacity is whatever the worker registered with (the cgroup quota); SED
// never refreshes it. A worker that slows down without changing its quota
// (CPU contention inside the container) is invisible to SED by design.
//
// SEDScheduler implements Scheduler only, not LearningScheduler: SelectWith
// treats it like LeastLoaded and P2C.
type SEDScheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker
}

// SEDConfig holds construction parameters.
type SEDConfig struct {
	Registry registry.Registry
	// CircuitChecker is optional; nil disables circuit-breaker filtering.
	CircuitChecker CircuitChecker
}

// NewSEDScheduler creates a SED scheduler.
func NewSEDScheduler(cfg SEDConfig) *SEDScheduler {
	return &SEDScheduler{
		registry:       cfg.Registry,
		circuitChecker: cfg.CircuitChecker,
	}
}

// Select implements Scheduler. Candidates come from eligibleCandidates, the
// admission rule shared by every scheduler under evaluation, so differences
// in makespan come from the choice rule alone.
//
// Ties are broken deterministically (the registry iterates a map, so
// candidate order is random): larger capacity first, then fewer active tasks,
// then the lexicographically smaller worker ID. The ActiveTasks level can
// only decide when a count is negative (clamped to zero for scoring, so two
// workers can tie while their raw counts differ); for valid non-negative
// counts equal score and equal capacity already imply equal ActiveTasks.
func (s *SEDScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	cands, err := eligibleCandidates(s.registry, s.circuitChecker, buildType, arch, clientOS)
	if err != nil {
		return nil, err
	}

	best := cands[0]
	bestCap := sedCapacity(best)
	for _, w := range cands[1:] {
		c := sedCapacity(w)
		if sedBetter(w, c, best, bestCap) {
			best, bestCap = w, c
		}
	}
	return best, nil
}

// sedBetter reports whether worker a (capacity ac) should be preferred over
// worker b (capacity bc).
func sedBetter(a *registry.WorkerInfo, ac int64, b *registry.WorkerInfo, bc int64) bool {
	aa, ba := sedActive(a), sedActive(b)
	if sedLess(aa, ac, ba, bc) {
		return true
	}
	if sedLess(ba, bc, aa, ac) {
		return false
	}
	if ac != bc {
		return ac > bc
	}
	if a.ActiveTasks != b.ActiveTasks {
		return a.ActiveTasks < b.ActiveTasks
	}
	return a.ID < b.ID
}

// sedActive returns the worker's active task count. The registry never
// produces a negative count itself but Add accepts one, so it is clamped to
// zero rather than wrapping when converted to unsigned in sedLess.
func sedActive(w *registry.WorkerInfo) int64 { return max(int64(w.ActiveTasks), 0) }

// sedCapacity returns the worker's CPU capacity in milli-cores: the cgroup
// quota when detected, otherwise cores*1000 (via effectiveCPUMillis), and
// defaultSEDCapacityMillis when the worker reports nothing usable.
func sedCapacity(w *registry.WorkerInfo) int64 {
	if c := int64(effectiveCPUMillis(w.Capabilities)); c > 0 {
		return c
	}
	return defaultSEDCapacityMillis
}

// sedLess reports whether (a1+1)/c1 < (a2+1)/c2 for non-negative a and
// positive c, by comparing the exact cross-products (a1+1)*c2 < (a2+1)*c1.
// Exact ties occur on the real cluster (1/500 == 2/1000) and floating-point
// division can order near-ties wrongly; the products can also exceed int64
// for legal int32 inputs, so they are computed in 128 bits.
func sedLess(a1, c1, a2, c2 int64) bool {
	hi1, lo1 := bits.Mul64(uint64(a1+1), uint64(c2))
	hi2, lo2 := bits.Mul64(uint64(a2+1), uint64(c1))
	return hi1 < hi2 || (hi1 == hi2 && lo1 < lo2)
}
