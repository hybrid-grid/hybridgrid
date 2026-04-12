package unity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

type BuildClient interface {
	Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error)
	Close() error
}

type ClientFactory func(address string, timeout time.Duration) (BuildClient, error)

type Dependencies struct {
	CoordinatorAddr func() string
	NewClient       ClientFactory
	RequestTimeout  time.Duration
	BuildTimeout    time.Duration
}

func NewCommand(deps Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unity",
		Short: "Build Unity projects with Hybrid-Grid",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	buildCmd := &cobra.Command{
		Use:   "build <platform>",
		Short: "Build Unity player artifacts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath, _ := cmd.Flags().GetString("project")
			buildMethod, _ := cmd.Flags().GetString("build-method")
			unityVersion, _ := cmd.Flags().GetString("unity-version")
			scriptingBackend, _ := cmd.Flags().GetString("scripting-backend")

			projectPath = strings.TrimSpace(projectPath)
			if projectPath == "" {
				return fmt.Errorf("project required")
			}

			buildMethod = strings.TrimSpace(buildMethod)
			if buildMethod == "" {
				return fmt.Errorf("build-method required")
			}

			info, err := os.Stat(projectPath)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("project must be a directory")
			}

			platform, err := parseTargetPlatform(args[0])
			if err != nil {
				return err
			}

			backend, err := normalizeScriptingBackend(scriptingBackend)
			if err != nil {
				return err
			}

			req, err := buildRequest(projectPath, platform, strings.TrimSpace(unityVersion), buildMethod, backend, deps.BuildTimeout)
			if err != nil {
				return err
			}

			resp, err := runBuild(cmd.Context(), deps, req)
			if err != nil {
				return err
			}

			return printBuildResult(cmd, resp)
		},
	}

	buildCmd.Flags().String("project", "", "path to Unity project")
	buildCmd.Flags().String("build-method", "", "Unity build method (for example: BuildScript.Build)")
	buildCmd.Flags().String("unity-version", "", "Unity editor version")
	buildCmd.Flags().String("scripting-backend", "", "scripting backend (mono, il2cpp)")

	_ = buildCmd.MarkFlagRequired("project")
	_ = buildCmd.MarkFlagRequired("build-method")

	cmd.AddCommand(buildCmd)
	return cmd
}

func parseTargetPlatform(value string) (pb.TargetPlatform, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "android":
		return pb.TargetPlatform_PLATFORM_ANDROID, nil
	case "ios":
		return pb.TargetPlatform_PLATFORM_IOS, nil
	case "windows":
		return pb.TargetPlatform_PLATFORM_WINDOWS, nil
	case "linux":
		return pb.TargetPlatform_PLATFORM_LINUX, nil
	case "macos":
		return pb.TargetPlatform_PLATFORM_MACOS, nil
	case "webgl":
		return pb.TargetPlatform_PLATFORM_WEBGL, nil
	default:
		return pb.TargetPlatform_PLATFORM_UNSPECIFIED, fmt.Errorf("invalid platform %q (supported: android, ios, windows, linux, macos, webgl)", value)
	}
}

func normalizeScriptingBackend(value string) (string, error) {
	backend := strings.TrimSpace(value)
	if backend == "" {
		return "", nil
	}

	backend = strings.ToLower(backend)
	switch backend {
	case "mono", "il2cpp":
		return backend, nil
	default:
		return "", fmt.Errorf("invalid scripting-backend %q (supported: mono, il2cpp)", value)
	}
}

func buildRequest(projectPath string, platform pb.TargetPlatform, unityVersion, buildMethod, scriptingBackend string, buildTimeout time.Duration) (*pb.BuildRequest, error) {
	archive, hash, err := createSourceArchive(projectPath)
	if err != nil {
		return nil, err
	}

	req := &pb.BuildRequest{
		TaskId:         generateTaskID(),
		SourceHash:     hash,
		BuildType:      pb.BuildType_BUILD_TYPE_UNITY,
		TargetPlatform: platform,
		SourceArchive:  archive,
		Config: &pb.BuildRequest_UnityConfig{
			UnityConfig: &pb.UnityConfig{
				UnityVersion:     unityVersion,
				BuildMethod:      buildMethod,
				ScriptingBackend: scriptingBackend,
				ExtraArgs:        map[string]string{},
			},
		},
	}

	if buildTimeout > 0 {
		req.TimeoutSeconds = int32(buildTimeout.Seconds())
	}

	return req, nil
}

func runBuild(ctx context.Context, deps Dependencies, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	if deps.CoordinatorAddr == nil {
		return nil, fmt.Errorf("coordinator resolver not configured")
	}
	if deps.NewClient == nil {
		return nil, fmt.Errorf("client factory not configured")
	}

	addr := strings.TrimSpace(deps.CoordinatorAddr())
	if addr == "" {
		return nil, fmt.Errorf("coordinator unavailable")
	}

	c, err := deps.NewClient(addr, deps.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	if ctx == nil {
		ctx = context.Background()
	}

	resp, err := c.Build(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	return resp, nil
}

func printBuildResult(cmd *cobra.Command, resp *pb.BuildResponse) error {
	if resp == nil {
		return fmt.Errorf("empty build response")
	}

	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		message := fmt.Sprintf("build failed with status %s", resp.Status.String())
		if resp.ExitCode != 0 {
			message = fmt.Sprintf("%s (exit %d)", message, resp.ExitCode)
		}
		if strings.TrimSpace(resp.Stderr) != "" {
			message = fmt.Sprintf("%s: %s", message, strings.TrimSpace(resp.Stderr))
		}
		return errors.New(message)
	}

	if resp.FromCache {
		cmd.Println("Build completed (cache hit)")
	} else {
		cmd.Println("Build completed")
	}

	for _, artifact := range resp.ArtifactList {
		cmd.Printf("%s (%d bytes)\n", artifact.Path, artifact.SizeBytes)
	}

	return nil
}

func generateTaskID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("task-%s-%d", hex.EncodeToString(b), time.Now().UnixNano()%10000)
}
