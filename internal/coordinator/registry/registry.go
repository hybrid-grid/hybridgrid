package registry

import (
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// WorkerState represents the current state of a worker.
type WorkerState int

const (
	WorkerStateIdle WorkerState = iota
	WorkerStateBusy
	WorkerStateUnhealthy
)

func (s WorkerState) String() string {
	switch s {
	case WorkerStateIdle:
		return "idle"
	case WorkerStateBusy:
		return "busy"
	case WorkerStateUnhealthy:
		return "unhealthy"
	}
	return "unknown"
}

// WorkerInfo stores information about a registered worker.
type WorkerInfo struct {
	ID              string
	Address         string
	Capabilities    *pb.WorkerCapabilities
	State           WorkerState
	DiscoverySource string // "mdns", "wan", "manual"
	LastHeartbeat   time.Time
	RegisteredAt    time.Time
	MaxParallel     int32 // Max concurrent tasks this worker can handle

	// Metrics
	ActiveTasks     int32
	TotalTasks      int64
	SuccessfulTasks int64
	FailedTasks     int64
	AvgCompileTime  time.Duration
}

// IsHealthy returns true if the worker is considered healthy.
func (w *WorkerInfo) IsHealthy(ttl time.Duration) bool {
	if w.State == WorkerStateUnhealthy {
		return false
	}
	return time.Since(w.LastHeartbeat) <= ttl
}

// Registry manages registered workers.
type Registry interface {
	// Add registers a new worker.
	Add(worker *WorkerInfo) error

	// Remove unregisters a worker by ID.
	Remove(id string) error

	// Get returns a worker by ID.
	Get(id string) (*WorkerInfo, bool)

	// List returns all registered workers.
	List() []*WorkerInfo

	// ListByCapability returns workers matching the given criteria.
	ListByCapability(buildType pb.BuildType, arch pb.Architecture) []*WorkerInfo

	// UpdateState updates a worker's state.
	UpdateState(id string, state WorkerState) error

	// UpdateHeartbeat updates the last heartbeat time.
	UpdateHeartbeat(id string) error

	// IncrementTasks increments the active task count.
	IncrementTasks(id string) error

	// DecrementTasks decrements the active task count.
	DecrementTasks(id string, success bool, compileTime time.Duration) error

	// Count returns the number of registered workers.
	Count() int
}

// InMemoryRegistry implements Registry with in-memory storage.
type InMemoryRegistry struct {
	mu      sync.RWMutex
	workers map[string]*WorkerInfo
	ttl     time.Duration
	stopCh  chan struct{}
}

// NewInMemoryRegistry creates a new in-memory registry.
func NewInMemoryRegistry(ttl time.Duration) *InMemoryRegistry {
	r := &InMemoryRegistry{
		workers: make(map[string]*WorkerInfo),
		ttl:     ttl,
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go r.cleanupLoop()

	return r
}

// Add registers a new worker or updates an existing one.
func (r *InMemoryRegistry) Add(worker *WorkerInfo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, exists := r.workers[worker.ID]; exists {
		// Update existing worker's capabilities and heartbeat
		existing.Capabilities = worker.Capabilities
		existing.Address = worker.Address
		existing.MaxParallel = worker.MaxParallel
		existing.LastHeartbeat = time.Now()
		// Reset unhealthy state when worker heartbeats
		if existing.State == WorkerStateUnhealthy {
			existing.State = WorkerStateIdle
		}
		return nil
	}

	worker.RegisteredAt = time.Now()
	worker.LastHeartbeat = time.Now()
	worker.State = WorkerStateIdle
	r.workers[worker.ID] = worker

	return nil
}

// Remove unregisters a worker.
func (r *InMemoryRegistry) Remove(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.workers[id]; !exists {
		return fmt.Errorf("worker %s not found", id)
	}

	delete(r.workers, id)
	return nil
}

// Get returns a worker by ID.
func (r *InMemoryRegistry) Get(id string) (*WorkerInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	worker, ok := r.workers[id]
	if !ok {
		return nil, false
	}

	// Return a copy to prevent data races
	copy := *worker
	return &copy, true
}

// List returns all registered workers.
func (r *InMemoryRegistry) List() []*WorkerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*WorkerInfo, 0, len(r.workers))
	for _, w := range r.workers {
		copy := *w
		result = append(result, &copy)
	}
	return result
}

// ListByCapability returns workers matching the given criteria.
func (r *InMemoryRegistry) ListByCapability(buildType pb.BuildType, arch pb.Architecture) []*WorkerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*WorkerInfo, 0)
	for _, w := range r.workers {
		if !w.IsHealthy(r.ttl) {
			continue
		}

		if r.matchesCapability(w.Capabilities, buildType, arch) {
			copy := *w
			result = append(result, &copy)
		}
	}
	return result
}

// matchesCapability checks if worker capabilities match the requirements.
func (r *InMemoryRegistry) matchesCapability(caps *pb.WorkerCapabilities, buildType pb.BuildType, arch pb.Architecture) bool {
	if caps == nil {
		log.Debug().Msg("matchesCapability: caps is nil")
		return false
	}

	// Check architecture
	if arch != pb.Architecture_ARCH_UNSPECIFIED {
		archMatch := caps.NativeArch == arch
		canCrossCompile := caps.DockerAvailable || (caps.Cpp != nil && caps.Cpp.CrossCompile)

		if !archMatch && !canCrossCompile {
			log.Debug().
				Str("worker", caps.WorkerId).
				Str("native_arch", caps.NativeArch.String()).
				Str("requested_arch", arch.String()).
				Bool("docker_available", caps.DockerAvailable).
				Bool("cross_compile", caps.Cpp != nil && caps.Cpp.CrossCompile).
				Msg("matchesCapability: architecture mismatch")
			return false
		}

		if !archMatch && canCrossCompile {
			log.Debug().
				Str("worker", caps.WorkerId).
				Str("native_arch", caps.NativeArch.String()).
				Str("requested_arch", arch.String()).
				Bool("docker", caps.DockerAvailable).
				Bool("cross_compile", caps.Cpp != nil && caps.Cpp.CrossCompile).
				Msg("matchesCapability: cross-compilation available")
		}
	}

	// Check build type support
	switch buildType {
	case pb.BuildType_BUILD_TYPE_CPP:
		hasCpp := caps.Cpp != nil && len(caps.Cpp.Compilers) > 0
		if !hasCpp {
			log.Debug().
				Str("worker", caps.WorkerId).
				Bool("cpp_nil", caps.Cpp == nil).
				Msg("matchesCapability: no C++ capability")
			return false
		}
		log.Debug().
			Str("worker", caps.WorkerId).
			Strs("compilers", caps.Cpp.Compilers).
			Msg("matchesCapability: C++ capability found")
	case pb.BuildType_BUILD_TYPE_GO:
		if caps.Go == nil {
			return false
		}
	case pb.BuildType_BUILD_TYPE_RUST:
		if caps.Rust == nil {
			return false
		}
	case pb.BuildType_BUILD_TYPE_NODEJS:
		if caps.Nodejs == nil {
			return false
		}
	case pb.BuildType_BUILD_TYPE_FLUTTER:
		if caps.Flutter == nil {
			return false
		}
	case pb.BuildType_BUILD_TYPE_UNSPECIFIED:
		// Any worker can handle unspecified
		return true
	}

	return true
}

// UpdateState updates a worker's state.
func (r *InMemoryRegistry) UpdateState(id string, state WorkerState) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	worker, ok := r.workers[id]
	if !ok {
		return fmt.Errorf("worker %s not found", id)
	}

	worker.State = state
	return nil
}

// UpdateHeartbeat updates the last heartbeat time.
func (r *InMemoryRegistry) UpdateHeartbeat(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	worker, ok := r.workers[id]
	if !ok {
		return fmt.Errorf("worker %s not found", id)
	}

	worker.LastHeartbeat = time.Now()
	if worker.State == WorkerStateUnhealthy {
		worker.State = WorkerStateIdle
	}
	return nil
}

// IncrementTasks increments the active task count.
func (r *InMemoryRegistry) IncrementTasks(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	worker, ok := r.workers[id]
	if !ok {
		return fmt.Errorf("worker %s not found", id)
	}

	worker.ActiveTasks++
	worker.TotalTasks++
	if worker.State == WorkerStateIdle {
		worker.State = WorkerStateBusy
	}
	return nil
}

// DecrementTasks decrements the active task count.
func (r *InMemoryRegistry) DecrementTasks(id string, success bool, compileTime time.Duration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	worker, ok := r.workers[id]
	if !ok {
		return fmt.Errorf("worker %s not found", id)
	}

	if worker.ActiveTasks > 0 {
		worker.ActiveTasks--
	}

	if success {
		worker.SuccessfulTasks++
	} else {
		worker.FailedTasks++
	}

	// Update average compile time (simple moving average)
	if worker.SuccessfulTasks > 0 {
		total := int64(worker.AvgCompileTime) * (worker.SuccessfulTasks - 1)
		worker.AvgCompileTime = time.Duration((total + int64(compileTime)) / worker.SuccessfulTasks)
	}

	if worker.ActiveTasks == 0 && worker.State == WorkerStateBusy {
		worker.State = WorkerStateIdle
	}

	return nil
}

// Count returns the number of registered workers.
func (r *InMemoryRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.workers)
}

// cleanupLoop periodically removes stale workers.
func (r *InMemoryRegistry) cleanupLoop() {
	ticker := time.NewTicker(r.ttl / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			r.cleanupStaleWorkers()
		case <-r.stopCh:
			return
		}
	}
}

// cleanupStaleWorkers marks workers as unhealthy if they haven't sent heartbeat.
func (r *InMemoryRegistry) cleanupStaleWorkers() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, w := range r.workers {
		if time.Since(w.LastHeartbeat) > r.ttl {
			w.State = WorkerStateUnhealthy
		}
	}
}

// Stop stops the cleanup goroutine.
func (r *InMemoryRegistry) Stop() {
	close(r.stopCh)
}
