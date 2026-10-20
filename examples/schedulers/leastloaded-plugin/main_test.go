package main

import "testing"

func TestSupportsArch(t *testing.T) {
	tests := []struct {
		name       string
		nativeArch string
		targetArch string
		want       bool
	}{
		{"exact match", "ARCH_ARM64", "ARCH_ARM64", true},
		{"mismatch", "ARCH_X86_64", "ARCH_ARM64", false},
		{"unspecified target accepts any", "ARCH_X86_64", "ARCH_UNSPECIFIED", true},
		{"empty target accepts any", "ARCH_ARMV7", "", true},
		{"unspecified native never matches a real target", "ARCH_UNSPECIFIED", "ARCH_ARM64", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := supportsArch(tt.nativeArch, tt.targetArch); got != tt.want {
				t.Errorf("supportsArch(%q, %q) = %v, want %v", tt.nativeArch, tt.targetArch, got, tt.want)
			}
		})
	}
}
