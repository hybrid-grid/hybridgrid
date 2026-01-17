package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

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
		buildType string
		output    string
		compiler  string
	)

	cmd := &cobra.Command{
		Use:   "build [files...]",
		Short: "Submit a build job to the coordinator",
		Long: `Submit source files for distributed compilation.

Examples:
  hgbuild build main.c            Compile single file
  hgbuild build *.c -o myapp      Compile multiple files
  hgbuild build --type=cpp src/   Compile directory`,
		Args: cobra.MinimumNArgs(1),
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

			fmt.Printf("Submitting build job to %s...\n", coordinator)
			fmt.Printf("Files: %v\n", args)
			fmt.Printf("Type:  %s\n", buildType)

			// TODO: Implement actual build submission
			// For now, show that the command structure works
			fmt.Println("\nBuild submission not fully implemented yet.")
			fmt.Println("Use hg-coord and hg-worker directly for compilation.")

			return nil
		},
	}

	cmd.Flags().StringVarP(&buildType, "type", "t", "cpp", "build type (cpp, flutter, unity, rust, go)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "output file/directory")
	cmd.Flags().StringVar(&compiler, "compiler", "", "compiler to use (auto-detect if empty)")

	return cmd
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
