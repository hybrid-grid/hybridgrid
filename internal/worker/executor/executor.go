package executor

import (
	"context"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
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
	PreprocessedSource []byte
	TargetArch         pb.Architecture
	Timeout            time.Duration
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

// Execute runs a compilation using the appropriate executor.
func (m *Manager) Execute(ctx context.Context, req *Request) (*Result, error) {
	executor := m.Select(req.TargetArch)
	return executor.Execute(ctx, req)
}
