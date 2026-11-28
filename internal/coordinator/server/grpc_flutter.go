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

type flutterCacheEntry = archiveCacheEntry

func (s *Server) handleFlutterBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	start := time.Now()
	meta := buildTaskMeta{taskID: req.TaskId, buildID: buildID, buildType: "flutter"}

	flutterConfig := req.GetFlutterConfig()
	flutterVersion := flutterConfig.GetFlutterVersion()
	cacheKey := cache.FlutterCacheKey(flutterConfig, req.SourceHash, flutterVersion)

	if cached := s.getFlutterCache(cacheKey); cached != nil {
		s.recordCacheServed(meta, start, cached.buildTimeMs, &s.flutterCacheHits, &s.flutterBuilds)
		return cached.response(), nil
	}

	atomic.AddInt64(&s.cacheMisses, 1)
	atomic.AddInt64(&s.flutterCacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	filter := func(w *registry.WorkerInfo) bool { return workerSupportsFlutterPlatform(w, req.TargetPlatform) }
	worker, queueWait, err := s.selectBuildWorker(ctx, filter, "flutter")
	if err != nil {
		s.countNoWorker()
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("target_platform", req.TargetPlatform.String()).
			Msg("No worker available for flutter build")
		return s.noWorkerResponse(meta, start, err), nil
	}

	worker, conn, failed := s.connectBuildWorker(ctx, meta, start, worker, filter, "flutter")
	if failed != nil {
		return failed, nil
	}

	release := s.beginWorkerTask(worker.ID, &s.flutterBuilds)
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
		s.setFlutterCache(cacheKey, buildResp)
	}

	return buildResp, nil
}

func (s *Server) getFlutterCache(key string) *flutterCacheEntry {
	s.flutterCacheMu.RLock()
	entry := s.flutterCache[key]
	s.flutterCacheMu.RUnlock()
	return entry
}

func (s *Server) setFlutterCache(key string, resp *pb.BuildResponse) {
	entry := newArchiveCacheEntry(resp)
	s.flutterCacheMu.Lock()
	s.flutterCache[key] = entry
	s.flutterCacheMu.Unlock()
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
