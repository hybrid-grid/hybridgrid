package server

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

// rewardNormalizationMs scales the learners' reward −log1p(ms)/log1p(scale).
// It equals the historical 120 s RequestTimeout used in Experiments 1-2 and is
// deliberately independent of the configurable timeout.
const rewardNormalizationMs = 120000.0

// failureReward scores a failed compile as if it had run for the whole request
// timeout (at least 120 s), so no successful compile can score worse than a
// failure. With the 120 s default it is exactly -1, as in Experiments 1-2.
func failureReward(requestTimeout time.Duration) float64 {
	ms := math.Max(rewardNormalizationMs, float64(requestTimeout.Milliseconds()))
	return -math.Log1p(ms) / math.Log1p(rewardNormalizationMs)
}

// Compile handles compilation requests by forwarding to workers.
//
// It runs as a sequence of phases: request validation and tracing setup
// (here), worker selection with booking and forwarding (dispatchAndForward,
// grpc_compile_dispatch.go), then result bookkeeping — registry release,
// scheduler reward, task log, counters, events and metrics
// (grpc_compile_finish.go).
func (s *Server) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	buildID := validation.NormalizeBuildID(req.BuildId)
	start := time.Now()

	// Start tracing span for the coordinator compile flow
	ctx, span := tracing.StartSpan(ctx, "coordinator.Compile",
		tracing.WithCompileAttributes(req.TaskId, req.Compiler, req.TargetArch.String(), len(req.PreprocessedSource)+len(req.RawSource)),
	)
	defer span.End()

	if req.TaskId == "" {
		span.SetStatus(otelcodes.Error, "task_id required")
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// Track cache miss (client checked cache first, this is a miss)
	atomic.AddInt64(&s.cacheMisses, 1)

	atomic.AddInt64(&s.queuedTasks, 1)
	defer atomic.AddInt64(&s.queuedTasks, -1)

	st := newCompileRun(ctx, req, buildID, start)

	atomic.AddInt64(&s.activeTasks, 1)
	defer atomic.AddInt64(&s.activeTasks, -1)
	// Leave the backpressure queue (if the dispatch loop entered it) when
	// Compile returns, after the completion bookkeeping below.
	defer func() {
		if st.queued {
			s.leaveDispatchQueue()
		}
	}()

	if failure := s.dispatchAndForward(ctx, span, st); failure != nil {
		s.countDispatchFailure()
		s.notifyDispatchFailed(st, failure, time.Now())
		return failure, nil
	}
	span.SetAttributes(tracing.AttrQueueTimeMs.Int64(st.queueTime.Milliseconds()))

	atomic.AddInt64(&s.totalTasks, 1)

	// A final rejected attempt already released its slot in the dispatch
	// loop; releasing again would drive the per-worker counter negative.
	if !st.unbooked {
		defer s.releaseWorkerSlot(st.m, st.worker.ID)
	}

	success := s.settleWorker(st)
	s.recordSchedulerReward(st, success)

	taskCompletedTime := time.Now()
	totalDuration := taskCompletedTime.Sub(start)
	span.SetAttributes(tracing.AttrDurationMs.Int64(totalDuration.Milliseconds()))

	s.logTaskCompleted(st, success, totalDuration)
	s.recordTaskResult(span, st, success)
	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskCompleted(completedEvent(st, success, taskCompletedTime))
	}
	return s.finishResponse(ctx, st, success, totalDuration)
}

// forwardCompile forwards the compile request to a worker.
func (s *Server) forwardCompile(ctx context.Context, worker *registry.WorkerInfo, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	// Get pooled connection to worker (reuses existing connections)
	conn, err := s.workerConns.get(ctx, worker.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to worker: %w", err)
	}

	client := pb.NewBuildServiceClient(conn)

	// Forward request with timeout. The client may ask for longer than
	// the configured default (large TU, cross-compile under qemu):
	// the larger of the two wins, otherwise the flag default applies.
	timeout := s.config.RequestTimeout
	if t := time.Duration(req.GetTimeoutSeconds()) * time.Second; t > timeout {
		timeout = t
	}
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return client.Compile(timeoutCtx, req)
}
