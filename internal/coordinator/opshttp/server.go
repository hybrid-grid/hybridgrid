// Package opshttp serves the coordinator's small HTTP ops surface:
// /health (liveness), /metrics (Prometheus), and /log-level (runtime
// log-level control, token-gated). It replaces the health/metrics/
// log-level mux that the embedded dashboard server used to register.
//
// The full dashboard UI, REST API, and WebSocket feed now live in the
// standalone hg-dashboard binary, which drives them off the
// TelemetryService gRPC API.
package opshttp

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
)

// Server is the coordinator's lightweight HTTP ops server.
type Server struct {
	Port      int
	AuthToken string

	server *http.Server
}

// Start registers the ops routes and serves them on s.Port. It blocks
// until Stop is called or the listener fails; callers should run it in
// its own goroutine and report any non-http.ErrServerClosed error.
func (s *Server) Start() error {
	mux := http.NewServeMux()

	// Health endpoint: 200 "ok" (mirrors the ops health contract).
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Prometheus metrics endpoint.
	mux.Handle("/metrics", promhttp.Handler())

	// Log level endpoint (gated by the coordinator auth token).
	mux.Handle("/log-level", logging.NewLogLevelHandler(s.AuthToken))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.Port),
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	log.Info().Int("port", s.Port).Msg("Ops HTTP server starting")
	return s.server.ListenAndServe()
}

// Stop gracefully shuts the server down with a bounded deadline.
func (s *Server) Stop() error {
	if s.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(ctx)
}
