package server

import (
	"context"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// lockedEventNotifier records events behind a mutex: handlers may keep
// emitting after the client has already gone away, so tests that poll
// for events cannot use the unsynchronised mockEventNotifier.
type lockedEventNotifier struct {
	mu        sync.Mutex
	started   []*TaskEvent
	completed []*TaskEvent
}

func (n *lockedEventNotifier) NotifyTaskStarted(e *TaskEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.started = append(n.started, e)
}

func (n *lockedEventNotifier) NotifyTaskCompleted(e *TaskEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.completed = append(n.completed, e)
}

func (n *lockedEventNotifier) snapshot() (started, completed []*TaskEvent) {
	n.mu.Lock()
	defer n.mu.Unlock()
	return append([]*TaskEvent(nil), n.started...), append([]*TaskEvent(nil), n.completed...)
}

// TestCompile_RejectedFinalAttemptReleasesSlotOnce: with the dispatch
// queue disabled, a worker that rejects every attempt surfaces as a
// failed compile. Each rejected attempt must release its booking exactly
// once — the final one used to be released by both the dispatch loop and
// the completion bookkeeping, freeing a slot owned by another task.
func TestCompile_RejectedFinalAttemptReleasesSlotOnce(t *testing.T) {
	var calls atomic.Int32
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	workerSrv := grpc.NewServer()
	pb.RegisterBuildServiceServer(workerSrv, &mockWorkerBuildService{
		compileFn: func(context.Context, *pb.CompileRequest) (*pb.CompileResponse, error) {
			calls.Add(1)
			return nil, status.Error(codes.ResourceExhausted, "too many concurrent tasks")
		},
	})
	go func() { _ = workerSrv.Serve(lis) }()
	defer workerSrv.Stop()

	s, client, cleanup := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 0, // queue disabled: pre-backpressure reject
	})
	defer cleanup()

	const workerID = "worker-always-rejects"
	if err := s.registry.Add(&registry.WorkerInfo{
		ID:      workerID,
		Address: lis.Addr().String(),
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 8,
	}); err != nil {
		t.Fatalf("add worker: %v", err)
	}
	// Another task is running on the same worker; its slot must survive.
	if err := s.registry.IncrementTasks(workerID); err != nil {
		t.Fatalf("book other task: %v", err)
	}

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "task-always-rejected",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", resp.Status)
	}
	if got := calls.Load(); got != maxDispatchAttempts {
		t.Fatalf("worker calls = %d, want %d", got, maxDispatchAttempts)
	}

	w, ok := s.registry.Get(workerID)
	if !ok {
		t.Fatal("worker missing from registry")
	}
	if w.ActiveTasks != 1 {
		t.Errorf("ActiveTasks = %d, want 1 (the other task's slot)", w.ActiveTasks)
	}
	if w.FailedTasks != maxDispatchAttempts {
		t.Errorf("FailedTasks = %d, want %d (one per rejected attempt)", w.FailedTasks, maxDispatchAttempts)
	}
	val, ok := s.activeTasksByWorker.Load(workerID)
	if !ok {
		t.Fatal("per-worker active counter missing")
	}
	if got := atomic.LoadInt64(val.(*int64)); got != 0 {
		t.Errorf("per-worker active counter = %d, want 0", got)
	}
}

// TestBuild_ConnectFailureReleasesSlot: selectBuildWorker books a slot
// before the coordinator dials the worker. If the dial fails the slot
// must be released and the build counted as failed.
func TestBuild_ConnectFailureReleasesSlot(t *testing.T) {
	// Reserve a port, then close it so nothing is listening there.
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	deadAddr := lis.Addr().String()
	lis.Close()

	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 30 * time.Second})
	defer cleanup()
	notifier := &lockedEventNotifier{}
	s.SetEventNotifier(notifier)

	const workerID = "worker-unreachable"
	if err := s.registry.Add(&registry.WorkerInfo{
		ID:      workerID,
		Address: deadAddr,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:     pb.Architecture_ARCH_X86_64,
			CustomBuilders: []string{"demo"},
		},
		MaxParallel: 2,
	}); err != nil {
		t.Fatalf("add worker: %v", err)
	}

	// The blocking dial only gives up when the request context ends, so
	// the client sees its own deadline; the handler settles just after.
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	_, _ = client.Build(ctx, &pb.BuildRequest{
		TaskId:        "custom-unreachable",
		SourceHash:    "aabb",
		SourceArchive: []byte("fake-tar"),
		CustomBuilder: "demo",
		BuildId:       "build-session-connect",
	})

	deadline := time.Now().Add(5 * time.Second)
	for {
		w, ok := s.registry.Get(workerID)
		if !ok {
			t.Fatal("worker missing from registry")
		}
		failed := atomic.LoadInt64(&s.failedTasks)
		_, completed := notifier.snapshot()
		if w.ActiveTasks == 0 && failed == 1 && len(completed) > 0 {
			// Let any (wrong) second settlement land before asserting.
			time.Sleep(100 * time.Millisecond)
			if total := atomic.LoadInt64(&s.totalTasks); total != 1 {
				t.Errorf("totalTasks = %d, want 1", total)
			}
			if failed := atomic.LoadInt64(&s.failedTasks); failed != 1 {
				t.Errorf("failedTasks = %d, want 1", failed)
			}
			if w2, _ := s.registry.Get(workerID); w2.ActiveTasks != 0 || w2.FailedTasks != 1 {
				t.Errorf("worker ActiveTasks/FailedTasks = %d/%d, want 0/1", w2.ActiveTasks, w2.FailedTasks)
			}
			started, completed := notifier.snapshot()
			if len(started) != 1 || len(completed) != 1 {
				t.Fatalf("events started/completed = %d/%d, want 1/1", len(started), len(completed))
			}
			ev := completed[0]
			if ev.Status != "failed" || ev.WorkerID != workerID || ev.BuildID != "build-session-connect" || ev.ID != "custom-unreachable" {
				t.Errorf("completion event = %+v, want failed on %s for build-session-connect", ev, workerID)
			}
			if started[0].BuildID != ev.BuildID {
				t.Errorf("started BuildID = %q, want %q", started[0].BuildID, ev.BuildID)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("slot not released after connect failure: ActiveTasks = %d, failedTasks = %d", w.ActiveTasks, failed)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestBuild_Unity_CompletionEventCarriesBuildID: unity builds must group
// under their build session like flutter and custom builds do.
func TestBuild_Unity_CompletionEventCarriesBuildID(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 30 * time.Second})
	defer cleanup()
	notifier := &mockEventNotifier{}
	s.SetEventNotifier(notifier)

	addr, workerCleanup := setupTestWorker(t, func(context.Context, *pb.BuildRequest) (*pb.BuildResponse, error) {
		return &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED, Artifacts: []byte("unity-artifact")}, nil
	})
	defer workerCleanup()

	if err := s.registry.Add(&registry.WorkerInfo{
		ID:      "unity-worker-1",
		Address: addr,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Unity: &pb.UnityCapability{
				Versions:     []string{"2022.3"},
				BuildTargets: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_ANDROID},
			},
		},
		MaxParallel: 1,
	}); err != nil {
		t.Fatalf("add worker: %v", err)
	}

	req := newUnityBuildRequest("unity-build-id", "hash-build-id")
	req.BuildId = "build-session-unity"
	resp, err := client.Build(context.Background(), req)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("status = %v, want COMPLETED (stderr: %s)", resp.Status, resp.Stderr)
	}
	if len(notifier.completed) != 1 {
		t.Fatalf("completed events = %d, want 1", len(notifier.completed))
	}
	if got := notifier.completed[0].BuildID; got != "build-session-unity" {
		t.Errorf("completion event BuildID = %q, want %q", got, "build-session-unity")
	}
	if len(notifier.started) != 1 || notifier.started[0].BuildID != notifier.completed[0].BuildID {
		t.Errorf("started and completed events must share the BuildID")
	}
}

func TestBackoffSleep(t *testing.T) {
	t.Run("waits when context is live", func(t *testing.T) {
		start := time.Now()
		if !backoffSleep(context.Background(), 1) {
			t.Fatal("backoffSleep returned false with a live context")
		}
		if elapsed := time.Since(start); elapsed < 25*time.Millisecond {
			t.Errorf("returned after %v, want at least 25ms", elapsed)
		}
	})
	t.Run("returns early when context is done", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		start := time.Now()
		if backoffSleep(ctx, 400) { // would be a 10s wait
			t.Fatal("backoffSleep returned true with a cancelled context")
		}
		if elapsed := time.Since(start); elapsed > time.Second {
			t.Errorf("took %v to notice cancellation", elapsed)
		}
	})
}

// TestCompile_RejectionThenSuccessReleasesSlotOnce guards against a stale
// unbooked flag: attempts rejected by the worker unbook themselves, but a
// later attempt that succeeds is booked afresh and must be released by
// the normal completion path exactly once.
func TestCompile_RejectionThenSuccessReleasesSlotOnce(t *testing.T) {
	var calls atomic.Int32
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	workerSrv := grpc.NewServer()
	pb.RegisterBuildServiceServer(workerSrv, &mockWorkerBuildService{
		compileFn: func(context.Context, *pb.CompileRequest) (*pb.CompileResponse, error) {
			if calls.Add(1) <= 2 {
				return nil, status.Error(codes.ResourceExhausted, "too many concurrent tasks")
			}
			return &pb.CompileResponse{Status: pb.TaskStatus_STATUS_COMPLETED, ObjectFile: []byte("obj")}, nil
		},
	})
	go func() { _ = workerSrv.Serve(lis) }()
	defer workerSrv.Stop()

	s, client, cleanup := setupTestServer(t, Config{
		Port:           0,
		HeartbeatTTL:   30 * time.Second,
		RequestTimeout: 30 * time.Second,
	})
	defer cleanup()

	const workerID = "worker-rejects-twice"
	if err := s.registry.Add(&registry.WorkerInfo{
		ID:      workerID,
		Address: lis.Addr().String(),
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 8,
	}); err != nil {
		t.Fatalf("add worker: %v", err)
	}
	if err := s.registry.IncrementTasks(workerID); err != nil {
		t.Fatalf("book other task: %v", err)
	}

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "task-rejected-twice",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("status = %v, want COMPLETED (stderr: %s)", resp.Status, resp.Stderr)
	}

	w, _ := s.registry.Get(workerID)
	if w.ActiveTasks != 1 {
		t.Errorf("ActiveTasks = %d, want 1 (the other task's slot)", w.ActiveTasks)
	}
	if w.FailedTasks != 2 || w.SuccessfulTasks != 1 {
		t.Errorf("Failed/SuccessfulTasks = %d/%d, want 2/1", w.FailedTasks, w.SuccessfulTasks)
	}
	val, _ := s.activeTasksByWorker.Load(workerID)
	if got := atomic.LoadInt64(val.(*int64)); got != 0 {
		t.Errorf("per-worker active counter = %d, want 0", got)
	}
}

// TestCompile_NoCapacityHonoursContextDeadline: with nothing eligible the
// handler retries for ~700ms. A client deadline shorter than that must end
// the wait promptly instead of sleeping through the whole backoff budget.
func TestCompile_NoCapacityHonoursContextDeadline(t *testing.T) {
	s := New(Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 30 * time.Second})
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	start := time.Now()
	resp, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "task-no-capacity",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch cancelled") {
		t.Errorf("response = %v %q, want FAILED with a dispatch-cancelled message", resp.Status, resp.Stderr)
	}
	if elapsed > 400*time.Millisecond {
		t.Errorf("Compile took %v after a 60ms deadline, want prompt return", elapsed)
	}
}

// TestBuild_NoCapacityHonoursContextDeadline: same for the archive build
// handlers' worker selection.
func TestBuild_NoCapacityHonoursContextDeadline(t *testing.T) {
	s := New(Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 30 * time.Second})
	defer s.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	start := time.Now()
	resp, err := s.handleCustomBuild(ctx, &pb.BuildRequest{
		TaskId:        "custom-no-capacity",
		SourceHash:    "aabb",
		SourceArchive: []byte("fake-tar"),
		CustomBuilder: "demo",
	}, "")
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("handleCustomBuild: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch cancelled") {
		t.Errorf("response = %v %q, want FAILED with a dispatch-cancelled message", resp.Status, resp.Stderr)
	}
	if elapsed > 400*time.Millisecond {
		t.Errorf("handleCustomBuild took %v after a 60ms deadline, want prompt return", elapsed)
	}
}
