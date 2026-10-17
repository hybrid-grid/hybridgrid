package scheduler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// PluginScheduler delegates worker selection to an external scheduler
// plugin process over gRPC (contract schedulerv1alpha1, see
// proto/hybridgrid/v1/scheduler_plugin.proto). Select is bounded by a
// per-call timeout; any error, timeout, empty answer, or unknown
// worker id falls back to the configured inner scheduler, so a broken
// or lagging plugin can never stall dispatch. Feedback is
// asynchronous and lossy by design: RecordOutcome enqueues without
// blocking and a slow plugin drops feedback instead of delaying task
// completion.
type PluginScheduler struct {
	client   pb.SchedulerPluginClient
	conn     *grpc.ClientConn
	inner    Scheduler
	reg      registry.Registry
	circuit  CircuitChecker
	timeout  time.Duration
	feedback chan *pb.FeedbackRequest
	done     chan struct{}
}

// PluginConfig configures NewPluginScheduler.
type PluginConfig struct {
	// Address of the plugin's gRPC endpoint (host:port or unix:path).
	Address string

	// Timeout bounds each Select call. Default 50ms when zero.
	Timeout time.Duration

	// Inner is the fallback scheduler used when the plugin fails,
	// times out, or answers with an unknown worker. Required.
	Inner Scheduler

	// Registry supplies the worker fleet snapshots.
	Registry registry.Registry

	// Circuit optionally reports open circuit breakers.
	Circuit CircuitChecker

	// Insecure disables TLS to the plugin (plaintext local socket).
	Insecure bool
}

// feedbackBuffer bounds queued feedback events; overflow drops.
const feedbackBuffer = 512

// NewPluginScheduler dials the plugin and starts the feedback pump.
func NewPluginScheduler(cfg PluginConfig) (*PluginScheduler, error) {
	if cfg.Inner == nil {
		return nil, fmt.Errorf("plugin scheduler requires a fallback (Inner) scheduler")
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 50 * time.Millisecond
	}
	dialOpts := []grpc.DialOption{grpc.WithBlock()}
	if cfg.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx, cfg.Address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("failed to dial scheduler plugin %s: %w", cfg.Address, err)
	}
	s := &PluginScheduler{
		client:   pb.NewSchedulerPluginClient(conn),
		conn:     conn,
		inner:    cfg.Inner,
		reg:      cfg.Registry,
		circuit:  cfg.Circuit,
		timeout:  cfg.Timeout,
		feedback: make(chan *pb.FeedbackRequest, feedbackBuffer),
		done:     make(chan struct{}),
	}
	go s.feedbackPump()
	return s, nil
}

// Select chooses a worker via the plugin, falling back to the inner
// scheduler on any plugin failure.
func (s *PluginScheduler) Select(buildType pb.BuildType, arch pb.Architecture, clientOS string) (*registry.WorkerInfo, error) {
	w, _, err := s.SelectWithDispatchInfo(buildType, arch, clientOS, TaskContext{})
	return w, err
}

// SelectWithDispatchInfo implements LearningScheduler. The plugin
// contract has no dispatch introspection, so DispatchInfo is
// zero-valued (the same as non-learning in-process schedulers).
func (s *PluginScheduler) SelectWithDispatchInfo(buildType pb.BuildType, arch pb.Architecture, clientOS string, taskCtx TaskContext) (*registry.WorkerInfo, DispatchInfo, error) {
	workers := s.reg.List()
	snapshots := make([]*pb.WorkerSnapshot, 0, len(workers))
	byID := make(map[string]*registry.WorkerInfo, len(workers))
	for _, w := range workers {
		byID[w.ID] = w
		snapshots = append(snapshots, workerSnapshot(w, s.circuit))
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	resp, err := s.client.Select(ctx, &pb.SelectRequest{
		BuildType:  buildTypeString(buildType),
		TargetArch: arch.String(),
		ClientOs:   clientOS,
		Context:    pluginTaskContext(taskCtx),
		Workers:    snapshots,
	})
	if err != nil {
		log.Debug().Err(err).Str("plugin", "scheduler").
			Msg("scheduler plugin Select failed; falling back")
		w, ferr := s.inner.Select(buildType, arch, clientOS)
		return w, DispatchInfo{}, ferr
	}
	if w, ok := byID[resp.GetWorkerId()]; ok && w != nil {
		return w, DispatchInfo{}, nil
	}
	log.Debug().Str("worker_id", resp.GetWorkerId()).
		Msg("scheduler plugin answered with an unknown worker; falling back")
	w, ferr := s.inner.Select(buildType, arch, clientOS)
	return w, DispatchInfo{}, ferr
}

// RecordOutcome feeds a completion to the plugin without ever blocking
// the caller: the event is dropped when the queue is full.
func (s *PluginScheduler) RecordOutcome(workerID string, reward float64, success bool, taskCtx TaskContext) {
	event := &pb.FeedbackRequest{
		BuildType:  "", // filled by callers that know it; reward carries the signal
		WorkerId:   workerID,
		Success:    success,
		DurationMs: -reward, // reward convention: negative latency (ms)
		Context:    pluginTaskContext(taskCtx),
	}
	select {
	case s.feedback <- event:
	default:
		// Drop: feedback is best-effort by contract.
	}
}

// RecordOutcomeFor routes a completion with its build type attached
// (used by the gRPC layer when it knows the task's type).
func (s *PluginScheduler) RecordOutcomeFor(buildType pb.BuildType, workerID string, reward float64, success bool, taskCtx TaskContext) {
	event := &pb.FeedbackRequest{
		BuildType:  buildTypeString(buildType),
		WorkerId:   workerID,
		Success:    success,
		DurationMs: -reward,
		Context:    pluginTaskContext(taskCtx),
	}
	select {
	case s.feedback <- event:
	default:
	}
}

// feedbackPump drains the queue to the plugin, one event at a time;
// slow plugins fall behind and the buffer absorbs bursts up to
// feedbackBuffer, after which events drop rather than block.
func (s *PluginScheduler) feedbackPump() {
	for {
		select {
		case event := <-s.feedback:
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			_, err := s.client.Feedback(ctx, event)
			cancel()
			if err != nil {
				log.Debug().Err(err).Str("plugin", "scheduler").
					Msg("scheduler plugin Feedback failed (dropped)")
			}
		case <-s.done:
			return
		}
	}
}

// Close stops the feedback pump and the connection.
func (s *PluginScheduler) Close() {
	close(s.done)
	_ = s.conn.Close()
}

// workerSnapshot projects a registry worker into the wire snapshot.
func workerSnapshot(w *registry.WorkerInfo, circuit CircuitChecker) *pb.WorkerSnapshot {
	snap := &pb.WorkerSnapshot{
		Id:          w.ID,
		Address:     w.Address,
		ActiveTasks: uint32(w.ActiveTasks),
		MaxParallel: uint32(w.MaxParallel),
		Healthy:     w.State.String() != "unhealthy",
		CpuCores:    uint32(w.Capabilities.GetCpuCores()),
		NativeOs:    w.Capabilities.GetOs(),
		NativeArch:  w.Capabilities.GetNativeArch().String(),
		BuildTypes:  buildTypesOf(w),
	}
	if circuit != nil {
		snap.CircuitOpen = circuit.IsOpen(w.ID)
	}
	return snap
}

// buildTypesOf lists the build types a worker serves.
func buildTypesOf(w *registry.WorkerInfo) []string {
	caps := w.Capabilities
	var types []string
	if caps == nil {
		return types
	}
	if caps.GetCpp() != nil {
		types = append(types, "cpp")
	}
	if caps.GetFlutter() != nil {
		types = append(types, "flutter")
	}
	if caps.GetUnity() != nil {
		types = append(types, "unity")
	}
	if caps.GetCocos() != nil {
		types = append(types, "cocos")
	}
	if caps.GetRust() != nil {
		types = append(types, "rust")
	}
	if caps.GetGo() != nil {
		types = append(types, "go")
	}
	types = append(types, caps.GetCustomBuilders()...)
	return types
}

// buildTypeString maps the proto enum to the wire string.
func buildTypeString(bt pb.BuildType) string {
	s := bt.String()
	return strings.ToLower(strings.TrimPrefix(s, "BUILD_TYPE_"))
}

// pluginTaskContext converts the scheduler TaskContext to the wire
// form.
func pluginTaskContext(ctx TaskContext) *pb.PluginTaskContext {
	return &pb.PluginTaskContext{
		SourceSizeBytes:    int64(ctx.SourceSizeBytes),
		RawSourceSizeBytes: int64(ctx.RawSourceSizeBytes),
		SourceFilename:     ctx.SourceFilename,
		Compiler:           ctx.Compiler,
	}
}
