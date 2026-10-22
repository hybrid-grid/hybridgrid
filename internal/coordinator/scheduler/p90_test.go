package scheduler

import (
	"errors"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// orderedP90Registry fixes candidate order so CPU tie behavior is observable.
type orderedP90Registry struct {
	registry.Registry
	workers []*registry.WorkerInfo
}

func (r *orderedP90Registry) ListByCapability(_ pb.BuildType, _ pb.Architecture) []*registry.WorkerInfo {
	return r.workers
}

func (r *orderedP90Registry) Count() int { return len(r.workers) }

func p90Worker(id string, cpu, active, parallel int32) *registry.WorkerInfo {
	return &registry.WorkerInfo{
		ID: id, ActiveTasks: active, MaxParallel: parallel,
		Capabilities: &pb.WorkerCapabilities{CpuMillis: cpu, NativeArch: pb.Architecture_ARCH_X86_64, Os: "linux"},
	}
}

func TestP90Scheduler_Selection(t *testing.T) {
	const threshold = 100
	buildType, arch := pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64
	tests := []struct {
		name    string
		workers []*registry.WorkerInfo
		size    int
		wantLL  string
		wantSED string
		wantErr error
	}{
		{"large picks strongest", []*registry.WorkerInfo{p90Worker("small", 500, 0, 2), p90Worker("strong", 1100, 1, 2)}, 101, "strong", "strong", nil},
		{"strongest full", []*registry.WorkerInfo{p90Worker("small", 500, 0, 2), p90Worker("mid", 800, 1, 2), p90Worker("strong", 1100, 2, 2)}, 101, "mid", "mid", nil},
		{"cpu tie keeps registry order", []*registry.WorkerInfo{p90Worker("first", 1000, 1, 2), p90Worker("second", 1000, 0, 2)}, 101, "first", "first", nil},
		{"unhealthy excluded", []*registry.WorkerInfo{p90Worker("bad", 2000, 0, 2), p90Worker("good", 500, 0, 2)}, 101, "good", "good", nil},
		{"small delegates", []*registry.WorkerInfo{p90Worker("least", 500, 0, 2), p90Worker("strong", 1100, 1, 2)}, 99, "least", "strong", nil},
		{"boundary delegates", []*registry.WorkerInfo{p90Worker("least", 500, 0, 2), p90Worker("strong", 1100, 1, 2)}, 100, "least", "strong", nil},
		{"small skips full", []*registry.WorkerInfo{p90Worker("full", 500, 1, 1), p90Worker("free", 1100, 1, 2)}, 99, "free", "free", nil},
		{"large all full falls back", []*registry.WorkerInfo{p90Worker("a", 500, 1, 1), p90Worker("b", 1100, 2, 2)}, 101, "", "", ErrNoMatchingWorkers},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "unhealthy excluded" {
				tt.workers[0].State = registry.WorkerStateUnhealthy
			}
			reg := &orderedP90Registry{workers: tt.workers}
			arms := []struct {
				name string
				s    Scheduler
				want string
			}{
				{"p90", NewP90Scheduler(reg, threshold), tt.wantLL},
				{"sed-p90", NewSEDP90Scheduler(SEDConfig{Registry: reg}, threshold), tt.wantSED},
			}
			for _, arm := range arms {
				t.Run(arm.name, func(t *testing.T) {
					if _, ok := arm.s.(LearningScheduler); ok {
						t.Fatal("baseline must not receive learning feedback")
					}
					w, info, err := SelectWith(arm.s, buildType, arch, "", TaskContext{SourceSizeBytes: tt.size})
					if !errors.Is(err, tt.wantErr) {
						t.Fatalf("error = %v, want %v", err, tt.wantErr)
					}
					if info != (DispatchInfo{}) {
						t.Fatalf("non-learning dispatch info = %+v", info)
					}
					if tt.wantErr == nil && (w == nil || w.ID != arm.want) {
						t.Fatalf("worker = %v, want %s", w, arm.want)
					}
				})
			}
		})
	}
}

func TestP90Scheduler_SmallMatchesBase(t *testing.T) {
	workers := [][]*registry.WorkerInfo{
		{p90Worker("a", 500, 0, 1), p90Worker("b", 1000, 1, 2)},
		{p90Worker("a", 500, 1, 2), p90Worker("b", 1000, 0, 2)},
		{p90Worker("a", 500, 1, 1), p90Worker("b", 1000, 1, 2)},
	}
	for _, state := range workers {
		reg := &orderedP90Registry{workers: state}
		pairs := []struct{ baseline, p90 Scheduler }{
			{NewLeastLoadedScheduler(reg), NewP90Scheduler(reg, 100)},
			{NewSEDScheduler(SEDConfig{Registry: reg}), NewSEDP90Scheduler(SEDConfig{Registry: reg}, 100)},
		}
		for _, pair := range pairs {
			want, err := pair.baseline.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
			if err != nil {
				t.Fatal(err)
			}
			got, _, err := SelectWith(pair.p90, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{SourceSizeBytes: 100})
			if err != nil || got.ID != want.ID {
				t.Fatalf("small task: got %v, %v; baseline %v", got, err, want)
			}
		}
	}
}

func TestP90Scheduler_CapabilityAndOS(t *testing.T) {
	reg := registry.NewInMemoryRegistry(time.Minute)
	defer reg.Stop()
	add := func(id string, cpu int32, arch pb.Architecture, osName string) {
		t.Helper()
		if err := reg.Add(&registry.WorkerInfo{
			ID: id, MaxParallel: 1,
			Capabilities: &pb.WorkerCapabilities{
				CpuMillis: cpu, NativeArch: arch, Os: osName,
				Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	add("wrong-arch", 3000, pb.Architecture_ARCH_ARM64, "linux")
	add("wrong-os", 2000, pb.Architecture_ARCH_X86_64, "windows")
	add("eligible", 500, pb.Architecture_ARCH_X86_64, "linux")
	for _, s := range []Scheduler{
		NewP90Scheduler(reg, 100),
		NewSEDP90Scheduler(SEDConfig{Registry: reg}, 100),
	} {
		w, _, err := SelectWith(s, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "linux", TaskContext{SourceSizeBytes: 101})
		if err != nil || w.ID != "eligible" {
			t.Fatalf("eligible worker = %v, %v", w, err)
		}
	}
}

type selectionSpy struct {
	selectCalls int
	worker      *registry.WorkerInfo
}

func (s *selectionSpy) Select(pb.BuildType, pb.Architecture, string) (*registry.WorkerInfo, error) {
	s.selectCalls++
	return s.worker, nil
}

type learnerSpy struct {
	selectionSpy
	contextCalls int
	context      TaskContext
}

func (s *learnerSpy) SelectWithDispatchInfo(_ pb.BuildType, _ pb.Architecture, _ string, ctx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	s.contextCalls++
	s.context = ctx
	return s.worker, DispatchInfo{QValueAtDispatch: 3}, nil
}

func (s *learnerSpy) RecordOutcome(string, float64, bool, TaskContext) {}

func TestSelectWith_PreservesExistingRouting(t *testing.T) {
	w := p90Worker("worker", 500, 0, 1)
	ctx := TaskContext{SourceSizeBytes: 101}
	plain := &selectionSpy{worker: w}
	got, info, err := SelectWith(plain, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
	if err != nil || got != w || plain.selectCalls != 1 || info != (DispatchInfo{}) {
		t.Fatalf("plain path: worker %v, info %+v, calls %d, err %v", got, info, plain.selectCalls, err)
	}
	learner := &learnerSpy{selectionSpy: selectionSpy{worker: w}}
	got, info, err = SelectWith(learner, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
	if err != nil || got != w || learner.selectCalls != 0 || learner.contextCalls != 1 || learner.context != ctx || info.QValueAtDispatch != 3 {
		t.Fatalf("learner path: worker %v, info %+v, calls %d/%d, err %v", got, info, learner.selectCalls, learner.contextCalls, err)
	}
}
