package server

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

// Build handles build requests.
func (s *Server) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	buildID := validation.NormalizeBuildID(req.BuildId)
	if req.TaskId == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id required")
	}

	resp, err := s.dispatchBuild(ctx, req, buildID)
	if err == nil {
		s.retainBuildArtifacts(buildID, buildKind(req), resp)
	}
	return resp, err
}

func (s *Server) dispatchBuild(ctx context.Context, req *pb.BuildRequest, buildID string) (*pb.BuildResponse, error) {
	if req.CustomBuilder != "" {
		return s.handleCustomBuild(ctx, req, buildID)
	}

	if req.GetFlutterConfig() != nil {
		return s.handleFlutterBuild(ctx, req, buildID)
	}

	if req.GetUnityConfig() != nil {
		return s.handleUnityBuild(ctx, req, buildID)
	}

	if req.GetNodeConfig() != nil {
		return s.handleTscBuild(ctx, req, buildID)
	}

	return &pb.BuildResponse{
		Status:   pb.TaskStatus_STATUS_FAILED,
		ExitCode: 1,
		Stderr:   "build type not implemented yet",
	}, nil
}

// StreamBuild is not implemented on the coordinator. It returns
// codes.Unimplemented immediately, without reading the client stream.
func (s *Server) StreamBuild(stream pb.BuildService_StreamBuildServer) error {
	return status.Error(codes.Unimplemented, "stream build not implemented")
}

func cloneArtifactList(list []*pb.ArtifactInfo) []*pb.ArtifactInfo {
	if list == nil {
		return nil
	}
	result := make([]*pb.ArtifactInfo, len(list))
	for i, a := range list {
		result[i] = &pb.ArtifactInfo{
			Name:      a.Name,
			Path:      a.Path,
			SizeBytes: a.SizeBytes,
			Checksum:  a.Checksum,
		}
	}
	return result
}
