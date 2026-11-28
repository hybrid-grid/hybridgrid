package server

import (
	"context"
	"fmt"
	"sort"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
)

// dispatchQueueSize bounds how many pending selection requests may queue
// up behind the single dispatchLoop goroutine before Compile() callers
// block trying to submit one. Generously sized relative to any burst
// this system exercises (make -jN cold starts up to a few dozen) so a
// legitimate burst never blocks on channel capacity itself — only on
// dispatchLoop's own (sub-microsecond) processing rate.
const dispatchQueueSize = 256

// maxQueuedCompiles bounds how many Compile() calls may sit in the
// dispatch queue (the backpressure path entered when the fast dispatch
// budget is exhausted) before the coordinator reverts to rejecting.
// Mirrors dispatchQueueSize: large enough for a make -j storm, small
// enough that a wedged cluster fails fast instead of accumulating
// unbounded waiters.
const maxQueuedCompiles = 256

// dispatchRequest is one pending "pick a worker and book it" decision.
// Compile() submits these to s.dispatchCh instead of calling
// scheduler.SelectWith and s.registry.IncrementTasks directly, so every
// selection decision across every concurrent RPC handler is funneled
// through the single dispatchLoop goroutine below.
//
// Why: SelectWith (reads a worker's ActiveTasks) and IncrementTasks
// (books the decision by writing it back) are two separate registry
// operations. When Compile() called them inline, concurrent handler
// goroutines could each read the same pre-booking ActiveTasks value
// before any of them had written their own increment — a classic
// check-then-act race. Under `make -jN` for N large enough relative to
// total worker capacity (empirically N ≳ 7 on a 5-worker/10-slot
// cluster — see PR hybrid-grid/hybridgrid#9), this let multiple
// in-flight decisions overbook the same
// (especially low-max_parallel) worker, which then rejected the excess
// with ResourceExhausted; a bounded reselection-with-backoff retry
// (maxDispatchAttempts) mitigated but did not eliminate the failures.
//
// Routing every decision through one goroutine removes the race
// entirely: only dispatchLoop ever reads or books a worker's
// ActiveTasks for a Compile() dispatch, so no two decisions can ever
// observe the same stale value. This does not add meaningful latency —
// a selection decision is a few in-memory comparisons (sub-microsecond)
// against ~10-500ms compile times — and does not serialize the actual
// compiles, which still run concurrently across all workers exactly as
// before; only the brief "who gets this file" moment is serialized.
//
// The Flutter/Unity build paths also route through this queue via
// filter-mode requests (see filter below): they share the same
// backpressure discipline and the same atomic select-then-book
// guarantee as Compile().
type dispatchRequest struct {
	buildType pb.BuildType
	arch      pb.Architecture
	clientOS  string
	ctx       scheduler.TaskContext
	// filter/label: when filter is non-nil this is a filter-mode
	// request (flutter, unity, and custom-builder paths): dispatchLoop
	// picks the least-loaded healthy worker with spare capacity that
	// matches the predicate instead of consulting the learning
	// scheduler.
	filter func(*registry.WorkerInfo) bool
	label  string
	result chan dispatchResult
}

// dispatchResult is dispatchLoop's answer to a dispatchRequest.
// activeAtDispatch mirrors the semantics the inline code used to
// provide: the worker's ActiveTasks read immediately before
// IncrementTasks booked this decision, i.e. load at decision time, not
// after. It must travel back through this struct rather than have the
// caller re-read worker.ActiveTasks after the fact, because by the time
// dispatch() returns, IncrementTasks has already run — a fresh read
// would see the post-booking count, not the pre-booking one the offline
// analysis field (worker_active_tasks_at_dispatch) is documented to mean.
type dispatchResult struct {
	worker           *registry.WorkerInfo
	info             scheduler.DispatchInfo
	activeAtDispatch int32
	seq              int64
	dispatchTS       time.Time
	candidates       []DispatchCandidate
	err              error
}

// dispatchLoop is the single goroutine that owns worker selection and
// booking for the Compile() path. It runs for the coordinator's
// lifetime, started once in New() and stopped by closing s.dispatchCh
// in Stop(). See dispatchRequest's doc comment for why this exists.
func (s *Server) dispatchLoop() {
	for req := range s.dispatchCh {
		var candidates []DispatchCandidate
		var (
			worker *registry.WorkerInfo
			info   scheduler.DispatchInfo
			err    error
		)
		if req.filter != nil {
			worker, err = s.selectByFilter(req.filter, req.label)
		} else {
			candidates = s.snapshotCandidates(req.buildType, req.arch, req.clientOS)
			worker, info, err = scheduler.SelectWith(s.scheduler, req.buildType, req.arch, req.clientOS, req.ctx)
		}
		result := dispatchResult{worker: worker, info: info, candidates: candidates, err: err}
		if err == nil {
			// Read-then-book, both on this single goroutine: no other
			// goroutine can interleave a selection between this read
			// and the IncrementTasks call below, because dispatchLoop
			// is the only caller of either.
			result.activeAtDispatch = worker.ActiveTasks
			if err = s.registry.IncrementTasks(worker.ID); err != nil {
				result.err = err
			} else {
				result.seq = atomic.AddInt64(&s.dispatchCount, 1)
				result.dispatchTS = time.Now().UTC()
				metrics.Default().DispatchDecisions.Inc()
			}
		}
		req.result <- result
	}
}

// snapshotCandidates includes unhealthy capability matches so their health is
// visible, and applies the scheduler's same-OS-or-Docker predicate.
func (s *Server) snapshotCandidates(buildType pb.BuildType, arch pb.Architecture, clientOS string) []DispatchCandidate {
	workers := s.registry.List()
	result := make([]DispatchCandidate, 0, len(workers))
	for _, worker := range workers {
		if !matchesDispatchCapability(worker.Capabilities, buildType, arch) {
			continue
		}
		if clientOS != "" && worker.Capabilities.Os != clientOS && !worker.Capabilities.DockerAvailable {
			continue
		}
		result = append(result, DispatchCandidate{
			Address: worker.Address, Active: worker.ActiveTasks,
			MaxParallel: worker.MaxParallel, CPUMillis: worker.Capabilities.CpuMillis,
			Healthy: worker.IsHealthy(s.config.HeartbeatTTL),
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Address < result[j].Address })
	return result
}

// matchesDispatchCapability mirrors registry.matchesCapability for the
// snapshot. The registry's ListByCapability omits unhealthy workers.
func matchesDispatchCapability(caps *pb.WorkerCapabilities, buildType pb.BuildType, arch pb.Architecture) bool {
	if caps == nil {
		return false
	}
	if arch != pb.Architecture_ARCH_UNSPECIFIED && caps.NativeArch != arch &&
		!caps.DockerAvailable && (caps.Cpp == nil || !caps.Cpp.CrossCompile) {
		return false
	}
	switch buildType {
	case pb.BuildType_BUILD_TYPE_CPP:
		return caps.Cpp != nil && len(caps.Cpp.Compilers) > 0
	case pb.BuildType_BUILD_TYPE_GO:
		return caps.Go != nil
	case pb.BuildType_BUILD_TYPE_RUST:
		return caps.Rust != nil
	case pb.BuildType_BUILD_TYPE_NODEJS:
		return caps.Nodejs != nil
	case pb.BuildType_BUILD_TYPE_FLUTTER:
		return caps.Flutter != nil
	default:
		return true
	}
}

// selectByFilter picks the matching worker with the fewest active
// tasks. Spare capacity is part of the match, so a full worker never
// wins a dispatch and the backpressure queue engages instead. Must
// only be called from dispatchLoop, where selection and the
// IncrementTasks booking are atomic with respect to every other
// dispatch.
func (s *Server) selectByFilter(filter func(*registry.WorkerInfo) bool, label string) (*registry.WorkerInfo, error) {
	var best *registry.WorkerInfo
	for _, w := range s.registry.List() {
		if !w.IsHealthy(s.config.HeartbeatTTL) || !filter(w) {
			continue
		}
		if w.MaxParallel > 0 && w.ActiveTasks >= w.MaxParallel {
			continue
		}
		if best == nil || w.ActiveTasks < best.ActiveTasks {
			best = w
		}
	}
	if best == nil {
		return nil, fmt.Errorf("no worker available for %s build", label)
	}
	return best, nil
}

// dispatch submits a selection request to dispatchLoop and blocks for
// its result. Safe to call from any number of concurrent goroutines —
// that concurrency is exactly what dispatchLoop serializes away.
//
// Registered on s.dispatchWG for the whole send-then-wait-for-result
// span (not just the send): Stop() waits on this WaitGroup before
// closing s.dispatchCh, so a close can never race a concurrent send —
// see Stop()'s comment for why this can't just rely on the gRPC
// server's GracefulStop() to provide that guarantee.
func (s *Server) dispatch(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx scheduler.TaskContext) (*registry.WorkerInfo, scheduler.DispatchInfo, int32, error) {
	res := s.dispatchWithMetadata(buildType, arch, clientOS, ctx)
	return res.worker, res.info, res.activeAtDispatch, res.err
}

func (s *Server) dispatchWithMetadata(buildType pb.BuildType, arch pb.Architecture, clientOS string, ctx scheduler.TaskContext) dispatchResult {
	s.dispatchWG.Add(1)
	defer s.dispatchWG.Done()

	req := dispatchRequest{
		buildType: buildType,
		arch:      arch,
		clientOS:  clientOS,
		ctx:       ctx,
		result:    make(chan dispatchResult, 1),
	}
	s.dispatchCh <- &req
	return <-req.result
}

// Dispatches returns the number of booked C/C++ dispatch attempts.
func (s *Server) Dispatches() int64 { return atomic.LoadInt64(&s.dispatchCount) }

// LogInjectedEvent writes an experiment event to the coordinator task log.
func (s *Server) LogInjectedEvent(kind, target string, dispatchCount int64, detail string) error {
	return s.taskLogger.LogInjectedEvent(kind, target, dispatchCount, detail)
}

// signalDispatchCapacity wakes every Compile() call waiting in the
// dispatch queue: a task completed (slot freed) or a worker registered
// (new capacity). Broadcast by close-and-replace, so a signal that
// arrives while no waiter is listening is not lost — the next waiter
// sees the already-closed channel immediately and retries once.
func (s *Server) signalDispatchCapacity() {
	s.dispatchSignalMu.Lock()
	if s.dispatchSignal != nil {
		close(s.dispatchSignal)
		s.dispatchSignal = make(chan struct{})
	}
	s.dispatchSignalMu.Unlock()
}

// dispatchCapacitySignal returns the current broadcast channel. Grab it
// once per wait; a channel replaced after the grab still fires on the
// next signal.
func (s *Server) dispatchCapacitySignal() chan struct{} {
	s.dispatchSignalMu.Lock()
	defer s.dispatchSignalMu.Unlock()
	return s.dispatchSignal
}

// awaitDispatchCapacity blocks until a capacity signal fires, deadline
// passes, or the client context is cancelled. Returns true when the
// caller should retry worker selection — races between woken waiters
// mean the retry can still find every slot taken, in which case the
// caller re-enters with the same deadline. Callers must have entered
// the queue via enterDispatchQueue (the waiter cap is checked there,
// once per queued Compile, not per wake).
func (s *Server) awaitDispatchCapacity(ctx context.Context, deadline time.Time) bool {
	// The re-check tick bounds signal loss: a capacity signal that
	// fires between the caller's last failed dispatch attempt and this
	// wait is invisible (the waiter grabs the replacement channel), so
	// without the tick it would sleep to the deadline despite free
	// capacity. With it, worst case the caller re-checks capacity one
	// tick late. It also makes the queue robust against any future
	// capacity-freeing path that forgets to signal.
	const recheckInterval = 250 * time.Millisecond
	remaining := time.Until(deadline)
	if remaining <= 0 {
		return false
	}
	timer := time.NewTimer(min(remaining, recheckInterval))
	defer timer.Stop()
	select {
	case <-s.dispatchCapacitySignal():
		return true
	case <-timer.C:
		// Tick: not a timeout — the deadline may still be far away.
		// Returning true lets the caller run a fresh dispatch budget,
		// which re-reads actual capacity.
		return true
	case <-ctx.Done():
		return false
	}
}

// enterDispatchQueue admits one Compile() into the backpressure queue,
// refusing when the waiter cap is reached so overload surfaces to the
// client instead of piling up. Pair with leaveDispatchQueue.
func (s *Server) enterDispatchQueue() bool {
	if atomic.AddInt64(&s.queuedCompiles, 1) > maxQueuedCompiles {
		atomic.AddInt64(&s.queuedCompiles, -1)
		return false
	}
	return true
}

// leaveDispatchQueue releases a queue admission taken by
// enterDispatchQueue.
func (s *Server) leaveDispatchQueue() {
	atomic.AddInt64(&s.queuedCompiles, -1)
}

// dispatchFiltered submits a filter-mode request to dispatchLoop and
// blocks for its result. See dispatchRequest.filter for the semantics.
func (s *Server) dispatchFiltered(filter func(*registry.WorkerInfo) bool, label string) (*registry.WorkerInfo, error) {
	s.dispatchWG.Add(1)
	defer s.dispatchWG.Done()

	req := dispatchRequest{
		filter: filter,
		label:  label,
		result: make(chan dispatchResult, 1),
	}
	s.dispatchCh <- &req
	res := <-req.result
	return res.worker, res.err
}

// selectBuildWorker runs filter-based worker selection for the
// archive-based build paths (flutter, unity, and custom builders)
// under the same backpressure discipline as Compile(): a fast retry
// budget first, then the bounded dispatch queue when every matching
// worker is at capacity, so a burst larger than the matching capacity
// stretches over time instead of hard-failing. Returns the queue wait
// so callers can report the coordinator-side queue_ms — the worker's
// own queue field is always zero on this path.
func (s *Server) selectBuildWorker(ctx context.Context, filter func(*registry.WorkerInfo) bool, label string) (*registry.WorkerInfo, time.Duration, error) {
	const maxAttempts = 8
	var (
		queued        bool
		queueDeadline time.Time
		queueStart    time.Time
	)
	for attempt := 1; ; attempt++ {
		worker, err := s.dispatchFiltered(filter, label)
		if err == nil {
			var queueWait time.Duration
			if queued {
				queueWait = time.Since(queueStart)
			}
			return worker, queueWait, nil
		}
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 25 * time.Millisecond)
			continue
		}
		// Fast budget exhausted (~700 ms): queue instead of failing,
		// bounded by DispatchQueueTimeout and the caller's context.
		if s.config.DispatchQueueTimeout > 0 && !queued {
			if s.enterDispatchQueue() {
				queued = true
				defer s.leaveDispatchQueue()
				queueStart = time.Now()
				queueDeadline = queueStart.Add(s.config.DispatchQueueTimeout)
				log.Warn().
					Str("build", label).
					Dur("timeout", s.config.DispatchQueueTimeout).
					Msg("All matching workers at capacity; queued for dispatch")
				attempt = 0
				continue
			}
		}
		if queued {
			if !s.awaitDispatchCapacity(ctx, queueDeadline) {
				return nil, 0, fmt.Errorf("dispatch queue timeout (%v) exceeded: no worker capacity became available: %w", s.config.DispatchQueueTimeout, err)
			}
			attempt = 0
			continue
		}
		return nil, 0, err
	}
}
