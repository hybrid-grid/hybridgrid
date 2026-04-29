package scheduler

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// TestHEFT_PicksMinEFTUnderColdStart verifies that with no observations
// the cold-start prior favours workers with more cores (Topcuoglu 2002
// Eq. 1 with average computation cost replaced by a hardware estimate
// for streaming arrivals; see linucb.go header for the cited adaptation).
func TestHEFT_PicksMinEFTUnderColdStart(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(reg.Stop)

	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "fast",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			CpuCores:   16,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))
	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "slow",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64,
			CpuCores:   1,
			Cpp:        &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))

	s := NewHEFTScheduler(HEFTConfig{Registry: reg})
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	assert.Equal(t, "fast", w.ID, "16-core worker should win EFT under cold-start prior")
	assert.True(t, info.WasExploration, "WasExploration must be true when no candidates have history")
}

// TestHEFT_RespectsObservedTimes — once we record a much faster compile
// time on the underdog, EFT should switch to it.
func TestHEFT_RespectsObservedTimes(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(reg.Stop)

	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "big",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64, CpuCores: 16,
			Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))
	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "small",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64, CpuCores: 4,
			Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 4,
	}))

	s := NewHEFTScheduler(HEFTConfig{Registry: reg, EWMAAlpha: 0.5})
	// Manually seed observations: small worker is empirically much faster.
	s.mu.Lock()
	s.wbarLocked("big").Update(2000)   // 2s
	s.wbarLocked("small").Update(200)  // 0.2s
	s.mu.Unlock()

	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	assert.Equal(t, "small", w.ID, "EFT should pick the empirically faster worker")
	assert.False(t, info.WasExploration, "with both arms warmed up, no exploration flag")
}

// TestHEFT_QueueAware — a busy worker is penalised by ActiveTasks × w̄.
func TestHEFT_QueueAware(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(reg.Stop)

	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "loaded",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64, CpuCores: 16,
			Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 8,
		ActiveTasks: 6, // queue depth 6
	}))
	require.NoError(t, reg.Add(&registry.WorkerInfo{
		ID: "idle",
		Capabilities: &pb.WorkerCapabilities{
			NativeArch: pb.Architecture_ARCH_X86_64, CpuCores: 8,
			Cpp: &pb.CppCapability{Compilers: []string{"gcc"}},
		},
		MaxParallel: 8,
		ActiveTasks: 0,
	}))

	s := NewHEFTScheduler(HEFTConfig{Registry: reg})
	// Both have similar w̄ (cold start) but queue should dominate.
	s.mu.Lock()
	s.wbarLocked("loaded").Update(500)
	s.wbarLocked("idle").Update(700) // idle is slightly slower per-task
	s.mu.Unlock()

	w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	// EFT(loaded)=500*7=3500; EFT(idle)=700*1=700 → idle wins
	assert.Equal(t, "idle", w.ID, "queue-clear time should dominate over per-task speed")
}

// TestHEFT_FastPathSingleCandidate — single worker bypasses scoring.
func TestHEFT_FastPathSingleCandidate(t *testing.T) {
	reg := newRegistryWithWorkers(t, 1)
	s := NewHEFTScheduler(HEFTConfig{Registry: reg})
	w, info, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	require.NoError(t, err)
	assert.Equal(t, "worker-a", w.ID)
	assert.False(t, info.WasExploration)
}

// TestHEFT_NoWorkers returns the documented error sentinel.
func TestHEFT_NoWorkers(t *testing.T) {
	reg := registry.NewInMemoryRegistry(60 * time.Second)
	t.Cleanup(reg.Stop)
	s := NewHEFTScheduler(HEFTConfig{Registry: reg})
	_, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
	assert.ErrorIs(t, err, ErrNoWorkers)
}

// TestHEFT_ConcurrentSafe under -race.
func TestHEFT_ConcurrentSafe(t *testing.T) {
	reg := newRegistryWithWorkers(t, 4)
	s := NewHEFTScheduler(HEFTConfig{Registry: reg, EWMAAlpha: 0.3})

	const goroutines = 16
	const iterations = 100
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				w, _, err := s.SelectWithDispatchInfo(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "", TaskContext{})
				if err != nil {
					return
				}
				s.RecordOutcome(w.ID, -float64(i), true, TaskContext{})
			}
		}()
	}
	wg.Wait()
}
