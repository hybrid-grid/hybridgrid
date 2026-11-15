package scheduler

import (
	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// P90Scheduler sends large tasks to the strongest free worker and delegates
// smaller tasks to LeastLoaded.
type P90Scheduler struct {
	registry  registry.Registry
	threshold int
	fallback  *LeastLoadedScheduler
}

func NewP90Scheduler(reg registry.Registry, threshold int) *P90Scheduler {
	return &P90Scheduler{registry: reg, threshold: threshold, fallback: NewLeastLoadedScheduler(reg)}
}

func (s *P90Scheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	return s.fallback.Select(buildType, arch, clientOS)
}

func (s *P90Scheduler) SelectWithContext(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, error) {
	if ctx.SourceSizeBytes > s.threshold {
		if w := strongestFree(s.registry, nil, buildType, arch, clientOS); w != nil {
			return w, nil
		}
	}
	return s.fallback.Select(buildType, arch, clientOS)
}

// SEDP90Scheduler sends large tasks to the strongest free worker and delegates
// smaller tasks to SED.
type SEDP90Scheduler struct {
	registry       registry.Registry
	circuitChecker CircuitChecker
	threshold      int
	fallback       *SEDScheduler
}

func NewSEDP90Scheduler(cfg SEDConfig, threshold int) *SEDP90Scheduler {
	return &SEDP90Scheduler{
		registry: cfg.Registry, circuitChecker: cfg.CircuitChecker,
		threshold: threshold, fallback: NewSEDScheduler(cfg),
	}
}

func (s *SEDP90Scheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	return s.fallback.Select(buildType, arch, clientOS)
}

func (s *SEDP90Scheduler) SelectWithContext(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx TaskContext) (*registry.WorkerInfo, error) {
	if ctx.SourceSizeBytes > s.threshold {
		if w := strongestFree(s.registry, s.circuitChecker, buildType, arch, clientOS); w != nil {
			return w, nil
		}
	}
	return s.fallback.Select(buildType, arch, clientOS)
}

// strongestFree uses the same capability, OS, health and capacity admission
// rules as SED. Strict comparison preserves registry order for CPU ties.
func strongestFree(reg registry.Registry, cc CircuitChecker, buildType pb.BuildType, arch pb.Architecture, clientOS string) *registry.WorkerInfo {
	candidates, err := eligibleCandidates(reg, cc, buildType, arch, clientOS)
	if err != nil {
		return nil
	}
	best := candidates[0]
	for _, w := range candidates[1:] {
		if w.Capabilities.CpuMillis > best.Capabilities.CpuMillis {
			best = w
		}
	}
	return best
}
