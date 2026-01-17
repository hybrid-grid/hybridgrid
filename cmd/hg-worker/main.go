package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	workerserver "github.com/h3nr1-d14z/hybridgrid/internal/worker/server"
)

var version = "v0.0.0-dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd := &cobra.Command{
		Use:   "hg-worker",
		Short: "Hybrid-Grid Build Worker Agent",
		Long: `hg-worker is the worker agent component of the Hybrid-Grid Build system.
It executes build tasks received from the coordinator.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hg-worker %s\n", version)
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the worker agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			coordinator, _ := cmd.Flags().GetString("coordinator")
			port, _ := cmd.Flags().GetInt("port")
			httpPort, _ := cmd.Flags().GetInt("http-port")
			token, _ := cmd.Flags().GetString("token")
			maxParallel, _ := cmd.Flags().GetInt("max-parallel")
			discoveryTimeout, _ := cmd.Flags().GetDuration("discovery-timeout")

			// Resolve coordinator address
			if coordinator == "" {
				log.Info().Dur("timeout", discoveryTimeout).Msg("No coordinator specified, trying mDNS discovery")

				browser := mdns.NewCoordBrowser(mdns.CoordBrowserConfig{
					Timeout: discoveryTimeout,
				})

				// Check for env var fallback
				envCoord := os.Getenv("HG_COORDINATOR")

				coordAddr, err := browser.DiscoverWithFallback(context.Background(), envCoord)
				if err != nil {
					return fmt.Errorf("coordinator discovery failed: %w\n\nHint: start coordinator with mDNS enabled, or specify --coordinator flag, or set HG_COORDINATOR env var", err)
				}
				coordinator = coordAddr
			}

			if maxParallel == 0 {
				maxParallel = runtime.NumCPU()
			}

			hostname, _ := os.Hostname()
			log.Info().
				Str("coordinator", coordinator).
				Str("hostname", hostname).
				Int("port", port).
				Int("max_parallel", maxParallel).
				Str("version", version).
				Msg("Starting Hybrid-Grid Worker")

			// Create and start worker gRPC server
			cfg := workerserver.DefaultConfig()
			cfg.Port = port
			cfg.MaxConcurrent = maxParallel

			srv := workerserver.New(cfg)
			caps := srv.Capabilities()
			caps.WorkerId = fmt.Sprintf("worker-%s", hostname)
			caps.MaxParallelTasks = int32(maxParallel)
			caps.Version = version

			// Connect to coordinator
			cli, err := client.New(client.Config{
				Address:   coordinator,
				AuthToken: token,
				Timeout:   30 * time.Second,
				Insecure:  true, // TODO: Add TLS support
			})
			if err != nil {
				return fmt.Errorf("failed to connect to coordinator: %w", err)
			}
			defer cli.Close()

			// Register with coordinator
			regReq := &pb.HandshakeRequest{
				Capabilities:  caps,
				AuthToken:     token,
				WorkerAddress: fmt.Sprintf("%s:%d", hostname, port),
			}
			resp, err := cli.Handshake(context.Background(), regReq)
			if err != nil {
				return fmt.Errorf("handshake failed: %w", err)
			}

			if !resp.Accepted {
				return fmt.Errorf("worker registration rejected: %s", resp.Message)
			}

			log.Info().
				Str("worker_id", resp.AssignedWorkerId).
				Int32("heartbeat_interval", resp.HeartbeatIntervalSeconds).
				Msg("Worker registered successfully")

			// Handle shutdown signals
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			errCh := make(chan error, 2)
			go func() {
				if err := srv.Start(); err != nil {
					errCh <- fmt.Errorf("gRPC server: %w", err)
				}
			}()

			// Start metrics HTTP server
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", promhttp.Handler())
			metricsMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})
			metricsServer := &http.Server{
				Addr:    fmt.Sprintf(":%d", httpPort),
				Handler: metricsMux,
			}
			go func() {
				log.Info().Int("port", httpPort).Msg("Worker metrics server started")
				if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- fmt.Errorf("metrics server: %w", err)
				}
			}()

			// Start heartbeat loop
			heartbeatInterval := time.Duration(resp.HeartbeatIntervalSeconds) * time.Second
			go func() {
				ticker := time.NewTicker(heartbeatInterval)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						hResp, err := cli.HealthCheck(context.Background())
						if err != nil {
							log.Warn().Err(err).Msg("Heartbeat failed")
						} else {
							log.Debug().Bool("healthy", hResp.Healthy).Msg("Heartbeat sent")
						}
					case <-sigCh:
						return
					}
				}
			}()

			select {
			case sig := <-sigCh:
				log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				metricsServer.Shutdown(ctx)
				srv.Stop()
				return nil
			case err := <-errCh:
				return fmt.Errorf("server error: %w", err)
			}
		},
	}

	serveCmd.Flags().Int("port", 50052, "Worker gRPC port")
	serveCmd.Flags().Int("http-port", 9090, "Worker HTTP/metrics port")
	serveCmd.Flags().String("coordinator", "", "Coordinator address (empty for mDNS auto-discovery)")
	serveCmd.Flags().String("config", "", "Path to config file")
	serveCmd.Flags().String("token", "", "Authentication token")
	serveCmd.Flags().Int("max-parallel", 0, "Max parallel tasks (0 = auto)")
	serveCmd.Flags().Duration("discovery-timeout", 10*time.Second, "mDNS discovery timeout")

	rootCmd.AddCommand(versionCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
