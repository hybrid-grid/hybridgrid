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
