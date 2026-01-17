package integration

import (
	"context"
	"fmt"
	"os/exec"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	coordserver "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	workerserver "github.com/h3nr1-d14z/hybridgrid/internal/worker/server"
)

func TestE2E_CoordinatorWorkerFlow(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping E2E test")
	}

	// Start coordinator
	coordCfg := coordserver.DefaultConfig()
	coordCfg.Port = 19000 // Use high port to avoid conflicts
	coordCfg.HeartbeatTTL = 30 * time.Second

	coord := coordserver.New(coordCfg)
	go func() {
		if err := coord.Start(); err != nil {
			t.Logf("Coordinator stopped: %v", err)
		}
	}()
	defer coord.Stop()

	// Wait for coordinator to start
	time.Sleep(100 * time.Millisecond)

	// Start worker
	workerCfg := workerserver.DefaultConfig()
	workerCfg.Port = 19001

	worker := workerserver.New(workerCfg)
	go func() {
		if err := worker.Start(); err != nil {
			t.Logf("Worker stopped: %v", err)
		}
	}()
	defer worker.Stop()

	// Wait for worker to start
	time.Sleep(100 * time.Millisecond)

	// Connect to coordinator
	cli, err := client.New(client.Config{
		Address:  "localhost:19000",
		Timeout:  10 * time.Second,
		Insecure: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	// Register worker with coordinator
	ctx := context.Background()
	caps := worker.Capabilities()
	caps.WorkerId = "test-worker-1"

	resp, err := cli.Handshake(ctx, &pb.HandshakeRequest{
		Capabilities:  caps,
		WorkerAddress: "localhost:19001",
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	if !resp.Accepted {
		t.Fatalf("Worker not accepted: %s", resp.Message)
	}

	t.Logf("Worker registered with ID: %s", resp.AssignedWorkerId)

	// Verify worker is in registry
	statusResp, err := cli.GetWorkerStatus(ctx)
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if statusResp.TotalWorkers != 1 {
		t.Errorf("Expected 1 worker, got %d", statusResp.TotalWorkers)
	}

	// Check health
	healthResp, err := cli.HealthCheck(ctx)
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if !healthResp.Healthy {
		t.Error("Expected coordinator to be healthy")
	}

	t.Log("E2E: Coordinator and worker communication successful")
}

func TestE2E_CompileThroughCoordinator(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping E2E compile test")
	}

	// Start coordinator
	coordCfg := coordserver.DefaultConfig()
	coordCfg.Port = 19002
	coordCfg.HeartbeatTTL = 30 * time.Second
	coordCfg.RequestTimeout = 30 * time.Second

	coord := coordserver.New(coordCfg)
	go func() {
		if err := coord.Start(); err != nil {
			t.Logf("Coordinator stopped: %v", err)
		}
	}()
	defer coord.Stop()

	time.Sleep(100 * time.Millisecond)

	// Start worker
	workerCfg := workerserver.DefaultConfig()
	workerCfg.Port = 19003

	worker := workerserver.New(workerCfg)
	go func() {
		if err := worker.Start(); err != nil {
			t.Logf("Worker stopped: %v", err)
		}
	}()
	defer worker.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect to coordinator
	cli, err := client.New(client.Config{
		Address:  "localhost:19002",
		Timeout:  30 * time.Second,
		Insecure: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	// Register worker
	ctx := context.Background()
	caps := worker.Capabilities()
	caps.WorkerId = "compile-test-worker"

	resp, err := cli.Handshake(ctx, &pb.HandshakeRequest{
		Capabilities:  caps,
		WorkerAddress: "localhost:19003",
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	if !resp.Accepted {
		t.Fatalf("Worker not accepted: %s", resp.Message)
	}

	// Wait for worker to be registered
	time.Sleep(100 * time.Millisecond)

	// Send compile request
	source := []byte(`int main() { return 0; }`)
	compileReq := &pb.CompileRequest{
		TaskId:             "test-compile-001",
		SourceHash:         "abc123",
		PreprocessedSource: source,
		CompilerArgs:       []string{"-O2"},
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		TimeoutSeconds:     30,
	}

	compileResp, err := cli.Compile(ctx, compileReq)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if compileResp.Status == pb.TaskStatus_STATUS_COMPLETED {
		if len(compileResp.ObjectFile) == 0 {
			t.Error("Expected non-empty object file")
		}
		t.Logf("Compilation succeeded: %d bytes, %dms", len(compileResp.ObjectFile), compileResp.CompilationTimeMs)
	} else {
		t.Logf("Compilation result: status=%v, exit_code=%d, stderr=%s",
			compileResp.Status, compileResp.ExitCode, compileResp.Stderr)
	}
}

func TestE2E_DirectWorkerCompile(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping direct worker test")
	}

	// Start worker directly
	workerCfg := workerserver.DefaultConfig()
	workerCfg.Port = 19004

	worker := workerserver.New(workerCfg)
	go func() {
		if err := worker.Start(); err != nil {
			t.Logf("Worker stopped: %v", err)
		}
	}()
	defer worker.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect directly to worker
	cli, err := client.New(client.Config{
		Address:  "localhost:19004",
		Timeout:  30 * time.Second,
		Insecure: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	// Compile directly with worker (use preprocessed source, no includes)
	source := []byte(`
int add(int a, int b) { return a + b; }
int main() { return add(1, 2); }
`)

	ctx := context.Background()
	compileReq := &pb.CompileRequest{
		TaskId:             "direct-compile-001",
		SourceHash:         "def456",
		PreprocessedSource: source,
		CompilerArgs:       []string{"-O2", "-Wall"},
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		TimeoutSeconds:     30,
	}

	resp, err := cli.Compile(ctx, compileReq)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Errorf("Expected STATUS_COMPLETED, got %v: %s", resp.Status, resp.Stderr)
	}

	if len(resp.ObjectFile) == 0 {
		t.Error("Expected non-empty object file")
	}

	// Verify it's a valid ELF or Mach-O file
	if len(resp.ObjectFile) >= 4 {
		magic := fmt.Sprintf("%x", resp.ObjectFile[:4])
		// ELF: 7f454c46, Mach-O: feedface or feedfacf or cffaedfe
		isELF := magic == "7f454c46"
		isMachO := magic == "feedface" || magic == "feedfacf" || magic == "cffaedfe" || magic == "cffa edfe"
		if !isELF && !isMachO {
			t.Logf("Object file magic: %s (not ELF or Mach-O, might be platform-specific)", magic)
		}
	}

	t.Logf("Direct worker compile: %d bytes object file, %dms", len(resp.ObjectFile), resp.CompilationTimeMs)
}

func TestE2E_CompileError(t *testing.T) {
	// Skip if gcc not available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not found, skipping compile error test")
	}

	// Start worker
	workerCfg := workerserver.DefaultConfig()
	workerCfg.Port = 19005

	worker := workerserver.New(workerCfg)
	go func() {
		if err := worker.Start(); err != nil {
			t.Logf("Worker stopped: %v", err)
		}
	}()
	defer worker.Stop()

	time.Sleep(100 * time.Millisecond)

	// Connect to worker
	cli, err := client.New(client.Config{
		Address:  "localhost:19005",
		Timeout:  30 * time.Second,
		Insecure: true,
	})
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer cli.Close()

	// Send invalid source
	source := []byte(`this is not valid C code { syntax error }`)

	ctx := context.Background()
	compileReq := &pb.CompileRequest{
		TaskId:             "error-compile-001",
		SourceHash:         "invalid",
		PreprocessedSource: source,
		CompilerArgs:       []string{"-c"},
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_UNSPECIFIED,
		TimeoutSeconds:     30,
	}

	resp, err := cli.Compile(ctx, compileReq)
	if err != nil {
		t.Fatalf("Compile call failed: %v", err)
	}

	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Errorf("Expected STATUS_FAILED, got %v", resp.Status)
	}

	if resp.ExitCode == 0 {
		t.Error("Expected non-zero exit code for compile error")
	}

	if resp.Stderr == "" {
		t.Error("Expected stderr output for compile error")
	}

	t.Logf("Compile error captured correctly: exit_code=%d", resp.ExitCode)
}
