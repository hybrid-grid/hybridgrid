package executor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const maxUnityLogBytes = 256 * 1024

type UnityExecutor struct {
	command string
}

func NewUnityExecutor() *UnityExecutor {
	command := detectUnityCommand()
	if command == "" {
		command = "Unity"
	}

	return &UnityExecutor{command: command}
}

func NewUnityExecutorWithCommand(command string) *UnityExecutor {
	return &UnityExecutor{command: command}
}

func (e *UnityExecutor) Name() string {
	return "unity"
}

func (e *UnityExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	return true
}

func (e *UnityExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if req == nil {
		return nil, fmt.Errorf("build request required")
	}
	if req.BuildType != pb.BuildType_BUILD_TYPE_UNITY {
		return nil, fmt.Errorf("unsupported build type: %s", req.BuildType.String())
	}

	unityConfig := req.UnityConfig
	if unityConfig == nil {
		return nil, fmt.Errorf("unity_config required")
	}
	if strings.TrimSpace(unityConfig.BuildMethod) == "" {
		return nil, fmt.Errorf("unity_config.build_method required")
	}

	execCtx := ctx
	var cancel context.CancelFunc
	if req.Timeout > 0 {
		execCtx, cancel = context.WithTimeout(ctx, req.Timeout)
		defer cancel()
	} else if req.TimeoutSeconds > 0 {
		execCtx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutSeconds)*time.Second)
		defer cancel()
	}

	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-unity-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	if len(req.SourceArchive) > 0 {
		if err := extractSourceArchive(workDir, req.SourceArchive); err != nil {
			return nil, fmt.Errorf("failed to extract source archive: %w", err)
		}
	}

	logPath := filepath.Join(workDir, "unity-build.log")
	args, err := buildUnityArgs(req, unityConfig, workDir, logPath)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(execCtx, e.command, args...)
	cmd.Dir = workDir

	stdout := newLimitedBuffer(maxUnityLogBytes)
	stderr := newLimitedBuffer(maxUnityLogBytes)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	start := time.Now()
	err = cmd.Run()
	buildTime := time.Since(start)

	unityLog := readUnityLog(logPath)
	result := &Result{
		Stdout:          strings.TrimSpace(stdout.String()),
		Stderr:          strings.TrimSpace(stderr.String()),
		CompilationTime: buildTime,
	}
	if unityLog != "" {
		result.Stdout = strings.TrimSpace(joinNonEmpty(result.Stdout, unityLog))
	}

	if execCtx.Err() == context.DeadlineExceeded {
		result.ExitCode = -1
		result.Success = false
		result.Stderr = strings.TrimSpace(joinNonEmpty(result.Stderr, "unity build timed out"))
		if unityLog != "" {
			result.Stderr = strings.TrimSpace(joinNonEmpty(result.Stderr, unityLog))
		}
		return result, nil
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = int32(exitErr.ExitCode())
			result.Success = false
			if unityLog != "" {
				result.Stderr = strings.TrimSpace(joinNonEmpty(result.Stderr, unityLog))
			}
			return result, nil
		}
		return nil, fmt.Errorf("execution failed: %w", err)
	}

	result.ExitCode = 0

	artifacts, archive, err := collectUnityArtifacts(workDir, req.TargetPlatform)
	if err != nil {
		result.Success = false
		result.ExitCode = 1
		result.Stderr = strings.TrimSpace(joinNonEmpty(result.Stderr, err.Error()))
		if unityLog != "" {
			result.Stderr = strings.TrimSpace(joinNonEmpty(result.Stderr, unityLog))
		}
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

func buildUnityArgs(req *Request, cfg *pb.UnityConfig, projectPath, logPath string) ([]string, error) {
	buildTarget, err := buildTargetToUnityString(req.TargetPlatform)
	if err != nil {
		return nil, err
	}

	args := []string{
		"-batchmode",
		"-quit",
		"-nographics",
		"-projectPath", projectPath,
		"-logFile", logPath,
		"-buildTarget", buildTarget,
		"-executeMethod", cfg.BuildMethod,
	}

	if cfg.ScriptingBackend != "" {
		args = append(args, "-scripting-backend", cfg.ScriptingBackend)
	}

	if len(cfg.ExtraArgs) > 0 {
		keys := make([]string, 0, len(cfg.ExtraArgs))
		for key := range cfg.ExtraArgs {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			value := cfg.ExtraArgs[key]
			args = append(args, key)
			if value != "" {
				args = append(args, value)
			}
		}
	}

	return args, nil
}

func buildTargetToUnityString(target pb.TargetPlatform) (string, error) {
	switch target {
	case pb.TargetPlatform_PLATFORM_ANDROID:
		return "Android", nil
	case pb.TargetPlatform_PLATFORM_IOS:
		return "iOS", nil
	case pb.TargetPlatform_PLATFORM_WINDOWS:
		return "StandaloneWindows64", nil
	case pb.TargetPlatform_PLATFORM_LINUX:
		return "StandaloneLinux64", nil
	case pb.TargetPlatform_PLATFORM_MACOS:
		return "StandaloneOSX", nil
	case pb.TargetPlatform_PLATFORM_WEBGL:
		return "WebGL", nil
	default:
		return "", fmt.Errorf("unsupported target platform: %s", target.String())
	}
}

func collectUnityArtifacts(workDir string, target pb.TargetPlatform) ([]*pb.ArtifactInfo, []byte, error) {
	searchRoot := filepath.Join(workDir, "Build")
	if _, err := os.Stat(searchRoot); err != nil {
		searchRoot = workDir
	}

	extensions := unityArtifactExtensions(target)
	entries := make([]flutterArtifact, 0)

	walkErr := filepath.WalkDir(searchRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(workDir, path)
		if err != nil {
			return err
		}

		if !isUnityArtifactMatch(path, relPath, target, extensions) {
			return nil
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
		return nil, nil, fmt.Errorf("no artifacts found for target platform: %s", target.String())
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

func unityArtifactExtensions(target pb.TargetPlatform) []string {
	switch target {
	case pb.TargetPlatform_PLATFORM_ANDROID:
		return []string{".apk", ".aab"}
	case pb.TargetPlatform_PLATFORM_IOS:
		return []string{".ipa"}
	case pb.TargetPlatform_PLATFORM_WINDOWS:
		return []string{".exe"}
	case pb.TargetPlatform_PLATFORM_LINUX:
		return []string{".x86_64", ".x86"}
	case pb.TargetPlatform_PLATFORM_MACOS:
		return []string{".app"}
	case pb.TargetPlatform_PLATFORM_WEBGL:
		return []string{".html", ".js", ".wasm", ".data"}
	default:
		return nil
	}
}

func isUnityArtifactMatch(absPath, relPath string, target pb.TargetPlatform, extensions []string) bool {
	if target == pb.TargetPlatform_PLATFORM_WEBGL {
		if strings.Contains(filepath.ToSlash(relPath), "Build/") {
			return true
		}
	}

	ext := strings.ToLower(filepath.Ext(absPath))
	for _, candidate := range extensions {
		if ext == strings.ToLower(candidate) {
			return true
		}
	}

	return false
}

func detectUnityCommand() string {
	patterns := unityEditorPathPatterns()
	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil || len(matches) == 0 {
			continue
		}
		sort.Strings(matches)
		for i := len(matches) - 1; i >= 0; i-- {
			if _, statErr := os.Stat(matches[i]); statErr == nil {
				return matches[i]
			}
		}
	}

	if path, err := exec.LookPath("Unity"); err == nil {
		return path
	}
	if path, err := exec.LookPath("unity"); err == nil {
		return path
	}

	return ""
}

func unityEditorPathPatterns() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/Applications/Unity/Hub/Editor/*/Unity.app/Contents/MacOS/Unity",
			"/Applications/Unity/Hub/Editor/*/Editor/Unity",
			"/Applications/Unity.app/Contents/MacOS/Unity",
			"/Applications/Unity/Unity.app/Contents/MacOS/Unity",
		}
	case "linux":
		home := os.Getenv("HOME")
		return []string{
			filepath.Join(home, "Unity/Hub/Editor/*/Editor/Unity"),
			filepath.Join(home, "Unity/Editor/Unity"),
			"/opt/unity/Editor/Unity",
			"/usr/bin/unity-editor",
		}
	case "windows":
		return []string{
			`C:\Program Files\Unity\Hub\Editor\*\Editor\Unity.exe`,
			`C:\Program Files\Unity\Editor\Unity.exe`,
		}
	default:
		return nil
	}
}

func readUnityLog(path string) string {
	data, err := os.ReadFile(path)
	if err != nil || len(data) == 0 {
		return ""
	}

	if len(data) > maxUnityLogBytes {
		return string(data[:maxUnityLogBytes]) + "\n[output truncated]"
	}

	return string(data)
}

func joinNonEmpty(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.TrimSpace(part) == "" {
			continue
		}
		nonEmpty = append(nonEmpty, part)
	}
	return strings.Join(nonEmpty, "\n")
}
