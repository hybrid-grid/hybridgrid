package server

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// startRejectingWorker serves a worker whose Compile always rejects with
// ResourceExhausted; onCall (optional) runs first on every call.
func startRejectingWorker(t *testing.T, onCall func(call int32)) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	var calls atomic.Int32
	srv := grpc.NewServer()
	pb.RegisterBuildServiceServer(srv, &mockWorkerBuildService{
		compileFn: func(context.Context, *pb.CompileRequest) (*pb.CompileResponse, error) {
			n := calls.Add(1)
			if onCall != nil {
				onCall(n)
			}
			return nil, status.Error(codes.ResourceExhausted, "too many concurrent tasks")
		},
	})
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(srv.Stop)
	return lis.Addr().String()
}

// assertEveryStartedTaskCompletedOnce checks that every task ID with a
// started event has exactly one completed event, and that it is a failure
// carrying the response's stderr.
func assertEveryStartedTaskCompletedOnce(t *testing.T, n *lockedEventNotifier, resp *pb.CompileResponse, wantTask, wantWorker, wantBuild string) {
	t.Helper()
	started, completed := n.snapshot()
	if len(started) == 0 {
		t.Fatal("no started events; test did not exercise a dispatch attempt")
	}
	completions := map[string][]*TaskEvent{}
	for _, e := range completed {
		completions[e.ID] = append(completions[e.ID], e)
	}
	seen := map[string]bool{}
	for _, e := range started {
		if seen[e.ID] {
			continue
		}
		seen[e.ID] = true
		if got := len(completions[e.ID]); got != 1 {
			t.Errorf("task %q: %d completed events, want exactly 1", e.ID, got)
		}
	}
	if len(completions) != 1 || len(completions[wantTask]) != 1 {
		t.Fatalf("completed events = %+v, want exactly one for %q", completed, wantTask)
	}
	ev := completions[wantTask][0]
	if ev.Status != "failed" {
		t.Errorf("Status = %q, want failed", ev.Status)
	}
	if ev.BuildType != "cpp" {
		t.Errorf("BuildType = %q, want cpp", ev.BuildType)
	}
	if ev.BuildID != wantBuild {
		t.Errorf("BuildID = %q, want %q", ev.BuildID, wantBuild)
	}
	if ev.WorkerID != wantWorker {
		t.Errorf("WorkerID = %q, want %q", ev.WorkerID, wantWorker)
	}
	if ev.ExitCode != 1 {
		t.Errorf("ExitCode = %d, want 1", ev.ExitCode)
	}
	if ev.ErrorMessage == "" || ev.ErrorMessage != resp.Stderr {
		t.Errorf("ErrorMessage = %q, want response stderr %q", ev.ErrorMessage, resp.Stderr)
	}
	if ev.StartedAtMs == 0 || ev.CompletedAtMs < ev.StartedAtMs {
		t.Errorf("timestamps started=%d completed=%d, want started set and completed >= started", ev.StartedAtMs, ev.CompletedAtMs)
	}
}

func registerWorker(t *testing.T, s *Server, id, addr string, maxParallel int32) {
	t.Helper()
	if err := s.registry.Add(&registry.WorkerInfo{
		ID:      id,
		Address: addr,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: maxParallel,
	}); err != nil {
		t.Fatalf("add worker: %v", err)
	}
}

// TestCompile_QueueTimeoutAfterRejectCompletesTask: a worker that always
// rejects sends the request into the dispatch queue, which then times out.
// Every dispatch attempt announced a started event, so the task must be
// closed with a failed completion rather than shown as running forever.
func TestCompile_QueueTimeoutAfterRejectCompletesTask(t *testing.T) {
	addr := startRejectingWorker(t, nil)
	s, client, cleanup := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 200 * time.Millisecond,
	})
	defer cleanup()
	notifier := &lockedEventNotifier{}
	s.SetEventNotifier(notifier)
	registerWorker(t, s, "worker-rejects", addr, 8)

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "task-queue-timeout",
		BuildId:            "build-queue-timeout",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch queue timeout") {
		t.Fatalf("resp = %v / %q, want FAILED queue timeout", resp.Status, resp.Stderr)
	}
	assertEveryStartedTaskCompletedOnce(t, notifier, resp, "task-queue-timeout", "worker-rejects", "build-queue-timeout")
}

// TestCompile_CancelledAfterRejectCompletesTask: the client's context ends
// while the rejected attempt is backing off.
func TestCompile_CancelledAfterRejectCompletesTask(t *testing.T) {
	addr := startRejectingWorker(t, nil)
	s, _, cleanup := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 5 * time.Second,
	})
	defer cleanup()
	notifier := &lockedEventNotifier{}
	s.SetEventNotifier(notifier)
	registerWorker(t, s, "worker-rejects", addr, 8)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	resp, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "task-cancelled",
		BuildId:            "build-cancelled",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch cancelled") {
		t.Fatalf("resp = %v / %q, want FAILED dispatch cancelled", resp.Status, resp.Stderr)
	}
	assertEveryStartedTaskCompletedOnce(t, notifier, resp, "task-cancelled", "worker-rejects", "build-cancelled")
}

// TestCompile_NoWorkerAfterRejectCompletesTask: an attempt is started and
// rejected, then the worker looks full (another booking lands during the
// rejected call) so the next selection finds no worker and the queue times
// out. The earlier started event must still be closed.
func TestCompile_NoWorkerAfterRejectCompletesTask(t *testing.T) {
	s, client, stop := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 200 * time.Millisecond,
	})
	defer stop()
	addr := startRejectingWorker(t, func(call int32) {
		if call == 1 {
			// Another task takes a slot: once this attempt's booking is
			// released the worker (MaxParallel 1) is still full.
			_ = s.registry.IncrementTasks("worker-rejects")
		}
	})
	notifier := &lockedEventNotifier{}
	s.SetEventNotifier(notifier)
	registerWorker(t, s, "worker-rejects", addr, 1)

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "task-no-worker-after-reject",
		BuildId:            "build-no-worker",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", resp.Status)
	}
	assertEveryStartedTaskCompletedOnce(t, notifier, resp, "task-no-worker-after-reject", "worker-rejects", "build-no-worker")
}
