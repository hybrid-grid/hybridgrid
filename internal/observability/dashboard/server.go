package dashboard

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

//go:embed assets/*
var assetsFS embed.FS

// Config holds dashboard server configuration.
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// StatsProvider provides statistics for the dashboard.
type StatsProvider interface {
	GetStats() *Stats
	GetWorkers() []*WorkerInfo
}

// Server is the HTTP dashboard server.
type Server struct {
	config   Config
	server   *http.Server
	hub      *Hub
	provider StatsProvider
}

// New creates a new dashboard server.
func New(cfg Config, provider StatsProvider) *Server {
	s := &Server{
		config:   cfg,
		hub:      NewHub(),
		provider: provider,
	}

	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// API endpoints
	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/workers", s.handleWorkers)
	mux.HandleFunc("/api/v1/events", s.handleEvents)

	// WebSocket endpoint
	mux.HandleFunc("/ws", s.handleWebSocket)

	// Static assets and dashboard
	assetsContent, _ := fs.Sub(assetsFS, "assets")
	mux.Handle("/", http.FileServer(http.FS(assetsContent)))

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return s
}

// Start starts the HTTP server and WebSocket hub.
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.hub.Run()

	// Start periodic stats broadcast
	go s.broadcastLoop()

	log.Info().Int("port", s.config.Port).Msg("Dashboard server starting")
	return s.server.ListenAndServe()
}

// Stop gracefully stops the server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()

	s.hub.Stop()
	return s.server.Shutdown(ctx)
}

// Hub returns the WebSocket hub for broadcasting events.
func (s *Server) Hub() *Hub {
	return s.hub
}

// EventNotifierFunc allows wiring up task events from the coordinator.
// To use: in main.go, create a wrapper that converts coordinator.TaskEvent to dashboard.TaskInfo.
type EventNotifierFunc struct {
	OnTaskStarted   func(task *TaskInfo)
	OnTaskCompleted func(task *TaskInfo)
}

// CreateEventNotifier creates event notifier callbacks for the coordinator.
func (s *Server) CreateEventNotifier() (onStart func(id, buildType, status, workerID string, startedAt int64), onComplete func(id, buildType, status, workerID string, startedAt, completedAt, durationMs int64, exitCode int32, errorMsg string)) {
	onStart = func(id, buildType, status, workerID string, startedAt int64) {
		s.hub.BroadcastTaskStarted(&TaskInfo{
			ID:        id,
			BuildType: buildType,
			Status:    status,
			WorkerID:  workerID,
			StartedAt: startedAt,
		})
	}
	onComplete = func(id, buildType, status, workerID string, startedAt, completedAt, durationMs int64, exitCode int32, errorMsg string) {
		s.hub.BroadcastTaskCompleted(&TaskInfo{
			ID:           id,
			BuildType:    buildType,
			Status:       status,
			WorkerID:     workerID,
			StartedAt:    startedAt,
			CompletedAt:  completedAt,
			DurationMs:   durationMs,
			ExitCode:     exitCode,
			ErrorMessage: errorMsg,
		})
	}
	return
}

// broadcastLoop periodically broadcasts stats to WebSocket clients.
func (s *Server) broadcastLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if s.provider != nil {
				stats := s.provider.GetStats()
				s.hub.BroadcastStats(stats)
			}
		case <-s.hub.done:
			return
		}
	}
}
