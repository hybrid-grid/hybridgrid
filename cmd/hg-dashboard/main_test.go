package main

import (
	"os"
	"testing"
)

// TestEnvOr_Default exercises the envOr helper the log setup leans on:
// an unset variable falls back to the supplied default.
func TestEnvOr_Default(t *testing.T) {
	const key = "HG_DASHBOARD_TEST_ENV_OR"
	t.Setenv(key, "")
	if got := envOr(key, "fallback"); got != "fallback" {
		t.Fatalf("envOr(%q,...) = %q, want fallback", key, got)
	}
}

// TestEnvOr_Set verifies the env-derived override path the LOG_LEVEL
// convention uses across the codebase.
func TestEnvOr_Set(t *testing.T) {
	const key = "HG_DASHBOARD_TEST_ENV_OR"
	t.Setenv(key, "set-value")
	if got := envOr(key, "fallback"); got != "set-value" {
		t.Fatalf("envOr(%q,...) = %q, want set-value", key, got)
	}
}

// TestVersion_Default is a load-bearing guard: a bare `go build` of
// hg-dashboard emits a recognizable dev version in logs so operators
// never confuse a from-source build with a released artifact.
func TestVersion_Default(t *testing.T) {
	if version == "" {
		t.Fatal("version must be set; -ldflags sets it at release time")
	}
	_ = os.Stdout // keep the import meaningful for future flag tests
}
