package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

func newTestMetrics() (*Metrics, *prometheus.Registry) {
	m := New()
	reg := prometheus.NewRegistry()
	m.Register(reg)
	return m, reg
}

func TestMetrics_New(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New returned nil")
	}
	if m.TasksTotal == nil {
		t.Error("TasksTotal is nil")
	}
	if m.CacheHits == nil {
		t.Error("CacheHits is nil")
	}
	if m.WorkersTotal == nil {
		t.Error("WorkersTotal is nil")
	}
}

func TestMetrics_RecordTaskComplete(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordTaskComplete(TaskStatusSuccess, "cpp", "worker-1", 1.5)
	m.RecordTaskComplete(TaskStatusError, "cpp", "worker-2", 0.5)
	m.RecordTaskComplete(TaskStatusSuccess, "go", "worker-1", 2.0)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_tasks_total" {
			found = true
			if len(mf.GetMetric()) != 3 {
				t.Errorf("Expected 3 metrics, got %d", len(mf.GetMetric()))
			}
		}
	}
	if !found {
		t.Error("hybridgrid_tasks_total metric not found")
	}
}

func TestMetrics_RecordCacheHitMiss(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheMiss()

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	hitCount, missCount := 0.0, 0.0
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_cache_hits_total" {
			hitCount = mf.GetMetric()[0].GetCounter().GetValue()
		}
		if mf.GetName() == "hybridgrid_cache_misses_total" {
			missCount = mf.GetMetric()[0].GetCounter().GetValue()
		}
	}

	if hitCount != 2 {
		t.Errorf("Cache hits = %f, want 2", hitCount)
	}
	if missCount != 1 {
		t.Errorf("Cache misses = %f, want 1", missCount)
	}
}

func TestMetrics_RecordFallback(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordFallback("no_workers")
	m.RecordFallback("circuit_open")
	m.RecordFallback("no_workers")

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_fallbacks_total" {
			found = true
			if len(mf.GetMetric()) != 2 {
				t.Errorf("Expected 2 metric series, got %d", len(mf.GetMetric()))
			}
		}
	}
	if !found {
		t.Error("hybridgrid_fallbacks_total metric not found")
	}
}

func TestMetrics_WorkerGauges(t *testing.T) {
	m, reg := newTestMetrics()

	m.SetWorkerCount("healthy", "mdns", 3)
	m.SetWorkerCount("unhealthy", "mdns", 1)
	m.SetActiveTaskCount("worker-1", 5)
	m.SetQueueDepth(10)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	for _, mf := range mfs {
		switch mf.GetName() {
		case "hybridgrid_workers_total":
			if len(mf.GetMetric()) != 2 {
				t.Errorf("workers_total: expected 2 series, got %d", len(mf.GetMetric()))
			}
		case "hybridgrid_active_tasks":
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 5 {
				t.Errorf("active_tasks = %f, want 5", val)
			}
		case "hybridgrid_queue_depth":
			val := mf.GetMetric()[0].GetGauge().GetValue()
			if val != 10 {
				t.Errorf("queue_depth = %f, want 10", val)
			}
		}
	}
}

func TestMetrics_CircuitState(t *testing.T) {
	m, reg := newTestMetrics()

	m.SetCircuitState("worker-1", CircuitStateClosed)
	m.SetCircuitState("worker-2", CircuitStateOpen)
	m.SetCircuitState("worker-3", CircuitStateHalfOpen)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_circuit_state" {
			found = true
			if len(mf.GetMetric()) != 3 {
				t.Errorf("Expected 3 workers, got %d", len(mf.GetMetric()))
			}
		}
	}
	if !found {
		t.Error("hybridgrid_circuit_state metric not found")
	}
}

func TestMetrics_RecordTransfer(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordTransfer("upload", 1024)
	m.RecordTransfer("download", 4096)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_network_transfer_bytes" {
			found = true
		}
	}
	if !found {
		t.Error("hybridgrid_network_transfer_bytes metric not found")
	}
}

func TestMetrics_RecordWorkerLatency(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordWorkerLatency("worker-1", 50)
	m.RecordWorkerLatency("worker-1", 75)
	m.RecordWorkerLatency("worker-2", 100)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_worker_latency_ms" {
			found = true
		}
	}
	if !found {
		t.Error("hybridgrid_worker_latency_ms metric not found")
	}
}

func TestMetrics_RemoveWorkerMetrics(t *testing.T) {
	m, reg := newTestMetrics()

	// Add metrics for worker-1
	m.SetActiveTaskCount("worker-1", 5)
	m.SetCircuitState("worker-1", CircuitStateClosed)
	m.RecordWorkerLatency("worker-1", 50)

	// Remove worker metrics
	m.RemoveWorkerMetrics("worker-1")

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	// After removal, these metrics should be empty for worker-1
	for _, mf := range mfs {
		switch mf.GetName() {
		case "hybridgrid_active_tasks", "hybridgrid_circuit_state":
			if len(mf.GetMetric()) > 0 {
				t.Errorf("%s should have no metrics after removal", mf.GetName())
			}
		}
	}
}

func TestMetrics_Handler(t *testing.T) {
	// Create a fresh registry for this test
	reg := prometheus.NewRegistry()
	m := New()
	m.Register(reg)

	// Record some metrics
	m.RecordTaskComplete(TaskStatusSuccess, "cpp", "worker-1", 1.0)
	m.RecordCacheHit()

	// Gather metrics
	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	// Check for expected metrics
	foundTasks := false
	foundCacheHits := false
	for _, mf := range mfs {
		switch mf.GetName() {
		case "hybridgrid_tasks_total":
			foundTasks = true
		case "hybridgrid_cache_hits_total":
			foundCacheHits = true
		}
	}

	if !foundTasks {
		t.Error("Missing hybridgrid_tasks_total metric")
	}
	if !foundCacheHits {
		t.Error("Missing hybridgrid_cache_hits_total metric")
	}

	// Test the actual HTTP handler
	handler := Handler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}
}

func TestMetrics_TaskDurationBuckets(t *testing.T) {
	m, reg := newTestMetrics()

	// Record tasks with various durations
	durations := []float64{0.05, 0.3, 0.8, 3.0, 15.0, 45.0}
	for _, d := range durations {
		m.RecordTaskComplete(TaskStatusSuccess, "cpp", "worker-1", d)
	}

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_task_duration_seconds" {
			histogram := mf.GetMetric()[0].GetHistogram()
			if histogram.GetSampleCount() != uint64(len(durations)) {
				t.Errorf("Sample count = %d, want %d", histogram.GetSampleCount(), len(durations))
			}
		}
	}
}

func TestMetrics_QueueTime(t *testing.T) {
	m, reg := newTestMetrics()

	m.RecordQueueTime("cpp", 0.05)
	m.RecordQueueTime("go", 0.1)

	mfs, err := reg.Gather()
	if err != nil {
		t.Fatalf("Gather failed: %v", err)
	}

	found := false
	for _, mf := range mfs {
		if mf.GetName() == "hybridgrid_queue_time_seconds" {
			found = true
		}
	}
	if !found {
		t.Error("hybridgrid_queue_time_seconds metric not found")
	}
}
