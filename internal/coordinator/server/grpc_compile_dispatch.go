package server

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
)

// maxDispatchAttempts bounds the fast select-book-forward budget of one
// Compile() before it falls back to the backpressure queue.
//
// It was 3, tuned for modest bursts. Empirically (follow-up measurements in
// PR hybrid-grid/hybridgrid#9; the budget's measured limits are in
// docs/thesis/bao-cao-tien-do-260821.md §4 and stated in
// docs/thesis/paper-hybridgrid-en.md §3.3), a 5-worker cluster with 10 total
// max_parallel slots hard-fails builds under `make -j7` and above with 3
// attempts: the linear backoff in the dispatch loop (25ms, 50ms) doesn't span
// enough of the dispatch storm at build-start cold start, when make launches
// all -jN local hgcc processes near-simultaneously. Raised to 8, which
// extends the cumulative backoff window to ~700ms (25+50+...+175ms across the
// 7 retries between 8 attempts) — comfortably above the ~10ms median compile
// time this cluster observes, so a slot should free up well before attempts
// run out.
//
// The race this originally compensated for (SelectWith/IncrementTasks reading
// and writing on separate, unsynchronized goroutines) is now eliminated at the
// source: selection is funneled through the single dispatchLoop goroutine (see
// dispatch()/dispatchRequest), so the registry's ActiveTasks is always
// consistent at decision time — no two concurrent Compile() calls can ever
// read the same stale value. A ResourceExhausted now means every eligible
// worker is genuinely at capacity, not a bookkeeping race, so the retry loop's
// remaining job is simply: wait a beat for a real slot to free up (compiles
// are typically ~10-500ms) and try again. Left at 8 with backoff for that
// legitimate case.
const maxDispatchAttempts = 8

// compileRun carries the per-request state of one Compile call between its
// phases: the immutable request context, the outputs of the dispatch loop and
// the backpressure-queue bookkeeping.
type compileRun struct {
	req            *pb.CompileRequest
	buildID        string
	start          time.Time
	m              *metrics.Metrics
	clientOSFilter string
	taskCtx        scheduler.TaskContext

	// Outputs of the last dispatch attempt.
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

	// queued marks entry into the backpressure path; queueDeadline
	// spans the whole queued phase (NOT reset per wake) so total wait
	// stays bounded.
	queued        bool
	queueDeadline time.Time

	// unbooked is set when the last attempt's booking (registry slot and
	// per-worker active counter) was already released by
	// handleWorkerRejected, so the completion bookkeeping must not
	// release it a second time.
	unbooked bool
}

// newCompileRun builds the request-scoped state, records the upload size
// metric and opens the scheduler-selection trace event.
func newCompileRun(ctx context.Context, req *pb.CompileRequest, buildID string, start time.Time) *compileRun {
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

	return &compileRun{
		req:            req,
		buildID:        buildID,
		start:          start,
		m:              m,
		clientOSFilter: clientOSFilter,
		taskCtx:        taskCtx,
	}
}

// loopAction tells the dispatch loop what to do after a failed attempt.
type loopAction int

const (
	// loopRetry continues with the next attempt of the current budget.
	loopRetry loopAction = iota
	// loopRetryFresh restarts the loop with a fresh attempt budget.
	loopRetryFresh
	// loopReturn ends Compile with the accompanying failure response.
	loopReturn
	// loopBreak leaves the loop and processes the rejected forward result.
	loopBreak
)

// queueOutcome is the result of queueForCapacity.
type queueOutcome int

const (
	// queueRetry means the caller should retry with a fresh fast budget.
	queueRetry queueOutcome = iota
	// queueTimedOut means the queue wait gave up (deadline or client cancel).
	queueTimedOut
	// queueDisabled means queueing is off or the queue is full.
	queueDisabled
)

// dispatchAndForward runs the select → book → forward loop, with bounded
// reselection when the worker's admission control rejects the dispatch
// (ResourceExhausted).
//
// Concurrent Compile handlers used to race between SelectWith (which reads
// ActiveTasks) and IncrementTasks (which books it), so a burst — e.g. make
// -jN cold start — could over-book a small worker. The worker rejects rather
// than queueing, and without reselection that scheduling artifact surfaces to
// the client as a compile failure. A rejected attempt is unbooked with
// success=false: the rejection is a genuine overload signal for the worker's
// stats. The learner sees no RecordOutcome for aborted attempts — re-Select
// overwrites the pendingX entry for this TaskID. See maxDispatchAttempts for
// the attempt budget.
//
// On success the outputs are left in st and nil is returned (st.err may still
// carry a non-ResourceExhausted forward error, or an exhausted-and-unqueued
// rejection, for the caller's result handling). A non-nil return is a
// terminal failure response for Compile to return as-is.
func (s *Server) dispatchAndForward(ctx context.Context, span trace.Span, st *compileRun) *pb.CompileResponse {
	for attempt := 1; ; attempt++ {
		selection := s.dispatchWithMetadata(pb.BuildType_BUILD_TYPE_CPP, st.req.TargetArch, st.clientOSFilter, st.taskCtx)
		if selection.err != nil {
			action, failure := s.handleNoWorker(ctx, span, st, attempt, selection.err)
			if action == loopReturn {
				return failure
			}
			if action == loopRetryFresh {
				attempt = 0
			}
			continue
		}
		s.bookAndAnnounce(ctx, span, st, selection)
		s.forwardAttempt(ctx, st)

		if st.err != nil && status.Code(st.err) == codes.ResourceExhausted {
			action, failure := s.handleWorkerRejected(ctx, span, st, attempt)
			if action == loopReturn {
				return failure
			}
			if action == loopBreak {
				return nil
			}
			if action == loopRetryFresh {
				attempt = 0
			}
			continue
		}
		return nil
	}
}

// bookAndAnnounce records a successful selection in st, bumps the per-worker
// active-task gauge, stamps queue/start times and notifies the dashboard that
// the task started. A rejected attempt re-notifies with the next worker;
// events are keyed by task ID so the dashboard just sees the assignment move.
func (s *Server) bookAndAnnounce(ctx context.Context, span trace.Span, st *compileRun, selection dispatchResult) {
	req := st.req
	worker := selection.worker
	st.worker, st.dispatchInfo, st.activeAtDispatch = worker, selection.info, selection.activeAtDispatch
	st.unbooked = false

	tracing.AddEvent(ctx, "scheduler.select.done")
	st.dispatchSeq, st.dispatchTS, st.candidates = selection.seq, selection.dispatchTS, selection.candidates
	span.SetAttributes(tracing.AttrWorkerID.String(worker.ID))

	// activeAtDispatch was captured inside dispatch(), before
	// IncrementTasks booked this decision — see dispatchResult's doc
	// comment for why it can't be re-read from worker.ActiveTasks here.
	val, _ := s.activeTasksByWorker.LoadOrStore(worker.ID, new(int64))
	count := atomic.AddInt64(val.(*int64), 1)
	st.m.SetActiveTaskCount(worker.ID, float64(count))

	st.queueTime = time.Since(st.start)
	st.taskStartTime = time.Now()

	log.Debug().
		Str("task_id", req.TaskId).
		Str("worker_id", worker.ID).
		Dur("queue_time", st.queueTime).
		Msg("Forwarding compile request")

	if s.eventNotifier != nil {
		s.eventNotifier.NotifyTaskStarted(&TaskEvent{
			ID:          req.TaskId,
			BuildType:   "cpp",
			BuildID:     st.buildID,
			Status:      "running",
			WorkerID:    worker.ID,
			StartedAtMs: st.taskStartTime.UnixMilli(),
			SourceFile:  req.GetSourceFilename(),
			Command:     compileCommand(req),
		})
	}
}

// forwardAttempt forwards the request to st.worker and records the result,
// the RPC latency and its metric in st.
func (s *Server) forwardAttempt(ctx context.Context, st *compileRun) {
	tracing.AddEvent(ctx, "forward.start")
	workerCallStart := time.Now()
	st.resp, st.err = s.forwardCompile(ctx, st.worker, st.req)
	st.workerLatency = time.Since(workerCallStart)
	st.m.RecordWorkerLatency(st.worker.ID, float64(st.workerLatency.Milliseconds()))
	tracing.AddEvent(ctx, "forward.done")
}

// releaseWorkerSlot decrements the coordinator-side active-task counter for
// a worker and republishes the gauge.
func (s *Server) releaseWorkerSlot(m *metrics.Metrics, workerID string) {
	if v, ok := s.activeTasksByWorker.Load(workerID); ok {
		c := atomic.AddInt64(v.(*int64), -1)
		m.SetActiveTaskCount(workerID, float64(c))
	}
}

// backoffSleep waits the linear 25ms-per-attempt backoff between attempts. It
// returns false without waiting the full interval when ctx is done, so a
// cancelled or timed-out client does not keep the handler retrying.
func backoffSleep(ctx context.Context, attempt int) bool {
	timer := time.NewTimer(time.Duration(attempt) * 25 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// dispatchCancelledResponse is the failure returned when the client's context
// ends while Compile is still waiting for a worker.
func dispatchCancelledResponse(ctx context.Context) *pb.CompileResponse {
	return &pb.CompileResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("dispatch cancelled while waiting for a worker: %v", ctx.Err()),
	}
}

// queueTimeoutResponse is the failure returned when the dispatch queue wait
// gives up.
func (s *Server) queueTimeoutResponse(err error) *pb.CompileResponse {
	return &pb.CompileResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("dispatch queue timeout (%v) exceeded: no worker capacity became available: %v", s.config.DispatchQueueTimeout, err),
	}
}

// queueForCapacity implements dispatch backpressure once the fast budget is
// exhausted: every eligible worker has been at capacity the whole time
// (~700 ms). Instead of failing the build, queue — wait for a completion or
// registration signal and retry with a fresh budget, bounded by
// DispatchQueueTimeout and the client's own context. A make -jN storm larger
// than cluster capacity thus stretches over time instead of hard-failing the
// over-capacity compiles. A DispatchQueueTimeout of 0 (ablation) or a full
// queue keeps the pre-backpressure reject behavior (queueDisabled).
//
// The first call enters the queue (st.queued) and returns queueRetry
// immediately, without awaiting first: capacity freed during the
// just-exhausted budget is caught by the fresh attempts — going straight to
// await-on-signal could miss a signal that already fired. enterMsg is the log
// line for that entry.
func (s *Server) queueForCapacity(ctx context.Context, st *compileRun, enterMsg string) queueOutcome {
	if s.config.DispatchQueueTimeout > 0 && !st.queued {
		if s.enterDispatchQueue() {
			st.queued = true
			st.queueDeadline = time.Now().Add(s.config.DispatchQueueTimeout)
			log.Warn().
				Str("task_id", st.req.TaskId).
				Dur("timeout", s.config.DispatchQueueTimeout).
				Msg(enterMsg)
			return queueRetry
		}
	}
	if st.queued {
		if !s.awaitDispatchCapacity(ctx, st.queueDeadline) {
			return queueTimedOut
		}
		// A slot may have freed: retry with a fresh fast budget.
		return queueRetry
	}
	return queueDisabled
}

// handleNoWorker handles dispatch() itself finding no eligible worker —
// every worker is genuinely at MaxParallel right now (this is the
// schedulers' own admission check, e.g. ErrNoMatchingWorkers; distinct from
// handleWorkerRejected, which fires only after a worker WAS selected and its
// own admission control rejected the forwarded compile). It used to return
// FAILED immediately with no retry at all, which made properly
// capacity-aware schedulers (P2C, LinUCB, epsilon-greedy, HEFT, and
// LeastLoaded) fail builds outright under a burst that a brief wait would
// have resolved: compiles finish in ~10-500ms, so a worker often frees a slot
// within a couple of backoff cycles. Retried with the same budget and backoff
// as handleWorkerRejected, so a single Compile() call never waits longer in
// total than before — see PR hybrid-grid/hybridgrid#9.
func (s *Server) handleNoWorker(ctx context.Context, span trace.Span, st *compileRun, attempt int, err error) (loopAction, *pb.CompileResponse) {
	req := st.req
	if attempt < maxDispatchAttempts {
		log.Warn().
			Str("task_id", req.TaskId).
			Int("attempt", attempt).
			Msg("No worker currently has capacity; retrying")
		if !backoffSleep(ctx, attempt) {
			span.SetStatus(otelcodes.Error, "dispatch cancelled")
			return loopReturn, dispatchCancelledResponse(ctx)
		}
		return loopRetry, nil
	}
	switch s.queueForCapacity(ctx, st, "All workers at capacity; queued for dispatch") {
	case queueRetry:
		return loopRetryFresh, nil
	case queueTimedOut:
		span.SetStatus(otelcodes.Error, "dispatch queue timeout")
		tracing.RecordError(ctx, err)
		log.Error().Err(err).
			Str("task_id", req.TaskId).
			Str("client_os", req.ClientOs).
			Bool("cross_compile", len(req.RawSource) > 0).
			Msg("Dispatch queue timeout exceeded; no worker capacity became available")
		return loopReturn, s.queueTimeoutResponse(err)
	}
	span.SetStatus(otelcodes.Error, "no worker available")
	tracing.RecordError(ctx, err)
	log.Error().Err(err).
		Str("task_id", req.TaskId).
		Str("client_os", req.ClientOs).
		Bool("cross_compile", len(req.RawSource) > 0).
		Msg("No worker available")
	return loopReturn, &pb.CompileResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   fmt.Sprintf("no worker available: %v", err),
	}
}

// handleWorkerRejected handles a selected worker rejecting the forwarded
// compile with ResourceExhausted (st.err): the booking is undone with
// success=false, then the attempt is retried, queued or — with the queue
// disabled (ablation) — surfaced as the pre-backpressure reject (loopBreak).
func (s *Server) handleWorkerRejected(ctx context.Context, span trace.Span, st *compileRun, attempt int) (loopAction, *pb.CompileResponse) {
	req, worker, err := st.req, st.worker, st.err
	s.registry.DecrementTasks(worker.ID, false, 0)
	s.releaseWorkerSlot(st.m, worker.ID)
	st.unbooked = true
	log.Warn().
		Str("task_id", req.TaskId).
		Str("worker_id", worker.ID).
		Int("attempt", attempt).
		Msg("Worker at capacity; reselecting")
	// Brief backoff lets the racing bookings land in the
	// registry before the next Select reads it.
	if attempt < maxDispatchAttempts {
		if !backoffSleep(ctx, attempt) {
			span.SetStatus(otelcodes.Error, "dispatch cancelled")
			return loopReturn, dispatchCancelledResponse(ctx)
		}
		return loopRetry, nil
	}
	// Fast budget exhausted while every selected worker rejects
	// at its own admission control: queue like the
	// no-eligible-worker path — a completion frees a slot and a
	// fresh budget retries, bounded by DispatchQueueTimeout.
	switch s.queueForCapacity(ctx, st, "Workers kept rejecting at capacity; queued for dispatch") {
	case queueRetry:
		return loopRetryFresh, nil
	case queueTimedOut:
		span.SetStatus(otelcodes.Error, "dispatch queue timeout")
		tracing.RecordError(ctx, err)
		log.Error().Err(err).
			Str("task_id", req.TaskId).
			Msg("Dispatch queue timeout exceeded while workers rejected at capacity")
		return loopReturn, s.queueTimeoutResponse(err)
	}
	// Queue disabled (ablation): pre-backpressure reject. The booking is
	// already released above (st.unbooked), so the caller's completion
	// bookkeeping records the failure without releasing it again.
	return loopBreak, nil
}
