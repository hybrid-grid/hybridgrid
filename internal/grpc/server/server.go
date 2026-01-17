package server

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// Config holds the gRPC server configuration.
type Config struct {
	Port          int
	AuthToken     string
	MaxConcurrent int
}

// Server implements the BuildService gRPC server.
type Server struct {
	pb.UnimplementedBuildServiceServer

	config Config
	server *grpc.Server

	mu      sync.RWMutex
	workers map[string]*pb.WorkerCapabilities
}

// New creates a new gRPC server instance.
func New(cfg Config) *Server {
	return &Server{
		config:  cfg,
		workers: make(map[string]*pb.WorkerCapabilities),
	}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer()
	pb.RegisterBuildServiceServer(s.server, s)

	return s.server.Serve(lis)
}

// Stop gracefully stops the gRPC server.
func (s *Server) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Handshake handles worker registration.
func (s *Server) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	if req.Capabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "capabilities required")
	}

	// Validate auth token if configured
	if s.config.AuthToken != "" && req.AuthToken != s.config.AuthToken {
		return &pb.HandshakeResponse{
			Accepted: false,
			Message:  "invalid auth token",
		}, nil
	}

	workerID := req.Capabilities.WorkerId
	if workerID == "" {
		workerID = fmt.Sprintf("worker-%s", req.Capabilities.Hostname)
	}

	s.mu.Lock()
	s.workers[workerID] = req.Capabilities
	s.mu.Unlock()

	return &pb.HandshakeResponse{
		Accepted:                 true,
		Message:                  "worker registered successfully",
		AssignedWorkerId:         workerID,
		HeartbeatIntervalSeconds: 30,
	}, nil
}

// Build handles a build request.
func (s *Server) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// TODO: Implement build logic
	// 1. Find suitable worker based on capabilities
	// 2. Forward request to worker
	// 3. Return response

	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "build not implemented yet",
	}, nil
}

// StreamBuild handles streaming build requests for large projects.
func (s *Server) StreamBuild(stream pb.BuildService_StreamBuildServer) error {
	var metadata *pb.BuildMetadata
	var sourceData []byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "failed to receive chunk: %v", err)
		}

		switch payload := chunk.Payload.(type) {
		case *pb.BuildChunk_Metadata:
			metadata = payload.Metadata
		case *pb.BuildChunk_SourceChunk:
			sourceData = append(sourceData, payload.SourceChunk...)
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata required")
	}

	// TODO: Implement streaming build logic
	return stream.SendAndClose(&pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "stream build not implemented yet",
	})
}

// Compile handles legacy C/C++ compilation requests.
func (s *Server) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	// TODO: Implement compile logic
	return &pb.CompileResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "compile not implemented yet",
	}, nil
}

// HealthCheck returns the server health status.
func (s *Server) HealthCheck(ctx context.Context, req *pb.HealthRequest) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{
		Healthy:     true,
		ActiveTasks: 0,
		QueuedTasks: 0,
	}, nil
}

// GetWorkerStatus returns the status of all registered workers.
func (s *Server) GetWorkerStatus(ctx context.Context, req *pb.WorkerStatusRequest) (*pb.WorkerStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workers := make([]*pb.WorkerStatusResponse_WorkerInfo, 0, len(s.workers))
	for id, caps := range s.workers {
		workers = append(workers, &pb.WorkerStatusResponse_WorkerInfo{
			WorkerId:    id,
			Host:        caps.Hostname,
			NativeArch:  caps.NativeArch,
			CpuCores:    caps.CpuCores,
			MemoryBytes: caps.MemoryBytes,
		})
	}

	return &pb.WorkerStatusResponse{
		Workers:        workers,
		TotalWorkers:   int32(len(workers)),
		HealthyWorkers: int32(len(workers)),
	}, nil
}

// GetWorkersForBuild returns workers capable of handling a specific build type.
func (s *Server) GetWorkersForBuild(ctx context.Context, req *pb.WorkersForBuildRequest) (*pb.WorkersForBuildResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var capable []string
	for id, caps := range s.workers {
		if s.workerCanHandle(caps, req.BuildType, req.TargetPlatform) {
			capable = append(capable, id)
		}
	}

	return &pb.WorkersForBuildResponse{
		WorkerIds:      capable,
		AvailableCount: int32(len(capable)),
	}, nil
}

// workerCanHandle checks if a worker can handle a specific build type and platform.
func (s *Server) workerCanHandle(caps *pb.WorkerCapabilities, buildType pb.BuildType, platform pb.TargetPlatform) bool {
	switch buildType {
	case pb.BuildType_BUILD_TYPE_CPP:
		return caps.Cpp != nil && len(caps.Cpp.Compilers) > 0
	case pb.BuildType_BUILD_TYPE_FLUTTER:
		if caps.Flutter == nil {
			return false
		}
		for _, p := range caps.Flutter.Platforms {
			if p == platform {
				return true
			}
		}
		return false
	case pb.BuildType_BUILD_TYPE_UNITY:
		if caps.Unity == nil {
			return false
		}
		for _, p := range caps.Unity.BuildTargets {
			if p == platform {
				return true
			}
		}
		return false
	case pb.BuildType_BUILD_TYPE_COCOS:
		if caps.Cocos == nil {
			return false
		}
		for _, p := range caps.Cocos.Platforms {
			if p == platform {
				return true
			}
		}
		return false
	case pb.BuildType_BUILD_TYPE_RUST:
		return caps.Rust != nil && len(caps.Rust.Toolchains) > 0
	case pb.BuildType_BUILD_TYPE_GO:
		return caps.Go != nil && caps.Go.Version != ""
	case pb.BuildType_BUILD_TYPE_NODEJS:
		return caps.Nodejs != nil && len(caps.Nodejs.Versions) > 0
	default:
		return false
	}
}
