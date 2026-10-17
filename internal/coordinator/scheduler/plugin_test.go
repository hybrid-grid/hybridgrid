package scheduler

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// fakePlugin is an in-process scheduler plugin used as the external
// endpoint in tests.
type fakePlugin struct {
	pb.UnimplementedSchedulerPluginServer
	selectFn func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error)
	feedFn   func(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error)
}

func (f *fakePlugin) Select(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
	return f.selectFn(ctx, req)
}

func (f *fakePlugin) Feedback(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error) {
	return f.feedFn(ctx, req)
}

// startPlugin launches a fake plugin on a random port and returns its
// address plus a stop function.
func startPlugin(t *testing.T, p *fakePlugin) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterSchedulerPluginServer(gs, p)
	go gs.Serve(lis)
	return lis.Addr().String(), gs.Stop
}

func TestPluginScheduler_SelectUsesPlugin(t *testing.T) {
	// Registry with two workers; the plugin will pick worker-b.
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-b", Address: "127.0.0.1:2",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 2,
	}); err != nil {
		t.Fatal(err)
	}

	var sawWorkers int32Shim
	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			sawWorkers.n = len(req.GetWorkers())
			return &pb.SelectResponse{WorkerId: "worker-b"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg, Insecure: true,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w.ID != "worker-b" {
		t.Fatalf("expected worker-b (plugin's answer), got %s", w.ID)
	}
	if sawWorkers.n != 2 {
		t.Fatalf("plugin should see 2 workers, saw %d", sawWorkers.n)
	}
}

func TestPluginScheduler_FallsBackWhenUnreachable(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	// Port 1 is never listening; the constructor must still succeed
	// fast (dial timeout) — actually it fails to dial: assert the
	// documented behavior.
	_, err := NewPluginScheduler(PluginConfig{
		Address: "127.0.0.1:1", Inner: NewLeastLoadedScheduler(reg), Registry: reg, Insecure: true,
	})
	if err == nil {
		t.Fatal("expected dial error for unreachable plugin")
	}
}

func TestPluginScheduler_FallsBackOnTimeout(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			time.Sleep(300 * time.Millisecond) // far beyond the timeout
			return &pb.SelectResponse{WorkerId: "worker-a"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Timeout: 30 * time.Millisecond,
		Inner: NewLeastLoadedScheduler(reg), Registry: reg, Insecure: true,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	start := time.Now()
	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w == nil || w.ID != "worker-a" {
		t.Fatalf("fallback should pick worker-a, got %+v", w)
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("timeout fallback took too long: %v", elapsed)
	}
}

func TestPluginScheduler_FallsBackOnUnknownWorker(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{WorkerId: "ghost-worker"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg, Insecure: true,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w == nil || w.ID != "worker-a" {
		t.Fatalf("unknown-worker answer must fall back to worker-a, got %+v", w)
	}
}

func TestPluginScheduler_FeedbackNeverBlocks(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	release := make(chan struct{})
	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{}, nil
		},
		feedFn: func(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error) {
			<-release // block every feedback call
			return &pb.FeedbackResponse{}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg, Insecure: true,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	// RecordOutcome must return promptly even with the plugin stuck.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < feedbackBuffer*2; i++ {
			s.RecordOutcome("worker-a", -10, true, TaskContext{})
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RecordOutcome blocked on a stuck plugin")
	}
	close(release)
}

// int32Shim is a tiny mutable int for capturing plugin observations.
type int32Shim struct{ n int }
