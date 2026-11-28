package server

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	otelcodes "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
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
	m := st.m
	taskCtx := st.taskCtx

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
		return failure, nil
	}
	worker, dispatchInfo, resp, err := st.worker, st.dispatchInfo, st.resp, st.err
	workerLatency, queueTime, taskStartTime := st.workerLatency, st.queueTime, st.taskStartTime
	activeAtDispatch, dispatchSeq, dispatchTS, candidates := st.activeAtDispatch, st.dispatchSeq, st.dispatchTS, st.candidates
	span.SetAttributes(tracing.AttrQueueTimeMs.Int64(queueTime.Milliseconds()))

	atomic.AddInt64(&s.totalTasks, 1)

	defer func() {
		if val, ok := s.activeTasksByWorker.Load(worker.ID); ok {
			count := atomic.AddInt64(val.(*int64), -1)
			m.SetActiveTaskCount(worker.ID, float64(count))
		}
	}()

	if resp != nil && len(resp.ObjectFile) > 0 {
		m.RecordTransfer("download", float64(len(resp.ObjectFile)))
	}

	// Track completion
	success := err == nil && resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED
	compileTime := time.Duration(0)
	if resp != nil {
		compileTime = time.Duration(resp.CompilationTimeMs) * time.Millisecond
	}
	s.registry.DecrementTasks(worker.ID, success, compileTime)
	// A cpp slot just freed — wake queued Compile() waiters.
	s.signalDispatchCapacity()

	// Feedback loop for online-learning schedulers. Reward convention:
	// higher is better. We use a normalised negative log-latency so the
	// reward magnitude does not dwarf LinUCB's UCB exploration bonus
	// during warm-up (code-review finding HIGH-2). Normalisation divisor
	// is log1p(RequestTimeoutMs) so the reward lies in roughly [-1, 0].
	// The log transform compresses the heavy tail (M1 P99/P50 ≈ 29×);
	// see docs/thesis/theory-notes.md §4.3 for the empirical-choice
	// rationale (this is NOT the Decima reward function).
	if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok {
		// The reward scale is fixed at the 2-minute default instead of
		// following RequestTimeout: raising the timeout (e.g. for long
		// C++ compiles) must not silently change what the learner sees.
		denom := math.Log1p(rewardNormalizationMs)
		var reward float64
		if success && resp != nil {
			reward = -math.Log1p(float64(resp.CompilationTimeMs)) / denom
		} else {
			reward = failureReward(s.config.RequestTimeout)
		}
		learner.RecordOutcome(worker.ID, reward, success, taskCtx)
	}

	taskCompletedTime := time.Now()
	totalDuration := taskCompletedTime.Sub(start)
	span.SetAttributes(tracing.AttrDurationMs.Int64(totalDuration.Milliseconds()))

	// Per-task structured log for offline analysis / RL training.
	if s.taskLogger != nil {
		var (
			workerCPUCores   int32
			workerCPUMillis  int32
			workerMemBytes   int64
			workerNativeArch string
		)
		if worker.Capabilities != nil {
			workerCPUCores = worker.Capabilities.CpuCores
			workerCPUMillis = worker.Capabilities.CpuMillis
			workerMemBytes = worker.Capabilities.MemoryBytes
			workerNativeArch = worker.Capabilities.NativeArch.String()
		}
		var (
			compileTimeMs int64
			exitCode      int32
		)
		if resp != nil {
			compileTimeMs = resp.CompilationTimeMs
			exitCode = resp.ExitCode
		}
		s.taskLogger.Log(&TaskLogRecord{
			TS:                          time.Now().UTC(),
			Event:                       "task_completed",
			DispatchSeq:                 dispatchSeq,
			DispatchTS:                  dispatchTS,
			ReceivedTS:                  start.UTC(),
			WorkerAddress:               worker.Address,
			Candidates:                  candidates,
			LinUCBCandidates:            dispatchInfo.Candidates,
			RewardOffset:                dispatchInfo.RewardOffset,
			TaskID:                      req.TaskId,
			BuildType:                   "cpp",
			BuildID:                     buildID,
			Scheduler:                   s.config.SchedulerType,
			SchedulerParams:             s.schedulerParams,
			WorkerID:                    worker.ID,
			WorkerArch:                  workerNativeArch,
			WorkerNativeArch:            workerNativeArch,
			WorkerCPUCores:              workerCPUCores,
			WorkerCPUMillis:             workerCPUMillis,
			WorkerMemBytes:              workerMemBytes,
			WorkerActiveTasksAtDispatch: activeAtDispatch,
			WorkerMaxParallel:           worker.MaxParallel,
			WorkerDiscoverySource:       worker.DiscoverySource,
			TargetArch:                  req.TargetArch.String(),
			ClientOS:                    req.ClientOs,
			SourceSizeBytes:             len(req.PreprocessedSource) + len(req.RawSource),
			PreprocessedSizeBytes:       len(req.PreprocessedSource),
			RawSourceSizeBytes:          len(req.RawSource),
			QueueTimeMs:                 queueTime.Milliseconds(),
			CompileTimeMs:               compileTimeMs,
			WorkerRPCLatencyMs:          workerLatency.Milliseconds(),
			TotalDurationMs:             totalDuration.Milliseconds(),
			Success:                     success,
			ExitCode:                    exitCode,
			FromCache:                   false,
			QValueAtDispatch:            dispatchInfo.QValueAtDispatch,
			WasExploration:              dispatchInfo.WasExploration,
		})
	}

	if success {
		atomic.AddInt64(&s.successTasks, 1)
		span.SetStatus(otelcodes.Ok, "compilation succeeded")
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
		span.SetStatus(otelcodes.Error, "compilation failed")
	}

	// Retain console output for the dashboard (bounded); the response
	// carries the full log because Compile is unary.
	if resp != nil {
		s.console.Put(req.TaskId, resp.Stdout, resp.Stderr)
	}
	// Notify task completed
	if s.eventNotifier != nil {
		event := &TaskEvent{
			ID:            req.TaskId,
			BuildType:     "cpp",
			BuildID:       buildID,
			WorkerID:      worker.ID,
			StartedAtMs:   taskStartTime.UnixMilli(),
			CompletedAtMs: taskCompletedTime.UnixMilli(),
			DurationMs:    taskCompletedTime.Sub(taskStartTime).Milliseconds(),
			// Coordinator-side queue time (Compile entry → dispatch),
			// the same span the JSONL task log records: with dispatch
			// backpressure the real wait happens here, before the
			// worker is involved — the worker's own QueueTimeMs is
			// always 0 (its admission control rejects rather than
			// queues).
			QueueTimeMs:   queueTime.Milliseconds(),
			CompileTimeMs: resp.GetCompilationTimeMs(),
		}
		if success {
			event.Status = "completed"
			if resp != nil {
				event.ExitCode = resp.ExitCode
			}
		} else {
			event.Status = "failed"
			event.ExitCode = 1
			if err != nil {
				event.ErrorMessage = err.Error()
			} else if resp != nil {
				event.ExitCode = resp.ExitCode
				event.ErrorMessage = resp.Stderr
			}
		}
		s.eventNotifier.NotifyTaskCompleted(event)
	}

	if err != nil {
		tracing.RecordError(ctx, err)
		log.Error().Err(err).Str("task_id", req.TaskId).Msg("Worker compilation failed")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", err),
		}, nil
	}

	// Set queue time
	resp.QueueTimeMs = int64(queueTime.Milliseconds())

	duration := totalDuration.Seconds()
	buildType := "cpp"
	if success {
		m.RecordTaskComplete(metrics.TaskStatusSuccess, buildType, worker.ID, duration)
	} else {
		m.RecordTaskComplete(metrics.TaskStatusError, buildType, worker.ID, duration)
	}
	m.RecordQueueTime(buildType, queueTime.Seconds())

	return resp, nil
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
