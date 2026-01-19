package executor

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
)

// Result represents the outcome of a compilation execution.
type Result struct {
	Success         bool
	ObjectCode      []byte
	Stdout          string
	Stderr          string
	ExitCode        int32
	CompilationTime time.Duration
}

// Request represents a compilation request to an executor.
type Request struct {
	TaskID             string
	Compiler           string
	Args               []string
	PreprocessedSource []byte // Mode 1: Already preprocessed source
	TargetArch         pb.Architecture
	Timeout            time.Duration

	// Cross-compilation mode (Mode 2)
	RawSource      []byte            // Raw source file (not preprocessed)
	SourceFilename string            // Original filename with extension (e.g., "main.cpp")
	IncludeFiles   map[string][]byte // Bundled project headers (path -> content)
	IncludePaths   []string          // -I paths for headers
}

// Executor defines the interface for compilation executors.
type Executor interface {
	// Execute runs a compilation task and returns the result.
	Execute(ctx context.Context, req *Request) (*Result, error)

	// Name returns the executor name for logging.
	Name() string

	// CanExecute returns true if this executor can handle the given architecture.
	CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool
}

// Manager manages multiple executors and selects the appropriate one.
type Manager struct {
	native     Executor
	docker     Executor
	msvc       Executor
	nativeArch pb.Architecture
}

// NewManager creates a new executor manager.
func NewManager(nativeArch pb.Architecture, dockerAvailable bool) *Manager {
	m := &Manager{
		nativeArch: nativeArch,
		native:     NewNativeExecutor(),
	}

	if dockerAvailable {
		docker, err := NewDockerExecutor()
		if err == nil {
			m.docker = docker
		}
	}

	// Try to initialize MSVC executor on Windows
	msvc, err := NewMSVCExecutor()
	if err == nil {
		m.msvc = msvc
	}

	return m
}

// Select chooses the appropriate executor for the given target architecture.
func (m *Manager) Select(targetArch pb.Architecture) Executor {
	// If target matches native arch, use native executor
	if targetArch == m.nativeArch || targetArch == pb.Architecture_ARCH_UNSPECIFIED {
		return m.native
	}

	// For cross-compilation, use Docker if available
	if m.docker != nil && m.docker.CanExecute(targetArch, m.nativeArch) {
		return m.docker
	}

	// Fall back to native (might fail, but let it try)
	return m.native
}

// SelectForCompiler chooses the appropriate executor based on compiler and target architecture.
func (m *Manager) SelectForCompiler(compiler string, targetArch pb.Architecture) Executor {
	// If compiler is MSVC (cl.exe), use MSVC executor
	if m.msvc != nil && isMSVCCompiler(compiler) {
		if m.msvc.CanExecute(targetArch, m.nativeArch) {
			return m.msvc
		}
	}

	// Otherwise use standard selection
	return m.Select(targetArch)
}

// GetMSVC returns the MSVC executor if available.
func (m *Manager) GetMSVC() Executor {
	return m.msvc
}

// isMSVCCompiler checks if the compiler is MSVC.
func isMSVCCompiler(compiler string) bool {
	lower := strings.ToLower(compiler)
	return strings.Contains(lower, "cl.exe") || lower == "cl"
}

// Execute runs a compilation using the appropriate executor.
func (m *Manager) Execute(ctx context.Context, req *Request) (*Result, error) {
	executor := m.Select(req.TargetArch)

	// Start tracing span
	ctx, span := tracing.StartSpan(ctx, "compile",
		trace.WithAttributes(
			tracing.AttrTaskID.String(req.TaskID),
			tracing.AttrCompiler.String(req.Compiler),
			tracing.AttrSourceFile.String(req.SourceFilename),
			tracing.AttrTargetArch.String(req.TargetArch.String()),
			attribute.String("executor", executor.Name()),
		),
	)
	defer span.End()

	// Execute compilation
	startTime := time.Now()
	result, err := executor.Execute(ctx, req)
	duration := time.Since(startTime)

	// Record result attributes
	if result != nil {
		span.SetAttributes(
			tracing.AttrExitCode.Int(int(result.ExitCode)),
			tracing.AttrDurationMs.Int64(duration.Milliseconds()),
			tracing.AttrObjectSize.Int(len(result.ObjectCode)),
		)

		if !result.Success {
			span.RecordError(err)
		}
	}

	if err != nil {
		span.RecordError(err)
	}

	return result, err
}
