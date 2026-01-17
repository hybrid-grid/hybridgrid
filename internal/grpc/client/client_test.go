package client

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const bufSize = 1024 * 1024

// mockBuildService is a simple mock implementation for testing.
type mockBuildService struct {
	pb.UnimplementedBuildServiceServer
}

func (m *mockBuildService) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	return &pb.HandshakeResponse{
		Accepted:                 true,
		Message:                  "mock accepted",
		AssignedWorkerId:         "mock-worker-1",
		HeartbeatIntervalSeconds: 30,
	}, nil
}

func (m *mockBuildService) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy:     true,
		ActiveTasks: 0,
		QueuedTasks: 0,
	}, nil
}

func (m *mockBuildService) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	return &pb.WorkerStatusResponse{
		Workers:        []*pb.WorkerStatusResponse_WorkerInfo{},
		TotalWorkers:   0,
		HealthyWorkers: 0,
	}, nil
}

func setupMockServer(t *testing.T) (*Client, func()) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterBuildServiceServer(srv, &mockBuildService{})

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

	client := &Client{
		config: Config{
			Address: "bufnet",
			Timeout: 5 * time.Second,
		},
		conn:   conn,
		client: pb.NewBuildServiceClient(conn),
	}

	cleanup := func() {
		conn.Close()
		srv.Stop()
	}

	return client, cleanup
}

func TestClient_Handshake(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
			CpuCores: 4,
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	if !resp.Accepted {
		t.Error("Expected handshake to be accepted")
	}
	if resp.AssignedWorkerId != "mock-worker-1" {
		t.Errorf("Expected worker ID 'mock-worker-1', got '%s'", resp.AssignedWorkerId)
	}
}

func TestClient_HandshakeWithCaps(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	resp, err := client.HandshakeWithCaps(context.Background(), &pb.WorkerCapabilities{
		Hostname: "test-worker",
		CpuCores: 4,
	})
	if err != nil {
		t.Fatalf("HandshakeWithCaps failed: %v", err)
	}

	if !resp.Accepted {
		t.Error("Expected handshake to be accepted")
	}
}

func TestClient_HealthCheck(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	resp, err := client.HealthCheck(context.Background())
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}

	if !resp.Healthy {
		t.Error("Expected server to be healthy")
	}
}

func TestClient_GetWorkerStatus(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	resp, err := client.GetWorkerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if resp.TotalWorkers != 0 {
		t.Errorf("Expected 0 workers, got %d", resp.TotalWorkers)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "valid insecure config",
			cfg: Config{
				Address:  "localhost:9000",
				Insecure: true,
				Timeout:  5 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "default timeout",
			cfg: Config{
				Address:  "localhost:9000",
				Insecure: true,
			},
			wantErr: false,
		},
		{
			name: "with auth token",
			cfg: Config{
				Address:   "localhost:9000",
				Insecure:  true,
				AuthToken: "test-token",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if client != nil {
				defer client.Close()
				if client.config.Timeout == 0 {
					t.Error("Timeout should not be 0 after creation")
				}
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	// Close should not error
	err := client.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Close nil conn should not error
	nilClient := &Client{}
	err = nilClient.Close()
	if err != nil {
		t.Errorf("Close() on nil conn error = %v", err)
	}
}

func TestClient_Handshake_WithAuthToken(t *testing.T) {
	client, cleanup := setupMockServer(t)
	defer cleanup()

	// Set auth token in config
	client.config.AuthToken = "test-token"

	// Handshake without explicit token should use config token
	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-worker",
		},
	})
	if err != nil {
		t.Fatalf("Handshake failed: %v", err)
	}

	if !resp.Accepted {
		t.Error("Expected handshake to be accepted")
	}
}

// Extended mock for more coverage
type extendedMockBuildService struct {
	pb.UnimplementedBuildServiceServer
}

func (m *extendedMockBuildService) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	return &pb.HandshakeResponse{
		Accepted:         true,
		AssignedWorkerId: "ext-mock-1",
	}, nil
}

func (m *extendedMockBuildService) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Healthy: true}, nil
}

func (m *extendedMockBuildService) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	return &pb.WorkerStatusResponse{
		Workers: []*pb.WorkerStatusResponse_WorkerInfo{
			{WorkerId: "worker-1", CpuCores: 4},
			{WorkerId: "worker-2", CpuCores: 8},
		},
		TotalWorkers:   2,
		HealthyWorkers: 2,
	}, nil
}

func (m *extendedMockBuildService) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_COMPLETED,
		ExitCode: 0,
	}, nil
}

func (m *extendedMockBuildService) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	return &pb.CompileResponse{
		Status:     pb.TaskStatus_STATUS_COMPLETED,
		ObjectFile: []byte("mock object"),
		WorkerId:   "worker-1",
	}, nil
}

func (m *extendedMockBuildService) GetWorkersForBuild(ctx context.Context, req *pb.WorkersForBuildRequest) (*pb.WorkersForBuildResponse, error) {
	return &pb.WorkersForBuildResponse{
		WorkerIds:      []string{"worker-1"},
		AvailableCount: 1,
	}, nil
}

func setupExtendedMockServer(t *testing.T) (*Client, func()) {
	lis := bufconn.Listen(bufSize)
	srv := grpc.NewServer()
	pb.RegisterBuildServiceServer(srv, &extendedMockBuildService{})

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

	client := &Client{
		config: Config{
			Address: "bufnet",
			Timeout: 5 * time.Second,
		},
		conn:   conn,
		client: pb.NewBuildServiceClient(conn),
	}

	cleanup := func() {
		conn.Close()
		srv.Stop()
	}

	return client, cleanup
}

func TestClient_Build(t *testing.T) {
	client, cleanup := setupExtendedMockServer(t)
	defer cleanup()

	resp, err := client.Build(context.Background(), &pb.BuildRequest{
		TaskId:    "test-task-1",
		BuildType: pb.BuildType_BUILD_TYPE_CPP,
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Errorf("Expected STATUS_COMPLETED, got %v", resp.Status)
	}
	if resp.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", resp.ExitCode)
	}
}

func TestClient_Compile(t *testing.T) {
	client, cleanup := setupExtendedMockServer(t)
	defer cleanup()

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "test-task-1",
		PreprocessedSource: []byte("int main() { return 0; }"),
		Compiler:           "gcc",
		CompilerArgs:       []string{"-c", "-o", "main.o"},
	})
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Errorf("Expected STATUS_COMPLETED, got %v", resp.Status)
	}
	if len(resp.ObjectFile) == 0 {
		t.Error("Expected object file content")
	}
}

func TestClient_GetWorkersForBuild(t *testing.T) {
	client, cleanup := setupExtendedMockServer(t)
	defer cleanup()

	resp, err := client.GetWorkersForBuild(context.Background(),
		pb.BuildType_BUILD_TYPE_CPP,
		pb.TargetPlatform_PLATFORM_LINUX,
	)
	if err != nil {
		t.Fatalf("GetWorkersForBuild failed: %v", err)
	}

	if len(resp.WorkerIds) == 0 {
		t.Error("Expected at least one worker")
	}
}

func TestClient_GetWorkerStatus_WithWorkers(t *testing.T) {
	client, cleanup := setupExtendedMockServer(t)
	defer cleanup()

	resp, err := client.GetWorkerStatus(context.Background())
	if err != nil {
		t.Fatalf("GetWorkerStatus failed: %v", err)
	}

	if resp.TotalWorkers != 2 {
		t.Errorf("Expected 2 workers, got %d", resp.TotalWorkers)
	}
	if resp.HealthyWorkers != 2 {
		t.Errorf("Expected 2 healthy workers, got %d", resp.HealthyWorkers)
	}
}

func TestConfig_Defaults(t *testing.T) {
	cfg := Config{}

	// Empty config should have zero values
	if cfg.Address != "" {
		t.Error("Default address should be empty")
	}
	if cfg.Timeout != 0 {
		t.Error("Default timeout should be 0")
	}
	if cfg.Insecure {
		t.Error("Default insecure should be false")
	}
}
