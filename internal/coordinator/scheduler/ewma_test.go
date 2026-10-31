package scheduler

import (
	"fmt"
	"math"
	"sort"
	"sync"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

const (
	ewmaBuild = pb.BuildType_BUILD_TYPE_CPP
	ewmaArch  = pb.Architecture_ARCH_X86_64
)

// Fixed ordering makes first-wins assertions independent of map iteration.
type ewmaOrderedRegistry struct {
	*registry.InMemoryRegistry
	order map[string]int
}

func (r *ewmaOrderedRegistry) ListByCapability(build pb.BuildType, arch pb.Architecture) []*registry.WorkerInfo {
	workers := r.InMemoryRegistry.ListByCapability(build, arch)
	sort.Slice(workers, func(i, j int) bool { return r.order[workers[i].ID] < r.order[workers[j].ID] })
	return workers
}

func newEWMARegistry(t *testing.T, ids ...string) *ewmaOrderedRegistry {
	t.Helper()
	r := &ewmaOrderedRegistry{InMemoryRegistry: newTestRegistry(), order: make(map[string]int)}
	t.Cleanup(r.Stop)
	for i, id := range ids {
		r.order[id] = i
		if err := r.Add(&registry.WorkerInfo{
			ID: id, MaxParallel: 4,
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: ewmaArch, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatalf("add %s: %v", id, err)
		}
	}
	return r
}

func ewmaReward(ms float64) float64 {
	return -math.Log1p(ms) / math.Log1p(ewmaRewardNormalizationMs)
}

func ewmaDispatch(t *testing.T, s *EWMAScheduler, id string, size int) (*registry.WorkerInfo, DispatchInfo) {
	t.Helper()
	w, info, err := s.SelectWithDispatchInfo(ewmaBuild, ewmaArch, "", TaskContext{TaskID: id, SourceSizeBytes: size})
	if err != nil {
		t.Fatalf("dispatch %s: %v", id, err)
	}
	return w, info
}

func ewmaObserve(t *testing.T, s *EWMAScheduler, id string, size int, ms float64) string {
	t.Helper()
	w, _ := ewmaDispatch(t, s, id, size)
	s.RecordOutcome(w.ID, ewmaReward(ms), true, TaskContext{TaskID: id})
	return w.ID
}

func TestEWMAScheduler_Params(t *testing.T) {
	for _, tc := range []struct {
		name      string
		warmStart int
		want      string
	}{
		{"default", 0, "alpha=0.2 warm_start=0"},
		{"configured", 100, "alpha=0.2 warm_start=100"},
		{"negative clamps to zero", -1, "alpha=0.2 warm_start=0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEWMAScheduler(EWMAConfig{WarmStartTasks: tc.warmStart})
			if got := s.Params(); got != tc.want {
				t.Fatalf("Params() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestEWMA_RewardInversion(t *testing.T) {
	for _, tc := range []struct {
		name string
		ms   float64
	}{
		{"success", 73.25},
		{"failure default timeout", 120000},
		{"failure extended timeout", 180000},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := ewmaRewardToDurationMs(ewmaReward(tc.ms))
			if math.Abs(got-tc.ms) > 1e-8 {
				t.Fatalf("inversion: got %.12g, want %.12g", got, tc.ms)
			}
		})
	}
}

func TestEWMA_FirstObservationAndAlpha(t *testing.T) {
	r := newEWMARegistry(t, "a")
	s := NewEWMAScheduler(EWMAConfig{Registry: r})
	for _, tc := range []struct {
		id   string
		size int
		ms   float64
		want float64
	}{
		{"first", 10, 100, math.Log(10)},
		{"second", 10, 200, 0.8*math.Log(10) + 0.2*math.Log(20)},
		{"failure", 0, 120000, 0.8*(0.8*math.Log(10)+0.2*math.Log(20)) + 0.2*math.Log(120000)},
	} {
		t.Run(tc.id, func(t *testing.T) {
			w, _ := ewmaDispatch(t, s, tc.id, tc.size)
			s.RecordOutcome(w.ID, ewmaReward(tc.ms), tc.id != "failure", TaskContext{TaskID: tc.id})
			if got := s.est["a"]; math.Abs(got-tc.want) > 1e-8 {
				t.Fatalf("estimate = %g, want %g", got, tc.want)
			}
			if len(s.pending) != 0 {
				t.Fatalf("pending size retained after %s", tc.id)
			}
		})
	}
	if s.n["a"] != 3 {
		t.Fatalf("observations = %d, want 3", s.n["a"])
	}
}

func TestEWMA_WarmStartLearnsAndExploresUnobserved(t *testing.T) {
	r := newEWMARegistry(t, "busy", "idle")
	if err := r.IncrementTasks("busy"); err != nil {
		t.Fatal(err)
	}
	s := NewEWMAScheduler(EWMAConfig{Registry: r, WarmStartTasks: 2})
	for i := 0; i < 2; i++ {
		id := fmt.Sprintf("warm-%d", i)
		w, info := ewmaDispatch(t, s, id, 10)
		wantQ := 0.0
		if i > 0 {
			wantQ = -math.Log(10)
		}
		if w.ID != "idle" || !info.WasExploration || math.Abs(info.QValueAtDispatch-wantQ) > 1e-8 {
			t.Fatalf("warm start %d: got %s, %+v", i, w.ID, info)
		}
		s.RecordOutcome(w.ID, ewmaReward(100), true, TaskContext{TaskID: id})
	}
	if s.n["idle"] != 2 || math.Abs(s.est["idle"]-math.Log(10)) > 1e-8 {
		t.Fatalf("warm-start outcome lost: n=%d, estimate=%g", s.n["idle"], s.est["idle"])
	}
	w, info := ewmaDispatch(t, s, "explore", 10)
	if w.ID != "busy" || !info.WasExploration || math.Abs(info.QValueAtDispatch+math.Log(2)) > 1e-8 {
		t.Fatalf("unobserved worker: got %s, %+v", w.ID, info)
	}
	s.RecordOutcome(w.ID, ewmaReward(1000), true, TaskContext{TaskID: "explore"})
	w, info = ewmaDispatch(t, s, "exploit", 10)
	if w.ID != "idle" || info.WasExploration || math.Abs(info.QValueAtDispatch+math.Log(10)) > 1e-8 {
		t.Fatalf("greedy choice: got %s, %+v", w.ID, info)
	}
}

func TestEWMA_SlowdownAndRecovery(t *testing.T) {
	r := newEWMARegistry(t, "slow", "fast")
	s := NewEWMAScheduler(EWMAConfig{Registry: r})
	if id := ewmaObserve(t, s, "slow-1", 10, 1000); id != "slow" {
		t.Fatal(id)
	}
	if id := ewmaObserve(t, s, "fast-1", 10, 100); id != "fast" {
		t.Fatal(id)
	}
	if math.Abs(s.est["slow"]-math.Log(100)) > 1e-8 || math.Abs(s.est["fast"]-math.Log(10)) > 1e-8 {
		t.Fatalf("initial estimates: slow=%g fast=%g", s.est["slow"], s.est["fast"])
	}
	w, info := ewmaDispatch(t, s, "avoid", 10)
	if w.ID != "fast" || math.Abs(info.QValueAtDispatch+math.Log(10)) > 1e-8 {
		t.Fatalf("slow worker selected or wrong score: %s, %+v", w.ID, info)
	}
	s.RecordOutcome(w.ID, ewmaReward(100), true, TaskContext{TaskID: "avoid"})
	// Fill the fast worker so the recovering worker receives observations.
	for i := 0; i < 4; i++ {
		if err := r.IncrementTasks("fast"); err != nil {
			t.Fatal(err)
		}
	}
	for i := 0; i < 12; i++ {
		if id := ewmaObserve(t, s, fmt.Sprintf("recover-%d", i), 10, 10); id != "slow" {
			t.Fatalf("recovery sent to %s", id)
		}
	}
	for i := 0; i < 4; i++ {
		if err := r.DecrementTasks("fast", true, 0); err != nil {
			t.Fatal(err)
		}
	}
	wantRecovered := math.Pow(1-ewmaAlpha, 12) * math.Log(100)
	if math.Abs(s.est["slow"]-wantRecovered) > 1e-8 {
		t.Fatalf("recovered estimate = %g, want %g", s.est["slow"], wantRecovered)
	}
	w, info = ewmaDispatch(t, s, "after-recovery", 10)
	if w.ID != "slow" || math.Abs(info.QValueAtDispatch+wantRecovered) > 1e-8 {
		t.Fatalf("estimate did not recover: slow=%g fast=%g, chose %s", s.est["slow"], s.est["fast"], w.ID)
	}
}

func TestEWMA_HeavyTailOutlier(t *testing.T) {
	r := newEWMARegistry(t, "fast")
	s := NewEWMAScheduler(EWMAConfig{Registry: r})
	ewmaObserve(t, s, "ordinary", 10, 100)    // 10 ms per byte.
	ewmaObserve(t, s, "outlier", 10, 10000)   // 100x slower.
	const wantLogEstimate = 3.223619130191664 // 0.8*ln(10) + 0.2*ln(1000).
	if got := s.est["fast"]; math.Abs(got-wantLogEstimate) > 1e-8 {
		t.Fatalf("log estimate = %.12g, want %.12g", got, wantLogEstimate)
	}
	// The corresponding linear EWMA would be 208 ms per byte.
	linearEstimate := 0.8*10 + 0.2*1000
	if got := math.Exp(s.est["fast"]); math.Abs(got-25.1188643150958) > 1e-8 || got > linearEstimate/4+1e-8 {
		t.Fatalf("estimated ms per byte = %.12g, linear estimate = %.12g", got, linearEstimate)
	}
}

func TestEWMA_LogScorePreservesProductOrdering(t *testing.T) {
	cases := []struct {
		active int32
		est    float64
	}{
		{0, math.Log(10)},
		{1, math.Log(6)},
		{3, math.Log(2)},
		{0, math.Log(100)},
		{2, math.Log(0.5)},
	}
	s := NewEWMAScheduler(EWMAConfig{})
	scores := make([]float64, len(cases))
	products := make([]float64, len(cases))
	for i, tc := range cases {
		w := &registry.WorkerInfo{ID: fmt.Sprintf("worker-%d", i), ActiveTasks: tc.active}
		s.est[w.ID] = tc.est
		scores[i] = s.score(w)
		products[i] = float64(tc.active+1) * math.Exp(tc.est)
		if math.Abs(scores[i]-math.Log(products[i])) > 1e-8 {
			t.Fatalf("case %d: log score = %g, log product = %g", i, scores[i], math.Log(products[i]))
		}
	}
	for i := range cases {
		for j := i + 1; j < len(cases); j++ {
			if (scores[i]-scores[j] < -1e-8) != (products[i]-products[j] < -1e-8) {
				t.Fatalf("cases %d and %d have different ordering", i, j)
			}
		}
	}
}

func TestEWMA_CapacityAndRegistryOrder(t *testing.T) {
	for _, tc := range []struct {
		name string
		full bool
		want string
	}{
		{"tie preserves registry order", false, "first"},
		{"full worker excluded", true, "second"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := newEWMARegistry(t, "first", "second")
			s := NewEWMAScheduler(EWMAConfig{Registry: r})
			s.n["first"], s.n["second"] = 1, 1
			s.est["first"], s.est["second"] = math.Log(10), math.Log(10)
			if tc.full {
				for i := 0; i < 4; i++ {
					if err := r.IncrementTasks("first"); err != nil {
						t.Fatal(err)
					}
				}
			}
			w, info := ewmaDispatch(t, s, tc.name, 1)
			if w.ID != tc.want || info.WasExploration {
				t.Fatalf("choice = %s, %+v; want %s", w.ID, info, tc.want)
			}
		})
	}
}

func TestEWMA_PendingCleanupAndConcurrentFeedback(t *testing.T) {
	r := newEWMARegistry(t, "a", "b")
	s := NewEWMAScheduler(EWMAConfig{Registry: r})
	const tasks = 100
	var wg sync.WaitGroup
	for i := 0; i < tasks; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("task-%d", i)
			ctx := TaskContext{TaskID: id, SourceSizeBytes: 10}
			w, _, err := s.SelectWithDispatchInfo(ewmaBuild, ewmaArch, "", ctx)
			if err != nil {
				t.Errorf("select: %v", err)
				return
			}
			s.RecordOutcome(w.ID, ewmaReward(100), true, ctx)
		}(i)
	}
	wg.Wait()
	if len(s.pending) != 0 || s.n["a"]+s.n["b"] != tasks {
		t.Fatalf("pending=%d observations=%d", len(s.pending), s.n["a"]+s.n["b"])
	}
	ctx := TaskContext{TaskID: "retry", SourceSizeBytes: 10}
	w, _ := ewmaDispatch(t, s, ctx.TaskID, ctx.SourceSizeBytes)
	s.RecordOutcome(w.ID, math.NaN(), false, ctx)
	if len(s.pending) != 0 {
		t.Fatal("invalid outcome retained pending entry")
	}
}
