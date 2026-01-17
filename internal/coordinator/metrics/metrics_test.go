package metrics

import (
	"sync"
	"testing"
)

func TestEWMA_Basic(t *testing.T) {
	e := NewEWMA(0.5)

	if e.IsInitialized() {
		t.Error("EWMA should not be initialized before any updates")
	}

	e.Update(100)
	if !e.IsInitialized() {
		t.Error("EWMA should be initialized after first update")
	}
	if e.Value() != 100 {
		t.Errorf("Value() = %f, want 100", e.Value())
	}

	e.Update(200)
	// With alpha=0.5: new = 0.5*200 + 0.5*100 = 150
	if e.Value() != 150 {
		t.Errorf("Value() = %f, want 150", e.Value())
	}

	e.Update(200)
	// new = 0.5*200 + 0.5*150 = 175
	if e.Value() != 175 {
		t.Errorf("Value() = %f, want 175", e.Value())
	}
}

func TestEWMA_HighAlpha(t *testing.T) {
	// High alpha = more weight to recent values
	e := NewEWMA(0.9)

	e.Update(100)
	e.Update(200)
	// new = 0.9*200 + 0.1*100 = 190
	if e.Value() != 190 {
		t.Errorf("Value() = %f, want 190", e.Value())
	}
}

func TestEWMA_LowAlpha(t *testing.T) {
	// Low alpha = more weight to historical values
	e := NewEWMA(0.1)

	e.Update(100)
	e.Update(200)
	// new = 0.1*200 + 0.9*100 = 110
	if e.Value() != 110 {
		t.Errorf("Value() = %f, want 110", e.Value())
	}
}

func TestEWMA_Reset(t *testing.T) {
	e := NewEWMA(0.5)
	e.Update(100)
	e.Reset()

	if e.IsInitialized() {
		t.Error("EWMA should not be initialized after reset")
	}
	if e.Value() != 0 {
		t.Errorf("Value() = %f, want 0 after reset", e.Value())
	}
}

func TestEWMA_Concurrent(t *testing.T) {
	e := NewEWMA(0.5)
	var wg sync.WaitGroup

	// Concurrent updates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val float64) {
			defer wg.Done()
			e.Update(val)
		}(float64(i))
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = e.Value()
		}()
	}

	wg.Wait()
	// Just checking it doesn't panic
}

func TestLatencyTracker_Basic(t *testing.T) {
	lt := NewLatencyTracker()

	// Unknown worker should return default
	latency := lt.Get("worker-1")
	if latency != DefaultLatencyMs {
		t.Errorf("Get() = %f, want default %f", latency, DefaultLatencyMs)
	}

	// Record some latencies
	lt.Record("worker-1", 50)
	if lt.Get("worker-1") != 50 {
		t.Errorf("Get() = %f, want 50", lt.Get("worker-1"))
	}

	lt.Record("worker-1", 100)
	// EWMA with alpha=0.5: 0.5*100 + 0.5*50 = 75
	if lt.Get("worker-1") != 75 {
		t.Errorf("Get() = %f, want 75", lt.Get("worker-1"))
	}
}

func TestLatencyTracker_MultipleWorkers(t *testing.T) {
	lt := NewLatencyTracker()

	lt.Record("worker-1", 50)
	lt.Record("worker-2", 100)
	lt.Record("worker-3", 150)

	if lt.Get("worker-1") != 50 {
		t.Errorf("worker-1 latency = %f, want 50", lt.Get("worker-1"))
	}
	if lt.Get("worker-2") != 100 {
		t.Errorf("worker-2 latency = %f, want 100", lt.Get("worker-2"))
	}
	if lt.Get("worker-3") != 150 {
		t.Errorf("worker-3 latency = %f, want 150", lt.Get("worker-3"))
	}
}

func TestLatencyTracker_Remove(t *testing.T) {
	lt := NewLatencyTracker()

	lt.Record("worker-1", 50)
	lt.Remove("worker-1")

	// After removal, should return default
	if lt.Get("worker-1") != DefaultLatencyMs {
		t.Errorf("Get() after Remove = %f, want default", lt.Get("worker-1"))
	}
}

func TestLatencyTracker_All(t *testing.T) {
	lt := NewLatencyTracker()

	lt.Record("worker-1", 50)
	lt.Record("worker-2", 100)

	all := lt.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d workers, want 2", len(all))
	}
	if all["worker-1"] != 50 {
		t.Errorf("All()[worker-1] = %f, want 50", all["worker-1"])
	}
	if all["worker-2"] != 100 {
		t.Errorf("All()[worker-2] = %f, want 100", all["worker-2"])
	}
}

func TestLatencyTracker_Reset(t *testing.T) {
	lt := NewLatencyTracker()

	lt.Record("worker-1", 50)
	lt.Record("worker-2", 100)
	lt.Reset()

	all := lt.All()
	if len(all) != 0 {
		t.Errorf("All() after Reset returned %d workers, want 0", len(all))
	}
}
