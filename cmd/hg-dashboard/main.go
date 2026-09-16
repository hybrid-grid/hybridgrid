// Command hg-dashboard serves the hybrid-grid dashboard's browser-facing
// surface (SPA, REST API and WebSocket) from a gRPC TelemetryService
// client to the coordinator. It is the standalone replacement for the
// coordinator's embedded dashboard, so the UI can be hosted apart from
// (or rewritten independently of) the coordinator.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/config"
	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
	"github.com/h3nr1-d14z/hybridgrid/internal/observability/ui"
	hgtls "github.com/h3nr1-d14z/hybridgrid/internal/security/tls"
)

// version is set at build time via -ldflags "-X main.version=…", the
// same convention as cmd/hg-coord. The dev default is intentional so a
// bare `go build ./cmd/hg-dashboard` is obvious in logs.
var version = "v0.0.0-dev"

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	// Console writer to stderr mirrors the coordinator startup; the
	// dashboard binary is a long-lived daemon the operator watches
	// interactively, so structured console coloring stays enabled.
	logCfg := config.LogConfig{
		Level:  envOr("LOG_LEVEL", "info"),
		Format: "console",
	}
	logger, logCloser, err := logging.SetupLogger(logCfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to setup logger")
	}
	log.Logger = logger
	defer logCloser.Close()

	rootCmd := &cobra.Command{
		Use:   "hg-dashboard",
		Short: "Hybrid-Grid Build Dashboard",
		Long: `hg-dashboard is the standalone dashboard for the Hybrid-Grid Build system.
It serves the SPA assets, the REST API and the WebSocket feed from the
coordinator's gRPC TelemetryService, so the UI can be hosted apart from
(or rewritten independently of) the coordinator binary.

Quick start:
  hg-dashboard serve                               Start on :8080 against localhost:9000
  hg-dashboard serve --port 9090 --insecure        Plaintext gRPC + :9090 HTTP
  hg-dashboard version                             Print version`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("hg-dashboard %s\n", version)
		},
	}

	var (
		port           int
		coordinator    string
		coordinatorTok string
		insecureDial   bool
		tlsCert        string
		tlsKey         string
		tlsCA          string
	)

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the dashboard server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if port < 1 || port > 65535 {
				return fmt.Errorf("invalid --port %d; must be 1-65535", port)
			}

			log.Info().
				Int("port", port).
				Str("coordinator", coordinator).
				Str("version", version).
				Msg("Starting Hybrid-Grid Dashboard")

			conn, err := dialCoordinator(coordinator, insecureDial, tlsCert, tlsKey, tlsCA)
			if err != nil {
				return fmt.Errorf("connect coordinator: %w", err)
			}
			defer func() { _ = conn.Close() }()

			cfg := ui.DefaultConfig()
			cfg.Port = port
			cfg.AuthToken = coordinatorTok

			srv := ui.New(cfg, pb.NewTelemetryServiceClient(conn))

			// Graceful shutdown: SIGINT/SIGTERM trip the HTTP
			// listener and the gRPC event pump together. The
			// ui Server's Stop blocks until both are down (or the
			// shutdown timeout elapses), so there is nothing else
			// to coordinate here.
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			errCh := make(chan error, 1)
			go func() {
				if err := srv.Start(); err != nil {
					errCh <- err
				}
			}()

			select {
			case <-ctx.Done():
				log.Info().Msg("Received shutdown signal")
				if err := srv.Stop(); err != nil {
					log.Warn().Err(err).Msg("Dashboard shutdown error")
				}
				return nil
			case err := <-errCh:
				return fmt.Errorf("dashboard server: %w", err)
			}
		},
	}

	serveCmd.Flags().IntVar(&port, "port", 8080, "HTTP port for the dashboard server")
	serveCmd.Flags().StringVar(&coordinator, "coordinator", "localhost:9000", "coordinator gRPC TelemetryService address")
	serveCmd.Flags().StringVar(&coordinatorTok, "coordinator-token", "", "auth token forwarded to the coordinator as auth_token")
	serveCmd.Flags().BoolVar(&insecureDial, "insecure", false, "connect to the coordinator with plaintext gRPC (no TLS)")
	serveCmd.Flags().StringVar(&tlsCert, "tls-cert", "", "path to TLS client certificate (PEM format)")
	serveCmd.Flags().StringVar(&tlsKey, "tls-key", "", "path to TLS client private key (PEM format)")
	serveCmd.Flags().StringVar(&tlsCA, "tls-ca", "", "path to CA certificate for coordinator verification")

	rootCmd.AddCommand(versionCmd, serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// dialCoordinator builds a gRPC client connection to the coordinator's
// TelemetryService endpoint. The trust model mirrors cmd/hgbuild:
// plaintext when --insecure, system TLS (with optional client cert +
// CA bundle) otherwise. The default stays secure (TLS) since the
// dashboard typically runs on a separate host from the coordinator.
func dialCoordinator(address string, insecureDial bool, tlsCert, tlsKey, tlsCA string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}

	if insecureDial {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		tlsCfg := hgtls.Config{
			Enabled:  true,
			CertFile: tlsCert,
			KeyFile:  tlsKey,
			ClientCA: tlsCA,
		}
		creds, err := hgtls.ClientCredentials(tlsCfg)
		if err != nil {
			return nil, fmt.Errorf("load TLS credentials: %w", err)
		}
		if creds != nil {
			opts = append(opts, grpc.WithTransportCredentials(creds))
		}
		log.Info().
			Str("coordinator", address).
			Bool("tls_client_cert", tlsCert != "").
			Bool("ca_bundle", tlsCA != "").
			Msg("Dialing coordinator over TLS")
	}

	conn, err := grpc.NewClient(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	return conn, nil
}

// envOr returns the named environment variable or fallback when unset.
// LOG_LEVEL is honoured here so the daemon picks up the same convention
// the rest of the codebase recognises without dragging in viper.
func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}
