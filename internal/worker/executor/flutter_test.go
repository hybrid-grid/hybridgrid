package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

var _ Executor = (*FlutterExecutor)(nil)

func TestFlutterExecutor_Execute_APKSuccess(t *testing.T) {
	argsFile := setupFakeFlutter(t)
	archive := writeTarArchive(t, map[string]string{
		"pubspec.yaml": "name: demo",
	})

	req := &Request{
		TaskID:         "flutter-apk",
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		FlutterConfig: &pb.FlutterConfig{
			BuildMode: "release",
			Flavor:    "demo",
			DartDefines: map[string]string{
				"FOO": "BAR",
				"ABC": "123",
			},
		},
	}

	executor := NewFlutterExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Expected success, stderr=%q", result.Stderr)
	}
	if result.ExitCode != 0 {
		t.Fatalf("Expected exit code 0, got %d", result.ExitCode)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(result.Artifacts))
	}
	if !strings.HasSuffix(result.Artifacts[0].Name, ".apk") {
		t.Fatalf("Expected apk artifact, got %q", result.Artifacts[0].Name)
	}
	if !strings.HasSuffix(result.ArtifactPath, ".apk") {
		t.Fatalf("Expected ArtifactPath ending in .apk, got %q", result.ArtifactPath)
	}

	args := readArgsFile(t, argsFile)
	assertArgsContain(t, args, "build", "apk", "--release", "--flavor", "demo", "--dart-define", "ABC=123", "FOO=BAR")
	assertArgOrder(t, args, "ABC=123", "FOO=BAR")
}

func TestFlutterExecutor_Execute_AppBundleSuccess(t *testing.T) {
	argsFile := setupFakeFlutter(t)
	archive := writeTarArchive(t, map[string]string{
		"pubspec.yaml": "name: demo",
	})

	req := &Request{
		TaskID:         "flutter-aab",
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		FlutterConfig: &pb.FlutterConfig{
			BuildMode: "release-aab",
		},
	}

	executor := NewFlutterExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Expected success, stderr=%q", result.Stderr)
	}
	if len(result.Artifacts) != 1 {
		t.Fatalf("Expected 1 artifact, got %d", len(result.Artifacts))
	}
	if !strings.HasSuffix(result.Artifacts[0].Name, ".aab") {
		t.Fatalf("Expected aab artifact, got %q", result.Artifacts[0].Name)
	}
	if !strings.HasSuffix(result.ArtifactPath, ".aab") {
		t.Fatalf("Expected ArtifactPath ending in .aab, got %q", result.ArtifactPath)
	}

	args := readArgsFile(t, argsFile)
	assertArgsContain(t, args, "build", "appbundle", "--release")
}

func TestFlutterExecutor_Execute_CommandError(t *testing.T) {
	argsFile := setupFakeFlutter(t)
	archive := writeTarArchive(t, map[string]string{
		"pubspec.yaml": "name: demo",
	})

	t.Setenv("HG_FLUTTER_EXIT_CODE", "2")

	req := &Request{
		TaskID:         "flutter-error",
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		FlutterConfig: &pb.FlutterConfig{
			BuildMode: "release",
		},
	}

	executor := NewFlutterExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if result.Success {
		t.Fatalf("Expected failure, stdout=%q", result.Stdout)
	}
	if result.ExitCode != 2 {
		t.Fatalf("Expected exit code 2, got %d", result.ExitCode)
	}
	if _, err := os.Stat(argsFile); err != nil {
		t.Fatalf("Expected args file, error=%v", err)
	}
}

func TestFlutterExecutor_Execute_GradleDaemonStopped(t *testing.T) {
	setupFakeFlutter(t)
	gradleStopMarker := filepath.Join(t.TempDir(), "gradle-stop-called")
	t.Setenv("HG_GRADLE_STOP_MARKER", gradleStopMarker)

	archive := writeTarArchiveWithGradlew(t, gradleStopMarker)

	req := &Request{
		TaskID:         "flutter-gradle-stop",
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		FlutterConfig: &pb.FlutterConfig{
			BuildMode: "release",
		},
	}

	executor := NewFlutterExecutor()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !result.Success {
		t.Fatalf("Expected success, stderr=%q", result.Stderr)
	}

	if _, err := os.Stat(gradleStopMarker); err != nil {
		t.Fatalf("Expected gradle --stop to be called (marker file missing): %v", err)
	}

	data, err := os.ReadFile(gradleStopMarker)
	if err != nil {
		t.Fatalf("Failed to read marker file: %v", err)
	}
	if !strings.Contains(string(data), "--stop") {
		t.Fatalf("Expected --stop argument, got: %q", string(data))
	}
}

func setupFakeFlutter(t *testing.T) string {
	binDir := t.TempDir()
	argsFile := filepath.Join(t.TempDir(), "args.txt")

	var scriptPath string
	var contents []byte

	if runtime.GOOS == "windows" {
		scriptPath = filepath.Join(binDir, "flutter.bat")
		contents = []byte("@echo off\r\n" +
			"if not \"%HG_FLUTTER_ARGS_FILE%\"==\"\" (\r\n" +
			"  echo %* > \"%HG_FLUTTER_ARGS_FILE%\"\r\n" +
			")\r\n" +
			"if \"%1\"==\"build\" (\r\n" +
			"  if \"%2\"==\"appbundle\" (\r\n" +
			"    mkdir build\\app\\outputs\\bundle\\release\r\n" +
			"    echo fake aab > build\\app\\outputs\\bundle\\release\\app-release.aab\r\n" +
			"  ) else (\r\n" +
			"    mkdir build\\app\\outputs\\flutter-apk\r\n" +
			"    echo fake apk > build\\app\\outputs\\flutter-apk\\app-release.apk\r\n" +
			"  )\r\n" +
			")\r\n" +
			"if not \"%HG_FLUTTER_EXIT_CODE%\"==\"\" exit /B %HG_FLUTTER_EXIT_CODE%\r\n" +
			"exit /B 0\r\n")
	} else {
		scriptPath = filepath.Join(binDir, "flutter")
		contents = []byte("#!/bin/sh\n" +
			"set -e\n" +
			"if [ -n \"$HG_FLUTTER_ARGS_FILE\" ]; then\n" +
			"  printf '%s\\n' \"$@\" > \"$HG_FLUTTER_ARGS_FILE\"\n" +
			"fi\n" +
			"if [ \"$1\" = \"build\" ]; then\n" +
			"  if [ \"$2\" = \"appbundle\" ]; then\n" +
			"    mkdir -p build/app/outputs/bundle/release\n" +
			"    printf 'fake aab' > build/app/outputs/bundle/release/app-release.aab\n" +
			"  else\n" +
			"    mkdir -p build/app/outputs/flutter-apk\n" +
			"    printf 'fake apk' > build/app/outputs/flutter-apk/app-release.apk\n" +
			"  fi\n" +
			"fi\n" +
			"if [ -n \"$HG_FLUTTER_EXIT_CODE\" ]; then\n" +
			"  exit \"$HG_FLUTTER_EXIT_CODE\"\n" +
			"fi\n" +
			"exit 0\n")
	}

	if err := os.WriteFile(scriptPath, contents, 0755); err != nil {
		t.Fatalf("Failed to write fake flutter: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(scriptPath, 0755); err != nil {
			t.Fatalf("Failed to chmod fake flutter: %v", err)
		}
	}

	pathEnv := binDir + string(os.PathListSeparator) + os.Getenv("PATH")
	t.Setenv("PATH", pathEnv)
	t.Setenv("HG_FLUTTER_ARGS_FILE", argsFile)
	return argsFile
}

func writeTarArchive(t *testing.T, files map[string]string) []byte {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	for name, content := range files {
		header := &tar.Header{
			Name: name,
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader error: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	return buf.Bytes()
}

func writeTarArchiveWithGradlew(t *testing.T, stopMarkerPath string) []byte {
	var gradlewScript string
	var gradlewName string

	if runtime.GOOS == "windows" {
		gradlewName = "android/gradlew.bat"
		gradlewScript = "@echo off\r\necho %* > \"" + stopMarkerPath + "\"\r\nexit /B 0\r\n"
	} else {
		gradlewName = "android/gradlew"
		gradlewScript = "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"" + stopMarkerPath + "\"\nexit 0\n"
	}

	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	files := map[string]string{
		"pubspec.yaml": "name: demo",
		gradlewName:    gradlewScript,
	}

	for name, content := range files {
		mode := int64(0644)
		if strings.Contains(name, "gradlew") {
			mode = 0755
		}
		header := &tar.Header{
			Name: name,
			Mode: mode,
			Size: int64(len(content)),
		}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("WriteHeader error: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("Write error: %v", err)
		}
	}

	if err := tarWriter.Close(); err != nil {
		t.Fatalf("Close error: %v", err)
	}

	return buf.Bytes()
}

func readArgsFile(t *testing.T, argsFile string) []string {
	data, err := os.ReadFile(argsFile)
	if err != nil {
		t.Fatalf("Read args file error: %v", err)
	}
	content := strings.TrimSpace(string(data))
	if content == "" {
		return nil
	}
	return strings.Split(content, "\n")
}

func assertArgsContain(t *testing.T, args []string, expected ...string) {
	for _, want := range expected {
		found := false
		for _, arg := range args {
			if arg == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("args missing %q in %v", want, args)
		}
	}
}

func assertArgOrder(t *testing.T, args []string, first, second string) {
	firstIndex := -1
	secondIndex := -1
	for i, arg := range args {
		if arg == first {
			firstIndex = i
		}
		if arg == second {
			secondIndex = i
		}
	}
	if firstIndex == -1 || secondIndex == -1 {
		t.Fatalf("args missing order check entries %q/%q in %v", first, second, args)
	}
	if firstIndex > secondIndex {
		t.Fatalf("expected %q before %q in %v", first, second, args)
	}
}
