package scheduler

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// addSEDWorker registers a C++ worker with an explicit cgroup quota
// (milli-cores) and parallelism cap.
func addSEDWorker(t *testing.T, reg *registry.InMemoryRegistry, id string, millis int32, maxParallel int32) {
	t.Helper()
	err := reg.Add(&registry.WorkerInfo{
		ID:          id,
		Address:     id + ":50051",
		MaxParallel: maxParallel,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			CpuMillis:  millis,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
	})
	if err != nil {
		t.Fatalf("add worker %s: %v", id, err)
	}
}

// addSEDWorkerCaps registers a C++ worker with caller-supplied CPU fields,
// for capacity-fallback cases.
func addSEDWorkerCaps(t *testing.T, reg *registry.InMemoryRegistry, id string, cores, millis int32, maxParallel int32) {
	t.Helper()
	err := reg.Add(&registry.WorkerInfo{
		ID:          id,
		Address:     id + ":50051",
		MaxParallel: maxParallel,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			CpuCores:   cores,
			CpuMillis:  millis,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
	})
	if err != nil {
		t.Fatalf("add worker %s: %v", id, err)
	}
}

// addBenchmarkCluster registers the five-worker heterogeneous cluster from
// test/stress/docker-compose-hetero.yml (quotas 0.5/0.6/0.8/1.0/1.1 cores,
// max-parallel 1/2/2/2/3).
func addBenchmarkCluster(t *testing.T, reg *registry.InMemoryRegistry) {
	t.Helper()
	addSEDWorker(t, reg, "w1", 500, 1)
	addSEDWorker(t, reg, "w2", 600, 2)
	addSEDWorker(t, reg, "w3", 800, 2)
	addSEDWorker(t, reg, "w4", 1000, 2)
	addSEDWorker(t, reg, "w5", 1100, 3)
}

func selectSED(t *testing.T, s Scheduler) string {
	t.Helper()
	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	return w.ID
}

func TestSEDScheduler_IsNotALearningScheduler(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	var s Scheduler = NewSEDScheduler(SEDConfig{Registry: reg})
	if _, ok := s.(LearningScheduler); ok {
		t.Fatal("SED must not implement LearningScheduler: it never learns from outcomes")
	}
}

func TestSEDScheduler_DispatchSequenceOnBenchmarkCluster(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addBenchmarkCluster(t, reg)

	s := NewSEDScheduler(SEDConfig{Registry: reg})

	// Hand-derived (score = (active+1)/millis, lowest wins, no task ever
	// completes). Step 6 is the exact tie w1 (1/500) vs w4 (2/1000): the
	// larger capacity, w4, must win.
	want := []string{"w5", "w4", "w3", "w2", "w5", "w4", "w1", "w3", "w5", "w2"}
	for i, wantID := range want {
		got := selectSED(t, s)
		if got != wantID {
			t.Fatalf("dispatch %d: got %s, want %s", i+1, got, wantID)
		}
		if err := reg.IncrementTasks(got); err != nil {
			t.Fatalf("IncrementTasks(%s): %v", got, err)
		}
	}

	// Total slots are 1+2+2+2+3 = 10; the 11th dispatch has nowhere to go.
	_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if !errors.Is(err, ErrNoMatchingWorkers) {
		t.Fatalf("dispatch 11: got %v, want ErrNoMatchingWorkers", err)
	}
}

// SED's reason to exist: LeastLoaded compares raw ActiveTasks and sends work
// to a slow idle worker even when a much faster worker has spare capacity.
func TestSEDScheduler_DiffersFromLeastLoaded(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addBenchmarkCluster(t, reg)

	// Only w1 (500 millis, idle) and w5 (1100 millis, one task running)
	// stay eligible; fill the rest.
	for _, id := range []string{"w2", "w3", "w4"} {
		for i := 0; i < 2; i++ {
			if err := reg.IncrementTasks(id); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := reg.IncrementTasks("w5"); err != nil {
		t.Fatal(err)
	}

	sed := NewSEDScheduler(SEDConfig{Registry: reg})
	ll := NewLeastLoadedScheduler(reg)

	// SED: w5 = 2/1100 = 0.00182 < w1 = 1/500 = 0.002.
	if got := selectSED(t, sed); got != "w5" {
		t.Errorf("SED chose %s, want w5", got)
	}
	// LeastLoaded: w1 has 0 active tasks, w5 has 1.
	if got := selectSED(t, ll); got != "w1" {
		t.Errorf("LeastLoaded chose %s, want w1", got)
	}
}

func TestSEDScheduler_NeverSelectsFullWorker(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	// The fastest worker is full; it must be skipped however attractive.
	addSEDWorker(t, reg, "fast-full", 4000, 1)
	addSEDWorker(t, reg, "slow-free", 500, 1)
	if err := reg.IncrementTasks("fast-full"); err != nil {
		t.Fatal(err)
	}

	s := NewSEDScheduler(SEDConfig{Registry: reg})
	for i := 0; i < 20; i++ {
		if got := selectSED(t, s); got != "slow-free" {
			t.Fatalf("iteration %d: got %s, want slow-free", i, got)
		}
	}
}

func TestSEDScheduler_ExcludesUnhealthyWorker(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorker(t, reg, "fast", 4000, 2)
	addSEDWorker(t, reg, "slow", 500, 2)
	if err := reg.UpdateState("fast", registry.WorkerStateUnhealthy); err != nil {
		t.Fatal(err)
	}

	s := NewSEDScheduler(SEDConfig{Registry: reg})
	if got := selectSED(t, s); got != "slow" {
		t.Errorf("got %s, want slow (fast is unhealthy)", got)
	}
}

func TestSEDScheduler_CircuitBreaker(t *testing.T) {
	tests := []struct {
		name string
		open map[string]bool
		want string
	}{
		{"open circuit excluded", map[string]bool{"fast": true}, "slow"},
		{"closed circuits ignored", map[string]bool{}, "fast"},
		// Every circuit open: relax the filter rather than stall (same
		// rule as the other schedulers via eligibleCandidates), so the
		// fastest worker is chosen again.
		{"all circuits open relaxes", map[string]bool{"fast": true, "slow": true}, "fast"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newTestRegistry()
			defer reg.Stop()
			addSEDWorker(t, reg, "fast", 4000, 2)
			addSEDWorker(t, reg, "slow", 500, 2)

			s := NewSEDScheduler(SEDConfig{
				Registry:       reg,
				CircuitChecker: &mockCircuitChecker{openWorkers: tt.open},
			})
			if got := selectSED(t, s); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestSEDScheduler_CapacityFallback(t *testing.T) {
	tests := []struct {
		name  string
		cores int32
		milli int32
		want  int64
	}{
		{"cgroup quota wins", 20, 1100, 1100},
		{"no quota falls back to cores*1000", 2, 0, 2000},
		{"no information defaults to 1000", 0, 0, 1000},
		{"negative quota and no cores defaults to 1000", 0, -5, 1000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &registry.WorkerInfo{Capabilities: &pb.WorkerCapabilities{CpuCores: tt.cores, CpuMillis: tt.milli}}
			if got := sedCapacity(w); got != tt.want {
				t.Errorf("sedCapacity = %d, want %d", got, tt.want)
			}
		})
	}

	t.Run("nil capabilities default to 1000", func(t *testing.T) {
		if got := sedCapacity(&registry.WorkerInfo{}); got != 1000 {
			t.Errorf("sedCapacity = %d, want 1000", got)
		}
	})
}

// A worker that reports nothing is assumed to be a 1000-millicore worker, so
// it must rank between a 500 and a 2000 worker rather than first or last.
func TestSEDScheduler_UnreportedCapacityRanksAsOneCore(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorkerCaps(t, reg, "unknown", 0, 0, 2)
	addSEDWorkerCaps(t, reg, "half", 0, 500, 2)
	s := NewSEDScheduler(SEDConfig{Registry: reg})
	if got := selectSED(t, s); got != "unknown" {
		t.Errorf("unknown(1000) vs half(500): got %s, want unknown", got)
	}

	reg2 := newTestRegistry()
	defer reg2.Stop()
	addSEDWorkerCaps(t, reg2, "unknown", 0, 0, 2)
	addSEDWorkerCaps(t, reg2, "double", 0, 2000, 2)
	s2 := NewSEDScheduler(SEDConfig{Registry: reg2})
	if got := selectSED(t, s2); got != "double" {
		t.Errorf("unknown(1000) vs double(2000): got %s, want double", got)
	}
}

func TestSEDScheduler_CoresFallbackWhenNoQuota(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	// 2 cores (no cgroup quota) = 2000 millis beats a 1500-millicore worker.
	addSEDWorkerCaps(t, reg, "cores2", 2, 0, 2)
	addSEDWorkerCaps(t, reg, "quota1500", 20, 1500, 2)
	s := NewSEDScheduler(SEDConfig{Registry: reg})
	if got := selectSED(t, s); got != "cores2" {
		t.Errorf("got %s, want cores2", got)
	}
}

func TestSEDScheduler_TieBreak(t *testing.T) {
	t.Run("equal score prefers larger capacity", func(t *testing.T) {
		reg := newTestRegistry()
		defer reg.Stop()
		addSEDWorker(t, reg, "a-small", 500, 2) // 1/500
		addSEDWorker(t, reg, "z-big", 1000, 2)  // 1/1000 ... made equal below
		// Make the scores equal: a-small idle = 1/500; z-big with one task = 2/1000.
		if err := reg.IncrementTasks("z-big"); err != nil {
			t.Fatal(err)
		}
		s := NewSEDScheduler(SEDConfig{Registry: reg})
		for i := 0; i < 50; i++ {
			if got := selectSED(t, s); got != "z-big" {
				t.Fatalf("iteration %d: got %s, want z-big (larger capacity wins an exact tie)", i, got)
			}
		}
	})

	t.Run("identical workers resolve by lexicographic ID", func(t *testing.T) {
		// The registry iterates a map, so candidate order is random per
		// registry; the result must not depend on it. Fresh registries
		// give fresh map layouts.
		for i := 0; i < 100; i++ {
			reg := newTestRegistry()
			for _, id := range []string{"w3", "w1", "w5", "w2", "w4"} {
				addSEDWorker(t, reg, id, 1000, 2)
			}
			got := selectSED(t, NewSEDScheduler(SEDConfig{Registry: reg}))
			reg.Stop()
			if got != "w1" {
				t.Fatalf("trial %d: got %s, want w1", i, got)
			}
		}
	})
}

func TestSEDScheduler_Errors(t *testing.T) {
	t.Run("no workers", func(t *testing.T) {
		reg := newTestRegistry()
		defer reg.Stop()
		s := NewSEDScheduler(SEDConfig{Registry: reg})
		_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
		if !errors.Is(err, ErrNoWorkers) {
			t.Errorf("got %v, want ErrNoWorkers", err)
		}
	})

	t.Run("no matching workers", func(t *testing.T) {
		reg := newTestRegistry()
		defer reg.Stop()
		if err := reg.Add(&registry.WorkerInfo{
			ID:      "go-only",
			Address: "go-only:50051",
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				CpuMillis:  1000,
				Go:         &pb.GoCapability{Version: "1.22"},
			},
		}); err != nil {
			t.Fatal(err)
		}
		s := NewSEDScheduler(SEDConfig{Registry: reg})
		_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
		if !errors.Is(err, ErrNoMatchingWorkers) {
			t.Errorf("got %v, want ErrNoMatchingWorkers", err)
		}
	})

	t.Run("all workers full", func(t *testing.T) {
		reg := newTestRegistry()
		defer reg.Stop()
		addSEDWorker(t, reg, "only", 1000, 1)
		if err := reg.IncrementTasks("only"); err != nil {
			t.Fatal(err)
		}
		s := NewSEDScheduler(SEDConfig{Registry: reg})
		_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
		if !errors.Is(err, ErrNoMatchingWorkers) {
			t.Errorf("got %v, want ErrNoMatchingWorkers", err)
		}
	})
}

// MaxParallel <= 0 means "4" in every scheduler's admission rule.
func TestSEDScheduler_DefaultMaxParallelIsFour(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorker(t, reg, "only", 1000, 0)
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	for i := 0; i < 4; i++ {
		id := selectSED(t, s)
		if err := reg.IncrementTasks(id); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, ""); !errors.Is(err, ErrNoMatchingWorkers) {
		t.Errorf("fifth dispatch: got %v, want ErrNoMatchingWorkers", err)
	}
}

// SED sees only the advertised quota and ActiveTasks, never observed speed:
// a worker that became very slow must still be chosen on its quota (the
// drift experiment relies on this).
func TestSEDScheduler_IgnoresCompletionDurations(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorker(t, reg, "fast", 1100, 3)
	addSEDWorker(t, reg, "slow", 500, 1)
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	// Completing tasks with absurd durations ("fast" became very slow) must
	// not change SED's choice: it only sees quota and ActiveTasks.
	for i := 0; i < 3; i++ {
		if err := reg.IncrementTasks("fast"); err != nil {
			t.Fatal(err)
		}
		if err := reg.DecrementTasks("fast", true, 3600e9); err != nil {
			t.Fatal(err)
		}
	}
	if got := selectSED(t, s); got != "fast" {
		t.Errorf("got %s, want fast", got)
	}
}

// sedLess compares exact 128-bit cross-products; extreme-but-legal inputs
// (int32 active counts, millicores up to int32*1000) must neither overflow nor tie.
func TestSEDScheduler_Less(t *testing.T) {
	tests := []struct {
		name           string
		a1, c1, a2, c2 int64
		want           bool
	}{
		{"lower load wins", 0, 1000, 1, 1000, true},
		{"higher capacity wins", 1, 500, 1, 1000, false},
		{"exact tie is not less", 1, 1000, 0, 500, false},
		{"adjacent huge capacities", 0, math.MaxInt32, 0, math.MaxInt32 - 1, true},
		{"large active counts", 1 << 30, math.MaxInt32, 1 << 30, math.MaxInt32 - 1, true},
		// (a+1)*c2 = 2^63-2^31 fits in int64 but (a+1)*c1 = 2^63 wraps negative, so a
		// signed int64 product answers false here. Both inputs are legal registry
		// values (ActiveTasks is int32, capacity up to int32*1000).
		{"signed int64 product wraps", 1<<31 - 1, 1 << 32, 1<<31 - 1, 1<<32 - 1, true},
		// Cross-products differ by exactly 1 (1e18-1 vs 1e18) while the two float64
		// quotients agree to ~1e-18, below float64 resolution.
		{"differs below float64 resolution", 999999998, 1000000000, 999999999, 1000000001, true},
		{"signed int64 product wraps, reversed", 1<<31 - 1, 1<<32 - 1, 1<<31 - 1, 1 << 32, false},
		{"differs below float64 resolution, reversed", 999999999, 1000000001, 999999998, 1000000000, false},
		// Exact pair from the other implementation: cross-products differ by 1 but
		// both float64 quotients round to the same value.
		{"float64-equal pair, first is larger", 2147483642, 2147483647000, 1882806283, 1882806287507, false},
		{"float64-equal pair, first is smaller", 1882806283, 1882806287507, 2147483642, 2147483647000, true},
		// (a+1)*c up to ~4.6e21 for the largest legal int32 count and capacity.
		{"max count and capacity, larger score", math.MaxInt32 - 1, math.MaxInt32*1000 - 1, 0, math.MaxInt32 * 1000, false},
		{"max count and capacity, smaller score", 0, math.MaxInt32 * 1000, math.MaxInt32 - 1, math.MaxInt32*1000 - 1, true},
		{"max count and capacity, tie", math.MaxInt32 - 1, math.MaxInt32 * 1000, math.MaxInt32 - 1, math.MaxInt32 * 1000, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sedLess(tt.a1, tt.c1, tt.a2, tt.c2); got != tt.want {
				t.Errorf("sedLess(%d/%d, %d/%d) = %v, want %v", tt.a1+1, tt.c1, tt.a2+1, tt.c2, got, tt.want)
			}
		})
	}
}

func TestSEDScheduler_ConcurrentSelect(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addBenchmarkCluster(t, reg)
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for g := 0; g < 8; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
				if err != nil {
					errs <- fmt.Errorf("select: %w", err)
					return
				}
				if w.ID == "" {
					errs <- errors.New("empty worker ID")
					return
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}

// SED is a plain Scheduler: the gRPC layer reaches it through SelectWith,
// which must hand back an empty DispatchInfo (no learner state to report).
func TestSEDScheduler_SelectWithReturnsZeroDispatchInfo(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addBenchmarkCluster(t, reg)
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	w, info, err := SelectWith(s, pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	if err != nil {
		t.Fatal(err)
	}
	if w.ID != "w5" {
		t.Errorf("got %s, want w5", w.ID)
	}
	if info != (DispatchInfo{}) {
		t.Errorf("DispatchInfo = %+v, want zero value", info)
	}
}

func TestSEDScheduler_ClientOSFilter(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	if err := reg.Add(&registry.WorkerInfo{
		ID:      "linux-nodocker",
		Address: "linux-nodocker:50051",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Os:         "linux",
			CpuMillis:  4000,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
	}); err != nil {
		t.Fatal(err)
	}
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	if _, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "linux"); err != nil {
		t.Errorf("same-OS client: %v", err)
	}
	_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "windows")
	if !errors.Is(err, ErrNoMatchingWorkers) {
		t.Errorf("cross-OS client without Docker: got %v, want ErrNoMatchingWorkers", err)
	}
}

// SED reads capacity from the registry snapshot on every Select: if a worker
// re-registers with a different quota, the next choice must follow it. (SED
// itself never refreshes or caches capacity.)
func TestSEDScheduler_ReadsRegisteredCapacityEachSelect(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorker(t, reg, "fast", 1100, 3)
	addSEDWorker(t, reg, "slow", 500, 3)
	s := NewSEDScheduler(SEDConfig{Registry: reg})

	if got := selectSED(t, s); got != "fast" {
		t.Fatalf("before re-registration: got %s, want fast", got)
	}
	// Re-registering an existing ID replaces its capabilities.
	addSEDWorker(t, reg, "fast", 100, 3)
	if got := selectSED(t, s); got != "slow" {
		t.Errorf("after quota drop to 100: got %s, want slow", got)
	}
}

// Each fallback capacity is bracketed by workers one milli-core either side,
// so Select itself must use exactly 1000 (nothing reported) and
// CpuCores*1000 (no cgroup quota).
func TestSEDScheduler_CapacityFallbackBracketed(t *testing.T) {
	tests := []struct {
		name  string
		cores int32
		other int32 // CpuMillis of the competing worker
		want  string
	}{
		{"no report beats 999", 0, 999, "probe"},
		{"no report loses to 1001", 0, 1001, "other"},
		{"2 cores beats 1999", 2, 1999, "probe"},
		{"2 cores loses to 2001", 2, 2001, "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := newTestRegistry()
			defer reg.Stop()
			addSEDWorkerCaps(t, reg, "probe", tt.cores, 0, 1)
			addSEDWorkerCaps(t, reg, "other", 0, tt.other, 1)
			s := NewSEDScheduler(SEDConfig{Registry: reg})
			if got := selectSED(t, s); got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// When every circuit is open the filter relaxes, but only for circuits:
// unhealthy and full workers stay excluded.
func TestSEDScheduler_RelaxationKeepsHealthAndCapacityFilters(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	addSEDWorker(t, reg, "best-unhealthy", 8000, 4)
	addSEDWorker(t, reg, "best-full", 8000, 1)
	addSEDWorker(t, reg, "fast", 1100, 3)
	addSEDWorker(t, reg, "slow", 500, 1)
	if err := reg.UpdateState("best-unhealthy", registry.WorkerStateUnhealthy); err != nil {
		t.Fatal(err)
	}
	if err := reg.IncrementTasks("best-full"); err != nil {
		t.Fatal(err)
	}

	s := NewSEDScheduler(SEDConfig{
		Registry: reg,
		CircuitChecker: &mockCircuitChecker{openWorkers: map[string]bool{
			"best-unhealthy": true, "best-full": true, "fast": true, "slow": true,
		}},
	})
	if got := selectSED(t, s); got != "fast" {
		t.Errorf("got %s, want fast (best healthy worker with room)", got)
	}
}

// Add accepts any ActiveTasks, so a negative count must behave like zero
// instead of wrapping to a huge unsigned load and being ranked last.
func TestSEDScheduler_NegativeActiveTasksTreatedAsZero(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	for _, id := range []string{"a-negative", "z-zero"} {
		if err := reg.Add(&registry.WorkerInfo{
			ID:          id,
			Address:     id,
			MaxParallel: 2,
			ActiveTasks: map[string]int32{"a-negative": -3, "z-zero": 0}[id],
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				CpuMillis:  1000,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	s := NewSEDScheduler(SEDConfig{Registry: reg})
	// Both score as load 1 at equal capacity, so the raw-ActiveTasks level
	// decides and a-negative wins. A negative count wrapped to unsigned would
	// rank a-negative last and pick z-zero.
	for i := 0; i < 20; i++ {
		if got := selectSED(t, s); got != "a-negative" {
			t.Fatalf("iteration %d: got %s, want a-negative", i, got)
		}
	}
}

// Tie-break order is score, larger capacity, fewer ActiveTasks, smaller ID.
// The ActiveTasks level is reachable only through a negative count (clamped
// to zero when scoring), so it is pinned here with one.
func TestSEDScheduler_TieBreakPrefersFewerActiveTasksBeforeID(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	for id, active := range map[string]int32{"a-zero": 0, "z-negative": -2} {
		if err := reg.Add(&registry.WorkerInfo{
			ID:          id,
			Address:     id,
			MaxParallel: 2,
			ActiveTasks: active,
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				CpuMillis:  1000,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	s := NewSEDScheduler(SEDConfig{Registry: reg})
	if got := selectSED(t, s); got != "z-negative" {
		t.Errorf("got %s, want z-negative (fewer raw ActiveTasks beats smaller ID)", got)
	}
}

// A negative count must score as zero, not as a negative load: the 500-milli
// worker with ActiveTasks=-3 scores 1/500 (not -2/500), so the idle 1000-milli
// worker must win.
func TestSEDScheduler_NegativeActiveTasksScoreAsZeroNotNegative(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()
	for _, w := range []struct {
		id     string
		millis int32
		active int32
	}{{"weak-negative", 500, -3}, {"strong-idle", 1000, 0}} {
		if err := reg.Add(&registry.WorkerInfo{
			ID:          w.id,
			Address:     w.id,
			MaxParallel: 2,
			ActiveTasks: w.active,
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				CpuMillis:  w.millis,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	s := NewSEDScheduler(SEDConfig{Registry: reg})
	if got := selectSED(t, s); got != "strong-idle" {
		t.Errorf("got %s, want strong-idle", got)
	}
}
