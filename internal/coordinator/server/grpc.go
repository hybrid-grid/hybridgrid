package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/resilience"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

// Config holds the coordinator gRPC server configuration.
type Config struct {
	Port           int
	AuthToken      string
	HeartbeatTTL   time.Duration
	RequestTimeout time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:           50051,
		HeartbeatTTL:   60 * time.Second,
		RequestTimeout: 120 * time.Second,
	}
}

// TaskEvent represents a task event for the dashboard.
type TaskEvent struct {
	ID           string
	BuildType    string
	Status       string
	WorkerID     string
	StartedAt    int64
	CompletedAt  int64
	DurationMs   int64
	ExitCode     int32
	FromCache    bool
	ErrorMessage string
}

// EventNotifier is called when task events occur.
type EventNotifier interface {
	NotifyTaskStarted(event *TaskEvent)
	NotifyTaskCompleted(event *TaskEvent)
}

// Server implements the coordinator gRPC server.
type Server struct {
	pb.UnimplementedBuildServiceServer

	config         Config
	server         *grpc.Server
	registry       registry.Registry
	scheduler      scheduler.Scheduler
	circuitManager *resilience.CircuitManager
	eventNotifier  EventNotifier

	activeTasks  int64
	queuedTasks  int64
	totalTasks   int64
	successTasks int64
	failedTasks  int64
	cacheHits    int64
	cacheMisses  int64
}

// New creates a new coordinator gRPC server.
func New(cfg Config) *Server {
	reg := registry.NewInMemoryRegistry(cfg.HeartbeatTTL)
	sched := scheduler.NewLeastLoadedScheduler(reg)
	circuitMgr := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())

	return &Server{
		config:         cfg,
		registry:       reg,
		scheduler:      sched,
		circuitManager: circuitMgr,
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterBuildServiceServer(s.server, s)

	log.Info().Int("port", s.config.Port).Msg("Coordinator gRPC server starting")
	return s.server.Serve(lis)
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if reg, ok := s.registry.(*registry.InMemoryRegistry); ok {
		reg.Stop()
	}
}

// Registry returns the worker registry.
func (s *Server) Registry() registry.Registry {
	return s.registry
}

// SetEventNotifier sets the event notifier for task events.
func (s *Server) SetEventNotifier(notifier EventNotifier) {
	s.eventNotifier = notifier
}

// Handshake handles worker registration.
func (s *Server) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	if req.Capabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "capabilities required")
	}

	// Validate auth token
	if s.config.AuthToken != "" && req.AuthToken != s.config.AuthToken {
		log.Warn().Str("hostname", req.Capabilities.Hostname).Msg("Worker rejected: invalid auth token")
		return &pb.HandshakeResponse{
			Accepted: false,
			Message:  "invalid auth token",
		}, nil
	}

	// Generate worker ID
	workerID := req.Capabilities.WorkerId
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%s-%d", req.Capabilities.Hostname, time.Now().UnixNano())
	}

	// Get worker address from context (or use provided)
	workerAddr := req.WorkerAddress
	if workerAddr == "" {
		workerAddr = fmt.Sprintf("%s:50052", req.Capabilities.Hostname)
	}

	// Get max parallel from capabilities (default to 4 if not set)
	maxParallel := req.Capabilities.MaxParallelTasks
	if maxParallel <= 0 {
		maxParallel = 4
	}

	// Register worker
	worker := &registry.WorkerInfo{
		ID:           workerID,
		Address:      workerAddr,
		Capabilities: req.Capabilities,
		MaxParallel:  maxParallel,
	}

	if err := s.registry.Add(worker); err != nil {
		// Worker might already exist, update heartbeat instead
		if err := s.registry.UpdateHeartbeat(workerID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to register worker: %v", err)
		}
	}

	// Log C++ capabilities for debugging
	var compilers []string
	if req.Capabilities.Cpp != nil {
		compilers = req.Capabilities.Cpp.Compilers
	}

	log.Info().
		Str("worker_id", workerID).
		Str("hostname", req.Capabilities.Hostname).
		Int32("cpu_cores", req.Capabilities.CpuCores).
		Int32("max_parallel", maxParallel).
		Str("arch", req.Capabilities.NativeArch.String()).
		Strs("cpp_compilers", compilers).
		Bool("docker", req.Capabilities.DockerAvailable).
		Msg("Worker registered")

	return &pb.HandshakeResponse{
		Accepted:                 true,
		Message:                  "worker registered successfully",
		AssignedWorkerId:         workerID,
		HeartbeatIntervalSeconds: int32(s.config.HeartbeatTTL.Seconds() / 2),
	}, nil
}

// Compile handles compilation requests by forwarding to workers.
func (s *Server) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	start := time.Now()

	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// Track cache miss (client checked cache first, this is a miss)
	atomic.AddInt64(&s.cacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	// Determine if we need OS filtering
	// Raw source mode (cross-compilation) doesn't need OS filtering
	// Preprocessed mode requires same-OS workers
	clientOSFilter := ""
	if len(req.RawSource) == 0 && len(req.PreprocessedSource) > 0 {
		// Preprocessed mode: filter by OS
		clientOSFilter = req.ClientOs
	}

	// Select worker
	worker, err := s.scheduler.Select(pb.BuildType_BUILD_TYPE_CPP, req.TargetArch, clientOSFilter)
	if err != nil {
		log.Error().Err(err).
			Str("task_id", req.TaskId).
			Str("client_os", req.ClientOs).
			Bool("cross_compile", len(req.RawSource) > 0).
			Msg("No worker available")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("no worker available: %v", err),
		}, nil
	}

	// Track task
	s.registry.IncrementTasks(worker.ID)
	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)

	defer func() {
		atomic.AddInt64(&s.activeTasks, -1)
	}()

	queueTime := time.Since(start)
	taskStartTime := time.Now()

	log.Debug().
		Str("task_id", req.TaskId).
		Str("worker_id", worker.ID).
		Dur("queue_time", queueTime).
		Msg("Forwarding compile request")

	// Notify task started
	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskStarted(&TaskEvent{
			ID:        req.TaskId,
			BuildType: "cpp",
			Status:    "running",
			WorkerID:  worker.ID,
			StartedAt: taskStartTime.Unix(),
		})
	}

	// Forward to worker
	resp, err := s.forwardCompile(ctx, worker, req)

	// Track completion
	success := err == nil && resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED
	compileTime := time.Duration(0)
	if resp != nil {
		compileTime = time.Duration(resp.CompilationTimeMs) * time.Millisecond
	}
	s.registry.DecrementTasks(worker.ID, success, compileTime)

	taskCompletedTime := time.Now()

	if success {
		atomic.AddInt64(&s.successTasks, 1)
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
	}

	// Notify task completed
	if s.eventNotifier != nil {
		event := &TaskEvent{
			ID:          req.TaskId,
			BuildType:   "cpp",
			WorkerID:    worker.ID,
			StartedAt:   taskStartTime.Unix(),
			CompletedAt: taskCompletedTime.Unix(),
			DurationMs:  taskCompletedTime.Sub(taskStartTime).Milliseconds(),
		}
		if success {
			event.Status = "completed"
			if resp != nil {
				event.ExitCode = resp.ExitCode
			}
		} else {
			event.Status = "failed"
			event.ExitCode = 1
			if err != nil {
				event.ErrorMessage = err.Error()
			} else if resp != nil {
				event.ExitCode = resp.ExitCode
				event.ErrorMessage = resp.Stderr
			}
		}
		s.eventNotifier.NotifyTaskCompleted(event)
	}

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Msg("Worker compilation failed")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	// Set queue time
	resp.QueueTimeMs = int64(queueTime.Milliseconds())

	return resp, nil
}

// forwardCompile forwards the compile request to a worker.
func (s *Server) forwardCompile(ctx context.Context, worker *registry.WorkerInfo, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	// Create connection to worker
	conn, err := grpc.DialContext(ctx, worker.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker: %w", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	// Forward request with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, s.config.RequestTimeout)
	defer cancel()

	return client.Compile(timeoutCtx, req)
}

// Build handles build requests.
func (s *Server) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// TODO: Implement for non-C++ build types
	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "build type not implemented yet",
	}, nil
}

// StreamBuild handles streaming build requests.
func (s *Server) StreamBuild(stream pb.BuildService_StreamBuildServer) error {
	var metadata *pb.BuildMetadata
	var sourceData []byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		switch payload := chunk.Payload.(type) {
		case *pb.BuildChunk_Metadata:
			metadata = payload.Metadata
		case *pb.BuildChunk_SourceChunk:
			sourceData = append(sourceData, payload.SourceChunk...)
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata required")
	}

	_ = sourceData // TODO: Use source data

	return stream.SendAndClose(&pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "stream build not implemented yet",
	})
}

// HealthCheck returns coordinator health status.
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	workers := s.registry.List()
	healthyCount := 0
	for _, w := range workers {
		if w.IsHealthy(s.config.HeartbeatTTL) {
			healthyCount++
		}
	}

	return &pb.HealthResponse{
		Healthy:     healthyCount > 0 || len(workers) == 0,
		ActiveTasks: int32(atomic.LoadInt64(&s.activeTasks)),
		QueuedTasks: int32(atomic.LoadInt64(&s.queuedTasks)),
	}, nil
}

// GetWorkerStatus returns status of all workers.
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	workers := s.registry.List()

	infos := make([]*pb.WorkerStatusResponse_WorkerInfo, 0, len(workers))
	healthyCount := 0

	for _, w := range workers {
		healthy := w.IsHealthy(s.config.HeartbeatTTL)
		if healthy {
			healthyCount++
		}

		info := &pb.WorkerStatusResponse_WorkerInfo{
			WorkerId:            w.ID,
			Host:                w.Capabilities.Hostname,
			NativeArch:          w.Capabilities.NativeArch,
			CpuCores:            w.Capabilities.CpuCores,
			MemoryBytes:         w.Capabilities.MemoryBytes,
			ActiveTasks:         w.ActiveTasks,
			TotalTasksCompleted: w.TotalTasks,
			LastHeartbeatUnix:   w.LastHeartbeat.Unix(),
		}
		infos = append(infos, info)
	}

	return &pb.WorkerStatusResponse{
		Workers:        infos,
		TotalWorkers:   int32(len(workers)),
		HealthyWorkers: int32(healthyCount),
	}, nil
}

// GetWorkersForBuild returns workers capable of handling a build type.
func (s *Server) GetWorkersForBuild(ctx context.Context, req *pb.WorkersForBuildRequest) (*pb.WorkersForBuildResponse, error) {
	workers := s.registry.ListByCapability(req.BuildType, pb.Architecture_ARCH_UNSPECIFIED)

	ids := make([]string, 0, len(workers))
	for _, w := range workers {
		ids = append(ids, w.ID)
	}

	return &pb.WorkersForBuildResponse{
		WorkerIds:      ids,
		AvailableCount: int32(len(ids)),
	}, nil
}

// ReportCacheHit handles cache hit reports from clients.
func (s *Server) ReportCacheHit(ctx context.Context, req *pb.ReportCacheHitRequest) (*pb.ReportCacheHitResponse, error) {
	if req.Hits > 0 {
		atomic.AddInt64(&s.cacheHits, int64(req.Hits))
	}
	return &pb.ReportCacheHitResponse{Acknowledged: true}, nil
}
