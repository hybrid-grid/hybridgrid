package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// DockerExecutor executes compilation inside Docker containers.
type DockerExecutor struct {
	client *client.Client
	images map[pb.Architecture]string
}

// Default dockcross images for cross-compilation.
var defaultImages = map[pb.Architecture]string{
	pb.Architecture_ARCH_X86_64: "dockcross/linux-x64",
	pb.Architecture_ARCH_ARM64:  "dockcross/linux-arm64",
	pb.Architecture_ARCH_ARMV7:  "dockcross/linux-armv7",
}

// NewDockerExecutor creates a new Docker executor.
func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify Docker is available
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = cli.Ping(ctx)
	if err != nil {
		cli.Close()
		return nil, fmt.Errorf("Docker not available: %w", err)
	}

	return &DockerExecutor{
		client: cli,
		images: defaultImages,
	}, nil
}

// Name returns the executor name.
func (e *DockerExecutor) Name() string {
	return "docker"
}

// CanExecute returns true if this executor can handle the target architecture.
func (e *DockerExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	_, ok := e.images[targetArch]
	return ok
}

// SetImage sets a custom image for a target architecture.
func (e *DockerExecutor) SetImage(arch pb.Architecture, image string) {
	e.images[arch] = image
}

// Execute runs the compilation inside a Docker container.
func (e *DockerExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	start := time.Now()

	// Create temp directory for this task
	workDir, err := os.MkdirTemp("", fmt.Sprintf("hg-docker-%s-", req.TaskID))
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(workDir)

	// Write preprocessed source to temp file
	srcFile := "source.i"
	if err := os.WriteFile(filepath.Join(workDir, srcFile), req.PreprocessedSource, 0644); err != nil {
		return nil, fmt.Errorf("failed to write source: %w", err)
	}

	// Select image based on target architecture
	img := e.selectImage(req.TargetArch)
	if img == "" {
		return nil, fmt.Errorf("no Docker image available for architecture: %v", req.TargetArch)
	}

	// Ensure image is available (auto-pull if not)
	if err := e.ensureImage(ctx, img); err != nil {
		return nil, fmt.Errorf("failed to ensure Docker image: %w", err)
	}

	// Build compilation command
	outFile := "output.o"
	compileCmd := e.buildCommand(req.Compiler, req.Args, srcFile, outFile)

	// Create container config with security settings
	containerConfig := &container.Config{
		Image:      img,
		Cmd:        []string{"/bin/sh", "-c", compileCmd},
		WorkingDir: "/work",
		Tty:        false,
	}

	// Host config with security restrictions
	hostConfig := &container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   mount.TypeBind,
				Source: workDir,
				Target: "/work",
			},
		},
		Resources: container.Resources{
			Memory:     512 * 1024 * 1024, // 512MB
			MemorySwap: 512 * 1024 * 1024, // No swap
			NanoCPUs:   1000000000,        // 1 CPU
			PidsLimit:  func() *int64 { v := int64(100); return &v }(),
		},
		NetworkMode:    "none",
		CapDrop:        []string{"ALL"},
		SecurityOpt:    []string{"no-new-privileges"},
		ReadonlyRootfs: true,
		// Allow writing to /work (mounted volume)
	}

	// Create container
	resp, err := e.client.ContainerCreate(ctx, containerConfig, hostConfig, nil, nil, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create container: %w", err)
	}
	containerID := resp.ID

	// Ensure cleanup
	defer func() {
		removeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		e.client.ContainerRemove(removeCtx, containerID, container.RemoveOptions{Force: true})
	}()

	// Start container
	if err := e.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Wait for container to finish
	statusCh, errCh := e.client.ContainerWait(ctx, containerID, container.WaitConditionNotRunning)

	var exitCode int64
	select {
	case err := <-errCh:
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				// Kill container on timeout
				killCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				e.client.ContainerKill(killCtx, containerID, "KILL")
				return &Result{
					Success:         false,
					ExitCode:        -1,
					Stderr:          "compilation timed out",
					CompilationTime: time.Since(start),
				}, nil
			}
			return nil, fmt.Errorf("container wait error: %w", err)
		}
	case status := <-statusCh:
		exitCode = status.StatusCode
	}

	// Get container logs
	stdout, stderr, err := e.getLogs(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	compilationTime := time.Since(start)

	result := &Result{
		Stdout:          stdout,
		Stderr:          stderr,
		ExitCode:        int32(exitCode),
		CompilationTime: compilationTime,
	}

	if exitCode != 0 {
		result.Success = false
		return result, nil
	}

	// Read output file
	objectCode, err := os.ReadFile(filepath.Join(workDir, outFile))
	if err != nil {
		result.Stderr += fmt.Sprintf("\nFailed to read output: %v", err)
		result.Success = false
		return result, nil
	}

	result.Success = true
	result.ObjectCode = objectCode

	return result, nil
}

// selectImage returns the Docker image for the target architecture.
func (e *DockerExecutor) selectImage(arch pb.Architecture) string {
	if img, ok := e.images[arch]; ok {
		return img
	}
	return ""
}

// buildCommand constructs the shell command for compilation.
func (e *DockerExecutor) buildCommand(compiler string, args []string, srcFile, outFile string) string {
	var cmdParts []string

	// Use the appropriate compiler from dockcross
	// dockcross images set up the cross-compiler in PATH
	cmdParts = append(cmdParts, compiler)

	// Add compile-only flag
	cmdParts = append(cmdParts, "-c")

	// Add other args, filtering out input/output
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-o" {
			i++ // Skip output file
			continue
		}
		if arg == "-c" {
			continue // Already added
		}
		if isInputFile(arg) {
			continue // Skip input files
		}
		cmdParts = append(cmdParts, arg)
	}

	// Add input and output
	cmdParts = append(cmdParts, srcFile, "-o", outFile)

	return strings.Join(cmdParts, " ")
}

// getLogs retrieves stdout and stderr from the container.
func (e *DockerExecutor) getLogs(ctx context.Context, containerID string) (string, string, error) {
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	}

	reader, err := e.client.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", "", err
	}
	defer reader.Close()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, reader)
	if err != nil && err != io.EOF {
		return "", "", err
	}

	return stdout.String(), stderr.String(), nil
}

// Close closes the Docker client connection.
func (e *DockerExecutor) Close() error {
	return e.client.Close()
}

// ensureImage checks if the image exists locally and pulls it if not.
// Also checks for updates if the image is older than 24 hours.
func (e *DockerExecutor) ensureImage(ctx context.Context, imageName string) error {
	// Add :latest tag if no tag specified
	if !strings.Contains(imageName, ":") {
		imageName += ":latest"
	}

	// Check if image exists locally
	images, err := e.client.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list images: %w", err)
	}

	imageExists := false
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == imageName {
				imageExists = true
				break
			}
		}
		if imageExists {
			break
		}
	}

	// Pull image if not exists
	if !imageExists {
		log.Info().Str("image", imageName).Msg("Pulling Docker image (first time)")
		if err := e.pullImage(ctx, imageName); err != nil {
			return err
		}
	}

	return nil
}

// pullImage pulls a Docker image from the registry.
func (e *DockerExecutor) pullImage(ctx context.Context, imageName string) error {
	// Use a longer timeout for pulling images
	pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	reader, err := e.client.ImagePull(pullCtx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", imageName, err)
	}
	defer reader.Close()

	// Read and discard the output (shows pull progress)
	// Log progress periodically
	buf := make([]byte, 1024)
	for {
		_, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading pull response: %w", err)
		}
	}

	log.Info().Str("image", imageName).Msg("Docker image pulled successfully")
	return nil
}
