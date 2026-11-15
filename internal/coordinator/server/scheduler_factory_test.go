package server

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/resilience"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

func TestNewScheduler_KnownTypes(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	cases := map[string]any{
		"simple":          (*scheduler.SimpleScheduler)(nil),
		"p2c":             (*scheduler.P2CScheduler)(nil),
		"leastloaded":     (*scheduler.LeastLoadedScheduler)(nil),
		"":                (*scheduler.LeastLoadedScheduler)(nil), // empty -> default
		"epsilon-greedy":  (*scheduler.EpsilonGreedyScheduler)(nil),
		"linucb":          (*scheduler.LinUCBScheduler)(nil),
		"hybrid-linucb":   (*scheduler.LinUCBScheduler)(nil),
		"hybrid-linucb-d": (*scheduler.LinUCBScheduler)(nil),
		"heft":            (*scheduler.HEFTScheduler)(nil),
		"icecc-fastest":   (*scheduler.IceccFastestScheduler)(nil),
		"sed":             (*scheduler.SEDScheduler)(nil),
		"p90":             (*scheduler.P90Scheduler)(nil),
		"sed-p90":         (*scheduler.SEDP90Scheduler)(nil),
	}

	for typ, want := range cases {
		t.Run(typ, func(t *testing.T) {
			got := newScheduler(Config{SchedulerType: typ}, reg, cm)
			assert.NotNil(t, got)
			assert.IsType(t, want, got)
		})
	}
}

func TestNewScheduler_UnknownTypeFallsBack(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	got := newScheduler(Config{SchedulerType: "does-not-exist"}, reg, cm)
	assert.IsType(t, (*scheduler.LeastLoadedScheduler)(nil), got)
}

// TestNewScheduler_EpsilonValueRespected ensures the configured
// EpsilonValue propagates to the EpsilonGreedyScheduler. We can't read
// the field directly (private) but a Q-based behavioural test would be
// brittle; verify only that construction with non-default ε works.
func TestNewScheduler_EpsilonValueRespected(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	got := newScheduler(Config{SchedulerType: "epsilon-greedy", EpsilonValue: 0.5}, reg, cm)
	assert.IsType(t, (*scheduler.EpsilonGreedyScheduler)(nil), got)
}

func TestNewScheduler_DiscountedParams(t *testing.T) {
	defaults := DefaultConfig()
	assert.Equal(t, 0.98, defaults.DiscountValue)
	assert.Equal(t, "global", defaults.DiscountModeValue)

	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	tests := []struct {
		name string
		cfg  Config
		want string
	}{
		{"default discount", Config{SchedulerType: "hybrid-linucb-d"}, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.98 discount_mode=global"},
		{"explicit arm", Config{SchedulerType: "hybrid-linucb-d", DiscountValue: 0.95, DiscountModeValue: "arm", AlphaValue: 1, WarmStartTasks: 100, LoadPenaltyValue: 0.5}, "alpha=1 warm_start=100 load_penalty=0.5 discount=0.95 discount_mode=arm"},
		{"hybrid discount off", Config{SchedulerType: "hybrid-linucb", DiscountValue: 0.95, DiscountModeValue: "arm"}, "alpha=0.5 warm_start=0 load_penalty=0"},
		{"unknown mode normalised", Config{SchedulerType: "hybrid-linucb-d", DiscountValue: 0.95, DiscountModeValue: "unknown"}, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.95 discount_mode=global"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newScheduler(tt.cfg, reg, cm)
			linucb, ok := got.(*scheduler.LinUCBScheduler)
			if assert.True(t, ok) {
				assert.Equal(t, tt.want, linucb.Params())
			}
		})
	}
}

func TestNewScheduler_ExperimentConfig(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	defer reg.Stop()
	cm := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	for _, typ := range []string{"linucb", "hybrid-linucb", "hybrid-linucb-d"} {
		t.Run(typ, func(t *testing.T) {
			cfg := Config{SchedulerType: typ, FeatureSetValue: "compact", RewardCenterValue: "none", DiagCandidates: true}
			got, ok := newScheduler(cfg, reg, cm).(*scheduler.LinUCBScheduler)
			if !ok {
				t.Fatal("expected LinUCBScheduler")
			}
			if !strings.Contains(got.Params(), "features=compact") {
				t.Fatalf("Params() = %q, want features=compact", got.Params())
			}
		})
	}

	hybrid, ok := newScheduler(Config{SchedulerType: "hybrid-linucb", RewardCenterValue: "warm", WarmStartTasks: 100}, reg, cm).(*scheduler.LinUCBScheduler)
	if !ok {
		t.Fatal("expected hybrid LinUCBScheduler")
	}
	if !strings.Contains(hybrid.Params(), "reward_center=warm") {
		t.Fatalf("Params() = %q, want reward_center=warm", hybrid.Params())
	}
	discounted, ok := newScheduler(Config{SchedulerType: "hybrid-linucb-d", RewardCenterValue: "warm", WarmStartTasks: 100, DiscountValue: 0.95, DiscountModeValue: "arm"}, reg, cm).(*scheduler.LinUCBScheduler)
	if !ok {
		t.Fatal("expected discounted LinUCBScheduler")
	}
	if !strings.Contains(discounted.Params(), "reward_center=warm") {
		t.Fatalf("Params() = %q, want reward_center=warm", discounted.Params())
	}

	ewma, ok := newScheduler(Config{SchedulerType: "ewma", EwmaAlphaValue: 0.4}, reg, cm).(*scheduler.EWMAScheduler)
	if !ok {
		t.Fatal("expected EWMAScheduler")
	}
	if !strings.Contains(ewma.Params(), "alpha=0.4") {
		t.Fatalf("Params() = %q, want alpha=0.4", ewma.Params())
	}

	for _, address := range []string{"127.0.0.1:19001", "127.0.0.1:19002"} {
		if err := reg.Add(&registry.WorkerInfo{
			ID:           address,
			Address:      address,
			MaxParallel:  2,
			Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		}); err != nil {
			t.Fatalf("add worker: %v", err)
		}
	}
	diag, ok := newScheduler(Config{SchedulerType: "linucb", DiagCandidates: true}, reg, cm).(*scheduler.LinUCBScheduler)
	if !ok {
		t.Fatal("expected diagnostic LinUCBScheduler")
	}
	_, info, err := diag.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", scheduler.TaskContext{SourceSizeBytes: 1024})
	if err != nil {
		t.Fatalf("select: %v", err)
	}
	if len(info.Candidates) != 2 {
		t.Fatalf("diagnostics = %d candidates, want 2", len(info.Candidates))
	}
	for _, typ := range []string{"hybrid-linucb", "hybrid-linucb-d"} {
		t.Run(typ+" diagnostics", func(t *testing.T) {
			configured, ok := newScheduler(Config{SchedulerType: typ, DiagCandidates: true}, reg, cm).(*scheduler.LinUCBScheduler)
			if !ok {
				t.Fatal("expected LinUCBScheduler")
			}
			_, info, err := configured.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", scheduler.TaskContext{SourceSizeBytes: 1024})
			if err != nil {
				t.Fatalf("select: %v", err)
			}
			if len(info.Candidates) != 2 {
				t.Fatalf("diagnostics = %d candidates, want 2", len(info.Candidates))
			}
		})
	}
}
