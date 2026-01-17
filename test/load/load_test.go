// Package load provides load testing for the Hybrid-Grid build system.
// Run with: go test -v -tags=load ./test/load/... -coordinator=localhost:9000
//
//go:build load

package load

import (
	"context"
	"flag"
	"fmt"
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
	numWorkers      = flag.Int("workers", 4, "Expected number of workers")
	numTasks        = flag.Int("tasks", 100, "Number of tasks to submit")
	concurrency     = flag.Int("concurrency", 10, "Number of concurrent requests")
	timeout         = flag.Duration("timeout", 5*time.Minute, "Test timeout")
)

// TestLoadBasic runs a basic load test against the coordinator.
func TestLoadBasic(t *testing.T) {
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

	// Check workers are available
	statusResp, err := client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	activeWorkers := len(statusResp.Workers)
	t.Logf("Connected workers: %d (expected: %d)", activeWorkers, *numWorkers)

	if activeWorkers < *numWorkers {
		t.Logf("WARNING: Fewer workers than expected")
	}

	// Run load test
	var (
		successCount int64
		failCount    int64
		totalLatency int64
		wg           sync.WaitGroup
		sem          = make(chan struct{}, *concurrency)
	)

	startTime := time.Now()

	for i := 0; i < *numTasks; i++ {
		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore

		go func(taskNum int) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore

			taskStart := time.Now()

			// Create a simple compile request
			req := &pb.CompileRequest{
				SourceFile:    fmt.Sprintf("test_%d.c", taskNum),
				SourceContent: []byte(fmt.Sprintf("int test_%d() { return %d; }", taskNum, taskNum)),
				Compiler:      "gcc",
				Args:          []string{"-c", "-o", fmt.Sprintf("test_%d.o", taskNum)},
				TargetArch:    pb.Architecture_ARCH_X86_64,
			}

			resp, err := client.Compile(ctx, req)
			latency := time.Since(taskStart).Milliseconds()

			if err != nil || !resp.Success {
				atomic.AddInt64(&failCount, 1)
				if err != nil {
					t.Logf("Task %d failed: %v", taskNum, err)
				} else {
					t.Logf("Task %d compile failed: %s", taskNum, resp.Stderr)
				}
			} else {
				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&totalLatency, latency)
			}
		}(i)
	}

	wg.Wait()
	totalTime := time.Since(startTime)

	// Report results
	success := atomic.LoadInt64(&successCount)
	fail := atomic.LoadInt64(&failCount)
	avgLatency := float64(0)
	if success > 0 {
		avgLatency = float64(atomic.LoadInt64(&totalLatency)) / float64(success)
	}

	t.Logf("\n=== Load Test Results ===")
	t.Logf("Total tasks:     %d", *numTasks)
	t.Logf("Successful:      %d (%.1f%%)", success, float64(success)/float64(*numTasks)*100)
	t.Logf("Failed:          %d (%.1f%%)", fail, float64(fail)/float64(*numTasks)*100)
	t.Logf("Total time:      %v", totalTime)
	t.Logf("Throughput:      %.2f tasks/sec", float64(*numTasks)/totalTime.Seconds())
	t.Logf("Avg latency:     %.2f ms", avgLatency)
	t.Logf("Concurrency:     %d", *concurrency)
	t.Logf("Workers:         %d", activeWorkers)

	// Verify success rate
	successRate := float64(success) / float64(*numTasks)
	if successRate < 0.95 {
		t.Errorf("Success rate %.1f%% is below 95%% threshold", successRate*100)
	}
}

// TestLoadSustained runs a sustained load test over a longer period.
func TestLoadSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping sustained load test in short mode")
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

	// Run for 60 seconds
	duration := 60 * time.Second
	ticker := time.NewTicker(100 * time.Millisecond) // 10 requests/sec
	defer ticker.Stop()

	var (
		successCount int64
		failCount    int64
		taskNum      int
	)

	deadline := time.Now().Add(duration)
	for time.Now().Before(deadline) {
		select {
		case <-ticker.C:
			taskNum++
			go func(n int) {
				req := &pb.CompileRequest{
					SourceFile:    fmt.Sprintf("sustained_%d.c", n),
					SourceContent: []byte(fmt.Sprintf("int sustained_%d() { return %d; }", n, n)),
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
			}(taskNum)
		case <-ctx.Done():
			t.Fatal("Test timed out")
		}
	}

	// Wait for in-flight requests
	time.Sleep(5 * time.Second)

	success := atomic.LoadInt64(&successCount)
	fail := atomic.LoadInt64(&failCount)
	total := success + fail

	t.Logf("\n=== Sustained Load Results ===")
	t.Logf("Duration:    %v", duration)
	t.Logf("Total:       %d", total)
	t.Logf("Successful:  %d (%.1f%%)", success, float64(success)/float64(total)*100)
	t.Logf("Failed:      %d", fail)
	t.Logf("Rate:        %.2f req/sec", float64(total)/duration.Seconds())

	if float64(success)/float64(total) < 0.90 {
		t.Errorf("Success rate below 90%% during sustained load")
	}
}

// TestLoadWorkerDistribution verifies tasks are distributed across workers.
func TestLoadWorkerDistribution(t *testing.T) {
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

	// Get initial worker status
	statusResp, err := client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get worker status: %v", err)
	}

	if len(statusResp.Workers) < 2 {
		t.Skip("Need at least 2 workers for distribution test")
	}

	workerTasksBefore := make(map[string]int64)
	for _, w := range statusResp.Workers {
		workerTasksBefore[w.WorkerId] = w.TotalTasks
	}

	// Submit tasks
	numTasks := 50
	var wg sync.WaitGroup
	for i := 0; i < numTasks; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			req := &pb.CompileRequest{
				SourceFile:    fmt.Sprintf("dist_%d.c", n),
				SourceContent: []byte(fmt.Sprintf("int dist_%d() { return %d; }", n, n)),
				Compiler:      "gcc",
				Args:          []string{"-c"},
				TargetArch:    pb.Architecture_ARCH_X86_64,
			}
			client.Compile(ctx, req)
		}(i)
	}
	wg.Wait()

	// Get final worker status
	statusResp, err = client.GetWorkerStatus(ctx, &pb.GetWorkerStatusRequest{})
	if err != nil {
		t.Fatalf("Failed to get final worker status: %v", err)
	}

	t.Logf("\n=== Worker Distribution ===")
	totalNew := int64(0)
	for _, w := range statusResp.Workers {
		before := workerTasksBefore[w.WorkerId]
		newTasks := w.TotalTasks - before
		totalNew += newTasks
		t.Logf("Worker %s: +%d tasks (total: %d)", w.WorkerId, newTasks, w.TotalTasks)
	}

	// Check that tasks were distributed
	if totalNew < int64(numTasks/2) {
		t.Errorf("Expected at least %d tasks to complete, got %d", numTasks/2, totalNew)
	}
}
