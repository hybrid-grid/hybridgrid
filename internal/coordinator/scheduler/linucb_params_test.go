package scheduler

import "testing"

func TestLinUCBScheduler_Params(t *testing.T) {
	tests := []struct {
		name string
		cfg  LinUCBConfig
		want string
	}{
		{"default", LinUCBConfig{}, "alpha=0.5 warm_start=0 load_penalty=0"},
		{"hybrid", LinUCBConfig{Alpha: 1, WarmStartTasks: 100, LoadPenalty: 0.5}, "alpha=1 warm_start=100 load_penalty=0.5"},
		{"discount global", LinUCBConfig{Discount: 0.95}, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.95 discount_mode=global"},
		{"discount arm", LinUCBConfig{Discount: 0.98, DiscountMode: DiscountModeArm}, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.98 discount_mode=arm"},
		{"gamma one", LinUCBConfig{Discount: 1, DiscountMode: DiscountModeArm}, "alpha=0.5 warm_start=0 load_penalty=0"},
		{"invalid gamma", LinUCBConfig{Discount: -0.5, DiscountMode: DiscountModeArm}, "alpha=0.5 warm_start=0 load_penalty=0"},
		{"gamma above one", LinUCBConfig{Discount: 1.5, DiscountMode: DiscountModeArm}, "alpha=0.5 warm_start=0 load_penalty=0"},
		{"zero alpha uses default", LinUCBConfig{Alpha: 0, Discount: 0.99}, "alpha=0.5 warm_start=0 load_penalty=0 discount=0.99 discount_mode=global"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLinUCBScheduler(tt.cfg).Params(); got != tt.want {
				t.Fatalf("Params() = %q, want %q", got, tt.want)
			}
		})
	}
}
