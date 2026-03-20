package server

import (
	"context"
	"net"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

const bufSize = 1024 * 1024

// setupTestServer creates a bufconn-based test server and returns a client + cleanup func.
func setupTestServer(t *testing.T, cfg Config) (*Server, pb.BuildServiceClient, func()) {
	t.Helper()

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
	require.NoError(t, err)

	client := pb.NewBuildServiceClient(conn)
	cleanup := func() {
		conn.Close()
		srv.Stop()
		s.Stop()
	}

	return s, client, cleanup
}

type mockWorkerBuildService struct {
	pb.UnimplementedBuildServiceServer
	buildFn func(context.Context, *pb.BuildRequest) (*pb.BuildResponse, error)
}

func (m *mockWorkerBuildService) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	if m.buildFn != nil {
		return m.buildFn(ctx, req)
	}
	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "mock build not configured",
	}, nil
}

func setupTestWorker(t *testing.T, buildFn func(context.Context, *pb.BuildRequest) (*pb.BuildResponse, error)) (string, func()) {
	t.Helper()

	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	srv := grpc.NewServer()
	pb.RegisterBuildServiceServer(srv, &mockWorkerBuildService{buildFn: buildFn})

	go func() {
		if err := srv.Serve(lis); err != nil {
			t.Logf("Worker server exited: %v", err)
		}
	}()

	cleanup := func() {
		srv.Stop()
		lis.Close()
	}

	return lis.Addr().String(), cleanup
}

func newFlutterBuildRequest(taskID, sourceHash string) *pb.BuildRequest {
	return &pb.BuildRequest{
		TaskId:         taskID,
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceHash:     sourceHash,
		SourceArchive:  []byte("flutter-archive"),
		Config: &pb.BuildRequest_FlutterConfig{
			FlutterConfig: &pb.FlutterConfig{
				FlutterVersion: "3.10.0",
				BuildMode:      "release",
			},
		},
	}
}

// --- DefaultConfig ---

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.Equal(t, 50051, cfg.Port)
	assert.Equal(t, 60*time.Second, cfg.HeartbeatTTL)
	assert.Equal(t, 120*time.Second, cfg.RequestTimeout)
	assert.Empty(t, cfg.AuthToken)
}

// --- Server creation ---

func TestNew(t *testing.T) {
	cfg := Config{Port: 9000, AuthToken: "token", HeartbeatTTL: 30 * time.Second}
	s := New(cfg)

	require.NotNil(t, s)
	assert.Equal(t, 9000, s.config.Port)
	assert.Equal(t, "token", s.config.AuthToken)
	assert.NotNil(t, s.registry)
	assert.NotNil(t, s.scheduler)
	assert.NotNil(t, s.circuitManager)

	s.Stop()
}

func TestStop_NotStarted(t *testing.T) {
	s := New(Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	// Should not panic when server hasn't started
	s.Stop()
}

func TestRegistry(t *testing.T) {
	s := New(Config{HeartbeatTTL: 30 * time.Second})
	defer s.Stop()

	assert.NotNil(t, s.Registry())
}

// --- Handshake ---

func TestHandshake_NilCapabilities(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestHandshake_Success(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-host",
			CpuCores: 8,
			Os:       "linux",
		},
	})
	require.NoError(t, err)

	assert.True(t, resp.Accepted)
	assert.NotEmpty(t, resp.AssignedWorkerId)
	assert.Contains(t, resp.AssignedWorkerId, "worker-test-host-")
	assert.NotEmpty(t, resp.Message)
	assert.Greater(t, resp.HeartbeatIntervalSeconds, int32(0))
}

func TestHandshake_WithCustomWorkerId(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "my-worker",
			Hostname: "host",
		},
	})
	require.NoError(t, err)

	assert.True(t, resp.Accepted)
	assert.Equal(t, "my-worker", resp.AssignedWorkerId)
}

func TestHandshake_WithWorkerAddress(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		WorkerAddress: "10.0.0.5:8080",
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "addr-worker",
			Hostname: "host",
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	// Verify the address was stored
	w, ok := s.registry.Get("addr-worker")
	require.True(t, ok)
	assert.Equal(t, "10.0.0.5:8080", w.Address)
}

func TestHandshake_DefaultWorkerAddress(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "default-addr-worker",
			Hostname: "my-host",
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	w, ok := s.registry.Get("default-addr-worker")
	require.True(t, ok)
	assert.Equal(t, "my-host:50052", w.Address)
}

func TestHandshake_MaxParallelDefaults(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	// MaxParallelTasks = 0 should default to 4
	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:         "mp-worker",
			Hostname:         "host",
			MaxParallelTasks: 0,
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	w, ok := s.registry.Get("mp-worker")
	require.True(t, ok)
	assert.Equal(t, int32(4), w.MaxParallel)
}

func TestHandshake_MaxParallelCustom(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:         "mp-worker2",
			Hostname:         "host",
			MaxParallelTasks: 12,
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	w, ok := s.registry.Get("mp-worker2")
	require.True(t, ok)
	assert.Equal(t, int32(12), w.MaxParallel)
}

func TestHandshake_AuthToken_Rejected(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, AuthToken: "secret", HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		AuthToken: "wrong",
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "host",
		},
	})
	require.NoError(t, err)
	assert.False(t, resp.Accepted)
	assert.Contains(t, resp.Message, "invalid auth token")
}

func TestHandshake_AuthToken_Accepted(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, AuthToken: "secret", HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		AuthToken: "secret",
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "host",
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
}

func TestHandshake_NoAuthToken_NoRestriction(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "host",
		},
	})
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
}

func TestHandshake_DuplicateWorker(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	req := &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "dup-worker",
			Hostname: "host",
		},
	}

	// First handshake
	resp, err := client.Handshake(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Accepted)

	// Second handshake (same worker ID) should also succeed (heartbeat update)
	resp, err = client.Handshake(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.Accepted)
}

// --- Compile ---

func TestCompile_EmptyTaskId(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	_, err := client.Compile(context.Background(), &pb.CompileRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestCompile_NoWorkersAvailable(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "task-1",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	// The coordinator returns a failed response, not a gRPC error
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	assert.Contains(t, resp.Stderr, "no worker available")
}

func TestCompile_TracksMetrics(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	// Start with 0 cache misses
	assert.Equal(t, int64(0), atomic.LoadInt64(&s.cacheMisses))

	// Compile triggers cache miss tracking even if no worker found
	ctx := context.Background()
	s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "task-metrics",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})

	assert.Equal(t, int64(1), atomic.LoadInt64(&s.cacheMisses))
}

// --- Build ---

func TestBuild_EmptyTaskId(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	_, err := client.Build(context.Background(), &pb.BuildRequest{})
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestBuild_NotImplemented(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Build(context.Background(), &pb.BuildRequest{
		TaskId:    "build-1",
		BuildType: pb.BuildType_BUILD_TYPE_CPP,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	assert.Contains(t, resp.Stderr, "not implemented")
}

func TestBuild_Flutter_NoWorker(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.Build(context.Background(), newFlutterBuildRequest("flutter-no-worker", "aabbccdd"))
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	assert.Contains(t, resp.Stderr, "no worker available")
}

func TestBuild_Flutter_CacheHit(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	req := newFlutterBuildRequest("flutter-cache-hit", "11223344")
	flutterConfig := req.GetFlutterConfig()
	cacheKey := cache.FlutterCacheKey(flutterConfig, req.SourceHash, flutterConfig.GetFlutterVersion())

	s.setFlutterCache(cacheKey, &pb.BuildResponse{
		Status:      pb.TaskStatus_STATUS_COMPLETED,
		ExitCode:    0,
		Stdout:      "cached build",
		Stderr:      "",
		Artifacts:   []byte("apk-bytes"),
		BuildTimeMs: 123,
		ArtifactList: []*pb.ArtifactInfo{
			{Name: "app-release.apk", SizeBytes: int64(len("apk-bytes"))},
		},
	})

	resp, err := client.Build(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_COMPLETED, resp.Status)
	assert.True(t, resp.FromCache)
	assert.Equal(t, "cached build", resp.Stdout)
	assert.Equal(t, []byte("apk-bytes"), resp.Artifacts)
	assert.Len(t, resp.ArtifactList, 1)
	assert.Equal(t, "app-release.apk", resp.ArtifactList[0].Name)
	assert.Equal(t, int64(1), atomic.LoadInt64(&s.cacheHits))
}

func TestBuild_Flutter_CacheMissForwardsAndCaches(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 2 * time.Second})
	defer cleanup()

	var buildCalls int64
	workerResp := &pb.BuildResponse{
		Status:      pb.TaskStatus_STATUS_COMPLETED,
		ExitCode:    0,
		Stdout:      "worker build",
		Stderr:      "",
		Artifacts:   []byte("worker-apk"),
		BuildTimeMs: 250,
		ArtifactList: []*pb.ArtifactInfo{
			{Name: "app-release.apk", SizeBytes: int64(len("worker-apk"))},
		},
	}

	addr, workerCleanup := setupTestWorker(t, func(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
		atomic.AddInt64(&buildCalls, 1)
		return workerResp, nil
	})
	defer workerCleanup()

	require.NoError(t, s.registry.Add(&registry.WorkerInfo{
		ID:      "flutter-worker-1",
		Address: addr,
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:    "flutter-worker-1",
			Hostname:    "flutter-host",
			CpuCores:    4,
			MemoryBytes: 8 * 1024 * 1024 * 1024,
			NativeArch:  pb.Architecture_ARCH_X86_64,
			Flutter: &pb.FlutterCapability{
				SdkVersion: "3.10.0",
				Platforms:  []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_ANDROID},
				AndroidSdk: true,
			},
		},
		MaxParallel: 4,
	}))

	req := newFlutterBuildRequest("flutter-cache-miss", "99aabbcc")
	resp, err := client.Build(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_COMPLETED, resp.Status)
	assert.False(t, resp.FromCache)
	assert.Equal(t, "worker build", resp.Stdout)
	assert.Equal(t, int64(1), atomic.LoadInt64(&buildCalls))

	cacheKey := cache.FlutterCacheKey(req.GetFlutterConfig(), req.SourceHash, req.GetFlutterConfig().GetFlutterVersion())
	cached := s.getFlutterCache(cacheKey)
	require.NotNil(t, cached)
	assert.Equal(t, "worker build", cached.stdout)
	assert.Equal(t, []byte("worker-apk"), cached.artifacts)

	req.TaskId = "flutter-cache-miss-2"
	resp, err = client.Build(context.Background(), req)
	require.NoError(t, err)
	assert.True(t, resp.FromCache)
	assert.Equal(t, int64(1), atomic.LoadInt64(&buildCalls))
}

// --- StreamBuild ---

func TestStreamBuild_NoMetadata(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	stream, err := client.StreamBuild(context.Background())
	require.NoError(t, err)

	// Send source without metadata first
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_SourceChunk{
			SourceChunk: []byte("int main() {}"),
		},
	})
	require.NoError(t, err)

	_, err = stream.CloseAndRecv()
	require.Error(t, err)
}

func TestStreamBuild_WithMetadata(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	stream, err := client.StreamBuild(context.Background())
	require.NoError(t, err)

	// Send metadata
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_Metadata{
			Metadata: &pb.BuildMetadata{
				TaskId:    "stream-1",
				BuildType: pb.BuildType_BUILD_TYPE_CPP,
			},
		},
	})
	require.NoError(t, err)

	// Send source
	err = stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_SourceChunk{
			SourceChunk: []byte("int main() { return 0; }"),
		},
	})
	require.NoError(t, err)

	resp, err := stream.CloseAndRecv()
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	assert.Contains(t, resp.Stderr, "not implemented")
}

// --- HealthCheck ---

func TestHealthCheck_NoWorkers(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.HealthCheck(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)

	// Healthy with 0 workers is true (idle coordinator)
	assert.True(t, resp.Healthy)
	assert.Equal(t, int32(0), resp.ActiveTasks)
	assert.Equal(t, int32(0), resp.QueuedTasks)
}

func TestHealthCheck_WithHealthyWorker(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 60 * time.Second})
	defer cleanup()

	// Register a worker
	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "h-worker",
			Hostname: "host",
		},
	})
	require.NoError(t, err)

	resp, err := client.HealthCheck(context.Background(), &pb.HealthRequest{})
	require.NoError(t, err)
	assert.True(t, resp.Healthy)
}

// --- GetWorkerStatus ---

func TestGetWorkerStatus_Empty(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.TotalWorkers)
	assert.Equal(t, int32(0), resp.HealthyWorkers)
	assert.Empty(t, resp.Workers)
}

func TestGetWorkerStatus_WithWorkers(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 60 * time.Second})
	defer cleanup()

	// Register workers
	for _, id := range []string{"w1", "w2"} {
		_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
			Capabilities: &pb.WorkerCapabilities{
				WorkerId:    id,
				Hostname:    id + "-host",
				CpuCores:    4,
				MemoryBytes: 8 * 1024 * 1024 * 1024,
				NativeArch:  pb.Architecture_ARCH_X86_64,
			},
		})
		require.NoError(t, err)
	}

	resp, err := client.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	require.NoError(t, err)
	assert.Equal(t, int32(2), resp.TotalWorkers)
	assert.Equal(t, int32(2), resp.HealthyWorkers)
	assert.Len(t, resp.Workers, 2)

	// Verify worker details
	found := map[string]bool{}
	for _, w := range resp.Workers {
		found[w.WorkerId] = true
		assert.Equal(t, int32(4), w.CpuCores)
		assert.Equal(t, int64(8*1024*1024*1024), w.MemoryBytes)
		assert.Equal(t, pb.Architecture_ARCH_X86_64, w.NativeArch)
	}
	assert.True(t, found["w1"])
	assert.True(t, found["w2"])
}

// --- GetWorkersForBuild ---

func TestGetWorkersForBuild_NoWorkers(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.GetWorkersForBuild(context.Background(), &pb.WorkersForBuildRequest{
		BuildType: pb.BuildType_BUILD_TYPE_CPP,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(0), resp.AvailableCount)
	assert.Empty(t, resp.WorkerIds)
}

func TestGetWorkersForBuild_WithCapableWorkers(t *testing.T) {
	_, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 60 * time.Second})
	defer cleanup()

	// Register a C++ worker
	_, err := client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:   "cpp-w",
			Hostname:   "host",
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc", "clang"},
			},
		},
	})
	require.NoError(t, err)

	// Register a Go-only worker
	_, err = client.Handshake(context.Background(), &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "go-w",
			Hostname: "host2",
			Go:       &pb.GoCapability{Version: "1.22"},
		},
	})
	require.NoError(t, err)

	resp, err := client.GetWorkersForBuild(context.Background(), &pb.WorkersForBuildRequest{
		BuildType: pb.BuildType_BUILD_TYPE_CPP,
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), resp.AvailableCount)
	assert.Contains(t, resp.WorkerIds, "cpp-w")
}

// --- ReportCacheHit ---

func TestReportCacheHit(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.ReportCacheHit(context.Background(), &pb.ReportCacheHitRequest{Hits: 5})
	require.NoError(t, err)
	assert.True(t, resp.Acknowledged)
	assert.Equal(t, int64(5), atomic.LoadInt64(&s.cacheHits))

	// Report more hits
	resp, err = client.ReportCacheHit(context.Background(), &pb.ReportCacheHitRequest{Hits: 3})
	require.NoError(t, err)
	assert.True(t, resp.Acknowledged)
	assert.Equal(t, int64(8), atomic.LoadInt64(&s.cacheHits))
}

func TestReportCacheHit_ZeroHits(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	resp, err := client.ReportCacheHit(context.Background(), &pb.ReportCacheHitRequest{Hits: 0})
	require.NoError(t, err)
	assert.True(t, resp.Acknowledged)
	assert.Equal(t, int64(0), atomic.LoadInt64(&s.cacheHits))
}

// --- EventNotifier ---

type mockEventNotifier struct {
	started   []*TaskEvent
	completed []*TaskEvent
}

func (m *mockEventNotifier) NotifyTaskStarted(event *TaskEvent) {
	m.started = append(m.started, event)
}
func (m *mockEventNotifier) NotifyTaskCompleted(event *TaskEvent) {
	m.completed = append(m.completed, event)
}

func TestSetEventNotifier(t *testing.T) {
	s := New(Config{HeartbeatTTL: 30 * time.Second})
	defer s.Stop()

	assert.Nil(t, s.eventNotifier)

	notifier := &mockEventNotifier{}
	s.SetEventNotifier(notifier)
	assert.NotNil(t, s.eventNotifier)
}

// --- Compile with event notifier (no worker, events on failure path) ---

func TestCompile_EventNotifier_NoWorker(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()

	notifier := &mockEventNotifier{}
	s.SetEventNotifier(notifier)

	// No workers registered; compile will fail at scheduler.Select
	resp, err := s.Compile(context.Background(), &pb.CompileRequest{
		TaskId:             "notified-task",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
	})
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)

	// No worker selected => no task started/completed events
	assert.Empty(t, notifier.started)
	assert.Empty(t, notifier.completed)
}

// --- Compile with a worker registered (worker unreachable) ---

func TestCompile_WorkerUnreachable(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{
		Port:           0,
		HeartbeatTTL:   60 * time.Second,
		RequestTimeout: 2 * time.Second,
	})
	defer cleanup()

	notifier := &mockEventNotifier{}
	s.SetEventNotifier(notifier)

	// Register a worker pointing to a non-existent address
	s.registry.Add(&registry.WorkerInfo{
		ID:      "unreachable-worker",
		Address: "127.0.0.1:19999",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc"},
			},
		},
		MaxParallel: 4,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := s.Compile(ctx, &pb.CompileRequest{
		TaskId:             "unreachable-task",
		PreprocessedSource: []byte("int main() {}"),
		Compiler:           "gcc",
		TargetArch:         pb.Architecture_ARCH_X86_64,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.TaskStatus_STATUS_FAILED, resp.Status)
	assert.Contains(t, resp.Stderr, "worker error")

	// Task events should fire
	assert.Len(t, notifier.started, 1)
	assert.Equal(t, "unreachable-task", notifier.started[0].ID)
	assert.Equal(t, "unreachable-worker", notifier.started[0].WorkerID)

	assert.Len(t, notifier.completed, 1)
	assert.Equal(t, "failed", notifier.completed[0].Status)
}
