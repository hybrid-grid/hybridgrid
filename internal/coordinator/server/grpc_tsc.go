package server

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/blobcache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

// depsRouting keeps dependency layers (node_modules) so a client
// uploads each once, and remembers which workers already hold which
// layer so it is forwarded to each machine once.
type depsRouting struct {
	store *blobcache.Cache

	mu   sync.Mutex
	held map[string]map[string]bool // worker id -> layer hash -> held
}

// SetDepsStore enables dependency layers; without a store, a request
// naming a layer must carry it every time.
func (s *Server) SetDepsStore(store *blobcache.Cache) {
	s.deps.store = store
}

func (d *depsRouting) workerHolds(workerID, hash string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.held[workerID][hash]
}

func (d *depsRouting) markHeld(workerID, hash string, held bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.held == nil {
		d.held = make(map[string]map[string]bool)
	}
	if d.held[workerID] == nil {
		d.held[workerID] = make(map[string]bool)
	}
	d.held[workerID][hash] = held
}

// accept stores a layer the client sent, or checks that a layer named
// only by hash is already here. A missing layer is FailedPrecondition:
// the client resends with the layer attached.
func (d *depsRouting) accept(req *pb.BuildRequest) error {
	hash := req.GetDepsHash()
	if hash == "" {
		return nil
	}
	if len(req.GetDepsArchive()) > 0 {
		if d.store != nil {
			if err := d.store.Put(hash, bytes.NewReader(req.GetDepsArchive())); err != nil {
				log.Warn().Err(err).Str("deps", hash).Msg("Failed to keep dependency layer; it will be forwarded from this request only")
			}
		}
		return nil
	}
	if d.store == nil || !d.store.Has(hash) {
		return status.Error(codes.FailedPrecondition, depslayer.MissingMessage)
	}
	return nil
}

// layer returns the layer bytes: from the request, else from the store.
func (d *depsRouting) layer(req *pb.BuildRequest) ([]byte, error) {
	if len(req.GetDepsArchive()) > 0 {
		return req.GetDepsArchive(), nil
	}
	if d.store == nil {
		return nil, errors.New("dependency layer store disabled")
	}
	f, _, err := d.store.Open(req.GetDepsHash())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func isMissingDeps(err error) bool {
	st, ok := status.FromError(err)
	return ok && st.Code() == codes.FailedPrecondition && st.Message() == depslayer.MissingMessage
}

func workerSupportsNode(w *registry.WorkerInfo) bool {
	return w != nil && w.Capabilities != nil && w.Capabilities.Nodejs != nil
}

// handleTscBuild runs one TypeScript project (`tsc -p`) on a worker.
// Requests carry a dependency layer by hash; it is attached to the
// forwarded request only for workers not known to hold it, and resent
// once if the worker reports it missing (e.g. its cache was evicted).
func (s *Server) handleTscBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	if err := s.deps.accept(req); err != nil {
		return nil, err
	}
	start := time.Now()
	meta := buildTaskMeta{taskID: req.TaskId, buildID: buildID, buildType: "tsc"}

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	worker, queueWait, err := s.selectBuildWorker(ctx, workerSupportsNode, "tsc")
	if err != nil {
		s.countNoWorker()
		log.Error().Err(err).Str("task_id", req.TaskId).Msg("No worker available for tsc build")
		return s.noWorkerResponse(meta, start, err), nil
	}
	worker, conn, failed := s.connectBuildWorker(ctx, meta, start, worker, workerSupportsNode, "tsc")
	if failed != nil {
		return failed, nil
	}

	release := s.beginWorkerTask(worker.ID)
	defer release()
	taskStartTime := time.Now()
	s.notifyBuildStarted(meta, worker.ID, taskStartTime)

	fwd := archiveForwardRequest(req)
	fwd.DepsHash = req.GetDepsHash()
	client := pb.NewBuildServiceClient(conn)
	attach := func() error {
		data, err := s.deps.layer(req)
		if err != nil {
			return fmt.Errorf("failed to load dependency layer: %w", err)
		}
		fwd.DepsArchive = data
		return nil
	}

	var buildResp *pb.BuildResponse
	if fwd.DepsHash != "" && !s.deps.workerHolds(worker.ID, fwd.DepsHash) {
		err = attach()
	}
	if err == nil {
		buildResp, err = client.Build(ctx, fwd)
		if isMissingDeps(err) && len(fwd.DepsArchive) == 0 {
			s.deps.markHeld(worker.ID, fwd.DepsHash, false)
			if err = attach(); err == nil {
				buildResp, err = client.Build(ctx, fwd)
			}
		}
	}
	if err == nil && fwd.DepsHash != "" {
		s.deps.markHeld(worker.ID, fwd.DepsHash, true)
	}

	s.finishWorkerBuild(meta, worker.ID, taskStartTime, queueWait, buildResp, err)
	if err != nil {
		log.Error().Err(err).Str("task_id", req.TaskId).Str("worker", worker.ID).Msg("Worker tsc build failed")
		return workerErrorResponse(err), nil
	}
	if buildResp != nil && buildResp.WorkerId == "" {
		buildResp.WorkerId = worker.ID
	}
	return buildResp, nil
}
