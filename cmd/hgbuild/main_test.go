package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
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

func TestInstallWrappers(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bin")
	if err := installWrappers(dir); err != nil {
		t.Fatalf("installWrappers failed: %v", err)
	}
	for _, tt := range []struct {
		name string
		mode string
	}{
		{"hgcc", wrapperCCEnv},
		{"hgcxx", wrapperCXXEnv},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(dir, tt.name)
			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("wrapper missing: %v", err)
			}
			if info.Mode().Perm()&0111 == 0 {
				t.Errorf("wrapper %s is not executable", path)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read wrapper: %v", err)
			}
			if !strings.Contains(string(data), "export "+tt.mode+"=1") || !strings.Contains(string(data), "exec ") {
				t.Errorf("wrapper does not select compiler mode: %s", data)
			}
		})
	}
}

func TestCompileTimeout_EnvOverride(t *testing.T) {
	saved := timeout
	defer func() { timeout = saved }()
	timeout = 2 * time.Minute

	tests := []struct {
		name        string
		env         string
		wantCompile time.Duration
		wantWorker  time.Duration
	}{
		{name: "unset keeps flag and 5m worker deadline", env: "", wantCompile: 2 * time.Minute, wantWorker: 5 * time.Minute},
		{name: "longer value raises both", env: "660s", wantCompile: 660 * time.Second, wantWorker: 660 * time.Second},
		{name: "shorter value keeps 5m worker floor", env: "3m", wantCompile: 3 * time.Minute, wantWorker: 5 * time.Minute},
		{name: "invalid falls back to flag", env: "soon", wantCompile: 2 * time.Minute, wantWorker: 5 * time.Minute},
		{name: "non-positive falls back to flag", env: "0s", wantCompile: 2 * time.Minute, wantWorker: 5 * time.Minute},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(timeoutEnv, tt.env)
			if got := compileTimeout(); got != tt.wantCompile {
				t.Errorf("compileTimeout() = %v, want %v", got, tt.wantCompile)
			}
			if got := workerTimeout(); got != tt.wantWorker {
				t.Errorf("workerTimeout() = %v, want %v", got, tt.wantWorker)
			}
		})
	}
}
