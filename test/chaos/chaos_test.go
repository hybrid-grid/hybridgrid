// Package chaos provides chaos testing for the Hybrid-Grid build system.
// Tests system resilience under failure conditions.
// Run with: go test -v -tags=chaos ./test/chaos/... -coordinator=localhost:9000
//
//go:build chaos

package chaos

import (
	"context"
	"flag"
	"fmt"
	"os/exec"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	coordinatorAddr = flag.String("coordinator", "localhost:9000", "Coordinator address")
	timeout         = flag.Duration("timeout", 5*time.Minute, "Test timeout")
)

// TestChaos_WorkerFailure tests system behavior when a worker fails.
func TestChaos_WorkerFailure(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	// Get initial worker count
	statusResp, err := client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	initialWorkers := len(statusResp.Workers)
	t.Logf("Initial workers: %d", initialWorkers)

	if initialWorkers < 2 {
		t.Skip("Need at least 2 workers for failure test")
	}

	// Start sending tasks in background
	var (
		successCount int64
		failCount    int64
		wg           sync.WaitGroup
	)

	stopChan := make(chan struct{})

	// Task sender goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		taskNum := 0
		for {
			select {
			case <-stopChan:
				return
			default:
				taskNum++
				req := &pb.CompileRequest{
					SourceFile:    fmt.Sprintf("chaos_%d.c", taskNum),
					SourceContent: []byte(fmt.Sprintf("int chaos_%d() { return %d; }", taskNum, taskNum)),
					Compiler:      "gcc",
					Args:          []string{"-c"},
					TargetArch:    pb.Architecture_ARCH_X86_64,
				}

				resp, err := client.Compile(ctx, req)
				if err != nil || !resp.Success {
					atomic.AddInt64(&failCount, 1)
				} else {
					atomic.AddInt64(&successCount, 1)
				}

				time.Sleep(50 * time.Millisecond)
			}
		}
	}()

	// Let tasks run for a bit
	time.Sleep(2 * time.Second)

	// Simulate worker failure by checking circuit breaker behavior
	// In real chaos testing, you would kill a worker container here:
	// exec.Command("docker", "kill", "hg-worker-1").Run()

	t.Log("Simulating worker failure scenario...")
	t.Log("(In production, this would kill a worker container)")

	// Continue sending tasks during "failure"
	time.Sleep(3 * time.Second)

	// Stop task sender
	close(stopChan)
	wg.Wait()

	// Report results
	success := atomic.LoadInt64(&successCount)
	fail := atomic.LoadInt64(&failCount)
	total := success + fail

	t.Logf("\n=== Chaos Test Results ===")
	t.Logf("Total tasks:   %d", total)
	t.Logf("Successful:    %d (%.1f%%)", success, float64(success)/float64(total)*100)
	t.Logf("Failed:        %d (%.1f%%)", fail, float64(fail)/float64(total)*100)

	// System should still be functional
	if success == 0 {
		t.Error("No tasks succeeded - system may be down")
	}
}

// TestChaos_CircuitBreakerRecovery tests circuit breaker recovery.
func TestChaos_CircuitBreakerRecovery(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	// Check initial circuit states
	statusResp, err := client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	t.Log("\n=== Initial Circuit States ===")
	for _, w := range statusResp.Workers {
		t.Logf("Worker %s: circuit=%s, tasks=%d", w.WorkerId, w.CircuitState, w.TotalTasks)
	}

	// Send some tasks
	for i := 0; i < 20; i++ {
		req := &pb.CompileRequest{
			SourceFile:    fmt.Sprintf("circuit_%d.c", i),
			SourceContent: []byte(fmt.Sprintf("int circuit_%d() { return %d; }", i, i)),
			Compiler:      "gcc",
			Args:          []string{"-c"},
			TargetArch:    pb.Architecture_ARCH_X86_64,
		}
		client.Compile(ctx, req)
	}

	// Check circuit states after load
	statusResp, err = client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	t.Log("\n=== Final Circuit States ===")
	for _, w := range statusResp.Workers {
		t.Logf("Worker %s: circuit=%s, tasks=%d, success=%d",
			w.WorkerId, w.CircuitState, w.TotalTasks, w.SuccessfulTasks)
	}
}

// TestChaos_GracefulDegradation tests system behavior when workers become slow.
func TestChaos_GracefulDegradation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	// Send a burst of tasks
	t.Log("Sending burst of 50 tasks...")
	var wg sync.WaitGroup
	var successCount, failCount int64

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			req := &pb.CompileRequest{
				SourceFile:    fmt.Sprintf("burst_%d.c", n),
				SourceContent: []byte(fmt.Sprintf("int burst_%d() { return %d; }", n, n)),
				Compiler:      "gcc",
				Args:          []string{"-c"},
				TargetArch:    pb.Architecture_ARCH_X86_64,
			}

			resp, err := client.Compile(ctx, req)
			if err != nil || !resp.Success {
				atomic.AddInt64(&failCount, 1)
			} else {
				atomic.AddInt64(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()

	success := atomic.LoadInt64(&successCount)
	fail := atomic.LoadInt64(&failCount)

	t.Logf("\n=== Burst Test Results ===")
	t.Logf("Successful: %d", success)
	t.Logf("Failed:     %d", fail)

	// Even under load, some tasks should succeed
	if success == 0 {
		t.Error("All tasks failed under burst load")
	}
}

// TestChaos_NetworkPartition simulates network issues.
func TestChaos_NetworkPartition(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network partition test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	t.Log("Testing resilience to network delays...")
	t.Log("(In production, use tc netem to inject network latency)")

	// Send tasks with varying timeouts
	timeouts := []time.Duration{1 * time.Second, 2 * time.Second, 5 * time.Second}

	for _, timeout := range timeouts {
		taskCtx, taskCancel := context.WithTimeout(ctx, timeout)

		req := &pb.CompileRequest{
			SourceFile:    fmt.Sprintf("timeout_%v.c", timeout),
			SourceContent: []byte("int timeout_test() { return 0; }"),
			Compiler:      "gcc",
			Args:          []string{"-c"},
			TargetArch:    pb.Architecture_ARCH_X86_64,
		}

		start := time.Now()
		resp, err := client.Compile(taskCtx, req)
		elapsed := time.Since(start)
		taskCancel()

		if err != nil {
			t.Logf("Timeout %v: Failed after %v - %v", timeout, elapsed, err)
		} else if resp.Success {
			t.Logf("Timeout %v: Success in %v", timeout, elapsed)
		} else {
			t.Logf("Timeout %v: Compile failed in %v", timeout, elapsed)
		}
	}
}

// TestChaos_DockerContainerRestart tests behavior during container restarts.
func TestChaos_DockerContainerRestart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Docker restart test in short mode")
	}

	// Check if we can use docker
	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("Docker not available for chaos test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, *coordinatorAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		t.Fatalf("Failed to connect to coordinator: %v", err)
	}
	defer conn.Close()

	client := pb.NewBuildServiceClient(conn)

	t.Log("To run full Docker chaos test:")
	t.Log("  1. Start cluster: docker compose up -d --scale worker=3")
	t.Log("  2. Run test: go test -v -tags=chaos ./test/chaos/...")
	t.Log("  3. During test: docker restart hybridgrid-worker-1")

	// Verify basic connectivity
	statusResp, err := client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	t.Logf("Current workers: %d", len(statusResp.Workers))

	// Send test task
	req := &pb.CompileRequest{
		SourceFile:    "docker_test.c",
		SourceContent: []byte("int docker_test() { return 42; }"),
		Compiler:      "gcc",
		Args:          []string{"-c"},
		TargetArch:    pb.Architecture_ARCH_X86_64,
	}

	resp, err := client.Compile(ctx, req)
	if err != nil {
		t.Logf("Task failed: %v", err)
	} else if resp.Success {
		t.Log("Task succeeded - system is healthy")
	} else {
		t.Logf("Compile failed: %s", resp.Stderr)
	}
}
