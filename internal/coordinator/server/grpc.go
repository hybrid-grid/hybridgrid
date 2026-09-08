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
	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

const maxGRPCMessageSize = 512 * 1024 * 1024

// dispatchQueueSize bounds how many pending selection requests may queue
// up behind the single dispatchLoop goroutine before Compile() callers
// block trying to submit one. Generously sized relative to any burst
// this system exercises (make -jN cold starts up to a few dozen) so a
// legitimate burst never blocks on channel capacity itself — only on
// dispatchLoop's own (sub-microsecond) processing rate.
const dispatchQueueSize = 256

// dispatchRequest is one pending "pick a worker and book it" decision.
// Compile() submits these to s.dispatchCh instead of calling
// scheduler.SelectWith and s.registry.IncrementTasks directly, so every
// selection decision across every concurrent RPC handler is funneled
// through the single dispatchLoop goroutine below.
//
// Why: SelectWith (reads a worker's ActiveTasks) and IncrementTasks
// (books the decision by writing it back) are two separate registry
// operations. When Compile() called them inline, concurrent handler
// goroutines could each read the same pre-booking ActiveTasks value
// before any of them had written their own increment — a classic
// check-then-act race. Under `make -jN` for N large enough relative to
// total worker capacity (empirically N ≳ 7 on a 5-worker/10-slot
// cluster — see undersubscription-explains-tie / cgroup-fix-verified-
// live), this let multiple in-flight decisions overbook the same
// (especially low-max_parallel) worker, which then rejected the excess
// with ResourceExhausted; a bounded reselection-with-backoff retry
// (maxDispatchAttempts) mitigated but did not eliminate the failures.
//
// Routing every decision through one goroutine removes the race
// entirely: only dispatchLoop ever reads or books a worker's
// ActiveTasks for a Compile() dispatch, so no two decisions can ever
// observe the same stale value. This does not add meaningful latency —
// a selection decision is a few in-memory comparisons (sub-microsecond)
// against ~10-500ms compile times — and does not serialize the actual
// compiles, which still run concurrently across all workers exactly as
// before; only the brief "who gets this file" moment is serialized.
//
// Scope: only the C/C++ Compile() path is routed through this queue.
// The Flutter/Unity build paths (selectFlutterWorker/selectUnityWorker)
// use the plain scheduler.Select (not the learning-aware SelectWith)
// and are outside this benchmark's scope per the paper's own stated
// scope; they are not touched here.
type dispatchRequest struct {
	buildType pb.BuildType
	arch      pb.Architecture
	clientOS  string
	ctx       scheduler.TaskContext
	result    chan dispatchResult
}

// dispatchResult is dispatchLoop's answer to a dispatchRequest.
// activeAtDispatch mirrors the semantics the inline code used to
// provide: the worker's ActiveTasks read immediately before
// IncrementTasks booked this decision, i.e. load at decision time, not
// after. It must travel back through this struct rather than have the
// caller re-read worker.ActiveTasks after the fact, because by the time
// dispatch() returns, IncrementTasks has already run — a fresh read
// would see the post-booking count, not the pre-booking one the offline
// analysis field (worker_active_tasks_at_dispatch) is documented to mean.
type dispatchResult struct {
	worker           *registry.WorkerInfo
	info             scheduler.DispatchInfo
	activeAtDispatch int32
	err              error
}

// dispatchLoop is the single goroutine that owns worker selection and
// booking for the Compile() path. It runs for the coordinator's
// lifetime, started once in New() and stopped by closing s.dispatchCh
// in Stop(). See dispatchRequest's doc comment for why this exists.
func (s *Server) dispatchLoop() {
	for req := range s.dispatchCh {
		worker, info, err := scheduler.SelectWith(s.scheduler, req.buildType, req.arch, req.clientOS, req.ctx)
		var activeAtDispatch int32
		if err == nil {
			// Read-then-book, both on this single goroutine: no other
			// goroutine can interleave a SelectWith between this read
			// and the IncrementTasks call below, because dispatchLoop
			// is the only caller of either for the Compile() path.
			activeAtDispatch = worker.ActiveTasks
			s.registry.IncrementTasks(worker.ID)
		}
		req.result <- dispatchResult{worker: worker, info: info, activeAtDispatch: activeAtDispatch, err: err}
	}
}

// dispatch submits a selection request to dispatchLoop and blocks for
// its result. Safe to call from any number of concurrent goroutines —
// that concurrency is exactly what dispatchLoop serializes away.
//
// Registered on s.dispatchWG for the whole send-then-wait-for-result
// span (not just the send): Stop() waits on this WaitGroup before
// closing s.dispatchCh, so a close can never race a concurrent send —
// see Stop()'s comment for why this can't just rely on the gRPC
// server's GracefulStop() to provide that guarantee.
func (s *Server) dispatch(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx scheduler.TaskContext) (*registry.WorkerInfo, scheduler.DispatchInfo, int32, error) {
	s.dispatchWG.Add(1)
	defer s.dispatchWG.Done()

	req := dispatchRequest{
		buildType: buildType,
		arch:      arch,
		clientOS:  clientOS,
		ctx:       ctx,
		result:    make(chan dispatchResult, 1),
	}
	s.dispatchCh <- &req
	res := <-req.result
	return res.worker, res.info, res.activeAtDispatch, res.err
}

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
	// Valid: "leastloaded" (default), "simple", "p2c", "epsilon-greedy",
	// "linucb", "hybrid-linucb", "heft".
	SchedulerType string
	// EpsilonValue is the exploration rate for epsilon-greedy. Ignored
	// for other schedulers. Default 0.1 (Sutton & Barto §2.3 baseline).
	EpsilonValue float64
	// AlphaValue is the LinUCB exploration coefficient α (Li 2010 Eq. 4).
	// Theoretical form is 1 + sqrt(ln(2/δ)/2); we default to 1.0 and
	// expect empirical tuning. Ignored for non-LinUCB schedulers.
	AlphaValue float64
	// WarmStartTasks is the hybrid-linucb warm-start window: that many
	// initial dispatches are routed least-loaded while the bandit learns
	// passively. 0 disables warm-start. Ignored by other schedulers.
	WarmStartTasks int
	// LoadPenaltyValue is the hybrid-linucb λ in p = mean + bonus −
	// λ·loadRatio. 0 disables the penalty. Ignored by other schedulers.
	LoadPenaltyValue float64
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
		// Hybrid-LinUCB defaults per docs/thesis/hybrid_linucb_proposal.md
		// §3; zero is a meaningful value ("off") for both, so defaults
		// live here rather than in the factory.
		WarmStartTasks:   100,
		LoadPenaltyValue: 0.5,
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
	case "hybrid-linucb":
		// Pass-through without zero-defaulting so --warm-start=0 or
		// --load-penalty=0 genuinely disables each mechanism for
		// ablation runs.
		return scheduler.NewLinUCBScheduler(scheduler.LinUCBConfig{
			Registry:       reg,
			CircuitChecker: cm,
			Alpha:          cfg.AlphaValue,
			WarmStartTasks: cfg.WarmStartTasks,
			LoadPenalty:    cfg.LoadPenaltyValue,
		})
	case "heft":
		return scheduler.NewHEFTScheduler(scheduler.HEFTConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "icecc-fastest":
		return scheduler.NewIceccFastestScheduler(scheduler.IceccConfig{
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
	BuildID      string
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
	dispatchCh     chan *dispatchRequest
	dispatchWG     sync.WaitGroup
	dispatchOnce   sync.Once

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

	s := &Server{
		config:         cfg,
		registry:       reg,
		scheduler:      sched,
		circuitManager: circuitMgr,
		workerConns:    newConnPool(dialOpts),
		taskLogger:     taskLogger,
		flutterCache:   make(map[string]*flutterCacheEntry),
		unityCache:     make(map[string]*unityCacheEntry),
		dispatchCh:     make(chan *dispatchRequest, dispatchQueueSize),
	}
	go s.dispatchLoop()
	return s
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
	// GracefulStop above waits for in-flight RPCs to finish when this
	// Server is fronted by its own s.server (the real cmd/hg-coord
	// path). It does NOT cover callers that reach dispatch() without
	// going through s.server — e.g. tests that register this Server on
	// a separately-constructed grpc.Server (bufconn-based test harnesses
	// never set s.server at all, since they never call s.Start()), or
	// any future direct dispatch() caller. dispatchWG closes that gap
	// unconditionally: every dispatch() call is registered on it for its
	// full send-plus-wait-for-result span, so waiting for it to drain
	// here guarantees no goroutine can still be sending to dispatchCh
	// when we close it, regardless of how this Server was wired up.
	s.dispatchWG.Wait()
	// sync.Once guards against a double Stop() call panicking on a
	// second close of an already-closed channel — cmd/hg-coord's
	// shutdown path only calls Stop() once today, but this is cheap
	// insurance against that changing (or a test calling it twice).
	if s.dispatchCh != nil {
		s.dispatchOnce.Do(func() { close(s.dispatchCh) })
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
	buildID := validation.NormalizeBuildID(req.BuildId)
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
		SourceSizeBytes:    len(req.PreprocessedSource) + len(req.RawSource),
		RawSourceSizeBytes: len(req.RawSource),
		SourceFilename:     req.SourceFilename,
		Compiler:           req.Compiler,
		TaskID:             req.TaskId,
	}
	m := metrics.Default()
	uploadBytes := len(req.PreprocessedSource)
	if len(req.RawSource) > 0 {
		uploadBytes = len(req.RawSource)
	}
	m.RecordTransfer("upload", float64(uploadBytes))

	// Select → book → forward, with bounded reselection when the worker's
	// admission control rejects the dispatch (ResourceExhausted).
	// Concurrent Compile handlers race between SelectWith (which reads
	// ActiveTasks) and IncrementTasks (which books it), so a burst — e.g.
	// make -jN cold start — can over-book a small worker. The worker
	// rejects rather than queueing, and without reselection that
	// scheduling artifact surfaces to the client as a compile failure.
	// A rejected attempt is unbooked with success=false: the rejection
	// is a genuine overload signal for the worker's stats. The learner
	// sees no RecordOutcome for aborted attempts — re-Select overwrites
	// the pendingX entry for this TaskID.
	//
	// maxDispatchAttempts was 3, tuned for modest bursts. Empirically
	// (undersubscription-explains-tie / cgroup-fix-verified-live
	// follow-up), a 5-worker cluster with 10 total max_parallel slots
	// hard-fails builds under `make -j7` and above with 3 attempts: the
	// linear backoff below (25ms, 50ms) doesn't span enough of the
	// dispatch storm at build-start cold start, when make launches all
	// -jN local hgcc processes near-simultaneously. Raised to 8, which
	// extends the cumulative backoff window to ~700ms (25+50+...+175ms
	// across the 7 retries between 8 attempts) — comfortably above the
	// ~10ms median compile time this cluster observes, so a slot should
	// free up well before attempts run out.
	// The race this originally compensated for (SelectWith/IncrementTasks
	// reading and writing on separate, unsynchronized goroutines) is now
	// eliminated at the source: selection is funneled through the single
	// dispatchLoop goroutine (see dispatch()/dispatchRequest below), so
	// the registry's ActiveTasks is always consistent at decision time —
	// no two concurrent Compile() calls can ever read the same stale
	// value. A ResourceExhausted here now means every eligible worker is
	// genuinely at capacity, not a bookkeeping race, so this retry loop's
	// remaining job is simply: wait a beat for a real slot to free up
	// (compiles are typically ~10-500ms) and try again. Left at 8 with
	// backoff for that legitimate case.
	const maxDispatchAttempts = 8
	var (
		worker           *registry.WorkerInfo
		dispatchInfo     scheduler.DispatchInfo
		resp             *pb.CompileResponse
		err              error
		workerLatency    time.Duration
		queueTime        time.Duration
		taskStartTime    time.Time
		activeAtDispatch int32
	)
	atomic.AddInt64(&s.activeTasks, 1)
	defer atomic.AddInt64(&s.activeTasks, -1)
	for attempt := 1; ; attempt++ {
		worker, dispatchInfo, activeAtDispatch, err = s.dispatch(pb.BuildType_BUILD_TYPE_CPP, req.TargetArch, clientOSFilter, taskCtx)
		if err != nil {
			// dispatch() itself found no eligible worker — every worker is
			// genuinely at MaxParallel right now (this is the schedulers'
			// own admission check, e.g. ErrNoMatchingWorkers; distinct from
			// the ResourceExhausted branch below, which fires only after a
			// worker WAS selected and its own admission control rejected
			// the forwarded compile). Previously this returned FAILED
			// immediately with no retry at all, which made properly
			// capacity-aware schedulers (P2C, LinUCB, epsilon-greedy, HEFT,
			// and — since the fix above — LeastLoaded) fail builds outright
			// under a burst that a brief wait would have resolved: compiles
			// finish in ~10-500ms, so a worker often frees a slot within a
			// couple of backoff cycles. Retried with the same budget and
			// backoff as the ResourceExhausted branch below, so a single
			// Compile() call never waits longer in total than before this
			// change — see undersubscription-explains-tie /
			// cgroup-fix-verified-live.
			if attempt < maxDispatchAttempts {
				log.Warn().
					Str("task_id", req.TaskId).
					Int("attempt", attempt).
					Msg("No worker currently has capacity; retrying")
				time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
				continue
			}
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

		// activeAtDispatch was captured inside dispatch(), before
		// IncrementTasks booked this decision — see dispatchResult's doc
		// comment for why it can't be re-read from worker.ActiveTasks here.
		val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
		count := atomic.AddInt64(val.(*int64), 1)
		m.SetActiveTaskCount(worker.ID, float64(count))

		queueTime = time.Since(start)
		taskStartTime = time.Now()

		log.Debug().
			Str("task_id", req.TaskId).
			Str("worker_id", worker.ID).
			Dur("queue_time", queueTime).
			Msg("Forwarding compile request")

		// Notify task started. A rejected attempt re-notifies with the
		// next worker; events are keyed by task ID so the dashboard just
		// sees the assignment move.
		if s.eventNotifier != nil {
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:        req.TaskId,
				BuildType: "cpp",
				BuildID:   buildID,
				Status:    "running",
				WorkerID:  worker.ID,
				StartedAt: taskStartTime.Unix(),
			})
		}

		tracing.AddEvent(ctx, "forward.start")
		workerCallStart := time.Now()
		resp, err = s.forwardCompile(ctx, worker, req)
		workerLatency = time.Since(workerCallStart)
		m.RecordWorkerLatency(worker.ID, float64(workerLatency.Milliseconds()))
		tracing.AddEvent(ctx, "forward.done")

		if err != nil && status.Code(err) == codes.ResourceExhausted && attempt < maxDispatchAttempts {
			s.registry.DecrementTasks(worker.ID, false, 0)
			if v, ok := s.activeTasksByWorker.Load(worker.ID); ok {
				c := atomic.AddInt64(v.(*int64), -1)
				m.SetActiveTaskCount(worker.ID, float64(c))
			}
			log.Warn().
				Str("task_id", req.TaskId).
				Str("worker_id", worker.ID).
				Int("attempt", attempt).
				Msg("Worker at capacity; reselecting")
			// Brief backoff lets the racing bookings land in the
			// registry before the next Select reads it.
			time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
			continue
		}
		break
	}
	span.SetAttributes(tracing.AttrQueueTimeMs.Int64(queueTime.Milliseconds()))

	atomic.AddInt64(&s.totalTasks, 1)

	defer func() {
		if val, ok := s.activeTasksByWorker.Load(worker.ID); ok {
			count := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(worker.ID, float64(count))
		}
	}()

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
	// higher is better. We use a normalised negative log-latency so the
	// reward magnitude does not dwarf LinUCB's UCB exploration bonus
	// during warm-up (code-review finding HIGH-2). Normalisation divisor
	// is log1p(RequestTimeoutMs) so the reward lies in roughly [-1, 0].
	// The log transform compresses the heavy tail (M1 P99/P50 ≈ 29×);
	// see docs/thesis/theory-notes.md §4.3 for the empirical-choice
	// rationale (this is NOT the Decima reward function).
	if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok {
		timeoutMs := float64(s.config.RequestTimeout.Milliseconds())
		if timeoutMs <= 0 {
			timeoutMs = 120000 // 2-minute fallback
		}
		denom := math.Log1p(timeoutMs)
		var reward float64
		if success && resp != nil {
			reward = -math.Log1p(float64(resp.CompilationTimeMs)) / denom
		} else {
			reward = -1.0
		}
		learner.RecordOutcome(worker.ID, reward, success, taskCtx)
	}

	taskCompletedTime := time.Now()
	totalDuration := taskCompletedTime.Sub(start)
	span.SetAttributes(tracing.AttrDurationMs.Int64(totalDuration.Milliseconds()))

	// Per-task structured log for offline analysis / RL training.
	if s.taskLogger != nil {
		var (
			workerCPUCores   int32
			workerCPUMillis  int32
			workerMemBytes   int64
			workerNativeArch string
		)
		if worker.Capabilities != nil {
			workerCPUCores = worker.Capabilities.CpuCores
			workerCPUMillis = worker.Capabilities.CpuMillis
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
			BuildID:                     buildID,
			Scheduler:                   s.config.SchedulerType,
			WorkerID:                    worker.ID,
			WorkerArch:                  workerNativeArch,
			WorkerNativeArch:            workerNativeArch,
			WorkerCPUCores:              workerCPUCores,
			WorkerCPUMillis:             workerCPUMillis,
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
			BuildID:     buildID,
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
	buildID := validation.NormalizeBuildID(req.BuildId)
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	if req.GetFlutterConfig() != nil {
		return s.handleFlutterBuild(ctx, req, buildID)
	}

	if req.GetUnityConfig() != nil {
		return s.handleUnityBuild(ctx, req, buildID)
	}

	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "build type not implemented yet",
	}, nil
}

func (s *Server) handleFlutterBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
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
				BuildID:   buildID,
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "flutter",
				BuildID:      buildID,
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
				BuildID:   buildID,
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "flutter",
				BuildID:      buildID,
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
			BuildID:   buildID,
			Status:    "running",
			WorkerID:  worker.ID,
			StartedAt: taskStartTime.Unix(),
		})
	}

	client := pb.NewBuildServiceClient(conn)
	buildResp, err := client.Build(ctx, &pb.BuildRequest{
		TaskId:         req.TaskId,
		BuildId:        req.BuildId,
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
			BuildID:     buildID,
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

func (s *Server) handleUnityBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
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
				BuildID:   buildID,
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "unity",
				BuildID:      buildID,
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
				BuildID:   buildID,
				Status:    "running",
				WorkerID:  "",
				StartedAt: taskStart,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:           req.TaskId,
				BuildType:    "unity",
				BuildID:      buildID,
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
			BuildID:   buildID,
			Status:    "running",
			WorkerID:  worker.ID,
			StartedAt: taskStartTime.Unix(),
		})
	}

	client := pb.NewBuildServiceClient(conn)
	buildResp, err := client.Build(ctx, &pb.BuildRequest{
		TaskId:         req.TaskId,
		BuildId:        req.BuildId,
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
			BuildID:     buildID,
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
