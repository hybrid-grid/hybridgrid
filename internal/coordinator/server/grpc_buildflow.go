package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
)

// Shared building blocks for the archive-based build handlers
// (handleFlutterBuild, handleUnityBuild, handleCustomBuild). Each
// handler keeps its own cache, worker filter, and counters; only the
// steps that are identical across them live here.

// buildTaskMeta identifies a build task in emitted TaskEvents.
type buildTaskMeta struct {
	taskID    string
	buildID   string
	buildType string
}

// archiveCacheEntry is the coordinator-side in-memory result of a
// flutter or unity build.
type archiveCacheEntry struct {
	artifacts    []byte
	artifactList []*pb.ArtifactInfo
	stdout       string
	stderr       string
	buildTimeMs  int64
}

func newArchiveCacheEntry(resp *pb.BuildResponse) *archiveCacheEntry {
	return &archiveCacheEntry{
		stdout:       resp.Stdout,
		stderr:       resp.Stderr,
		artifacts:    resp.Artifacts,
		artifactList: cloneArtifactList(resp.ArtifactList),
		buildTimeMs:  resp.BuildTimeMs,
	}
}

// response builds the cache-hit BuildResponse, copying the mutable
// parts so callers cannot alias the stored entry.
func (e *archiveCacheEntry) response() *pb.BuildResponse {
	return &pb.BuildResponse{
		Status:       pb.TaskStatus_STATUS_COMPLETED,
		ExitCode:     0,
		Stdout:       e.stdout,
		Stderr:       e.stderr,
		Artifacts:    append([]byte(nil), e.artifacts...),
		ArtifactList: cloneArtifactList(e.artifactList),
		BuildTimeMs:  e.buildTimeMs,
		FromCache:    true,
	}
}

// recordCacheServed accounts for a build answered from a coordinator
// cache: the shared hit/total/success counters, any per-type counters
// in extra, and a started+completed event pair with no worker.
func (s *Server) recordCacheServed(meta buildTaskMeta, start time.Time, buildTimeMs int64, extra ...*int64) {
	atomic.AddInt64(&s.cacheHits, 1)
	for _, c := range extra {
		atomic.AddInt64(c, 1)
	}
	atomic.AddInt64(&s.totalTasks, 1)
	atomic.AddInt64(&s.successTasks, 1)

	if s.eventNotifier == nil {
		return
	}
	startedMs := start.UnixMilli()
	s.eventNotifier.NotifyTaskStarted(startedEvent(meta, "", startedMs))
	s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
		ID:            meta.taskID,
		BuildType:     meta.buildType,
		BuildID:       meta.buildID,
		Status:        "completed",
		WorkerID:      "",
		StartedAtMs:   startedMs,
		CompletedAtMs: time.Now().UnixMilli(),
		DurationMs:    buildTimeMs,
		QueueTimeMs:   0,
		CompileTimeMs: buildTimeMs,
		ExitCode:      0,
		FromCache:     true,
		ErrorMessage:  "",
	})
}

func startedEvent(meta buildTaskMeta, workerID string, startedMs int64) *TaskEvent {
	return &TaskEvent{
		ID:          meta.taskID,
		BuildType:   meta.buildType,
		BuildID:     meta.buildID,
		Status:      "running",
		WorkerID:    workerID,
		StartedAtMs: startedMs,
	}
}

// countNoWorker bumps the counters for a build that could not be
// placed on any worker.
func (s *Server) countNoWorker() {
	atomic.AddInt64(&s.totalTasks, 1)
	atomic.AddInt64(&s.failedTasks, 1)
}

// noWorkerResponse emits the started+failed event pair for a build
// that found no worker and returns the failure response.
func (s *Server) noWorkerResponse(meta buildTaskMeta, start time.Time, err error) *pb.BuildResponse {
	msg := fmt.Sprintf("no worker available: %v", err)
	if s.eventNotifier != nil {
		startedMs := start.UnixMilli()
		s.eventNotifier.NotifyTaskStarted(startedEvent(meta, "", startedMs))
		s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
			ID:            meta.taskID,
			BuildType:     meta.buildType,
			BuildID:       meta.buildID,
			Status:        "failed",
			WorkerID:      "",
			StartedAtMs:   startedMs,
			CompletedAtMs: time.Now().UnixMilli(),
			DurationMs:    0,
			ExitCode:      1,
			FromCache:     false,
			ErrorMessage:  msg,
		})
	}
	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   msg,
	}
}

// connectBuildWorker returns a connection to worker. On failure it
// logs, settles the task as failed and returns the failure response to
// hand back to the client.
//
// selectBuildWorker already booked a registry slot on worker, so a
// failed connect must release it: otherwise every build that selects an
// unreachable worker leaks one slot until the worker looks permanently
// full. The failure is also counted and reported like any other failed
// build, so it does not vanish from the stats and the dashboard.
func (s *Server) connectBuildWorker(ctx context.Context, meta buildTaskMeta, start time.Time, worker *registry.WorkerInfo) (*grpc.ClientConn, *pb.BuildResponse) {
	conn, err := s.workerConns.get(ctx, worker.Address)
	if err != nil {
		log.Error().Err(err).Str("worker", worker.ID).Msg("Failed to connect to worker")
		msg := fmt.Sprintf("failed to connect to worker: %v", err)

		s.registry.DecrementTasks(worker.ID, false, 0)
		s.signalDispatchCapacity()
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.failedTasks, 1)
		if s.eventNotifier != nil {
			startedMs := start.UnixMilli()
			s.eventNotifier.NotifyTaskStarted(startedEvent(meta, worker.ID, startedMs))
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:            meta.taskID,
				BuildType:     meta.buildType,
				BuildID:       meta.buildID,
				Status:        "failed",
				WorkerID:      worker.ID,
				StartedAtMs:   startedMs,
				CompletedAtMs: time.Now().UnixMilli(),
				ExitCode:      1,
				ErrorMessage:  msg,
			})
		}
		return nil, &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   msg,
		}
	}
	return conn, nil
}

// beginWorkerTask marks a build as running on workerID: global and
// per-worker active counters, the active-task gauge, and the
// per-type counters in extra. The returned release func undoes the
// active accounting and must be deferred by the caller.
func (s *Server) beginWorkerTask(workerID string, extra ...*int64) (release func()) {
	m := metrics.Default()

	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)
	for _, c := range extra {
		atomic.AddInt64(c, 1)
	}

	val, _ := s.activeTasksByWorker.LoadOrStore(workerID, new(int64))
	count := atomic.AddInt64(val.(*int64), 1)
	m.SetActiveTaskCount(workerID, float64(count))

	return func() {
		atomic.AddInt64(&s.activeTasks, -1)
		if val, ok := s.activeTasksByWorker.Load(workerID); ok {
			c := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(workerID, float64(c))
		}
	}
}

// notifyBuildStarted emits the running event for a dispatched build.
func (s *Server) notifyBuildStarted(meta buildTaskMeta, workerID string, startedAt time.Time) {
	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskStarted(startedEvent(meta, workerID, startedAt.UnixMilli()))
	}
}

// archiveForwardRequest is the subset of an archive build request
// that flutter and unity forward to the worker.
func archiveForwardRequest(req *pb.BuildRequest) *pb.BuildRequest {
	return &pb.BuildRequest{
		TaskId:         req.TaskId,
		BuildId:        req.BuildId,
		SourceHash:     req.SourceHash,
		SourceArchive:  req.SourceArchive,
		BuildType:      req.BuildType,
		TargetPlatform: req.TargetPlatform,
		Config:         req.Config,
		TimeoutSeconds: req.TimeoutSeconds,
	}
}

// buildSucceeded reports whether a worker Build call completed.
func buildSucceeded(resp *pb.BuildResponse, err error) bool {
	return err == nil && resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED
}

// buildCompletionEvent builds the completion event for a build that
// ran on a worker.
func buildCompletionEvent(meta buildTaskMeta, workerID string, startedAt, completedAt time.Time, queueWait time.Duration, resp *pb.BuildResponse, err error) *TaskEvent {
	event := &TaskEvent{
		ID:            meta.taskID,
		BuildType:     meta.buildType,
		BuildID:       meta.buildID,
		WorkerID:      workerID,
		StartedAtMs:   startedAt.UnixMilli(),
		CompletedAtMs: completedAt.UnixMilli(),
		DurationMs:    completedAt.Sub(startedAt).Milliseconds(),
		QueueTimeMs:   queueWait.Milliseconds(),
		CompileTimeMs: resp.GetBuildTimeMs(),
		FromCache:     false,
	}
	if buildSucceeded(resp, err) {
		event.Status = "completed"
		event.ExitCode = resp.ExitCode
		return event
	}
	event.Status = "failed"
	event.ExitCode = 1
	if err != nil {
		event.ErrorMessage = err.Error()
	} else if resp != nil {
		event.ExitCode = resp.ExitCode
		event.ErrorMessage = resp.Stderr
	}
	return event
}

// finishWorkerBuild settles a worker Build call: releases the
// worker's registry slot, wakes queued dispatch waiters, updates the
// success/failure counters, stores console output, and emits the
// completion event (which carries meta.buildID as given).
func (s *Server) finishWorkerBuild(meta buildTaskMeta, workerID string, startedAt time.Time, queueWait time.Duration, resp *pb.BuildResponse, err error) {
	success := buildSucceeded(resp, err)

	s.registry.DecrementTasks(workerID, success, time.Duration(0))
	// A slot just freed — queued cpp Compile() waiters can use it
	// (ActiveTasks accounting is shared across build types).
	s.signalDispatchCapacity()

	completedAt := time.Now()

	if success {
		atomic.AddInt64(&s.successTasks, 1)
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
	}

	if resp != nil {
		s.console.Put(meta.taskID, resp.Stdout, resp.Stderr)
	}
	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskCompleted(
			buildCompletionEvent(meta, workerID, startedAt, completedAt, queueWait, resp, err))
	}
}

// workerErrorResponse is the failure response for a transport-level
// worker Build error.
func workerErrorResponse(err error) *pb.BuildResponse {
	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("worker error: %v", err),
	}
}
