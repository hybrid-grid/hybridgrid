package executor

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

var _ Executor = (*UnityExecutor)(nil)

func TestNewUnityExecutor(t *testing.T) {
	executor := NewUnityExecutor()
	if executor == nil {
		t.Fatal("NewUnityExecutor() returned nil")
	}
	if strings.TrimSpace(executor.command) == "" {
		t.Fatal("NewUnityExecutor() returned empty command")
	}
}

func TestNewUnityExecutorWithCommand(t *testing.T) {
	executor := NewUnityExecutorWithCommand("/tmp/custom-unity")
	if executor == nil {
		t.Fatal("NewUnityExecutorWithCommand() returned nil")
	}
	if executor.command != "/tmp/custom-unity" {
		t.Fatalf("executor.command = %q, want %q", executor.command, "/tmp/custom-unity")
	}
}

func TestUnityExecutor_Name(t *testing.T) {
	executor := NewUnityExecutorWithCommand("Unity")
	if executor.Name() != "unity" {
		t.Fatalf("Name() = %q, want %q", executor.Name(), "unity")
	}
}

func TestUnityExecutor_CanExecute(t *testing.T) {
	executor := NewUnityExecutorWithCommand("Unity")
	if !executor.CanExecute(pb.Architecture_ARCH_ARM64, pb.Architecture_ARCH_X86_64) {
		t.Fatal("CanExecute() = false, want true")
	}
}

func TestUnityExecutor_Execute_ValidatesBuildType(t *testing.T) {
	executor := NewUnityExecutorWithCommand("Unity")
	_, err := executor.Execute(context.Background(), &Request{
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		UnityConfig:    &pb.UnityConfig{BuildMethod: "BuildScript.Build"},
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
	})
	if err == nil {
		t.Fatal("Execute() expected error for non-unity build type")
	}
}

func TestUnityExecutor_Execute_ValidatesConfigPresent(t *testing.T) {
	executor := NewUnityExecutorWithCommand("Unity")
	_, err := executor.Execute(context.Background(), &Request{
		BuildType:      pb.BuildType_BUILD_TYPE_UNITY,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
	})
	if err == nil {
		t.Fatal("Execute() expected error for missing unity config")
	}
}

func TestBuildTargetToUnityString(t *testing.T) {
	tests := []struct {
		name      string
		target    pb.TargetPlatform
		want      string
		wantError bool
	}{
		{name: "android", target: pb.TargetPlatform_PLATFORM_ANDROID, want: "Android"},
		{name: "ios", target: pb.TargetPlatform_PLATFORM_IOS, want: "iOS"},
		{name: "windows", target: pb.TargetPlatform_PLATFORM_WINDOWS, want: "StandaloneWindows64"},
		{name: "linux", target: pb.TargetPlatform_PLATFORM_LINUX, want: "StandaloneLinux64"},
		{name: "macos", target: pb.TargetPlatform_PLATFORM_MACOS, want: "StandaloneOSX"},
		{name: "webgl", target: pb.TargetPlatform_PLATFORM_WEBGL, want: "WebGL"},
		{name: "unspecified", target: pb.TargetPlatform_PLATFORM_UNSPECIFIED, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildTargetToUnityString(tt.target)
			if tt.wantError {
				if err == nil {
					t.Fatal("buildTargetToUnityString() expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("buildTargetToUnityString() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("buildTargetToUnityString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestUnityExecutor_Execute_CommandConstruction(t *testing.T) {
	unityCmd, argsFile := setupFakeUnity(t)
	archive := writeTarArchive(t, map[string]string{
		"ProjectSettings/ProjectVersion.txt": "m_EditorVersion: 2022.3.10f1",
	})

	executor := NewUnityExecutorWithCommand(unityCmd)
	req := &Request{
		TaskID:         "unity-build",
		BuildType:      pb.BuildType_BUILD_TYPE_UNITY,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		UnityConfig: &pb.UnityConfig{
			BuildMethod:      "BuildScript.Build",
			ScriptingBackend: "il2cpp",
			ExtraArgs: map[string]string{
				"-customArg": "customValue",
				"-flagOnly":  "",
			},
		},
		TimeoutSeconds: 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Execute() expected success, stderr=%q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Execute() exit code = %d, want 0", result.ExitCode)
	}
	if len(result.Artifacts) == 0 {
		t.Fatal("Execute() returned no artifacts")
	}
	if !strings.HasSuffix(result.ArtifactPath, ".apk") {
		t.Fatalf("ArtifactPath = %q, want suffix .apk", result.ArtifactPath)
	}

	args := readArgsFile(t, argsFile)
	assertArgsContain(t, args,
		"-batchmode",
		"-quit",
		"-nographics",
		"-projectPath",
		"-logFile",
		"-buildTarget",
		"Android",
		"-executeMethod",
		"BuildScript.Build",
		"-scripting-backend",
		"il2cpp",
		"-customArg",
		"customValue",
		"-flagOnly",
	)
	assertArgOrder(t, args, "-customArg", "-flagOnly")
}

func TestManager_SelectForRequest_UnityRoute(t *testing.T) {
	m := &Manager{
		native: &fakeExecutor{name: "native"},
		unity:  &fakeExecutor{name: "unity"},
	}

	got := m.SelectForRequest(&Request{BuildType: pb.BuildType_BUILD_TYPE_UNITY})
	if got.Name() != "unity" {
		t.Fatalf("SelectForRequest() executor = %s, want unity", got.Name())
	}
}

func setupFakeUnity(t *testing.T) (string, string) {
	binDir := t.TempDir()
	argsFile := filepath.Join(t.TempDir(), "unity-args.txt")

	var scriptPath string
	var contents []byte

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(binDir, "unity.bat")
		contents = []byte("@echo off\r\n" +
			"set PROJECT=\r\n" +
			"set LOGFILE=\r\n" +
			"set TARGET=\r\n" +
			"if not \"%HG_UNITY_ARGS_FILE%\"==\"\" (\r\n" +
			"  for %%a in (%*) do echo %%a>>\"%HG_UNITY_ARGS_FILE%\"\r\n" +
			")\r\n" +
			":loop\r\n" +
			"if \"%1\"==\"\" goto done\r\n" +
			"if \"%1\"==\"-projectPath\" (\r\n" +
			"  set PROJECT=%2\r\n" +
			")\r\n" +
			"if \"%1\"==\"-logFile\" (\r\n" +
			"  set LOGFILE=%2\r\n" +
			")\r\n" +
			"if \"%1\"==\"-buildTarget\" (\r\n" +
			"  set TARGET=%2\r\n" +
			")\r\n" +
			"shift\r\n" +
			"goto loop\r\n" +
			":done\r\n" +
			"if not \"%LOGFILE%\"==\"\" echo unity-log>\"%LOGFILE%\"\r\n" +
			"if not \"%PROJECT%\"==\"\" (\r\n" +
			"  mkdir \"%PROJECT%\\Build\" 2>nul\r\n" +
			"  if \"%TARGET%\"==\"Android\" (\r\n" +
			"    echo fake apk>\"%PROJECT%\\Build\\app-release.apk\"\r\n" +
			"  )\r\n" +
			")\r\n" +
			"if not \"%HG_UNITY_EXIT_CODE%\"==\"\" exit /B %HG_UNITY_EXIT_CODE%\r\n" +
			"exit /B 0\r\n")
	} else {
		scriptPath = filepath.Join(binDir, "unity")
		contents = []byte("#!/bin/sh\n" +
			"set -e\n" +
			"if [ -n \"$HG_UNITY_ARGS_FILE\" ]; then\n" +
			"  printf '%s\\n' \"$@\" > \"$HG_UNITY_ARGS_FILE\"\n" +
			"fi\n" +
			"PROJECT=\"\"\n" +
			"LOG_FILE=\"\"\n" +
			"TARGET=\"\"\n" +
			"while [ $# -gt 0 ]; do\n" +
			"  case \"$1\" in\n" +
			"    -projectPath) PROJECT=\"$2\"; shift 2 ;;\n" +
			"    -logFile) LOG_FILE=\"$2\"; shift 2 ;;\n" +
			"    -buildTarget) TARGET=\"$2\"; shift 2 ;;\n" +
			"    *) shift ;;\n" +
			"  esac\n" +
			"done\n" +
			"if [ -n \"$LOG_FILE\" ]; then\n" +
			"  printf 'unity-log' > \"$LOG_FILE\"\n" +
			"fi\n" +
			"if [ -n \"$PROJECT\" ]; then\n" +
			"  mkdir -p \"$PROJECT/Build\"\n" +
			"  if [ \"$TARGET\" = \"Android\" ]; then\n" +
			"    printf 'fake apk' > \"$PROJECT/Build/app-release.apk\"\n" +
			"  fi\n" +
			"fi\n" +
			"if [ -n \"$HG_UNITY_EXIT_CODE\" ]; then\n" +
			"  exit \"$HG_UNITY_EXIT_CODE\"\n" +
			"fi\n" +
			"exit 0\n")
	}

	if err := os.WriteFile(scriptPath, contents, 0755); err != nil {
		t.Fatalf("Failed to write fake unity: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(scriptPath, 0755); err != nil {
			t.Fatalf("Failed to chmod fake unity: %v", err)
		}
	}

	t.Setenv("HG_UNITY_ARGS_FILE", argsFile)
	return scriptPath, argsFile
}
