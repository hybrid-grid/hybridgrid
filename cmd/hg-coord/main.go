package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	coordserver "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/dashboard"
)

var version = "v0.0.0-dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	rootCmd := &cobra.Command{
		Use:   "hg-coord",
		Short: "Hybrid-Grid Build Coordinator",
		Long: `hg-coord is the coordinator component of the Hybrid-Grid Build system.
It manages worker registration, task scheduling, and provides the dashboard.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hg-coord %s\n", version)
		},
	}

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the coordinator server",
		RunE: func(cmd *cobra.Command, args []string) error {
			grpcPort, _ := cmd.Flags().GetInt("grpc-port")
			httpPort, _ := cmd.Flags().GetInt("http-port")
			token, _ := cmd.Flags().GetString("token")
			noMdns, _ := cmd.Flags().GetBool("no-mdns")

			log.Info().
				Int("grpc_port", grpcPort).
				Int("http_port", httpPort).
				Str("version", version).
				Msg("Starting Hybrid-Grid Coordinator")

			cfg := coordserver.DefaultConfig()
			cfg.Port = grpcPort
			cfg.AuthToken = token
			cfg.HeartbeatTTL = 60 * time.Second
			cfg.RequestTimeout = 120 * time.Second

			srv := coordserver.New(cfg)

			// Handle shutdown signals
			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

			errCh := make(chan error, 2)
			go func() {
				if err := srv.Start(); err != nil {
					errCh <- fmt.Errorf("gRPC server: %w", err)
				}
			}()

			// Start HTTP dashboard server
			dashCfg := dashboard.DefaultConfig()
			dashCfg.Port = httpPort
			dashSrv := dashboard.New(dashCfg, srv.NewStatsProvider())

			// Wire up event notifications from coordinator to dashboard
			onStart, onComplete := dashSrv.CreateEventNotifier()
			srv.SetEventNotifier(&eventNotifierWrapper{onStart: onStart, onComplete: onComplete})

			go func() {
				if err := dashSrv.Start(); err != nil {
					errCh <- fmt.Errorf("dashboard server: %w", err)
				}
			}()

			log.Info().Int("port", httpPort).Msg("Dashboard server started")

			// Start mDNS announcer (unless disabled)
			var mdnsAnnouncer *mdns.CoordAnnouncer
			if !noMdns {
				hostname, _ := os.Hostname()
				mdnsAnnouncer = mdns.NewCoordAnnouncer(mdns.CoordAnnouncerConfig{
					Instance:   fmt.Sprintf("hg-coord-%s", hostname),
					GRPCPort:   grpcPort,
					HTTPPort:   httpPort,
					Version:    version,
					InstanceID: fmt.Sprintf("%s-%d", hostname, os.Getpid()),
				})

				if err := mdnsAnnouncer.Start(); err != nil {
					log.Warn().Err(err).Msg("Failed to start mDNS announcer (continuing without)")
				} else {
					log.Info().
						Str("service", mdns.CoordServiceType).
						Msg("Coordinator discoverable via mDNS")
				}
			}

			select {
			case sig := <-sigCh:
				log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
				if mdnsAnnouncer != nil {
					mdnsAnnouncer.Stop()
				}
				dashSrv.Stop()
				srv.Stop()
				return nil
			case err := <-errCh:
				return fmt.Errorf("server error: %w", err)
			}
		},
	}

	serveCmd.Flags().Int("grpc-port", 9000, "gRPC server port")
	serveCmd.Flags().Int("http-port", 8080, "HTTP/Dashboard port")
	serveCmd.Flags().String("config", "", "Path to config file")
	serveCmd.Flags().String("token", "", "Authentication token")
	serveCmd.Flags().Bool("no-mdns", false, "Disable mDNS advertisement")

	rootCmd.AddCommand(versionCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// eventNotifierWrapper adapts dashboard callbacks to coordinator's EventNotifier interface.
type eventNotifierWrapper struct {
	onStart    func(id, buildType, status, workerID string, startedAt int64)
	onComplete func(id, buildType, status, workerID string, startedAt, completedAt, durationMs int64, exitCode int32, errorMsg string)
}

func (w *eventNotifierWrapper) NotifyTaskStarted(event *coordserver.TaskEvent) {
	if w.onStart != nil {
		w.onStart(event.ID, event.BuildType, event.Status, event.WorkerID, event.StartedAt)
	}
}

func (w *eventNotifierWrapper) NotifyTaskCompleted(event *coordserver.TaskEvent) {
	if w.onComplete != nil {
		w.onComplete(event.ID, event.BuildType, event.Status, event.WorkerID, event.StartedAt, event.CompletedAt, event.DurationMs, event.ExitCode, event.ErrorMessage)
	}
}
