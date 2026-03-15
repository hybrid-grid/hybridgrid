package executor

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// TestNativeExecutor_Execute_RawSourceMode tests raw source compilation
func TestNativeExecutor_Execute_RawSourceMode(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	e := NewNativeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	source := []byte(`int add(int a, int b) { return a + b; }`)

	req := &Request{
		TaskID:         "test-raw-001",
		Compiler:       "gcc",
		Args:           []string{"-O2", "-Wall"},
		RawSource:      source,
		SourceFilename: "test.c",
		TargetArch:     pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:        30 * time.Second,
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

// TestNativeExecutor_Execute_RawSourceWithIncludes tests raw source with bundled includes
func TestNativeExecutor_Execute_RawSourceWithIncludes(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	e := NewNativeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	source := []byte(`#include "math_ops.h"
int main() { return add(1, 2); }`)

	includeFiles := map[string][]byte{
		"math_ops.h": []byte(`int add(int a, int b);`),
	}

	req := &Request{
		TaskID:         "test-raw-002",
		Compiler:       "gcc",
		Args:           []string{"-O2"},
		RawSource:      source,
		SourceFilename: "main.c",
		IncludeFiles:   includeFiles,
		TargetArch:     pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:        30 * time.Second,
	}

	result, err := e.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Success {
		t.Errorf("Compilation failed: stderr=%s", result.Stderr)
	}

	if len(result.ObjectCode) == 0 {
		t.Error("Expected non-empty object code")
	}
}

// TestNativeExecutor_setupRawSourceMode tests the raw source setup logic
func TestNativeExecutor_setupRawSourceMode(t *testing.T) {
	e := NewNativeExecutor()

	tests := []struct {
		name         string
		req          *Request
		wantFilename string
		wantArgsHas  []string
	}{
		{
			name: "c source with includes",
			req: &Request{
				TaskID:         "test-setup-001",
				Compiler:       "gcc",
				Args:           []string{"-O2", "-Wall"},
				RawSource:      []byte("int main() {}"),
				SourceFilename: "main.c",
				IncludeFiles: map[string][]byte{
					"header.h": []byte("#define VALUE 42"),
				},
			},
			wantFilename: "main.c",
			wantArgsHas:  []string{"-c", "-O2", "-Wall", "-o"},
		},
		{
			name: "cpp source",
			req: &Request{
				TaskID:         "test-setup-002",
				Compiler:       "g++",
				Args:           []string{"-std=c++17"},
				RawSource:      []byte("int main() {}"),
				SourceFilename: "main.cpp",
			},
			wantFilename: "main.cpp",
			wantArgsHas:  []string{"-c", "-std=c++17"},
		},
		{
			name: "no filename defaults to source.c",
			req: &Request{
				TaskID:    "test-setup-003",
				Compiler:  "gcc",
				Args:      []string{},
				RawSource: []byte("int main() {}"),
			},
			wantFilename: "source.c",
			wantArgsHas:  []string{"-c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			outFile := "output.o"

			srcFile, args, err := e.setupRawSourceMode(workDir, tt.req, outFile)
			if err != nil {
				t.Fatalf("setupRawSourceMode() error = %v", err)
			}

			if srcFile == "" {
				t.Error("Expected non-empty srcFile")
			}

			for _, want := range tt.wantArgsHas {
				found := false
				for _, arg := range args {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("setupRawSourceMode() args missing %q in %v", want, args)
				}
			}
		})
	}
}

// TestNativeExecutor_Execute_CommandNotFound tests behavior when compiler not found
func TestNativeExecutor_Execute_CommandNotFound(t *testing.T) {
	e := NewNativeExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &Request{
		TaskID:             "test-not-found",
		Compiler:           "nonexistent-compiler-xyz",
		Args:               []string{"-c"},
		PreprocessedSource: []byte("int main() {}"),
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:            5 * time.Second,
	}

	_, err := e.Execute(ctx, req)
	if err == nil {
		t.Error("Expected error for nonexistent compiler")
	}
}

func TestNativeExecutor_buildArgs_ReplacesOutputAndSkipsInputFiles(t *testing.T) {
	e := NewNativeExecutor()
	args := e.buildArgs([]string{"-O2", "-o", "old.o", "input.c", "-Wall"}, "source.i", "output.o")

	for _, want := range []string{"-c", "-O2", "-Wall", "source.i", "-o", "output.o"} {
		found := false
		for _, arg := range args {
			if arg == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("buildArgs() missing %q in %v", want, args)
		}
	}

	for _, notWant := range []string{"old.o", "input.c"} {
		for _, arg := range args {
			if arg == notWant {
				t.Fatalf("buildArgs() unexpectedly contains %q in %v", notWant, args)
			}
		}
	}
}

func TestNativeExecutor_setupRawSourceMode_WritesSourceAndIncludes(t *testing.T) {
	e := NewNativeExecutor()
	workDir := t.TempDir()
	req := &Request{
		TaskID:         "setup-raw-files",
		Compiler:       "gcc",
		Args:           []string{"-O2", "-I", "ignored", "-Ialso-ignored"},
		RawSource:      []byte("#include \"nested/header.h\"\nint main() { return 0; }"),
		SourceFilename: "main.c",
		IncludeFiles: map[string][]byte{
			"nested/header.h": []byte("#define VALUE 1"),
		},
	}

	srcFile, args, err := e.setupRawSourceMode(workDir, req, "output.o")
	if err != nil {
		t.Fatalf("setupRawSourceMode() error = %v", err)
	}

	if _, err := os.Stat(srcFile); err != nil {
		t.Fatalf("source file not written: %v", err)
	}

	includePath := filepath.Join(workDir, "includes", "nested", "header.h")
	content, err := os.ReadFile(includePath)
	if err != nil {
		t.Fatalf("include file not written: %v", err)
	}
	if string(content) != "#define VALUE 1" {
		t.Fatalf("include file content = %q", string(content))
	}

	hasBundledInclude := false
	for _, arg := range args {
		if arg == "-I"+filepath.Join(workDir, "includes") {
			hasBundledInclude = true
			break
		}
	}
	if !hasBundledInclude {
		t.Fatalf("expected bundled include path in args: %v", args)
	}
}

func TestNativeExecutor_Execute_TimeoutReturnsFailedResult(t *testing.T) {
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping test")
	}

	e := NewNativeExecutor()
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
	defer cancel()

	result, err := e.Execute(ctx, &Request{
		TaskID:             "timeout-result",
		Compiler:           "gcc",
		Args:               []string{"-c"},
		PreprocessedSource: []byte("int main() { return 0; }"),
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		Timeout:            time.Second,
	})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result.Success {
		t.Fatal("Execute() success = true, want false")
	}
	if result.ExitCode != -1 {
		t.Fatalf("Execute() exit code = %d, want -1", result.ExitCode)
	}
}
