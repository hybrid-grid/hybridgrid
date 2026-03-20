package flutter

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

const (
	outputAPK       = "apk"
	outputAppBundle = "appbundle"

	defaultBuildMode = "release"
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
		Use:   "flutter",
		Short: "Build Flutter projects with Hybrid-Grid",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Build Flutter application artifacts",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	buildCmd.AddCommand(
		newBuildTargetCmd(outputAPK, deps),
		newBuildTargetCmd(outputAppBundle, deps),
	)

	cmd.AddCommand(buildCmd)
	return cmd
}

func newBuildTargetCmd(outputType string, deps Dependencies) *cobra.Command {
	var (
		projectPath string
		buildMode   string
		flavor      string
	)

	cmd := &cobra.Command{
		Use:   outputType,
		Short: fmt.Sprintf("Build Flutter %s artifacts", outputType),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectPath = strings.TrimSpace(projectPath)
			if projectPath == "" {
				return fmt.Errorf("project required")
			}

			info, err := os.Stat(projectPath)
			if err != nil {
				return fmt.Errorf("project not found: %w", err)
			}
			if !info.IsDir() {
				return fmt.Errorf("project must be a directory")
			}

			mode, err := normalizeBuildMode(buildMode)
			if err != nil {
				return err
			}

			req, err := buildRequest(projectPath, outputType, mode, strings.TrimSpace(flavor), deps.BuildTimeout)
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

	cmd.Flags().StringVar(&projectPath, "project", "", "path to Flutter project")
	cmd.Flags().StringVar(&buildMode, "build-mode", defaultBuildMode, "build mode (debug, profile, release)")
	cmd.Flags().StringVar(&flavor, "flavor", "", "build flavor")

	return cmd
}

func normalizeBuildMode(value string) (string, error) {
	mode := strings.TrimSpace(value)
	if mode == "" {
		return defaultBuildMode, nil
	}

	mode = strings.ToLower(mode)
	switch mode {
	case "debug", "profile", "release":
		return mode, nil
	default:
		return "", fmt.Errorf("invalid build-mode %q (supported: debug, profile, release)", value)
	}
}

func buildRequest(projectPath, outputType, buildMode, flavor string, buildTimeout time.Duration) (*pb.BuildRequest, error) {
	archive, hash, err := createSourceArchive(projectPath)
	if err != nil {
		return nil, err
	}

	mode, err := formatBuildMode(outputType, buildMode)
	if err != nil {
		return nil, err
	}

	req := &pb.BuildRequest{
		TaskId:         generateTaskID(),
		SourceHash:     hash,
		BuildType:      pb.BuildType_BUILD_TYPE_FLUTTER,
		TargetPlatform: pb.TargetPlatform_PLATFORM_ANDROID,
		SourceArchive:  archive,
		Config: &pb.BuildRequest_FlutterConfig{
			FlutterConfig: &pb.FlutterConfig{
				BuildMode: mode,
				Flavor:    flavor,
			},
		},
	}

	if buildTimeout > 0 {
		req.TimeoutSeconds = int32(buildTimeout.Seconds())
	}

	return req, nil
}

func formatBuildMode(outputType, buildMode string) (string, error) {
	switch outputType {
	case outputAPK, outputAppBundle:
	default:
		return "", fmt.Errorf("unsupported output type: %s", outputType)
	}

	mode := strings.TrimSpace(buildMode)
	if mode == "" {
		mode = defaultBuildMode
	}

	return fmt.Sprintf("%s-%s", outputType, mode), nil
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
