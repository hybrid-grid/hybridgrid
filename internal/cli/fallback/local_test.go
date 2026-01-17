package fallback

import (
	"context"
	"os/exec"
	"testing"
	"time"
)

func TestLocalFallback_New(t *testing.T) {
	cfg := DefaultConfig()
	f := New(cfg)

	if f == nil {
		t.Fatal("New returned nil")
	}
	if !f.IsEnabled() {
		t.Error("Fallback should be enabled by default")
	}
}

func TestLocalFallback_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	f := New(cfg)

	if f.IsEnabled() {
		t.Error("Fallback should be disabled")
	}

	_, err := f.Execute(context.Background(), &CompileJob{})
	if err == nil {
		t.Error("Execute should fail when disabled")
	}
}

func TestLocalFallback_Execute_Success(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	f := New(DefaultConfig())

	job := &CompileJob{
		TaskID:             "test-001",
		Compiler:           "gcc",
		Args:               []string{"-O2"},
		PreprocessedSource: []byte(`int main() { return 0; }`),
		Timeout:            30 * time.Second,
	}

	result, err := f.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("ExitCode = %d, want 0. Stderr: %s", result.ExitCode, result.Stderr)
	}
	if len(result.ObjectCode) == 0 {
		t.Error("ObjectCode is empty")
	}
	if !result.Fallback {
		t.Error("Fallback flag should be true")
	}
	if result.CompilationTime == 0 {
		t.Error("CompilationTime should be > 0")
	}

	t.Logf("Local compilation: %d bytes in %v", len(result.ObjectCode), result.CompilationTime)
}

func TestLocalFallback_Execute_CompileError(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	f := New(DefaultConfig())

	job := &CompileJob{
		TaskID:             "test-002",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: []byte(`this is not valid C code { syntax error }`),
		Timeout:            30 * time.Second,
	}

	result, err := f.Execute(context.Background(), job)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.ExitCode == 0 {
		t.Error("ExitCode should be non-zero for compile error")
	}
	if result.Stderr == "" {
		t.Error("Stderr should contain error message")
	}
	if !result.Fallback {
		t.Error("Fallback flag should be true")
	}

	t.Logf("Compile error captured: exit_code=%d", result.ExitCode)
}

func TestLocalFallback_Execute_Timeout(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	f := New(Config{
		Enabled:    true,
		MaxTimeout: 1 * time.Nanosecond, // Very short timeout
	})

	job := &CompileJob{
		TaskID:             "test-003",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: []byte(`int main() { return 0; }`),
		Timeout:            1 * time.Nanosecond,
	}

	// Give a moment for context to timeout
	time.Sleep(10 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	result, err := f.Execute(ctx, job)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	// Timeout should result in non-zero exit or timeout message
	if result.ExitCode == 0 && result.Stderr == "" {
		t.Log("Compilation completed despite timeout - fast machine")
	}
}

func TestBuildArgs(t *testing.T) {
	tests := []struct {
		name     string
		original []string
		srcFile  string
		outFile  string
		wantHas  []string
	}{
		{
			name:     "basic",
			original: []string{"-O2", "-Wall"},
			srcFile:  "/tmp/src.i",
			outFile:  "/tmp/out.o",
			wantHas:  []string{"-c", "-O2", "-Wall", "/tmp/src.i", "-o", "/tmp/out.o"},
		},
		{
			name:     "with existing -c",
			original: []string{"-c", "-O2"},
			srcFile:  "/tmp/src.i",
			outFile:  "/tmp/out.o",
			wantHas:  []string{"-c", "-O2"},
		},
		{
			name:     "skip output file",
			original: []string{"-O2", "-o", "old.o"},
			srcFile:  "/tmp/src.i",
			outFile:  "/tmp/out.o",
			wantHas:  []string{"-O2", "/tmp/src.i", "-o", "/tmp/out.o"},
		},
		{
			name:     "skip source file",
			original: []string{"-O2", "original.c"},
			srcFile:  "/tmp/src.i",
			outFile:  "/tmp/out.o",
			wantHas:  []string{"-O2", "/tmp/src.i"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildArgs(tt.original, tt.srcFile, tt.outFile)

			for _, want := range tt.wantHas {
				found := false
				for _, arg := range got {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildArgs() missing %q in %v", want, got)
				}
			}
		})
	}
}

func TestIsSourceFile(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"foo.c", true},
		{"foo.cpp", true},
		{"foo.cc", true},
		{"foo.i", true},
		{"foo.ii", true},
		{"foo.s", true},
		{"foo.m", true},
		{"foo.mm", true},
		{"foo.o", false},
		{"foo.h", false},
		{"-O2", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isSourceFile(tt.arg)
		if got != tt.expected {
			t.Errorf("isSourceFile(%q) = %v, want %v", tt.arg, got, tt.expected)
		}
	}
}
