package server

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const bufSize = 1024 * 1024

func setupTestServer(t *testing.T, cfg Config) (pb.BuildServiceClient, func()) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	s := New(cfg)
	pb.RegisterBuildServiceServer(srv, s)

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("Server exited: %v", err)
		}
	}()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}

	client := pb.NewBuildServiceClient(conn)
	cleanup := func() {
		conn.Close()
		srv.Stop()
	}

	return client, cleanup
}

func TestHandshake_Success(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
			CpuCores: 4,
			Os:       "linux",
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	if !resp.Accepted {
		t.Error("Expected worker to be accepted")
	}
	if resp.AssignedWorkerId == "" {
		t.Error("Expected assigned worker ID")
	}
	if resp.HeartbeatIntervalSeconds == 0 {
		t.Error("Expected heartbeat interval")
	}
}

func TestHandshake_WithAuthToken(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{
		Port:      0,
		AuthToken: "secret-token",
	})
	defer cleanup()

	// Test with wrong token
	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
		},
		AuthToken: "wrong-token",
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	if resp.Accepted {
		t.Error("Expected worker to be rejected with wrong token")
	}

	// Test with correct token
	resp, err = client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
		},
		AuthToken: "secret-token",
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}
	if !resp.Accepted {
		t.Error("Expected worker to be accepted with correct token")
	}
}

func TestHandshake_NilCapabilities(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{})
	if err == nil {
		t.Error("Expected error for nil capabilities")
	}
}

func TestHealthCheck(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.HealthCheck(context.Background(), &pb.HealthRequest{})
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if !resp.Healthy {
		t.Error("Expected server to be healthy")
	}
}

func TestGetWorkerStatus_Empty(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if resp.TotalWorkers != 0 {
		t.Errorf("Expected 0 workers, got %d", resp.TotalWorkers)
	}
}

func TestGetWorkerStatus_AfterHandshake(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	// Register a worker
	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
			CpuCores: 4,
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	// Check status
	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if resp.TotalWorkers != 1 {
		t.Errorf("Expected 1 worker, got %d", resp.TotalWorkers)
	}
}

func TestBuild_NoTaskId(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	_, err := client.Build(context.Background(), &pb.BuildRequest{})
	if err == nil {
		t.Error("Expected error for empty task_id")
	}
}

func TestCompile_NoTaskId(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	_, err := client.Compile(context.Background(), &pb.CompileRequest{})
	if err == nil {
		t.Error("Expected error for empty task_id")
	}
}

func TestGetWorkersForBuild_NoWorkers(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.GetWorkersForBuild(context.Background(), &pb.WorkersForBuildRequest{
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
	})
	if err != nil {
		t.Fatalf("GetWorkersForBuild failed: %v", err)
	}

	if resp.AvailableCount != 0 {
		t.Errorf("Expected 0 available workers, got %d", resp.AvailableCount)
	}
}

func TestNew(t *testing.T) {
	cfg := Config{
		Port:          9000,
		AuthToken:     "test-token",
		MaxConcurrent: 10,
	}
	s := New(cfg)

	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.config.Port != 9000 {
		t.Errorf("Port = %d, want 9000", s.config.Port)
	}
	if s.config.AuthToken != "test-token" {
		t.Errorf("AuthToken = %q, want 'test-token'", s.config.AuthToken)
	}
	if s.workers == nil {
		t.Error("workers map should be initialized")
	}
}

func TestStop_NotStarted(t *testing.T) {
	s := New(Config{Port: 0})
	// Stop without starting should not panic
	s.Stop()
}

func TestBuild_WithTaskId(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.Build(context.Background(), &pb.BuildRequest{
		TaskId:    "test-task-1",
		BuildType: pb.BuildType_BUILD_TYPE_CPP,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Build is not implemented yet, so should return failed status
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Errorf("Status = %v, want STATUS_FAILED", resp.Status)
	}
}

func TestCompile_WithTaskId(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "test-task-1",
		PreprocessedSource: []byte("int main() { return 0; }"),
		Compiler:           "gcc",
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Compile is not implemented yet, so should return failed status
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Errorf("Status = %v, want STATUS_FAILED", resp.Status)
	}
}

func TestHandshake_WithWorkerId(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "custom-worker-id",
			Hostname: "test-host",
			CpuCores: 8,
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	if !resp.Accepted {
		t.Error("Expected worker to be accepted")
	}
	if resp.AssignedWorkerId != "custom-worker-id" {
		t.Errorf("AssignedWorkerId = %q, want 'custom-worker-id'", resp.AssignedWorkerId)
	}
}

func TestGetWorkerStatus_WorkerDetails(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	// Register a worker with details
	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:    "detail-worker",
			Hostname:    "detail-host",
			CpuCores:    16,
			MemoryBytes: 32 * 1024 * 1024 * 1024, // 32GB
			NativeArch:  pb.Architecture_ARCH_X86_64,
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if len(resp.Workers) != 1 {
		t.Fatalf("Expected 1 worker, got %d", len(resp.Workers))
	}

	worker := resp.Workers[0]
	if worker.WorkerId != "detail-worker" {
		t.Errorf("WorkerId = %q, want 'detail-worker'", worker.WorkerId)
	}
	if worker.Host != "detail-host" {
		t.Errorf("Host = %q, want 'detail-host'", worker.Host)
	}
	if worker.CpuCores != 16 {
		t.Errorf("CpuCores = %d, want 16", worker.CpuCores)
	}
	if worker.MemoryBytes != 32*1024*1024*1024 {
		t.Errorf("MemoryBytes = %d, want 32GB", worker.MemoryBytes)
	}
	if worker.NativeArch != pb.Architecture_ARCH_X86_64 {
		t.Errorf("NativeArch = %v, want ARCH_X86_64", worker.NativeArch)
	}
}

func TestWorkerCanHandle_CPP(t *testing.T) {
	s := New(Config{})

	// Worker without CPP capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_CPP, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker without CPP caps should not handle CPP builds")
	}

	// Worker with CPP capability
	caps = &pb.WorkerCapabilities{
		Cpp: &pb.CppCapability{
			Compilers: []string{"gcc", "clang"},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_CPP, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker with CPP caps should handle CPP builds")
	}
}

func TestWorkerCanHandle_Flutter(t *testing.T) {
	s := New(Config{})

	// Worker without Flutter capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_FLUTTER, pb.TargetPlatform_PLATFORM_ANDROID) {
		t.Error("Worker without Flutter caps should not handle Flutter builds")
	}

	// Worker with Flutter but wrong platform
	caps = &pb.WorkerCapabilities{
		Flutter: &pb.FlutterCapability{
			Platforms: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_IOS},
		},
	}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_FLUTTER, pb.TargetPlatform_PLATFORM_ANDROID) {
		t.Error("Worker with Flutter/iOS should not handle Android builds")
	}

	// Worker with correct Flutter platform
	caps = &pb.WorkerCapabilities{
		Flutter: &pb.FlutterCapability{
			Platforms: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_ANDROID, pb.TargetPlatform_PLATFORM_IOS},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_FLUTTER, pb.TargetPlatform_PLATFORM_ANDROID) {
		t.Error("Worker with Flutter/Android should handle Android builds")
	}
}

func TestWorkerCanHandle_Unity(t *testing.T) {
	s := New(Config{})

	// Worker without Unity capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_UNITY, pb.TargetPlatform_PLATFORM_WINDOWS) {
		t.Error("Worker without Unity caps should not handle Unity builds")
	}

	// Worker with Unity and matching platform
	caps = &pb.WorkerCapabilities{
		Unity: &pb.UnityCapability{
			BuildTargets: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_WINDOWS},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_UNITY, pb.TargetPlatform_PLATFORM_WINDOWS) {
		t.Error("Worker with Unity/Windows should handle Windows builds")
	}
}

func TestWorkerCanHandle_Cocos(t *testing.T) {
	s := New(Config{})

	// Worker without Cocos capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_COCOS, pb.TargetPlatform_PLATFORM_WEB) {
		t.Error("Worker without Cocos caps should not handle Cocos builds")
	}

	// Worker with Cocos and matching platform
	caps = &pb.WorkerCapabilities{
		Cocos: &pb.CocosCapability{
			Platforms: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_WEB},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_COCOS, pb.TargetPlatform_PLATFORM_WEB) {
		t.Error("Worker with Cocos/Web should handle Web builds")
	}
}

func TestWorkerCanHandle_Rust(t *testing.T) {
	s := New(Config{})

	// Worker without Rust capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_RUST, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker without Rust caps should not handle Rust builds")
	}

	// Worker with Rust capability
	caps = &pb.WorkerCapabilities{
		Rust: &pb.RustCapability{
			Toolchains: []string{"stable", "nightly"},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_RUST, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker with Rust caps should handle Rust builds")
	}
}

func TestWorkerCanHandle_Go(t *testing.T) {
	s := New(Config{})

	// Worker without Go capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_GO, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker without Go caps should not handle Go builds")
	}

	// Worker with Go capability but empty version
	caps = &pb.WorkerCapabilities{
		Go: &pb.GoCapability{
			Version: "",
		},
	}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_GO, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker with empty Go version should not handle Go builds")
	}

	// Worker with Go capability
	caps = &pb.WorkerCapabilities{
		Go: &pb.GoCapability{
			Version: "1.21.0",
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_GO, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker with Go caps should handle Go builds")
	}
}

func TestWorkerCanHandle_NodeJS(t *testing.T) {
	s := New(Config{})

	// Worker without NodeJS capability
	caps := &pb.WorkerCapabilities{}
	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_NODEJS, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker without NodeJS caps should not handle NodeJS builds")
	}

	// Worker with NodeJS capability
	caps = &pb.WorkerCapabilities{
		Nodejs: &pb.NodeCapability{
			Versions: []string{"18.0.0", "20.0.0"},
		},
	}
	if !s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_NODEJS, pb.TargetPlatform_PLATFORM_LINUX) {
		t.Error("Worker with NodeJS caps should handle NodeJS builds")
	}
}

func TestWorkerCanHandle_Unspecified(t *testing.T) {
	s := New(Config{})
	caps := &pb.WorkerCapabilities{}

	if s.workerCanHandle(caps, pb.BuildType_BUILD_TYPE_UNSPECIFIED, pb.TargetPlatform_PLATFORM_UNSPECIFIED) {
		t.Error("Should not handle unspecified build type")
	}
}

func TestGetWorkersForBuild_WithCapableWorkers(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	// Register a worker with CPP capability
	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "cpp-worker",
			Hostname: "cpp-host",
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc"},
			},
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	// Query for CPP workers
	resp, err := client.GetWorkersForBuild(context.Background(), &pb.WorkersForBuildRequest{
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
	})
	if err != nil {
		t.Fatalf("GetWorkersForBuild failed: %v", err)
	}

	if resp.AvailableCount != 1 {
		t.Errorf("Expected 1 available worker, got %d", resp.AvailableCount)
	}
	if len(resp.WorkerIds) != 1 || resp.WorkerIds[0] != "cpp-worker" {
		t.Errorf("Expected ['cpp-worker'], got %v", resp.WorkerIds)
	}
}

func TestStreamBuild_NoMetadata(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	stream, err := client.StreamBuild(context.Background())
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Send only source chunk without metadata
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_SourceChunk{
			SourceChunk: []byte("int main() {}"),
		},
	})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	// Close send
	resp, err := stream.CloseAndRecv()
	if err == nil {
		t.Error("Expected error for missing metadata")
	}
	if resp != nil && resp.Status == pb.TaskStatus_STATUS_COMPLETED {
		t.Error("Expected non-completed status")
	}
}

func TestStreamBuild_WithMetadata(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	stream, err := client.StreamBuild(context.Background())
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Send metadata
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_Metadata{
			Metadata: &pb.BuildMetadata{
				TaskId:    "stream-task-1",
				BuildType: pb.BuildType_BUILD_TYPE_CPP,
			},
		},
	})
	if err != nil {
		t.Fatalf("Send metadata failed: %v", err)
	}

	// Send source chunk
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_SourceChunk{
			SourceChunk: []byte("int main() { return 0; }"),
		},
	})
	if err != nil {
		t.Fatalf("Send source failed: %v", err)
	}

	// Close and receive response
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("CloseAndRecv failed: %v", err)
	}

	// StreamBuild is not implemented yet, so should return failed status
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Errorf("Status = %v, want STATUS_FAILED", resp.Status)
	}
}

func TestMultipleWorkersRegistration(t *testing.T) {
	client, cleanup := setupTestServer(t, Config{Port: 0})
	defer cleanup()

	// Register multiple workers
	workers := []string{"worker-1", "worker-2", "worker-3"}
	for _, id := range workers {
		_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
			Capabilities: &pb.WorkerCapabilities{
				WorkerId: id,
				Hostname: id + "-host",
			},
		})
		if err != nil {
			t.Fatalf("Handshake for %s failed: %v", id, err)
		}
	}

	// Check all workers are registered
	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if resp.TotalWorkers != 3 {
		t.Errorf("Expected 3 workers, got %d", resp.TotalWorkers)
	}
	if resp.HealthyWorkers != 3 {
		t.Errorf("Expected 3 healthy workers, got %d", resp.HealthyWorkers)
	}
}
