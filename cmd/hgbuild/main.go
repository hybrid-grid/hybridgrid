package main

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/build"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/flutter"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/output"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/sourcearchive"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/taskid"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/unity"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/graph"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

var (
	version           = "v0.0.0-dev"
	cfgFile           string
	coordinator       string
	insecure          bool
	noFallback        bool
	timeout           time.Duration
	verbose           bool
	tlsCert           string
	tlsKey            string
	tlsCA             string
	tracingEnable     bool
	tracingEndpoint   string
	tracingSampleRate float64
	tracerShutdown    func() error
)

const (
	wrapperCCEnv  = "HG_WRAP_CC_MODE"
	wrapperCXXEnv = "HG_WRAP_CXX_MODE"
	noFallbackEnv = "HG_NO_FALLBACK"
)

func main() {
	injectWrappedCompilerMode()

	cfg := config.DefaultConfig()

	// Setup logger
	logger, logCloser, err := logging.SetupLogger(cfg.Log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup logger")
	}
	log.Logger = logger
	defer logCloser.Close()

	rootCmd := &cobra.Command{
		Use:   "hgbuild",
		Short: "Hybrid-Grid Build - Distributed multi-platform build system",
		Long: `hgbuild is a CLI client for the Hybrid-Grid Build system.
It intercepts compiler commands and distributes them to remote workers.

Quick start:
  hgbuild make -j8            Wrap make with distributed compilation
  hgbuild cc -c main.c        Compile C file (drop-in gcc replacement)
  hgbuild c++ -c main.cpp     Compile C++ file (drop-in g++ replacement)
  hgbuild status              Check coordinator status
  hgbuild workers             List available workers

Environment:
  HG_COORDINATOR    Coordinator address (default: auto-discover via mDNS)
  HG_CC             C compiler to use (default: gcc)
  HG_CXX            C++ compiler to use (default: g++)`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			clientCfg := config.ClientConfig{
				CoordinatorAddr: coordinator,
				Timeout:         timeout,
			}
			if err := clientCfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			tlsCfg := config.TLSConfig{
				Enabled:            tlsCert != "" || tlsKey != "" || tlsCA != "",
				CertFile:           tlsCert,
				KeyFile:            tlsKey,
				ClientCA:           tlsCA,
				InsecureSkipVerify: insecure,
			}
			if err := tlsCfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			tracingCfg := config.TracingConfig{
				Enable:     tracingEnable,
				Endpoint:   tracingEndpoint,
				SampleRate: tracingSampleRate,
			}
			if err := tracingCfg.Validate(); err != nil {
				return fmt.Errorf("invalid configuration: %w", err)
			}

			if tracingEnable {
				ctx := cmd.Context()
				if ctx == nil {
					ctx = context.Background()
				}

				tracingCfg := config.TracingToLibConfig(config.TracingConfig{
					Enable:      tracingEnable,
					Endpoint:    tracingEndpoint,
					ServiceName: "hgbuild",
					SampleRate:  tracingSampleRate,
					Insecure:    true,
					Timeout:     10 * time.Second,
					BatchSize:   512,
				})

				tp, err := tracing.Init(ctx, tracingCfg)
				if err != nil {
					return fmt.Errorf("failed to initialize tracing: %w", err)
				}
				if tp != nil {
					tracerShutdown = func() error {
						shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						return tp.Shutdown(shutdownCtx)
					}
					log.Info().
						Str("endpoint", tracingEndpoint).
						Float64("sample_rate", tracingSampleRate).
						Msg("Tracing enabled")
				}
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if tracerShutdown != nil {
				if err := tracerShutdown(); err != nil {
					log.Warn().Err(err).Msg("Failed to shutdown tracer provider")
				}
				tracerShutdown = nil
			}
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&coordinator, "coordinator", "C", "", "coordinator address (auto-discover if empty)")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "use insecure connection")
	rootCmd.PersistentFlags().BoolVar(&noFallback, "no-fallback", false, "disable local fallback when coordinator is unavailable")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 2*time.Minute, "connection timeout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().StringVar(&tlsCert, "tls-cert", "", "Path to TLS certificate file (PEM format)")
	rootCmd.PersistentFlags().StringVar(&tlsKey, "tls-key", "", "Path to TLS private key file (PEM format)")
	rootCmd.PersistentFlags().StringVar(&tlsCA, "tls-ca", "", "Path to CA certificate for server verification")
	rootCmd.PersistentFlags().BoolVar(&tracingEnable, "tracing-enable", false, "Enable OpenTelemetry tracing")
	rootCmd.PersistentFlags().StringVar(&tracingEndpoint, "tracing-endpoint", "localhost:4317", "OTLP gRPC endpoint")
	rootCmd.PersistentFlags().Float64Var(&tracingSampleRate, "tracing-sample-rate", 0.01, "Tracing sample rate (0.0-1.0)")

	// Commands
	rootCmd.AddCommand(
		newVersionCmd(),
		newStatusCmd(),
		newWorkersCmd(),
		newBuildCmd(),
		newFlutterCmd(),
		newUnityCmd(),
		newConfigCmd(),
		newCacheCmd(),
		newGraphCmd(),
		// Compiler wrappers
		newCCCmd(),
		newCXXCmd(),
		// Build wrappers
		newMakeCmd(),
		newNinjaCmd(),
		newWrapCmd(),
	)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hgbuild %s\n", version)
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show coordinator and worker status",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.AutoDetectColors()

			c, err := client.New(newClientConfig(coordinator, timeout))
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resp, err := c.HealthCheck(ctx)
			if err != nil {
				return fmt.Errorf("health check failed: %w", err)
			}

			// Use the new colored output
			output.PrintStatus(output.StatusInfo{
				Address:     coordinator,
				Healthy:     resp.Healthy,
				ActiveTasks: int(resp.ActiveTasks),
				QueuedTasks: int(resp.QueuedTasks),
			})

			return nil
		},
	}
}

func newWorkersCmd() *cobra.Command {
	var verbose bool

	cmd := &cobra.Command{
		Use:   "workers",
		Short: "List available workers",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.AutoDetectColors()

			c, err := client.New(newClientConfig(coordinator, timeout))
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()

			resp, err := c.GetWorkerStatus(ctx)
			if err != nil {
				return fmt.Errorf("failed to get workers: %w", err)
			}

			// Convert to output.WorkerInfo
			workers := make([]output.WorkerInfo, len(resp.Workers))
			for i, w := range resp.Workers {
				workers[i] = output.WorkerInfo{
					ID:           w.WorkerId,
					Arch:         w.NativeArch.String(),
					Cores:        int(w.CpuCores),
					MemoryGB:     float64(w.MemoryBytes) / (1024 * 1024 * 1024),
					ActiveTasks:  int(w.ActiveTasks),
					CircuitState: w.CircuitState,
				}
			}

			// Use colored table output
			if verbose {
				output.PrintWorkersTable(workers, int(resp.TotalWorkers), int(resp.HealthyWorkers))
			} else {
				output.PrintWorkersTableCompact(workers, int(resp.TotalWorkers), int(resp.HealthyWorkers))
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed info")
	return cmd
}

func newBuildCmd() *cobra.Command {
	var (
		buildType  string
		output     string
		compiler   string
		compArgs   []string
		verbose    bool
		targetArch string
	)

	cmd := &cobra.Command{
		Use:   "build [files...]",
		Short: "Submit a build job to the coordinator",
		Long: `Submit source files for distributed compilation.

Examples:
  hgbuild build main.c                    Compile single file
  hgbuild build main.c -o main.o          Compile with output name
  hgbuild build -c gcc main.c -- -O2      Compile with compiler args
  hgbuild build *.c -o myapp              Compile multiple files`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Unknown build types are plugin-builder ids: route the
			// project through the custom-builder path instead of the
			// per-TU compile loop.
			if !isCoreBuildType(buildType) {
				return runCustomBuild(buildType, args)
			}
			c, err := client.New(newClientConfig(coordinator, 5*time.Minute))
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Client cache: same store and key space the Service path uses.
			// A cache failure only disables reuse; the build still runs.
			buildCfg := build.DefaultConfig()
			cacheStore, err := cache.NewStore(buildCfg.CacheDir, buildCfg.CacheMaxSize, buildCfg.CacheTTLHours)
			if err != nil {
				log.Warn().Err(err).Msg("Client cache disabled")
			}

			// One build session per invocation: all files submitted below
			// share it, and the wrapping make's HG_BUILD_ID groups the
			// per-TU invocations of one logical build (see taskid).
			buildID := taskid.BuildSessionID()
			if verbose {
				fmt.Printf("Build session: %s\n", buildID)
			}

			// Process each file
			successCount := 0
			failCount := 0
			var cacheHits int32

			for _, file := range args {
				// Read source file
				source, err := os.ReadFile(file)
				if err != nil {
					fmt.Printf("Error reading %s: %v\n", file, err)
					failCount++
					continue
				}

				// Detect compiler if not specified
				comp := compiler
				if comp == "" {
					comp = detectCompiler(file)
				}

				// Parse target architecture
				arch := parseArch(targetArch)

				// Cache key covers raw source, compiler, flags and target
				// arch — the same bytes the Service path would hash.
				cacheKey := cacheKeyFor(comp, compArgs, arch, source)

				// Cross-arch: local preprocessing would bake this machine's
				// headers into the .i — silently wrong on the other arch
				// (jmp_buf/bits macros resolve for the wrong target). Send raw
				// source plus project includes; the worker's own toolchain
				// preprocesses with arch-correct headers.
				rawCrossArch := arch != pb.Architecture_ARCH_UNSPECIFIED && arch != clientArch()

				var includeFiles map[string][]byte
				if rawCrossArch {
					includeFiles = collectCrossArchIncludes(compArgs)
				} else if prep := preprocessLocal(file, comp, compArgs); prep != nil {
					// Same-arch: preprocess locally so workers receive a
					// self-contained translation unit (mirrors the Service
					// path in internal/cli/build). Falls back to raw source.
					source = prep
				}

				// Generate task ID
				taskID := taskid.NewTaskID()

				// Determine output file
				outFile := output
				if outFile == "" {
					outFile = strings.TrimSuffix(file, filepath.Ext(file)) + ".o"
				}

				if verbose {
					fmt.Printf("Compiling %s → %s (compiler: %s)\n", file, outFile, comp)
				} else {
					fmt.Printf("Compiling %s...", file)
				}

				// Cache hit: the object bytes are already local, so skip the
				// network entirely. Report it so the dashboard counts it.
				if cacheKey != "" && cacheStore != nil {
					if obj, ok := cacheStore.GetBytes(cacheKey); ok {
						if err := os.WriteFile(outFile, obj, 0644); err != nil { // #nosec G703 -- outFile is the compiler's own -o argument, exactly like gcc writing it
							fmt.Printf(" FAILED (write error: %v)\n", err)
							failCount++
							continue
						}
						fmt.Printf(" OK (cached)\n")
						cacheHits++
						successCount++
						continue
					}
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

				// Create compile request
				req := &pb.CompileRequest{
					TaskId:         taskID,
					BuildId:        buildID,
					Compiler:       comp,
					CompilerArgs:   compArgs,
					TargetArch:     arch,
					TimeoutSeconds: 300,
				}
				if rawCrossArch {
					req.RawSource = source
					req.SourceFilename = filepath.Base(file)
					req.IncludeFiles = includeFiles
					req.ClientOs = runtime.GOOS
					req.ClientArch = clientArch()
				} else {
					req.PreprocessedSource = source
				}

				// Send to coordinator
				start := time.Now()
				resp, err := c.Compile(ctx, req)
				cancel()

				if err != nil {
					fmt.Printf(" FAILED (%v)\n", err)
					failCount++
					continue
				}

				if resp.Status == pb.TaskStatus_STATUS_COMPLETED {
					// Write output file
					if err := os.WriteFile(outFile, resp.ObjectFile, 0644); err != nil {
						fmt.Printf(" FAILED (write error: %v)\n", err)
						failCount++
						continue
					}

					// Store only successful objects; a cache write failure
					// must never fail the build.
					if cacheKey != "" && cacheStore != nil && len(resp.ObjectFile) > 0 {
						if err := cacheStore.PutBytes(cacheKey, resp.ObjectFile); err != nil && verbose {
							fmt.Printf("  [cache] store failed: %v\n", err)
						}
					}

					duration := time.Since(start)
					if verbose {
						fmt.Printf(" OK (%.2fs, %d bytes, queue: %dms, compile: %dms)\n",
							duration.Seconds(), len(resp.ObjectFile),
							resp.QueueTimeMs, resp.CompilationTimeMs)
					} else {
						fmt.Printf(" OK (%.2fs)\n", duration.Seconds())
					}
					successCount++
				} else {
					fmt.Printf(" FAILED (exit %d)\n", resp.ExitCode)
					if resp.Stderr != "" {
						fmt.Printf("  stderr: %s\n", resp.Stderr)
					}
					failCount++
				}
			}

			// One report per invocation: a per-file RPC costs a full RTT each
			// and would defeat the point of a local cache hit.
			if cacheHits > 0 {
				_ = c.ReportCacheHit(context.Background(), cacheHits)
			}

			// Summary
			fmt.Printf("\nResults: %d succeeded, %d failed\n", successCount, failCount)

			if failCount > 0 {
				return fmt.Errorf("%d files failed to compile", failCount)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&buildType, "type", "t", "cpp", "build type (cpp, flutter, unity, rust, go, or any plugin builder id)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (for single file builds)")
	cmd.Flags().StringVar(&compiler, "compiler", "", "compiler to use (auto-detect if empty)")
	cmd.Flags().StringSliceVar(&compArgs, "args", nil, "compiler arguments")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().StringVar(&targetArch, "arch", "", "target architecture (x86_64, arm64)")

	return cmd
}

func newFlutterCmd() *cobra.Command {
	deps := flutter.Dependencies{
		CoordinatorAddr: getCoordinatorAddress,
		NewClient: func(address string, requestTimeout time.Duration) (flutter.BuildClient, error) {
			cfg := newClientConfig(address, requestTimeout)
			return client.New(cfg)
		},
		RequestTimeout: 25 * time.Minute,
		BuildTimeout:   20 * time.Minute,
	}

	return flutter.NewCommand(deps)
}

func newUnityCmd() *cobra.Command {
	deps := unity.Dependencies{
		CoordinatorAddr: getCoordinatorAddress,
		NewClient: func(address string, requestTimeout time.Duration) (unity.BuildClient, error) {
			cfg := newClientConfig(address, requestTimeout)
			return client.New(cfg)
		},
		RequestTimeout: 25 * time.Minute,
		BuildTimeout:   30 * time.Minute,
	}

	return unity.NewCommand(deps)
}

// isCoreBuildType reports whether buildType names a built-in type the
// build command handles itself (cpp compiles per-TU; flutter and unity
// have dedicated subcommands). Everything else is a plugin-builder id.
func isCoreBuildType(buildType string) bool {
	switch buildType {
	case "cpp", "flutter", "unity", "rust", "go":
		return true
	}
	return false
}

// customSourceExclude keeps VCS metadata, dependency caches, and stale
// build outputs out of the plugin-builder source archive.
func customSourceExclude(relPath string, isDir bool) bool {
	relPath = strings.TrimPrefix(relPath, "./")
	if relPath == "" {
		return false
	}
	for _, part := range strings.Split(relPath, "/") {
		switch part {
		case ".git", ".hg", ".svn", "node_modules", "__pycache__",
			".venv", "venv", "dist", "out", "build", "target", "bin", "obj":
			return true
		}
	}
	return strings.HasSuffix(relPath, ".DS_Store")
}

// runCustomBuild submits a whole-project build to a plugin builder:
// archive the project, send it with the builder id, then unpack the
// returned artifacts into outDir (default "hybridgrid-out").
func runCustomBuild(builderID string, args []string) error {
	project := "."
	if len(args) > 0 {
		project = args[0]
	}

	archive, hash, err := sourcearchive.Create(project, customSourceExclude)
	if err != nil {
		return fmt.Errorf("failed to archive project: %w", err)
	}

	c, err := client.New(newClientConfig(coordinator, 30*time.Minute))
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer c.Close()

	buildID := taskid.BuildSessionID()
	taskID := taskid.NewTaskID()
	fmt.Printf("Custom build %q (session %s, task %s, source %d bytes)\n",
		builderID, buildID, taskID, len(archive))

	// Builder args come from the positional args after the project
	// path, so `hgbuild build -t mytool . --release --verbose` passes
	// --release --verbose straight through to the plugin.
	var builderArgs []string
	if len(args) > 1 {
		builderArgs = args[1:]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	resp, err := c.Build(ctx, &pb.BuildRequest{
		TaskId:        taskID,
		BuildId:       buildID,
		SourceHash:    hash,
		SourceArchive: archive,
		CustomBuilder: builderID,
		Config: &pb.BuildRequest_CustomConfig{
			CustomConfig: &pb.CustomConfig{Args: builderArgs},
		},
		TimeoutSeconds: 1800,
	})
	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	if resp.Stdout != "" {
		fmt.Print(resp.Stdout)
		if !strings.HasSuffix(resp.Stdout, "\n") {
			fmt.Println()
		}
	}

	if resp.Status != pb.TaskStatus_STATUS_COMPLETED {
		if resp.Stderr != "" {
			fmt.Fprintf(os.Stderr, "builder stderr:\n%s\n", resp.Stderr)
		}
		return fmt.Errorf("builder %q failed (exit %d)", builderID, resp.ExitCode)
	}

	if len(resp.Artifacts) == 0 {
		fmt.Println("Build succeeded (no artifacts).")
		return nil
	}

	outDir := "hybridgrid-out"
	if err := extractArtifactArchive(outDir, resp.Artifacts); err != nil {
		return fmt.Errorf("failed to unpack artifacts: %w", err)
	}
	fmt.Printf("Artifacts written to %s:\n", outDir)
	for _, a := range resp.ArtifactList {
		checksum := a.Checksum
		if len(checksum) > 12 {
			checksum = checksum[:12]
		}
		fmt.Printf("  %-40s %8d bytes  sha256:%s…\n", a.Path, a.SizeBytes, checksum)
	}
	return nil
}

// extractArtifactArchive unpacks a plain-tar artifact archive into dir,
// refusing entries that escape dir (path traversal).
func extractArtifactArchive(dir string, data []byte) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	reader := tar.NewReader(bytes.NewReader(data))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		target := filepath.Join(absDir, filepath.Clean(filepath.FromSlash(header.Name)))
		if !strings.HasPrefix(target, absDir+string(os.PathSeparator)) {
			return fmt.Errorf("artifact %q escapes the output directory", header.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeTarFile(target, reader, header); err != nil {
			return err
		}
	}
}

// writeTarFile streams one tar entry to disk with its declared mode.
func writeTarFile(target string, reader io.Reader, header *tar.Header) error {
	file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode&0o777))
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, reader)
	return err
}

// detectCompiler returns an appropriate compiler based on file extension.
func detectCompiler(file string) string {
	ext := strings.ToLower(filepath.Ext(file))
	switch ext {
	case ".c":
		return "gcc"
	case ".cpp", ".cc", ".cxx":
		return "g++"
	case ".m":
		return "clang"
	case ".mm":
		return "clang++"
	default:
		return "gcc"
	}
}

// parseArch converts architecture string to proto enum.
func parseArch(arch string) pb.Architecture {
	switch strings.ToLower(arch) {
	case "x86_64", "amd64", "x64":
		return pb.Architecture_ARCH_X86_64
	case "arm64", "aarch64":
		return pb.Architecture_ARCH_ARM64
	case "arm", "armv7":
		return pb.Architecture_ARCH_ARMV7
	case "":
		// Default to local architecture
		return getLocalArch()
	default:
		return pb.Architecture_ARCH_UNSPECIFIED
	}
}

func getLocalArch() pb.Architecture {
	switch runtime.GOARCH {
	case "amd64":
		return pb.Architecture_ARCH_X86_64
	case "arm64":
		return pb.Architecture_ARCH_ARM64
	case "arm":
		return pb.Architecture_ARCH_ARMV7
	default:
		return pb.Architecture_ARCH_UNSPECIFIED
	}
}

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	showCmd := &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(cfgFile)
			if err != nil {
				fmt.Printf("No config file found, using defaults\n\n")
				cfg = config.DefaultConfig()
			}

			coordAddr := cfg.Client.CoordinatorAddr
			if coordAddr == "" {
				coordAddr = fmt.Sprintf("localhost:%d", cfg.Coordinator.GRPCPort)
			}
			fmt.Printf("Coordinator: %s\n", coordAddr)
			fmt.Printf("Cache Dir:   %s\n", cfg.Cache.Dir)
			fmt.Printf("Cache Size:  %d MB\n", cfg.Cache.MaxSize)
			fmt.Printf("Log Level:   %s\n", cfg.Log.Level)

			return nil
		},
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Create default configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			path := cfgFile
			if path == "" {
				home, _ := os.UserHomeDir()
				path = filepath.Join(home, ".hybridgrid", "config.yaml")
			}

			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}

			if err := config.WriteExample(path); err != nil {
				return err
			}

			fmt.Printf("Config file created: %s\n", path)
			return nil
		},
	}

	cmd.AddCommand(showCmd, initCmd)
	return cmd
}

func newCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage local cache",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	statsCmd := &cobra.Command{
		Use:   "stats",
		Short: "Show cache statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			output.AutoDetectColors()

			// Use the same cache directory as build service
			buildCfg := build.DefaultConfig()

			store, err := cache.NewStore(buildCfg.CacheDir, buildCfg.CacheMaxSize, buildCfg.CacheTTLHours)
			if err != nil {
				return fmt.Errorf("failed to open cache: %w", err)
			}

			stats := store.Stats()

			// Use colored output
			output.PrintCacheStats(output.CacheStats{
				Directory: buildCfg.CacheDir,
				Entries:   stats.Entries,
				TotalSize: stats.TotalSize,
				MaxSize:   stats.MaxSize,
				TotalHits: stats.TotalHits,
			})

			return nil
		},
	}

	clearCmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Use the same cache directory as build service
			buildCfg := build.DefaultConfig()

			store, err := cache.NewStore(buildCfg.CacheDir, buildCfg.CacheMaxSize, buildCfg.CacheTTLHours)
			if err != nil {
				return fmt.Errorf("failed to open cache: %w", err)
			}

			if err := store.Clear(); err != nil {
				return fmt.Errorf("failed to clear cache: %w", err)
			}

			fmt.Println("Cache cleared")
			return nil
		},
	}

	cmd.AddCommand(statsCmd, clearCmd)
	return cmd
}

// =============================================================================
// Graph Command
// =============================================================================

func newGraphCmd() *cobra.Command {
	var (
		inputFile  string
		outputFile string
		format     string
	)

	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Generate build dependency graph visualization",
		Long: `Parse build files and generate dependency graph visualizations.

Supports:
  - Makefile      Standard GNU Make files
  - compile_commands.json  CMake/Clang compilation database

Output formats:
  - html   Interactive D3.js visualization (default)
  - dot    Graphviz DOT format
  - json   Raw JSON data

Examples:
  hgbuild graph --input Makefile --output graph.html
  hgbuild graph --input compile_commands.json --format dot > deps.dot
  hgbuild graph -i build/compile_commands.json -o deps.html`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if inputFile == "" {
				// Try to auto-detect
				if _, err := os.Stat("Makefile"); err == nil {
					inputFile = "Makefile"
				} else if _, err := os.Stat("compile_commands.json"); err == nil {
					inputFile = "compile_commands.json"
				} else {
					return fmt.Errorf("no input file specified and no Makefile or compile_commands.json found")
				}
			}

			// Parse the build file
			parser := graph.NewParser()
			g, err := parser.ParseAuto(inputFile)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", inputFile, err)
			}

			if g.NodeCount() == 0 {
				fmt.Println("Warning: no dependencies found in build file")
			}

			fmt.Printf("Parsed %d nodes and %d edges\n", g.NodeCount(), g.EdgeCount())

			// Determine output format
			if format == "" {
				format = "html"
			}

			// Generate output
			switch format {
			case "html":
				if outputFile == "" {
					outputFile = "build-graph.html"
				}
				if err := graph.RenderHTML(g, outputFile); err != nil {
					return fmt.Errorf("failed to render HTML: %w", err)
				}
				fmt.Printf("Generated: %s\n", outputFile)

			case "dot":
				if outputFile == "" {
					// Print to stdout
					fmt.Print(g.ToDOT())
				} else {
					if err := graph.RenderDOT(g, outputFile); err != nil {
						return fmt.Errorf("failed to render DOT: %w", err)
					}
					fmt.Printf("Generated: %s\n", outputFile)
				}

			case "json":
				if outputFile == "" {
					data, _ := g.ToJSON()
					fmt.Println(string(data))
				} else {
					if err := graph.RenderJSON(g, outputFile); err != nil {
						return fmt.Errorf("failed to render JSON: %w", err)
					}
					fmt.Printf("Generated: %s\n", outputFile)
				}

			default:
				return fmt.Errorf("unknown format: %s (supported: html, dot, json)", format)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&inputFile, "input", "i", "", "input file (Makefile or compile_commands.json)")
	cmd.Flags().StringVarP(&outputFile, "output", "o", "", "output file (default: stdout for dot/json, build-graph.html for html)")
	cmd.Flags().StringVarP(&format, "format", "f", "html", "output format (html, dot, json)")

	return cmd
}

// =============================================================================
// Compiler Wrappers (cc, c++)
// =============================================================================

func newCCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cc [flags] [files...]",
		Short: "C compiler wrapper (drop-in gcc replacement)",
		Long: `Distributed C compiler wrapper. Use as a drop-in replacement for gcc.

Examples:
  hgbuild cc -c main.c -o main.o
  CC="hgbuild cc" make
  CC="hgbuild cc" cmake --build .`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompiler("gcc", "HG_CC", args)
		},
	}
}

func newCXXCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "c++ [flags] [files...]",
		Short: "C++ compiler wrapper (drop-in g++ replacement)",
		Long: `Distributed C++ compiler wrapper. Use as a drop-in replacement for g++.

Examples:
  hgbuild c++ -c main.cpp -o main.o
  CXX="hgbuild c++" make
  CXX="hgbuild c++" cmake --build .`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompiler("g++", "HG_CXX", args)
		},
	}
}

// filterHgbuildFlags removes hgbuild-specific flags from compiler arguments.
// These flags are parsed by hgbuild but should not be passed to the compiler.
func filterHgbuildFlags(args []string) []string {
	var filtered []string
	skipNext := false

	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}

		// Skip hgbuild-specific flags
		switch {
		case arg == "--coordinator" || arg == "-C":
			skipNext = true // skip next arg (the value)
			continue
		case strings.HasPrefix(arg, "--coordinator="):
			continue
		case arg == "--timeout":
			skipNext = true
			continue
		case strings.HasPrefix(arg, "--timeout="):
			continue
		case arg == "--insecure":
			continue
		case arg == "--no-fallback":
			noFallback = true
			continue
		case arg == "--verbose" || arg == "-v":
			// Set verbose flag and skip
			verbose = true
			continue
		}

		filtered = append(filtered, arg)
	}

	return filtered
}

// runCompiler handles distributed compilation for cc/c++ commands.
func runCompiler(defaultCompiler, envVar string, args []string) error {
	// Check HG_VERBOSE environment variable
	if os.Getenv("HG_VERBOSE") == "1" {
		verbose = true
	}

	// Filter out hgbuild-specific flags from compiler args
	compilerArgs := filterHgbuildFlags(args)

	// Determine compiler from env or default
	comp := os.Getenv(envVar)
	if comp == "" {
		comp = defaultCompiler
	}

	// Parse arguments
	fullArgs := append([]string{comp}, compilerArgs...)
	parsed := compiler.Parse(fullArgs)

	if parsed == nil {
		return fmt.Errorf("failed to parse compiler arguments")
	}

	// Check if this is distributable
	if !parsed.IsDistributable() {
		// Run locally for linking, preprocessing-only, etc.
		if verbose {
			fmt.Fprintf(os.Stderr, "[local] Non-distributable: %s\n", strings.Join(fullArgs, " "))
		}
		return runLocalCompiler(comp, compilerArgs)
	}

	// Get coordinator address (auto-discover if not specified)
	coordAddr := getCoordinatorAddress()
	if coordAddr == "" {
		if !fallbackEnabled() {
			return fmt.Errorf("coordinator unavailable and fallback disabled")
		}

		// No coordinator available, run locally
		if verbose {
			fmt.Fprintln(os.Stderr, "[local] Coordinator unavailable")
		} else {
			fmt.Fprintln(os.Stderr, "Warning: coordinator not available, compiling locally")
		}
		return runLocalCompiler(comp, compilerArgs)
	}

	// Create build service with defaults
	cfg := build.DefaultConfig()
	cfg.CoordinatorAddr = coordAddr
	cfg.Insecure = insecure
	cfg.Timeout = 5 * time.Minute
	cfg.FallbackEnabled = fallbackEnabled()
	cfg.Verbose = verbose

	svc, err := build.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create build service: %w", err)
	}
	defer svc.Close()

	// Connect to coordinator
	clientCfg := newClientConfig(coordAddr, timeout)
	c, err := client.New(clientCfg)
	if err != nil {
		if !fallbackEnabled() {
			return fmt.Errorf("coordinator unavailable and fallback disabled: %w", err)
		}

		// Fallback to local
		if verbose {
			fmt.Fprintf(os.Stderr, "[local] Connection failed: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Warning: coordinator not available, compiling locally")
		}
		return runLocalCompiler(comp, compilerArgs)
	}
	svc.SetClient(c)

	// Determine output file
	outputFile := parsed.OutputFile
	if outputFile == "" && len(parsed.InputFiles) > 0 {
		// Default: replace extension with .o
		base := strings.TrimSuffix(parsed.InputFiles[0], filepath.Ext(parsed.InputFiles[0]))
		outputFile = base + ".o"
	}

	// Build request
	req := &build.Request{
		TaskID:     taskid.NewTaskID(),
		BuildID:    wrappedBuildID(),
		SourceFile: parsed.InputFiles[0],
		OutputFile: outputFile,
		Args:       parsed,
		TargetArch: parseArch(parsed.TargetArch),
		Timeout:    5 * time.Minute,
	}

	ctx := context.Background()
	result, err := svc.Build(ctx, req)
	if err != nil {
		return err
	}

	// Check exit code
	if result.ExitCode != 0 {
		if result.Stderr != "" {
			fmt.Fprint(os.Stderr, result.Stderr)
		}
		os.Exit(result.ExitCode)
	}

	// Write output file
	if err := writeOutputFile(outputFile, result.ObjectFile); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	// Print status
	if verbose {
		status := "[remote]"
		if result.CacheHit {
			status = "[cache]"
		} else if result.Fallback {
			status = "[local]"
		}
		fmt.Fprintf(os.Stderr, "%s %s -> %s (%.2fs)\n",
			status, parsed.InputFiles[0], outputFile, result.Duration.Seconds())
	}

	return nil
}

// runLocalCompiler runs the compiler locally (for non-distributable operations).
func runLocalCompiler(compiler string, args []string) error {
	resolvedCompiler, err := resolveLocalCompilerPath(compiler)
	if err != nil {
		return err
	}

	// #nosec G204,G702 -- compiler path is validated and resolved before execution.
	cmd := exec.Command(resolvedCompiler, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func writeOutputFile(outputFile string, data []byte) error {
	cleanPath, err := resolveOutputPath(outputFile)
	if err != nil {
		return err
	}

	baseDir := filepath.Dir(cleanPath)
	fileName := filepath.Base(cleanPath)
	// #nosec G301,G302,G703 -- cleanPath is validated by resolveOutputPath before directory creation.
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	root, err := os.OpenRoot(baseDir)
	if err != nil {
		return fmt.Errorf("failed to open output directory: %w", err)
	}
	defer root.Close()

	file, err := root.OpenFile(fileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open output file: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(data); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	if err := file.Sync(); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func resolveOutputPath(outputFile string) (string, error) {
	if outputFile == "" {
		return "", fmt.Errorf("output path cannot be empty")
	}

	baseDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	cleanPath := validation.SanitizePath(baseDir, outputFile)
	if cleanPath == "" {
		return "", fmt.Errorf("invalid output path: %s", outputFile)
	}

	if !filepath.IsAbs(cleanPath) {
		cleanPath = filepath.Join(baseDir, cleanPath)
	}

	return cleanPath, nil
}

func resolveLocalCompilerPath(compiler string) (string, error) {
	if compiler == "" {
		return "", fmt.Errorf("compiler cannot be empty")
	}

	if strings.ContainsAny(compiler, "\r\n") {
		return "", fmt.Errorf("compiler contains invalid characters")
	}

	if strings.ContainsAny(compiler, `/\`) {
		cleanPath := validation.SanitizePath("", compiler)
		if cleanPath == "" {
			return "", fmt.Errorf("invalid compiler path: %s", compiler)
		}

		absPath, err := filepath.Abs(cleanPath)
		if err != nil {
			return "", fmt.Errorf("failed to resolve compiler path: %w", err)
		}

		// #nosec G304,G703 -- absPath is validated by SanitizePath and filepath.Abs before stat.
		info, err := os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("failed to stat compiler path: %w", err)
		}
		if info.IsDir() {
			return "", fmt.Errorf("compiler path is a directory: %s", compiler)
		}

		return absPath, nil
	}

	if !isSafeExecutableName(compiler) {
		return "", fmt.Errorf("invalid compiler name: %s", compiler)
	}

	resolvedCompiler, err := exec.LookPath(compiler)
	if err != nil {
		return "", fmt.Errorf("failed to find compiler %s: %w", compiler, err)
	}

	return resolvedCompiler, nil
}

func isSafeExecutableName(name string) bool {
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			continue
		}
		switch r {
		case '.', '_', '-', '+':
			continue
		default:
			return false
		}
	}

	return true
}

// getCoordinatorAddress gets the coordinator address from flags, env, or mDNS.
func getCoordinatorAddress() string {
	// 1. Check command-line flag
	if coordinator != "" {
		return coordinator
	}

	// 2. Check environment variable
	if addr := os.Getenv("HG_COORDINATOR"); addr != "" {
		return addr
	}

	// 3. Try mDNS auto-discovery
	browser := mdns.NewCoordBrowser(mdns.CoordBrowserConfig{
		Timeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	coord, err := browser.Discover(ctx)
	if err == nil && coord != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "[mdns] Discovered coordinator at %s\n", coord.Address)
		}
		return coord.Address
	}

	return ""
}

// =============================================================================
// Build Wrappers (make, ninja, wrap)
// =============================================================================

func newMakeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "make [make-args...]",
		Short: "Run make with distributed compilation",
		Long: `Wrap make with distributed compilation by setting CC/CXX automatically.

Examples:
  hgbuild make
  hgbuild make -j8
  hgbuild make clean all`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return wrapBuildCommand("make", filterHgbuildWrapperFlags(args))
		},
	}
}

func newNinjaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ninja [ninja-args...]",
		Short: "Run ninja with distributed compilation",
		Long: `Wrap ninja with distributed compilation by setting CC/CXX automatically.

Examples:
  hgbuild ninja
  hgbuild ninja -j8`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return wrapBuildCommand("ninja", filterHgbuildWrapperFlags(args))
		},
	}
}

func newWrapCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "wrap <command> [args...]",
		Short: "Wrap any build command with distributed compilation",
		Long: `Wrap any build command with distributed compilation.
Sets CC and CXX to use hgbuild for distributed compilation.

Examples:
  hgbuild wrap cmake --build .
  hgbuild wrap ./build.sh`,
		DisableFlagParsing: true,
		SilenceUsage:       true,
		RunE: func(cmd *cobra.Command, args []string) error {
			filteredArgs := filterHgbuildWrapperFlags(args)
			if len(filteredArgs) == 0 {
				return fmt.Errorf("no command specified")
			}
			return wrapBuildCommand(filteredArgs[0], filteredArgs[1:])
		},
	}
}

func filterHgbuildWrapperFlags(args []string) []string {
	filtered := make([]string, 0, len(args))

	for _, arg := range args {
		switch {
		case arg == "--no-fallback":
			noFallback = true
			continue
		case strings.HasPrefix(arg, "--no-fallback="):
			value := strings.TrimSpace(strings.TrimPrefix(arg, "--no-fallback="))
			if parsed, err := strconv.ParseBool(value); err == nil {
				if parsed {
					noFallback = true
				}
				continue
			}
		}

		filtered = append(filtered, arg)
	}

	return filtered
}

// wrapBuildCommand wraps a build command with CC/CXX set to hgbuild.
func wrapBuildCommand(command string, args []string) error {
	// Find our own executable
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find hgbuild executable: %w", err)
	}

	wrapperDir, err := os.MkdirTemp("", "hgbuild-wrap-")
	if err != nil {
		return fmt.Errorf("failed to create wrapper dir: %w", err)
	}
	defer os.RemoveAll(wrapperDir)

	ccValue, err := createCompilerWrapper(wrapperDir, "hgcc", self, wrapperCCEnv)
	if err != nil {
		return err
	}

	cxxValue, err := createCompilerWrapper(wrapperDir, "hgcxx", self, wrapperCXXEnv)
	if err != nil {
		return err
	}

	// Build environment. Every translation unit the wrapped tool compiles
	// inherits one HG_BUILD_ID, so the coordinator groups them as one build.
	env := ensureBuildSession(os.Environ())

	// Set CC and CXX environment variables (for build systems that respect them)
	env = setEnv(env, "CC", ccValue)
	env = setEnv(env, "CXX", cxxValue)

	// Pass through coordinator address if specified
	if coordinator != "" {
		env = setEnv(env, "HG_COORDINATOR", coordinator)
	}

	if !fallbackEnabled() {
		env = setEnv(env, noFallbackEnv, "1")
	}

	// Pass through verbose flag
	if verbose {
		env = setEnv(env, "HG_VERBOSE", "1")
	}

	// For make, also pass CC/CXX as command-line arguments
	// This overrides Makefile assignments (which take precedence over env vars)
	finalArgs := args
	if command == "make" {
		finalArgs = append([]string{"CC=" + ccValue, "CXX=" + cxxValue}, args...)
	}

	if verbose {
		fmt.Fprintf(os.Stderr, "[wrap] CC=%s\n", ccValue)
		fmt.Fprintf(os.Stderr, "[wrap] CXX=%s\n", cxxValue)
		fmt.Fprintf(os.Stderr, "[wrap] Running: %s %s\n", command, strings.Join(finalArgs, " "))
	}

	// Execute wrapped command
	cmd := exec.Command(command, finalArgs...)
	cmd.Env = env
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func injectWrappedCompilerMode() {
	mode := ""
	switch {
	case os.Getenv(wrapperCCEnv) == "1":
		mode = "cc"
	case os.Getenv(wrapperCXXEnv) == "1":
		mode = "c++"
	}

	if mode == "" {
		return
	}

	if len(os.Args) > 1 && (os.Args[1] == "cc" || os.Args[1] == "c++") {
		return
	}

	os.Args = append([]string{os.Args[0], mode}, os.Args[1:]...)
}

func createCompilerWrapper(dir, name, self, modeEnv string) (string, error) {
	path := filepath.Join(dir, name)
	content := fmt.Sprintf("#!/bin/sh\nexport %s=1\nexec %q \"$@\"\n", modeEnv, self)
	if err := os.WriteFile(path, []byte(content), 0755); err != nil {
		return "", fmt.Errorf("failed to write compiler wrapper %s: %w", name, err)
	}

	return path, nil
}

// setEnv sets an environment variable in the env slice.
// ensureBuildSession returns env with HG_BUILD_ID set, generating a fresh
// session id when the caller did not export one.
func ensureBuildSession(env []string) []string {
	if envValue(env, taskid.BuildIDEnvVar) != "" {
		return env
	}
	return setEnv(env, taskid.BuildIDEnvVar, taskid.BuildSessionID())
}

// wrappedBuildID is the build session for a single cc/c++ invocation. It is
// only set when HG_BUILD_ID is exported (by `hgbuild make` or the user): a
// lone compiler call must not mint its own one-task "build", which would
// crowd real builds out of the coordinator's bounded build history.
func wrappedBuildID() string {
	if os.Getenv(taskid.BuildIDEnvVar) == "" {
		return ""
	}
	return taskid.BuildSessionID()
}

// envValue returns the value of key in an os.Environ-style slice.
func envValue(env []string, key string) string {
	prefix := key + "="
	for _, e := range env {
		if strings.HasPrefix(e, prefix) {
			return strings.TrimPrefix(e, prefix)
		}
	}
	return ""
}

func setEnv(env []string, key, value string) []string {
	prefix := key + "="
	for i, e := range env {
		if strings.HasPrefix(e, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func fallbackEnabled() bool {
	if noFallback {
		return false
	}

	value := strings.TrimSpace(os.Getenv(noFallbackEnv))
	return value == ""
}

func newClientConfig(address string, requestTimeout time.Duration) client.Config {
	clientCfg := client.Config{
		Address:       address,
		Insecure:      insecure,
		Timeout:       requestTimeout,
		EnableTracing: tracingEnable,
	}

	if tlsCert != "" && tlsKey != "" {
		clientCfg.TLS.Enabled = true
		clientCfg.TLS.CertFile = tlsCert
		clientCfg.TLS.KeyFile = tlsKey
		clientCfg.TLS.ClientCA = tlsCA
		clientCfg.Insecure = false
	}

	return clientCfg
}

// preprocessLocal expands #include directives with the local toolchain
// so the worker compiles a self-contained unit. Returns nil on failure;
// the caller then falls back to sending raw source.
func preprocessLocal(file, comp string, compArgs []string) []byte {
	parsed := compiler.Parse(append([]string{comp}, compArgs...))
	if parsed == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := compiler.NewPreprocessor(compiler.DefaultPreprocessorConfig()).Preprocess(ctx, parsed, file)
	if err != nil {
		fmt.Printf(" (local preprocess failed: %v; sending raw source)", err)
		return nil
	}
	return res.PreprocessedSource
}

// cacheKeyFor builds the client-side cache key for one translation unit.
// Package-level because the compile command shadows the compiler package
// name with its --compiler flag variable.
func cacheKeyFor(comp string, compArgs []string, arch pb.Architecture, source []byte) string {
	parsed := compiler.Parse(append([]string{comp}, compArgs...))
	if parsed == nil {
		return ""
	}
	return build.CacheKeyRaw(parsed, arch, source)
}

// clientArch returns the machine architecture as the protobuf enum.
func clientArch() pb.Architecture {
	switch runtime.GOARCH {
	case "amd64":
		return pb.Architecture_ARCH_X86_64
	case "arm64":
		return pb.Architecture_ARCH_ARM64
	case "arm":
		return pb.Architecture_ARCH_ARMV7
	default:
		return pb.Architecture_ARCH_UNSPECIFIED
	}
}

// collectCrossArchIncludes bundles project-local headers (-I paths) so a
// raw-source cross-arch request can carry them. The worker's own toolchain
// supplies system headers for its architecture. Returns nil on failure —
// the build proceeds without project includes and likely fails loudly at
// the worker rather than silently miscompiling.
func collectCrossArchIncludes(compArgs []string) map[string][]byte {
	parsed := compiler.Parse(append([]string{"cc"}, compArgs...))
	if parsed == nil {
		return nil
	}
	incs, err := build.CollectIncludeFiles(parsed)
	if err != nil {
		log.Warn().Err(err).Msg("collecting include files for cross-arch build; continuing without them")
		return nil
	}
	return incs
}
