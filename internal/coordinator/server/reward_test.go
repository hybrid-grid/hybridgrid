package server

import (
	"math"
	"testing"
	"time"
)

func TestFailureReward_KeepsHistoricalScaleAndOrdering(t *testing.T) {
	successReward := func(d time.Duration) float64 {
		return -math.Log1p(float64(d.Milliseconds())) / math.Log1p(rewardNormalizationMs)
	}
	tests := []struct {
		name    string
		timeout time.Duration
		want    float64
	}{
		{name: "historical 120s timeout scores exactly -1", timeout: 120 * time.Second, want: -1},
		{name: "shorter timeout keeps the -1 floor", timeout: 30 * time.Second, want: -1},
		{name: "unset timeout keeps -1", timeout: 0, want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := failureReward(tt.timeout); got != tt.want {
				t.Errorf("failureReward(%v) = %v, want %v", tt.timeout, got, tt.want)
			}
		})
	}

	// With a 600 s timeout every success that finishes in time must still
	// score better than a failure.
	long := failureReward(600 * time.Second)
	if long >= -1 {
		t.Fatalf("failureReward(600s) = %v, want below -1", long)
	}
	for _, d := range []time.Duration{time.Second, 120 * time.Second, 193 * time.Second, 599 * time.Second} {
		if s := successReward(d); s <= long {
			t.Errorf("success after %v scores %v, not better than failure %v", d, s, long)
		}
	}
}
