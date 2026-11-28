package server

import (
	"context"
	"fmt"
	"math"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// Handshake handles worker registration.
func (s *Server) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	if req.Capabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "capabilities required")
	}

	// Validate auth token
	if s.config.AuthToken != "" && req.AuthToken != s.config.AuthToken {
		log.Warn().Str("hostname", req.Capabilities.Hostname).Msg("Worker rejected: invalid auth token")
		return &pb.HandshakeResponse{
			Accepted: false,
			Message:  "invalid auth token",
		}, nil
	}

	// Generate worker ID
	workerID := req.Capabilities.WorkerId
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%s-%d", req.Capabilities.Hostname, time.Now().UnixNano())
	}

	// Get worker address from context (or use provided)
	workerAddr := req.WorkerAddress
	if workerAddr == "" {
		workerAddr = fmt.Sprintf("%s:50052", req.Capabilities.Hostname)
	}

	// Get max parallel from capabilities (default to 4 if not set)
	maxParallel := req.Capabilities.MaxParallelTasks
	if maxParallel <= 0 {
		maxParallel = 4
	}

	// Register worker
	worker := &registry.WorkerInfo{
		ID:           workerID,
		Address:      workerAddr,
		Capabilities: req.Capabilities,
		MaxParallel:  maxParallel,
	}

	if err := s.registry.Add(worker); err != nil {
		// Worker might already exist, update heartbeat instead
		if err := s.registry.UpdateHeartbeat(workerID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to register worker: %v", err)
		}
	} else {
		// Fresh capacity joined the cluster — wake queued waiters.
		s.signalDispatchCapacity()
	}

	// Log C++ capabilities for debugging
	var compilers []string
	if req.Capabilities.Cpp != nil {
		compilers = req.Capabilities.Cpp.Compilers
	}

	log.Info().
		Str("worker_id", workerID).
		Str("hostname", req.Capabilities.Hostname).
		Int32("cpu_cores", req.Capabilities.CpuCores).
		Int32("max_parallel", maxParallel).
		Str("arch", req.Capabilities.NativeArch.String()).
		Strs("cpp_compilers", compilers).
		Bool("docker", req.Capabilities.DockerAvailable).
		Msg("Worker registered")

	return &pb.HandshakeResponse{
		Accepted:                 true,
		Message:                  "worker registered successfully",
		AssignedWorkerId:         workerID,
		HeartbeatIntervalSeconds: int32(s.config.HeartbeatTTL.Seconds() / 2),
	}, nil
}

// HealthCheck returns coordinator health status.
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	workers := s.registry.List()
	healthyCount := 0
	for _, w := range workers {
		if w.IsHealthy(s.config.HeartbeatTTL) {
			healthyCount++
		}
	}

	return &pb.HealthResponse{
		Healthy:     healthyCount > 0 || len(workers) == 0,
		ActiveTasks: clampInt64ToInt32(atomic.LoadInt64(&s.activeTasks)),
		QueuedTasks: clampInt64ToInt32(atomic.LoadInt64(&s.queuedTasks)),
	}, nil
}

// GetWorkerStatus returns status of all workers.
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	workers := s.registry.List()

	infos := make([]*pb.WorkerStatusResponse_WorkerInfo, 0, len(workers))
	healthyCount := 0

	for _, w := range workers {
		healthy := w.IsHealthy(s.config.HeartbeatTTL)
		if healthy {
			healthyCount++
		}

		caps := workerCapabilitiesOrDefault(w)

		info := &pb.WorkerStatusResponse_WorkerInfo{
			WorkerId:            w.ID,
			Host:                caps.Hostname,
			NativeArch:          caps.NativeArch,
			CpuCores:            caps.CpuCores,
			MemoryBytes:         caps.MemoryBytes,
			ActiveTasks:         w.ActiveTasks,
			TotalTasksCompleted: w.TotalTasks,
			LastHeartbeatUnix:   w.LastHeartbeat.Unix(),
			Healthy:             &healthy,
			CircuitState:        "CLOSED",
		}
		if s.circuitManager != nil {
			info.CircuitState = string(s.circuitManager.GetState(w.ID))
		}
		infos = append(infos, info)
	}

	return &pb.WorkerStatusResponse{
		Workers:        infos,
		TotalWorkers:   int32(len(workers)),
		HealthyWorkers: int32(healthyCount),
	}, nil
}

// GetWorkersForBuild returns workers capable of handling a build type.
func (s *Server) GetWorkersForBuild(ctx context.Context, req *pb.WorkersForBuildRequest) (*pb.WorkersForBuildResponse, error) {
	workers := s.registry.ListByCapability(req.BuildType, pb.Architecture_ARCH_UNSPECIFIED)

	ids := make([]string, 0, len(workers))
	for _, w := range workers {
		ids = append(ids, w.ID)
	}

	return &pb.WorkersForBuildResponse{
		WorkerIds:      ids,
		AvailableCount: int32(len(ids)),
	}, nil
}

// ReportCacheHit handles cache hit reports from clients.
func (s *Server) ReportCacheHit(ctx context.Context, req *pb.ReportCacheHitRequest) (*pb.ReportCacheHitResponse, error) {
	if req.Hits > 0 {
		atomic.AddInt64(&s.cacheHits, int64(req.Hits))
	}
	return &pb.ReportCacheHitResponse{Acknowledged: true}, nil
}

func clampInt64ToInt32(v int64) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}

func workerCapabilitiesOrDefault(worker *registry.WorkerInfo) *pb.WorkerCapabilities {
	if worker != nil && worker.Capabilities != nil {
		return worker.Capabilities
	}

	return &pb.WorkerCapabilities{}
}
