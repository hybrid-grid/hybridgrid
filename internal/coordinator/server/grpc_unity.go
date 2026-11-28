package server

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

type unityCacheEntry = archiveCacheEntry

func (s *Server) handleUnityBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	start := time.Now()
	meta := buildTaskMeta{taskID: req.TaskId, buildID: buildID, buildType: "unity"}

	unityConfig := req.GetUnityConfig()
	unityVersion := unityConfig.GetUnityVersion()
	cacheKey := cache.UnityCacheKey(unityConfig, req.SourceHash, unityVersion, req.TargetPlatform)

	if cached := s.getUnityCache(cacheKey); cached != nil {
		s.recordCacheServed(meta, start, cached.buildTimeMs, &s.unityCacheHits, &s.unityBuilds)
		return cached.response(), nil
	}

	atomic.AddInt64(&s.cacheMisses, 1)
	atomic.AddInt64(&s.unityCacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	worker, queueWait, err := s.selectBuildWorker(ctx,
		func(w *registry.WorkerInfo) bool { return workerSupportsUnity(w, req.TargetPlatform, unityVersion) }, "unity")
	if err != nil {
		s.countNoWorker()
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("target_platform", req.TargetPlatform.String()).
			Msg("No worker available for unity build")
		return s.noWorkerResponse(meta, start, err), nil
	}

	conn, failed := s.connectBuildWorker(ctx, meta, start, worker)
	if failed != nil {
		return failed, nil
	}

	release := s.beginWorkerTask(worker.ID, &s.unityBuilds)
	defer release()

	taskStartTime := time.Now()
	s.notifyBuildStarted(meta, worker.ID, taskStartTime)

	buildResp, err := pb.NewBuildServiceClient(conn).Build(ctx, archiveForwardRequest(req))

	s.finishWorkerBuild(meta, worker.ID, taskStartTime, queueWait, buildResp, err)

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).
			Msg("Worker build failed")
		return workerErrorResponse(err), nil
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
	entry := newArchiveCacheEntry(resp)
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
