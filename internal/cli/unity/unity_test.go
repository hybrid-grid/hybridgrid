package unity

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
	return &pb.BuildResponse{Status: pb.TaskStatus_STATUS_COMPLETED}, nil
}

func (f *fakeClient) Close() error {
	return nil
}

func TestNewCommand_CreatesCommandTree(t *testing.T) {
	cmd := NewCommand(Dependencies{})
	if cmd == nil {
		t.Fatal("expected command")
	}
	if cmd.Use != "unity" {
		t.Fatalf("unexpected use: %s", cmd.Use)
	}

	buildCmd, _, err := cmd.Find([]string{"build"})
	if err != nil {
		t.Fatalf("expected build command: %v", err)
	}
	if buildCmd == nil || !strings.HasPrefix(buildCmd.Use, "build") {
		t.Fatalf("unexpected build command: %#v", buildCmd)
	}
}

func TestParseTargetPlatform_Valid(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected pb.TargetPlatform
	}{
		{name: "android", input: "android", expected: pb.TargetPlatform_PLATFORM_ANDROID},
		{name: "ios", input: "ios", expected: pb.TargetPlatform_PLATFORM_IOS},
		{name: "windows", input: "windows", expected: pb.TargetPlatform_PLATFORM_WINDOWS},
		{name: "linux", input: "linux", expected: pb.TargetPlatform_PLATFORM_LINUX},
		{name: "macos", input: "macos", expected: pb.TargetPlatform_PLATFORM_MACOS},
		{name: "webgl", input: "webgl", expected: pb.TargetPlatform_PLATFORM_WEBGL},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTargetPlatform(tt.input)
			if err != nil {
				t.Fatalf("parseTargetPlatform(%q) returned error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Fatalf("unexpected platform: got %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestParseTargetPlatform_Invalid(t *testing.T) {
	_, err := parseTargetPlatform("playstation")
	if err == nil {
		t.Fatal("expected error for invalid platform")
	}
	if !strings.Contains(err.Error(), "invalid platform") {
		t.Fatalf("expected invalid platform error, got %v", err)
	}
}

func TestUnityCLI_RequiresBuildMethod(t *testing.T) {
	client := &fakeClient{}
	cmd := NewCommand(Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			return client, nil
		},
		RequestTimeout: time.Second,
	})

	projectDir := t.TempDir()
	err := os.WriteFile(filepath.Join(projectDir, "Assets.meta"), []byte(""), 0644)
	if err != nil {
		t.Fatalf("failed to write test project file: %v", err)
	}

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"build", "android", "--project", projectDir})

	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing build-method")
	}
	if !strings.Contains(err.Error(), "build-method") {
		t.Fatalf("expected build-method error, got %v", err)
	}
}

func TestUnityCLI_RequiresProject(t *testing.T) {
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
	cmd.SetArgs([]string{"build", "android", "--build-method", "BuildScript.Build"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing project")
	}
	if !strings.Contains(err.Error(), "project") {
		t.Fatalf("expected project error, got %v", err)
	}
}

func TestUnityCLI_ProjectPathValidation(t *testing.T) {
	client := &fakeClient{}
	cmd := NewCommand(Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			return client, nil
		},
		RequestTimeout: time.Second,
	})

	missingProject := filepath.Join(t.TempDir(), "missing")

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"build", "android", "--project", missingProject, "--build-method", "BuildScript.Build"})

	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing project path")
	}
	if !strings.Contains(err.Error(), "project not found") {
		t.Fatalf("expected project not found error, got %v", err)
	}
}

func TestUnityCLI_BuildRequestUsesUnityConfig(t *testing.T) {
	client := &fakeClient{}
	cmd := NewCommand(Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			return client, nil
		},
		RequestTimeout: time.Second,
	})

	projectDir := t.TempDir()
	err := os.WriteFile(filepath.Join(projectDir, "ProjectSettings.asset"), []byte("player settings"), 0644)
	if err != nil {
		t.Fatalf("failed to write test project file: %v", err)
	}

	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{
		"build",
		"android",
		"--project",
		projectDir,
		"--build-method",
		"BuildScript.Build",
		"--unity-version",
		"2022.3.0f1",
		"--scripting-backend",
		"il2cpp",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if client.lastRequest == nil {
		t.Fatal("expected build request to be sent")
	}

	if client.lastRequest.BuildType != pb.BuildType_BUILD_TYPE_UNITY {
		t.Fatalf("unexpected build type: %v", client.lastRequest.BuildType)
	}
	if client.lastRequest.TargetPlatform != pb.TargetPlatform_PLATFORM_ANDROID {
		t.Fatalf("unexpected target platform: %v", client.lastRequest.TargetPlatform)
	}

	unityConfig := client.lastRequest.GetUnityConfig()
	if unityConfig == nil {
		t.Fatal("expected unity config")
	}
	if unityConfig.BuildMethod != "BuildScript.Build" {
		t.Fatalf("unexpected build method: %s", unityConfig.BuildMethod)
	}
	if unityConfig.UnityVersion != "2022.3.0f1" {
		t.Fatalf("unexpected unity version: %s", unityConfig.UnityVersion)
	}
	if unityConfig.ScriptingBackend != "il2cpp" {
		t.Fatalf("unexpected scripting backend: %s", unityConfig.ScriptingBackend)
	}

	if len(client.lastRequest.SourceArchive) == 0 {
		t.Fatal("expected source archive data")
	}
	if client.lastRequest.SourceHash == "" {
		t.Fatal("expected source hash")
	}
}

func TestRunBuild_UsesBuildClient(t *testing.T) {
	client := &fakeClient{}
	deps := Dependencies{
		CoordinatorAddr: func() string { return "coordinator:9000" },
		NewClient: func(address string, timeout time.Duration) (BuildClient, error) {
			if address != "coordinator:9000" {
				t.Fatalf("unexpected address: %s", address)
			}
			return client, nil
		},
		RequestTimeout: time.Second,
	}

	req := &pb.BuildRequest{TaskId: "task-123"}
	resp, err := runBuild(context.Background(), deps, req)
	if err != nil {
		t.Fatalf("runBuild failed: %v", err)
	}
	if resp == nil {
		t.Fatal("expected response")
	}
	if client.lastRequest != req {
		t.Fatal("expected request to be passed to client")
	}
}
