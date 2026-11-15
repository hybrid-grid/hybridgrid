package scheduler

import (
	"fmt"
	"math"
	"strings"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func nearExperiment(t *testing.T, got, want float64) {
	t.Helper()
	if math.Abs(got-want) > 1e-9 {
		t.Fatalf("got %.12g, want %.12g", got, want)
	}
}

func experimentSelect(t *testing.T, s *LinUCBScheduler, id string) (*registry.WorkerInfo, DispatchInfo) {
	t.Helper()
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP,
		pb.Architecture_ARCH_X86_64, "", TaskContext{TaskID: id})
	if err != nil {
		t.Fatal(err)
	}
	return w, info
}

func TestLinUCB_ExperimentalDefaultsPreserveState(t *testing.T) {
	r := newEWMARegistry(t, "a", "b", "c")
	base := NewLinUCBScheduler(LinUCBConfig{Registry: r, WarmStartTasks: 2, LoadPenalty: 0.5})
	explicit := NewLinUCBScheduler(LinUCBConfig{Registry: r, WarmStartTasks: 2, LoadPenalty: 0.5,
		FeatureSet: "full", RewardCenter: "none", DiagCandidates: false})
	for i := 0; i < 9; i++ {
		ctx := TaskContext{TaskID: fmt.Sprintf("default-%d", i), SourceSizeBytes: 1000 + 17*i,
			RawSourceSizeBytes: 100 + i, SourceFilename: "unit.cpp"}
		w1, d1, err := base.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		if err != nil {
			t.Fatal(err)
		}
		w2, d2, err := explicit.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		if err != nil {
			t.Fatal(err)
		}
		if w1.ID != w2.ID || d1.WasExploration != d2.WasExploration || d1.Candidates != nil || d2.Candidates != nil {
			t.Fatalf("dispatch %d differs: %s %+v / %s %+v", i, w1.ID, d1, w2.ID, d2)
		}
		nearExperiment(t, d1.QValueAtDispatch, d2.QValueAtDispatch)
		reward := -0.2 - float64(i%3)*0.1
		base.RecordOutcome(w1.ID, reward, true, ctx)
		explicit.RecordOutcome(w2.ID, reward, true, ctx)
	}
	for _, id := range []string{"a", "b", "c"} {
		x := base.featureVector(&registry.WorkerInfo{ID: id}, pb.Architecture_ARCH_X86_64, TaskContext{})
		base.score(id, x)
		explicit.score(id, x)
		a, b := base.arms[id], explicit.arms[id]
		if a == nil || b == nil || a.count != b.count {
			t.Fatalf("arm %s count differs", id)
		}
		for i := 0; i < base.dim; i++ {
			nearExperiment(t, a.b.AtVec(i), b.b.AtVec(i))
			nearExperiment(t, a.theta.AtVec(i), b.theta.AtVec(i))
			for j := 0; j < base.dim; j++ {
				nearExperiment(t, a.A.At(i, j), b.A.At(i, j))
				nearExperiment(t, a.Ainv.At(i, j), b.Ainv.At(i, j))
			}
		}
	}
}

func TestLinUCB_CompactFeaturesAndSelection(t *testing.T) {
	r := newEWMARegistry(t, "a", "b")
	s := NewLinUCBScheduler(LinUCBConfig{Registry: r, FeatureSet: "compact"})
	if s.dim != 4 || s.Params() != "alpha=0.5 warm_start=0 load_penalty=0 features=compact" {
		t.Fatalf("compact setup: dim=%d params=%q", s.dim, s.Params())
	}
	w := &registry.WorkerInfo{ID: "sample", MaxParallel: 4, ActiveTasks: 2}
	ctx := TaskContext{SourceSizeBytes: 1 << 16, RawSourceSizeBytes: 1 << 15}
	x := s.featureVector(w, pb.Architecture_ARCH_X86_64, ctx)
	for i, want := range []float64{1, math.Log1p(1<<16) / sizeNormDenom, 0.5,
		math.Log1p(1<<15) / rawSizeNormDenom} {
		nearExperiment(t, x.AtVec(i), want)
	}
	capped := s.featureVector(w, pb.Architecture_ARCH_X86_64,
		TaskContext{SourceSizeBytes: 8 << 20, RawSourceSizeBytes: 2 << 20})
	nearExperiment(t, capped.AtVec(1), 1)
	nearExperiment(t, capped.AtVec(3), 1)
	for i := 0; i < 12; i++ {
		ctx.TaskID = fmt.Sprintf("compact-%d", i)
		chosen, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", ctx)
		if err != nil {
			t.Fatal(err)
		}
		s.RecordOutcome(chosen.ID, -0.3, true, ctx)
	}
	for id, arm := range s.arms {
		rows, cols := arm.A.Dims()
		if rows != 4 || cols != 4 || arm.b.Len() != 4 {
			t.Fatalf("arm %s has wrong dimension", id)
		}
	}
}

func TestLinUCB_WarmCenterFreezeAndReplay(t *testing.T) {
	r := newEWMARegistry(t, "a", "b")
	s := NewLinUCBScheduler(LinUCBConfig{Registry: r, WarmStartTasks: 2,
		RewardCenter: "warm", FeatureSet: "compact"})
	first, firstInfo := experimentSelect(t, s, "first")
	if first.ID != "a" || firstInfo.Candidates != nil {
		t.Fatalf("first warm choice: %s %+v", first.ID, firstInfo)
	}
	if err := r.IncrementTasks("a"); err != nil {
		t.Fatal(err)
	}
	second, _ := experimentSelect(t, s, "second")
	third, thirdInfo := experimentSelect(t, s, "third")
	if second.ID != "b" || third.ID != "b" || !thirdInfo.WasExploration {
		t.Fatalf("warm start did not continue past dispatch 2: %s %s %+v", second.ID, third.ID, thirdInfo)
	}
	s.RecordOutcome(second.ID, -3, false, TaskContext{TaskID: "second"})
	if s.centerFrozen || len(s.warmOutcomes) != 1 || s.arms["b"].count != 0 {
		t.Fatal("updated an arm before the second completion")
	}
	s.RecordOutcome(first.ID, -1, true, TaskContext{TaskID: "first"})
	if !s.centerFrozen || len(s.warmOutcomes) != 0 {
		t.Fatal("did not freeze at two recorded outcomes")
	}
	nearExperiment(t, s.rewardOffset, -2)
	// The two warm vectors are [1,0,0,0]. Ridge gives theta = residual / 2.
	for id, want := range map[string]float64{"a": 0.5, "b": -0.5} {
		arm := s.arms[id]
		if arm.count != 1 {
			t.Fatalf("arm %s count=%d", id, arm.count)
		}
		nearExperiment(t, arm.A.At(0, 0), 2)
		nearExperiment(t, arm.b.AtVec(0), 2*want)
		nearExperiment(t, arm.Ainv.At(0, 0)*arm.b.AtVec(0), want)
	}
	if err := r.DecrementTasks("a", true, 0); err != nil {
		t.Fatal(err)
	}
	chosen, info := experimentSelect(t, s, "after")
	if chosen.ID != "a" || info.WasExploration || info.Candidates != nil {
		t.Fatalf("post-freeze choice: %s %+v", chosen.ID, info)
	}
	nearExperiment(t, info.RewardOffset, -2)
	nearExperiment(t, info.QValueAtDispatch, -2+0.5+0.5*math.Sqrt(0.5))
	if s.Params() != "alpha=0.5 warm_start=2 load_penalty=0 features=compact reward_center=warm" {
		t.Fatalf("Params()=%q", s.Params())
	}
	// A later observation uses the same frozen offset.
	s.RecordOutcome(chosen.ID, -4, true, TaskContext{TaskID: "after"})
	nearExperiment(t, s.arms["a"].b.AtVec(0), -1)
}

func TestLinUCB_WarmCenterValidationAndIncomplete(t *testing.T) {
	for _, tc := range []struct {
		name     string
		warm     int
		discount float64
		mode     DiscountMode
		want     string
	}{
		{"no warm start", 0, 0, "", "none"},
		{"global discount", 2, 0.98, DiscountModeGlobal, "none"},
		{"ineffective global discount (gamma 1 is no discounting)", 2, 1, DiscountModeGlobal, "warm"},
		{"arm discount", 2, 0.98, DiscountModeArm, "warm"},
		{"no discount", 2, 0, DiscountModeGlobal, "warm"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewLinUCBScheduler(LinUCBConfig{WarmStartTasks: tc.warm,
				Discount: tc.discount, DiscountMode: tc.mode, RewardCenter: "warm"})
			if s.rewardCenter != tc.want {
				t.Fatalf("center=%q, want %q", s.rewardCenter, tc.want)
			}
		})
	}
	for _, value := range []string{"", "full", "compact"} {
		if !ValidFeatureSet(value) {
			t.Fatalf("valid feature set %q rejected", value)
		}
	}
	if ValidFeatureSet("other") || ValidRewardCenter("other") || !ValidRewardCenter("warm") {
		t.Fatal("invalid validation result")
	}
	invalid := NewLinUCBScheduler(LinUCBConfig{FeatureSet: "other", RewardCenter: "other"})
	if invalid.featureSet != "full" || invalid.rewardCenter != "none" || invalid.dim != featureDim() {
		t.Fatal("invalid constructor values did not fall back")
	}
	r := newEWMARegistry(t, "a", "b")
	s := NewLinUCBScheduler(LinUCBConfig{Registry: r, WarmStartTasks: 3, RewardCenter: "warm"})
	w, _ := experimentSelect(t, s, "one")
	s.RecordOutcome(w.ID, -1, true, TaskContext{TaskID: "one"})
	for i := 0; i < 4; i++ {
		_, info := experimentSelect(t, s, fmt.Sprintf("uncompleted-%d", i))
		if !info.WasExploration || info.RewardOffset != 0 || info.Candidates != nil {
			t.Fatalf("incomplete warm start: %+v", info)
		}
	}
	if s.centerFrozen || len(s.warmOutcomes) != 1 || s.arms[w.ID].count != 0 {
		t.Fatal("incomplete warm start advanced learning")
	}
}

func TestLinUCB_WarmCenterArmDiscountReplay(t *testing.T) {
	r := newEWMARegistry(t, "a", "b")
	s := NewLinUCBScheduler(LinUCBConfig{Registry: r, FeatureSet: "compact",
		WarmStartTasks: 2, RewardCenter: "warm", Discount: 0.5, DiscountMode: DiscountModeArm})
	for i, reward := range []float64{-1, -3} {
		id := fmt.Sprintf("arm-%d", i)
		w, _ := experimentSelect(t, s, id)
		if w.ID != "a" {
			t.Fatalf("warm choice %d = %s", i, w.ID)
		}
		s.RecordOutcome(w.ID, reward, true, TaskContext{TaskID: id})
	}
	a := s.arms["a"]
	if !s.centerFrozen || a.count != 2 {
		t.Fatal("arm discount replay was incomplete")
	}
	nearExperiment(t, s.rewardOffset, -2)
	nearExperiment(t, a.A.At(0, 0), 2.5)
	nearExperiment(t, a.b.AtVec(0), -0.5)
	nearExperiment(t, a.Ainv.At(0, 0)*a.b.AtVec(0), -0.2)
}

func TestLinUCB_CandidateDiagnostics(t *testing.T) {
	r := &ewmaOrderedRegistry{InMemoryRegistry: newTestRegistry(),
		order: map[string]int{"a": 0, "b": 1, "c": 2}}
	t.Cleanup(r.Stop)
	for _, id := range []string{"a", "b", "c"} {
		if err := r.Add(&registry.WorkerInfo{
			ID: id, Address: id + ":1234", MaxParallel: 4,
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
		}); err != nil {
			t.Fatal(err)
		}
	}
	s := NewLinUCBScheduler(LinUCBConfig{Registry: r, DiagCandidates: true,
		WarmStartTasks: 1, LoadPenalty: 0.5})
	_, warm := experimentSelect(t, s, "warm")
	if warm.Candidates != nil {
		t.Fatal("diagnostics during warm start")
	}
	chosen, info := experimentSelect(t, s, "bandit")
	if len(info.Candidates) != 3 {
		t.Fatalf("candidate count=%d", len(info.Candidates))
	}
	workers := r.ListByCapability(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	bestP := math.Inf(-1)
	bestIndex := -1
	for i, d := range info.Candidates {
		if d.Address != workers[i].Address || len(d.X) != 12 || d.ArmCount != 0 {
			t.Fatalf("candidate %d mismatch: %+v", i, d)
		}
		nearExperiment(t, d.P, d.Mean-d.LoadPenalty+d.Bonus)
		nearExperiment(t, d.Bonus, s.alpha*math.Sqrt(d.XAinvX))
		if d.P > bestP {
			bestP, bestIndex = d.P, i
		}
	}
	if chosen.ID != workers[bestIndex].ID {
		t.Fatalf("chosen %s, best candidate %s", chosen.ID, workers[bestIndex].ID)
	}
	nearExperiment(t, info.QValueAtDispatch, bestP)
	// Mutating diagnostics cannot mutate the cached feature vector.
	info.Candidates[bestIndex].X[0] = 99
	nearExperiment(t, s.pendingX["bandit"].AtVec(0), 1)
	without := NewLinUCBScheduler(LinUCBConfig{Registry: r})
	_, quiet := experimentSelect(t, without, "quiet")
	if quiet.Candidates != nil {
		t.Fatal("diagnostics allocated when disabled")
	}
	for _, id := range []string{"b", "c"} {
		for i := 0; i < 4; i++ {
			if err := r.IncrementTasks(id); err != nil {
				t.Fatal(err)
			}
		}
	}
	_, single := experimentSelect(t, s, "single")
	if single.Candidates != nil {
		t.Fatal("diagnostics allocated on single-candidate path")
	}
}

func TestLinUCBExperiment_CenteringFollowsEffectiveDiscount(t *testing.T) {
	for _, tc := range []struct {
		name     string
		discount float64
		mode     DiscountMode
		want     string
	}{
		{"no discount", 0, "", "warm"},
		{"gamma one, global clock", 1, DiscountModeGlobal, "warm"},
		{"invalid gamma, global clock", 1.5, DiscountModeGlobal, "warm"},
		{"gamma 0.98, arm clock", 0.98, DiscountModeArm, "warm"},
		{"gamma 0.98, global clock", 0.98, DiscountModeGlobal, "none"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewLinUCBScheduler(LinUCBConfig{
				Registry: newTestRegistry(), WarmStartTasks: 10, RewardCenter: "warm",
				Discount: tc.discount, DiscountMode: tc.mode,
			})
			if s.rewardCenter != tc.want {
				t.Fatalf("rewardCenter = %q, want %q", s.rewardCenter, tc.want)
			}
			if got := strings.Contains(s.Params(), "reward_center=warm"); got != (tc.want == "warm") {
				t.Fatalf("Params() = %q does not match the active centering %q", s.Params(), tc.want)
			}
		})
	}
}

func TestLinUCBExperiment_ParamsReportDiagnostics(t *testing.T) {
	on := NewLinUCBScheduler(LinUCBConfig{Registry: newTestRegistry(), DiagCandidates: true})
	off := NewLinUCBScheduler(LinUCBConfig{Registry: newTestRegistry()})
	if !strings.Contains(on.Params(), "diag_candidates=on") {
		t.Fatalf("Params() = %q lacks the diagnostics marker", on.Params())
	}
	if strings.Contains(off.Params(), "diag_candidates") {
		t.Fatalf("Params() = %q must not mention diagnostics by default", off.Params())
	}
}
