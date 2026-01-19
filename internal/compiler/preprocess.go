package compiler

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Preprocessor handles local C/C++ preprocessing.
type Preprocessor struct {
	defaultTimeout time.Duration
}

// PreprocessorConfig holds preprocessor configuration.
type PreprocessorConfig struct {
	Timeout time.Duration
}

// DefaultPreprocessorConfig returns sensible defaults.
func DefaultPreprocessorConfig() PreprocessorConfig {
	return PreprocessorConfig{
		Timeout: 60 * time.Second,
	}
}

// NewPreprocessor creates a new preprocessor.
func NewPreprocessor(cfg PreprocessorConfig) *Preprocessor {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	return &Preprocessor{
		defaultTimeout: timeout,
	}
}

// PreprocessResult holds the result of preprocessing.
type PreprocessResult struct {
	// PreprocessedSource is the fully expanded source code
	PreprocessedSource []byte
	// Warnings from the preprocessor (if any)
	Warnings string
}

// Preprocess runs the C/C++ preprocessor on the source file.
// It expands all #include directives and macros, producing a self-contained source.
func (p *Preprocessor) Preprocess(ctx context.Context, args *ParsedArgs, sourceFile string) (*PreprocessResult, error) {
	if args == nil {
		return nil, fmt.Errorf("parsed args cannot be nil")
	}

	// Determine compiler to use for preprocessing
	compiler := args.Compiler
	if compiler == "" {
		compiler = "gcc"
	}

	// Build preprocessing command
	cmdArgs := p.buildPreprocessArgs(args, sourceFile)

	// Create command with timeout
	timeout := p.defaultTimeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(execCtx, compiler, cmdArgs...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run preprocessor
	err := cmd.Run()

	// Check for timeout
	if execCtx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("preprocessing timed out after %v", timeout)
	}

	// Check for errors
	if err != nil {
		return nil, &PreprocessError{
			SourceFile: sourceFile,
			Stderr:     stderr.String(),
			Err:        err,
		}
	}

	return &PreprocessResult{
		PreprocessedSource: stdout.Bytes(),
		Warnings:           extractWarnings(stderr.String()),
	}, nil
}

// buildPreprocessArgs constructs the preprocessor command line.
func (p *Preprocessor) buildPreprocessArgs(args *ParsedArgs, sourceFile string) []string {
	cmdArgs := []string{"-E"} // Preprocess only

	// Add include paths
	for _, inc := range args.IncludeDirs {
		cmdArgs = append(cmdArgs, "-I"+inc)
	}

	// Add defines
	for _, def := range args.Defines {
		cmdArgs = append(cmdArgs, "-D"+def)
	}

	// Add language standard if specified
	if args.Standard != "" {
		cmdArgs = append(cmdArgs, "-std="+args.Standard)
	}

	// Add language flag if specified (for .h files or ambiguous extensions)
	if args.Language != "" {
		cmdArgs = append(cmdArgs, "-x", args.Language)
	}

	// Add other flags that affect preprocessing
	for _, flag := range args.Flags {
		// Include flags that affect preprocessing
		if isPreprocessingFlag(flag) {
			cmdArgs = append(cmdArgs, flag)
		}
	}

	// Add source file
	cmdArgs = append(cmdArgs, sourceFile)

	return cmdArgs
}

// isPreprocessingFlag returns true if the flag affects preprocessing.
func isPreprocessingFlag(flag string) bool {
	// Flags that affect preprocessing behavior
	prefixes := []string{
		"-U",         // Undefine macro
		"-include",   // Force include
		"-imacros",   // Include macros only
		"-isystem",   // System include path
		"-idirafter", // Include path (after -I)
		"-iprefix",   // Prefix for -iwithprefix
		"-nostdinc",  // Don't search standard include dirs
		"-trigraphs", // Enable trigraphs
		"-fno-",      // Various -fno- flags
		"-f",         // Various -f flags that might affect preprocessing
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(flag, prefix) {
			return true
		}
	}

	// Exact matches
	exactFlags := map[string]bool{
		"-nostdinc":    true,
		"-nostdinc++":  true,
		"-trigraphs":   true,
		"-ansi":        true,
		"-traditional": true,
	}

	return exactFlags[flag]
}

// extractWarnings extracts warning messages from stderr.
func extractWarnings(stderr string) string {
	var warnings []string
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && strings.Contains(line, "warning:") {
			warnings = append(warnings, line)
		}
	}
	return strings.Join(warnings, "\n")
}

// PreprocessError represents a preprocessing failure.
type PreprocessError struct {
	SourceFile string
	Stderr     string
	Err        error
}

func (e *PreprocessError) Error() string {
	// Format a helpful error message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("preprocessing failed for %s\n", e.SourceFile))

	// Parse stderr for specific errors
	if e.Stderr != "" {
		// Look for common error patterns
		lines := strings.Split(e.Stderr, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Include error lines (not just warnings)
			if strings.Contains(line, "error:") || strings.Contains(line, "fatal error:") {
				sb.WriteString("  ")
				sb.WriteString(line)
				sb.WriteString("\n")
			}
		}

		// Add hint for missing headers
		if strings.Contains(e.Stderr, "file not found") || strings.Contains(e.Stderr, "No such file") {
			sb.WriteString("  Hint: Check include paths with -I flag\n")
		}
	}

	return sb.String()
}

func (e *PreprocessError) Unwrap() error {
	return e.Err
}
