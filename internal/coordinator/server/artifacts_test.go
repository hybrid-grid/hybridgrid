package server

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/artifacts"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

func newArtifactStore(t *testing.T, maxFile int64) *artifacts.Store {
	t.Helper()
	store, err := artifacts.New(artifacts.Config{Dir: t.TempDir(), MaxFileBytes: maxFile})
	require.NoError(t, err)
	return store
}

// upload sends header + data in chunks of `chunk` bytes and returns the result.
func upload(client pb.BuildServiceClient, header *pb.ArtifactChunk, data string, chunk int) (*pb.UploadArtifactResponse, error) {
	stream, err := client.UploadArtifact(context.Background())
	if err != nil {
		return nil, err
	}
	first := true
	for len(data) > 0 || first {
		n := min(chunk, len(data))
		msg := &pb.ArtifactChunk{Data: []byte(data[:n])}
		if first {
			msg.BuildId, msg.Name, msg.AuthToken = header.BuildId, header.Name, header.AuthToken
			first = false
		}
		data = data[n:]
		if err := stream.Send(msg); err != nil {
			if err == io.EOF {
				break // server closed early; its status comes from CloseAndRecv
			}
			return nil, err
		}
	}
	return stream.CloseAndRecv()
}

func TestUploadArtifact_StoresChunkedUpload(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second})
	defer cleanup()
	store := newArtifactStore(t, 0)
	s.SetArtifactStore(store)

	resp, err := upload(client, &pb.ArtifactChunk{BuildId: "build-1", Name: "out/app"}, "hello world", 3)
	require.NoError(t, err)
	assert.Equal(t, "app", resp.Name)
	assert.Equal(t, int64(11), resp.SizeBytes)
	assert.Len(t, resp.Sha256, 64)

	f, meta, err := store.Open("build-1", "app")
	require.NoError(t, err)
	defer f.Close()
	data, _ := io.ReadAll(f)
	assert.Equal(t, "hello world", string(data))
	assert.Equal(t, artifacts.SourceClient, meta.Source)
}

func TestUploadArtifact_Rejections(t *testing.T) {
	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, AuthToken: "secret"})
	defer cleanup()

	// Store disabled.
	_, err := upload(client, &pb.ArtifactChunk{BuildId: "b", Name: "a", AuthToken: "secret"}, "x", 8)
	assert.Equal(t, codes.Unavailable, status.Code(err))

	s.SetArtifactStore(newArtifactStore(t, 4))

	tests := []struct {
		name   string
		header *pb.ArtifactChunk
		data   string
		want   codes.Code
	}{
		{"wrong token", &pb.ArtifactChunk{BuildId: "b", Name: "a", AuthToken: "nope"}, "x", codes.Unauthenticated},
		{"bad build id", &pb.ArtifactChunk{BuildId: "../x", Name: "a", AuthToken: "secret"}, "x", codes.InvalidArgument},
		{"bad name", &pb.ArtifactChunk{BuildId: "b", Name: "..", AuthToken: "secret"}, "x", codes.InvalidArgument},
		{"too large", &pb.ArtifactChunk{BuildId: "b", Name: "big", AuthToken: "secret"}, strings.Repeat("x", 64), codes.ResourceExhausted},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := upload(client, tt.header, tt.data, 8)
			assert.Equal(t, tt.want, status.Code(err), "err = %v", err)
		})
	}
}

// A successful remote build keeps its output archive under the build id;
// a build without a build id, or a failed one, keeps nothing.
func TestBuild_RetainsRemoteArtifacts(t *testing.T) {
	addr, workerCleanup := setupTestWorker(t, func(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
		return &pb.BuildResponse{
			Status:       pb.TaskStatus_STATUS_COMPLETED,
			Artifacts:    []byte("tar-zst-bytes"),
			ArtifactList: []*pb.ArtifactInfo{{Name: "app.apk", Path: "build/app.apk", SizeBytes: 42}},
		}, nil
	})
	defer workerCleanup()

	s, client, cleanup := setupTestServer(t, Config{Port: 0, HeartbeatTTL: 30 * time.Second, RequestTimeout: 30 * time.Second})
	defer cleanup()
	store := newArtifactStore(t, 0)
	s.SetArtifactStore(store)
	require.NoError(t, s.registry.Add(&registry.WorkerInfo{
		ID:      "builder",
		Address: addr,
		Capabilities: &pb.WorkerCapabilities{
			NativeArch:            pb.Architecture_ARCH_X86_64,
			CustomBuilders:        []string{"demo"},
			CustomBuilderVersions: map[string]string{"demo": "v1"},
		},
		MaxParallel: 2,
	}))

	resp, err := client.Build(context.Background(), &pb.BuildRequest{
		TaskId: "t1", BuildId: "release-7", SourceHash: "h", SourceArchive: []byte("src"), CustomBuilder: "demo",
	})
	require.NoError(t, err)
	require.Equal(t, pb.TaskStatus_STATUS_COMPLETED, resp.Status)

	list, err := store.List("release-7")
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "builder-demo-artifacts.tar", list[0].Name, "plain tar payload is named .tar")
	assert.Equal(t, artifacts.SourceWorker, list[0].Source)
	assert.Equal(t, []artifacts.File{{Path: "build/app.apk", SizeBytes: 42}}, list[0].Files)

	// No build id: nothing to attach the output to.
	_, err = client.Build(context.Background(), &pb.BuildRequest{
		TaskId: "t2", SourceHash: "h2", SourceArchive: []byte("src2"), CustomBuilder: "demo",
	})
	require.NoError(t, err)
	entries, err := store.List("t2")
	require.NoError(t, err)
	assert.Empty(t, entries)
}

func TestArchiveExtAndBuildKind(t *testing.T) {
	assert.Equal(t, ".tar.zst", archiveExt([]byte{0x28, 0xb5, 0x2f, 0xfd, 0x00}))
	assert.Equal(t, ".tar.gz", archiveExt([]byte{0x1f, 0x8b, 0x08}))
	assert.Equal(t, ".tar", archiveExt([]byte("ustar...")))
	assert.Equal(t, ".tar", archiveExt(nil))

	flutter := func(platforms ...string) *pb.BuildRequest {
		return &pb.BuildRequest{Config: &pb.BuildRequest_FlutterConfig{FlutterConfig: &pb.FlutterConfig{TargetPlatforms: platforms}}}
	}
	assert.Equal(t, "flutter", buildKind(flutter()))
	assert.Equal(t, "flutter-android-arm64", buildKind(flutter("android-arm64")))
	assert.Equal(t, "flutter-android-arm+android-x64", buildKind(flutter("android-arm", "android-x64")))
	assert.Equal(t, "builder-demo", buildKind(&pb.BuildRequest{CustomBuilder: "demo"}))
}
