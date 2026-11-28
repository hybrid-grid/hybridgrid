package server

import (
	"context"
	"net"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// taskCounters is a point-in-time copy of the coordinator task counters.
type taskCounters struct{ total, success, failed int64 }

func readTaskCounters(s *Server) taskCounters {
	return taskCounters{
		total:   atomic.LoadInt64(&s.totalTasks),
		success: atomic.LoadInt64(&s.successTasks),
		failed:  atomic.LoadInt64(&s.failedTasks),
	}
}

// assertFailureCounted checks that exactly one failed task was added since
// before: total and failed up by one, success unchanged.
func assertFailureCounted(t *testing.T, s *Server, before taskCounters) {
	t.Helper()
	after := readTaskCounters(s)
	if got := after.total - before.total; got != 1 {
		t.Errorf("totalTasks delta = %d, want 1", got)
	}
	if got := after.failed - before.failed; got != 1 {
		t.Errorf("failedTasks delta = %d, want 1", got)
	}
	if got := after.success - before.success; got != 0 {
		t.Errorf("successTasks delta = %d, want 0", got)
	}
}

func compileRequest(taskID string) *pb.CompileRequest {
	return &pb.CompileRequest{
		TaskId:             taskID,
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	}
}

// TestCompile_NoWorkerCountsFailedTask: a request that never finds a worker
// is still a failed task in the coordinator stats.
func TestCompile_NoWorkerCountsFailedTask(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	before := readTaskCounters(s)
	resp, err := s.Compile(context.Background(), compileRequest("task-no-worker"))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "no worker available") {
		t.Fatalf("resp = %v / %q, want FAILED no worker available", resp.Status, resp.Stderr)
	}
	assertFailureCounted(t, s, before)
}

// TestCompile_QueueTimeoutCountsFailedTask: the dispatch queue gives up
// while the only worker keeps rejecting with ResourceExhausted.
func TestCompile_QueueTimeoutCountsFailedTask(t *testing.T) {
	addr := startRejectingWorker(t, nil)
	s, client, cleanup := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 200 * time.Millisecond,
	})
	defer cleanup()
	registerWorker(t, s, "worker-rejects", addr, 8)

	before := readTaskCounters(s)
	resp, err := client.Compile(context.Background(), compileRequest("task-queue-timeout-stats"))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch queue timeout") {
		t.Fatalf("resp = %v / %q, want FAILED queue timeout", resp.Status, resp.Stderr)
	}
	assertFailureCounted(t, s, before)
}

// TestCompile_ClientDeadlineCountsFailedTask: the client's deadline expires
// while Compile is still waiting for a worker (backing off after a
// rejection, and later while parked in the dispatch queue).
func TestCompile_ClientDeadlineCountsFailedTask(t *testing.T) {
	t.Run("backoff after rejection", func(t *testing.T) {
		addr := startRejectingWorker(t, nil)
		s, _, cleanup := setupTestServer(t, Config{
			Port:                 0,
			HeartbeatTTL:         30 * time.Second,
			RequestTimeout:       30 * time.Second,
			DispatchQueueTimeout: 5 * time.Second,
		})
		defer cleanup()
		registerWorker(t, s, "worker-rejects", addr, 8)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		before := readTaskCounters(s)
		resp, err := s.Compile(ctx, compileRequest("task-deadline-backoff"))
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		if resp.Status != pb.TaskStatus_STATUS_FAILED || !strings.Contains(resp.Stderr, "dispatch cancelled") {
			t.Fatalf("resp = %v / %q, want FAILED dispatch cancelled", resp.Status, resp.Stderr)
		}
		assertFailureCounted(t, s, before)
	})

	t.Run("no worker while queued", func(t *testing.T) {
		s, _, cleanup := setupTestServer(t, Config{
			Port:                 0,
			HeartbeatTTL:         30 * time.Second,
			DispatchQueueTimeout: 10 * time.Second,
		})
		defer cleanup()
		registerWorker(t, s, "worker-full", "127.0.0.1:0", 1)
		if err := s.registry.IncrementTasks("worker-full"); err != nil { // never freed
			t.Fatalf("book slot: %v", err)
		}

		// Past the ~700 ms fast budget, so the request is parked in the queue.
		ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
		defer cancel()
		before := readTaskCounters(s)
		resp, err := s.Compile(ctx, compileRequest("task-deadline-queued"))
		if err != nil {
			t.Fatalf("Compile: %v", err)
		}
		if resp.Status != pb.TaskStatus_STATUS_FAILED {
			t.Fatalf("status = %v, want FAILED", resp.Status)
		}
		assertFailureCounted(t, s, before)
	})
}

// TestCompile_RejectedWithQueueDisabledCountsOnce: with the queue disabled
// the final rejection takes the normal completion bookkeeping, which already
// counts the task; the dispatch-failure accounting must not count it again.
func TestCompile_RejectedWithQueueDisabledCountsOnce(t *testing.T) {
	addr := startRejectingWorker(t, nil)
	s, client, cleanup := setupTestServer(t, Config{
		Port:                 0,
		HeartbeatTTL:         30 * time.Second,
		RequestTimeout:       30 * time.Second,
		DispatchQueueTimeout: 0,
	})
	defer cleanup()
	registerWorker(t, s, "worker-rejects", addr, 8)

	before := readTaskCounters(s)
	resp, err := client.Compile(context.Background(), compileRequest("task-rejected-no-queue"))
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", resp.Status)
	}
	assertFailureCounted(t, s, before)
}

// TestCompile_TaskCountersInvariant: after a success, a dispatch-stage
// failure and a worker-side compile failure, totalTasks must equal
// successTasks + failedTasks.
func TestCompile_TaskCountersInvariant(t *testing.T) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	workerSrv := grpc.NewServer()
	pb.RegisterBuildServiceServer(workerSrv, &mockWorkerBuildService{
		compileFn: func(_ context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
			if req.TaskId == "task-compile-error" {
				return &pb.CompileResponse{
					Status:   pb.TaskStatus_STATUS_FAILED,
					ExitCode: 1,
					Stderr:   "error: expected ';'",
				}, nil
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
	registerWorker(t, s, "worker-1", lis.Addr().String(), 1)

	if resp, err := client.Compile(context.Background(), compileRequest("task-ok")); err != nil || resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("success compile: resp=%v err=%v", resp, err)
	}
	if resp, err := client.Compile(context.Background(), compileRequest("task-compile-error")); err != nil || resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("worker-side failure: resp=%v err=%v", resp, err)
	}
	// Fill the only slot so the next request fails at dispatch.
	if err := s.registry.IncrementTasks("worker-1"); err != nil {
		t.Fatalf("book slot: %v", err)
	}
	if resp, err := client.Compile(context.Background(), compileRequest("task-no-capacity")); err != nil || resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("dispatch failure: resp=%v err=%v", resp, err)
	}

	c := readTaskCounters(s)
	if c.total != 3 || c.success != 1 || c.failed != 2 {
		t.Errorf("counters total/success/failed = %d/%d/%d, want 3/1/2", c.total, c.success, c.failed)
	}
	if c.total != c.success+c.failed {
		t.Errorf("invariant broken: total=%d != success=%d + failed=%d", c.total, c.success, c.failed)
	}
}
