package server

import (
	"context"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/blobcache"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

func tscRequest(taskID string) *pb.BuildRequest {
	return &pb.BuildRequest{
		TaskId:    taskID,
		BuildType: pb.BuildType_BUILD_TYPE_NODEJS,
		DepsHash:  "abc123",
		Config:    &pb.BuildRequest_NodeConfig{NodeConfig: &pb.NodeConfig{TscProject: "packages/lib/tsconfig.json"}},
	}
}

func TestTscBuild_DependencyLayerTravelsOncePerMachine(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 5 * time.Second})
	defer cleanup()
	store, err := blobcache.New(blobcache.Config{Dir: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	s.SetDepsStore(store)

	var (
		mu       sync.Mutex
		received []int // layer bytes each worker call carried
		lost     bool  // the worker's cache was evicted
	)
	addr, stop := setupTestWorker(t, func(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
		mu.Lock()
		defer mu.Unlock()
		received = append(received, len(req.DepsArchive))
		if lost && len(req.DepsArchive) == 0 {
			return nil, status.Error(codes.FailedPrecondition, depslayer.MissingMessage)
		}
		lost = false
		return &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, nil
	})
	defer stop()
	if err := s.registry.Add(&registry.WorkerInfo{
		ID: "node-1", Address: addr, MaxParallel: 4,
		Capabilities: &pb.WorkerCapabilities{WorkerId: "node-1", Nodejs: &pb.NodeCapability{}},
	}); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()

	// 1. Named by hash, unknown to the coordinator: the client must resend.
	if _, err := client.Build(ctx, tscRequest("t1")); status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("unknown layer: %v", err)
	}
	// 2. Resent with the layer: stored, and forwarded to the worker.
	req := tscRequest("t1")
	req.DepsArchive = []byte("layer-bytes")
	resp, err := client.Build(ctx, req)
	if err != nil || resp.Status != pb.TaskStatus_STATUS_COMPLETED || resp.WorkerId != "node-1" {
		t.Fatalf("with layer: %v %+v", err, resp)
	}
	// 3. Later tasks name it only; the worker already holds it.
	if _, err := client.Build(ctx, tscRequest("t2")); err != nil {
		t.Fatal(err)
	}
	// 4. The worker lost it: the coordinator resends from its store.
	mu.Lock()
	lost = true
	mu.Unlock()
	if resp, err := client.Build(ctx, tscRequest("t3")); err != nil || resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("after eviction: %v %+v", err, resp)
	}

	mu.Lock()
	defer mu.Unlock()
	want := []int{len("layer-bytes"), 0, 0, len("layer-bytes")}
	if len(received) != len(want) {
		t.Fatalf("worker calls = %v, want %v", received, want)
	}
	for i := range want {
		if received[i] != want[i] {
			t.Fatalf("worker calls = %v, want %v", received, want)
		}
	}
}

func TestTscBuild_NeedsANodeWorker(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, DispatchQueueTimeout: 0})
	defer cleanup()
	addFlutterWorker(t, s, "flutter-only", "127.0.0.1:1")
	req := tscRequest("t1")
	req.DepsHash = ""
	resp, err := client.Build(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED (no node worker)", resp.Status)
	}
}
