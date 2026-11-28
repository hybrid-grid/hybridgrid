package server

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/artifacts"
	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

// SetArtifactStore enables build-output retention; nil disables it.
func (s *Server) SetArtifactStore(store *artifacts.Store) {
	s.artifacts = store
}

// UploadArtifact stores one client-supplied build output under its build.
// Unlike the compile RPCs this writes to the coordinator's disk, so it
// enforces the coordinator token when one is configured; the store caps
// the size, and an aborted or oversized upload leaves nothing behind.
func (s *Server) UploadArtifact(stream pb.BuildService_UploadArtifactServer) error {
	if s.artifacts == nil {
		return status.Error(codes.Unavailable, "artifact store disabled on this coordinator")
	}
	first, err := stream.Recv()
	if errors.Is(err, io.EOF) {
		return status.Error(codes.InvalidArgument, "empty upload")
	}
	if err != nil {
		return err
	}
	if s.config.AuthToken != "" && first.GetAuthToken() != s.config.AuthToken {
		return status.Error(codes.Unauthenticated, "invalid auth token")
	}
	buildID := validation.NormalizeBuildID(first.GetBuildId())
	if buildID == "" {
		return status.Error(codes.InvalidArgument, "valid build_id required")
	}
	name, err := artifacts.SanitizeName(first.GetName())
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid artifact name %q", first.GetName())
	}

	// Stream chunks straight into the store instead of buffering the file.
	pr, pw := io.Pipe()
	go func() {
		if _, err := pw.Write(first.GetData()); err != nil {
			return
		}
		for {
			chunk, err := stream.Recv()
			if errors.Is(err, io.EOF) {
				_ = pw.Close()
				return
			}
			if err != nil {
				_ = pw.CloseWithError(err)
				return
			}
			if _, err := pw.Write(chunk.GetData()); err != nil {
				return
			}
		}
	}()

	art, err := s.artifacts.Put(buildID, name, artifacts.SourceClient, pr, nil)
	// Unblock the reader goroutine if Put stopped early (size cap, I/O error).
	_ = pr.CloseWithError(io.ErrClosedPipe)
	if err != nil {
		if errors.Is(err, artifacts.ErrTooLarge) {
			return status.Errorf(codes.ResourceExhausted, "artifact %q exceeds the coordinator's size limit", name)
		}
		log.Warn().Err(err).Str("build_id", buildID).Str("name", name).Msg("Artifact upload failed")
		return status.Errorf(codes.Internal, "failed to store artifact: %v", err)
	}

	log.Info().Str("build_id", buildID).Str("name", art.Name).Int64("bytes", art.SizeBytes).Msg("Artifact stored")
	return stream.SendAndClose(&pb.UploadArtifactResponse{Name: art.Name, SizeBytes: art.SizeBytes, Sha256: art.SHA256})
}

// retainBuildArtifacts keeps a successful remote build's output archive
// (Flutter/Unity/custom builder) under its build id, so it can be
// downloaded from the dashboard after the client has gone.
func (s *Server) retainBuildArtifacts(buildID, kind string, resp *pb.BuildResponse) {
	if s.artifacts == nil || buildID == "" || resp == nil ||
		resp.GetStatus() != pb.TaskStatus_STATUS_COMPLETED || len(resp.GetArtifacts()) == 0 {
		return
	}
	files := make([]artifacts.File, 0, len(resp.GetArtifactList()))
	for _, a := range resp.GetArtifactList() {
		path := a.GetPath()
		if path == "" {
			path = a.GetName()
		}
		files = append(files, artifacts.File{Path: path, SizeBytes: a.GetSizeBytes()})
	}
	name := kind + "-artifacts" + archiveExt(resp.GetArtifacts())
	if _, err := s.artifacts.Put(buildID, name, artifacts.SourceWorker, bytes.NewReader(resp.GetArtifacts()), files); err != nil {
		log.Warn().Err(err).Str("build_id", buildID).Str("name", name).Msg("Failed to retain build artifacts")
	}
}

// buildKind names a remote build for its artifact file.
func buildKind(req *pb.BuildRequest) string {
	switch {
	case req.GetCustomBuilder() != "":
		return "builder-" + req.GetCustomBuilder()
	case req.GetFlutterConfig() != nil:
		// Per-ABI builds of one build must not overwrite each other.
		if platforms := req.GetFlutterConfig().GetTargetPlatforms(); len(platforms) > 0 {
			return "flutter-" + strings.Join(platforms, "+")
		}
		return "flutter"
	case req.GetUnityConfig() != nil:
		return "unity"
	default:
		return "build"
	}
}

// archiveExt names an artifact archive by its actual format: workers send
// a plain tar today, but the proto allows zstd, and builder plugins may gzip.
func archiveExt(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte{0x28, 0xb5, 0x2f, 0xfd}):
		return ".tar.zst"
	case bytes.HasPrefix(data, []byte{0x1f, 0x8b}):
		return ".tar.gz"
	default:
		return ".tar"
	}
}
