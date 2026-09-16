package telemetry

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// Service implements the gRPC TelemetryService over a Store and a
// StatsProvider. It is read-only: no method can alter build-plane
// state. When Token is non-empty every request must carry it,
// mirroring the build plane's per-request auth.
type Service struct {
	pb.UnimplementedTelemetryServiceServer

	store    *Store
	provider StatsProvider

	Token string

	subsMu sync.Mutex
	subs   map[chan *pb.TelemetryEvent]struct{}
}

// NewService creates a telemetry service. provider may be nil, in
// which case stats and worker listings report empty snapshots.
func NewService(store *Store, provider StatsProvider) *Service {
	return &Service{
		store:    store,
		provider: provider,
		subs:     make(map[chan *pb.TelemetryEvent]struct{}),
	}
}

// CreateEventNotifier creates task-event callbacks with the same
// signatures the embedded dashboard's notifier uses, so the
// coordinator can feed both from one hook. Timestamps are Unix
// milliseconds.
func (s *Service) CreateEventNotifier() (onStart func(id, buildID, buildType, status, workerID string, startedAtMs int64), onComplete func(id, buildID, buildType, status, workerID string, startedAtMs, completedAtMs, durationMs, queueMs, compileMs int64, exitCode int32, errorMsg string)) {
	onStart = func(id, buildID, buildType, status, workerID string, startedAtMs int64) {
		s.RecordAndNotify(&TaskInfo{
			ID:          id,
			BuildID:     buildID,
			BuildType:   buildType,
			Status:      status,
			WorkerID:    workerID,
			StartedAtMs: startedAtMs,
		})
	}
	onComplete = func(id, buildID, buildType, status, workerID string, startedAtMs, completedAtMs, durationMs, queueMs, compileMs int64, exitCode int32, errorMsg string) {
		s.RecordAndNotify(&TaskInfo{
			ID:            id,
			BuildID:       buildID,
			BuildType:     buildType,
			Status:        status,
			WorkerID:      workerID,
			StartedAtMs:   startedAtMs,
			CompletedAtMs: completedAtMs,
			DurationMs:    durationMs,
			QueueMs:       queueMs,
			CompileMs:     compileMs,
			ExitCode:      exitCode,
			ErrorMessage:  errorMsg,
		})
	}
	return
}

// RunStatsLoop broadcasts a stats event to stream subscribers every
// two seconds until ctx is cancelled, mirroring the cadence the
// embedded dashboard's WebSocket feed uses.
func (s *Service) RunStatsLoop(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if s.provider == nil {
				continue
			}
			payload, err := json.Marshal(s.provider.GetStats())
			if err != nil {
				log.Error().Err(err).Msg("telemetry: failed to marshal stats event")
				continue
			}
			s.broadcast(&pb.TelemetryEvent{
				Type:        "stats",
				Timestamp:   time.Now().Unix(),
				PayloadJson: string(payload),
			})
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) authorize(token string) error {
	if s.Token != "" && token != s.Token {
		return status.Error(codes.Unauthenticated, "invalid auth token")
	}
	return nil
}

func (s *Service) broadcast(ev *pb.TelemetryEvent) {
	s.subsMu.Lock()
	defer s.subsMu.Unlock()
	for ch := range s.subs {
		select {
		case ch <- ev:
		default:
			// Subscriber too slow; drop rather than block the plane.
		}
	}
}

// notifyTask pushes a task event to stream subscribers.
func (s *Service) notifyTask(task *TaskInfo) {
	evType := "task_started"
	if task.Status != "running" {
		evType = "task_completed"
	}
	payload, err := json.Marshal(task)
	if err != nil {
		log.Error().Err(err).Msg("telemetry: failed to marshal task event")
		return
	}
	s.broadcast(&pb.TelemetryEvent{
		Type:        evType,
		Timestamp:   time.Now().Unix(),
		PayloadJson: string(payload),
	})
}

// GetStats returns a cluster statistics snapshot.
func (s *Service) GetStats(ctx context.Context, req *pb.GetTelemetryStatsRequest) (*pb.GetTelemetryStatsResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	resp := &pb.GetTelemetryStatsResponse{Stats: &pb.TelemetryStats{Timestamp: time.Now().Unix()}}
	if s.provider != nil {
		resp.Stats = statsToProto(s.provider.GetStats())
	}
	return resp, nil
}

// ListWorkers returns the registered workers.
func (s *Service) ListWorkers(ctx context.Context, req *pb.ListTelemetryWorkersRequest) (*pb.ListTelemetryWorkersResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	resp := &pb.ListTelemetryWorkersResponse{Workers: []*pb.TelemetryWorker{}}
	if s.provider != nil {
		for _, w := range s.provider.GetWorkers() {
			resp.Workers = append(resp.Workers, workerToProto(w))
		}
	}
	return resp, nil
}

// ListTasks returns retained tasks, newest first, optionally bounded.
func (s *Service) ListTasks(ctx context.Context, req *pb.ListTelemetryTasksRequest) (*pb.ListTelemetryTasksResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	tasks := s.store.Tasks()
	if limit := int(req.GetLimit()); limit > 0 && len(tasks) > limit {
		tasks = tasks[:limit]
	}
	resp := &pb.ListTelemetryTasksResponse{Tasks: make([]*pb.TelemetryTask, 0, len(tasks))}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, taskToProto(t))
	}
	return resp, nil
}

// ListBuilds returns logical build aggregates, newest first.
func (s *Service) ListBuilds(ctx context.Context, req *pb.ListTelemetryBuildsRequest) (*pb.ListTelemetryBuildsResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	builds := s.store.Builds()
	resp := &pb.ListTelemetryBuildsResponse{Builds: make([]*pb.TelemetryBuild, 0, len(builds))}
	for _, b := range builds {
		resp.Builds = append(resp.Builds, buildToProto(b))
	}
	return resp, nil
}

// GetBuild returns one build with its retained tasks.
func (s *Service) GetBuild(ctx context.Context, req *pb.GetTelemetryBuildRequest) (*pb.GetTelemetryBuildResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	build, tasks, ok := s.store.BuildDetail(req.GetBuildId())
	if !ok {
		return nil, status.Errorf(codes.NotFound, "build %q not found", req.GetBuildId())
	}
	resp := &pb.GetTelemetryBuildResponse{
		Build: buildToProto(build),
		Tasks: make([]*pb.TelemetryTask, 0, len(tasks)),
	}
	for _, t := range tasks {
		resp.Tasks = append(resp.Tasks, taskToProto(t))
	}
	return resp, nil
}

// GetConsole returns retained console output for one task.
func (s *Service) GetConsole(ctx context.Context, req *pb.GetTelemetryConsoleRequest) (*pb.GetTelemetryConsoleResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	provider, ok := s.provider.(ConsoleProvider)
	if !ok {
		return nil, status.Error(codes.Unavailable, "console output not available")
	}
	stdout, stderr, truncated, found := provider.GetConsole(req.GetTaskId())
	if !found {
		return nil, status.Errorf(codes.NotFound, "task %q not found", req.GetTaskId())
	}
	return &pb.GetTelemetryConsoleResponse{Console: &pb.TelemetryConsole{
		TaskId:    req.GetTaskId(),
		Stdout:    stdout,
		Stderr:    stderr,
		Truncated: truncated,
	}}, nil
}

// StreamEvents replays the bounded recent-task ring as events, then
// pushes live task and stats events until the client disconnects.
func (s *Service) StreamEvents(req *pb.StreamTelemetryEventsRequest, stream pb.TelemetryService_StreamEventsServer) error {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return err
	}

	ch := make(chan *pb.TelemetryEvent, 64)
	s.subsMu.Lock()
	s.subs[ch] = struct{}{}
	s.subsMu.Unlock()
	defer func() {
		s.subsMu.Lock()
		delete(s.subs, ch)
		s.subsMu.Unlock()
	}()

	for _, task := range s.store.RecentTasks() {
		evType := "task_started"
		if task.Status != "running" {
			evType = "task_completed"
		}
		payload, err := json.Marshal(task)
		if err != nil {
			continue
		}
		if err := stream.Send(&pb.TelemetryEvent{
			Type:        evType,
			Timestamp:   time.Now().Unix(),
			PayloadJson: string(payload),
		}); err != nil {
			return err
		}
	}

	for {
		select {
		case <-stream.Context().Done():
			return nil
		case ev := <-ch:
			if err := stream.Send(ev); err != nil {
				return err
			}
		}
	}
}

// RecordAndNotify records a task state change and pushes it to
// stream subscribers. The coordinator's event hooks call this via
// CreateEventNotifier; exported for tests and future producers.
func (s *Service) RecordAndNotify(task *TaskInfo) {
	s.store.RecordTask(task)
	s.notifyTask(task)
}

func statsToProto(s *Stats) *pb.TelemetryStats {
	if s == nil {
		return &pb.TelemetryStats{}
	}
	return &pb.TelemetryStats{
		TotalTasks:          s.TotalTasks,
		SuccessTasks:        s.SuccessTasks,
		FailedTasks:         s.FailedTasks,
		ActiveTasks:         s.ActiveTasks,
		QueuedTasks:         s.QueuedTasks,
		CacheHits:           s.CacheHits,
		CacheMisses:         s.CacheMisses,
		CacheHitRate:        s.CacheHitRate,
		FlutterBuilds:       s.FlutterBuilds,
		FlutterCacheHits:    s.FlutterCacheHits,
		FlutterCacheMisses:  s.FlutterCacheMisses,
		FlutterCacheHitRate: s.FlutterCacheHitRate,
		UnityBuilds:         s.UnityBuilds,
		UnityCacheHits:      s.UnityCacheHits,
		UnityCacheMisses:    s.UnityCacheMisses,
		UnityCacheHitRate:   s.UnityCacheHitRate,
		TotalWorkers:        int32(s.TotalWorkers),
		HealthyWorkers:      int32(s.HealthyWorkers),
		UptimeSeconds:       s.UptimeSeconds,
		Timestamp:           s.Timestamp,
	}
}

func workerToProto(w *WorkerInfo) *pb.TelemetryWorker {
	return &pb.TelemetryWorker{
		Id:                w.ID,
		Host:              w.Host,
		Address:           w.Address,
		Os:                w.OS,
		Architecture:      w.Architecture,
		Architectures:     w.Architectures,
		CpuCores:          w.CPUCores,
		MemoryGb:          w.MemoryGB,
		MaxParallelTasks:  w.MaxParallelTasks,
		ActiveTasks:       w.ActiveTasks,
		TotalTasks:        w.TotalTasks,
		SuccessRate:       w.SuccessRate,
		AvgLatencyMs:      w.AvgLatencyMs,
		CircuitState:      w.CircuitState,
		DiscoverySource:   w.DiscoverySource,
		Version:           w.Version,
		DockerAvailable:   w.DockerAvailable,
		FlutterAvailable:  w.FlutterAvailable,
		FlutterSdkVersion: w.FlutterSDKVersion,
		FlutterPlatforms:  w.FlutterPlatforms,
		UnityAvailable:    w.UnityAvailable,
		UnityVersions:     w.UnityVersions,
		UnityPlatforms:    w.UnityPlatforms,
		Compilers:         w.Compilers,
		BuildTypes:        w.BuildTypes,
		Healthy:           w.Healthy,
		LastSeen:          w.LastSeen,
	}
}

func taskToProto(t *TaskInfo) *pb.TelemetryTask {
	return &pb.TelemetryTask{
		Id:            t.ID,
		BuildType:     t.BuildType,
		BuildId:       t.BuildID,
		Status:        t.Status,
		WorkerId:      t.WorkerID,
		StartedAtMs:   t.StartedAtMs,
		CompletedAtMs: t.CompletedAtMs,
		DurationMs:    t.DurationMs,
		QueueMs:       t.QueueMs,
		CompileMs:     t.CompileMs,
		ExitCode:      t.ExitCode,
		FromCache:     t.FromCache,
		ErrorMessage:  t.ErrorMessage,
	}
}

func buildToProto(b *BuildInfo) *pb.TelemetryBuild {
	return &pb.TelemetryBuild{
		Id:             b.ID,
		BuildType:      b.BuildType,
		Status:         b.Status,
		TotalTasks:     int32(b.TotalTasks),
		CompletedTasks: int32(b.CompletedTasks),
		FailedTasks:    int32(b.FailedTasks),
		RunningTasks:   int32(b.RunningTasks),
		FromCacheCount: int32(b.FromCacheCount),
		FirstTaskAtMs:  b.FirstTaskAtMs,
		LastTaskAtMs:   b.LastTaskAtMs,
		Truncated:      b.Truncated,
	}
}
