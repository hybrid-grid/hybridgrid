package server

import (
	"context"
	"fmt"
	"io"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/resilience"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/interceptors"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
	hgtls "github.com/h3nr1-d14z/hybridgrid/internal/security/tls"
)

const maxGRPCMessageSize = 512 * 1024 * 1024

// connPool caches gRPC client connections to workers by address.
type connPool struct {
	mu       sync.Mutex
	conns    map[string]*grpc.ClientConn
	dialOpts []grpc.DialOption
}

func newConnPool(dialOpts []grpc.DialOption) *connPool {
	return &connPool{
		conns:    make(map[string]*grpc.ClientConn),
		dialOpts: dialOpts,
	}
}

// get returns an existing connection or creates a new one.
func (p *connPool) get(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if conn, ok := p.conns[addr]; ok {
		p.mu.Unlock()
		return conn, nil
	}
	p.mu.Unlock()

	opts := make([]grpc.DialOption, len(p.dialOpts))
	copy(opts, p.dialOpts)
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if existing, ok := p.conns[addr]; ok {
		_ = conn.Close()
		return existing, nil
	}
	p.conns[addr] = conn
	return conn, nil
}

// closeAll closes all pooled connections.
func (p *connPool) closeAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for addr, conn := range p.conns {
		conn.Close()
		delete(p.conns, addr)
	}
}

// Config holds the coordinator gRPC server configuration.
type Config struct {
	Port            int
	AuthToken       string
	HeartbeatTTL    time.Duration
	RequestTimeout  time.Duration
	TLS             hgtls.Config
	Tracing         tracing.Config
	EnableRequestID bool
	// SchedulerType selects the scheduler implementation.
	// Valid: "leastloaded" (default), "simple", "p2c", "epsilon-greedy".
	SchedulerType string
	// EpsilonValue is the exploration rate for epsilon-greedy. Ignored
	// for other schedulers. Default 0.1 (Sutton & Barto §2.3 baseline).
	EpsilonValue float64
	// AlphaValue is the LinUCB exploration coefficient α (Li 2010 Eq. 4).
	// Theoretical form is 1 + sqrt(ln(2/δ)/2); we default to 1.0 and
	// expect empirical tuning. Ignored for non-LinUCB schedulers.
	AlphaValue float64
	// TaskLogPath is the path to the JSON Lines per-task log file.
	// Empty or "stdout" routes records to standard output.
	TaskLogPath string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:            50051,
		HeartbeatTTL:    60 * time.Second,
		RequestTimeout:  120 * time.Second,
		EnableRequestID: true,
		SchedulerType:   "leastloaded",
	}
}

// newScheduler constructs a scheduler.Scheduler from the configured type.
// Unknown types fall back to LeastLoaded for backward compatibility.
func newScheduler(cfg Config, reg registry.Registry, cm *resilience.CircuitManager) scheduler.Scheduler {
	switch cfg.SchedulerType {
	case "simple":
		return scheduler.NewSimpleScheduler(reg)
	case "p2c":
		return scheduler.NewP2CScheduler(scheduler.P2CConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "epsilon-greedy":
		eps := cfg.EpsilonValue
		if eps == 0 {
			eps = 0.1 // Sutton & Barto §2.3 default
		}
		return scheduler.NewEpsilonGreedyScheduler(scheduler.EpsilonGreedyConfig{
			Registry:       reg,
			CircuitChecker: cm,
			Epsilon:        eps,
		})
	case "linucb":
		return scheduler.NewLinUCBScheduler(scheduler.LinUCBConfig{
			Registry:       reg,
			CircuitChecker: cm,
			Alpha:          cfg.AlphaValue,
		})
	case "heft":
		return scheduler.NewHEFTScheduler(scheduler.HEFTConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "leastloaded", "":
		return scheduler.NewLeastLoadedScheduler(reg)
	default:
		log.Warn().Str("requested", cfg.SchedulerType).Msg("Unknown scheduler type; falling back to leastloaded")
		return scheduler.NewLeastLoadedScheduler(reg)
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
	workerConns    *connPool
	taskLogger     *TaskLogger

	activeTasks         int64
	queuedTasks         int64
	totalTasks          int64
	successTasks        int64
	failedTasks         int64
	cacheHits           int64
	cacheMisses         int64
	flutterBuilds       int64
	flutterCacheHits    int64
	flutterCacheMisses  int64
	flutterCacheMu      sync.RWMutex
	flutterCache        map[string]*flutterCacheEntry
	unityBuilds         int64
	unityCacheHits      int64
	unityCacheMisses    int64
	unityCacheMu        sync.RWMutex
	unityCache          map[string]*unityCacheEntry
	activeTasksByWorker sync.Map
}

type flutterCacheEntry struct {
	artifacts    []byte
	artifactList []*pb.ArtifactInfo
	stdout       string
	stderr       string
	buildTimeMs  int64
}

type unityCacheEntry struct {
	artifacts    []byte
	artifactList []*pb.ArtifactInfo
	stdout       string
	stderr       string
	buildTimeMs  int64
}

// New creates a new coordinator gRPC server.
func New(cfg Config) *Server {
	reg := registry.NewInMemoryRegistry(cfg.HeartbeatTTL)
	circuitMgr := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())
	sched := newScheduler(cfg, reg, circuitMgr)
	log.Info().Str("scheduler", cfg.SchedulerType).Msg("Scheduler initialized")

	taskLogger, err := NewTaskLogger(cfg.TaskLogPath)
	if err != nil {
		log.Warn().Err(err).Str("path", cfg.TaskLogPath).Msg("Failed to open task log; falling back to stdout")
		taskLogger, _ = NewTaskLogger("")
	}

	m := metrics.Default()
	circuitMgr.OnStateChange(func(workerID string, from, to resilience.CircuitState) {
		var stateValue metrics.CircuitStateValue
		switch to {
		case resilience.CircuitClosed:
			stateValue = metrics.CircuitStateClosed
		case resilience.CircuitHalfOpen:
			stateValue = metrics.CircuitStateHalfOpen
		case resilience.CircuitOpen:
			stateValue = metrics.CircuitStateOpen
		}
		m.SetCircuitState(workerID, stateValue)
	})

	// Build dial options for worker connections
	var dialOpts []grpc.DialOption
	if cfg.TLS.Enabled {
		creds, err := hgtls.ClientCredentials(cfg.TLS)
		if err == nil && creds != nil {
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds))
		} else {
			dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
	} else {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	dialOpts = append(dialOpts, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(maxGRPCMessageSize),
		grpc.MaxCallSendMsgSize(maxGRPCMessageSize),
	))
	if cfg.Tracing.Enable {
		dialOpts = append(dialOpts, tracing.DialOptions()...)
	}

	return &Server{
		config:         cfg,
		registry:       reg,
		scheduler:      sched,
		circuitManager: circuitMgr,
		workerConns:    newConnPool(dialOpts),
		taskLogger:     taskLogger,
		flutterCache:   make(map[string]*flutterCacheEntry),
		unityCache:     make(map[string]*unityCacheEntry),
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	var opts []grpc.ServerOption

	// Add TLS credentials if configured
	if s.config.TLS.Enabled {
		creds, err := hgtls.ServerCredentials(s.config.TLS)
		if err != nil {
			return fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		if creds != nil {
			opts = append(opts, grpc.Creds(creds))
			log.Info().
				Bool("mtls", s.config.TLS.RequireClientCert).
				Str("min_version", s.config.TLS.MinVersionName()).
				Msg("TLS enabled for coordinator gRPC server")
		}
	}

	// Add tracing interceptors if enabled
	if s.config.Tracing.Enable {
		opts = append(opts, tracing.ServerOptions()...)
		log.Info().Msg("OpenTelemetry tracing enabled for coordinator gRPC server")
	}

	// Add request ID interceptors if enabled
	if s.config.EnableRequestID {
		opts = append(opts,
			grpc.UnaryInterceptor(interceptors.UnaryRequestIDInterceptor()),
			grpc.StreamInterceptor(interceptors.StreamRequestIDInterceptor()),
		)
		log.Info().Msg("Request ID interceptor enabled for coordinator gRPC server")
	}

	s.server = grpc.NewServer(opts...)
	pb.RegisterBuildServiceServer(s.server, s)

	log.Info().Int("port", s.config.Port).Msg("Coordinator gRPC server starting")
	return s.server.Serve(lis)
}

// Stop gracefully stops the server and releases resources.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
	if s.workerConns != nil {
		s.workerConns.closeAll()
	}
	if reg, ok := s.registry.(*registry.InMemoryRegistry); ok {
		reg.Stop()
	}
	if s.taskLogger != nil {
		_ = s.taskLogger.Close()
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

	// Start tracing span for the coordinator compile flow
	ctx, span := tracing.StartSpan(ctx, "coordinator.Compile",
		tracing.WithCompileAttributes(req.TaskId, req.Compiler, req.TargetArch.String(), len(req.PreprocessedSource)+len(req.RawSource)),
	)
	defer span.End()

	if req.TaskId == "" {
		span.SetStatus(otelcodes.Error, "task_id required")
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// Track cache miss (client checked cache first, this is a miss)
	atomic.AddInt64(&s.cacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	// Determine OS filtering strategy:
	// - Raw source mode: no OS filter needed, workers with Docker can cross-compile
	//   using dockcross images. Workers with matching OS use native compiler.
	// - Preprocessed mode: must match OS since headers are already expanded.
	clientOSFilter := ""
	if len(req.RawSource) == 0 && len(req.PreprocessedSource) > 0 {
		clientOSFilter = req.ClientOs
	}

	// Select worker with tracing
	tracing.AddEvent(ctx, "scheduler.select.start")
	taskCtx := scheduler.TaskContext{
		SourceSizeBytes: len(req.PreprocessedSource) + len(req.RawSource),
	}
	worker, dispatchInfo, err := scheduler.SelectWith(s.scheduler, pb.BuildType_BUILD_TYPE_CPP, req.TargetArch, clientOSFilter, taskCtx)
	if err != nil {
		span.SetStatus(otelcodes.Error, "no worker available")
		tracing.RecordError(ctx, err)
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
	tracing.AddEvent(ctx, "scheduler.select.done")
	span.SetAttributes(tracing.AttrWorkerID.String(worker.ID))

	// Snapshot dispatch-time worker state for offline analysis. Captured
	// before IncrementTasks so the value reflects load at the scheduling
	// decision, not after this task has been booked. The read races with
	// concurrent dispatches but the log is for offline analysis only.
	activeAtDispatch := worker.ActiveTasks

	// Track task
	s.registry.IncrementTasks(worker.ID)
	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)

	m := metrics.Default()
	val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
	count := atomic.AddInt64(val.(*int64), 1)
	m.SetActiveTaskCount(worker.ID, float64(count))

	defer func() {
		atomic.AddInt64(&s.activeTasks, -1)
		if val, ok := s.activeTasksByWorker.Load(worker.ID); ok {
			count := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(worker.ID, float64(count))
		}
	}()

	queueTime := time.Since(start)
	taskStartTime := time.Now()
	span.SetAttributes(tracing.AttrQueueTimeMs.Int64(queueTime.Milliseconds()))

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

	uploadBytes := len(req.PreprocessedSource)
	if len(req.RawSource) > 0 {
		uploadBytes = len(req.RawSource)
	}
	m.RecordTransfer("upload", float64(uploadBytes))

	tracing.AddEvent(ctx, "forward.start")
	workerCallStart := time.Now()
	resp, err := s.forwardCompile(ctx, worker, req)
	workerLatency := time.Since(workerCallStart)
	m.RecordWorkerLatency(worker.ID, float64(workerLatency.Milliseconds()))
	tracing.AddEvent(ctx, "forward.done")

	if resp != nil && len(resp.ObjectFile) > 0 {
		m.RecordTransfer("download", float64(len(resp.ObjectFile)))
	}

	// Track completion
	success := err == nil && resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED
	compileTime := time.Duration(0)
	if resp != nil {
		compileTime = time.Duration(resp.CompilationTimeMs) * time.Millisecond
	}
	s.registry.DecrementTasks(worker.ID, success, compileTime)

	// Feedback loop for online-learning schedulers. Reward convention:
	// higher is better, so use negative log-latency (Decima §4.2 precedent).
	// Failed tasks receive a punishing constant so the learner downweights
	// the offending worker without conflating compile-time noise.
	if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok {
		var reward float64
		if success && resp != nil {
			reward = -math.Log1p(float64(resp.CompilationTimeMs))
		} else {
			reward = -math.Log1p(float64(s.config.RequestTimeout.Milliseconds()))
		}
		learner.RecordOutcome(worker.ID, reward, success, taskCtx)
	}

	taskCompletedTime := time.Now()
	totalDuration := taskCompletedTime.Sub(start)
	span.SetAttributes(tracing.AttrDurationMs.Int64(totalDuration.Milliseconds()))

	// Per-task structured log for offline analysis / RL training.
	if s.taskLogger != nil {
		var (
			workerCPUCores    int32
			workerMemBytes    int64
			workerNativeArch  string
		)
		if worker.Capabilities != nil {
			workerCPUCores = worker.Capabilities.CpuCores
			workerMemBytes = worker.Capabilities.MemoryBytes
			workerNativeArch = worker.Capabilities.NativeArch.String()
		}
		var (
			compileTimeMs int64
			exitCode      int32
		)
		if resp != nil {
			compileTimeMs = resp.CompilationTimeMs
			exitCode = resp.ExitCode
		}
		s.taskLogger.Log(&TaskLogRecord{
			TS:                          time.Now().UTC(),
			Event:                       "task_completed",
			TaskID:                      req.TaskId,
			BuildType:                   "cpp",
			Scheduler:                   s.config.SchedulerType,
			WorkerID:                    worker.ID,
			WorkerArch:                  workerNativeArch,
			WorkerNativeArch:            workerNativeArch,
			WorkerCPUCores:              workerCPUCores,
			WorkerMemBytes:              workerMemBytes,
			WorkerActiveTasksAtDispatch: activeAtDispatch,
			WorkerMaxParallel:           worker.MaxParallel,
			WorkerDiscoverySource:       worker.DiscoverySource,
			TargetArch:                  req.TargetArch.String(),
			ClientOS:                    req.ClientOs,
			SourceSizeBytes:             len(req.PreprocessedSource) + len(req.RawSource),
			PreprocessedSizeBytes:       len(req.PreprocessedSource),
			RawSourceSizeBytes:          len(req.RawSource),
			QueueTimeMs:                 queueTime.Milliseconds(),
			CompileTimeMs:               compileTimeMs,
			WorkerRPCLatencyMs:          workerLatency.Milliseconds(),
			TotalDurationMs:             totalDuration.Milliseconds(),
			Success:                     success,
			ExitCode:                    exitCode,
			FromCache:                   false,
			QValueAtDispatch:            dispatchInfo.QValueAtDispatch,
			WasExploration:              dispatchInfo.WasExploration,
		})
	}

	if success {
		atomic.AddInt64(&s.successTasks, 1)
		span.SetStatus(otelcodes.Ok, "compilation succeeded")
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
		span.SetStatus(otelcodes.Error, "compilation failed")
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
		tracing.RecordError(ctx, err)
		log.Error().Err(err).Str("task_id", req.TaskId).Msg("Worker compilation failed")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	// Set queue time
	resp.QueueTimeMs = int64(queueTime.Milliseconds())

	duration := totalDuration.Seconds()
	buildType := "cpp"
	if success {
		m.RecordTaskComplete(metrics.TaskStatusSuccess, buildType, worker.ID, duration)
	} else {
		m.RecordTaskComplete(metrics.TaskStatusError, buildType, worker.ID, duration)
	}
	m.RecordQueueTime(buildType, queueTime.Seconds())

	return resp, nil
}

// forwardCompile forwards the compile request to a worker.
func (s *Server) forwardCompile(ctx context.Context, worker *registry.WorkerInfo, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	// Get pooled connection to worker (reuses existing connections)
	conn, err := s.workerConns.get(ctx, worker.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker: %w", err)
	}

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

	if req.GetFlutterConfig() != nil {
		return s.handleFlutterBuild(ctx, req)
	}

	if req.GetUnityConfig() != nil {
		return s.handleUnityBuild(ctx, req)
	}

	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "build type not implemented yet",
	}, nil
}

func (s *Server) handleFlutterBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	start := time.Now()
	m := metrics.Default()

	flutterConfig := req.GetFlutterConfig()
	flutterVersion := flutterConfig.GetFlutterVersion()
	cacheKey := cache.FlutterCacheKey(flutterConfig, req.SourceHash, flutterVersion)

	if cached := s.getFlutterCache(cacheKey); cached != nil {
		atomic.AddInt64(&s.cacheHits, 1)
		atomic.AddInt64(&s.flutterCacheHits, 1)
		atomic.AddInt64(&s.flutterBuilds, 1)
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.successTasks, 1)

		if s.eventNotifier != nil {
			taskStart := start.Unix()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:        req.TaskId,
				BuildType: "flutter",
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "flutter",
				Status:       "completed",
				WorkerID:     "",
				StartedAt:    taskStart,
				CompletedAt:  time.Now().Unix(),
				DurationMs:   cached.buildTimeMs,
				ExitCode:     0,
				FromCache:    true,
				ErrorMessage: "",
			})
		}

		return &pb.BuildResponse{
			Status:       pb.TaskStatus_STATUS_COMPLETED,
			ExitCode:     0,
			Stdout:       cached.stdout,
			Stderr:       cached.stderr,
			Artifacts:    append([]byte(nil), cached.artifacts...),
			ArtifactList: cloneArtifactList(cached.artifactList),
			BuildTimeMs:  cached.buildTimeMs,
			FromCache:    true,
		}, nil
	}

	atomic.AddInt64(&s.cacheMisses, 1)
	atomic.AddInt64(&s.flutterCacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	worker, err := s.selectFlutterWorker(req.TargetPlatform)
	if err != nil {
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.failedTasks, 1)
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("target_platform", req.TargetPlatform.String()).
			Msg("No worker available for flutter build")

		if s.eventNotifier != nil {
			taskStart := start.Unix()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:        req.TaskId,
				BuildType: "flutter",
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "flutter",
				Status:       "failed",
				WorkerID:     "",
				StartedAt:    taskStart,
				CompletedAt:  time.Now().Unix(),
				DurationMs:   0,
				ExitCode:     1,
				FromCache:    false,
				ErrorMessage: fmt.Sprintf("no worker available: %v", err),
			})
		}

		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("no worker available: %v", err),
		}, nil
	}

	conn, err := s.workerConns.get(ctx, worker.Address)
	if err != nil {
		log.Error().Err(err).Str("worker", worker.ID).Msg("Failed to connect to worker")
		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("failed to connect to worker: %v", err),
		}, nil
	}

	s.registry.IncrementTasks(worker.ID)
	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)
	atomic.AddInt64(&s.flutterBuilds, 1)

	val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
	count := atomic.AddInt64(val.(*int64), 1)
	m.SetActiveTaskCount(worker.ID, float64(count))

	defer func() {
		atomic.AddInt64(&s.activeTasks, -1)
		if val, ok := s.activeTasksByWorker.Load(worker.ID); ok {
			c := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(worker.ID, float64(c))
		}
	}()

	taskStartTime := time.Now()

	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskStarted(&TaskEvent{
			ID:        req.TaskId,
			BuildType: "flutter",
			Status:    "running",
			WorkerID:  worker.ID,
			StartedAt: taskStartTime.Unix(),
		})
	}

	client := pb.NewBuildServiceClient(conn)
	buildResp, err := client.Build(ctx, &pb.BuildRequest{
		TaskId:         req.TaskId,
		SourceHash:     req.SourceHash,
		SourceArchive:  req.SourceArchive,
		BuildType:      req.BuildType,
		TargetPlatform: req.TargetPlatform,
		Config:         req.Config,
		TimeoutSeconds: req.TimeoutSeconds,
	})

	success := err == nil && buildResp != nil && buildResp.Status == pb.TaskStatus_STATUS_COMPLETED

	s.registry.DecrementTasks(worker.ID, success, time.Duration(0))

	taskCompletedTime := time.Now()

	if success {
		atomic.AddInt64(&s.successTasks, 1)
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
	}

	if s.eventNotifier != nil {
		event := &TaskEvent{
			ID:          req.TaskId,
			BuildType:   "flutter",
			WorkerID:    worker.ID,
			StartedAt:   taskStartTime.Unix(),
			CompletedAt: taskCompletedTime.Unix(),
			DurationMs:  taskCompletedTime.Sub(taskStartTime).Milliseconds(),
			FromCache:   false,
		}
		if success {
			event.Status = "completed"
			if buildResp != nil {
				event.ExitCode = buildResp.ExitCode
			}
		} else {
			event.Status = "failed"
			event.ExitCode = 1
			if err != nil {
				event.ErrorMessage = err.Error()
			} else if buildResp != nil {
				event.ExitCode = buildResp.ExitCode
				event.ErrorMessage = buildResp.Stderr
			}
		}
		s.eventNotifier.NotifyTaskCompleted(event)
	}

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).
			Msg("Worker build failed")
		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	if buildResp != nil && len(buildResp.Artifacts) > 0 {
		s.setFlutterCache(cacheKey, buildResp)
	}

	return buildResp, nil
}

func (s *Server) handleUnityBuild(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	start := time.Now()
	m := metrics.Default()

	unityConfig := req.GetUnityConfig()
	unityVersion := unityConfig.GetUnityVersion()
	cacheKey := cache.UnityCacheKey(unityConfig, req.SourceHash, unityVersion, req.TargetPlatform)

	if cached := s.getUnityCache(cacheKey); cached != nil {
		atomic.AddInt64(&s.cacheHits, 1)
		atomic.AddInt64(&s.unityCacheHits, 1)
		atomic.AddInt64(&s.unityBuilds, 1)
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.successTasks, 1)

		if s.eventNotifier != nil {
			taskStart := start.Unix()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:        req.TaskId,
				BuildType: "unity",
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "unity",
				Status:       "completed",
				WorkerID:     "",
				StartedAt:    taskStart,
				CompletedAt:  time.Now().Unix(),
				DurationMs:   cached.buildTimeMs,
				ExitCode:     0,
				FromCache:    true,
				ErrorMessage: "",
			})
		}

		return &pb.BuildResponse{
			Status:       pb.TaskStatus_STATUS_COMPLETED,
			ExitCode:     0,
			Stdout:       cached.stdout,
			Stderr:       cached.stderr,
			Artifacts:    append([]byte(nil), cached.artifacts...),
			ArtifactList: cloneArtifactList(cached.artifactList),
			BuildTimeMs:  cached.buildTimeMs,
			FromCache:    true,
		}, nil
	}

	atomic.AddInt64(&s.cacheMisses, 1)
	atomic.AddInt64(&s.unityCacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	worker, err := s.selectUnityWorker(req.TargetPlatform)
	if err != nil {
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.failedTasks, 1)
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("target_platform", req.TargetPlatform.String()).
			Msg("No worker available for unity build")

		if s.eventNotifier != nil {
			taskStart := start.Unix()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:        req.TaskId,
				BuildType: "unity",
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "unity",
				Status:       "failed",
				WorkerID:     "",
				StartedAt:    taskStart,
				CompletedAt:  time.Now().Unix(),
				DurationMs:   0,
				ExitCode:     1,
				FromCache:    false,
				ErrorMessage: fmt.Sprintf("no worker available: %v", err),
			})
		}

		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("no worker available: %v", err),
		}, nil
	}

	conn, err := s.workerConns.get(ctx, worker.Address)
	if err != nil {
		log.Error().Err(err).Str("worker", worker.ID).Msg("Failed to connect to worker")
		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("failed to connect to worker: %v", err),
		}, nil
	}

	s.registry.IncrementTasks(worker.ID)
	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)
	atomic.AddInt64(&s.unityBuilds, 1)

	val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
	count := atomic.AddInt64(val.(*int64), 1)
	m.SetActiveTaskCount(worker.ID, float64(count))

	defer func() {
		atomic.AddInt64(&s.activeTasks, -1)
		if val, ok := s.activeTasksByWorker.Load(worker.ID); ok {
			c := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(worker.ID, float64(c))
		}
	}()

	taskStartTime := time.Now()

	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskStarted(&TaskEvent{
			ID:        req.TaskId,
			BuildType: "unity",
			Status:    "running",
			WorkerID:  worker.ID,
			StartedAt: taskStartTime.Unix(),
		})
	}

	client := pb.NewBuildServiceClient(conn)
	buildResp, err := client.Build(ctx, &pb.BuildRequest{
		TaskId:         req.TaskId,
		SourceHash:     req.SourceHash,
		SourceArchive:  req.SourceArchive,
		BuildType:      req.BuildType,
		TargetPlatform: req.TargetPlatform,
		Config:         req.Config,
		TimeoutSeconds: req.TimeoutSeconds,
	})

	success := err == nil && buildResp != nil && buildResp.Status == pb.TaskStatus_STATUS_COMPLETED

	s.registry.DecrementTasks(worker.ID, success, time.Duration(0))

	taskCompletedTime := time.Now()

	if success {
		atomic.AddInt64(&s.successTasks, 1)
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
	}

	if s.eventNotifier != nil {
		event := &TaskEvent{
			ID:          req.TaskId,
			BuildType:   "unity",
			WorkerID:    worker.ID,
			StartedAt:   taskStartTime.Unix(),
			CompletedAt: taskCompletedTime.Unix(),
			DurationMs:  taskCompletedTime.Sub(taskStartTime).Milliseconds(),
			FromCache:   false,
		}
		if success {
			event.Status = "completed"
			if buildResp != nil {
				event.ExitCode = buildResp.ExitCode
			}
		} else {
			event.Status = "failed"
			event.ExitCode = 1
			if err != nil {
				event.ErrorMessage = err.Error()
			} else if buildResp != nil {
				event.ExitCode = buildResp.ExitCode
				event.ErrorMessage = buildResp.Stderr
			}
		}
		s.eventNotifier.NotifyTaskCompleted(event)
	}

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).
			Msg("Worker build failed")
		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	if buildResp != nil && len(buildResp.Artifacts) > 0 {
		s.setUnityCache(cacheKey, buildResp)
	}

	return buildResp, nil
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
		ActiveTasks: clampInt64ToInt32(atomic.LoadInt64(&s.activeTasks)),
		QueuedTasks: clampInt64ToInt32(atomic.LoadInt64(&s.queuedTasks)),
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

		caps := workerCapabilitiesOrDefault(w)

		info := &pb.WorkerStatusResponse_WorkerInfo{
			WorkerId:            w.ID,
			Host:                caps.Hostname,
			NativeArch:          caps.NativeArch,
			CpuCores:            caps.CpuCores,
			MemoryBytes:         caps.MemoryBytes,
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

func clampInt64ToInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func workerCapabilitiesOrDefault(worker *registry.WorkerInfo) *pb.WorkerCapabilities {
	if worker != nil && worker.Capabilities != nil {
		return worker.Capabilities
	}

	return &pb.WorkerCapabilities{}
}

func (s *Server) getFlutterCache(key string) *flutterCacheEntry {
	s.flutterCacheMu.RLock()
	entry := s.flutterCache[key]
	s.flutterCacheMu.RUnlock()
	return entry
}

func (s *Server) setFlutterCache(key string, resp *pb.BuildResponse) {
	entry := &flutterCacheEntry{
		stdout:       resp.Stdout,
		stderr:       resp.Stderr,
		artifacts:    resp.Artifacts,
		artifactList: cloneArtifactList(resp.ArtifactList),
		buildTimeMs:  resp.BuildTimeMs,
	}
	s.flutterCacheMu.Lock()
	s.flutterCache[key] = entry
	s.flutterCacheMu.Unlock()
}

func (s *Server) selectFlutterWorker(targetPlatform pb.TargetPlatform) (*registry.WorkerInfo, error) {
	workers := s.registry.List()
	for _, w := range workers {
		if workerSupportsFlutterPlatform(w, targetPlatform) {
			if w.IsHealthy(s.config.HeartbeatTTL) {
				return w, nil
			}
		}
	}
	return nil, fmt.Errorf("no flutter worker available for platform %s", targetPlatform)
}

func workerSupportsFlutterPlatform(worker *registry.WorkerInfo, platform pb.TargetPlatform) bool {
	if worker == nil || worker.Capabilities == nil || worker.Capabilities.Flutter == nil {
		return false
	}
	if len(worker.Capabilities.Flutter.Platforms) == 0 {
		return true
	}
	for _, p := range worker.Capabilities.Flutter.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}

func (s *Server) getUnityCache(key string) *unityCacheEntry {
	s.unityCacheMu.RLock()
	entry := s.unityCache[key]
	s.unityCacheMu.RUnlock()
	return entry
}

func (s *Server) setUnityCache(key string, resp *pb.BuildResponse) {
	entry := &unityCacheEntry{
		stdout:       resp.Stdout,
		stderr:       resp.Stderr,
		artifacts:    resp.Artifacts,
		artifactList: cloneArtifactList(resp.ArtifactList),
		buildTimeMs:  resp.BuildTimeMs,
	}
	s.unityCacheMu.Lock()
	s.unityCache[key] = entry
	s.unityCacheMu.Unlock()
}

func (s *Server) selectUnityWorker(targetPlatform pb.TargetPlatform) (*registry.WorkerInfo, error) {
	workers := s.registry.List()
	for _, w := range workers {
		if workerSupportsUnityPlatform(w, targetPlatform) {
			if w.IsHealthy(s.config.HeartbeatTTL) {
				return w, nil
			}
		}
	}
	return nil, fmt.Errorf("no unity worker available for platform %s", targetPlatform)
}

func workerSupportsUnityPlatform(worker *registry.WorkerInfo, platform pb.TargetPlatform) bool {
	if worker == nil || worker.Capabilities == nil || worker.Capabilities.Unity == nil {
		return false
	}
	if len(worker.Capabilities.Unity.BuildTargets) == 0 {
		return true
	}
	for _, p := range worker.Capabilities.Unity.BuildTargets {
		if p == platform {
			return true
		}
	}
	return false
}

func cloneArtifactList(list []*pb.ArtifactInfo) []*pb.ArtifactInfo {
	if list == nil {
		return nil
	}
	result := make([]*pb.ArtifactInfo, len(list))
	for i, a := range list {
		result[i] = &pb.ArtifactInfo{
			Name:      a.Name,
			Path:      a.Path,
			SizeBytes: a.SizeBytes,
			Checksum:  a.Checksum,
		}
	}
	return result
}
