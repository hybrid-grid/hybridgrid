package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/netip"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	"github.com/h3nr1-d14z/hybridgrid/internal/discovery/mdns"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
	workerserver "github.com/h3nr1-d14z/hybridgrid/internal/worker/server"
)

var version = "v0.0.0-dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	cfg := config.DefaultConfig()

	// Setup logger
	logger, logCloser, err := logging.SetupLogger(cfg.Log)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup logger")
	}
	log.Logger = logger
	defer logCloser.Close()

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
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			coordinator, _ := cmd.Flags().GetString("coordinator")
			port, _ := cmd.Flags().GetInt("port")
			httpPort, _ := cmd.Flags().GetInt("http-port")
			token, _ := cmd.Flags().GetString("token")
			maxParallel, _ := cmd.Flags().GetInt("max-parallel")
			discoveryTimeout, _ := cmd.Flags().GetDuration("discovery-timeout")
			advertiseAddr, _ := cmd.Flags().GetString("advertise-address")
			tlsCert, _ := cmd.Flags().GetString("tls-cert")
			tlsKey, _ := cmd.Flags().GetString("tls-key")
			tlsCA, _ := cmd.Flags().GetString("tls-ca")
			tlsRequireClientCert, _ := cmd.Flags().GetBool("tls-require-client-cert")
			tracingEnable, _ := cmd.Flags().GetBool("tracing-enable")
			tracingEndpoint, _ := cmd.Flags().GetString("tracing-endpoint")
			tracingSampleRate, _ := cmd.Flags().GetFloat64("tracing-sample-rate")
			tracingInsecure, _ := cmd.Flags().GetBool("tracing-insecure")
			tracingServiceName, _ := cmd.Flags().GetString("tracing-service-name")
			tracingTimeout, _ := cmd.Flags().GetDuration("tracing-timeout")
			tracingBatchSize, _ := cmd.Flags().GetInt("tracing-batch-size")

			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid configuration: worker.port must be 1-65535, got %d", port)
			}
			if httpPort < 1 || httpPort > 65535 {
				return fmt.Errorf("invalid configuration: worker.http_port must be 1-65535, got %d", httpPort)
			}
			if port == httpPort {
				return fmt.Errorf("invalid configuration: worker.port and worker.http_port must be different, got %d for both", port)
			}

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
				}

				log.Info().
					Str("endpoint", tracingEndpoint).
					Str("service", tracingServiceName).
					Float64("sample_rate", tracingSampleRate).
					Msg("Tracing enabled")
			}

			// Create and start worker gRPC server
			cfg := workerserver.DefaultConfig()
			cfg.Port = port
			cfg.MaxConcurrent = maxParallel
			cfg.EnableRequestID = true
			cfg.Tracing.Enable = tracingEnable
			cfg.Tracing.Endpoint = tracingEndpoint
			cfg.Tracing.ServiceName = tracingServiceName
			cfg.Tracing.SampleRate = tracingSampleRate
			cfg.Tracing.Insecure = tracingInsecure
			cfg.Tracing.Timeout = tracingTimeout
			cfg.Tracing.BatchSize = tracingBatchSize

			if tlsCert != "" && tlsKey != "" {
				cfg.TLS.Enabled = true
				cfg.TLS.CertFile = tlsCert
				cfg.TLS.KeyFile = tlsKey
				cfg.TLS.ClientCA = tlsCA
				cfg.TLS.RequireClientCert = tlsRequireClientCert

				log.Info().
					Str("cert", tlsCert).
					Bool("mtls", tlsRequireClientCert).
					Msg("TLS enabled")
			} else {
				log.Debug().Msg("TLS disabled")
			}

			srv := workerserver.New(cfg)
			caps := srv.Capabilities()
			caps.WorkerId = fmt.Sprintf("worker-%s", hostname)
			caps.MaxParallelTasks = int32(maxParallel)
			caps.Version = version

			// Connect to coordinator
			cli, err := client.New(client.Config{
				Address:       coordinator,
				AuthToken:     token,
				Timeout:       30 * time.Second,
				Insecure:      !cfg.TLS.Enabled,
				EnableTracing: tracingEnable,
				TLS:           cfg.TLS,
			})
			if err != nil {
				return fmt.Errorf("failed to connect to coordinator: %w", err)
			}
			defer cli.Close()

			// Determine worker address to advertise
			workerAddr := advertiseAddr
			if workerAddr == "" {
				// Auto-detect outbound IP by checking which interface routes to coordinator
				if ip := getOutboundIP(coordinator); ip != "" {
					workerAddr = fmt.Sprintf("%s:%d", ip, port)
					log.Info().Str("detected_ip", ip).Msg("Auto-detected outbound IP for advertisement")
				} else {
					workerAddr = fmt.Sprintf("%s:%d", hostname, port)
					log.Warn().Str("fallback", workerAddr).Msg("Could not detect IP, using hostname")
				}
			}

			// Register with coordinator
			regReq := &pb.HandshakeRequest{
				Capabilities:  caps,
				AuthToken:     token,
				WorkerAddress: workerAddr,
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
			defer signal.Stop(sigCh)

			stopHeartbeatCh := make(chan struct{})
			var stopHeartbeatOnce sync.Once
			stopHeartbeat := func() {
				stopHeartbeatOnce.Do(func() {
					close(stopHeartbeatCh)
				})
			}

			errCh := make(chan error, 2)
			go func() {
				if err := srv.Start(); err != nil {
					errCh <- fmt.Errorf("gRPC server: %w", err)
				}
			}()

			// Start metrics HTTP server
			metricsMux := http.NewServeMux()
			metricsMux.Handle("/metrics", promhttp.Handler())
			metricsMux.Handle("/log-level", logging.NewLogLevelHandler())
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

			// Start heartbeat loop using Handshake to update registry
			heartbeatInterval := time.Duration(resp.HeartbeatIntervalSeconds) * time.Second
			go func() {
				ticker := time.NewTicker(heartbeatInterval)
				defer ticker.Stop()

				for {
					select {
					case <-ticker.C:
						// Re-send handshake to update heartbeat in coordinator registry
						hResp, err := cli.Handshake(context.Background(), regReq)
						if err != nil {
							log.Warn().Err(err).Msg("Heartbeat failed")
						} else {
							log.Debug().Bool("accepted", hResp.Accepted).Msg("Heartbeat sent")
						}
					case <-stopHeartbeatCh:
						return
					}
				}
			}()

			select {
			case sig := <-sigCh:
				stopHeartbeat()
				log.Info().Str("signal", sig.String()).Msg("Received shutdown signal")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()
				metricsServer.Shutdown(ctx)
				srv.Stop()
				return nil
			case err := <-errCh:
				stopHeartbeat()
				return fmt.Errorf("server error: %w", err)
			}
		},
	}

	serveCmd.Flags().Int("port", 50052, "Worker gRPC port")
	serveCmd.Flags().Int("http-port", 9090, "Worker HTTP/metrics port")
	serveCmd.Flags().String("coordinator", "", "Coordinator address (empty for mDNS auto-discovery)")
	serveCmd.Flags().String("advertise-address", "", "Address to advertise to coordinator (default: hostname:port)")
	serveCmd.Flags().String("token", "", "Authentication token")
	serveCmd.Flags().Int("max-parallel", 0, "Max parallel tasks (0 = auto)")
	serveCmd.Flags().Duration("discovery-timeout", 10*time.Second, "mDNS discovery timeout")
	serveCmd.Flags().String("tls-cert", "", "Path to TLS certificate file (PEM format)")
	serveCmd.Flags().String("tls-key", "", "Path to TLS private key file (PEM format)")
	serveCmd.Flags().String("tls-ca", "", "Path to CA certificate for client verification (mTLS)")
	serveCmd.Flags().Bool("tls-require-client-cert", false, "Require client certificates (mTLS)")
	serveCmd.Flags().Bool("tracing-enable", false, "Enable OpenTelemetry tracing")
	serveCmd.Flags().String("tracing-endpoint", "localhost:4317", "OTLP gRPC endpoint")
	serveCmd.Flags().Float64("tracing-sample-rate", 0.1, "Tracing sample rate (0.0-1.0)")
	serveCmd.Flags().Bool("tracing-insecure", true, "Use insecure connection for tracing")
	serveCmd.Flags().String("tracing-service-name", "hybridgrid-worker", "Service name in traces")
	serveCmd.Flags().Duration("tracing-timeout", 10*time.Second, "Timeout for OTLP exports")
	serveCmd.Flags().Int("tracing-batch-size", 512, "Max spans to batch before export")

	rootCmd.AddCommand(versionCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// getOutboundIP returns the preferred outbound IP for reaching the target address.
// This finds which local IP would be used to connect to the coordinator.
func getOutboundIP(target string) string {
	host := target
	if parsedHost, _, err := net.SplitHostPort(target); err == nil {
		host = parsedHost
	}

	ip, err := netip.ParseAddr(host)
	if err != nil {
		return ""
	}

	remote := net.UDPAddrFromAddrPort(netip.AddrPortFrom(ip, 80))
	conn, err := net.DialUDP("udp", nil, remote)
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}
