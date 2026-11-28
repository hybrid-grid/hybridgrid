package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
)

type unityCacheEntry struct {
	artifacts    []byte
	artifactList []*pb.ArtifactInfo
	stdout       string
	stderr       string
	buildTimeMs  int64
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
			taskStartMs := start.UnixMilli()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:          req.TaskId,
				BuildType:   "unity",
				BuildID:     buildID,
				Status:      "running",
				WorkerID:    "",
				StartedAtMs: taskStartMs,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:            req.TaskId,
				BuildType:     "unity",
				BuildID:       buildID,
				Status:        "completed",
				WorkerID:      "",
				StartedAtMs:   taskStartMs,
				CompletedAtMs: time.Now().UnixMilli(),
				DurationMs:    cached.buildTimeMs,
				QueueTimeMs:   0,
				CompileTimeMs: cached.buildTimeMs,
				ExitCode:      0,
				FromCache:     true,
				ErrorMessage:  "",
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

	worker, queueWait, err := s.selectBuildWorker(ctx,
		func(w *registry.WorkerInfo) bool { return workerSupportsUnity(w, req.TargetPlatform, unityVersion) }, "unity")
	if err != nil {
		atomic.AddInt64(&s.totalTasks, 1)
		atomic.AddInt64(&s.failedTasks, 1)
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("target_platform", req.TargetPlatform.String()).
			Msg("No worker available for unity build")

		if s.eventNotifier != nil {
			taskStartMs := start.UnixMilli()
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:          req.TaskId,
				BuildType:   "unity",
				BuildID:     buildID,
				Status:      "running",
				WorkerID:    "",
				StartedAtMs: taskStartMs,
			})
			s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
				ID:            req.TaskId,
				BuildType:     "unity",
				BuildID:       buildID,
				Status:        "failed",
				WorkerID:      "",
				StartedAtMs:   taskStartMs,
				CompletedAtMs: time.Now().UnixMilli(),
				DurationMs:    0,
				ExitCode:      1,
				FromCache:     false,
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
			ID:          req.TaskId,
			BuildType:   "unity",
			BuildID:     buildID,
			Status:      "running",
			WorkerID:    worker.ID,
			StartedAtMs: taskStartTime.UnixMilli(),
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
	// A slot just freed (flutter/unity) — queued cpp Compile() waiters
	// can use it (ActiveTasks accounting is shared across build types).
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
			BuildType:     "unity",
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

// workerSupportsUnity checks platform and, when the request pins an editor
// version, that the worker advertises it. A worker with no versions listed
// can only be vetted by platform, so it passes the version check.
func workerSupportsUnity(worker *registry.WorkerInfo, platform pb.TargetPlatform, version string) bool {
	if worker == nil || worker.Capabilities == nil || worker.Capabilities.Unity == nil {
		return false
	}
	caps := worker.Capabilities.Unity
	if version != "" && len(caps.Versions) > 0 {
		found := false
		for _, v := range caps.Versions {
			if v == version {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if len(caps.BuildTargets) == 0 {
		return true
	}
	for _, p := range caps.BuildTargets {
		if p == platform {
			return true
		}
	}
	return false
}
