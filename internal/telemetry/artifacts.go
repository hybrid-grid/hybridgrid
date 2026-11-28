package telemetry

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/artifacts"
)

// downloadChunkSize keeps each streamed message well under gRPC's 4 MiB
// default receive limit.
const downloadChunkSize = 256 << 10

// SetArtifactStore exposes the coordinator's stored build outputs to the
// dashboard; nil (the default) reports them as unavailable.
func (s *Service) SetArtifactStore(store *artifacts.Store) {
	s.artifacts = store
}

// ListArtifacts returns a build's stored outputs, newest first.
func (s *Service) ListArtifacts(_ context.Context, req *pb.ListTelemetryArtifactsRequest) (*pb.ListTelemetryArtifactsResponse, error) {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return nil, err
	}
	if s.artifacts == nil {
		return nil, status.Error(codes.Unavailable, "artifact store disabled on this coordinator")
	}
	list, err := s.artifacts.List(req.GetBuildId())
	if err != nil {
		return nil, artifactError(err)
	}
	out := make([]*pb.TelemetryArtifact, 0, len(list))
	for i := range list {
		out = append(out, artifactToProto(&list[i]))
	}
	return &pb.ListTelemetryArtifactsResponse{Artifacts: out}, nil
}

// DownloadArtifact streams one stored artifact; the first chunk carries
// its metadata so the caller can set headers before the bytes arrive.
func (s *Service) DownloadArtifact(req *pb.DownloadTelemetryArtifactRequest, stream pb.TelemetryService_DownloadArtifactServer) error {
	if err := s.authorize(req.GetAuthToken()); err != nil {
		return err
	}
	if s.artifacts == nil {
		return status.Error(codes.Unavailable, "artifact store disabled on this coordinator")
	}
	f, meta, err := s.artifacts.Open(req.GetBuildId(), req.GetName())
	if err != nil {
		return artifactError(err)
	}
	defer f.Close()

	first := true
	buf := make([]byte, downloadChunkSize)
	for {
		n, readErr := f.Read(buf)
		if n > 0 || first {
			chunk := &pb.TelemetryArtifactChunk{Data: append([]byte(nil), buf[:n]...)}
			if first {
				chunk.Artifact = artifactToProto(meta)
				first = false
			}
			if err := stream.Send(chunk); err != nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			return nil
		}
		if readErr != nil {
			return status.Errorf(codes.Internal, "failed to read artifact: %v", readErr)
		}
	}
}

func artifactError(err error) error {
	switch {
	case errors.Is(err, artifacts.ErrNotFound):
		return status.Error(codes.NotFound, "artifact not found")
	case errors.Is(err, artifacts.ErrInvalidBuildID), errors.Is(err, artifacts.ErrInvalidName):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Errorf(codes.Internal, "artifact store: %v", err)
	}
}

func artifactToProto(a *artifacts.Artifact) *pb.TelemetryArtifact {
	files := make([]*pb.TelemetryArtifactFile, 0, len(a.Files))
	for _, f := range a.Files {
		files = append(files, &pb.TelemetryArtifactFile{Path: f.Path, SizeBytes: f.SizeBytes})
	}
	return &pb.TelemetryArtifact{
		BuildId:     a.BuildID,
		Name:        a.Name,
		SizeBytes:   a.SizeBytes,
		Sha256:      a.SHA256,
		CreatedAtMs: a.CreatedAtMs,
		Source:      a.Source,
		Files:       files,
	}
}
