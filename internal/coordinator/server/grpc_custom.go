package server

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
)

// handleCustomBuild routes a plugin-builder build (BuildRequest.
// custom_builder) to a worker advertising that builder, under the same
// backpressure discipline as the other archive builds. There is
// deliberately no coordinator-side result cache for custom builders:
// cache keys are builder-specific semantics, so v1 leaves caching to
// the builder itself (it owns its outputs and can fingerprint them).
func (s *Server) handleCustomBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	start := time.Now()
	m := metrics.Default()

	builderID := req.CustomBuilder
	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	// Per-builder result cache: keyed on builder id + the worker-side
	// builder content hash + source hash, so a repeat run with the
	// same project and builder version skips the worker entirely.
	cacheKey := s.builderCacheKey(builderID, req.SourceHash, req.SourceArchive)
	if cacheKey != "" {
		if cached, ok := s.builderCacheLookup(cacheKey); ok {
			atomic.AddInt64(&s.cacheHits, 1)
			atomic.AddInt64(&s.builderCacheHits, 1)
			atomic.AddInt64(&s.totalTasks, 1)
			atomic.AddInt64(&s.successTasks, 1)
			taskStartMs := start.UnixMilli()
			if s.eventNotifier != nil {
				s.eventNotifier.NotifyTaskStarted(&TaskEvent{
					ID:          req.TaskId,
					BuildType:   builderID,
					BuildID:     buildID,
					Status:      "running",
					WorkerID:    "",
					StartedAtMs: taskStartMs,
				})
				s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
					ID:            req.TaskId,
					BuildType:     builderID,
					BuildID:       buildID,
					Status:        "completed",
					WorkerID:      "",
					StartedAtMs:   taskStartMs,
					CompletedAtMs: time.Now().UnixMilli(),
					DurationMs:    cached.GetBuildTimeMs(),
					QueueTimeMs:   0,
					CompileTimeMs: cached.GetBuildTimeMs(),
					ExitCode:      0,
					FromCache:     true,
				})
			}
			log.Info().Str("builder", builderID).Str("task_id", req.TaskId).
				Msg("Custom build served from builder cache")
			// The stored response has FromCache unset (the worker
			// built it); mark the replay so dashboards count it.
			cached.FromCache = true
			return cached, nil
		}
		atomic.AddInt64(&s.cacheMisses, 1)
	}

	worker, queueWait, err := s.selectBuildWorker(ctx,
		func(w *registry.WorkerInfo) bool {
			return w != nil && w.Capabilities != nil &&
				containsString(w.Capabilities.GetCustomBuilders(), builderID)
		}, builderID)
	if err != nil {
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.failedTasks, 1)
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("builder", builderID).
			Msg("No worker available for custom build")

		if s.eventNotifier != nil {
			taskStartMs := start.UnixMilli()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:          req.TaskId,
				BuildType:   builderID,
				BuildID:     buildID,
				Status:      "running",
				WorkerID:    "",
				StartedAtMs: taskStartMs,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:            req.TaskId,
				BuildType:     builderID,
				BuildID:       buildID,
				Status:        "failed",
				WorkerID:      "",
				StartedAtMs:   taskStartMs,
				CompletedAtMs: time.Now().UnixMilli(),
				DurationMs:    0,
				ExitCode:      1,
				ErrorMessage:  fmt.Sprintf("no worker available: %v", err),
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

	atomic.AddInt64(&s.activeTasks, 1)
	atomic.AddInt64(&s.totalTasks, 1)

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
			ID:          req.TaskId,
			BuildType:   builderID,
			BuildID:     buildID,
			Status:      "running",
			WorkerID:    worker.ID,
			StartedAtMs: taskStartTime.UnixMilli(),
		})
	}

	client := pb.NewBuildServiceClient(conn)
	buildResp, err := client.Build(ctx, req)

	success := err == nil && buildResp != nil && buildResp.Status == pb.TaskStatus_STATUS_COMPLETED

	s.registry.DecrementTasks(worker.ID, success, time.Duration(0))
	s.signalDispatchCapacity()

	taskCompletedTime := time.Now()

	if success {
		atomic.AddInt64(&s.successTasks, 1)
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
	}

	if buildResp != nil {
		s.console.Put(req.TaskId, buildResp.Stdout, buildResp.Stderr)
	}
	if s.eventNotifier != nil {
		event := &TaskEvent{
			ID:            req.TaskId,
			BuildType:     builderID,
			BuildID:       buildID,
			WorkerID:      worker.ID,
			StartedAtMs:   taskStartTime.UnixMilli(),
			CompletedAtMs: taskCompletedTime.UnixMilli(),
			DurationMs:    taskCompletedTime.Sub(taskStartTime).Milliseconds(),
			QueueTimeMs:   queueWait.Milliseconds(),
			CompileTimeMs: buildResp.GetBuildTimeMs(),
			FromCache:     false,
		}
		if success {
			event.Status = "completed"
			event.ExitCode = buildResp.ExitCode
		} else {
			event.Status = "failed"
			event.ExitCode = 1
			if err != nil {
				event.ErrorMessage = err.Error()
			} else {
				event.ExitCode = buildResp.GetExitCode()
				event.ErrorMessage = buildResp.Stderr
			}
		}
		s.eventNotifier.NotifyTaskCompleted(event)
	}

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).
			Msg("Worker custom build failed")
		return &pb.BuildResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	// Success with an active cache key: persist for future identical
	// runs. Only completed responses are cached — partial failures
	// must not poison the key.
	if cacheKey != "" && buildResp != nil && buildResp.Status == pb.TaskStatus_STATUS_COMPLETED && buildResp.ExitCode == 0 {
		s.builderCacheStore(cacheKey, buildResp)
	}

	return buildResp, nil
}

// containsString reports whether s contains v.
func containsString(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}

// builderCacheKey builds the per-builder result cache key from the
// builder id, the advertising worker's builder content hash, and the
// source hash (falling back to hashing the uploaded archive). The
// components are hashed together: the raw composite contains ":"
// separators and arbitrary client hashes, which the Windows cache-key
// validator rejects. Empty when the cache is off or no version/hash is
// known — both mean "do not cache".
func (s *Server) builderCacheKey(builderID, sourceHash string, archive []byte) string {
	if s.builderCache == nil {
		return ""
	}
	if sourceHash == "" && len(archive) > 0 {
		sourceHash = fmt.Sprintf("%x", sha256.Sum256(archive))
	}
	if sourceHash == "" {
		return ""
	}
	version := ""
	for _, w := range s.registry.List() {
		if w.Capabilities == nil || !containsString(w.Capabilities.GetCustomBuilders(), builderID) {
			continue
		}
		if v := w.Capabilities.GetCustomBuilderVersions()[builderID]; v != "" {
			version = v
			break
		}
	}
	if version == "" {
		return ""
	}
	h := sha256.Sum256([]byte(builderID + "\x00" + version + "\x00" + sourceHash))
	return "builder-cache_" + hex.EncodeToString(h[:])
}

// builderCacheLookup returns a cached BuildResponse for key.
func (s *Server) builderCacheLookup(key string) (*pb.BuildResponse, bool) {
	data, ok := s.builderCache.GetBytes(key)
	if !ok {
		return nil, false
	}
	var resp pb.BuildResponse
	if err := (protojson.UnmarshalOptions{}).Unmarshal(data, &resp); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Builder cache entry unreadable; treating as miss")
		return nil, false
	}
	return &resp, true
}

// builderCacheStore persists a successful BuildResponse under key.
func (s *Server) builderCacheStore(key string, resp *pb.BuildResponse) {
	data, err := (protojson.MarshalOptions{}).Marshal(resp)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Builder cache store failed to marshal")
		return
	}
	if err := s.builderCache.PutBytes(key, data); err != nil {
		log.Warn().Err(err).Str("key", key).Msg("Builder cache store failed")
	}
}
