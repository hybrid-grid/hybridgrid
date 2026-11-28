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

	// Determine OS filtering strategy:
	// - Raw source mode: no OS filter needed, workers with Docker can cross-compile
	//   using dockcross images. Workers with matching OS use native compiler.
	// - Preprocessed mode: must match OS since headers are already expanded.
	clientOSFilter := ""
	if len(req.RawSource) == 0 && len(req.PreprocessedSource) > 0 {
		clientOSFilter = req.ClientOs
	}

	// Select worker with tracing
	tracing.AddEvent(ctx, "scheduler.select.start")
	taskCtx := scheduler.TaskContext{
		SourceSizeBytes:    len(req.PreprocessedSource) + len(req.RawSource),
		RawSourceSizeBytes: len(req.RawSource),
		SourceFilename:     req.SourceFilename,
		Compiler:           req.Compiler,
		TaskID:             req.TaskId,
	}
	m := metrics.Default()
	uploadBytes := len(req.PreprocessedSource)
	if len(req.RawSource) > 0 {
		uploadBytes = len(req.RawSource)
	}
	m.RecordTransfer("upload", float64(uploadBytes))

	// Select → book → forward, with bounded reselection when the worker's
	// admission control rejects the dispatch (ResourceExhausted).
	// Concurrent Compile handlers race between SelectWith (which reads
	// ActiveTasks) and IncrementTasks (which books it), so a burst — e.g.
	// make -jN cold start — can over-book a small worker. The worker
	// rejects rather than queueing, and without reselection that
	// scheduling artifact surfaces to the client as a compile failure.
	// A rejected attempt is unbooked with success=false: the rejection
	// is a genuine overload signal for the worker's stats. The learner
	// sees no RecordOutcome for aborted attempts — re-Select overwrites
	// the pendingX entry for this TaskID.
	//
	// maxDispatchAttempts was 3, tuned for modest bursts. Empirically
	// (follow-up measurements in PR hybrid-grid/hybridgrid#9; the budget's
	// measured limits are in docs/thesis/bao-cao-tien-do-260821.md §4 and
	// stated in docs/thesis/paper-hybridgrid-en.md §3.3), a 5-worker
	// cluster with 10 total max_parallel slots
	// hard-fails builds under `make -j7` and above with 3 attempts: the
	// linear backoff below (25ms, 50ms) doesn't span enough of the
	// dispatch storm at build-start cold start, when make launches all
	// -jN local hgcc processes near-simultaneously. Raised to 8, which
	// extends the cumulative backoff window to ~700ms (25+50+...+175ms
	// across the 7 retries between 8 attempts) — comfortably above the
	// ~10ms median compile time this cluster observes, so a slot should
	// free up well before attempts run out.
	// The race this originally compensated for (SelectWith/IncrementTasks
	// reading and writing on separate, unsynchronized goroutines) is now
	// eliminated at the source: selection is funneled through the single
	// dispatchLoop goroutine (see dispatch()/dispatchRequest below), so
	// the registry's ActiveTasks is always consistent at decision time —
	// no two concurrent Compile() calls can ever read the same stale
	// value. A ResourceExhausted here now means every eligible worker is
	// genuinely at capacity, not a bookkeeping race, so this retry loop's
	// remaining job is simply: wait a beat for a real slot to free up
	// (compiles are typically ~10-500ms) and try again. Left at 8 with
	// backoff for that legitimate case.
	const maxDispatchAttempts = 8
	var (
		worker           *registry.WorkerInfo
		dispatchInfo     scheduler.DispatchInfo
		resp             *pb.CompileResponse
		err              error
		workerLatency    time.Duration
		queueTime        time.Duration
		taskStartTime    time.Time
		activeAtDispatch int32
		dispatchSeq      int64
		dispatchTS       time.Time
		candidates       []DispatchCandidate
	)
	atomic.AddInt64(&s.activeTasks, 1)
	defer atomic.AddInt64(&s.activeTasks, -1)
	// queued marks entry into the backpressure path; queueDeadline
	// spans the whole queued phase (NOT reset per wake) so total wait
	// stays bounded.
	var (
		queued        bool
		queueDeadline time.Time
	)
	for attempt := 1; ; attempt++ {
		selection := s.dispatchWithMetadata(pb.BuildType_BUILD_TYPE_CPP, req.TargetArch, clientOSFilter, taskCtx)
		worker, dispatchInfo, activeAtDispatch, err = selection.worker, selection.info, selection.activeAtDispatch, selection.err
		if err != nil {
			// dispatch() itself found no eligible worker — every worker is
			// genuinely at MaxParallel right now (this is the schedulers'
			// own admission check, e.g. ErrNoMatchingWorkers; distinct from
			// the ResourceExhausted branch below, which fires only after a
			// worker WAS selected and its own admission control rejected
			// the forwarded compile). Previously this returned FAILED
			// immediately with no retry at all, which made properly
			// capacity-aware schedulers (P2C, LinUCB, epsilon-greedy, HEFT,
			// and — since the fix above — LeastLoaded) fail builds outright
			// under a burst that a brief wait would have resolved: compiles
			// finish in ~10-500ms, so a worker often frees a slot within a
			// couple of backoff cycles. Retried with the same budget and
			// backoff as the ResourceExhausted branch below, so a single
			// Compile() call never waits longer in total than before this
			// change — see PR hybrid-grid/hybridgrid#9.
			if attempt < maxDispatchAttempts {
				log.Warn().
					Str("task_id", req.TaskId).
					Int("attempt", attempt).
					Msg("No worker currently has capacity; retrying")
				time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
				continue
			}
			// Fast budget exhausted (~700 ms): every eligible worker
			// has been at capacity the whole time. Instead of failing
			// the build, queue — wait for a completion or registration
			// signal and retry with a fresh budget, bounded by
			// DispatchQueueTimeout and the client's own context. This
			// is dispatch backpressure: a make -jN storm larger than
			// cluster capacity now stretches over time instead of
			// hard-failing the over-capacity compiles. A
			// DispatchQueueTimeout of 0 (ablation) or a full queue
			// keeps the pre-backpressure reject behavior.
			if s.config.DispatchQueueTimeout > 0 && !queued {
				if s.enterDispatchQueue() {
					queued = true
					defer s.leaveDispatchQueue()
					queueDeadline = time.Now().Add(s.config.DispatchQueueTimeout)
					log.Warn().
						Str("task_id", req.TaskId).
						Dur("timeout", s.config.DispatchQueueTimeout).
						Msg("All workers at capacity; queued for dispatch")
					// A fresh budget immediately (no await first):
					// capacity freed during the just-exhausted budget
					// is caught by these attempts — going straight to
					// await-on-signal could miss a signal that already
					// fired.
					attempt = 0
					continue
				}
			}
			if queued {
				if !s.awaitDispatchCapacity(ctx, queueDeadline) {
					span.SetStatus(otelcodes.Error, "dispatch queue timeout")
					tracing.RecordError(ctx, err)
					log.Error().Err(err).
						Str("task_id", req.TaskId).
						Str("client_os", req.ClientOs).
						Bool("cross_compile", len(req.RawSource) > 0).
						Msg("Dispatch queue timeout exceeded; no worker capacity became available")
					return &pb.CompileResponse{
						Status:   pb.TaskStatus_STATUS_FAILED,
						ExitCode: 1,
						Stderr:   fmt.Sprintf("dispatch queue timeout (%v) exceeded: no worker capacity became available: %v", s.config.DispatchQueueTimeout, err),
					}, nil
				}
				// A slot may have freed: retry with a fresh fast budget.
				attempt = 0
				continue
			}
			span.SetStatus(otelcodes.Error, "no worker available")
			tracing.RecordError(ctx, err)
			log.Error().Err(err).
				Str("task_id", req.TaskId).
				Str("client_os", req.ClientOs).
				Bool("cross_compile", len(req.RawSource) > 0).
				Msg("No worker available")
			return &pb.CompileResponse{
				Status:   pb.TaskStatus_STATUS_FAILED,
				ExitCode: 1,
				Stderr:   fmt.Sprintf("no worker available: %v", err),
			}, nil
		}
		tracing.AddEvent(ctx, "scheduler.select.done")
		dispatchSeq, dispatchTS, candidates = selection.seq, selection.dispatchTS, selection.candidates
		span.SetAttributes(tracing.AttrWorkerID.String(worker.ID))

		// activeAtDispatch was captured inside dispatch(), before
		// IncrementTasks booked this decision — see dispatchResult's doc
		// comment for why it can't be re-read from worker.ActiveTasks here.
		val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
		count := atomic.AddInt64(val.(*int64), 1)
		m.SetActiveTaskCount(worker.ID, float64(count))

		queueTime = time.Since(start)
		taskStartTime = time.Now()

		log.Debug().
			Str("task_id", req.TaskId).
			Str("worker_id", worker.ID).
			Dur("queue_time", queueTime).
			Msg("Forwarding compile request")

		// Notify task started. A rejected attempt re-notifies with the
		// next worker; events are keyed by task ID so the dashboard just
		// sees the assignment move.
		if s.eventNotifier != nil {
			s.eventNotifier.NotifyTaskStarted(&TaskEvent{
				ID:          req.TaskId,
				BuildType:   "cpp",
				BuildID:     buildID,
				Status:      "running",
				WorkerID:    worker.ID,
				StartedAtMs: taskStartTime.UnixMilli(),
			})
		}

		tracing.AddEvent(ctx, "forward.start")
		workerCallStart := time.Now()
		resp, err = s.forwardCompile(ctx, worker, req)
		workerLatency = time.Since(workerCallStart)
		m.RecordWorkerLatency(worker.ID, float64(workerLatency.Milliseconds()))
		tracing.AddEvent(ctx, "forward.done")

		if err != nil && status.Code(err) == codes.ResourceExhausted {
			s.registry.DecrementTasks(worker.ID, false, 0)
			if v, ok := s.activeTasksByWorker.Load(worker.ID); ok {
				c := atomic.AddInt64(v.(*int64), -1)
				m.SetActiveTaskCount(worker.ID, float64(c))
			}
			log.Warn().
				Str("task_id", req.TaskId).
				Str("worker_id", worker.ID).
				Int("attempt", attempt).
				Msg("Worker at capacity; reselecting")
			// Brief backoff lets the racing bookings land in the
			// registry before the next Select reads it.
			if attempt < maxDispatchAttempts {
				time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
				continue
			}
			// Fast budget exhausted while every selected worker rejects
			// at its own admission control: queue like the
			// no-eligible-worker path — a completion frees a slot and a
			// fresh budget retries, bounded by DispatchQueueTimeout.
			if s.config.DispatchQueueTimeout > 0 && !queued {
				if s.enterDispatchQueue() {
					queued = true
					defer s.leaveDispatchQueue()
					queueDeadline = time.Now().Add(s.config.DispatchQueueTimeout)
					log.Warn().
						Str("task_id", req.TaskId).
						Dur("timeout", s.config.DispatchQueueTimeout).
						Msg("Workers kept rejecting at capacity; queued for dispatch")
					attempt = 0
					continue
				}
			}
			if queued {
				if !s.awaitDispatchCapacity(ctx, queueDeadline) {
					span.SetStatus(otelcodes.Error, "dispatch queue timeout")
					tracing.RecordError(ctx, err)
					log.Error().Err(err).
						Str("task_id", req.TaskId).
						Msg("Dispatch queue timeout exceeded while workers rejected at capacity")
					return &pb.CompileResponse{
						Status:   pb.TaskStatus_STATUS_FAILED,
						ExitCode: 1,
						Stderr:   fmt.Sprintf("dispatch queue timeout (%v) exceeded: no worker capacity became available: %v", s.config.DispatchQueueTimeout, err),
					}, nil
				}
				// A slot may have freed: retry with a fresh fast budget.
				attempt = 0
				continue
			}
			// Queue disabled (ablation): pre-backpressure reject.
			break
		}
		break
	}
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
