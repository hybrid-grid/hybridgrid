package scheduler

import (
	"math"
	"testing"
)

func TestEWMA_ConfiguredAlpha(t *testing.T) {
	for _, tc := range []struct {
		name  string
		alpha float64
		want  float64
		param string
	}{
		{"unset", 0, 0.2, "alpha=0.2 warm_start=0"},
		{"sweep", 0.4, 0.4, "alpha=0.4 warm_start=0"},
		{"one", 1, 1, "alpha=1 warm_start=0"},
		{"negative", -1, 0.2, "alpha=0.2 warm_start=0"},
		{"over one", 1.1, 0.2, "alpha=0.2 warm_start=0"},
		{"nan", math.NaN(), 0.2, "alpha=0.2 warm_start=0"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := NewEWMAScheduler(EWMAConfig{Alpha: tc.alpha})
			if math.Abs(s.alpha-tc.want) > 1e-9 || s.Params() != tc.param {
				t.Fatalf("alpha=%g, params=%q", s.alpha, s.Params())
			}
		})
	}
	r := newEWMARegistry(t, "a")
	s := NewEWMAScheduler(EWMAConfig{Registry: r, Alpha: 0.4})
	ewmaObserve(t, s, "first", 10, 100)
	ewmaObserve(t, s, "second", 10, 200)
	want := 0.6*math.Log(10) + 0.4*math.Log(20)
	if math.Abs(s.est["a"]-want) > 1e-9 {
		t.Fatalf("estimate=%g, want %g", s.est["a"], want)
	}
}
