package server

import (
	"context"
	"net"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func addFlutterWorker(t *testing.T, s *Server, id, addr string) {
	t.Helper()
	if err := s.registry.Add(&registry.WorkerInfo{
		ID: id, Address: addr, MaxParallel: 4,
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: id, CpuCores: 4, NativeArch: pb.Architecture_ARCH_X86_64,
			Flutter: &pb.FlutterCapability{Platforms: []pb.TargetPlatform{pb.TargetPlatform_PLATFORM_ANDROID}},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

// deadAddr is a loopback address nothing listens on.
func deadAddr(t *testing.T) string {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := lis.Addr().String()
	lis.Close()
	return addr
}

func TestBuild_DeadWorkerFailsFast(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 2 * time.Second})
	defer cleanup()
	addFlutterWorker(t, s, "dead", deadAddr(t))

	// The client allows a long build; a dead worker must not use it up.
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	start := time.Now()
	resp, err := client.Build(ctx, newFlutterBuildRequest("dead-1", "deadbeef"))
	took := time.Since(start)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if resp.Status != pb.TaskStatus_STATUS_FAILED {
		t.Fatalf("status = %v, want FAILED", resp.Status)
	}
	if took > 3*workerDialTimeout {
		t.Fatalf("dead worker held the build for %v", took)
	}
	w, _ := s.registry.Get("dead")
	if w.State != registry.WorkerStateUnhealthy || w.ActiveTasks != 0 {
		t.Fatalf("dead worker state=%v active=%d, want unhealthy with no booked slot", w.State, w.ActiveTasks)
	}
	// Its next heartbeat brings it back.
	if err := s.registry.UpdateHeartbeat("dead"); err != nil {
		t.Fatal(err)
	}
	if w, _ := s.registry.Get("dead"); w.State == registry.WorkerStateUnhealthy {
		t.Fatal("heartbeat did not clear the unhealthy mark")
	}
}

func TestConnectBuildWorker_FailsOverToLiveWorker(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()
	s.workerConns.dialTimeout = time.Second

	liveAddr, stop := setupTestWorker(t, func(context.Context, *pb.BuildRequest) (*pb.BuildResponse, error) {
		return &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, nil
	})
	defer stop()
	addFlutterWorker(t, s, "dead", deadAddr(t))
	addFlutterWorker(t, s, "live", liveAddr)

	// selectBuildWorker picked the dead worker and booked a slot on it.
	if err := s.registry.IncrementTasks("dead"); err != nil {
		t.Fatal(err)
	}
	dead, _ := s.registry.Get("dead")
	onlyLive := func(w *registry.WorkerInfo) bool { return w.ID == "live" }
	meta := buildTaskMeta{taskID: "t1", buildType: "flutter"}

	got, conn, failed := s.connectBuildWorker(context.Background(), meta, time.Now(), dead, onlyLive, "flutter")
	if failed != nil || conn == nil {
		t.Fatalf("failover failed: %v", failed)
	}
	if got.ID != "live" {
		t.Fatalf("connected to %s, want live", got.ID)
	}
	if w, _ := s.registry.Get("dead"); w.State != registry.WorkerStateUnhealthy || w.ActiveTasks != 0 {
		t.Fatalf("dead worker state=%v active=%d", w.State, w.ActiveTasks)
	}
	if w, _ := s.registry.Get("live"); w.ActiveTasks != 1 {
		t.Fatalf("live worker active=%d, want the slot booked by failover", w.ActiveTasks)
	}
}

func TestGetWorkerStatus_ReportsPerWorkerHealth(t *testing.T) {
	s, _, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()
	addFlutterWorker(t, s, "alive", "127.0.0.1:1")
	addFlutterWorker(t, s, "gone", "127.0.0.1:2")
	if err := s.registry.UpdateState("gone", registry.WorkerStateUnhealthy); err != nil {
		t.Fatal(err)
	}

	resp, err := s.GetWorkerStatus(context.Background(), &pb.WorkerStatusRequest{})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]bool{}
	for _, w := range resp.Workers {
		if w.Healthy == nil {
			t.Fatalf("%s: healthy not set", w.WorkerId)
		}
		if w.CircuitState == "" {
			t.Fatalf("%s: circuit state not set", w.WorkerId)
		}
		got[w.WorkerId] = *w.Healthy
	}
	if !got["alive"] || got["gone"] || resp.HealthyWorkers != 1 {
		t.Fatalf("health = %v, healthy count %d", got, resp.HealthyWorkers)
	}
}
