package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/artifacts"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/resilience"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/interceptors"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
	hgtls "github.com/h3nr1-d14z/hybridgrid/internal/security/tls"
	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

const maxGRPCMessageSize = 512 * 1024 * 1024

// messageSizeOptions lets the server accept and send messages up to
// maxGRPCMessageSize, matching what clients and the worker connections
// are configured for. Without them gRPC's 4 MiB default rejects any
// Flutter, Unity or plugin build whose source archive exceeds 4 MiB.
func messageSizeOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.MaxRecvMsgSize(maxGRPCMessageSize),
		grpc.MaxSendMsgSize(maxGRPCMessageSize),
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
	// Valid: "leastloaded" (default), "sed", "p90", "sed-p90", "ewma", "simple", "p2c",
	// "epsilon-greedy", "linucb", "hybrid-linucb", "hybrid-linucb-d", "heft", "icecc-fastest".
	SchedulerType string
	// SizeThreshold is the TaskContext.SourceSizeBytes cutoff in bytes for p90 arms.
	SizeThreshold int
	// EpsilonValue is the exploration rate for epsilon-greedy. Ignored
	// for other schedulers. Default 0.1 (Sutton & Barto §2.3 baseline).
	EpsilonValue float64
	// AlphaValue is the LinUCB exploration coefficient α (Li 2010 Eq. 4).
	// Theoretical form is 1 + sqrt(ln(2/δ)/2); we default to 1.0 and
	// expect empirical tuning. Ignored for non-LinUCB schedulers.
	AlphaValue float64
	// FeatureSetValue selects full or compact LinUCB features.
	FeatureSetValue string
	// RewardCenterValue selects the residual reward baseline for hybrid LinUCB.
	RewardCenterValue string
	// DiagCandidates records the score decomposition for each LinUCB candidate.
	DiagCandidates bool
	// EwmaAlphaValue is the EWMA scheduler's smoothing factor.
	EwmaAlphaValue float64
	// WarmStartTasks is the hybrid-linucb and ewma warm-start window: that
	// many initial dispatches are routed least-loaded while learning from
	// outcomes. 0 disables warm-start. Ignored by other schedulers.
	WarmStartTasks int
	// LoadPenaltyValue is the hybrid-linucb λ in p = mean + bonus −
	// λ·loadRatio. 0 disables the penalty. Ignored by other schedulers.
	LoadPenaltyValue float64
	// DiscountValue is the HG-LinUCB-D discount gamma. Zero uses 0.98 in the factory.
	DiscountValue float64
	// DiscountModeValue selects the global or per-arm discount clock.
	DiscountModeValue string
	// TaskLogPath is the path to the JSON Lines per-task log file.
	// Empty or "stdout" routes records to standard output.
	TaskLogPath string
	// DispatchQueueTimeout bounds how long a Compile() call may wait
	// in the dispatch queue after the fast dispatch budget (~700 ms)
	// is exhausted, before the coordinator gives up and returns the
	// "no worker available" failure. 0 disables queueing entirely
	// (ablation: the pre-backpressure reject-after-budget behavior).
	DispatchQueueTimeout time.Duration
	// BuilderCacheDir persists per-builder build results (keyed by
	// builder id + builder content hash + source hash) so repeat
	// plugin-builder runs with unchanged inputs skip the worker.
	// Empty disables the cache (default: off).
	BuilderCacheDir string
	// SchedulerPluginAddr is the gRPC endpoint of an external scheduler
	// plugin (contract schedulerv1alpha1). Used when SchedulerType is
	// "external"; the plugin decides dispatches within
	// SchedulerPluginTimeout and the in-process fallback (leastloaded)
	// answers when it fails, times out, or returns an unknown worker.
	SchedulerPluginAddr string
	// SchedulerPluginTLS enables TLS/mTLS to the external scheduler
	// plugin. Nil dials plaintext (local socket).
	SchedulerPluginTLS *tls.Config
	// SchedulerPluginTimeout bounds each plugin Select call. Default
	// 50ms when zero.
	SchedulerPluginTimeout time.Duration
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

		DiscountValue:     0.98,
		DiscountModeValue: "global",
		FeatureSetValue:   "full",
		RewardCenterValue: "none",
		EwmaAlphaValue:    0.2,

		DispatchQueueTimeout: 30 * time.Second,
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
			FeatureSet:     cfg.FeatureSetValue,
			RewardCenter:   cfg.RewardCenterValue,
			DiagCandidates: cfg.DiagCandidates,
		})
	case "hybrid-linucb":
		// Pass-through without zero-defaulting so --warm-start=0 or
		// --load-penalty=0 genuinely disables each mechanism for
		// ablation runs.
		return scheduler.NewLinUCBScheduler(scheduler.LinUCBConfig{
			Registry:       reg,
			CircuitChecker: cm,
			Alpha:          cfg.AlphaValue,
			FeatureSet:     cfg.FeatureSetValue,
			RewardCenter:   cfg.RewardCenterValue,
			DiagCandidates: cfg.DiagCandidates,
			WarmStartTasks: cfg.WarmStartTasks,
			LoadPenalty:    cfg.LoadPenaltyValue,
		})
	case "hybrid-linucb-d":
		discount := cfg.DiscountValue
		if discount == 0 {
			discount = 0.98
		}
		mode := cfg.DiscountModeValue
		if mode == "" {
			mode = string(scheduler.DiscountModeGlobal)
		}
		return scheduler.NewLinUCBScheduler(scheduler.LinUCBConfig{
			Registry:       reg,
			CircuitChecker: cm,
			Alpha:          cfg.AlphaValue,
			FeatureSet:     cfg.FeatureSetValue,
			RewardCenter:   cfg.RewardCenterValue,
			DiagCandidates: cfg.DiagCandidates,
			WarmStartTasks: cfg.WarmStartTasks,
			LoadPenalty:    cfg.LoadPenaltyValue,
			Discount:       discount,
			DiscountMode:   scheduler.DiscountMode(mode),
		})
	case "heft":
		return scheduler.NewHEFTScheduler(scheduler.HEFTConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "sed":
		return scheduler.NewSEDScheduler(scheduler.SEDConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "ewma":
		return scheduler.NewEWMAScheduler(scheduler.EWMAConfig{
			Registry:       reg,
			CircuitChecker: cm,
			WarmStartTasks: cfg.WarmStartTasks,
			Alpha:          cfg.EwmaAlphaValue,
		})
	case "p90":
		return scheduler.NewP90Scheduler(reg, cfg.SizeThreshold)
	case "sed-p90":
		return scheduler.NewSEDP90Scheduler(scheduler.SEDConfig{
			Registry:       reg,
			CircuitChecker: cm,
		}, cfg.SizeThreshold)
	case "icecc-fastest":
		return scheduler.NewIceccFastestScheduler(scheduler.IceccConfig{
			Registry:       reg,
			CircuitChecker: cm,
		})
	case "external":
		plugin, err := scheduler.NewPluginScheduler(scheduler.PluginConfig{
			Address:   cfg.SchedulerPluginAddr,
			Timeout:   cfg.SchedulerPluginTimeout,
			Inner:     scheduler.NewLeastLoadedScheduler(reg),
			Registry:  reg,
			Circuit:   cm,
			TLSConfig: cfg.SchedulerPluginTLS,
		})
		if err != nil {
			log.Error().Err(err).Msg("Failed to connect scheduler plugin; using leastloaded")
			return scheduler.NewLeastLoadedScheduler(reg)
		}
		return plugin
	case "leastloaded", "":
		return scheduler.NewLeastLoadedScheduler(reg)
	default:
		log.Warn().Str("requested", cfg.SchedulerType).Msg("Unknown scheduler type; falling back to leastloaded")
		return scheduler.NewLeastLoadedScheduler(reg)
	}
}

// TaskEvent represents a task event for the dashboard.
type TaskEvent struct {
	ID        string
	BuildType string
	BuildID   string
	Status    string
	WorkerID  string
	// StartedAtMs and CompletedAtMs are Unix milliseconds so the
	// dashboard can place tasks on a sub-second timeline (second
	// granularity collapses builds where every TU lands within one
	// second). The JSONL task log is unaffected — it carries its own
	// TS and durations.
	StartedAtMs   int64
	CompletedAtMs int64
	DurationMs    int64
	QueueTimeMs   int64
	CompileTimeMs int64
	ExitCode      int32
	FromCache     bool
	ErrorMessage  string
	// SourceFile and Command identify a C/C++ compile for humans: the
	// client-sent base file name and the invocation forwarded to the
	// worker. Both are empty for Flutter/Unity tasks.
	SourceFile string
	Command    string
}

// maxEventCommandLen bounds TaskEvent.Command so a pathological
// argument list cannot bloat telemetry retention or webhook payloads.
const maxEventCommandLen = 2048

// compileCommand renders the forwarded compiler invocation for display.
func compileCommand(req *pb.CompileRequest) string {
	cmd := strings.TrimSpace(req.GetCompiler() + " " + strings.Join(req.GetCompilerArgs(), " "))
	if len(cmd) > maxEventCommandLen {
		// Drop a rune split by the byte cut so the result stays valid UTF-8.
		cmd = strings.ToValidUTF8(cmd[:maxEventCommandLen], "") + "…"
	}
	return cmd
}

// EventNotifier is called when task events occur.
type EventNotifier interface {
	NotifyTaskStarted(event *TaskEvent)
	NotifyTaskCompleted(event *TaskEvent)
}

// Server implements the coordinator gRPC server.
type Server struct {
	pb.UnimplementedBuildServiceServer

	config          Config
	server          *grpc.Server
	registry        registry.Registry
	scheduler       scheduler.Scheduler
	schedulerParams string
	circuitManager  *resilience.CircuitManager
	eventNotifier   EventNotifier
	telemetry       *telemetry.Service
	artifacts       *artifacts.Store
	deps            depsRouting
	workerConns     *connPool
	taskLogger      *TaskLogger
	console         *consoleStore
	dispatchCh      chan *dispatchRequest
	dispatchWG      sync.WaitGroup
	dispatchOnce    sync.Once

	// dispatchSignal broadcasts "a slot may have freed" to Compile()
	// calls waiting in the dispatch queue (backpressure path). It is
	// closed and replaced under dispatchSignalMu; waiters select on
	// the channel they grabbed. queuedCompiles counts current waiters
	// against maxQueuedCompiles.
	dispatchSignalMu sync.Mutex
	dispatchSignal   chan struct{}
	queuedCompiles   int64
	dispatchCount    int64

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

	// builderCache persists per-builder results when BuilderCacheDir
	// is configured; nil means the cache is off.
	builderCache     *cache.Store
	builderCacheHits int64
}

// New creates a new coordinator gRPC server.
func New(cfg Config) *Server {
	reg := registry.NewInMemoryRegistry(cfg.HeartbeatTTL)
	circuitMgr := resilience.NewCircuitManager(resilience.DefaultCircuitConfig())
	sched := newScheduler(cfg, reg, circuitMgr)
	schedulerParams := ""
	if withParams, ok := sched.(interface{ Params() string }); ok {
		schedulerParams = withParams.Params()
	} else {
		// Schedulers without a Params() method: derive from cfg for
		// provenance display.
		switch cfg.SchedulerType {
		case "hybrid-linucb", "hybrid-linucb-d":
			schedulerParams = fmt.Sprintf("alpha=%s warm_start=%d load_penalty=%s",
				strconv.FormatFloat(cfg.AlphaValue, 'g', -1, 64),
				cfg.WarmStartTasks,
				strconv.FormatFloat(cfg.LoadPenaltyValue, 'g', -1, 64))
		case "epsilon-greedy":
			eps := cfg.EpsilonValue
			if eps == 0 {
				eps = 0.1 // Sutton & Barto §2.3 default, same as newScheduler
			}
			schedulerParams = "epsilon=" + strconv.FormatFloat(eps, 'g', -1, 64)
		}
	}
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
		config:          cfg,
		registry:        reg,
		scheduler:       sched,
		schedulerParams: schedulerParams,
		circuitManager:  circuitMgr,
		workerConns:     newConnPool(dialOpts),
		taskLogger:      taskLogger,
		console:         newConsoleStore(),
		flutterCache:    make(map[string]*flutterCacheEntry),
		unityCache:      make(map[string]*unityCacheEntry),
		dispatchCh:      make(chan *dispatchRequest, dispatchQueueSize),
		dispatchSignal:  make(chan struct{}),
	}
	if cfg.BuilderCacheDir != "" {
		// 512MB / 30 days mirrors the CLI cache defaults; builder
		// outputs (archives + artifacts) are small relative to .o files.
		store, err := cache.NewStore(cfg.BuilderCacheDir, 512, 24*30)
		if err != nil {
			log.Warn().Err(err).Str("dir", cfg.BuilderCacheDir).
				Msg("Builder cache disabled; starting without it")
		} else {
			s.builderCache = store
			log.Info().Str("dir", cfg.BuilderCacheDir).Msg("Builder cache enabled")
		}
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

	opts := messageSizeOptions()

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
	if s.telemetry != nil {
		pb.RegisterTelemetryServiceServer(s.server, s.telemetry)
	}

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
	// An external scheduler plugin owns a gRPC connection and a
	// feedback pump goroutine; both stop here.
	if ps, ok := s.scheduler.(*scheduler.PluginScheduler); ok {
		ps.Close()
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

// SchedulerInfo reports the active scheduler type and its parameter
// string — the same values the task log records — so out-of-process
// consumers (telemetry, dashboard) can label builds with the policy
// that produced them.
func (s *Server) SchedulerInfo() (string, string) {
	return s.config.SchedulerType, s.schedulerParams
}

// SetTelemetry attaches a telemetry service to the coordinator gRPC
// server. Must be called before Start; the service is registered on
// the same port as the build plane.
func (s *Server) SetTelemetry(svc *telemetry.Service) {
	s.telemetry = svc
}
