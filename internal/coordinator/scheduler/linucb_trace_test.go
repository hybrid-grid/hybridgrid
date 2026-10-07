package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// This file is the deterministic trace driver for Hybrid-LinUCB and the test
// that pins today's behaviour against testdata/hybrid_linucb_golden.json.
//
// The golden file was generated from the UNMODIFIED scheduler before the
// discounted variant was written; it is the reference for "discount off is
// identical to the old code". Regenerate it only deliberately:
//
//	UPDATE_GOLDEN=1 go test -run TestLinUCBTrace_Golden ./internal/coordinator/scheduler/
//
// Comparison policy: selections, exploration flags and counts are compared
// exactly; floating-point state with a relative tolerance, because gonum
// kernels may differ in the last bits across CPUs. Bit-exact equality between
// two scheduler configurations is checked at run time on one machine instead
// (see sameTrace).

const (
	traceSteps      = 400
	traceGoldenPath = "testdata/hybrid_linucb_golden.json"
	traceTolerance  = 1e-12
)

// sortedRegistry makes candidate order deterministic: the real registry
// iterates a Go map, so the scheduler would otherwise see workers in a
// different order on every run.
type sortedRegistry struct {
	registry.Registry
}

func (r sortedRegistry) ListByCapability(buildType pb.BuildType, arch pb.Architecture) []*registry.WorkerInfo {
	ws := r.Registry.ListByCapability(buildType, arch)
	sort.Slice(ws, func(i, j int) bool { return ws[i].ID < ws[j].ID })
	return ws
}

// traceArm is the exported state of one bandit arm.
type traceArm struct {
	A     [][]float64 `json:"a"`
	Ainv  [][]float64 `json:"ainv"`
	B     []float64   `json:"b"`
	Count int64       `json:"count"`
}

// traceResult is everything the golden file pins.
type traceResult struct {
	Selections  []string             `json:"selections"`
	QValues     []float64            `json:"q_values"`
	Exploration []bool               `json:"exploration"`
	Dispatches  int64                `json:"dispatches"`
	Arms        map[string]*traceArm `json:"arms"`
}

// newTraceCluster registers the five-worker heterogeneous benchmark cluster
// (quotas 0.5/0.6/0.8/1.0/1.1 cores, max-parallel 1/2/2/2/3).
func newTraceCluster(t *testing.T) *registry.InMemoryRegistry {
	t.Helper()
	reg := registry.NewInMemoryRegistry(time.Hour)
	t.Cleanup(reg.Stop)
	specs := []struct {
		id          string
		millis      int32
		maxParallel int32
	}{
		{"w1", 500, 1}, {"w2", 600, 2}, {"w3", 800, 2}, {"w4", 1000, 2}, {"w5", 1100, 3},
	}
	for _, sp := range specs {
		err := reg.Add(&registry.WorkerInfo{
			ID:          sp.id,
			Address:     sp.id + ":50051",
			MaxParallel: sp.maxParallel,
			Capabilities: &pb.WorkerCapabilities{
				NativeArch:  pb.Architecture_ARCH_X86_64,
				CpuCores:    20,
				CpuMillis:   sp.millis,
				MemoryBytes: int64(sp.millis) * 1 << 20,
				Cpp:         &pb.CppCapability{Compilers: []string{"gcc", "g++"}},
			},
		})
		if err != nil {
			t.Fatalf("add worker %s: %v", sp.id, err)
		}
	}
	return reg
}

// newTraceScheduler builds the Hybrid-LinUCB used by the golden trace
// (alpha 0.5, warm-start 100, load penalty 0.5) on a deterministic registry.
func newTraceScheduler(reg *registry.InMemoryRegistry) *LinUCBScheduler {
	return NewLinUCBScheduler(LinUCBConfig{
		Registry:       sortedRegistry{reg},
		Alpha:          0.5,
		WarmStartTasks: 100,
		LoadPenalty:    0.5,
	})
}

type traceTask struct {
	ctx    TaskContext
	worker string
	doneAt int
	reward float64
	dur    time.Duration
}

// runTrace drives s through traceSteps dispatches with delayed completions.
// Everything is seeded: sizes, completion lags, reward noise. The reward of a
// task is fixed at dispatch so it does not depend on completion order.
func runTrace(t *testing.T, s *LinUCBScheduler, reg *registry.InMemoryRegistry) traceResult {
	t.Helper()
	rng := rand.New(rand.NewSource(20261007))
	speed := map[string]float64{"w1": 0.5, "w2": 0.6, "w3": 0.8, "w4": 1.0, "w5": 1.1}

	res := traceResult{Arms: map[string]*traceArm{}}
	var inflight []*traceTask

	complete := func(i int) {
		task := inflight[i]
		inflight = append(inflight[:i], inflight[i+1:]...)
		if err := reg.DecrementTasks(task.worker, true, task.dur); err != nil {
			t.Fatalf("DecrementTasks: %v", err)
		}
		s.RecordOutcome(task.worker, task.reward, true, task.ctx)
	}

	for step := 0; step < traceSteps; step++ {
		// Complete everything that is due.
		for i := 0; i < len(inflight); {
			if inflight[i].doneAt <= step {
				complete(i)
				continue
			}
			i++
		}

		size := 500 + rng.Intn(60000)
		name := "unit.c"
		compiler := "gcc"
		if step%3 == 0 {
			name, compiler = "unit.cpp", "g++"
		}
		ctx := TaskContext{
			TaskID:             fmt.Sprintf("task-%04d", step),
			SourceSizeBytes:    size,
			RawSourceSizeBytes: size / 10,
			SourceFilename:     name,
			Compiler:           compiler,
		}
		lag := 1 + rng.Intn(6)
		noise := 1 + 0.1*rng.NormFloat64()

		var w *registry.WorkerInfo
		var info DispatchInfo
		for {
			var err error
			w, info, err = s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
			if err == nil {
				break
			}
			if !errors.Is(err, ErrNoMatchingWorkers) || len(inflight) == 0 {
				t.Fatalf("step %d: select: %v", step, err)
			}
			complete(0) // every worker is full: free the oldest task and retry
		}
		if err := reg.IncrementTasks(w.ID); err != nil {
			t.Fatalf("IncrementTasks: %v", err)
		}

		ms := (200 + float64(size)/100) / speed[w.ID] * noise
		res.Selections = append(res.Selections, w.ID)
		res.QValues = append(res.QValues, info.QValueAtDispatch)
		res.Exploration = append(res.Exploration, info.WasExploration)
		inflight = append(inflight, &traceTask{
			ctx:    ctx,
			worker: w.ID,
			doneAt: step + lag,
			reward: -math.Log1p(ms) / math.Log1p(120000),
			dur:    time.Duration(ms) * time.Millisecond,
		})
	}
	for len(inflight) > 0 {
		complete(0)
	}

	res.Dispatches = atomic.LoadInt64(&s.totalDispatches)
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, arm := range s.arms {
		res.Arms[id] = dumpArm(arm)
	}
	return res
}

func dumpArm(arm *linUCBArm) *traceArm {
	d, _ := arm.A.Dims()
	out := &traceArm{Count: arm.count, B: make([]float64, d)}
	out.A = make([][]float64, d)
	out.Ainv = make([][]float64, d)
	for i := 0; i < d; i++ {
		out.A[i] = make([]float64, d)
		out.Ainv[i] = make([]float64, d)
		for j := 0; j < d; j++ {
			out.A[i][j] = arm.A.At(i, j)
			out.Ainv[i][j] = arm.Ainv.At(i, j)
		}
		out.B[i] = arm.b.AtVec(i)
	}
	return out
}

// sameTrace reports whether two traces are identical bit for bit.
func sameTrace(a, b traceResult) bool {
	return reflect.DeepEqual(a, b)
}

// closeTo reports whether a and b agree to a relative tolerance (absolute
// near zero).
func closeTo(a, b float64) bool {
	return math.Abs(a-b) <= traceTolerance*math.Max(1, math.Max(math.Abs(a), math.Abs(b)))
}

// compareTraces returns a description of the first difference, or "".
// Selections, exploration flags, counts and the dispatch counter must match
// exactly; floating-point values within traceTolerance.
func compareTraces(got, want traceResult) string {
	if !reflect.DeepEqual(got.Selections, want.Selections) {
		for i := range want.Selections {
			if i >= len(got.Selections) || got.Selections[i] != want.Selections[i] {
				return fmt.Sprintf("selection %d differs", i)
			}
		}
		return "selection count differs"
	}
	if !reflect.DeepEqual(got.Exploration, want.Exploration) {
		return "exploration flags differ"
	}
	if got.Dispatches != want.Dispatches {
		return fmt.Sprintf("dispatches %d != %d", got.Dispatches, want.Dispatches)
	}
	for i := range want.QValues {
		if !closeTo(got.QValues[i], want.QValues[i]) {
			return fmt.Sprintf("q value %d: %g != %g", i, got.QValues[i], want.QValues[i])
		}
	}
	if len(got.Arms) != len(want.Arms) {
		return "arm set differs"
	}
	for id, w := range want.Arms {
		g, ok := got.Arms[id]
		if !ok {
			return "missing arm " + id
		}
		if g.Count != w.Count {
			return fmt.Sprintf("arm %s count %d != %d", id, g.Count, w.Count)
		}
		for i := range w.B {
			if !closeTo(g.B[i], w.B[i]) {
				return fmt.Sprintf("arm %s b[%d]: %g != %g", id, i, g.B[i], w.B[i])
			}
			for j := range w.A[i] {
				if !closeTo(g.A[i][j], w.A[i][j]) || !closeTo(g.Ainv[i][j], w.Ainv[i][j]) {
					return fmt.Sprintf("arm %s A/Ainv[%d][%d] differ", id, i, j)
				}
			}
		}
	}
	return ""
}

func traceOnce(t *testing.T) traceResult {
	t.Helper()
	reg := newTraceCluster(t)
	return runTrace(t, newTraceScheduler(reg), reg)
}

func TestLinUCBTrace_RepeatableOnOneMachine(t *testing.T) {
	a, b := traceOnce(t), traceOnce(t)
	if !sameTrace(a, b) {
		t.Fatalf("two runs of the same trace differ: %s", compareTraces(b, a))
	}
	if len(a.Selections) != traceSteps {
		t.Fatalf("got %d selections, want %d", len(a.Selections), traceSteps)
	}
	if len(a.Arms) != 5 {
		t.Fatalf("got %d arms, want 5 (every worker should have been tried)", len(a.Arms))
	}
}

func TestLinUCBTrace_Golden(t *testing.T) {
	got := traceOnce(t)

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		data, err := json.MarshalIndent(got, "", " ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(traceGoldenPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(traceGoldenPath, append(data, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("wrote %s", traceGoldenPath)
		return
	}

	data, err := os.ReadFile(traceGoldenPath)
	if err != nil {
		t.Fatalf("read golden: %v (generate with UPDATE_GOLDEN=1)", err)
	}
	var want traceResult
	if err := json.Unmarshal(data, &want); err != nil {
		t.Fatalf("parse golden: %v", err)
	}
	if diff := compareTraces(got, want); diff != "" {
		t.Fatalf("behaviour changed vs the golden trace: %s", diff)
	}
}
