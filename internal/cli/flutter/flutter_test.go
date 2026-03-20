package flutter

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

type fakeClient struct {
	lastRequest *pb.BuildRequest
}

func (f *fakeClient) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	f.lastRequest = req
	return &pb.BuildResponse{
		Status: pb.TaskStatus_STATUS_COMPLETED,
	}, nil
}

func (f *fakeClient) Close() error {
	return nil
}

func TestFlutterCLI_RequiresProject(t *testing.T) {
	client := &fakeClient{}
	cmd := NewCommand(Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			return client, nil
		},
		RequestTimeout: time.Second,
	})

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"build", "apk"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing project")
	}
	if !strings.Contains(err.Error(), "project") {
		t.Fatalf("expected project error, got %v", err)
	}
}

func TestFlutterCLI_BuildModeFlavor(t *testing.T) {
	client := &fakeClient{}
	cmd := NewCommand(Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			return client, nil
		},
		RequestTimeout: time.Second,
	})

	projectDir := t.TempDir()
	err := os.WriteFile(filepath.Join(projectDir, "pubspec.yaml"), []byte("name: demo\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write test project file: %v", err)
	}

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"build",
		"appbundle",
		"--project",
		projectDir,
		"--build-mode",
		"profile",
		"--flavor",
		"staging",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if client.lastRequest == nil {
		t.Fatal("expected build request to be sent")
	}

	if client.lastRequest.BuildType != pb.BuildType_BUILD_TYPE_FLUTTER {
		t.Fatalf("unexpected build type: %v", client.lastRequest.BuildType)
	}
	if client.lastRequest.TargetPlatform != pb.TargetPlatform_PLATFORM_ANDROID {
		t.Fatalf("unexpected target platform: %v", client.lastRequest.TargetPlatform)
	}

	flutterConfig := client.lastRequest.GetFlutterConfig()
	if flutterConfig == nil {
		t.Fatal("expected flutter config")
	}
	if flutterConfig.BuildMode != "appbundle-profile" {
		t.Fatalf("unexpected build mode: %s", flutterConfig.BuildMode)
	}
	if flutterConfig.Flavor != "staging" {
		t.Fatalf("unexpected flavor: %s", flutterConfig.Flavor)
	}

	if len(client.lastRequest.SourceArchive) == 0 {
		t.Fatal("expected source archive data")
	}
	if client.lastRequest.SourceHash == "" {
		t.Fatal("expected source hash")
	}
}
