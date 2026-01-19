package registry

import (
	"sync"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func newTestRegistry() *InMemoryRegistry {
	return NewInMemoryRegistry(30 * time.Second)
}

func TestAdd(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	worker := &WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
		Capabilities: &pb.WorkerCapabilities{
			Hostname: "test-host",
			CpuCores: 4,
		},
	}

	err := r.Add(worker)
	if err != nil {
		t.Fatalf("Add failed: %v", err)
	}

	if r.Count() != 1 {
		t.Errorf("Expected count 1, got %d", r.Count())
	}

	// Adding same worker again should succeed (update/heartbeat behavior)
	err = r.Add(worker)
	if err != nil {
		t.Errorf("Re-adding worker should succeed (heartbeat): %v", err)
	}

	// Count should still be 1 (update, not duplicate)
	if r.Count() != 1 {
		t.Errorf("Expected count 1 after re-add, got %d", r.Count())
	}
}

func TestRemove(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	worker := &WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
	}

	r.Add(worker)

	err := r.Remove("worker-1")
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if r.Count() != 0 {
		t.Errorf("Expected count 0, got %d", r.Count())
	}

	// Try to remove non-existent
	err = r.Remove("worker-1")
	if err == nil {
		t.Error("Expected error for non-existent worker")
	}
}

func TestGet(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	worker := &WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
	}

	r.Add(worker)

	got, ok := r.Get("worker-1")
	if !ok {
		t.Fatal("Expected to find worker")
	}

	if got.ID != "worker-1" {
		t.Errorf("Expected ID 'worker-1', got '%s'", got.ID)
	}

	// Test not found
	_, ok = r.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent worker")
	}
}

func TestList(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	for i := 0; i < 3; i++ {
		r.Add(&WorkerInfo{
			ID:      string(rune('a' + i)),
			Address: "localhost:50051",
		})
	}

	list := r.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 workers, got %d", len(list))
	}
}

func TestListByCapability(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	// Add C++ capable worker
	r.Add(&WorkerInfo{
		ID:      "cpp-worker",
		Address: "localhost:50051",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc", "clang"},
			},
		},
	})

	// Add Go capable worker
	r.Add(&WorkerInfo{
		ID:      "go-worker",
		Address: "localhost:50052",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Go: &pb.GoCapability{
				Version: "1.22.0",
			},
		},
	})

	// Query for C++ workers
	cppWorkers := r.ListByCapability(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if len(cppWorkers) != 1 {
		t.Errorf("Expected 1 C++ worker, got %d", len(cppWorkers))
	}

	if len(cppWorkers) > 0 && cppWorkers[0].ID != "cpp-worker" {
		t.Errorf("Expected cpp-worker, got %s", cppWorkers[0].ID)
	}

	// Query for Go workers
	goWorkers := r.ListByCapability(pb.BuildType_BUILD_TYPE_GO, pb.Architecture_ARCH_UNSPECIFIED)
	if len(goWorkers) != 1 {
		t.Errorf("Expected 1 Go worker, got %d", len(goWorkers))
	}
}

func TestUpdateState(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	r.Add(&WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
	})

	err := r.UpdateState("worker-1", WorkerStateBusy)
	if err != nil {
		t.Fatalf("UpdateState failed: %v", err)
	}

	got, _ := r.Get("worker-1")
	if got.State != WorkerStateBusy {
		t.Errorf("Expected state Busy, got %s", got.State)
	}

	// Test non-existent
	err = r.UpdateState("nonexistent", WorkerStateBusy)
	if err == nil {
		t.Error("Expected error for non-existent worker")
	}
}

func TestUpdateHeartbeat(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	r.Add(&WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
	})

	// Mark as unhealthy
	r.UpdateState("worker-1", WorkerStateUnhealthy)

	// Heartbeat should restore to idle
	err := r.UpdateHeartbeat("worker-1")
	if err != nil {
		t.Fatalf("UpdateHeartbeat failed: %v", err)
	}

	got, _ := r.Get("worker-1")
	if got.State != WorkerStateIdle {
		t.Errorf("Expected state Idle after heartbeat, got %s", got.State)
	}
}

func TestTaskTracking(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	r.Add(&WorkerInfo{
		ID:      "worker-1",
		Address: "localhost:50051",
	})

	// Start task
	err := r.IncrementTasks("worker-1")
	if err != nil {
		t.Fatalf("IncrementTasks failed: %v", err)
	}

	got, _ := r.Get("worker-1")
	if got.ActiveTasks != 1 {
		t.Errorf("Expected 1 active task, got %d", got.ActiveTasks)
	}
	if got.State != WorkerStateBusy {
		t.Errorf("Expected state Busy, got %s", got.State)
	}

	// Complete task
	err = r.DecrementTasks("worker-1", true, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("DecrementTasks failed: %v", err)
	}

	got, _ = r.Get("worker-1")
	if got.ActiveTasks != 0 {
		t.Errorf("Expected 0 active tasks, got %d", got.ActiveTasks)
	}
	if got.SuccessfulTasks != 1 {
		t.Errorf("Expected 1 successful task, got %d", got.SuccessfulTasks)
	}
	if got.State != WorkerStateIdle {
		t.Errorf("Expected state Idle, got %s", got.State)
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := newTestRegistry()
	defer r.Stop()

	// Add some workers
	for i := 0; i < 10; i++ {
		r.Add(&WorkerInfo{
			ID:      string(rune('a' + i)),
			Address: "localhost:50051",
		})
	}

	var wg sync.WaitGroup
	concurrency := 100

	// Concurrent reads
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			r.List()
			r.Get("a")
			r.Count()
		}()
	}

	// Concurrent writes
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func(n int) {
			defer wg.Done()
			r.UpdateHeartbeat(string(rune('a' + (n % 10))))
			r.UpdateState(string(rune('a'+(n%10))), WorkerStateIdle)
		}(i)
	}

	wg.Wait()

	// Should not panic or race
	if r.Count() != 10 {
		t.Errorf("Expected 10 workers, got %d", r.Count())
	}
}

func TestWorkerInfoIsHealthy(t *testing.T) {
	ttl := 30 * time.Second

	tests := []struct {
		name     string
		worker   *WorkerInfo
		expected bool
	}{
		{
			name: "healthy",
			worker: &WorkerInfo{
				State:         WorkerStateIdle,
				LastHeartbeat: time.Now(),
			},
			expected: true,
		},
		{
			name: "unhealthy state",
			worker: &WorkerInfo{
				State:         WorkerStateUnhealthy,
				LastHeartbeat: time.Now(),
			},
			expected: false,
		},
		{
			name: "stale heartbeat",
			worker: &WorkerInfo{
				State:         WorkerStateIdle,
				LastHeartbeat: time.Now().Add(-60 * time.Second),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.worker.IsHealthy(ttl)
			if got != tt.expected {
				t.Errorf("IsHealthy() = %v, want %v", got, tt.expected)
			}
		})
	}
}
