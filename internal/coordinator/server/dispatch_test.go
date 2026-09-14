package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/scheduler"
)

// TestDispatch_NeverOverbooksUnderConcurrency is a regression test for the
// race documented in docs/thesis/hybrid-linucb-implementation.md
// (ResourceExhausted blocker) and PR hybrid-grid/hybridgrid#9: concurrent
// Compile() handlers used to call
// scheduler.SelectWith (read ActiveTasks) and registry.IncrementTasks
// (write it) as two separate, unsynchronized steps, so multiple goroutines
// racing on the same tightly-capacity-limited worker could all read the
// same stale "idle" snapshot and all pick it — overbooking a worker beyond
// its MaxParallel purely from a coordinator-side bookkeeping race, not
// genuine cluster-wide capacity exhaustion.
//
// This drives many concurrent dispatch()+release cycles against workers
// with MaxParallel=1 each and just enough total capacity (workers == max
// concurrent callers) that a correct scheduler should NEVER need to
// overbook any one of them. It tracks, per worker, how many callers
// currently hold that worker's slot (an independent atomic counter, not
// the registry's own bookkeeping) and fails if that count ever exceeds
// MaxParallel — the exact invariant the race used to violate.
func TestDispatch_NeverOverbooksUnderConcurrency(t *testing.T) {
	const numWorkers = 5
	const numGoroutines = 5 // == numWorkers: correct dispatch never needs to overbook
	const waves = 40        // repeat many times to surface a timing-dependent race

	s := New(Config{SchedulerType: "leastloaded", HeartbeatTTL: 30 * time.Second})
	defer s.Stop()

	workerIDs := make([]string, numWorkers)
	for i := 0; i < numWorkers; i++ {
		id := "worker-" + string(rune('A'+i))
		workerIDs[i] = id
		require := func(err error) {
			if err != nil {
				t.Fatalf("failed to register %s: %v", id, err)
			}
		}
		require(s.registry.Add(&registry.WorkerInfo{
			ID:      id,
			Address: "127.0.0.1:0",
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
			MaxParallel: 1,
			State:       registry.WorkerStateIdle,
		}))
	}

	held := make(map[string]*int64, numWorkers)
	for _, id := range workerIDs {
		held[id] = new(int64)
	}
	var mu sync.Mutex // guards nothing but keeps `held` map access simple/obviously-safe
	getCounter := func(id string) *int64 {
		mu.Lock()
		defer mu.Unlock()
		return held[id]
	}

	var overbooked int64
	for wave := 0; wave < waves; wave++ {
		var wg sync.WaitGroup
		wg.Add(numGoroutines)
		for g := 0; g < numGoroutines; g++ {
			go func() {
				defer wg.Done()
				worker, _, _, err := s.dispatch(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", scheduler.TaskContext{})
				if err != nil {
					t.Errorf("dispatch failed: %v", err)
					return
				}
				counter := getCounter(worker.ID)
				n := atomic.AddInt64(counter, 1)
				if n > 1 {
					atomic.AddInt64(&overbooked, 1)
				}
				// Simulate the task finishing quickly and releasing the
				// slot, mirroring the real Compile() flow's eventual
				// DecrementTasks after the worker responds.
				time.Sleep(time.Microsecond)
				atomic.AddInt64(counter, -1)
				s.registry.DecrementTasks(worker.ID, true, 0)
			}()
		}
		wg.Wait()
	}

	if overbooked > 0 {
		t.Errorf("worker was concurrently dispatched beyond MaxParallel=1 in %d cases out of %d waves — dispatch is not serialized correctly", overbooked, waves)
	}
}

// TestDispatch_ConcurrentCallersGetDistinctWorkers is a sanity companion to
// the overbooking test above: with exactly as many concurrent callers as
// workers (each MaxParallel=1), a correctly serialized dispatcher should
// assign every caller a DIFFERENT worker in each wave, not just avoid
// overbooking a single one while leaving others idle.
func TestDispatch_ConcurrentCallersGetDistinctWorkers(t *testing.T) {
	const numWorkers = 5

	s := New(Config{SchedulerType: "leastloaded", HeartbeatTTL: 30 * time.Second})
	defer s.Stop()

	for i := 0; i < numWorkers; i++ {
		id := "worker-" + string(rune('A'+i))
		if err := s.registry.Add(&registry.WorkerInfo{
			ID:      id,
			Address: "127.0.0.1:0",
			Capabilities: &pb.WorkerCapabilities{
				NativeArch: pb.Architecture_ARCH_X86_64,
				Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
			},
			MaxParallel: 1,
			State:       registry.WorkerStateIdle,
		}); err != nil {
			t.Fatalf("failed to register %s: %v", id, err)
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	seen := make(map[string]bool, numWorkers)
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			worker, _, _, err := s.dispatch(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", scheduler.TaskContext{})
			if err != nil {
				t.Errorf("dispatch failed: %v", err)
				return
			}
			mu.Lock()
			seen[worker.ID] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	if len(seen) != numWorkers {
		t.Errorf("expected all %d workers to be used exactly once, got %d distinct workers: %v", numWorkers, len(seen), seen)
	}
}
