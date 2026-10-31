package main

import (
	"math"
	"strings"
	"testing"
)

func TestValidateDiscountFlags(t *testing.T) {
	tests := []struct {
		name      string
		scheduler string
		discount  float64
		mode      string
		wantError string
	}{
		{"default global", "hybrid-linucb-d", 0.98, "global", ""},
		{"arm", "hybrid-linucb-d", 0.95, "arm", ""},
		{"one", "hybrid-linucb-d", 1, "global", ""},
		{"zero", "hybrid-linucb-d", 0, "global", "--discount"},
		{"negative", "hybrid-linucb-d", -0.1, "global", "--discount"},
		{"greater than one", "hybrid-linucb-d", 1.5, "global", "--discount"},
		{"NaN", "hybrid-linucb-d", math.NaN(), "global", "--discount"},
		{"positive infinity", "hybrid-linucb-d", math.Inf(1), "global", "--discount"},
		{"negative infinity", "hybrid-linucb-d", math.Inf(-1), "global", "--discount"},
		{"invalid mode", "hybrid-linucb-d", 0.98, "bogus", "--discount-mode"},
		{"empty mode", "hybrid-linucb-d", 0.98, "", "--discount-mode"},
		{"other scheduler ignores invalid discount and mode", "hybrid-linucb", 0, "bogus", ""},
		{"other scheduler ignores NaN", "leastloaded", math.NaN(), "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDiscountFlags(tt.scheduler, tt.discount, tt.mode)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("validateDiscountFlags() = %v, want nil", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("validateDiscountFlags() = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}

func TestValidateSizeThreshold(t *testing.T) {
	tests := []struct {
		name      string
		scheduler string
		threshold int
		wantError bool
	}{
		{"p90 missing", "p90", 0, true},
		{"sed-p90 negative", "sed-p90", -1, true},
		{"p90 positive", "p90", 1, false},
		{"sed-p90 positive", "sed-p90", 1024, false},
		{"leastloaded ignores negative", "leastloaded", -1, false},
		{"sed ignores zero", "sed", 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSizeThreshold(tt.scheduler, tt.threshold)
			if tt.wantError {
				if err == nil || !strings.Contains(err.Error(), "--size-threshold") {
					t.Fatalf("validateSizeThreshold() = %v, want --size-threshold error", err)
				}
			} else if err != nil {
				t.Fatalf("validateSizeThreshold() = %v, want nil", err)
			}
		})
	}
}

func TestValidateExperimentFlags(t *testing.T) {
	type flags struct {
		scheduler    string
		rewardCenter string
		features     string
		discountMode string
		ewmaAlpha    float64
		discount     float64
		diag         bool
		warmStart    int
	}
	defaults := flags{scheduler: "leastloaded", rewardCenter: "none", features: "full", ewmaAlpha: 0.2, warmStart: 100, discount: 0.98, discountMode: "global"}
	tests := []struct {
		name      string
		change    func(*flags)
		wantError string
	}{
		{"defaults", func(*flags) {}, ""},
		{"invalid reward center", func(f *flags) { f.rewardCenter = "other" }, "--reward-center"},
		{"invalid features", func(f *flags) { f.features = "other" }, "--features"},
		{"zero ewma alpha", func(f *flags) { f.ewmaAlpha = 0 }, "--ewma-alpha"},
		{"negative ewma alpha", func(f *flags) { f.ewmaAlpha = -0.1 }, "--ewma-alpha"},
		{"large ewma alpha", func(f *flags) { f.ewmaAlpha = 1.1 }, "--ewma-alpha"},
		{"NaN ewma alpha", func(f *flags) { f.ewmaAlpha = math.NaN() }, "--ewma-alpha"},
		{"ewma alpha", func(f *flags) { f.scheduler = "ewma"; f.ewmaAlpha = 0.4 }, ""},
		{"ewma alpha unrelated", func(f *flags) { f.ewmaAlpha = 0.4 }, "--ewma-alpha"},
		{"compact linucb", func(f *flags) { f.scheduler = "linucb"; f.features = "compact" }, ""},
		{"compact hybrid", func(f *flags) { f.scheduler = "hybrid-linucb"; f.features = "compact" }, ""},
		{"compact discounted", func(f *flags) { f.scheduler = "hybrid-linucb-d"; f.features = "compact" }, ""},
		{"compact unrelated", func(f *flags) { f.features = "compact" }, "--features"},
		{"diagnostics linucb", func(f *flags) { f.scheduler = "linucb"; f.diag = true }, ""},
		{"diagnostics hybrid", func(f *flags) { f.scheduler = "hybrid-linucb"; f.diag = true }, ""},
		{"diagnostics discounted", func(f *flags) { f.scheduler = "hybrid-linucb-d"; f.diag = true }, ""},
		{"diagnostics unrelated", func(f *flags) { f.diag = true }, "--diag-candidates"},
		{"warm hybrid", func(f *flags) { f.scheduler = "hybrid-linucb"; f.rewardCenter = "warm" }, ""},
		{"warm linucb", func(f *flags) { f.scheduler = "linucb"; f.rewardCenter = "warm" }, "--reward-center"},
		{"warm unrelated", func(f *flags) { f.rewardCenter = "warm" }, "--reward-center"},
		{"warm without start", func(f *flags) { f.scheduler = "hybrid-linucb"; f.rewardCenter = "warm"; f.warmStart = 0 }, "--warm-start"},
		{"warm discounted global", func(f *flags) { f.scheduler = "hybrid-linucb-d"; f.rewardCenter = "warm" }, "--discount-mode"},
		{"warm discounted arm", func(f *flags) { f.scheduler = "hybrid-linucb-d"; f.rewardCenter = "warm"; f.discountMode = "arm" }, ""},
		{"warm discounted gamma one", func(f *flags) { f.scheduler = "hybrid-linucb-d"; f.rewardCenter = "warm"; f.discount = 1 }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := defaults
			tt.change(&f)
			err := validateExperimentFlags(f.scheduler, f.rewardCenter, f.features, f.ewmaAlpha, f.diag, f.warmStart, f.discount, f.discountMode)
			if tt.wantError == "" {
				if err != nil {
					t.Fatalf("validateExperimentFlags() = %v, want nil", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("validateExperimentFlags() = %v, want error containing %q", err, tt.wantError)
			}
		})
	}
}
