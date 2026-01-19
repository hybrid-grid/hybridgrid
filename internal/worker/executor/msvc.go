//go:build windows

package executor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/capability"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
)

// MSVCExecutor executes compilation using Microsoft Visual C++ (cl.exe).
type MSVCExecutor struct {
	msvcInfo   *capability.MSVCInfo
	env        []string
	clExePath  string
	targetArch string // x64, x86, arm64
}

// NewMSVCExecutor creates a new MSVC executor.
func NewMSVCExecutor() (*MSVCExecutor, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("MSVC executor only available on Windows")
	}

	msvc := capability.DetectMSVC()
	if msvc == nil || !msvc.Available {
		return nil, fmt.Errorf("MSVC not found on this system")
	}

	// Default to x64 target
	targetArch := "x64"
	clExePath := msvc.GetClExeForArch(targetArch)
	if clExePath == "" {
		// Try x86 if x64 not available
		targetArch = "x86"
		clExePath = msvc.GetClExeForArch(targetArch)
		if clExePath == "" {
			return nil, fmt.Errorf("cl.exe not found for any architecture")
		}
	}

	// Setup environment from vcvars
	env, err := setupVCVarsEnv(msvc, targetArch)
	if err != nil {
		return nil, fmt.Errorf("failed to setup MSVC environment: %w", err)
	}

	return &MSVCExecutor{
		msvcInfo:   msvc,
		env:        env,
		clExePath:  clExePath,
		targetArch: targetArch,
	}, nil
}

// Name returns the executor name.
func (e *MSVCExecutor) Name() string {
	return "msvc"
}

// CanExecute returns true if this executor can handle the target architecture.
func (e *MSVCExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	// MSVC can compile for Windows targets
	switch targetArch {
	case pb.Architecture_ARCH_X86_64:
		return contains(e.msvcInfo.Architectures, "x64")
	case pb.Architecture_ARCH_ARMV7:
		return contains(e.msvcInfo.Architectures, "arm")
	case pb.Architecture_ARCH_ARM64:
		return contains(e.msvcInfo.Architectures, "arm64")
	case pb.Architecture_ARCH_UNSPECIFIED:
		return true
	}
	return false
}

// Execute runs the compilation using MSVC cl.exe.
func (e *MSVCExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	// Create temp directory for this task
	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-msvc-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	var srcFile string
	var args []string

	// Determine output file path (MSVC uses .obj extension)
	outFile := filepath.Join(workDir, "output.obj")

	// Check if using raw source mode (cross-compilation) or preprocessed mode
	if len(req.RawSource) > 0 {
		// Mode 2: Raw source
		srcFile, args, err = e.setupRawSourceMode(workDir, req, outFile)
		if err != nil {
			return nil, fmt.Errorf("failed to setup raw source: %w", err)
		}
	} else {
		// Mode 1: Preprocessed source
		srcFile = filepath.Join(workDir, "source.i")
		if err := os.WriteFile(srcFile, req.PreprocessedSource, 0644); err != nil {
			return nil, fmt.Errorf("failed to write source: %w", err)
		}
		args = e.buildArgs(req.Args, srcFile, outFile)
	}

	// Create command with context for timeout
	cmd := exec.CommandContext(ctx, e.clExePath, args...)
	cmd.Dir = workDir
	cmd.Env = e.env

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

// buildArgs constructs MSVC compiler arguments, translating from GCC if needed.
func (e *MSVCExecutor) buildArgs(originalArgs []string, srcFile, outFile string) []string {
	// First, translate GCC flags to MSVC flags
	translated := compiler.TranslateToMSVC(originalArgs)

	// Add standard MSVC flags
	args := compiler.MSVCToCLFlags(translated)

	// Remove any input/output files from translated args
	filtered := make([]string, 0, len(args))
	skipNext := false
	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if strings.HasPrefix(arg, "/Fo") || strings.HasPrefix(arg, "/Fe") {
			continue // Skip output files, we'll add our own
		}
		if arg == "/c" {
			continue // We'll add this
		}
		if isInputFile(arg) {
			continue // Skip input files
		}
		filtered = append(filtered, arg)
	}

	// Build final args
	result := make([]string, 0, len(filtered)+4)
	result = append(result, "/c") // Compile only
	result = append(result, filtered...)
	result = append(result, srcFile)
	result = append(result, "/Fo"+outFile)

	return result
}

// setupRawSourceMode sets up the working directory for raw source compilation with MSVC.
func (e *MSVCExecutor) setupRawSourceMode(workDir string, req *Request, outFile string) (string, []string, error) {
	// Determine source filename
	srcFilename := req.SourceFilename
	if srcFilename == "" {
		srcFilename = "source.c"
	}

	// Create includes directory
	includesDir := filepath.Join(workDir, "includes")
	if err := os.MkdirAll(includesDir, 0755); err != nil {
		return "", nil, fmt.Errorf("failed to create includes dir: %w", err)
	}

	// Write bundled include files
	for path, content := range req.IncludeFiles {
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

	// Build MSVC arguments
	args := make([]string, 0, len(req.Args)+10)

	// Add compile-only flag
	args = append(args, "/c")

	// Add MSVC standard flags
	args = append(args, "/nologo", "/EHsc")

	// Add bundled includes directory first
	if len(req.IncludeFiles) > 0 {
		args = append(args, "/I"+includesDir)
	}

	// Translate and add original arguments
	translated := compiler.TranslateToMSVC(req.Args)
	for _, arg := range translated {
		// Skip compile-only and output flags (we handle them)
		if arg == "/c" || strings.HasPrefix(arg, "/Fo") || strings.HasPrefix(arg, "/Fe") {
			continue
		}
		if isInputFile(arg) {
			continue
		}
		args = append(args, arg)
	}

	// Add source file
	args = append(args, srcFile)

	// Add output file
	args = append(args, "/Fo"+outFile)

	return srcFile, args, nil
}

// validMSVCArchitectures are the only allowed architecture values.
var validMSVCArchitectures = []string{"x64", "x86", "arm64", "arm"}

// SetTargetArch sets the target architecture for MSVC compilation.
func (e *MSVCExecutor) SetTargetArch(arch string) error {
	// Validate architecture parameter
	archLower := strings.ToLower(arch)
	if !contains(validMSVCArchitectures, archLower) {
		return fmt.Errorf("invalid architecture %q: must be one of %v", arch, validMSVCArchitectures)
	}

	clExePath := e.msvcInfo.GetClExeForArch(archLower)
	if clExePath == "" {
		return fmt.Errorf("architecture %s not available", arch)
	}

	env, err := setupVCVarsEnv(e.msvcInfo, arch)
	if err != nil {
		return fmt.Errorf("failed to setup environment for %s: %w", arch, err)
	}

	e.targetArch = arch
	e.clExePath = clExePath
	e.env = env
	return nil
}

// allowedVSBasePaths are the only directories from which vcvars can be executed.
var allowedVSBasePaths = []string{
	`C:\Program Files\Microsoft Visual Studio`,
	`C:\Program Files (x86)\Microsoft Visual Studio`,
}

// isAllowedVSPath validates that a path is within allowed VS installation directories.
func isAllowedVSPath(path string) bool {
	cleanPath := filepath.Clean(path)
	for _, base := range allowedVSBasePaths {
		if strings.HasPrefix(strings.ToLower(cleanPath), strings.ToLower(base)) {
			return true
		}
	}
	return false
}

// setupVCVarsEnv extracts environment variables from vcvars batch file.
func setupVCVarsEnv(msvc *capability.MSVCInfo, targetArch string) ([]string, error) {
	// Validate InstallDir is in allowed VS paths (prevent path injection)
	if !isAllowedVSPath(msvc.InstallDir) {
		return nil, fmt.Errorf("VS installation path %q is not in allowed directories", msvc.InstallDir)
	}

	// Get vcvars script path
	var vcvarsScript string
	switch targetArch {
	case "x64":
		vcvarsScript = filepath.Join(msvc.InstallDir, "VC", "Auxiliary", "Build", "vcvars64.bat")
	case "x86":
		vcvarsScript = filepath.Join(msvc.InstallDir, "VC", "Auxiliary", "Build", "vcvars32.bat")
	case "arm64":
		vcvarsScript = filepath.Join(msvc.InstallDir, "VC", "Auxiliary", "Build", "vcvarsamd64_arm64.bat")
	default:
		vcvarsScript = filepath.Join(msvc.InstallDir, "VC", "Auxiliary", "Build", "vcvars64.bat")
	}

	// Double-check the constructed path is still within allowed directories
	if !isAllowedVSPath(vcvarsScript) {
		return nil, fmt.Errorf("vcvars path %q escapes allowed directories", vcvarsScript)
	}

	if _, err := os.Stat(vcvarsScript); err != nil {
		return nil, fmt.Errorf("vcvars script not found: %s", vcvarsScript)
	}

	// Run vcvars and capture environment
	// cmd /c "vcvars64.bat && set"
	cmd := exec.Command("cmd", "/c", vcvarsScript, "&&", "set")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to run vcvars: %w", err)
	}

	// Parse KEY=VALUE lines
	var env []string
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		line = strings.TrimPrefix(line, "\r")
		if strings.Contains(line, "=") && !strings.HasPrefix(line, "=") {
			env = append(env, line)
		}
	}

	if len(env) == 0 {
		return nil, fmt.Errorf("no environment variables captured from vcvars")
	}

	return env, nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, val) {
			return true
		}
	}
	return false
}
