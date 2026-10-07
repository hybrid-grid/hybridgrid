package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"gonum.org/v1/gonum/mat"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

const discountGamma = 0.8

// newDiscountTraceScheduler keeps the P0 configuration and changes only the
// two knobs introduced by the API contract.
func newDiscountTraceScheduler(reg *registry.InMemoryRegistry, gamma float64, mode DiscountMode) *LinUCBScheduler {
	return NewLinUCBScheduler(LinUCBConfig{
		Registry: sortedRegistry{reg}, Alpha: 0.5, WarmStartTasks: 100,
		LoadPenalty: 0.5, Discount: gamma, DiscountMode: mode,
	})
}

func newDiscountScheduler(t *testing.T, workers int, gamma float64, mode DiscountMode) *LinUCBScheduler {
	t.Helper()
	return NewLinUCBScheduler(LinUCBConfig{
		Registry: sortedRegistry{newRegistryWithWorkers(t, workers)},
		Alpha:    0.5, Discount: gamma, DiscountMode: mode,
	})
}

func discountSelect(t *testing.T, s *LinUCBScheduler, id string, size int) string {
	t.Helper()
	w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{TaskID: id, SourceSizeBytes: size})
	if err != nil {
		t.Fatalf("select %s: %v", id, err)
	}
	return w.ID
}

func discountPending(t *testing.T, s *LinUCBScheduler, id string) *mat.VecDense {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	x := s.pendingX[id]
	if x == nil {
		t.Fatalf("missing pending x for %s", id)
	}
	return mat.VecDenseCopyOf(x)
}

func discountArm(t *testing.T, s *LinUCBScheduler, id string) *traceArm {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	a := s.arms[id]
	if a == nil {
		t.Fatalf("missing arm %s", id)
	}
	return dumpArm(a)
}

func discountMax(A [][]float64) float64 {
	var m float64
	for _, row := range A {
		for _, v := range row {
			m = math.Max(m, math.Abs(v))
		}
	}
	return math.Max(1, m)
}

func discountAssertState(t *testing.T, got *traceArm, wantA [][]float64, wantB []float64) {
	t.Helper()
	tol := 1e-10 * discountMax(wantA)
	for i := range wantB {
		if math.Abs(got.B[i]-wantB[i]) > tol || math.IsNaN(got.B[i]) {
			t.Errorf("b[%d] = %.17g, want %.17g (tol %.3g)", i, got.B[i], wantB[i], tol)
		}
		for j := range wantA[i] {
			if math.Abs(got.A[i][j]-wantA[i][j]) > tol || math.IsNaN(got.A[i][j]) {
				t.Errorf("A[%d,%d] = %.17g, want %.17g (tol %.3g)", i, j, got.A[i][j], wantA[i][j], tol)
			}
		}
	}
}

func discountIdentity(d int) ([][]float64, []float64) {
	A := make([][]float64, d)
	for i := range A {
		A[i] = make([]float64, d)
		A[i][i] = 1
	}
	return A, make([]float64, d)
}

type discountObservation struct {
	step   int64
	x      []float64
	reward float64
}

func discountCoordinates(x *mat.VecDense) []float64 {
	d := x.Len()
	out := make([]float64, d)
	for i := range out {
		out[i] = x.AtVec(i)
	}
	return out
}

// This reference is the D4 weighted least-squares sum, using only float64
// arithmetic over recorded observations. It does not call scheduler helpers.
func discountClosedForm(d int, gamma float64, now int64, obs []discountObservation) ([][]float64, []float64) {
	A, b := discountIdentity(d)
	for _, o := range obs {
		lag := now - o.step
		if lag < 0 {
			lag = 0
		}
		w := math.Pow(gamma, float64(lag))
		for i := 0; i < d; i++ {
			b[i] += w * o.reward * o.x[i]
			for j := 0; j < d; j++ {
				A[i][j] += w * o.x[i] * o.x[j]
			}
		}
	}
	return A, b
}

// This reference applies the advisor's decay once per decision and adds an
// observation at its arrival with decision-time lag weight. It is independent
// of the closed-form sum and of the production lazy catch-up implementation.
func discountEagerGlobal(d int, gamma float64, arrivals map[int64][]discountObservation, final int64) ([][]float64, []float64) {
	A, b := discountIdentity(d)
	for step := int64(1); step <= final; step++ {
		for i := 0; i < d; i++ {
			b[i] *= gamma
			for j := 0; j < d; j++ {
				A[i][j] *= gamma
				if i == j {
					A[i][j] += 1 - gamma
				}
			}
		}
		for _, o := range arrivals[step] {
			w := math.Pow(gamma, float64(step-o.step))
			for i := 0; i < d; i++ {
				b[i] += w * o.reward * o.x[i]
				for j := 0; j < d; j++ {
					A[i][j] += w * o.x[i] * o.x[j]
				}
			}
		}
	}
	return A, b
}

func discountEagerArm(d int, gamma float64, obs []discountObservation) ([][]float64, []float64) {
	A, b := discountIdentity(d)
	for _, o := range obs {
		for i := 0; i < d; i++ {
			b[i] = gamma*b[i] + o.reward*o.x[i]
			for j := 0; j < d; j++ {
				A[i][j] = gamma*A[i][j] + o.x[i]*o.x[j]
				if i == j {
					A[i][j] += 1 - gamma
				}
			}
		}
	}
	return A, b
}

func discountSameBits(a, b traceResult) bool {
	if !sameTrace(a, b) {
		return false
	}
	for i := range a.QValues {
		if math.Float64bits(a.QValues[i]) != math.Float64bits(b.QValues[i]) {
			return false
		}
	}
	for id, x := range a.Arms {
		y := b.Arms[id]
		for i := range x.B {
			if math.Float64bits(x.B[i]) != math.Float64bits(y.B[i]) {
				return false
			}
			for j := range x.A[i] {
				if math.Float64bits(x.A[i][j]) != math.Float64bits(y.A[i][j]) || math.Float64bits(x.Ainv[i][j]) != math.Float64bits(y.Ainv[i][j]) {
					return false
				}
			}
		}
	}
	return true
}

// T1 oracle: all off configurations are bit-identical on one machine; the
// committed pre-change golden permits only the documented 1e-12 tolerance.
func TestLinUCBDiscount_GoldenOff(t *testing.T) {
	data, err := os.ReadFile(traceGoldenPath)
	if err != nil {
		t.Fatal(err)
	}
	var golden traceResult
	if err := json.Unmarshal(data, &golden); err != nil {
		t.Fatal(err)
	}
	if len(golden.Selections) != traceSteps {
		t.Fatalf("golden has %d selections, want %d", len(golden.Selections), traceSteps)
	}
	reg := newTraceCluster(t)
	baseline := runTrace(t, newTraceScheduler(reg), reg)
	if diff := compareTraces(baseline, golden); diff != "" {
		t.Fatalf("unset Discount vs golden: %s", diff)
	}
	for _, mode := range []DiscountMode{"", DiscountModeGlobal, DiscountModeArm} {
		for _, gamma := range []float64{0, 1} {
			name := fmt.Sprintf("mode=%q/gamma=%g", mode, gamma)
			t.Run(name, func(t *testing.T) {
				reg := newTraceCluster(t)
				got := runTrace(t, newDiscountTraceScheduler(reg, gamma, mode), reg)
				if !discountSameBits(got, baseline) {
					t.Fatalf("off path differs bit for bit from unset Discount: %s", compareTraces(got, baseline))
				}
				if diff := compareTraces(got, golden); diff != "" {
					t.Fatalf("golden: %s", diff)
				}
			})
		}
	}
}

// T2 oracle: D4's closed form includes three different feedback lags and
// reversed completion order; the cached x is the value at each dispatch.
func TestLinUCBDiscount_GlobalClosedFormLag(t *testing.T) {
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeGlobal)
	ids := []string{"lag-1", "lag-2", "lag-3"}
	obs := make([]discountObservation, len(ids))
	for i, id := range ids {
		discountSelect(t, s, id, 100+i*1300)
		obs[i] = discountObservation{step: int64(i + 1), x: discountCoordinates(discountPending(t, s, id)), reward: -float64(i + 1)}
	}
	discountSelect(t, s, "clock-4", 400)
	for _, i := range []int{2, 0, 1} {
		s.RecordOutcome("worker-a", obs[i].reward, true, TaskContext{TaskID: ids[i]})
	}
	discountSelect(t, s, "clock-5", 500)
	s.score("worker-a", mat.NewVecDense(s.dim, nil))
	got := discountArm(t, s, "worker-a")
	wantA, wantB := discountClosedForm(s.dim, discountGamma, 5, obs)
	discountAssertState(t, got, wantA, wantB)
	if got.Count != 3 {
		t.Fatalf("count = %d, want 3", got.Count)
	}
	// Legacy white-box callers supply pendingX alone. The missing step
	// means this observation has weight one at the current decision.
	legacy := "legacy-no-step"
	legacyX := mat.NewVecDense(s.dim, nil)
	legacyX.SetVec(0, 1)
	s.mu.Lock()
	s.pendingX[legacy] = legacyX
	s.mu.Unlock()
	s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: legacy})
	obs = append(obs, discountObservation{step: 5, x: discountCoordinates(legacyX), reward: -1})
	wantA, wantB = discountClosedForm(s.dim, discountGamma, 5, obs)
	discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
}

// T3b oracle: an arm whose recorded step is ahead of the dispatch clock
// neither ages backward on score nor uses the stale clock for feedback weight.
func TestLinUCBDiscount_GlobalLastStepNeverMovesBackward(t *testing.T) {
	const gamma = 0.9
	s := newDiscountScheduler(t, 2, gamma, DiscountModeGlobal)
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("history-%d", i)
		discountSelect(t, s, id, 100+i*1000)
		s.RecordOutcome("worker-a", -float64(i+1), true, TaskContext{TaskID: id})
	}
	id := "future-last-step"
	discountSelect(t, s, id, 750)
	x := discountPending(t, s, id)
	s.score("worker-a", x) // materialize theta before the unchanged-state check
	s.mu.Lock()
	clock := atomic.LoadInt64(&s.totalDispatches)
	arm := s.arms["worker-a"]
	arm.lastStep = clock + 5
	lastStep := arm.lastStep
	s0 := s.pendingStep[id]
	beforeA := mat.DenseCopyOf(arm.A)
	beforeAinv := mat.DenseCopyOf(arm.Ainv)
	beforeB := mat.VecDenseCopyOf(arm.b)
	beforeTheta := mat.VecDenseCopyOf(arm.theta)
	beforeCount, beforeDirty := arm.count, arm.dirty
	s.mu.Unlock()
	if s0 >= lastStep {
		t.Fatalf("pending step %d must precede arm lastStep %d", s0, lastStep)
	}

	s.score("worker-a", x)
	s.mu.Lock()
	arm = s.arms["worker-a"]
	unchanged := arm.lastStep == lastStep && arm.count == beforeCount && arm.dirty == beforeDirty &&
		mat.Equal(arm.A, beforeA) && mat.Equal(arm.Ainv, beforeAinv) &&
		mat.Equal(arm.b, beforeB) && mat.Equal(arm.theta, beforeTheta)
	s.mu.Unlock()
	if !unchanged {
		t.Fatal("score moved an arm backward or changed its state while clock < lastStep")
	}

	const reward = -1.25
	s.RecordOutcome("worker-a", reward, true, TaskContext{TaskID: id})
	wantA := make([][]float64, s.dim)
	wantB := make([]float64, s.dim)
	w := math.Pow(gamma, float64(lastStep-s0))
	for i := 0; i < s.dim; i++ {
		wantA[i] = make([]float64, s.dim)
		xi := x.AtVec(i)
		wantB[i] = beforeB.AtVec(i) + w*reward*xi
		for j := 0; j < s.dim; j++ {
			wantA[i][j] = beforeA.At(i, j) + w*xi*x.AtVec(j)
		}
	}
	discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
	s.mu.Lock()
	arm = s.arms["worker-a"]
	gotStep, gotCount := arm.lastStep, arm.count
	s.mu.Unlock()
	if gotStep != lastStep || gotCount != beforeCount+1 {
		t.Fatalf("outcome moved arm backward or changed count incorrectly: lastStep=%d want=%d count=%d want=%d", gotStep, lastStep, gotCount, beforeCount+1)
	}
}

// T3c oracle: the outcome reads the dispatch clock after acquiring s.mu.
// The other worker's single-candidate Select advances that clock without s.mu.
func TestLinUCBDiscount_GlobalClockReadInsideLock(t *testing.T) {
	const gamma = 0.9
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: sortedRegistry{reg}, Discount: gamma, DiscountMode: DiscountModeGlobal})
	x := mat.NewVecDense(s.dim, nil)
	x.SetVec(0, 1)
	x.SetVec(1, 0.5)
	x.SetVec(2, -0.25)
	const reward = -2.0
	s.mu.Lock()
	s.pendingX["t1"] = mat.VecDenseCopyOf(x)
	s.pendingStep["t1"] = 1
	atomic.StoreInt64(&s.totalDispatches, 1)
	s.armForLocked("worker-a").lastStep = 1
	s.mu.Unlock()
	if err := reg.Remove("worker-a"); err != nil {
		t.Fatal(err)
	}

	s.mu.Lock()
	done := make(chan struct{})
	go func() {
		defer close(done)
		s.RecordOutcome("worker-a", reward, true, TaskContext{TaskID: "t1"})
	}()
	time.Sleep(100 * time.Millisecond)
	w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	clock := atomic.LoadInt64(&s.totalDispatches)
	s.mu.Unlock()
	if err != nil || w == nil || w.ID != "worker-b" || clock != 2 {
		t.Fatalf("single-candidate Select: worker=%v err=%v clock=%d, want worker-b and clock 2", w, err, clock)
	}
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("RecordOutcome did not finish within 5 seconds after releasing s.mu")
	}
	wantA, wantB := discountClosedForm(s.dim, gamma, 2, []discountObservation{{step: 1, x: discountCoordinates(x), reward: reward}})
	discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
}

// T3 oracle: an eager advisor recurrence after each decision agrees with
// lazy arm state, including feedback arriving out of dispatch order.
func TestLinUCBDiscount_GlobalEagerReference(t *testing.T) {
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeGlobal)
	arrivals := make(map[int64][]discountObservation)
	var saved [3]discountObservation
	for step := int64(1); step <= 7; step++ {
		id := fmt.Sprintf("eager-%d", step)
		discountSelect(t, s, id, int(step*900))
		if step <= 3 {
			saved[step-1] = discountObservation{step: step, x: discountCoordinates(discountPending(t, s, id)), reward: -float64(step)}
		}
		for _, index := range map[int64][]int{4: {1}, 5: {2}, 6: {0}}[step] {
			o := saved[index]
			arrivals[step] = append(arrivals[step], o)
			s.RecordOutcome("worker-a", o.reward, true, TaskContext{TaskID: fmt.Sprintf("eager-%d", o.step)})
		}
		s.score("worker-a", mat.NewVecDense(s.dim, nil))
		wantA, wantB := discountEagerGlobal(s.dim, discountGamma, arrivals, step)
		discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
	}
}

// T4 oracle: an idle arm ages by exactly k decisions, and repeated reads
// at the same clock are idempotent. Outcomes alone cannot move the clock.
func TestLinUCBDiscount_GlobalIdleCatchUp(t *testing.T) {
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeGlobal)
	id := "initial"
	discountSelect(t, s, id, 1000)
	x := discountPending(t, s, id)
	s.RecordOutcome("worker-a", -2, true, TaskContext{TaskID: id})
	before := discountArm(t, s, "worker-a")
	clock := atomic.LoadInt64(&s.totalDispatches)
	s.RecordOutcome("worker-a", -3, true, TaskContext{TaskID: "unknown"})
	if atomic.LoadInt64(&s.totalDispatches) != clock {
		t.Fatal("RecordOutcome moved decision clock")
	}
	for i := 0; i < 9; i++ {
		discountSelect(t, s, fmt.Sprintf("idle-%d", i), i+1)
	}
	s.score("worker-a", x)
	after := discountArm(t, s, "worker-a")
	wantA, wantB := discountIdentity(s.dim)
	g := math.Pow(discountGamma, 9)
	for i := 0; i < s.dim; i++ {
		wantB[i] = g * before.B[i]
		for j := 0; j < s.dim; j++ {
			base := 0.0
			if i == j {
				base = 1
			}
			wantA[i][j] = base + g*(before.A[i][j]-base)
		}
	}
	discountAssertState(t, after, wantA, wantB)
	s.mu.Lock()
	lastStep := s.arms["worker-a"].lastStep
	s.mu.Unlock()
	if lastStep != atomic.LoadInt64(&s.totalDispatches) {
		t.Fatalf("global arm aged through step %d, clock is %d", lastStep, atomic.LoadInt64(&s.totalDispatches))
	}
	s.score("worker-a", x)
	if !reflect.DeepEqual(after, discountArm(t, s, "worker-a")) {
		t.Fatal("second score at the same step changed the arm")
	}
	if after.Count != before.Count {
		t.Fatal("catch-up changed observation count")
	}
	assertDiscountInvariants(t, s, "worker-a")
}

func assertDiscountInvariants(t *testing.T, s *LinUCBScheduler, id string) {
	t.Helper()
	s.score(id, mat.NewVecDense(s.dim, nil))
	s.mu.Lock()
	a := s.arms[id]
	A, inv := mat.DenseCopyOf(a.A), mat.DenseCopyOf(a.Ainv)
	b, theta := mat.VecDenseCopyOf(a.b), mat.VecDenseCopyOf(a.theta)
	s.mu.Unlock()
	d := s.dim
	maxA := 1.0
	for i := 0; i < d; i++ {
		if math.IsNaN(b.AtVec(i)) || math.IsInf(b.AtVec(i), 0) || math.IsNaN(theta.AtVec(i)) || math.IsInf(theta.AtVec(i), 0) {
			t.Fatalf("non-finite b/theta for %s", id)
		}
		for j := 0; j < d; j++ {
			v := A.At(i, j)
			if math.IsNaN(v) || math.IsInf(v, 0) || math.IsNaN(inv.At(i, j)) || math.IsInf(inv.At(i, j), 0) {
				t.Fatalf("non-finite A/Ainv for %s", id)
			}
			maxA = math.Max(maxA, math.Abs(v))
			if math.Abs(v-A.At(j, i)) > 1e-9*maxA {
				t.Errorf("A asymmetric for %s at %d,%d", id, i, j)
			}
		}
	}
	var eig mat.EigenSym
	sym := mat.NewSymDense(d, nil)
	for i := 0; i < d; i++ {
		for j := 0; j <= i; j++ {
			sym.SetSym(i, j, (A.At(i, j)+A.At(j, i))/2)
		}
	}
	if !eig.Factorize(sym, false) {
		t.Fatalf("eigendecomposition failed for %s", id)
	}
	for _, v := range eig.Values(nil) {
		if v < 1-1e-9 {
			t.Errorf("A eigenvalue %.17g < 1 for %s", v, id)
		}
	}
	for i := 0; i < d; i++ {
		var wantTheta float64
		for j := 0; j < d; j++ {
			wantTheta += inv.At(i, j) * b.AtVec(j)
		}
		if math.Abs(theta.AtVec(i)-wantTheta) > 1e-9*maxA {
			t.Errorf("theta != Ainv*b for %s at %d", id, i)
		}
		for j := 0; j < d; j++ {
			var product float64
			for k := 0; k < d; k++ {
				product += A.At(i, k) * inv.At(k, j)
			}
			want := 0.0
			if i == j {
				want = 1
			}
			if math.Abs(product-want) > 1e-9*maxA {
				t.Errorf("A*Ainv[%d,%d] = %.17g for %s", i, j, product, id)
			}
		}
	}
}

// T5 oracle: 2000 seeded dispatches preserve SPD, inverse, theta and
// finiteness at each supported discount scale.
func TestLinUCBDiscount_GlobalRandomInvariants(t *testing.T) {
	for _, gamma := range []float64{0.5, 0.95, 0.99} {
		t.Run(fmt.Sprintf("gamma=%g", gamma), func(t *testing.T) {
			s := newDiscountScheduler(t, 5, gamma, DiscountModeGlobal)
			rng := rand.New(rand.NewSource(73))
			for i := 0; i < 2000; i++ {
				id := fmt.Sprintf("random-%d", i)
				w := discountSelect(t, s, id, rng.Intn(4<<20)+1)
				s.RecordOutcome(w, -rng.Float64(), true, TaskContext{TaskID: id})
			}
			for i := 0; i < 5; i++ {
				assertDiscountInvariants(t, s, idForTest(i))
			}
		})
	}
}

// T6 oracle: invalid inputs are dropped without consumption or mutation;
// pure helpers leave inputs intact and zero underflow returns exact I, 0.
func TestLinUCBDiscount_GlobalGuardsAndPureHelpers(t *testing.T) {
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeGlobal)
	for i := 0; i < 12; i++ {
		id := fmt.Sprintf("guard-%d", i)
		w := discountSelect(t, s, id, i+100)
		s.mu.Lock()
		if len(s.pendingX) != len(s.pendingStep) {
			t.Fatal("pending maps diverged after Select")
		}
		s.mu.Unlock()
		if i%2 == 0 {
			s.RecordOutcome(w, -1, true, TaskContext{TaskID: id})
		}
		s.mu.Lock()
		if len(s.pendingX) != len(s.pendingStep) {
			t.Fatal("pending maps diverged after RecordOutcome")
		}
		s.mu.Unlock()
	}
	before := discountArm(t, s, "worker-a")
	clock := atomic.LoadInt64(&s.totalDispatches)
	for _, reward := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		s.RecordOutcome("worker-a", reward, true, TaskContext{TaskID: "guard-1"})
	}
	s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: "not-dispatched"})
	if !reflect.DeepEqual(before, discountArm(t, s, "worker-a")) || atomic.LoadInt64(&s.totalDispatches) != clock {
		t.Fatal("invalid/unknown outcome changed arm or clock")
	}
	s.mu.Lock()
	_, xStill := s.pendingX["guard-1"]
	_, stepStill := s.pendingStep["guard-1"]
	balanced := len(s.pendingX) == len(s.pendingStep)
	s.mu.Unlock()
	if !xStill || !stepStill || !balanced {
		t.Fatal("invalid reward consumed pending state or maps diverged")
	}
	A := mat.NewDense(2, 2, []float64{2, 0.25, 0.25, 3})
	b := mat.NewVecDense(2, []float64{4, -2})
	x := mat.NewVecDense(2, []float64{1, 2})
	A0, b0, x0 := mat.DenseCopyOf(A), mat.VecDenseCopyOf(b), mat.VecDenseCopyOf(x)
	for _, tc := range []struct{ reward, weight float64 }{{math.NaN(), 1}, {math.Inf(1), 1}, {1, math.NaN()}, {1, math.Inf(1)}} {
		_, _, _, ok := discountUpdate(A, b, x, tc.reward, tc.weight)
		if ok {
			t.Errorf("discountUpdate accepted reward=%g weight=%g", tc.reward, tc.weight)
		}
	}
	badX := mat.NewVecDense(2, []float64{math.Inf(1), 1})
	if _, _, _, ok := discountUpdate(A, b, badX, 1, 1); ok {
		t.Error("discountUpdate accepted non-finite x")
	}
	badA := mat.NewDense(2, 2, []float64{math.NaN(), 0, 0, 1})
	if _, _, _, ok := discountUpdate(badA, b, x, 1, 1); ok {
		t.Error("discountUpdate accepted non-finite A")
	}
	badB := mat.NewVecDense(2, []float64{math.Inf(-1), 0})
	if _, _, _, ok := discountUpdate(A, badB, x, 1, 1); ok {
		t.Error("discountUpdate accepted non-finite b")
	}
	if _, _, _, ok := discountUpdate(mat.NewDense(2, 2, []float64{-10, 0, 0, 1}), b, x, 1, 1); ok {
		t.Error("discountUpdate accepted non-SPD candidate")
	}
	if !mat.Equal(A, A0) || !mat.Equal(b, b0) || !mat.Equal(x, x0) {
		t.Fatal("discountUpdate mutated input")
	}
	validA, validB, validInv, ok := discountUpdate(A, b, x, -2, 0.25)
	if !ok {
		t.Fatal("discountUpdate rejected a finite SPD candidate")
	}
	for i := 0; i < 2; i++ {
		if math.Abs(validB.AtVec(i)-(b.AtVec(i)-0.5*x.AtVec(i))) > 1e-12 {
			t.Fatal("discountUpdate b formula differs from weighted reward")
		}
		for j := 0; j < 2; j++ {
			want := A.At(i, j) + 0.25*x.AtVec(i)*x.AtVec(j)
			if math.Abs(validA.At(i, j)-want) > 1e-12 {
				t.Fatal("discountUpdate A formula differs from weighted outer product")
			}
			var product float64
			for k := 0; k < 2; k++ {
				product += validA.At(i, k) * validInv.At(k, j)
			}
			identity := 0.0
			if i == j {
				identity = 1
			}
			if math.Abs(product-identity) > 1e-12 || validInv.At(i, j) != validInv.At(j, i) {
				t.Fatal("discountUpdate inverse is wrong or asymmetric")
			}
		}
	}
	bad := "nonfinite-global-x"
	s.mu.Lock()
	s.pendingX[bad] = mat.NewVecDense(s.dim, nil)
	s.pendingX[bad].SetVec(0, math.Inf(1))
	s.pendingStep[bad] = clock
	s.mu.Unlock()
	s.RecordOutcome("worker-a", 1, true, TaskContext{TaskID: bad})
	if !reflect.DeepEqual(before, discountArm(t, s, "worker-a")) {
		t.Fatal("global non-finite candidate partially committed")
	}
	s.mu.Lock()
	balanced = len(s.pendingX) == len(s.pendingStep)
	s.mu.Unlock()
	if !balanced {
		t.Fatal("failed update did not consume both pending maps")
	}
	decayedA, decayedB := discountDecay(A, b, 0.5, 100000)
	for i := 0; i < 2; i++ {
		if decayedB.AtVec(i) != 0 {
			t.Fatal("underflow did not produce exact zero b")
		}
		for j := 0; j < 2; j++ {
			want := 0.0
			if i == j {
				want = 1
			}
			if decayedA.At(i, j) != want {
				t.Fatal("underflow did not produce exact identity A")
			}
		}
	}
	if !mat.Equal(A, A0) || !mat.Equal(b, b0) {
		t.Fatal("discountDecay mutated input")
	}
	finiteA, finiteB := discountDecay(A, b, discountGamma, 3)
	g := math.Pow(discountGamma, 3)
	for i := 0; i < 2; i++ {
		if math.Abs(finiteB.AtVec(i)-g*b.AtVec(i)) > 1e-12 {
			t.Fatal("discountDecay b formula differs from advisor recurrence")
		}
		for j := 0; j < 2; j++ {
			identity := 0.0
			if i == j {
				identity = 1
			}
			if math.Abs(finiteA.At(i, j)-(identity+g*(A.At(i, j)-identity))) > 1e-12 {
				t.Fatal("discountDecay A formula differs from advisor recurrence")
			}
		}
	}
}

// T7 oracle: only candidate-bearing Select calls tick; fast-path outcomes
// have no cache; the second dispatch of one TaskID replaces x and step.
func TestLinUCBDiscount_FastPathOverwriteAndFailedSelect(t *testing.T) {
	single := newDiscountScheduler(t, 1, discountGamma, DiscountModeGlobal)
	id := TaskContext{TaskID: "fast", SourceSizeBytes: 10}
	w, _, err := single.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", id)
	if err != nil {
		t.Fatal(err)
	}
	if atomic.LoadInt64(&single.totalDispatches) != 1 || len(single.pendingX) != 0 || len(single.pendingStep) != 0 {
		t.Fatal("fast path clock/cache mismatch")
	}
	single.RecordOutcome(w.ID, -1, true, id)
	if len(single.arms) != 0 {
		t.Fatal("fast-path outcome created an arm")
	}
	empty := newDiscountScheduler(t, 0, discountGamma, DiscountModeGlobal)
	_, _, err = empty.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", id)
	if !errors.Is(err, ErrNoWorkers) || atomic.LoadInt64(&empty.totalDispatches) != 0 {
		t.Fatalf("empty Select: err=%v clock=%d", err, atomic.LoadInt64(&empty.totalDispatches))
	}
	_, _, err = single.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_ARM64, "", id)
	if !errors.Is(err, ErrNoMatchingWorkers) || atomic.LoadInt64(&single.totalDispatches) != 1 {
		t.Fatalf("unmatched Select: err=%v clock=%d", err, atomic.LoadInt64(&single.totalDispatches))
	}
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeGlobal)
	discountSelect(t, s, "same", 10)
	first := discountPending(t, s, "same")
	s.mu.Lock()
	firstStep := s.pendingStep["same"]
	s.mu.Unlock()
	discountSelect(t, s, "same", 1000000)
	latest := discountPending(t, s, "same")
	s.mu.Lock()
	latestStep := s.pendingStep["same"]
	s.mu.Unlock()
	if mat.Equal(first, latest) || firstStep != 1 || latestStep != 2 {
		t.Fatal("re-selection did not replace both caches")
	}
	discountSelect(t, s, "later", 20)
	s.RecordOutcome("worker-a", -2, true, TaskContext{TaskID: "same"})
	got := discountArm(t, s, "worker-a")
	A, b := discountClosedForm(s.dim, discountGamma, 3, []discountObservation{{step: 2, x: discountCoordinates(latest), reward: -2}})
	discountAssertState(t, got, A, b)
	if len(s.pendingX) != len(s.pendingStep) {
		t.Fatal("pending maps diverged after overwrite consumption")
	}
}

// D12 special-path oracle: re-registering an ID retains its arm, and
// candidate-bearing dispatches made while absent still age it globally.
func TestLinUCBDiscount_ReregistrationKeepsArm(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: sortedRegistry{reg}, Discount: discountGamma, DiscountMode: DiscountModeGlobal})
	id := "before-reregistration"
	discountSelect(t, s, id, 100)
	x := discountPending(t, s, id)
	s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: id})
	if err := reg.Remove("worker-a"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 7; i++ {
		discountSelect(t, s, "", i+1)
	}
	if err := reg.Add(newHybridTestWorker("worker-a", 4, 4, 0)); err != nil {
		t.Fatal(err)
	}
	s.score("worker-a", x)
	got := discountArm(t, s, "worker-a")
	A, b := discountClosedForm(s.dim, discountGamma, 8, []discountObservation{{step: 1, x: discountCoordinates(x), reward: -1}})
	discountAssertState(t, got, A, b)
	if got.Count != 1 {
		t.Fatalf("re-registration reset count to %d", got.Count)
	}
}

// T8 oracle: the delegated worker's feature and dispatch step are cached,
// its reward updates the arm, and the next warm-start score ages that arm.
func TestLinUCBDiscount_WarmStartLearnsAndCatchesUp(t *testing.T) {
	reg := registry.NewInMemoryRegistry(time.Hour)
	t.Cleanup(reg.Stop)
	for _, tc := range []struct {
		id     string
		active int32
	}{{"busy", 3}, {"idle", 0}, {"half", 2}} {
		if err := reg.Add(newHybridTestWorker(tc.id, 4, 4, tc.active)); err != nil {
			t.Fatal(err)
		}
	}
	s := NewLinUCBScheduler(LinUCBConfig{Registry: sortedRegistry{reg}, Discount: discountGamma, DiscountMode: DiscountModeGlobal, WarmStartTasks: 10})
	ctx := TaskContext{TaskID: "warm-1", SourceSizeBytes: 1000}
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
	if err != nil || w.ID != "idle" || !info.WasExploration {
		t.Fatalf("warm-start selection: worker=%v info=%+v err=%v", w, info, err)
	}
	x := discountPending(t, s, ctx.TaskID)
	s.mu.Lock()
	step := s.pendingStep[ctx.TaskID]
	s.mu.Unlock()
	if step != 1 {
		t.Fatalf("cached step = %d, want 1", step)
	}
	s.RecordOutcome(w.ID, -2, true, ctx)
	before := discountArm(t, s, "idle")
	if before.Count != 1 || before.B[0] == 0 {
		t.Fatal("warm-start outcome did not learn")
	}
	ctx2 := TaskContext{TaskID: "warm-2", SourceSizeBytes: 2000}
	w, _, err = s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx2)
	if err != nil || w.ID != "idle" {
		t.Fatalf("second warm-start: worker=%v err=%v", w, err)
	}
	if discountPending(t, s, ctx2.TaskID) == nil {
		t.Fatal("missing second warm-start cache")
	}
	after := discountArm(t, s, "idle")
	A, b := discountClosedForm(s.dim, discountGamma, 2, []discountObservation{{step: 1, x: discountCoordinates(x), reward: -2}})
	discountAssertState(t, after, A, b)
	if after.Count != 1 {
		t.Fatal("score catch-up changed count")
	}
}

func runDiscountConcurrent(t *testing.T, mode DiscountMode) {
	t.Helper()
	s := newDiscountScheduler(t, 5, discountGamma, mode)
	const goroutines, perGoroutine = 8, 32
	var wg sync.WaitGroup
	var accepted atomic.Int64
	var failures atomic.Int64
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < perGoroutine; i++ {
				id := fmt.Sprintf("race-%d-%d", g, i)
				ctx := TaskContext{TaskID: id, SourceSizeBytes: 1 + g*perGoroutine + i}
				w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
				if err != nil || w == nil {
					failures.Add(1)
					return
				}
				s.RecordOutcome(w.ID, -0.5, true, ctx)
				accepted.Add(1)
			}
		}(g)
	}
	wg.Wait()
	if failures.Load() != 0 {
		t.Fatalf("%d concurrent Select calls failed", failures.Load())
	}
	s.mu.Lock()
	var count int64
	for _, arm := range s.arms {
		count += arm.count
	}
	steps := len(s.pendingStep)
	s.mu.Unlock()
	if count != accepted.Load() || atomic.LoadInt64(&s.totalDispatches) != accepted.Load() {
		t.Fatalf("accepted=%d count=%d clock=%d", accepted.Load(), count, atomic.LoadInt64(&s.totalDispatches))
	}
	if mode == DiscountModeArm && steps != 0 {
		t.Fatal("arm mode used pendingStep")
	}
}

// T9 oracle: every unique concurrent TaskID is consumed once and every
// candidate-bearing Select increments the clock once. Safe under -race.
func TestLinUCBDiscount_GlobalConcurrent(t *testing.T) { runDiscountConcurrent(t, DiscountModeGlobal) }

// T10 oracle: invalid library gamma values silently select the old path.
func TestLinUCBDiscount_InvalidGammaOff(t *testing.T) {
	baselineReg := newTraceCluster(t)
	baseline := runTrace(t, newTraceScheduler(baselineReg), baselineReg)
	for _, gamma := range []float64{-0.1, 1.1, math.NaN(), math.Inf(1)} {
		t.Run(fmt.Sprintf("gamma=%g", gamma), func(t *testing.T) {
			for _, mode := range []DiscountMode{DiscountModeGlobal, DiscountModeArm} {
				reg := newTraceCluster(t)
				got := runTrace(t, newDiscountTraceScheduler(reg, gamma, mode), reg)
				if !sameTrace(got, baseline) {
					t.Fatalf("invalid gamma %g mode %q did not behave as off: %s", gamma, mode, compareTraces(got, baseline))
				}
			}
		})
	}
}

func newDiscountPilot(t *testing.T, gamma float64, mode DiscountMode) *LinUCBScheduler {
	t.Helper()
	reg := registry.NewInMemoryRegistry(time.Hour)
	t.Cleanup(reg.Stop)
	for i := 0; i < 3; i++ {
		w := newHybridTestWorker(fmt.Sprintf("pilot-%d", i), 4, 4, 0)
		w.MaxParallel = 100000
		if err := reg.Add(w); err != nil {
			t.Fatal(err)
		}
	}
	return NewLinUCBScheduler(LinUCBConfig{Registry: sortedRegistry{reg}, Alpha: 0.5, WarmStartTasks: 0, LoadPenalty: 0, Discount: gamma, DiscountMode: mode})
}

// The exact section 5 noise-free pilot: identical worker capabilities,
// log-uniform size, immediate feedback, no ActiveTasks mutation or noise.
func discountPilotShares(t *testing.T, seed int64, gamma float64, mode DiscountMode, recoverArm bool) (drift, recovery float64) {
	t.Helper()
	s := newDiscountPilot(t, gamma, mode)
	rng := rand.New(rand.NewSource(seed))
	means := []float64{-0.30, -0.35, -0.40}
	var driftCount, recoveryCount int
	for decision := 1; decision <= 1900; decision++ {
		size := int(math.Exp(rng.Float64()*math.Log(4<<20))) + 1
		ctx := TaskContext{TaskID: fmt.Sprintf("pilot-task-%d", decision), SourceSizeBytes: size}
		w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		if err != nil {
			t.Fatalf("pilot decision %d: %v", decision, err)
		}
		arm := int(w.ID[len("pilot-")] - '0')
		if decision >= 1501 && decision <= 1650 && arm == 0 {
			driftCount++
		}
		if decision >= 1751 && decision <= 1900 && arm == 0 {
			recoveryCount++
		}
		reward := means[arm] - 0.1*math.Min(1, math.Log1p(float64(size))/sizeNormDenom)
		if arm == 0 && decision > 1500 && (!recoverArm || decision <= 1650) {
			reward -= 0.5
		}
		s.RecordOutcome(w.ID, reward, true, ctx)
	}
	return float64(driftCount) / 150, float64(recoveryCount) / 150
}

// T12 oracle: predeclared permanent-drift thresholds hold for every seed.
func TestLinUCBDiscount_GlobalPermanentDrift(t *testing.T) {
	if testing.Short() {
		t.Skip("predeclared 1900-decision pilot")
	}
	for seed := int64(101); seed <= 105; seed++ {
		off, _ := discountPilotShares(t, seed, 1, DiscountModeGlobal, false)
		fast, _ := discountPilotShares(t, seed, 0.95, DiscountModeGlobal, false)
		t.Logf("seed=%d gamma=1 drift=%.4f gamma=0.95 drift=%.4f", seed, off, fast)
		if off < 0.30 || fast > off-0.25 {
			t.Errorf("seed %d: permanent drift thresholds failed: off=%.4f discounted=%.4f", seed, off, fast)
		}
	}
}

// T13 oracle: global aging returns to the recovered arm at the fixed rate.
func TestLinUCBDiscount_GlobalRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("predeclared 1900-decision pilot")
	}
	for seed := int64(101); seed <= 105; seed++ {
		_, recovery := discountPilotShares(t, seed, 0.98, DiscountModeGlobal, true)
		t.Logf("seed=%d gamma=0.98 recovery=%.4f", seed, recovery)
		if recovery < 0.30 {
			t.Errorf("seed %d: recovery %.4f < 0.30", seed, recovery)
		}
	}
}

// T14 oracle: D12's own-observation closed form and an eager recurrence
// agree; 1000 dispatches without feedback leave the idle arm at I, 0.
func TestLinUCBDiscount_ArmClosedFormAndIdle(t *testing.T) {
	s := newDiscountScheduler(t, 3, discountGamma, DiscountModeArm)
	otherID := "arm-other"
	discountSelect(t, s, otherID, 987)
	s.RecordOutcome("worker-b", -1, true, TaskContext{TaskID: otherID})
	otherBefore := discountArm(t, s, "worker-b")
	var obs []discountObservation
	for i := 0; i < 8; i++ {
		id := fmt.Sprintf("arm-%d", i)
		discountSelect(t, s, id, i*500+1)
		o := discountObservation{x: discountCoordinates(discountPending(t, s, id)), reward: -float64(i+1) / 4}
		obs = append(obs, o)
		s.RecordOutcome("worker-a", o.reward, true, TaskContext{TaskID: id})
		wantA, wantB := discountEagerArm(s.dim, discountGamma, obs)
		discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
	}
	if !reflect.DeepEqual(otherBefore, discountArm(t, s, "worker-b")) {
		t.Fatal("updating one arm decayed another arm")
	}
	closedA, closedB := discountIdentity(s.dim)
	for i, o := range obs {
		w := math.Pow(discountGamma, float64(len(obs)-1-i))
		for j := 0; j < s.dim; j++ {
			closedB[j] += w * o.reward * o.x[j]
			for k := 0; k < s.dim; k++ {
				closedA[j][k] += w * o.x[j] * o.x[k]
			}
		}
	}
	discountAssertState(t, discountArm(t, s, "worker-a"), closedA, closedB)
	for i := 0; i < 1000; i++ {
		discountSelect(t, s, "", i+1)
	}
	// An additional accepted outcome after the idle gap must apply exactly
	// one arm recurrence, regardless of the global decision clock.
	id := "arm-after-idle"
	discountSelect(t, s, id, 777)
	o := discountObservation{x: discountCoordinates(discountPending(t, s, id)), reward: -0.75}
	obs = append(obs, o)
	s.RecordOutcome("worker-a", o.reward, true, TaskContext{TaskID: id})
	wantA, wantB := discountEagerArm(s.dim, discountGamma, obs)
	discountAssertState(t, discountArm(t, s, "worker-a"), wantA, wantB)
	if !reflect.DeepEqual(otherBefore, discountArm(t, s, "worker-b")) {
		t.Fatal("arm mode aged another arm after idle decisions and an outcome")
	}
	s.score("worker-c", mat.NewVecDense(s.dim, nil))
	idleA, idleB := discountIdentity(s.dim)
	discountAssertState(t, discountArm(t, s, "worker-c"), idleA, idleB)
	for _, id := range []string{"worker-a", "worker-b", "worker-c"} {
		assertDiscountInvariants(t, s, id)
	}
}

// T15 oracle: arm mode never stores steps or ages on score; invalid and
// non-finite updates are transactional, while Select still counts decisions.
func TestLinUCBDiscount_ArmGuards(t *testing.T) {
	s := newDiscountScheduler(t, 2, discountGamma, DiscountModeArm)
	id := "arm-guard"
	discountSelect(t, s, id, 100)
	if len(s.pendingStep) != 0 || atomic.LoadInt64(&s.totalDispatches) != 1 {
		t.Fatal("arm mode stored step or failed to count decision")
	}
	s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: id})
	before := discountArm(t, s, "worker-a")
	for i := 0; i < 10; i++ {
		discountSelect(t, s, "", i+1)
	}
	s.score("worker-a", mat.NewVecDense(s.dim, nil))
	if !reflect.DeepEqual(before, discountArm(t, s, "worker-a")) {
		t.Fatal("arm mode aged at score time")
	}
	clock := atomic.LoadInt64(&s.totalDispatches)
	for _, r := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		s.RecordOutcome("worker-a", r, true, TaskContext{TaskID: id})
	}
	s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: "unknown"})
	if !reflect.DeepEqual(before, discountArm(t, s, "worker-a")) || atomic.LoadInt64(&s.totalDispatches) != clock {
		t.Fatal("invalid/unknown arm outcome changed state or clock")
	}
	bad := "nonfinite-x"
	s.mu.Lock()
	s.pendingX[bad] = mat.NewVecDense(s.dim, nil)
	s.pendingX[bad].SetVec(0, math.Inf(1))
	s.mu.Unlock()
	s.RecordOutcome("worker-a", 1, true, TaskContext{TaskID: bad})
	if !reflect.DeepEqual(before, discountArm(t, s, "worker-a")) || len(s.pendingStep) != 0 {
		t.Fatal("non-finite candidate partially committed or arm mode stored a step")
	}
}

// T15 race analogue: unique TaskIDs give exact counts under concurrency.
func TestLinUCBDiscount_ArmConcurrent(t *testing.T) { runDiscountConcurrent(t, DiscountModeArm) }

// T16 scheduler oracle: empty mode acts like global, while arm and global
// produce distinct state after idle decisions; T1 pins both off modes.
func TestLinUCBDiscount_ModeSemantics(t *testing.T) {
	state := func(mode DiscountMode) *traceArm {
		s := newDiscountScheduler(t, 2, discountGamma, mode)
		id := "mode-sample"
		discountSelect(t, s, id, 100)
		s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: id})
		for i := 0; i < 5; i++ {
			discountSelect(t, s, "", 100+i)
		}
		s.score("worker-a", mat.NewVecDense(s.dim, nil))
		return discountArm(t, s, "worker-a")
	}
	defaultState, globalState, armState := state(""), state(DiscountModeGlobal), state(DiscountModeArm)
	if !reflect.DeepEqual(defaultState, globalState) {
		t.Fatal("empty mode did not default to global")
	}
	if reflect.DeepEqual(globalState, armState) {
		t.Fatal("mode flag ignored: states are identical after idle decisions")
	}
}

// T17 oracle: predeclared arm-mode permanent-drift shares, with recovery
// measured only for reporting because the pilot predicts possible failure.
func TestLinUCBDiscount_ArmPilot(t *testing.T) {
	if testing.Short() {
		t.Skip("predeclared 1900-decision pilot")
	}
	for seed := int64(101); seed <= 105; seed++ {
		off, _ := discountPilotShares(t, seed, 1, DiscountModeArm, false)
		fast, _ := discountPilotShares(t, seed, 0.98, DiscountModeArm, false)
		_, recovery := discountPilotShares(t, seed, 0.98, DiscountModeArm, true)
		t.Logf("seed=%d arm off drift=%.4f arm gamma=0.98 drift=%.4f recovery=%.4f (report only)", seed, off, fast, recovery)
		if off < 0.30 || fast > 0.15 {
			t.Errorf("seed %d: arm permanent drift thresholds failed: off=%.4f discounted=%.4f", seed, off, fast)
		}
	}
}

// Mutant coverage (every entry is killed by a failing assertion):
// arm: score-time decay, decay all arms, global clock -> ArmGuards, ArmClosedFormAndIdle.
// arm: omit (1-gamma)I -> ArmClosedFormAndIdle, GlobalRandomInvariants.
// both: ignore mode flag -> ModeSemantics, ArmClosedFormAndIdle.
// global: decay only observed arm -> GlobalIdleCatchUp, GlobalRecovery.
// global: omit (1-gamma)I -> GlobalIdleCatchUp, GlobalRandomInvariants.
// global: decay b but not A, or A but not b -> GlobalIdleCatchUp, GlobalEagerReference.
// global: add xxT before decay -> GlobalClosedFormLag, GlobalEagerReference.
// global: weight 1 instead of gamma^(t-s) -> GlobalClosedFormLag.
// global: arm.lastStep moved backward -> GlobalLastStepNeverMovesBackward.
// global: clock read outside the lock -> GlobalClockReadInsideLock.
// global: completion-time clock -> GlobalClosedFormLag, GlobalEagerReference.
// global: gamma replaced by 1-gamma -> GlobalIdleCatchUp, GlobalClosedFormLag.
// global: skip catch-up in score -> GlobalIdleCatchUp, GlobalEagerReference.
// global: skip catch-up in selectWarmStart -> WarmStartLearnsAndCatchesUp.
// global: Sherman-Morrison in discounted mode -> GlobalClosedFormLag, GlobalRandomInvariants.
// global: discounted path active at gamma=1 -> GoldenOff.
// global: RecordOutcome advances clock -> GlobalIdleCatchUp, GlobalGuardsAndPureHelpers.
// global: non-transactional NaN commit -> GlobalGuardsAndPureHelpers, ArmGuards.

// The stored inverse must match the stored A right after an outcome, before
// any score call can refresh it. Outcomes arrive with a lag, so the update
// weight is below 1: a Sherman-Morrison update of the old inverse (weight 1)
// would be wrong here although it is exact for zero lag.
func TestLinUCBDiscount_AinvMatchesAImmediatelyAfterOutcome(t *testing.T) {
	for _, mode := range []DiscountMode{DiscountModeGlobal, DiscountModeArm} {
		t.Run(string(mode), func(t *testing.T) {
			s := newDiscountScheduler(t, 3, 0.9, mode)
			type pending struct{ id, worker string }
			var queue []pending
			for step := 0; step < 14; step++ {
				id := fmt.Sprintf("ainv-%s-%d", mode, step)
				queue = append(queue, pending{id, discountSelect(t, s, id, 500+step*3000)})
				if step < 3 {
					continue
				}
				p := queue[0]
				queue = queue[1:]
				s.RecordOutcome(p.worker, -0.5-0.05*float64(step), true, TaskContext{TaskID: p.id})

				s.mu.Lock()
				arm := s.arms[p.worker]
				A, inv := mat.DenseCopyOf(arm.A), mat.DenseCopyOf(arm.Ainv)
				s.mu.Unlock()
				maxA := 1.0
				for i := 0; i < s.dim; i++ {
					for j := 0; j < s.dim; j++ {
						maxA = math.Max(maxA, math.Abs(A.At(i, j)))
					}
				}
				for i := 0; i < s.dim; i++ {
					for j := 0; j < s.dim; j++ {
						var product float64
						for k := 0; k < s.dim; k++ {
							product += A.At(i, k) * inv.At(k, j)
						}
						want := 0.0
						if i == j {
							want = 1
						}
						if math.Abs(product-want) > 1e-9*maxA {
							t.Fatalf("step %d, %s: A*Ainv[%d,%d] = %.17g right after the outcome", step, p.worker, i, j, product)
						}
					}
				}
			}
		})
	}
}

func TestLinUCBDiscount_GlobalScoreContinuesWhenCatchUpFails(t *testing.T) {
	s := newDiscountScheduler(t, 2, 0.9, DiscountModeGlobal)
	var x *mat.VecDense
	for i := 0; i < 3; i++ {
		id := fmt.Sprintf("score-setup-%d", i)
		discountSelect(t, s, id, 100+i*200)
		x = discountPending(t, s, id)
		s.RecordOutcome("worker-a", -float64(i+1), true, TaskContext{TaskID: id})
	}
	beforeMean, beforeBonus := s.score("worker-a", x)
	if beforeMean == 0 || beforeBonus == 0 {
		t.Fatalf("setup has trivial score: mean=%g bonus=%g", beforeMean, beforeBonus)
	}
	atomic.AddInt64(&s.totalDispatches, 1)
	s.mu.Lock()
	arm := s.arms["worker-a"]
	lastStep := arm.lastStep
	for i := 0; i < s.dim; i++ {
		for j := 0; j < s.dim; j++ {
			arm.A.Set(i, j, 0)
		}
		arm.A.Set(i, i, -5)
	}
	corruptState := dumpArm(arm)
	// Compute the score from the untouched cached inverse and theta.
	wantMean := mat.Dot(arm.theta, x)
	var quadratic float64
	for i := 0; i < s.dim; i++ {
		for j := 0; j < s.dim; j++ {
			quadratic += x.AtVec(i) * arm.Ainv.At(i, j) * x.AtVec(j)
		}
	}
	wantBonus := s.alpha * math.Sqrt(quadratic)
	s.mu.Unlock()
	if wantMean != beforeMean || math.Abs(wantBonus-beforeBonus) > 1e-10*discountMax(corruptState.A) {
		t.Fatalf("reference score changed: got (%g, %g), want (%g, %g)", wantMean, wantBonus, beforeMean, beforeBonus)
	}
	mean, bonus := s.score("worker-a", x)
	if mean != beforeMean || bonus != beforeBonus {
		t.Errorf("failed catch-up score = (%g, %g), want exactly (%g, %g)", mean, bonus, beforeMean, beforeBonus)
	}
	s.mu.Lock()
	after := dumpArm(arm)
	gotStep := arm.lastStep
	s.mu.Unlock()
	if !reflect.DeepEqual(after, corruptState) || gotStep != lastStep {
		t.Errorf("failed catch-up changed A, Ainv, b, count or lastStep: lastStep=%d, want %d", gotStep, lastStep)
	}
}

func TestLinUCBDiscount_FailedFirstUpdateDoesNotCreateArm(t *testing.T) {
	for _, mode := range []DiscountMode{DiscountModeGlobal, DiscountModeArm} {
		t.Run(string(mode), func(t *testing.T) {
			s := newDiscountScheduler(t, 2, 0.9, mode)
			good := mat.NewVecDense(s.dim, nil)
			good.SetVec(0, 0.5)
			s.mu.Lock()
			s.pendingX["good"] = good
			if mode == DiscountModeGlobal {
				s.pendingStep["good"] = 1
			}
			s.mu.Unlock()
			s.RecordOutcome("worker-a", -1, true, TaskContext{TaskID: "good"})
			before := discountArm(t, s, "worker-a")
			s.mu.Lock()
			original := s.arms["worker-a"]
			beforeStep, beforeDirty := original.lastStep, original.dirty
			beforeTheta := mat.VecDenseCopyOf(original.theta)
			s.mu.Unlock()

			for _, tc := range []struct{ id, worker string }{{"bad-new", "new-worker"}, {"bad-existing", "worker-a"}} {
				bad := mat.NewVecDense(s.dim, nil)
				bad.SetVec(0, math.Inf(1))
				s.mu.Lock()
				s.pendingX[tc.id] = bad
				if mode == DiscountModeGlobal {
					s.pendingStep[tc.id] = 1
				}
				s.mu.Unlock()
				s.RecordOutcome(tc.worker, -1, true, TaskContext{TaskID: tc.id})
				s.mu.Lock()
				_, hasX := s.pendingX[tc.id]
				_, hasStep := s.pendingStep[tc.id]
				_, hasNewArm := s.arms["new-worker"]
				after := dumpArm(s.arms["worker-a"])
				afterStep, afterDirty := s.arms["worker-a"].lastStep, s.arms["worker-a"].dirty
				afterTheta := mat.VecDenseCopyOf(s.arms["worker-a"].theta)
				armCount := len(s.arms)
				s.mu.Unlock()
				if hasX || hasStep {
					t.Errorf("%s: failed update left pending cache: x=%t step=%t", tc.id, hasX, hasStep)
				}
				if hasNewArm || armCount != 1 {
					t.Errorf("%s: failed first update created new-worker arm (arms=%d)", tc.id, armCount)
				}
				if !reflect.DeepEqual(after, before) || afterStep != beforeStep || afterDirty != beforeDirty || !mat.Equal(afterTheta, beforeTheta) {
					t.Errorf("%s: failed update changed existing arm state", tc.id)
				}
			}
		})
	}
}

func TestLinUCBDiscount_GlobalOutcomeAfterIdleCatchesUpInsideUpdate(t *testing.T) {
	reg := newRegistryWithWorkers(t, 2)
	s := NewLinUCBScheduler(LinUCBConfig{Registry: sortedRegistry{reg}, Discount: 0.9, DiscountMode: DiscountModeGlobal})
	s.mu.Lock()
	arm := s.armForLocked("worker-a")
	arm.A.Set(0, 0, 2)
	arm.A.Set(1, 1, 3)
	arm.Ainv.Set(0, 0, 0.5)
	arm.Ainv.Set(1, 1, 1.0/3)
	arm.b.SetVec(0, 1)
	arm.b.SetVec(1, -2)
	arm.lastStep = 1
	arm.count = 2
	x := mat.NewVecDense(s.dim, nil)
	x.SetVec(0, 0.5)
	x.SetVec(1, -1)
	s.pendingX["t1"] = x
	s.pendingStep["t1"] = 4
	s.mu.Unlock()
	atomic.StoreInt64(&s.totalDispatches, 1)
	if err := reg.Remove("worker-a"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 5; i++ {
		if got := discountSelect(t, s, "", i+1); got != "worker-b" {
			t.Fatalf("idle dispatch selected %s, want worker-b", got)
		}
	}
	if clock := atomic.LoadInt64(&s.totalDispatches); clock != 6 {
		t.Fatalf("dispatch clock = %d, want 6", clock)
	}
	s.RecordOutcome("worker-a", -1.5, true, TaskContext{TaskID: "t1"})
	got := discountArm(t, s, "worker-a")
	wantA, wantB := discountIdentity(s.dim)
	oldA, oldB := discountIdentity(s.dim)
	oldA[0][0], oldA[1][1] = 2, 3
	oldB[0], oldB[1] = 1, -2
	aged := math.Pow(0.9, 5)
	lagged := math.Pow(0.9, 2)
	for i := 0; i < s.dim; i++ {
		wantB[i] = aged*oldB[i] + lagged*(-1.5)*x.AtVec(i)
		for j := 0; j < s.dim; j++ {
			base := 0.0
			if i == j {
				base = 1
			}
			wantA[i][j] = base + aged*(oldA[i][j]-base) + lagged*x.AtVec(i)*x.AtVec(j)
		}
	}
	discountAssertState(t, got, wantA, wantB)
	s.mu.Lock()
	step, count := s.arms["worker-a"].lastStep, s.arms["worker-a"].count
	s.mu.Unlock()
	if step != 6 || count != 3 {
		t.Errorf("outcome after idle: lastStep=%d count=%d, want 6 and 3", step, count)
	}
}

func TestLinUCBDiscount_UnknownModeNormalizedToGlobal(t *testing.T) {
	unknown := newDiscountScheduler(t, 2, 0.9, "bogus")
	global := newDiscountScheduler(t, 2, 0.9, DiscountModeGlobal)
	if unknown.discountMode != DiscountModeGlobal {
		t.Errorf("unknown mode normalized to %q, want %q", unknown.discountMode, DiscountModeGlobal)
	}
	for _, s := range []*LinUCBScheduler{unknown, global} {
		for i, step := range []int64{1, 4} {
			id := fmt.Sprintf("mode-%d", i)
			x := mat.NewVecDense(s.dim, nil)
			x.SetVec(0, float64(i+1)/2)
			s.mu.Lock()
			s.pendingX[id] = x
			s.pendingStep[id] = step
			s.mu.Unlock()
			atomic.StoreInt64(&s.totalDispatches, step+int64(i))
			s.RecordOutcome("worker-a", -float64(i+1), true, TaskContext{TaskID: id})
		}
	}
	if got, want := discountArm(t, unknown, "worker-a"), discountArm(t, global, "worker-a"); !reflect.DeepEqual(got, want) {
		t.Errorf("unknown mode state differs from global: got %+v, want %+v", got, want)
	}
	unknown.mu.Lock()
	unknownStep := unknown.arms["worker-a"].lastStep
	unknown.mu.Unlock()
	global.mu.Lock()
	globalStep := global.arms["worker-a"].lastStep
	global.mu.Unlock()
	if unknownStep != globalStep {
		t.Errorf("unknown mode lastStep=%d, global lastStep=%d", unknownStep, globalStep)
	}
}
