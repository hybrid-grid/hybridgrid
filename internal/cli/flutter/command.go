package flutter

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/sourcearchive"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/taskid"
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
		outputDir   string
		distribute  bool
		abis        []string
	)

	cmd := &cobra.Command{
		Use:   outputType,
		Short: fmt.Sprintf("Build Flutter %s artifacts", outputType),
		// A failed build is not a usage mistake; keep the real error visible.
		SilenceUsage: true,
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

			if distribute {
				if outputType != outputAPK {
					return errors.New("--distribute-abis only applies to apk: an app bundle must contain every ABI")
				}
				abiList, err := normalizeABIs(abis)
				if err != nil {
					return err
				}
				return runDistributedBuild(cmd, deps, req, abiList, outputDir)
			}

			resp, err := runBuild(cmd.Context(), deps, req)
			if err != nil {
				return err
			}

			if err := printBuildResult(cmd, resp); err != nil {
				return err
			}
			return saveOutputs(cmd, outputDir, resp)
		},
	}

	cmd.Flags().StringVar(&projectPath, "project", "", "path to Flutter project")
	cmd.Flags().StringVar(&buildMode, "build-mode", defaultBuildMode, "build mode (debug, profile, release)")
	cmd.Flags().StringVar(&flavor, "flavor", "", "build flavor")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "", "write the built APK/AAB files into this directory")
	if outputType == outputAPK {
		cmd.Flags().BoolVar(&distribute, "distribute-abis", false, "build each ABI as its own task, in parallel on different workers (one APK per ABI)")
		cmd.Flags().StringSliceVar(&abis, "abis", nil, "ABIs for --distribute-abis: arm64, arm, x64 (default all three)")
	}

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
	archive, hash, err := sourcearchive.Create(projectPath, shouldExclude)
	if err != nil {
		return nil, err
	}

	mode, err := formatBuildMode(outputType, buildMode)
	if err != nil {
		return nil, err
	}

	req := &pb.BuildRequest{
		TaskId:         taskid.NewTaskID(),
		BuildId:        taskid.BuildSessionID(),
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

// connect resolves the coordinator and opens a build client.
func connect(deps Dependencies) (BuildClient, error) {
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
	return c, nil
}

func runBuild(ctx context.Context, deps Dependencies, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	c, err := connect(deps)
	if err != nil {
		return nil, err
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

// saveOutputs writes the response's APK/AAB files into dir when --output
// was given. Without it the files stay downloadable from the dashboard.
func saveOutputs(cmd *cobra.Command, dir string, resp *pb.BuildResponse) error {
	if dir == "" || resp == nil {
		return nil
	}
	paths, err := extractOutputs(resp.GetArtifacts(), dir)
	if err != nil {
		return err
	}
	for _, p := range paths {
		cmd.Printf("Saved %s\n", p)
	}
	return nil
}

// runDistributedBuild builds one APK per ABI as parallel tasks of one
// build and reports each, the wall time, and the serial time it saved.
func runDistributedBuild(cmd *cobra.Command, deps Dependencies, base *pb.BuildRequest, abis []string, outputDir string) error {
	c, err := connect(deps)
	if err != nil {
		return err
	}
	defer c.Close()

	ctx := cmd.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	cmd.Printf("Building %d ABIs in parallel (build %s): %s\n", len(abis), base.GetBuildId(), strings.Join(abis, ", "))
	start := time.Now()
	results := runDistributed(ctx, c, abiRequests(base, abis))
	wall := time.Since(start)

	var serial time.Duration
	var failed []string
	for _, r := range results {
		serial += r.Took
		switch {
		case r.Err != nil:
			failed = append(failed, r.ABI)
			cmd.Printf("  %-14s failed: %v\n", r.ABI, r.Err)
		case r.Resp.GetStatus() != pb.TaskStatus_STATUS_COMPLETED:
			failed = append(failed, r.ABI)
			cmd.Printf("  %-14s failed on %s (exit %d): %s\n", r.ABI, r.Resp.GetWorkerId(), r.Resp.GetExitCode(), lastLine(r.Resp.GetStderr()))
		default:
			cache := ""
			if r.Resp.GetFromCache() {
				cache = " (cache hit)"
			}
			cmd.Printf("  %-14s %-22s %6.1fs%s\n", r.ABI, r.Resp.GetWorkerId(), r.Took.Seconds(), cache)
			for _, a := range r.Resp.GetArtifactList() {
				cmd.Printf("      %s (%d bytes)\n", a.GetPath(), a.GetSizeBytes())
			}
			if err := saveOutputs(cmd, outputDir, r.Resp); err != nil {
				return err
			}
		}
	}
	cmd.Printf("Done in %.1fs (tasks summed %.1fs)\n", wall.Seconds(), serial.Seconds())
	if len(failed) > 0 {
		return fmt.Errorf("%d of %d ABI builds failed: %s", len(failed), len(abis), strings.Join(failed, ", "))
	}
	return nil
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		return s[i+1:]
	}
	return s
}
