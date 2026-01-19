package server

import (
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/capability"
	"github.com/h3nr1-d14z/hybridgrid/internal/worker/executor"
)

// Config holds the worker gRPC server configuration.
type Config struct {
	Port           int
	MaxConcurrent  int
	DefaultTimeout time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:           50052,
		MaxConcurrent:  4,
		DefaultTimeout: 120 * time.Second,
	}
}

// Server implements the worker gRPC server.
type Server struct {
	pb.UnimplementedBuildServiceServer

	config       Config
	server       *grpc.Server
	executor     *executor.Manager
	capabilities *pb.WorkerCapabilities

	activeTasks  int64
	totalTasks   int64
	successTasks int64
	failedTasks  int64
	totalTimeMs  int64
}

// New creates a new worker gRPC server.
func New(cfg Config) *Server {
	caps := capability.Detect()
	// Set max parallel tasks from config
	caps.MaxParallelTasks = int32(cfg.MaxConcurrent)

	return &Server{
		config:       cfg,
		capabilities: caps,
		executor:     executor.NewManager(caps.NativeArch, caps.DockerAvailable),
	}
}

// Capabilities returns the worker's capabilities.
func (s *Server) Capabilities() *pb.WorkerCapabilities {
	return s.capabilities
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterBuildServiceServer(s.server, s)

	log.Info().Int("port", s.config.Port).Msg("Worker gRPC server starting")
	return s.server.Serve(lis)
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Handshake is not used by workers (they connect to coordinator).
func (s *Server) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "workers don't handle handshakes")
}

// Compile handles compilation requests.
func (s *Server) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	start := time.Now()

	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// Check concurrency limit
	active := atomic.AddInt64(&s.activeTasks, 1)
	defer atomic.AddInt64(&s.activeTasks, -1)

	if active > int64(s.config.MaxConcurrent) {
		return nil, status.Error(codes.ResourceExhausted, "too many concurrent tasks")
	}

	atomic.AddInt64(&s.totalTasks, 1)

	log.Debug().
		Str("task_id", req.TaskId).
		Str("compiler", req.Compiler).
		Str("target_arch", req.TargetArch.String()).
		Msg("Starting compilation")

	// Determine timeout
	timeout := s.config.DefaultTimeout
	if req.TimeoutSeconds > 0 {
		timeout = time.Duration(req.TimeoutSeconds) * time.Second
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Build executor request
	execReq := &executor.Request{
		TaskID:             req.TaskId,
		Compiler:           req.Compiler,
		Args:               req.CompilerArgs,
		PreprocessedSource: req.PreprocessedSource,
		TargetArch:         req.TargetArch,
		Timeout:            timeout,
		// Cross-compilation mode fields
		RawSource:      req.RawSource,
		SourceFilename: req.SourceFilename,
		IncludeFiles:   req.IncludeFiles,
		IncludePaths:   req.IncludePaths,
	}

	// Execute compilation
	result, err := s.executor.Execute(execCtx, execReq)
	if err != nil {
		atomic.AddInt64(&s.failedTasks, 1)
		log.Error().Err(err).Str("task_id", req.TaskId).Msg("Compilation execution error")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("execution error: %v", err),
		}, nil
	}

	compilationTime := time.Since(start)
	atomic.AddInt64(&s.totalTimeMs, compilationTime.Milliseconds())

	resp := &pb.CompileResponse{
		ObjectFile:        result.ObjectCode,
		Stdout:            result.Stdout,
		Stderr:            result.Stderr,
		ExitCode:          result.ExitCode,
		CompilationTimeMs: compilationTime.Milliseconds(),
	}

	if result.Success {
		resp.Status = pb.TaskStatus_STATUS_COMPLETED
		atomic.AddInt64(&s.successTasks, 1)
		log.Debug().
			Str("task_id", req.TaskId).
			Dur("duration", compilationTime).
			Int("output_size", len(result.ObjectCode)).
			Msg("Compilation succeeded")
	} else {
		resp.Status = pb.TaskStatus_STATUS_FAILED
		atomic.AddInt64(&s.failedTasks, 1)
		log.Debug().
			Str("task_id", req.TaskId).
			Dur("duration", compilationTime).
			Int32("exit_code", result.ExitCode).
			Msg("Compilation failed")
	}

	return resp, nil
}

// Build is not implemented for workers (coordinator handles it).
func (s *Server) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	return nil, status.Error(codes.Unimplemented, "use Compile for compilation tasks")
}

// StreamBuild is not implemented for workers.
func (s *Server) StreamBuild(stream pb.BuildService_StreamBuildServer) error {
	return status.Error(codes.Unimplemented, "stream build not supported by workers")
}

// HealthCheck returns worker health status.
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	active := atomic.LoadInt64(&s.activeTasks)
	healthy := active < int64(s.config.MaxConcurrent)

	return &pb.HealthResponse{
		Healthy:     healthy,
		ActiveTasks: int32(active),
		QueuedTasks: 0,
	}, nil
}

// GetWorkerStatus returns this worker's status.
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	total := atomic.LoadInt64(&s.totalTasks)
	success := atomic.LoadInt64(&s.successTasks)
	avgTime := int64(0)
	if success > 0 {
		avgTime = atomic.LoadInt64(&s.totalTimeMs) / success
	}

	info := &pb.WorkerStatusResponse_WorkerInfo{
		WorkerId:            s.capabilities.WorkerId,
		Host:                s.capabilities.Hostname,
		NativeArch:          s.capabilities.NativeArch,
		CpuCores:            s.capabilities.CpuCores,
		MemoryBytes:         s.capabilities.MemoryBytes,
		ActiveTasks:         int32(atomic.LoadInt64(&s.activeTasks)),
		TotalTasksCompleted: total,
		AvgLatencyMs:        float32(avgTime),
	}

	return &pb.WorkerStatusResponse{
		Workers:        []*pb.WorkerStatusResponse_WorkerInfo{info},
		TotalWorkers:   1,
		HealthyWorkers: 1,
	}, nil
}

// GetWorkersForBuild not applicable for workers.
func (s *Server) GetWorkersForBuild(ctx context.Context, req *pb.WorkersForBuildRequest) (*pb.WorkersForBuildResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not applicable for workers")
}
