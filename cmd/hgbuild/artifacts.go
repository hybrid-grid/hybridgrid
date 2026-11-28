package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
)

// artifactUploadTimeout bounds one upload; large binaries on a slow link
// need far longer than a compile RPC.
const artifactUploadTimeout = 30 * time.Minute

var (
	// artifactPaths collects --artifact values from the build wrappers.
	artifactPaths []string
	// artifactFlagErr reports a malformed --artifact (missing value).
	artifactFlagErr error
)

func newArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "artifact",
		Short: "Manage build artifacts stored on the coordinator",
	}
	var buildID string
	upload := &cobra.Command{
		Use:   "upload --build BUILD_ID FILE...",
		Short: "Upload build outputs so they can be downloaded from the dashboard",
		Long: `Upload files to the coordinator under a build, for download from the
dashboard's build page. "hgbuild make --artifact FILE" does this
automatically after a successful build; use this command for build
systems hgbuild does not wrap.

Examples:
  hgbuild artifact upload --build "$HG_BUILD_ID" build/app
  hgbuild artifact upload --build release-42 dist/app.tar.gz dist/app.sha256`,
		Args:         cobra.MinimumNArgs(1),
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if buildID == "" {
				buildID = os.Getenv("HG_BUILD_ID")
			}
			if buildID == "" {
				return errors.New("--build is required (or export HG_BUILD_ID)")
			}
			return uploadArtifacts(buildID, args)
		},
	}
	upload.Flags().StringVar(&buildID, "build", "", "build id to attach the files to (default: $HG_BUILD_ID)")
	cmd.AddCommand(upload)
	return cmd
}

// uploadArtifacts sends each file to the coordinator under buildID and
// reports what was stored. Any failure fails the command: the user asked
// for these outputs to be kept, so silently losing one would be worse.
func uploadArtifacts(buildID string, paths []string) error {
	if buildID == "" {
		return errors.New("cannot upload artifacts: no build id")
	}
	address := getCoordinatorAddress()
	if address == "" {
		return errors.New("cannot upload artifacts: no coordinator (set --coordinator or HG_COORDINATOR)")
	}
	c, err := client.New(newClientConfig(address, artifactUploadTimeout))
	if err != nil {
		return fmt.Errorf("failed to connect to coordinator: %w", err)
	}
	defer c.Close()

	for _, path := range paths {
		if err := uploadOne(c, buildID, path); err != nil {
			return err
		}
	}
	return nil
}

func uploadOne(c *client.Client, buildID, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open artifact: %w", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat artifact: %w", err)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("artifact %s is not a regular file", path)
	}

	ctx, cancel := context.WithTimeout(context.Background(), artifactUploadTimeout)
	defer cancel()
	resp, err := c.UploadArtifact(ctx, buildID, filepath.Base(path), info.Size(), f)
	if err != nil {
		return fmt.Errorf("failed to upload artifact %s: %w", path, err)
	}
	fmt.Fprintf(os.Stderr, "[artifact] %s (%s) → %s  sha256:%.12s\n", resp.GetName(), humanBytes(resp.GetSizeBytes()), buildID, resp.GetSha256())
	return nil
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
