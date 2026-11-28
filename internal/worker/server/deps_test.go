package server

import (
	"io"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

func TestResolveDeps(t *testing.T) {
	s := &Server{deps: newDepsCache(Config{DepsCacheDir: t.TempDir()})}

	if open, err := s.resolveDeps(&pb.BuildRequest{}); open != nil || err != nil {
		t.Fatalf("no layer: %v, %v", open != nil, err)
	}

	_, err := s.resolveDeps(&pb.BuildRequest{DepsHash: "abc"})
	if st, _ := status.FromError(err); st.Code() != codes.FailedPrecondition || st.Message() != depslayer.MissingMessage {
		t.Fatalf("missing layer: %v", err)
	}

	if _, err := s.resolveDeps(&pb.BuildRequest{DepsHash: "abc", DepsArchive: []byte("layer")}); err != nil {
		t.Fatal(err)
	}
	open, err := s.resolveDeps(&pb.BuildRequest{DepsHash: "abc"})
	if err != nil {
		t.Fatalf("cached layer: %v", err)
	}
	rc, err := open()
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	if data, _ := io.ReadAll(rc); string(data) != "layer" {
		t.Fatalf("layer = %q", data)
	}

	if _, err := s.resolveDeps(&pb.BuildRequest{DepsHash: "../escape", DepsArchive: []byte("x")}); err == nil {
		t.Fatal("unsafe hash accepted")
	}
}
