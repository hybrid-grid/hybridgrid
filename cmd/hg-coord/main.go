package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	coordserver "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/dashboard"
	observabilitymetrics "github.com/h3nr1-d14z/hybridgrid/internal/observability/metrics"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
)

var version = "v0.0.0-dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Load default config (validation will happen in serveCmd.RunE after CLI flags applied)
	cfg := config.DefaultConfig()

	// Setup logger
	logger, logCloser, err := logging.SetupLogger(cfg.Log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup logger")
	}
	log.Logger = logger
	defer logCloser.Close()

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
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			grpcPort, _ := cmd.Flags().GetInt("grpc-port")
			httpPort, _ := cmd.Flags().GetInt("http-port")
			token, _ := cmd.Flags().GetString("token")
			noMdns, _ := cmd.Flags().GetBool("no-mdns")
			schedulerType, _ := cmd.Flags().GetString("scheduler")
			taskLogPath, _ := cmd.Flags().GetString("task-log")
			epsilonValue, _ := cmd.Flags().GetFloat64("epsilon")
			alphaValue, _ := cmd.Flags().GetFloat64("alpha")

			// Validate scheduler choice (fail fast rather than silent fallback).
			validSchedulers := map[string]bool{"leastloaded": true, "simple": true, "p2c": true, "epsilon-greedy": true, "linucb": true}
			if !validSchedulers[schedulerType] {
				return fmt.Errorf("invalid --scheduler %q; must be one of: leastloaded, simple, p2c, epsilon-greedy, linucb", schedulerType)
			}
			if epsilonValue < 0 || epsilonValue > 1 {
				return fmt.Errorf("invalid --epsilon %v; must be in [0, 1]", epsilonValue)
			}
			if alphaValue < 0 || alphaValue > 10 {
				return fmt.Errorf("invalid --alpha %v; must be in [0, 10]", alphaValue)
			}

			// Validate port ranges
			if grpcPort < 1 || grpcPort > 65535 {
				return fmt.Errorf("invalid configuration: coordinator.grpc_port must be 1-65535, got %d", grpcPort)
			}
			if httpPort < 1 || httpPort > 65535 {
				return fmt.Errorf("invalid configuration: coordinator.http_port must be 1-65535, got %d", httpPort)
			}
			if grpcPort == httpPort {
				return fmt.Errorf("invalid configuration: coordinator.grpc_port and coordinator.http_port must be different, got %d for both", grpcPort)
			}

			// TLS flags
			tlsCert, _ := cmd.Flags().GetString("tls-cert")
			tlsKey, _ := cmd.Flags().GetString("tls-key")
			tlsCA, _ := cmd.Flags().GetString("tls-ca")
			tlsRequireClientCert, _ := cmd.Flags().GetBool("tls-require-client-cert")

			// Tracing flags
			tracingEnable, _ := cmd.Flags().GetBool("tracing-enable")
			tracingEndpoint, _ := cmd.Flags().GetString("tracing-endpoint")
			tracingSampleRate, _ := cmd.Flags().GetFloat64("tracing-sample-rate")
			tracingInsecure, _ := cmd.Flags().GetBool("tracing-insecure")
			tracingServiceName, _ := cmd.Flags().GetString("tracing-service-name")
			tracingTimeout, _ := cmd.Flags().GetDuration("tracing-timeout")
			tracingBatchSize, _ := cmd.Flags().GetInt("tracing-batch-size")

			log.Info().
				Int("grpc_port", grpcPort).
				Int("http_port", httpPort).
				Str("version", version).
				Msg("Starting Hybrid-Grid Coordinator")

			// Initialize tracing if enabled
			if tracingEnable {
				tracingCfg := config.TracingToLibConfig(config.TracingConfig{
					Enable:      tracingEnable,
					Endpoint:    tracingEndpoint,
					ServiceName: tracingServiceName,
					SampleRate:  tracingSampleRate,
					Insecure:    tracingInsecure,
					Timeout:     tracingTimeout,
					BatchSize:   tracingBatchSize,
				})

				tp, err := tracing.Init(ctx, tracingCfg)
				if err != nil {
					return fmt.Errorf("failed to initialize tracing: %w", err)
				}
				if tp != nil {
					defer func() {
						shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						if err := tp.Shutdown(shutdownCtx); err != nil {
							log.Warn().Err(err).Msg("Failed to shutdown tracer provider")
						}
					}()
					log.Info().
						Str("endpoint", tracingEndpoint).
						Str("service", tracingServiceName).
						Float64("sample_rate", tracingSampleRate).
						Msg("Tracing enabled")
				}
			}

			cfg := coordserver.DefaultConfig()
			cfg.Port = grpcPort
			cfg.AuthToken = token
			cfg.HeartbeatTTL = 60 * time.Second
			cfg.RequestTimeout = 120 * time.Second
			cfg.EnableRequestID = true
			cfg.SchedulerType = schedulerType
			cfg.TaskLogPath = taskLogPath
			cfg.EpsilonValue = epsilonValue
			cfg.AlphaValue = alphaValue
			cfg.Tracing.Enable = tracingEnable
			cfg.Tracing.Endpoint = tracingEndpoint
			cfg.Tracing.ServiceName = tracingServiceName
			cfg.Tracing.SampleRate = tracingSampleRate
			cfg.Tracing.Insecure = tracingInsecure
			cfg.Tracing.Timeout = tracingTimeout
			cfg.Tracing.BatchSize = tracingBatchSize

			// Configure TLS from CLI flags
			anyTLSFlags := tlsCert != "" || tlsKey != "" || tlsCA != "" || tlsRequireClientCert
			cfg.TLS.CertFile = tlsCert
			cfg.TLS.KeyFile = tlsKey
			cfg.TLS.ClientCA = tlsCA
			cfg.TLS.RequireClientCert = tlsRequireClientCert
			cfg.TLS.Enabled = anyTLSFlags

			// Validate TLS configuration if any TLS flags were provided
			if cfg.TLS.Enabled {
				if err := cfg.TLS.Validate(); err != nil {
					return err
				}

				log.Info().
					Str("cert", tlsCert).
					Bool("mtls", tlsRequireClientCert).
					Msg("TLS enabled")
			} else {
				log.Debug().Msg("TLS disabled")
			}

			// Initialize Prometheus metrics
			_ = observabilitymetrics.Default()
			log.Info().Msg("Prometheus metrics initialized")

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
	serveCmd.Flags().String("token", "", "Authentication token")
	serveCmd.Flags().Bool("no-mdns", false, "Disable mDNS advertisement")
	serveCmd.Flags().String("scheduler", "leastloaded", "Scheduler type: leastloaded, simple, p2c, epsilon-greedy, linucb")
	serveCmd.Flags().String("task-log", "", "Path to per-task JSON Lines log file (default: stdout)")
	serveCmd.Flags().Float64("epsilon", 0.1, "Exploration rate for epsilon-greedy scheduler (in [0, 1]; ignored otherwise)")
	serveCmd.Flags().Float64("alpha", 1.0, "LinUCB exploration coefficient α (in [0, 10]; ignored for other schedulers)")
	serveCmd.Flags().String("tls-cert", "", "Path to TLS certificate file (PEM format)")
	serveCmd.Flags().String("tls-key", "", "Path to TLS private key file (PEM format)")
	serveCmd.Flags().String("tls-ca", "", "Path to CA certificate for client verification (mTLS)")
	serveCmd.Flags().Bool("tls-require-client-cert", false, "Require client certificates (mTLS)")
	serveCmd.Flags().Bool("tracing-enable", false, "Enable OpenTelemetry tracing")
	serveCmd.Flags().String("tracing-endpoint", "localhost:4317", "OTLP gRPC endpoint")
	serveCmd.Flags().Float64("tracing-sample-rate", 0.1, "Tracing sample rate (0.0-1.0)")
	serveCmd.Flags().Bool("tracing-insecure", true, "Use insecure connection for tracing")
	serveCmd.Flags().String("tracing-service-name", "hybridgrid-coordinator", "Service name in traces")
	serveCmd.Flags().Duration("tracing-timeout", 10*time.Second, "Timeout for OTLP exports")
	serveCmd.Flags().Int("tracing-batch-size", 512, "Max spans to batch before export")

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
