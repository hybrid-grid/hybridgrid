package server

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
)

// compileSuccess reports whether a forwarded compile completed successfully.
func compileSuccess(resp *pb.CompileResponse, err error) bool {
	return err == nil && resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED
}

// settleWorker closes the registry booking for the finished compile: it
// records the artifact download size, derives the success flag, releases the
// worker slot with the compile time and wakes queued Compile() waiters. It
// returns the success flag.
func (s *Server) settleWorker(st *compileRun) bool {
	resp := st.resp
	if resp != nil && len(resp.ObjectFile) > 0 {
		st.m.RecordTransfer("download", float64(len(resp.ObjectFile)))
	}

	// Track completion
	success := compileSuccess(resp, st.err)
	compileTime := time.Duration(0)
	if resp != nil {
		compileTime = time.Duration(resp.CompilationTimeMs) * time.Millisecond
	}
	// A rejected final attempt was already unbooked (with success=false)
	// by handleWorkerRejected; decrementing again would free a slot that
	// belongs to another task and double-count the failure.
	if !st.unbooked {
		s.registry.DecrementTasks(st.worker.ID, success, compileTime)
	}
	// A cpp slot just freed — wake queued Compile() waiters.
	s.signalDispatchCapacity()
	return success
}

// compileReward is the learner reward for one finished compile. Reward
// convention: higher is better. We use a normalised negative log-latency so
// the reward magnitude does not dwarf LinUCB's UCB exploration bonus during
// warm-up (code-review finding HIGH-2). Normalisation divisor is
// log1p(rewardNormalizationMs) so the reward lies in roughly [-1, 0]. The log
// transform compresses the heavy tail (M1 P99/P50 ≈ 29×); see
// docs/thesis/theory-notes.md §4.3 for the empirical-choice rationale (this
// is NOT the Decima reward function).
//
// The reward scale is fixed at the 2-minute default instead of following
// RequestTimeout: raising the timeout (e.g. for long C++ compiles) must not
// silently change what the learner sees.
func compileReward(success bool, resp *pb.CompileResponse, requestTimeout time.Duration) float64 {
	denom := math.Log1p(rewardNormalizationMs)
	if success && resp != nil {
		return -math.Log1p(float64(resp.CompilationTimeMs)) / denom
	}
	return failureReward(requestTimeout)
}

// recordSchedulerReward feeds the outcome back to online-learning
// schedulers; it is a no-op for schedulers that do not learn.
func (s *Server) recordSchedulerReward(st *compileRun, success bool) {
	if learner, ok := s.scheduler.(scheduler.LearningScheduler); ok {
		reward := compileReward(success, st.resp, s.config.RequestTimeout)
		learner.RecordOutcome(st.worker.ID, reward, success, st.taskCtx)
	}
}

// logTaskCompleted writes the per-task structured log record for offline
// analysis / RL training. It is a no-op when no task logger is configured.
func (s *Server) logTaskCompleted(st *compileRun, success bool, totalDuration time.Duration) {
	if s.taskLogger == nil {
		return
	}
	req, worker, resp := st.req, st.worker, st.resp
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
		DispatchSeq:                 st.dispatchSeq,
		DispatchTS:                  st.dispatchTS,
		ReceivedTS:                  st.start.UTC(),
		WorkerAddress:               worker.Address,
		Candidates:                  st.candidates,
		LinUCBCandidates:            st.dispatchInfo.Candidates,
		RewardOffset:                st.dispatchInfo.RewardOffset,
		TaskID:                      req.TaskId,
		BuildType:                   "cpp",
		BuildID:                     st.buildID,
		Scheduler:                   s.config.SchedulerType,
		SchedulerParams:             s.schedulerParams,
		WorkerID:                    worker.ID,
		WorkerArch:                  workerNativeArch,
		WorkerNativeArch:            workerNativeArch,
		WorkerCPUCores:              workerCPUCores,
		WorkerCPUMillis:             workerCPUMillis,
		WorkerMemBytes:              workerMemBytes,
		WorkerActiveTasksAtDispatch: st.activeAtDispatch,
		WorkerMaxParallel:           worker.MaxParallel,
		WorkerDiscoverySource:       worker.DiscoverySource,
		TargetArch:                  req.TargetArch.String(),
		ClientOS:                    req.ClientOs,
		SourceSizeBytes:             len(req.PreprocessedSource) + len(req.RawSource),
		PreprocessedSizeBytes:       len(req.PreprocessedSource),
		RawSourceSizeBytes:          len(req.RawSource),
		QueueTimeMs:                 st.queueTime.Milliseconds(),
		CompileTimeMs:               compileTimeMs,
		WorkerRPCLatencyMs:          st.workerLatency.Milliseconds(),
		TotalDurationMs:             totalDuration.Milliseconds(),
		Success:                     success,
		ExitCode:                    exitCode,
		FromCache:                   false,
		QValueAtDispatch:            st.dispatchInfo.QValueAtDispatch,
		WasExploration:              st.dispatchInfo.WasExploration,
	})
}

// recordTaskResult updates the success/failure counters and span status and
// retains the console output for the dashboard.
func (s *Server) recordTaskResult(span trace.Span, st *compileRun, success bool) {
	if success {
		atomic.AddInt64(&s.successTasks, 1)
		span.SetStatus(otelcodes.Ok, "compilation succeeded")
	} else {
		atomic.AddInt64(&s.failedTasks, 1)
		span.SetStatus(otelcodes.Error, "compilation failed")
	}

	// Retain console output for the dashboard (bounded); the response
	// carries the full log because Compile is unary.
	if st.resp != nil {
		s.console.Put(st.req.TaskId, st.resp.Stdout, st.resp.Stderr)
	}
}

// countDispatchFailure counts a request that ended in dispatchAndForward's
// terminal failure (no worker, queue timeout, client cancellation) as one
// failed task, like the build handlers do for their no-worker failures.
// Those paths return before recordTaskResult, so without this the stats
// under-report failed compiles. It must be called exactly once per such
// failure and never for a request that reaches recordTaskResult, which
// counts its own outcome (including the queue-disabled final rejection).
func (s *Server) countDispatchFailure() {
	atomic.AddInt64(&s.totalTasks, 1)
	atomic.AddInt64(&s.failedTasks, 1)
}

// completedEvent builds the dashboard completion event for a finished
// compile.
func completedEvent(st *compileRun, success bool, completedAt time.Time) *TaskEvent {
	resp, err := st.resp, st.err
	event := &TaskEvent{
		ID:            st.req.TaskId,
		BuildType:     "cpp",
		BuildID:       st.buildID,
		WorkerID:      st.worker.ID,
		StartedAtMs:   st.taskStartTime.UnixMilli(),
		CompletedAtMs: completedAt.UnixMilli(),
		DurationMs:    completedAt.Sub(st.taskStartTime).Milliseconds(),
		// Coordinator-side queue time (Compile entry → dispatch),
		// the same span the JSONL task log records: with dispatch
		// backpressure the real wait happens here, before the
		// worker is involved — the worker's own QueueTimeMs is
		// always 0 (its admission control rejects rather than
		// queues).
		QueueTimeMs:   st.queueTime.Milliseconds(),
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
	return event
}

// notifyDispatchFailed closes the dashboard task opened by bookAndAnnounce
// when the dispatch loop ends the request with a terminal failure (queue
// timeout, client cancellation, no worker left after a rejected attempt).
// Those paths return before the normal completion bookkeeping, so without
// this the started event would leave the task "running" forever. A request
// that never reached a worker (st.worker == nil) announced nothing and gets
// no completion.
func (s *Server) notifyDispatchFailed(st *compileRun, failure *pb.CompileResponse, completedAt time.Time) {
	if s.eventNotifier == nil || st.worker == nil {
		return
	}
	s.eventNotifier.NotifyTaskCompleted(&TaskEvent{
		ID:            st.req.TaskId,
		BuildType:     "cpp",
		BuildID:       st.buildID,
		Status:        "failed",
		WorkerID:      st.worker.ID,
		StartedAtMs:   st.taskStartTime.UnixMilli(),
		CompletedAtMs: completedAt.UnixMilli(),
		DurationMs:    completedAt.Sub(st.taskStartTime).Milliseconds(),
		QueueTimeMs:   st.queueTime.Milliseconds(),
		ExitCode:      1,
		ErrorMessage:  failure.Stderr,
	})
}

// finishResponse builds Compile's return value from the forwarded result: a
// worker error becomes a FAILED response, otherwise the worker response is
// returned with the coordinator queue time and the completion metrics
// recorded.
func (s *Server) finishResponse(ctx context.Context, st *compileRun, success bool, totalDuration time.Duration) (*pb.CompileResponse, error) {
	if st.err != nil {
		tracing.RecordError(ctx, st.err)
		log.Error().Err(st.err).Str("task_id", st.req.TaskId).Msg("Worker compilation failed")
		return &pb.CompileResponse{
			Status:   pb.TaskStatus_STATUS_FAILED,
			ExitCode: 1,
			Stderr:   fmt.Sprintf("worker error: %v", st.err),
		}, nil
	}

	// Set queue time
	st.resp.QueueTimeMs = int64(st.queueTime.Milliseconds())

	duration := totalDuration.Seconds()
	buildType := "cpp"
	if success {
		st.m.RecordTaskComplete(metrics.TaskStatusSuccess, buildType, st.worker.ID, duration)
	} else {
		st.m.RecordTaskComplete(metrics.TaskStatusError, buildType, st.worker.ID, duration)
	}
	st.m.RecordQueueTime(buildType, st.queueTime.Seconds())

	return st.resp, nil
}
