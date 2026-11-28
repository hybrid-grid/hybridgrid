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
)

// handleCustomBuild routes a plugin-builder build (BuildRequest.
// custom_builder) to a worker advertising that builder, under the same
// backpressure discipline as the other archive builds. There is
// deliberately no coordinator-side result cache for custom builders:
// cache keys are builder-specific semantics, so v1 leaves caching to
// the builder itself (it owns its outputs and can fingerprint them).
func (s *Server) handleCustomBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	start := time.Now()
	builderID := req.CustomBuilder
	meta := buildTaskMeta{taskID: req.TaskId, buildID: buildID, buildType: builderID}

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	// Per-builder result cache: keyed on builder id + the worker-side
	// builder content hash + source hash, so a repeat run with the
	// same project and builder version skips the worker entirely.
	cacheKey := s.builderCacheKey(builderID, req.SourceHash, req.SourceArchive)
	if cacheKey != "" {
		if cached, ok := s.builderCacheLookup(cacheKey); ok {
			s.recordCacheServed(meta, start, cached.GetBuildTimeMs(), &s.builderCacheHits)
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
		s.countNoWorker()
		log.Error().Err(err).Str("task_id", req.TaskId).
			Str("builder", builderID).
			Msg("No worker available for custom build")
		return s.noWorkerResponse(meta, start, err), nil
	}

	conn, failed := s.connectBuildWorker(ctx, worker)
	if failed != nil {
		return failed, nil
	}

	release := s.beginWorkerTask(worker.ID)
	defer release()

	taskStartTime := time.Now()
	s.notifyBuildStarted(meta, worker.ID, taskStartTime)

	buildResp, err := pb.NewBuildServiceClient(conn).Build(ctx, req)

	s.finishWorkerBuild(meta, worker.ID, taskStartTime, queueWait, buildResp, err)

	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).
			Msg("Worker custom build failed")
		return workerErrorResponse(err), nil
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
