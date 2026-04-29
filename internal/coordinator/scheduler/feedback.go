package scheduler

import (
	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// DispatchInfo carries learner-internal state observed at the moment the
// scheduler made a selection. It is logged into TaskLogRecord and never
// affects scheduling. The fields are populated only by learning
// schedulers; non-learning schedulers report zero values.
type DispatchInfo struct {
	// QValueAtDispatch is the learner's current estimate for the chosen
	// worker. For ε-greedy this is the running mean reward; for LinUCB
	// it is θᵀx + α√(xᵀA⁻¹x). Zero when no learner is configured.
	QValueAtDispatch float64
	// WasExploration is true when the learner picked a non-greedy action
	// (uniform random for ε-greedy). Lets evaluation split traffic into
	// explore vs exploit windows.
	WasExploration bool
}

// LearningScheduler extends Scheduler with an online-learning loop.
//
// A scheduler implementing this interface receives reward observations
// after each task completes and can expose introspection at dispatch
// time. Existing non-learning schedulers (Simple, LeastLoaded, P2C)
// remain unchanged; the gRPC layer uses a type assertion to detect a
// learner and feed it back.
type LearningScheduler interface {
	Scheduler

	// SelectWithDispatchInfo selects a worker and returns the learner's
	// internal state at decision time. Implementations must satisfy:
	// when err == nil, worker != nil and DispatchInfo is fully populated.
	SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, DispatchInfo, error)

	// RecordOutcome updates the learner's estimates from an observed
	// task outcome. The reward sign convention is "higher is better"
	// (typically negative latency). The success flag is provided in
	// case implementations want to discount or skip failed tasks.
	RecordOutcome(workerID string, reward float64, success bool)
}

// SelectWith dispatches to LearningScheduler when available, falling
// back to the base Scheduler.Select for non-learning schedulers. The
// returned DispatchInfo is zero-valued in the fallback case.
func SelectWith(s Scheduler, buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, DispatchInfo, error) {
	if learner, ok := s.(LearningScheduler); ok {
		return learner.SelectWithDispatchInfo(buildType, arch, clientOS)
	}
	w, err := s.Select(buildType, arch, clientOS)
	return w, DispatchInfo{}, err
}
