package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/cache"
	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
)

var (
	version     = "v0.0.0-dev"
	cfgFile     string
	coordinator string
	insecure    bool
	timeout     time.Duration
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "hgbuild",
		Short: "Hybrid-Grid Build - Distributed multi-platform build system",
		Long: `hgbuild is a CLI client for the Hybrid-Grid Build system.
It intercepts compiler commands and distributes them to remote workers.

Quick start:
  hgbuild status              Check coordinator status
  hgbuild workers             List available workers
  hgbuild build <file>        Submit a build job
  hgbuild cache stats         View cache statistics`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default: ~/.hybridgrid/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&coordinator, "coordinator", "C", "localhost:50051", "coordinator address")
	rootCmd.PersistentFlags().BoolVar(&insecure, "insecure", true, "use insecure connection")
	rootCmd.PersistentFlags().DurationVar(&timeout, "timeout", 10*time.Second, "connection timeout")

	// Commands
	rootCmd.AddCommand(
		newVersionCmd(),
		newStatusCmd(),
		newWorkersCmd(),
		newBuildCmd(),
		newConfigCmd(),
		newCacheCmd(),
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
	rand.Read(b)
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
			cfg, _ := config.Load(cfgFile)
			if cfg == nil {
				cfg = config.DefaultConfig()
			}

			store, err := cache.NewStore(cfg.Cache.Dir, cfg.Cache.MaxSize, cfg.Cache.TTLHours)
			if err != nil {
				return fmt.Errorf("failed to open cache: %w", err)
			}

			stats := store.Stats()
			fmt.Printf("Cache Directory: %s\n", cfg.Cache.Dir)
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
			cfg, _ := config.Load(cfgFile)
			if cfg == nil {
				cfg = config.DefaultConfig()
			}

			store, err := cache.NewStore(cfg.Cache.Dir, cfg.Cache.MaxSize, cfg.Cache.TTLHours)
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
