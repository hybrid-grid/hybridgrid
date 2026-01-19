package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/cli/build"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
)

var (
	version     = "v0.0.0-dev"
	cfgFile     string
	coordinator string
	insecure    bool
	timeout     time.Duration
	verbose     bool
)

func main() {
	// Configure logging
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

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
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.hybridgrid/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&coordinator, "coordinator", "C", "", "coordinator address (auto-discover if empty)")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "use insecure connection")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 2*time.Minute, "connection timeout")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	// Commands
	rootCmd.AddCommand(
		newVersionCmd(),
		newStatusCmd(),
		newWorkersCmd(),
		newBuildCmd(),
		newConfigCmd(),
		newCacheCmd(),
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
			c, err := client.New(client.Config{
				Address:  coordinator,
				Insecure: insecure,
				Timeout:  timeout,
			})
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

			fmt.Printf("Coordinator: %s\n", coordinator)
			fmt.Printf("Status:      %s\n", statusEmoji(resp.Healthy))
			fmt.Printf("Active:      %d tasks\n", resp.ActiveTasks)
			fmt.Printf("Queued:      %d tasks\n", resp.QueuedTasks)

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
			c, err := client.New(client.Config{
				Address:  coordinator,
				Insecure: insecure,
				Timeout:  timeout,
			})
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

			if len(resp.Workers) == 0 {
				fmt.Println("No workers connected")
				return nil
			}

			fmt.Printf("Workers: %d total, %d healthy\n\n", resp.TotalWorkers, resp.HealthyWorkers)

			if verbose {
				fmt.Printf("%-20s %-10s %-8s %-8s %-10s %-8s\n",
					"ID", "ARCH", "CORES", "MEM(GB)", "TASKS", "STATUS")
				fmt.Println("-------------------- ---------- -------- -------- ---------- --------")
			} else {
				fmt.Printf("%-20s %-10s %-8s %-8s\n", "ID", "ARCH", "CORES", "STATUS")
				fmt.Println("-------------------- ---------- -------- --------")
			}

			for _, w := range resp.Workers {
				status := "healthy"
				if w.CircuitState != "" && w.CircuitState != "CLOSED" {
					status = w.CircuitState
				}

				if verbose {
					fmt.Printf("%-20s %-10s %-8d %-8.1f %-10d %-8s\n",
						truncate(w.WorkerId, 20),
						w.NativeArch.String(),
						w.CpuCores,
						float64(w.MemoryBytes)/(1024*1024*1024),
						w.ActiveTasks,
						status)
				} else {
					fmt.Printf("%-20s %-10s %-8d %-8s\n",
						truncate(w.WorkerId, 20),
						w.NativeArch.String(),
						w.CpuCores,
						status)
				}
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
			c, err := client.New(client.Config{
				Address:  coordinator,
				Insecure: insecure,
				Timeout:  5 * time.Minute, // Builds take longer
			})
			if err != nil {
				return fmt.Errorf("failed to connect: %w", err)
			}
			defer c.Close()

			// Process each file
			successCount := 0
			failCount := 0

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

				// Generate task ID
				taskID := generateTaskID()

				// Determine output file
				outFile := output
				if outFile == "" {
					outFile = strings.TrimSuffix(file, filepath.Ext(file)) + ".o"
				}

				// Parse target architecture
				arch := parseArch(targetArch)

				if verbose {
					fmt.Printf("Compiling %s → %s (compiler: %s)\n", file, outFile, comp)
				} else {
					fmt.Printf("Compiling %s...", file)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

				// Create compile request
				req := &pb.CompileRequest{
					TaskId:             taskID,
					Compiler:           comp,
					CompilerArgs:       compArgs,
					PreprocessedSource: source,
					TargetArch:         arch,
					TimeoutSeconds:     300,
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

			// Summary
			fmt.Printf("\nResults: %d succeeded, %d failed\n", successCount, failCount)

			if failCount > 0 {
				return fmt.Errorf("%d files failed to compile", failCount)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&buildType, "type", "t", "cpp", "build type (cpp, flutter, unity, rust, go)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file (for single file builds)")
	cmd.Flags().StringVar(&compiler, "compiler", "", "compiler to use (auto-detect if empty)")
	cmd.Flags().StringSliceVar(&compArgs, "args", nil, "compiler arguments")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
	cmd.Flags().StringVar(&targetArch, "arch", "", "target architecture (x86_64, arm64)")

	return cmd
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

// generateTaskID creates a unique task identifier.
func generateTaskID() string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-only if crypto/rand fails
		return fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("task-%s-%d", hex.EncodeToString(b), time.Now().UnixNano()%10000)
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
			// Use the same cache directory as build service
			buildCfg := build.DefaultConfig()

			store, err := cache.NewStore(buildCfg.CacheDir, buildCfg.CacheMaxSize, buildCfg.CacheTTLHours)
			if err != nil {
				return fmt.Errorf("failed to open cache: %w", err)
			}

			stats := store.Stats()
			fmt.Printf("Cache Directory: %s\n", buildCfg.CacheDir)
			fmt.Printf("Entries:         %d\n", stats.Entries)
			fmt.Printf("Size:            %.2f MB / %.0f MB\n",
				float64(stats.TotalSize)/(1024*1024), float64(stats.MaxSize)/(1024*1024))
			fmt.Printf("Total Hits:      %d\n", stats.TotalHits)

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

// Helper functions

func statusEmoji(healthy bool) string {
	if healthy {
		return "healthy ✓"
	}
	return "unhealthy ✗"
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
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

	for i, arg := range args {
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
		case arg == "--config" || arg == "-c":
			// Note: -c is also compiler flag for compile-only, but we only skip
			// if next arg looks like a file path
			if i+1 < len(args) && (strings.HasSuffix(args[i+1], ".yaml") || strings.HasSuffix(args[i+1], ".yml")) {
				skipNext = true
				continue
			}
		case strings.HasPrefix(arg, "--config="):
			continue
		case arg == "--timeout":
			skipNext = true
			continue
		case strings.HasPrefix(arg, "--timeout="):
			continue
		case arg == "--insecure":
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
		// No coordinator available, run locally
		if verbose {
			fmt.Fprintf(os.Stderr, "[local] No coordinator available\n")
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
	cfg.FallbackEnabled = true
	cfg.Verbose = verbose

	svc, err := build.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create build service: %w", err)
	}
	defer svc.Close()

	// Connect to coordinator
	clientCfg := client.Config{
		Address:  coordAddr,
		Insecure: insecure,
		Timeout:  timeout,
	}
	c, err := client.New(clientCfg)
	if err != nil {
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
		TaskID:     generateTaskID(),
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
	if err := os.WriteFile(outputFile, result.ObjectFile, 0644); err != nil {
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
	cmd := exec.Command(compiler, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
			return wrapBuildCommand("make", args)
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
			return wrapBuildCommand("ninja", args)
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
			if len(args) == 0 {
				return fmt.Errorf("no command specified")
			}
			return wrapBuildCommand(args[0], args[1:])
		},
	}
}

// wrapBuildCommand wraps a build command with CC/CXX set to hgbuild.
func wrapBuildCommand(command string, args []string) error {
	// Find our own executable
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to find hgbuild executable: %w", err)
	}

	ccValue := self + " cc"
	cxxValue := self + " c++"

	// Build environment
	env := os.Environ()

	// Set CC and CXX environment variables (for build systems that respect them)
	env = setEnv(env, "CC", ccValue)
	env = setEnv(env, "CXX", cxxValue)

	// Pass through coordinator address if specified
	if coordinator != "" {
		env = setEnv(env, "HG_COORDINATOR", coordinator)
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

// setEnv sets an environment variable in the env slice.
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
