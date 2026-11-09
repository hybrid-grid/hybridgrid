package executor

import (
	"context"
	"runtime"
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

	// Flutter artifact fields (optional, used by FlutterExecutor)
	Artifacts       []*pb.ArtifactInfo
	ArtifactArchive []byte
	ArtifactPath    string
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

	// Client info for OS-aware executor selection
	ClientOs string // OS where the build was initiated (e.g., "linux", "darwin", "windows")

	// Flutter build fields (optional, used by FlutterExecutor)
	FlutterConfig  *pb.FlutterConfig // Flutter-specific configuration
	UnityConfig    *pb.UnityConfig
	SourceArchive  []byte            // Tar archive of Flutter project source
	TargetPlatform pb.TargetPlatform // Target platform (e.g., PLATFORM_ANDROID)
	BuildType      pb.BuildType      // Build type (e.g., BUILD_TYPE_FLUTTER)
	TimeoutSeconds int32             // Timeout in seconds for Flutter builds
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
	flutter    Executor
	unity      Executor
	nativeArch pb.Architecture
}

// NewManager creates a new executor manager. dockerImages overrides the
// default per-arch Docker images (nil keeps the dockcross defaults);
// dockerCCs overrides the compiler command inside those images per arch
// (nil keeps the client-requested compiler).
func NewManager(nativeArch pb.Architecture, dockerAvailable bool, dockerImages map[pb.Architecture]string, dockerCCs map[pb.Architecture]string) *Manager {
	m := &Manager{
		nativeArch: nativeArch,
		native:     NewNativeExecutor(),
	}

	if dockerAvailable {
		docker, err := NewDockerExecutor()
		if err == nil {
			for arch, image := range dockerImages {
				docker.SetImage(arch, image)
			}
			for arch, cc := range dockerCCs {
				docker.SetToolchain(arch, cc)
			}
			m.docker = docker
		}
	}

	// Try to initialize MSVC executor on Windows. Split per platform because
	// off Windows the constructor can only fail, so the attempt is dead code
	// there (staticcheck SA4023).
	m.initMSVC()

	// Initialize Flutter executor
	m.flutter = NewFlutterExecutor()
	m.unity = NewUnityExecutor()

	return m
}

// SetDockerToolchain overrides the compiler command used inside the
// docker executor's image for a target architecture (no-op when the
// docker executor is unavailable).
func (m *Manager) SetDockerToolchain(arch pb.Architecture, cc string) {
	if d, ok := m.docker.(*DockerExecutor); ok {
		d.SetToolchain(arch, cc)
	}
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

// SelectForRequest chooses the executor considering client OS.
// If the client OS differs from this worker's OS and raw source is provided,
// use Docker to compile in a matching environment.
func (m *Manager) SelectForRequest(req *Request) Executor {
	// Route Flutter build requests to FlutterExecutor
	if req.BuildType == pb.BuildType_BUILD_TYPE_FLUTTER {
		return m.flutter
	}

	if req.BuildType == pb.BuildType_BUILD_TYPE_UNITY {
		return m.unity
	}

	// If client OS is set and differs from this worker's OS,
	// raw source needs Docker for cross-OS compilation
	if req.ClientOs != "" && req.ClientOs != runtime.GOOS && len(req.RawSource) > 0 {
		if m.docker != nil && m.docker.CanExecute(req.TargetArch, m.nativeArch) {
			return m.docker
		}
	}

	// Otherwise use standard arch-based selection
	return m.Select(req.TargetArch)
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

// Close releases resources held by managed executors.
func (m *Manager) Close() error {
	if m.docker == nil {
		return nil
	}

	if d, ok := m.docker.(interface{ Close() error }); ok {
		return d.Close()
	}
	return nil
}

// GetMSVC returns the MSVC executor if available.
func (m *Manager) GetMSVC() Executor {
	return m.msvc
}

// GetFlutter returns the Flutter executor.
func (m *Manager) GetFlutter() Executor {
	return m.flutter
}

func (m *Manager) GetUnity() Executor {
	return m.unity
}

// isMSVCCompiler checks if the compiler is MSVC.
func isMSVCCompiler(compiler string) bool {
	lower := strings.ToLower(compiler)
	return strings.Contains(lower, "cl.exe") || lower == "cl"
}

// Execute runs a compilation using the appropriate executor.
func (m *Manager) Execute(ctx context.Context, req *Request) (*Result, error) {
	executor := m.SelectForRequest(req)

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
