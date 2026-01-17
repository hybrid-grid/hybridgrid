package fallback

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rs/zerolog/log"
)

// CompileJob represents a local compilation job.
type CompileJob struct {
	TaskID             string
	Compiler           string
	Args               []string
	PreprocessedSource []byte
	WorkDir            string
	Timeout            time.Duration
}

// CompileResult represents the result of local compilation.
type CompileResult struct {
	ObjectCode      []byte
	Stdout          string
	Stderr          string
	ExitCode        int
	CompilationTime time.Duration
	Fallback        bool
	FallbackReason  string
}

// LocalFallback handles local compilation when remote workers fail.
type LocalFallback struct {
	enabled    bool
	workDir    string
	maxTimeout time.Duration
}

// Config holds local fallback configuration.
type Config struct {
	Enabled    bool
	WorkDir    string
	MaxTimeout time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Enabled:    true,
		WorkDir:    "",
		MaxTimeout: 300 * time.Second, // 5 minutes max
	}
}

// New creates a new local fallback handler.
func New(cfg Config) *LocalFallback {
	workDir := cfg.WorkDir
	if workDir == "" {
		workDir = os.TempDir()
	}

	return &LocalFallback{
		enabled:    cfg.Enabled,
		workDir:    workDir,
		maxTimeout: cfg.MaxTimeout,
	}
}

// IsEnabled returns true if local fallback is enabled.
func (f *LocalFallback) IsEnabled() bool {
	return f.enabled
}

// Execute runs a compilation job locally.
func (f *LocalFallback) Execute(ctx context.Context, job *CompileJob) (*CompileResult, error) {
	if !f.enabled {
		return nil, fmt.Errorf("local fallback is disabled")
	}

	start := time.Now()

	// Create temp directory for this job
	taskDir, err := os.MkdirTemp(f.workDir, fmt.Sprintf("hg-fallback-%s-", job.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(taskDir)

	// Write preprocessed source
	srcFile := filepath.Join(taskDir, "source.i")
	if err := os.WriteFile(srcFile, job.PreprocessedSource, 0644); err != nil {
		return nil, fmt.Errorf("failed to write source: %w", err)
	}

	// Determine output file
	outFile := filepath.Join(taskDir, "output.o")

	// Build command arguments
	args := buildArgs(job.Args, srcFile, outFile)

	// Determine timeout
	timeout := job.Timeout
	if timeout > f.maxTimeout {
		timeout = f.maxTimeout
	}
	if timeout <= 0 {
		timeout = f.maxTimeout
	}

	// Create command with timeout context
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, job.Compiler, args...)
	cmd.Dir = taskDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Debug().
		Str("task_id", job.TaskID).
		Str("compiler", job.Compiler).
		Strs("args", args).
		Msg("Executing local fallback compilation")

	err = cmd.Run()
	compilationTime := time.Since(start)

	result := &CompileResult{
		Stdout:          stdout.String(),
		Stderr:          stderr.String(),
		CompilationTime: compilationTime,
		Fallback:        true,
		FallbackReason:  "remote workers unavailable",
	}

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		result.Stderr = fmt.Sprintf("compilation timed out after %v", timeout)
		log.Warn().
			Str("task_id", job.TaskID).
			Dur("timeout", timeout).
			Msg("Local fallback timed out")
		return result, nil
	}

	// Check exit code
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
			result.Stderr = fmt.Sprintf("exec error: %v\n%s", err, result.Stderr)
		}
		log.Debug().
			Str("task_id", job.TaskID).
			Int("exit_code", result.ExitCode).
			Msg("Local fallback compilation failed")
		return result, nil
	}

	// Read output file
	objectCode, err := os.ReadFile(outFile)
	if err != nil {
		result.ExitCode = 1
		result.Stderr = fmt.Sprintf("failed to read output: %v\n%s", err, result.Stderr)
		return result, nil
	}

	result.ObjectCode = objectCode
	result.ExitCode = 0

	log.Info().
		Str("task_id", job.TaskID).
		Dur("duration", compilationTime).
		Int("output_size", len(objectCode)).
		Msg("Local fallback compilation succeeded")

	return result, nil
}

// buildArgs constructs compiler arguments for local compilation.
func buildArgs(original []string, srcFile, outFile string) []string {
	var args []string

	// Add compile-only flag if not present
	hasCompileOnly := false
	for _, arg := range original {
		if arg == "-c" {
			hasCompileOnly = true
			break
		}
	}
	if !hasCompileOnly {
		args = append(args, "-c")
	}

	// Add original args, filtering out input/output files
	for i := 0; i < len(original); i++ {
		arg := original[i]
		if arg == "-o" {
			i++ // Skip output file
			continue
		}
		if isSourceFile(arg) {
			continue
		}
		args = append(args, arg)
	}

	// Add input and output
	args = append(args, srcFile, "-o", outFile)

	return args
}

// isSourceFile checks if an argument looks like a source file.
func isSourceFile(arg string) bool {
	if len(arg) == 0 || arg[0] == '-' {
		return false
	}

	ext := filepath.Ext(arg)
	switch ext {
	case ".c", ".cpp", ".cc", ".cxx", ".C",
		".i", ".ii", ".s", ".S",
		".m", ".mm":
		return true
	}
	return false
}
