package executor

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/rs/zerolog/log"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// DockerResourceLimits holds configurable resource limits for Docker containers.
type DockerResourceLimits struct {
	MemoryBytes int64 // Container memory limit (default 512MB)
	NanoCPUs    int64 // CPU limit in nano-CPUs (default 1e9 = 1 CPU)
	PidsLimit   int64 // Max number of PIDs (default 100)
}

// DefaultDockerResourceLimits returns the default resource limits.
func DefaultDockerResourceLimits() DockerResourceLimits {
	return DockerResourceLimits{
		MemoryBytes: 512 * 1024 * 1024,
		NanoCPUs:    1_000_000_000,
		PidsLimit:   100,
	}
}

// DockerExecutor executes compilation inside Docker containers.
type DockerExecutor struct {
	client *client.Client
	images map[pb.Architecture]string
	// toolchains overrides the compiler command used inside the image
	// for a target arch (e.g. aarch64-linux-gnu-gcc for a dockcross
	// image whose plain gcc is the host compiler).
	toolchains map[pb.Architecture]string
	// probeResults caches image+cc dumpmachine verdicts: true = matches
	// the target arch, false = known mismatch. Probes run once per pair.
	probeResults sync.Map
	limits       DockerResourceLimits
}

// Default dockcross images for cross-compilation.
var defaultImages = map[pb.Architecture]string{
	pb.Architecture_ARCH_X86_64: "dockcross/linux-x64",
	pb.Architecture_ARCH_ARM64:  "dockcross/linux-arm64",
	pb.Architecture_ARCH_ARMV7:  "dockcross/linux-armv7",
}

// NewDockerExecutor creates a new Docker executor with default resource limits.
func NewDockerExecutor() (*DockerExecutor, error) {
	return NewDockerExecutorWithLimits(DefaultDockerResourceLimits())
}

// NewDockerExecutorWithLimits creates a new Docker executor with custom resource limits.
func NewDockerExecutorWithLimits(limits DockerResourceLimits) (*DockerExecutor, error) {
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
		return nil, fmt.Errorf("docker not available: %w", err)
	}

	return &DockerExecutor{
		client:     cli,
		images:     defaultImages,
		toolchains: make(map[pb.Architecture]string),
		limits:     limits,
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

// SetToolchain overrides the compiler command executed inside the image
// for a target architecture. Cross-toolchain images often keep the host
// gcc as "gcc"; this is how an operator points at the prefixed cross
// compiler (e.g. "aarch64-linux-gnu-gcc").
func (e *DockerExecutor) SetToolchain(arch pb.Architecture, cc string) {
	e.toolchains[arch] = cc
}

// toolchainFor returns the compiler command to run inside the image for
// a target arch: the operator override when set, otherwise the compiler
// the client asked for.
func (e *DockerExecutor) toolchainFor(arch pb.Architecture, requested string) string {
	if cc, ok := e.toolchains[arch]; ok && cc != "" {
		return cc
	}
	return requested
}

// archPlatform maps a target architecture to its Docker platform suffix
// ("" keeps the daemon default, i.e. the host architecture).
func archPlatform(arch pb.Architecture) string {
	switch arch {
	case pb.Architecture_ARCH_X86_64:
		return "linux/amd64"
	case pb.Architecture_ARCH_ARM64:
		return "linux/arm64"
	case pb.Architecture_ARCH_ARMV7:
		return "linux/arm/v7"
	default:
		return ""
	}
}

// archTriplePrefix is the dumpmachine prefix a toolchain must report to
// be trusted for the target architecture.
func archTriplePrefix(arch pb.Architecture) string {
	switch arch {
	case pb.Architecture_ARCH_X86_64:
		return "x86_64"
	case pb.Architecture_ARCH_ARM64:
		return "aarch64"
	case pb.Architecture_ARCH_ARMV7:
		return "arm"
	default:
		return ""
	}
}

// tripleMatches reports whether a compiler's dumpmachine output targets
// the given architecture.
func tripleMatches(dumpmachine string, arch pb.Architecture) bool {
	want := archTriplePrefix(arch)
	if want == "" {
		return true // unspecified target: accept anything
	}
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(dumpmachine)), want)
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

	// Determine source file and write it
	var srcFile string
	if len(req.RawSource) > 0 {
		// Raw source mode: write source with original filename
		srcFile = req.SourceFilename
		if srcFile == "" {
			srcFile = "source.c"
		}
		if err := os.WriteFile(filepath.Join(workDir, srcFile), req.RawSource, 0644); err != nil {
			return nil, fmt.Errorf("failed to write source: %w", err)
		}
		// Write bundled include files
		for path, content := range req.IncludeFiles {
			fullPath := filepath.Join(workDir, path)
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				return nil, fmt.Errorf("failed to create include dir: %w", err)
			}
			if err := os.WriteFile(fullPath, content, 0644); err != nil {
				return nil, fmt.Errorf("failed to write include file %s: %w", path, err)
			}
		}
	} else {
		// Preprocessed source mode
		srcFile = "source.i"
		if err := os.WriteFile(filepath.Join(workDir, srcFile), req.PreprocessedSource, 0644); err != nil {
			return nil, fmt.Errorf("failed to write source: %w", err)
		}
	}

	// Select image based on target architecture
	img := e.selectImage(req.TargetArch)
	if img == "" {
		return nil, fmt.Errorf("no Docker image available for architecture: %v", req.TargetArch)
	}

	// Compiler command inside the image (operator override or client's).
	cc := e.toolchainFor(req.TargetArch, req.Compiler)

	// Ensure image is available (auto-pull if not)
	if err := e.ensureImage(ctx, img, req.TargetArch); err != nil {
		return nil, fmt.Errorf("failed to ensure Docker image: %w", err)
	}

	// Refuse loudly when the image's compiler targets the wrong arch —
	// silent x86_64 objects for arm64 tasks are exactly the bug class
	// raw mode exists to prevent.
	if err := e.probeToolchain(ctx, img, cc, req.TargetArch); err != nil {
		return nil, err
	}

	// Build compilation command
	outFile := "output.o"
	var compileCmd []string
	if len(req.RawSource) > 0 {
		compileCmd = e.buildRawCommand(cc, req.Args, srcFile, outFile, req.IncludePaths)
	} else {
		compileCmd = e.buildCommand(cc, req.Args, srcFile, outFile)
	}

	// Create container config with security settings
	containerConfig := &container.Config{
		Image:      img,
		Cmd:        compileCmd,
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
			Memory:     e.limits.MemoryBytes,
			MemorySwap: e.limits.MemoryBytes, // No swap
			NanoCPUs:   e.limits.NanoCPUs,
			PidsLimit:  func() *int64 { v := e.limits.PidsLimit; return &v }(),
		},
		NetworkMode: "none",
	}

	// Apply Linux-specific security features
	// These are not available on Windows Docker (even with WSL2 backend)
	if runtime.GOOS != "windows" {
		hostConfig.CapDrop = []string{"ALL"}
		hostConfig.SecurityOpt = []string{"no-new-privileges"}
		hostConfig.ReadonlyRootfs = true
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

// buildCommand constructs the argv used for compilation.
func (e *DockerExecutor) buildCommand(compiler string, args []string, srcFile, outFile string) []string {
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

	return cmdParts
}

// buildRawCommand constructs the argv used for raw source compilation in Docker.
func (e *DockerExecutor) buildRawCommand(compiler string, args []string, srcFile, outFile string, includePaths []string) []string {
	var cmdParts []string

	cmdParts = append(cmdParts, compiler)
	cmdParts = append(cmdParts, "-c")

	// Add include paths pointing to /work
	for _, p := range includePaths {
		cmdParts = append(cmdParts, "-I/work/"+p)
	}

	// Add other args, filtering out input/output/include paths
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "-o" {
			i++
			continue
		}
		if arg == "-c" {
			continue
		}
		if arg == "-I" {
			i++ // skip -I <path>
			continue
		}
		if strings.HasPrefix(arg, "-I") {
			continue // skip -Ipath
		}
		if isInputFile(arg) {
			continue
		}
		cmdParts = append(cmdParts, arg)
	}

	cmdParts = append(cmdParts, srcFile, "-o", outFile)
	return cmdParts
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
func (e *DockerExecutor) ensureImage(ctx context.Context, imageName string, arch pb.Architecture) error {
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
		if err := e.pullImage(ctx, imageName, arch); err != nil {
			return err
		}
	}

	return nil
}

// IsWSL2DockerAvailable checks if Docker is running with WSL2 backend on Windows.
// Returns true if Docker is available and using Linux containers via WSL2.
func IsWSL2DockerAvailable() bool {
	if runtime.GOOS != "windows" {
		return false
	}

	// Check if Docker is running with Linux containers (WSL2 backend)
	cmd := exec.Command("docker", "info", "--format", "{{.OSType}}")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	// If OSType is "linux", Docker is using WSL2 backend
	return strings.TrimSpace(string(output)) == "linux"
}

// pullImage pulls a Docker image from the registry, requesting the
// variant matching the target architecture. Without it a multi-arch tag
// resolves to the daemon's host architecture and an arm64 task would
// silently compile with an x86_64 toolchain.
func (e *DockerExecutor) pullImage(ctx context.Context, imageName string, arch pb.Architecture) error {
	// Use a longer timeout for pulling images
	pullCtx, cancel := context.WithTimeout(ctx, 10*time.Minute)
	defer cancel()

	opts := image.PullOptions{}
	if platform := archPlatform(arch); platform != "" {
		opts.Platform = platform
	}
	reader, err := e.client.ImagePull(pullCtx, imageName, opts)
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

// probeToolchain runs "<cc> -dumpmachine" inside the image once per
// (image, cc) pair and refuses the task when the reported triple does
// not match the target architecture. Without this a cross-compile task
// can silently produce objects for the wrong arch (the default
// dockcross images' plain gcc is the host compiler).
func (e *DockerExecutor) probeToolchain(ctx context.Context, img, cc string, arch pb.Architecture) error {
	if archTriplePrefix(arch) == "" {
		return nil // unspecified target: nothing to compare
	}
	key := img + "|" + cc
	if v, ok := e.probeResults.Load(key); ok {
		if v.(bool) {
			return nil
		}
		return e.toolchainMismatchErr(img, cc, arch)
	}

	probeCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	dump, err := e.runInImage(probeCtx, img, []string{"sh", "-c", cc + " -dumpmachine"})
	if err != nil {
		// A missing sh/cc surfaces here; treat like a mismatch with a
		// clear message instead of a cryptic exec failure later.
		return fmt.Errorf("docker toolchain probe failed for image %q (cc %q): %w — pin a toolchain that targets %s, e.g. --docker-image-arm64 with a single-arch image, or set the matching -cc override",
			img, cc, err, arch.String())
	}
	ok := tripleMatches(dump, arch)
	e.probeResults.Store(key, ok)
	if !ok {
		return e.toolchainMismatchErr(img, cc, arch)
	}
	return nil
}

func (e *DockerExecutor) toolchainMismatchErr(img, cc string, arch pb.Architecture) error {
	want := archTriplePrefix(arch)
	return fmt.Errorf(
		"docker image %q: compiler %q does not target %s (expected %s-* triple) — refusing to compile silently wrong output; point --docker-image-%s-cc at the cross compiler or pin a single-arch image",
		img, cc, arch.String(), want, strings.ToLower(strings.TrimPrefix(arch.String(), "ARCH_")))
}

// runInImage runs argv in a throwaway container over the image and
// returns trimmed stdout.
func (e *DockerExecutor) runInImage(ctx context.Context, img string, argv []string) (string, error) {
	cfg := &container.Config{Image: img, Cmd: argv}
	hostCfg := &container.HostConfig{NetworkMode: "none"}
	resp, err := e.client.ContainerCreate(ctx, cfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", err
	}
	id := resp.ID
	defer func() {
		rmCtx, rmCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer rmCancel()
		_ = e.client.ContainerRemove(rmCtx, id, container.RemoveOptions{Force: true})
	}()
	if err := e.client.ContainerStart(ctx, id, container.StartOptions{}); err != nil {
		return "", err
	}
	statusCh, errCh := e.client.ContainerWait(ctx, id, container.WaitConditionNotRunning)
	var outBuf bytes.Buffer
	logsCh, logsErr := e.client.ContainerLogs(ctx, id, container.LogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if logsErr != nil {
		return "", fmt.Errorf("container logs error: %w", logsErr)
	}
	copyDone := make(chan struct{})
	go func() {
		// Non-TTY log streams are multiplexed with 8-byte frame headers;
		// stdcopy de-muxes stdout/stderr into the plain bytes we want.
		_, _ = stdcopy.StdCopy(&outBuf, &outBuf, logsCh)
		_ = logsCh.Close()
		close(copyDone)
	}()
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case err := <-errCh:
			return "", fmt.Errorf("container wait error: %w", err)
		case st := <-statusCh:
			<-copyDone // logs drain just after exit; buffer is final here
			if st.StatusCode != 0 {
				return "", fmt.Errorf("probe exited %d: %s", st.StatusCode, strings.TrimSpace(outBuf.String()))
			}
			return strings.TrimSpace(outBuf.String()), nil
		}
	}
}
