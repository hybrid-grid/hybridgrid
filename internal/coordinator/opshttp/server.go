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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"

	"github.com/h3nr1-d14z/hybridgrid/internal/logging"
	"github.com/h3nr1-d14z/hybridgrid/internal/security/auth"
)

// EventSink accepts validated experiment events.
type EventSink interface {
	LogInjectedEvent(kind, target string, dispatchCount int64, detail string) error
}

// DispatchCounter reports booked C/C++ dispatch attempts.
type DispatchCounter interface {
	Dispatches() int64
}

// Server is the coordinator's lightweight HTTP ops server.
type Server struct {
	Port            int
	AuthToken       string
	EventSink       EventSink
	DispatchCounter DispatchCounter

	server *http.Server
}

// Start registers the ops routes and serves them on s.Port. It blocks
// until Stop is called or the listener fails; callers should run it in
// its own goroutine and report any non-http.ErrServerClosed error.
func (s *Server) Start() error {
	s.server = &http.Server{
		Addr: fmt.Sprintf(":%d", s.Port), Handler: s.Handler(),
		ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second,
	}
	log.Info().Int("port", s.Port).Msg("Ops HTTP server starting")
	return s.server.ListenAndServe()
}

// Handler exposes the ops routes, including the experiment endpoints.
func (s *Server) Handler() http.Handler {
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
	mux.HandleFunc("/dispatch-count", s.handleDispatchCount)
	mux.HandleFunc("/events", s.handleEvents)
	return mux
}

func (s *Server) handleDispatchCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.DispatchCounter == nil {
		http.Error(w, "dispatch counter unavailable", http.StatusServiceUnavailable)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(struct {
		Dispatches int64 `json:"dispatches"`
	}{s.DispatchCounter.Dispatches()})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if s.AuthToken != "" {
		provided := strings.TrimSpace(r.Header.Get("Authorization"))
		provided = strings.TrimPrefix(provided, "Bearer ")
		if !auth.ValidateToken(provided, s.AuthToken) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	if s.EventSink == nil {
		http.Error(w, "task logger unavailable", http.StatusServiceUnavailable)
		return
	}
	var event struct {
		Kind          string `json:"kind"`
		Target        string `json:"target"`
		DispatchCount int64  `json:"dispatch_count"`
		Detail        string `json:"detail"`
	}
	decoder := json.NewDecoder(io.LimitReader(r.Body, 4097))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		http.Error(w, "invalid event", http.StatusBadRequest)
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF || decoder.InputOffset() > 4096 {
		http.Error(w, "invalid event", http.StatusBadRequest)
		return
	}
	if event.Kind != "drift_on" && event.Kind != "drift_off" && event.Kind != "drift_note" ||
		event.Target == "" || len(event.Target) > 128 || len(event.Detail) > 512 || event.DispatchCount < 0 {
		http.Error(w, "invalid event", http.StatusBadRequest)
		return
	}
	if err := s.EventSink.LogInjectedEvent(event.Kind, event.Target, event.DispatchCount, event.Detail); err != nil {
		http.Error(w, "task logger unavailable", http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
