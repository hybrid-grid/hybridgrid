package main

import (
	"os"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
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

func TestFilterHgbuildFlags_SetsCoordinatorAndTimeout(t *testing.T) {
	origCoord, origTimeout := coordinator, timeout
	defer func() { coordinator, timeout = origCoord, origTimeout }()
	coordinator = ""
	timeout = 2 * time.Minute

	// Space form — the shape that used to be silently swallowed
	// (cc runs with DisableFlagParsing so cobra never sees it).
	filtered := filterHgbuildFlags([]string{"--coordinator", "localhost:19002", "-c", "a.c"})
	if coordinator != "localhost:19002" {
		t.Fatalf("space-form --coordinator ignored: got %q", coordinator)
	}
	if !containsArg(filtered, "-c") || !containsArg(filtered, "a.c") {
		t.Fatalf("compiler args must survive filtering: %v", filtered)
	}
	if containsArg(filtered, "--coordinator") || containsArg(filtered, "localhost:19002") {
		t.Fatalf("hgbuild flags leaked into compiler args: %v", filtered)
	}

	// Equals forms + timeout.
	filtered = filterHgbuildFlags([]string{"--coordinator=h1:1", "--timeout=10m", "-c", "a.c"})
	if coordinator != "h1:1" {
		t.Fatalf("equals-form --coordinator ignored: got %q", coordinator)
	}
	if timeout != 10*time.Minute {
		t.Fatalf("equals-form --timeout ignored: got %v", timeout)
	}
	if containsArg(filtered, "--timeout=10m") {
		t.Fatalf("hgbuild flags leaked into compiler args: %v", filtered)
	}
}

func TestParseArchStrict_RejectsUnknown(t *testing.T) {
	cases := []struct {
		in      string
		want    pb.Architecture
		wantErr bool
	}{
		{"x86_64", pb.Architecture_ARCH_X86_64, false},
		{"ARM64", pb.Architecture_ARCH_ARM64, false},
		{"", getLocalArch(), false},
		{"riscv", pb.Architecture_ARCH_UNSPECIFIED, true},
		{"arm6", pb.Architecture_ARCH_UNSPECIFIED, true},
	}
	for _, c := range cases {
		got, err := parseArchStrict(c.in)
		if (err != nil) != c.wantErr {
			t.Fatalf("parseArchStrict(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
		}
		if got != c.want {
			t.Fatalf("parseArchStrict(%q) = %v, want %v", c.in, got, c.want)
		}
	}
	// Tolerant wrapper (compiler -march values like "native") must not panic.
	if got := parseArch("native"); got != pb.Architecture_ARCH_UNSPECIFIED {
		t.Fatalf("parseArch(native) = %v, want UNSPECIFIED", got)
	}
}
