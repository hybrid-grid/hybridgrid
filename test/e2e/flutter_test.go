package e2e

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
)

const flutterBuildCallTimeout = 9 * time.Minute

func TestFlutterE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Flutter E2E test in short mode")
	}

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found, skipping Flutter E2E test")
	}

	if err := runCmd(repoRoot(t), 30*time.Second, "docker", "compose", "version"); err != nil {
		t.Skipf("docker compose not available: %v", err)
	}

	if err := runCmd(repoRoot(t), 30*time.Second, "docker", "image", "inspect", "hybridgrid/flutter-android:3.19.6"); err != nil {
		t.Skipf("required image hybridgrid/flutter-android:3.19.6 not found: %v", err)
	}

	if os.Getenv("INTEGRATION_TEST") != "1" {
		t.Skip("skipping Flutter E2E test unless INTEGRATION_TEST=1 is set")
	}

	root := repoRoot(t)
	appPath := filepath.Join(root, "test", "e2e", "flutter", "testapp")
	if info, err := os.Stat(appPath); err != nil || !info.IsDir() {
		t.Fatalf("flutter test app not found at %s", appPath)
	}

	tmpDir := t.TempDir()
	linuxWorker := filepath.Join(tmpDir, "hg-worker")

	workerEnv := []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
	}
	if err := runCmdWithEnv(root, 2*time.Minute, workerEnv, "go", "build", "-o", linuxWorker, "./cmd/hg-worker"); err != nil {
		t.Fatalf("failed to build linux hg-worker binary: %v", err)
	}

	grpcPort := "19090"
	httpPort := "18080"
	composePath := filepath.Join(tmpDir, "docker-compose.flutter-e2e.yml")
	composeYAML := fmt.Sprintf(`services:
  coordinator:
    build:
      context: %s
      dockerfile: test/e2e/Dockerfile.worker
      args:
        VERSION: dev
        COMMIT: local
        BUILD_TIME: test
    command: hg-coord serve --grpc-port=9000 --http-port=8080
    ports:
      - "%s:9000"
      - "%s:8080"
    healthcheck:
      test: ["CMD-SHELL", "curl -sf http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 6
      start_period: 10s

  flutter-worker:
    image: hybridgrid/flutter-android:3.19.6
    command: /bin/sh -lc "flutter precache --android --android_gen_snapshot && /usr/local/bin/hg-worker serve --coordinator=coordinator:9000 --port=50052 --http-port=9090 --advertise-address=flutter-worker:50052 --max-parallel=1"
    environment:
      - FLUTTER_ROOT=/sdks/flutter
      - ANDROID_SDK_ROOT=/opt/android-sdk-linux
      - ANDROID_HOME=/opt/android-sdk-linux
    volumes:
      - %s:/usr/local/bin/hg-worker:ro
    depends_on:
      coordinator:
        condition: service_healthy
`, root, grpcPort, httpPort, linuxWorker)

	if err := os.WriteFile(composePath, []byte(composeYAML), 0600); err != nil {
		t.Fatalf("failed to write compose file: %v", err)
	}

	composeArgs := []string{"compose", "-f", composePath, "-p", "hg-flutter-e2e", "up", "-d", "--build", "--pull=never"}
	if err := runCmd(root, 15*time.Minute, "docker", composeArgs...); err != nil {
		t.Fatalf("failed to start docker compose stack: %v", err)
	}
	t.Cleanup(func() {
		_ = runCmd(root, 2*time.Minute, "docker", "compose", "-f", composePath, "-p", "hg-flutter-e2e", "down", "-v", "--remove-orphans")
	})

	coordAddr := "localhost:" + grpcPort
	if err := waitForFlutterWorker(t, coordAddr, 6*time.Minute); err != nil {
		t.Fatalf("flutter worker did not become ready: %v", err)
	}

	archive, hash, err := makeFlutterSourceArchive(appPath)
	if err != nil {
		t.Fatalf("failed to prepare flutter source archive: %v", err)
	}

	buildResp, err := runDirectFlutterBuild(coordAddr, archive, hash)
	if err != nil {
		t.Logf("flutter compose log excerpt:\n%s", flutterComposeLogExcerpt(t, root, composePath))
		t.Fatalf("first direct build request failed: %v", err)
	}
	if buildResp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("expected completed first build status, got %s (stderr: %s)", buildResp.Status.String(), buildResp.Stderr)
	}
	if buildResp.FromCache {
		t.Fatal("expected first direct build to be a cache miss")
	}

	secondResp, err := runDirectFlutterBuild(coordAddr, archive, hash)
	if err != nil {
		t.Logf("flutter compose log excerpt:\n%s", flutterComposeLogExcerpt(t, root, composePath))
		t.Fatalf("second direct build request failed: %v", err)
	}
	if secondResp.Status != pb.TaskStatus_STATUS_COMPLETED {
		t.Fatalf("expected completed second build status, got %s (stderr: %s)", secondResp.Status.String(), secondResp.Stderr)
	}
	if !secondResp.FromCache {
		t.Fatal("expected cache hit on second direct build")
	}
	if len(buildResp.Artifacts) == 0 {
		t.Fatal("expected non-empty artifact archive in build response")
	}

	apkData, err := extractFirstAPKFromTar(buildResp.Artifacts)
	if err != nil {
		t.Fatalf("failed to extract apk artifact: %v", err)
	}
	if err := validateAPKZip(apkData); err != nil {
		t.Fatalf("apk zip validation failed: %v", err)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
}

func runCmd(workDir string, timeout time.Duration, name string, args ...string) error {
	_, err := runCmdOutput(workDir, timeout, name, args...)
	return err
}

func runCmdWithEnv(workDir string, timeout time.Duration, extraEnv []string, name string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), extraEnv...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, string(output))
	}
	return nil
}

func runCmdOutput(workDir string, timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = workDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("%s %s failed: %w\n%s", name, strings.Join(args, " "), err, string(output))
	}

	return string(output), nil
}

func waitForFlutterWorker(t *testing.T, coordinatorAddr string, timeout time.Duration) error {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		cli, err := client.New(client.Config{
			Address:  coordinatorAddr,
			Timeout:  5 * time.Second,
			Insecure: true,
		})
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		workersResp, err := cli.GetWorkersForBuild(
			context.Background(),
			pb.BuildType_BUILD_TYPE_FLUTTER,
			pb.TargetPlatform_PLATFORM_ANDROID,
		)
		_ = cli.Close()
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		if workersResp.GetAvailableCount() > 0 {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timed out waiting for flutter-capable worker")
}

func makeFlutterSourceArchive(projectPath string) ([]byte, string, error) {
	root, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if shouldExcludePath(rel, d.IsDir()) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if d.IsDir() || !d.Type().IsRegular() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		fi, err := f.Stat()
		if err != nil {
			_ = f.Close()
			return err
		}

		h, err := tar.FileInfoHeader(fi, "")
		if err != nil {
			_ = f.Close()
			return err
		}
		h.Name = rel

		if err := tw.WriteHeader(h); err != nil {
			_ = f.Close()
			return err
		}

		if _, err := io.Copy(tw, f); err != nil {
			_ = f.Close()
			return err
		}

		if err := f.Close(); err != nil {
			return err
		}

		return nil
	})
	if walkErr != nil {
		_ = tw.Close()
		return nil, "", fmt.Errorf("failed to archive project: %w", walkErr)
	}

	if err := tw.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to finalize archive: %w", err)
	}

	data := buf.Bytes()
	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

func shouldExcludePath(relPath string, isDir bool) bool {
	parts := strings.Split(strings.TrimPrefix(relPath, "./"), "/")
	if len(parts) > 0 {
		switch parts[0] {
		case "build", ".dart_tool", ".git":
			return true
		}
	}

	if strings.HasPrefix(relPath, "android/.gradle") {
		return true
	}
	if strings.HasPrefix(relPath, "android/app/build/") {
		return true
	}

	if !isDir && strings.HasSuffix(relPath, ".DS_Store") {
		return true
	}

	return false
}

func runDirectFlutterBuild(coordinatorAddr string, sourceArchive []byte, sourceHash string) (*pb.BuildResponse, error) {
	cli, err := client.New(client.Config{
		Address:  coordinatorAddr,
		Timeout:  flutterBuildCallTimeout,
		Insecure: true,
	})
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	ctx, cancel := context.WithTimeout(context.Background(), flutterBuildCallTimeout)
	defer cancel()

	return cli.Build(ctx, &pb.BuildRequest{
		TaskId:         fmt.Sprintf("flutter-e2e-validate-%d", time.Now().UnixNano()),
		SourceHash:     sourceHash,
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  sourceArchive,
		Config: &pb.BuildRequest_FlutterConfig{
			FlutterConfig: &pb.FlutterConfig{
				BuildMode: "apk-debug",
			},
		},
		TimeoutSeconds: int32(flutterBuildCallTimeout / time.Second),
	})
}

func flutterComposeLogExcerpt(t *testing.T, root, composePath string) string {
	t.Helper()

	output, err := runCmdOutput(root, 90*time.Second, "docker", "compose", "-f", composePath, "-p", "hg-flutter-e2e", "logs", "--no-color", "coordinator", "flutter-worker")
	if err != nil {
		return fmt.Sprintf("failed to get docker compose logs: %v", err)
	}

	keywords := []string{"flutter pub get", "flutter build", "gradle", "license", "download"}
	lines := strings.Split(output, "\n")
	matches := make([]string, 0, 128)
	for _, line := range lines {
		lower := strings.ToLower(line)
		for _, keyword := range keywords {
			if strings.Contains(lower, keyword) {
				matches = append(matches, line)
				break
			}
		}
		if len(matches) >= 200 {
			break
		}
	}

	if len(matches) == 0 {
		return "no matching flutter/gradle/license/download lines found in compose logs"
	}

	return strings.Join(matches, "\n")
}

func extractFirstAPKFromTar(archive []byte) ([]byte, error) {
	tr := tar.NewReader(bytes.NewReader(archive))
	for {
		h, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read artifact tar: %w", err)
		}
		if h.Typeflag != tar.TypeReg {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(h.Name), ".apk") {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, fmt.Errorf("failed to read apk from tar: %w", err)
		}
		return data, nil
	}

	return nil, fmt.Errorf("no apk file found in artifact archive")
}

func validateAPKZip(apk []byte) error {
	readerAt := bytes.NewReader(apk)
	zr, err := zip.NewReader(readerAt, int64(len(apk)))
	if err != nil {
		return fmt.Errorf("invalid zip structure: %w", err)
	}
	if len(zr.File) == 0 {
		return fmt.Errorf("zip archive has no entries")
	}
	return nil
}
