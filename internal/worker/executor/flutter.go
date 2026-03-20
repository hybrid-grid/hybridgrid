package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const maxFlutterLogBytes = 256 * 1024

type limitedBuffer struct {
	max       int
	buf       bytes.Buffer
	truncated bool
}

func newLimitedBuffer(max int) *limitedBuffer {
	return &limitedBuffer{max: max}
}

func (b *limitedBuffer) Write(p []byte) (int, error) {
	if b.max <= 0 {
		b.truncated = true
		return len(p), nil
	}

	remaining := b.max - b.buf.Len()
	if remaining <= 0 {
		b.truncated = true
		return len(p), nil
	}

	if len(p) > remaining {
		_, _ = b.buf.Write(p[:remaining])
		b.truncated = true
		return len(p), nil
	}

	_, _ = b.buf.Write(p)
	return len(p), nil
}

func (b *limitedBuffer) String() string {
	if !b.truncated {
		return b.buf.String()
	}
	return b.buf.String() + "\n[output truncated]"
}

type flutterOutputType string

const (
	flutterOutputAPK       flutterOutputType = "apk"
	flutterOutputAppBundle flutterOutputType = "appbundle"
)

type FlutterExecutor struct {
	command string
}

func NewFlutterExecutor() *FlutterExecutor {
	return &FlutterExecutor{command: "flutter"}
}

func NewFlutterExecutorWithCommand(command string) *FlutterExecutor {
	return &FlutterExecutor{command: command}
}

func (e *FlutterExecutor) Name() string {
	return "flutter"
}

func (e *FlutterExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	return true
}

func (e *FlutterExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req == nil {
		return nil, fmt.Errorf("build request required")
	}
	if req.BuildType != pb.BuildType_BUILD_TYPE_FLUTTER {
		return nil, fmt.Errorf("unsupported build type: %s", req.BuildType.String())
	}

	flutterConfig := req.FlutterConfig
	if flutterConfig == nil {
		return nil, fmt.Errorf("flutter_config required")
	}

	if req.TargetPlatform != pb.TargetPlatform_PLATFORM_ANDROID {
		return nil, fmt.Errorf("unsupported target platform: %s", req.TargetPlatform.String())
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if req.TimeoutSeconds > 0 {
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-flutter-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)
	defer stopGradleDaemon(workDir)

	if len(req.SourceArchive) > 0 {
		if err := extractSourceArchive(workDir, req.SourceArchive); err != nil {
			return nil, fmt.Errorf("failed to extract source archive: %w", err)
		}
	}

	outputType, modeFlag, err := resolveFlutterBuildOptions(flutterConfig.BuildMode)
	if err != nil {
		return nil, err
	}

	pubGetOut := newLimitedBuffer(maxFlutterLogBytes)
	pubGetErr := newLimitedBuffer(maxFlutterLogBytes)
	if err := runFlutterPubGet(execCtx, e.command, workDir, pubGetOut, pubGetErr); err != nil {
		result := &Result{
			Success:  false,
			ExitCode: 1,
			Stderr:   pubGetErr.String(),
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = int32(exitErr.ExitCode())
		}
		return result, nil
	}

	args := buildFlutterArgs(outputType, modeFlag, flutterConfig)
	cmd := exec.CommandContext(execCtx, e.command, args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GRADLE_OPTS=-Dorg.gradle.vfs.watch=false")

	stdout := newLimitedBuffer(maxFlutterLogBytes)
	stderr := newLimitedBuffer(maxFlutterLogBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err = cmd.Run()
	buildTime := time.Since(start)

	result := &Result{
		Stdout:          pubGetOut.String() + stdout.String(),
		Stderr:          pubGetErr.String() + stderr.String(),
		CompilationTime: buildTime,
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		result.Success = false
		return result, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = int32(exitErr.ExitCode())
			result.Success = false
			return result, nil
		}
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	result.ExitCode = 0

	artifacts, archive, err := collectFlutterArtifacts(workDir, outputType)
	if err != nil {
		result.Success = false
		result.ExitCode = 1
		result.Stderr = strings.TrimSpace(result.Stderr + "\n" + err.Error())
		return result, nil
	}

	result.Success = true
	result.Artifacts = artifacts
	result.ArtifactArchive = archive
	if len(artifacts) > 0 {
		result.ArtifactPath = artifacts[0].Path
	}

	return result, nil
}

func runFlutterPubGet(ctx context.Context, flutterCmd, workDir string, stdout, stderr *limitedBuffer) error {
	cmd := exec.CommandContext(ctx, flutterCmd, "pub", "get")
	cmd.Dir = workDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	cmd.Env = append(os.Environ(), "GRADLE_OPTS=-Dorg.gradle.vfs.watch=false")
	return cmd.Run()
}

// stopGradleDaemon kills the Gradle daemon to ensure build isolation.
// Best-effort: errors are swallowed since daemon cleanup is non-critical.
func stopGradleDaemon(workDir string) {
	gradlewPath := filepath.Join(workDir, "android", "gradlew")
	if _, err := os.Stat(gradlewPath); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, gradlewPath, "--stop")
	cmd.Dir = filepath.Join(workDir, "android")
	_ = cmd.Run()
}

func resolveFlutterBuildOptions(buildMode string) (flutterOutputType, string, error) {
	output := flutterOutputAPK
	mode := ""

	if buildMode != "" {
		tokens := strings.FieldsFunc(strings.ToLower(buildMode), func(r rune) bool {
			switch r {
			case '-', '_', ' ':
				return true
			default:
				return false
			}
		})
		for _, token := range tokens {
			switch token {
			case "apk":
				output = flutterOutputAPK
			case "aab", "appbundle", "bundle":
				output = flutterOutputAppBundle
			case "debug", "profile", "release":
				mode = token
			case "":
				continue
			default:
				return output, "", fmt.Errorf("unsupported build_mode %q", buildMode)
			}
		}
	}

	if mode == "" {
		mode = "release"
	}

	return output, "--" + mode, nil
}

func buildFlutterArgs(outputType flutterOutputType, modeFlag string, cfg *pb.FlutterConfig) []string {
	args := []string{"build", string(outputType)}
	if modeFlag != "" {
		args = append(args, modeFlag)
	}
	if cfg.Flavor != "" {
		args = append(args, "--flavor", cfg.Flavor)
	}
	if len(cfg.DartDefines) > 0 {
		keys := make([]string, 0, len(cfg.DartDefines))
		for key := range cfg.DartDefines {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			args = append(args, "--dart-define", fmt.Sprintf("%s=%s", key, cfg.DartDefines[key]))
		}
	}
	return args
}

func collectFlutterArtifacts(workDir string, outputType flutterOutputType) ([]*pb.ArtifactInfo, []byte, error) {
	searchRoot := filepath.Join(workDir, "build")
	if _, err := os.Stat(searchRoot); err != nil {
		return nil, nil, fmt.Errorf("build output directory not found")
	}

	var wantExt string
	if outputType == flutterOutputAppBundle {
		wantExt = ".aab"
	} else {
		wantExt = ".apk"
	}

	entries := make([]flutterArtifact, 0)
	walkErr := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != wantExt {
			return nil
		}

		relPath, err := filepath.Rel(workDir, path)
		if err != nil {
			return err
		}

		checksum, size, err := checksumFile(path)
		if err != nil {
			return err
		}

		entries = append(entries, flutterArtifact{
			info: &pb.ArtifactInfo{
				Name:      filepath.Base(path),
				Path:      filepath.ToSlash(relPath),
				SizeBytes: size,
				Checksum:  checksum,
			},
			absPath: path,
		})
		return nil
	})
	if walkErr != nil {
		return nil, nil, fmt.Errorf("failed to scan artifacts: %w", walkErr)
	}

	if len(entries) == 0 {
		return nil, nil, fmt.Errorf("no %s artifacts found", wantExt)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].info.Path < entries[j].info.Path
	})

	archive, err := archiveArtifacts(workDir, entries)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to archive artifacts: %w", err)
	}

	artifactList := make([]*pb.ArtifactInfo, 0, len(entries))
	for _, entry := range entries {
		artifactList = append(artifactList, entry.info)
	}

	return artifactList, archive, nil
}

type flutterArtifact struct {
	info    *pb.ArtifactInfo
	absPath string
}

func archiveArtifacts(root string, entries []flutterArtifact) ([]byte, error) {
	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)
	defer tarWriter.Close()

	for _, entry := range entries {
		file, err := os.Open(entry.absPath)
		if err != nil {
			return nil, err
		}

		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return nil, err
		}

		relPath, err := filepath.Rel(root, entry.absPath)
		if err != nil {
			_ = file.Close()
			return nil, err
		}

		header := &tar.Header{
			Name:    filepath.ToSlash(relPath),
			Mode:    int64(info.Mode().Perm()),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			_ = file.Close()
			return nil, err
		}

		if _, err := io.Copy(tarWriter, file); err != nil {
			_ = file.Close()
			return nil, err
		}

		_ = file.Close()
	}

	if err := tarWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func checksumFile(path string) (string, int64, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	hash := sha256.New()
	written, err := io.Copy(hash, file)
	if err != nil {
		return "", 0, err
	}

	return hex.EncodeToString(hash.Sum(nil)), written, nil
}

func extractSourceArchive(dest string, data []byte) error {
	if len(data) == 0 {
		return nil
	}

	if isZstdArchive(data) {
		zstdPath, err := exec.LookPath("zstd")
		if err != nil {
			return fmt.Errorf("zstd not found for compressed archive")
		}

		cmd := exec.Command(zstdPath, "-d", "-q", "-c")
		cmd.Stdin = bytes.NewReader(data)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to create decompressor stdout pipe: %w", err)
		}

		var stderr bytes.Buffer
		cmd.Stderr = &stderr

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start zstd decompressor: %w", err)
		}

		if err := extractTarStream(dest, tar.NewReader(stdout)); err != nil {
			_ = cmd.Wait()
			return err
		}

		if err := cmd.Wait(); err != nil {
			return fmt.Errorf("zstd decompression failed: %v", err)
		}

		return nil
	}

	return extractTarStream(dest, tar.NewReader(bytes.NewReader(data)))
}

func extractTarStream(dest string, reader *tar.Reader) error {
	for {
		header, err := reader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read tar archive: %w", err)
		}

		name := filepath.Clean(header.Name)
		if name == "." {
			continue
		}
		if filepath.IsAbs(name) || strings.HasPrefix(name, "..") {
			return fmt.Errorf("invalid archive path: %s", header.Name)
		}

		path := filepath.Join(dest, name)
		if !isWithinDir(dest, path) {
			return fmt.Errorf("invalid archive path: %s", header.Name)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return fmt.Errorf("failed to create parent dir: %w", err)
			}
			file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return fmt.Errorf("failed to create file: %w", err)
			}
			if _, err := io.Copy(file, reader); err != nil {
				_ = file.Close()
				return fmt.Errorf("failed to write file: %w", err)
			}
			if err := file.Close(); err != nil {
				return fmt.Errorf("failed to close file: %w", err)
			}
		default:
			continue
		}
	}

	return nil
}

func isWithinDir(base, target string) bool {
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if rel == ".." {
		return false
	}
	return !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func isZstdArchive(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	return data[0] == 0x28 && data[1] == 0xB5 && data[2] == 0x2F && data[3] == 0xFD
}
