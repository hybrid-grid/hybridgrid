package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// NativeExecutor executes compilation directly on the host system.
type NativeExecutor struct{}

// NewNativeExecutor creates a new native executor.
func NewNativeExecutor() *NativeExecutor {
	return &NativeExecutor{}
}

// Name returns the executor name.
func (e *NativeExecutor) Name() string {
	return "native"
}

// CanExecute returns true if this executor can handle the target architecture.
func (e *NativeExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	// Native executor can only compile for native architecture
	return targetArch == nativeArch || targetArch == pb.Architecture_ARCH_UNSPECIFIED
}

// Execute runs the compilation command directly.
func (e *NativeExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	// Create temp directory for this task
	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-worker-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	var srcFile string
	var args []string

	// Determine output file path
	outFile := filepath.Join(workDir, "output.o")

	// Check if using raw source mode (cross-compilation) or preprocessed mode
	if len(req.RawSource) > 0 {
		// Mode 2: Raw source - need to set up includes and compile from scratch
		// srcFile is embedded in args, so we discard it
		_, args, err = e.setupRawSourceMode(workDir, req, outFile)
		if err != nil {
			return nil, fmt.Errorf("failed to setup raw source: %w", err)
		}
	} else {
		// Mode 1: Preprocessed source (legacy)
		srcFile = filepath.Join(workDir, "source.i")
		if err := os.WriteFile(srcFile, req.PreprocessedSource, 0644); err != nil {
			return nil, fmt.Errorf("failed to write source: %w", err)
		}
		args = e.buildArgs(req.Args, srcFile, outFile)
	}

	// Create command with context for timeout
	cmd := exec.CommandContext(ctx, req.Compiler, args...)
	cmd.Dir = workDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run compilation
	err = cmd.Run()
	compilationTime := time.Since(start)

	result := &Result{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		CompilationTime: compilationTime,
	}

	// Check for context cancellation (timeout)
	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		result.Success = false
		return result, nil
	}

	// Get exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = int32(exitErr.ExitCode())
		} else {
			return nil, fmt.Errorf("execution failed: %w", err)
		}
		result.Success = false
		return result, nil
	}

	// Read output file
	objectCode, err := os.ReadFile(outFile)
	if err != nil {
		result.Stderr += fmt.Sprintf("\nFailed to read output: %v", err)
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.ObjectCode = objectCode
	result.ExitCode = 0

	return result, nil
}

// buildArgs constructs compiler arguments, replacing input/output paths.
func (e *NativeExecutor) buildArgs(originalArgs []string, srcFile, outFile string) []string {
	args := make([]string, 0, len(originalArgs)+4)

	// Add compile-only flag if not present
	hasCompileOnly := false
	hasOutput := false
	skipNext := false

	for i, arg := range originalArgs {
		if skipNext {
			skipNext = false
			continue
		}

		switch arg {
		case "-c":
			hasCompileOnly = true
			args = append(args, arg)
		case "-o":
			hasOutput = true
			skipNext = true // Skip the next arg (output file)
		default:
			// Skip input files (we'll add our own)
			if !isInputFile(arg) {
				args = append(args, arg)
			}
		}
		_ = i // suppress unused variable warning
	}

	if !hasCompileOnly {
		args = append(args, "-c")
	}

	// Add input file
	args = append(args, srcFile)

	// Add output file
	if !hasOutput {
		args = append(args, "-o", outFile)
	} else {
		args = append(args, "-o", outFile)
	}

	return args
}

// isInputFile checks if an argument looks like an input source file.
func isInputFile(arg string) bool {
	if len(arg) == 0 || arg[0] == '-' {
		return false
	}

	ext := filepath.Ext(arg)
	switch ext {
	case ".c", ".cc", ".cpp", ".cxx", ".C", ".i", ".ii", ".s", ".S":
		return true
	}
	return false
}

// setupRawSourceMode sets up the working directory for raw source compilation.
// This mode is used for cross-compilation where source is not preprocessed.
func (e *NativeExecutor) setupRawSourceMode(workDir string, req *Request, outFile string) (string, []string, error) {
	// Determine source filename
	srcFilename := req.SourceFilename
	if srcFilename == "" {
		srcFilename = "source.c" // Default to C
	}

	// Create includes directory
	includesDir := filepath.Join(workDir, "includes")
	if err := os.MkdirAll(includesDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create includes dir: %w", err)
	}

	// Write bundled include files
	for path, content := range req.IncludeFiles {
		// Create subdirectories as needed
		fullPath := filepath.Join(includesDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return "", nil, fmt.Errorf("failed to create include dir for %s: %w", path, err)
		}
		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return "", nil, fmt.Errorf("failed to write include %s: %w", path, err)
		}
	}

	// Write source file
	srcFile := filepath.Join(workDir, srcFilename)
	if err := os.WriteFile(srcFile, req.RawSource, 0644); err != nil {
		return "", nil, fmt.Errorf("failed to write source: %w", err)
	}

	// Build arguments
	args := make([]string, 0, len(req.Args)+10)

	// Add compile-only flag
	args = append(args, "-c")

	// Add bundled includes directory first (highest priority)
	if len(req.IncludeFiles) > 0 {
		args = append(args, "-I"+includesDir)
	}

	// Add original arguments (skip -c, -o, input files, and -I paths we'll handle)
	skipNext := false
	for _, arg := range req.Args {
		if skipNext {
			skipNext = false
			continue
		}
		switch {
		case arg == "-c":
			// Already added
		case arg == "-o":
			skipNext = true // Skip output file
		case isInputFile(arg):
			// Skip input files
		default:
			args = append(args, arg)
		}
	}

	// Add source file
	args = append(args, srcFile)

	// Add output file
	args = append(args, "-o", outFile)

	return srcFile, args, nil
}
