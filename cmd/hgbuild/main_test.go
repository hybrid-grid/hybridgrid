package main

import (
	"os"
	"strings"
	"testing"

	"github.com/h3nr1-d14z/hybridgrid/internal/cli/taskid"
)

func TestFallbackEnabled_Default(t *testing.T) {
	t.Setenv(noFallbackEnv, "")
	noFallback = false

	if !fallbackEnabled() {
		t.Fatal("fallback should be enabled by default")
	}
}

func TestFallbackEnabled_DisabledByFlag(t *testing.T) {
	t.Setenv(noFallbackEnv, "")
	noFallback = true
	defer func() {
		noFallback = false
	}()

	if fallbackEnabled() {
		t.Fatal("fallback should be disabled when flag is set")
	}
}

func TestFallbackEnabled_DisabledByEnvironment(t *testing.T) {
	t.Setenv(noFallbackEnv, "1")
	noFallback = false

	if fallbackEnabled() {
		t.Fatal("fallback should be disabled when environment variable is set")
	}
}

func TestFilterHgbuildFlags_SetsNoFallback(t *testing.T) {
	noFallback = false
	defer func() {
		noFallback = false
	}()

	args := []string{"--no-fallback", "-c", "main.c"}
	filtered := filterHgbuildFlags(args)

	if !containsArg(filtered, "-c") || !containsArg(filtered, "main.c") {
		t.Fatal("expected compiler args to remain after filtering")
	}
	if containsArg(filtered, "--no-fallback") {
		t.Fatal("expected --no-fallback to be removed from compiler args")
	}
	if !noFallback {
		t.Fatal("expected --no-fallback to update global fallback state")
	}
	if fallbackEnabled() {
		t.Fatal("fallback should be disabled after filtering --no-fallback")
	}
}

func TestFilterHgbuildWrapperFlags_SetsNoFallback(t *testing.T) {
	noFallback = false
	defer func() {
		noFallback = false
	}()

	args := []string{"--no-fallback", "-j8"}
	filtered := filterHgbuildWrapperFlags(args)

	if containsArg(filtered, "--no-fallback") {
		t.Fatal("expected --no-fallback to be removed from wrapper args")
	}
	if !containsArg(filtered, "-j8") {
		t.Fatal("expected wrapper args to preserve build command arguments")
	}
	if !noFallback {
		t.Fatal("expected --no-fallback to update global fallback state for wrapper mode")
	}
}

func TestFilterHgbuildWrapperFlags_WrapCommandStillPresent(t *testing.T) {
	noFallback = false
	defer func() {
		noFallback = false
	}()

	args := []string{"--no-fallback", "make", "-j4"}
	filtered := filterHgbuildWrapperFlags(args)

	if len(filtered) != 2 {
		t.Fatalf("expected filtered args length 2, got %d", len(filtered))
	}
	if filtered[0] != "make" {
		t.Fatalf("expected command to remain first arg, got %q", filtered[0])
	}
	if filtered[1] != "-j4" {
		t.Fatalf("expected command args to remain unchanged, got %q", filtered[1])
	}
}

func TestSetEnv_ReplacesExistingValue(t *testing.T) {
	env := []string{"A=1", "B=2"}
	env = setEnv(env, "B", "3")

	for _, item := range env {
		if item == "B=3" {
			return
		}
	}

	t.Fatal("expected existing environment variable to be replaced")
}

func TestSetEnv_AppendsMissingValue(t *testing.T) {
	env := []string{"A=1"}
	env = setEnv(env, "B", "2")

	for _, item := range env {
		if item == "B=2" {
			return
		}
	}

	t.Fatal("expected missing environment variable to be appended")
}

func containsArg(args []string, target string) bool {
	for _, arg := range args {
		if arg == target {
			return true
		}
	}

	return false
}

func TestMain(m *testing.M) {
	code := m.Run()
	noFallback = false
	os.Exit(code)
}

func TestEnsureBuildSession(t *testing.T) {
	tests := []struct {
		name      string
		env       []string
		wantExact string
	}{
		{"keeps an explicit id", []string{"PATH=/bin", "HG_BUILD_ID=release-42"}, "release-42"},
		{"generates one when missing", []string{"PATH=/bin"}, ""},
		{"replaces an empty value", []string{"HG_BUILD_ID="}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := envValue(ensureBuildSession(tt.env), taskid.BuildIDEnvVar)
			if tt.wantExact != "" && got != tt.wantExact {
				t.Fatalf("HG_BUILD_ID = %q, want %q", got, tt.wantExact)
			}
			if tt.wantExact == "" && !strings.HasPrefix(got, "build-") {
				t.Fatalf("HG_BUILD_ID = %q, want a generated build- id", got)
			}
		})
	}
}

func TestWrappedBuildID(t *testing.T) {
	t.Setenv(taskid.BuildIDEnvVar, "")
	if got := wrappedBuildID(); got != "" {
		t.Fatalf("without HG_BUILD_ID got %q, want empty so lone cc calls do not each become a build", got)
	}

	t.Setenv(taskid.BuildIDEnvVar, "demo-1")
	if got := wrappedBuildID(); got != "demo-1" {
		t.Fatalf("got %q, want demo-1", got)
	}
}
