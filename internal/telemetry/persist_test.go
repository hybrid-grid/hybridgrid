package telemetry

import (
	"context"
	"path/filepath"
	"testing"
)

func TestPersistenceRoundtrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history.json")

	s1 := NewStore()
	s1.SetScheduler("hybrid-linucb-d", "alpha=0.5 discount=0.98 discount-mode=arm")
	s1.RecordTask(&TaskInfo{ID: "t1", BuildID: "b1", Status: "completed", WorkerID: "w1", StartedAtMs: 10, CompletedAtMs: 20})
	s1.SetScheduler("leastloaded", "")
	s1.RecordTask(&TaskInfo{ID: "t2", BuildID: "b2", Status: "completed", WorkerID: "w2", StartedAtMs: 30, CompletedAtMs: 40})
	if err := s1.writeSnapshot(path); err != nil {
		t.Fatalf("writeSnapshot: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	s2 := NewStore()
	if err := s2.EnablePersistence(ctx, path); err != nil {
		t.Fatalf("EnablePersistence: %v", err)
	}
	// Stop the loop before TempDir cleanup: the final flush writes into
	// the same directory and Windows cannot unlink freshly recreated
	// files mid-RemoveAll.
	cancel()
	s2.WaitPersistLoop()

	builds := s2.Builds()
	if len(builds) != 2 {
		t.Fatalf("Builds() = %d builds, want 2", len(builds))
	}
	byID := map[string]*BuildInfo{}
	for _, b := range builds {
		byID[b.ID] = b
	}
	if got := byID["b1"].Scheduler; got != "hybrid-linucb-d" {
		t.Errorf("b1 scheduler = %q, want hybrid-linucb-d", got)
	}
	if got := byID["b1"].SchedulerParams; got != "alpha=0.5 discount=0.98 discount-mode=arm" {
		t.Errorf("b1 scheduler_params = %q", got)
	}
	if got := byID["b2"].Scheduler; got != "leastloaded" {
		t.Errorf("b2 scheduler = %q, want leastloaded", got)
	}
	// Task detail must survive so per-worker distribution stays readable.
	if _, tasks, ok := s2.BuildDetail("b1"); !ok || len(tasks) != 1 || tasks[0].WorkerID != "w1" {
		t.Fatalf("BuildDetail(b1) = ok=%v tasks=%d, want 1 task on w1", ok, len(tasks))
	}
}

func TestBuildsStampedWithActiveScheduler(t *testing.T) {
	s := NewStore()
	s.RecordTask(&TaskInfo{ID: "t0", BuildID: "pre", Status: "completed"})
	s.SetScheduler("sed", "")
	s.RecordTask(&TaskInfo{ID: "t1", BuildID: "post", Status: "completed"})

	byID := map[string]*BuildInfo{}
	for _, b := range s.Builds() {
		byID[b.ID] = b
	}
	if byID["pre"].Scheduler != "" {
		t.Errorf("pre scheduler = %q, want empty", byID["pre"].Scheduler)
	}
	if byID["post"].Scheduler != "sed" {
		t.Errorf("post scheduler = %q, want sed", byID["post"].Scheduler)
	}
}
