package ui

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strconv"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// ArtifactFile is one entry inside an archive artifact.
type ArtifactFile struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
}

// Artifact is a stored build output as the SPA sees it.
type Artifact struct {
	BuildID     string         `json:"build_id"`
	Name        string         `json:"name"`
	SizeBytes   int64          `json:"size_bytes"`
	SHA256      string         `json:"sha256"`
	CreatedAtMs int64          `json:"created_at_ms"`
	Source      string         `json:"source"`
	Files       []ArtifactFile `json:"files"`
}

func artifactFromProto(a *pb.TelemetryArtifact) Artifact {
	files := make([]ArtifactFile, 0, len(a.GetFiles()))
	for _, f := range a.GetFiles() {
		files = append(files, ArtifactFile{Path: f.GetPath(), SizeBytes: f.GetSizeBytes()})
	}
	return Artifact{
		BuildID:     a.GetBuildId(),
		Name:        a.GetName(),
		SizeBytes:   a.GetSizeBytes(),
		SHA256:      a.GetSha256(),
		CreatedAtMs: a.GetCreatedAtMs(),
		Source:      a.GetSource(),
		Files:       files,
	}
}

// handleBuildArtifacts lists a build's stored outputs, newest first.
// 503 when the coordinator runs without an artifact store.
func (s *Server) handleBuildArtifacts(w http.ResponseWriter, r *http.Request) {
	resp, err := s.client.ListArtifacts(r.Context(), &pb.ListTelemetryArtifactsRequest{
		AuthToken: s.config.AuthToken,
		BuildId:   r.PathValue("id"),
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	list := make([]Artifact, 0, len(resp.GetArtifacts()))
	for _, a := range resp.GetArtifacts() {
		list = append(list, artifactFromProto(a))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"build_id": r.PathValue("id"), "artifacts": list})
}

// handleArtifactDownload streams one artifact from the coordinator to the
// browser without buffering it, as an attachment under its stored name.
func (s *Server) handleArtifactDownload(w http.ResponseWriter, r *http.Request) {
	stream, err := s.client.DownloadArtifact(r.Context(), &pb.DownloadTelemetryArtifactRequest{
		AuthToken: s.config.AuthToken,
		BuildId:   r.PathValue("id"),
		Name:      r.PathValue("name"),
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	// A server stream reports NOT_FOUND etc. on the first Recv, so headers
	// are only committed once the artifact is known to exist.
	first, err := stream.Recv()
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	meta := first.GetArtifact()
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": meta.GetName()})
	if disposition == "" {
		disposition = "attachment"
	}
	h := w.Header()
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Disposition", disposition)
	h.Set("Content-Length", strconv.FormatInt(meta.GetSizeBytes(), 10))
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("X-Checksum-Sha256", meta.GetSha256())
	h.Set("Cache-Control", "no-store")

	if _, err := w.Write(first.GetData()); err != nil {
		return
	}
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return
		}
		if err != nil {
			// Headers are already sent: the short body against
			// Content-Length tells the browser the download failed.
			return
		}
		if _, err := w.Write(chunk.GetData()); err != nil {
			return
		}
	}
}
