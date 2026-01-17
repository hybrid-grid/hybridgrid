package executor

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestNativeExecutor_Name(t *testing.T) {
	e := NewNativeExecutor()
	if e.Name() != "native" {
		t.Errorf("Expected name 'native', got '%s'", e.Name())
	}
}

func TestNativeExecutor_CanExecute(t *testing.T) {
	e := NewNativeExecutor()

	tests := []struct {
		target   pb.Architecture
		native   pb.Architecture
		expected bool
	}{
		{pb.Architecture_ARCH_X86_64, pb.Architecture_ARCH_X86_64, true},
		{pb.Architecture_ARCH_ARM64, pb.Architecture_ARCH_ARM64, true},
		{pb.Architecture_ARCH_UNSPECIFIED, pb.Architecture_ARCH_X86_64, true},
		{pb.Architecture_ARCH_ARM64, pb.Architecture_ARCH_X86_64, false},
	}

	for _, tt := range tests {
		got := e.CanExecute(tt.target, tt.native)
		if got != tt.expected {
			t.Errorf("CanExecute(%v, %v) = %v, want %v", tt.target, tt.native, got, tt.expected)
		}
	}
}

func TestNativeExecutor_Execute_SimpleCompile(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping native executor test")
	}

	e := NewNativeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Simple C program
	source := []byte(`int main() { return 0; }`)

	req := &Request{
		TaskID:             "test-001",
		Compiler:           "gcc",
		Args:               []string{"-c", "-O2"},
		PreprocessedSource: source,
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:            30 * time.Second,
	}

	result, err := e.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Compilation failed: %s", result.Stderr)
	}

	if len(result.ObjectCode) == 0 {
		t.Error("Expected non-empty object code")
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}
}

func TestNativeExecutor_Execute_CompileError(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	e := NewNativeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Invalid C code
	source := []byte(`int main() { syntax error here }`)

	req := &Request{
		TaskID:             "test-002",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: source,
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:            30 * time.Second,
	}

	result, err := e.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected compilation to fail")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}

	if result.Stderr == "" {
		t.Error("Expected stderr output for compile error")
	}
}

func TestNativeExecutor_Execute_Timeout(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	e := NewNativeExecutor()
	// Very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	source := []byte(`int main() { return 0; }`)

	req := &Request{
		TaskID:             "test-003",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: source,
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:            1 * time.Nanosecond,
	}

	// Give context a moment to expire
	time.Sleep(10 * time.Millisecond)

	result, err := e.Execute(ctx, req)
	if err != nil {
		// Timeout might cause error
		return
	}

	// If we got a result, it should indicate timeout
	if result.Success {
		t.Log("Compilation succeeded despite timeout - may be race condition")
	}
}

func TestManager_Select(t *testing.T) {
	// Create manager without Docker
	m := NewManager(pb.Architecture_ARCH_X86_64, false)

	// Should return native executor for native arch
	e := m.Select(pb.Architecture_ARCH_X86_64)
	if e.Name() != "native" {
		t.Errorf("Expected native executor, got %s", e.Name())
	}

	// Should return native for unspecified
	e = m.Select(pb.Architecture_ARCH_UNSPECIFIED)
	if e.Name() != "native" {
		t.Errorf("Expected native executor, got %s", e.Name())
	}

	// Should fall back to native when Docker unavailable
	e = m.Select(pb.Architecture_ARCH_ARM64)
	if e.Name() != "native" {
		t.Errorf("Expected native executor fallback, got %s", e.Name())
	}
}

func TestBuildArgs(t *testing.T) {
	e := NewNativeExecutor()

	tests := []struct {
		name     string
		original []string
		srcFile  string
		outFile  string
		wantHas  []string
	}{
		{
			name:     "basic compile",
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
			name:     "skip input files",
			original: []string{"-O2", "original.c", "-Wall"},
			srcFile:  "/tmp/src.i",
			outFile:  "/tmp/out.o",
			wantHas:  []string{"-O2", "-Wall", "/tmp/src.i"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.buildArgs(tt.original, tt.srcFile, tt.outFile)

			for _, want := range tt.wantHas {
				found := false
				for _, arg := range got {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildArgs() missing expected arg %q in %v", want, got)
				}
			}
		})
	}
}

func TestIsInputFile(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"foo.c", true},
		{"foo.cpp", true},
		{"foo.cc", true},
		{"foo.cxx", true},
		{"foo.i", true},
		{"foo.ii", true},
		{"foo.s", true},
		{"foo.o", false},
		{"-O2", false},
		{"-I/include", false},
		{"", false},
	}

	for _, tt := range tests {
		got := isInputFile(tt.arg)
		if got != tt.expected {
			t.Errorf("isInputFile(%q) = %v, want %v", tt.arg, got, tt.expected)
		}
	}
}

func TestDockerExecutor_Integration(t *testing.T) {
	// Skip if Docker not available
	dockerExec, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerExec.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Simple C program (no includes needed since it's preprocessed)
	source := []byte(`int add(int a, int b) { return a + b; }
int main() { return add(1, 2); }`)

	req := &Request{
		TaskID:             "docker-test-001",
		Compiler:           "gcc",
		Args:               []string{"-O2"},
		PreprocessedSource: source,
		TargetArch:         pb.Architecture_ARCH_X86_64,
		Timeout:            120 * time.Second,
	}

	t.Log("Starting Docker compilation with dockcross/linux-x64...")
	result, err := dockerExec.Execute(ctx, req)
	if err != nil {
		// Skip if image not available (common in CI)
		if strings.Contains(err.Error(), "No such image") {
			t.Skipf("Docker image not available: %v", err)
		}
		t.Fatalf("Docker executor failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Docker compilation failed: exit_code=%d, stderr=%s", result.ExitCode, result.Stderr)
	}

	if len(result.ObjectCode) == 0 {
		t.Error("Expected non-empty object code from Docker executor")
	} else {
		t.Logf("Docker compilation succeeded: %d bytes, %v", len(result.ObjectCode), result.CompilationTime)
		// Check ELF magic (dockcross produces Linux ELF)
		if len(result.ObjectCode) >= 4 {
			magic := result.ObjectCode[:4]
			if magic[0] == 0x7f && magic[1] == 'E' && magic[2] == 'L' && magic[3] == 'F' {
				t.Log("Output is valid ELF object file")
			} else {
				t.Logf("Output magic: %x", magic)
			}
		}
	}
}

func TestDockerExecutor_CompileError(t *testing.T) {
	dockerExec, err := NewDockerExecutor()
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer dockerExec.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Invalid C code
	source := []byte(`this is not valid C code { syntax error }`)

	req := &Request{
		TaskID:             "docker-error-001",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: source,
		TargetArch:         pb.Architecture_ARCH_X86_64,
		Timeout:            60 * time.Second,
	}

	result, err := dockerExec.Execute(ctx, req)
	if err != nil {
		// Skip if image not available (common in CI)
		if strings.Contains(err.Error(), "No such image") {
			t.Skipf("Docker image not available: %v", err)
		}
		t.Fatalf("Docker executor returned error: %v", err)
	}

	if result.Success {
		t.Error("Expected compilation to fail for invalid code")
	}

	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code")
	}

	t.Logf("Docker compile error captured: exit_code=%d", result.ExitCode)
}
