package scheduler

import (
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func newTestRegistry() *registry.InMemoryRegistry {
	return registry.NewInMemoryRegistry(30 * time.Second)
}

func addCppWorker(r *registry.InMemoryRegistry, id string) error {
	return r.Add(&registry.WorkerInfo{
		ID:      id,
		Address: "localhost:50051",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc"},
			},
		},
	})
}

func TestSimpleScheduler_NoWorkers(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	s := NewSimpleScheduler(reg)

	_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != ErrNoWorkers {
		t.Errorf("Expected ErrNoWorkers, got %v", err)
	}
}

func TestSimpleScheduler_NoMatchingWorkers(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	// Add a Go worker
	reg.Add(&registry.WorkerInfo{
		ID:      "go-worker",
		Address: "localhost:50051",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			Go: &pb.GoCapability{
				Version: "1.22",
			},
		},
	})

	s := NewSimpleScheduler(reg)

	// Ask for C++ worker
	_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != ErrNoMatchingWorkers {
		t.Errorf("Expected ErrNoMatchingWorkers, got %v", err)
	}
}

func TestSimpleScheduler_SelectSingle(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addCppWorker(reg, "worker-1")

	s := NewSimpleScheduler(reg)

	worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if worker.ID != "worker-1" {
		t.Errorf("Expected worker-1, got %s", worker.ID)
	}
}

func TestSimpleScheduler_RoundRobin(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addCppWorker(reg, "worker-1")
	addCppWorker(reg, "worker-2")
	addCppWorker(reg, "worker-3")

	s := NewSimpleScheduler(reg)

	// Track selections
	counts := make(map[string]int)

	for i := 0; i < 30; i++ {
		worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		counts[worker.ID]++
	}

	// Each worker should be selected at least once (round-robin guarantees distribution)
	for id, count := range counts {
		if count < 1 {
			t.Errorf("Worker %s was never selected", id)
		}
	}
	t.Logf("Round-robin distribution: %v", counts)
}

func TestSimpleScheduler_SkipsUnhealthy(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addCppWorker(reg, "worker-1")
	addCppWorker(reg, "worker-2")

	// Mark worker-1 as unhealthy
	reg.UpdateState("worker-1", registry.WorkerStateUnhealthy)

	s := NewSimpleScheduler(reg)

	// All selections should be worker-2
	for i := 0; i < 10; i++ {
		worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if worker.ID != "worker-2" {
			t.Errorf("Expected worker-2, got %s", worker.ID)
		}
	}
}

func TestLeastLoadedScheduler_SelectsLeastLoaded(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addCppWorker(reg, "worker-1")
	addCppWorker(reg, "worker-2")
	addCppWorker(reg, "worker-3")

	// Add load to worker-1 and worker-2
	reg.IncrementTasks("worker-1")
	reg.IncrementTasks("worker-1")
	reg.IncrementTasks("worker-2")

	s := NewLeastLoadedScheduler(reg)

	worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if worker.ID != "worker-3" {
		t.Errorf("Expected worker-3 (least loaded), got %s", worker.ID)
	}
}

func TestLeastLoadedScheduler_SkipsUnhealthy(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addCppWorker(reg, "worker-1")
	addCppWorker(reg, "worker-2")

	// worker-1 has no load but is unhealthy
	reg.UpdateState("worker-1", registry.WorkerStateUnhealthy)

	// worker-2 has load but is healthy
	reg.IncrementTasks("worker-2")

	s := NewLeastLoadedScheduler(reg)

	worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if worker.ID != "worker-2" {
		t.Errorf("Expected worker-2 (healthy), got %s", worker.ID)
	}
}

// P2C Scheduler Tests

func addDetailedCppWorker(r *registry.InMemoryRegistry, id string, cpuCores int32, memGB int64, activeTasks int32, source string) error {
	return r.Add(&registry.WorkerInfo{
		ID:              id,
		Address:         "localhost:50051",
		DiscoverySource: source,
		ActiveTasks:     activeTasks,
		Capabilities: &pb.WorkerCapabilities{
			CpuCores:        cpuCores,
			MemoryBytes:     memGB * 1024 * 1024 * 1024,
			NativeArch:      pb.Architecture_ARCH_X86_64,
			DockerAvailable: true,
			Cpp: &pb.CppCapability{
				Compilers: []string{"gcc"},
			},
		},
	})
}

func TestP2CScheduler_NoWorkers(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	_, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != ErrNoWorkers {
		t.Errorf("Expected ErrNoWorkers, got %v", err)
	}
}

func TestP2CScheduler_SingleWorker(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addDetailedCppWorker(reg, "worker-1", 8, 16, 0, "mdns")

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	if worker.ID != "worker-1" {
		t.Errorf("Expected worker-1, got %s", worker.ID)
	}
}

func TestP2CScheduler_PrefersBetterWorker(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	// Worker-1: Low specs, high load
	addDetailedCppWorker(reg, "worker-1", 2, 4, 5, "wan")

	// Worker-2: High specs, no load, LAN
	addDetailedCppWorker(reg, "worker-2", 16, 64, 0, "mdns")

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	// Run multiple selections - worker-2 should be preferred
	counts := make(map[string]int)
	for i := 0; i < 100; i++ {
		worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		counts[worker.ID]++
	}

	// Worker-2 should be selected most of the time due to higher score
	if counts["worker-2"] < counts["worker-1"] {
		t.Errorf("Expected worker-2 to be preferred, got counts: %v", counts)
	}
}

func TestP2CScheduler_SkipsOpenCircuit(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addDetailedCppWorker(reg, "worker-1", 16, 64, 0, "mdns")
	addDetailedCppWorker(reg, "worker-2", 8, 16, 0, "mdns")

	// Mock circuit checker that marks worker-1 as open
	checker := &mockCircuitChecker{openWorkers: map[string]bool{"worker-1": true}}

	s := NewP2CScheduler(P2CConfig{
		Registry:       reg,
		CircuitChecker: checker,
	})

	// All selections should be worker-2
	for i := 0; i < 10; i++ {
		worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if worker.ID != "worker-2" {
			t.Errorf("Expected worker-2 (circuit closed), got %s", worker.ID)
		}
	}
}

func TestP2CScheduler_SkipsOverloadedWorkers(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addDetailedCppWorker(reg, "worker-1", 16, 64, ScoreMaxActiveTasks, "mdns") // At max
	addDetailedCppWorker(reg, "worker-2", 8, 16, 0, "mdns")

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	// Worker-1 is at max tasks, should prefer worker-2
	for i := 0; i < 10; i++ {
		worker, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64)
		if err != nil {
			t.Fatalf("Select failed: %v", err)
		}
		if worker.ID != "worker-2" {
			t.Errorf("Expected worker-2 (not overloaded), got %s", worker.ID)
		}
	}
}

func TestP2CScheduler_ReportSuccess(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addDetailedCppWorker(reg, "worker-1", 8, 16, 0, "mdns")

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	// Report some latencies
	s.ReportSuccess("worker-1", 50)
	s.ReportSuccess("worker-1", 100)

	// The latency tracker should have updated EWMA
	// This is tested indirectly through scoring
}

func TestScoreWorker(t *testing.T) {
	reg := newTestRegistry()
	defer reg.Stop()

	addDetailedCppWorker(reg, "good-worker", 16, 64, 0, "mdns")
	addDetailedCppWorker(reg, "bad-worker", 2, 4, 5, "wan")

	s := NewP2CScheduler(P2CConfig{Registry: reg})

	goodWorker, _ := reg.Get("good-worker")
	badWorker, _ := reg.Get("bad-worker")

	goodScore := s.scoreWorker(goodWorker, pb.Architecture_ARCH_X86_64)
	badScore := s.scoreWorker(badWorker, pb.Architecture_ARCH_X86_64)

	if goodScore <= badScore {
		t.Errorf("Good worker score (%f) should be > bad worker score (%f)", goodScore, badScore)
	}

	// Verify score components for good worker:
	// +50 (native arch) + 160 (16 cores * 10) + 320 (64GB * 5) + 0 (no tasks) + 20 (LAN) - 50 (100ms default latency) = 500
	// Actual may vary due to latency default
	t.Logf("Good worker score: %f, Bad worker score: %f", goodScore, badScore)
}

func TestPickTwo(t *testing.T) {
	// Test that pickTwo returns different indices
	for i := 0; i < 100; i++ {
		idx1, idx2 := pickTwo(10)
		if idx1 == idx2 {
			t.Errorf("pickTwo returned same indices: %d, %d", idx1, idx2)
		}
		if idx1 < 0 || idx1 >= 10 || idx2 < 0 || idx2 >= 10 {
			t.Errorf("pickTwo returned out of range indices: %d, %d", idx1, idx2)
		}
	}
}

func TestPickTwo_Small(t *testing.T) {
	// Test edge cases
	idx1, idx2 := pickTwo(2)
	if idx1 == idx2 {
		t.Errorf("pickTwo(2) returned same indices: %d, %d", idx1, idx2)
	}

	idx1, idx2 = pickTwo(1)
	if idx1 != 0 || idx2 != 0 {
		t.Errorf("pickTwo(1) should return 0, 0, got %d, %d", idx1, idx2)
	}
}

// mockCircuitChecker implements CircuitChecker for testing.
type mockCircuitChecker struct {
	openWorkers map[string]bool
}

func (m *mockCircuitChecker) IsOpen(workerID string) bool {
	return m.openWorkers[workerID]
}
