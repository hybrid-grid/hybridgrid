// Package ui serves the hybrid-grid dashboard's browser-facing surface
// (SPA assets, REST API and WebSocket) from a gRPC TelemetryService
// client instead of an in-process provider. It is the standalone
// hg-dashboard binary's only HTTP concern; the coordinator keeps the
// gRPC build plane and its small ops HTTP server.
//
// Wire shapes are byte-identical to the API the embedded dashboard
// always exposed: the same JSON tag layouts live in
// internal/telemetry so the gRPC contract and the browser API cannot
// drift apart. The ui package reuses them as aliases and converts
// proto responses back into those types before encoding.
package ui

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

// The SPA's built output. Run `npm run build` inside web/ (or `make
// build-ui`) before `go build`/`go:embed` picks up web/dist — it is
// gitignored generated output, not committed source.
//
//go:embed all:web/dist
var assetsFS embed.FS

// The dashboard's wire types live in internal/telemetry so the gRPC
// telemetry contract and the browser-facing REST API cannot drift
// apart. Re-aliasing them here keeps every conversion readable and
// gives the JSON encoder the exact struct tags the SPA depends on.
type (
	Stats      = telemetry.Stats
	WorkerInfo = telemetry.WorkerInfo
	TaskInfo   = telemetry.TaskInfo
	BuildInfo  = telemetry.BuildInfo
)

// Client is the subset of the generated TelemetryServiceClient the
// ui server calls. Declared as an interface so tests can substitute a
// fake while the production path takes the real generated client via
// pb.NewTelemetryServiceClient without an adapter (the generated
// client widens cleanly through this method set).
type Client interface {
	GetStats(ctx context.Context, in *pb.GetTelemetryStatsRequest, opts ...grpc.CallOption) (*pb.GetTelemetryStatsResponse, error)
	ListWorkers(ctx context.Context, in *pb.ListTelemetryWorkersRequest, opts ...grpc.CallOption) (*pb.ListTelemetryWorkersResponse, error)
	ListTasks(ctx context.Context, in *pb.ListTelemetryTasksRequest, opts ...grpc.CallOption) (*pb.ListTelemetryTasksResponse, error)
	ListBuilds(ctx context.Context, in *pb.ListTelemetryBuildsRequest, opts ...grpc.CallOption) (*pb.ListTelemetryBuildsResponse, error)
	GetBuild(ctx context.Context, in *pb.GetTelemetryBuildRequest, opts ...grpc.CallOption) (*pb.GetTelemetryBuildResponse, error)
	GetConsole(ctx context.Context, in *pb.GetTelemetryConsoleRequest, opts ...grpc.CallOption) (*pb.GetTelemetryConsoleResponse, error)
	StreamEvents(ctx context.Context, in *pb.StreamTelemetryEventsRequest, opts ...grpc.CallOption) (grpc.ServerStreamingClient[pb.TelemetryEvent], error)
}

// Config holds the standalone dashboard server's configuration.
type Config struct {
	Port            int
	AuthToken       string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

// DefaultConfig returns sensible defaults matching the embedded
// dashboard's long-standing values.
func DefaultConfig() Config {
	return Config{
		Port:            8080,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
		ShutdownTimeout: 10 * time.Second,
	}
}

// Server is the browser-facing HTTP server. It serves SPA assets,
// the REST API and the WebSocket feed, backed entirely by a gRPC
// TelemetryService client to the coordinator.
type Server struct {
	config   Config
	client   Client
	hub      *Hub
	server   *http.Server
	upgrader websocket.Upgrader

	// streamCtx / streamCancel govern the long-lived StreamEvents
	// pump that feeds the WebSocket hub. They are captured so Stop
	// can tear the pump down with the HTTP listener.
	streamCtx    context.Context
	streamCancel context.CancelFunc
}

// New wires a standalone dashboard server against the given
// telemetry client. The client is expected to live at least as long
// as the server; the caller (cmd/hg-dashboard) owns closing it.
func New(cfg Config, client Client) *Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // The dashboard is a same-LAN admin surface.
		},
	}
	s := &Server{
		config:   cfg,
		client:   client,
		hub:      NewHub(),
		upgrader: upgrader,
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      s.routes(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
	return s
}

// routes wires every HTTP concern the SPA, REST API and WebSocket
// need. It mirrors the path table the embedded dashboard exposed so
// existing clients keep working unchanged against the new binary.
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// Health: cheap liveness probe, byte-identical to the embedded
	// dashboard (200 "OK\n"). The coordinator keeps /metrics and
	// /log-level on its own ops port; this binary is dashboard-only.
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK\n"))
	})

	mux.HandleFunc("/api/v1/stats", s.handleStats)
	mux.HandleFunc("/api/v1/workers", s.handleWorkers)
	mux.HandleFunc("/api/v1/events", s.handleEvents)
	mux.HandleFunc("/api/v1/tasks", s.handleTasks)
	mux.HandleFunc("/api/v1/builds", s.handleBuilds)
	mux.HandleFunc("GET /api/v1/builds/{id}", s.handleBuildByID)
	mux.HandleFunc("/api/v1/builds/{id}", s.handleBuildByID)
	mux.HandleFunc("/api/v1/tasks/{id}/console", s.handleTaskConsole)

	mux.HandleFunc("/ws", s.handleWebSocket)

	assetsContent, _ := fs.Sub(assetsFS, "web/dist")
	mux.Handle("/", spaHandler(assetsContent))

	return mux
}

// spaHandler serves the built SPA, falling back to index.html for any
// path that isn't a real file — the TanStack Router routes (/builds,
// /builds/{id}, /workers, ...) are client-side only, so a hard
// refresh or direct link on one of them has no matching file on disk.
func spaHandler(content fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(content))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(content, strings.TrimPrefix(r.URL.Path, "/")); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}

// Start brings up the WebSocket hub, opens the gRPC event stream that
// feeds it, and serves HTTP. It blocks until Stop is called or the
// listener errors. Callers typically run it in its own goroutine.
func (s *Server) Start() error {
	s.streamCtx, s.streamCancel = context.WithCancel(context.Background())

	go s.hub.Run()
	go s.pumpStream(s.streamCtx)

	log.Info().Int("port", s.config.Port).Msg("hg-dashboard server starting")
	return s.server.ListenAndServe()
}

// pumpStream keeps the gRPC StreamEvents subscription alive: it dials
// it, translates each TelemetryEvent into a WS message and broadcasts
// it. Replay on (re)connect happens because the server-side Telemetry
// service replays recent task events at the start of every stream.
// On any transport error the next iteration re-subscribes after a
// short back-off.
func (s *Server) pumpStream(ctx context.Context) {
	const backoff = 2 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		// runStream only returns on error (including ctx cancel via
		// Recv), so the nil check would be vacuous; suppress the
		// reconnect warning during intentional shutdown instead.
		if err := s.runStream(ctx); ctx.Err() == nil {
			log.Warn().Err(err).Msg("telemetry stream ended; reconnecting")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
	}
}

// runStream reads telemetry events off one StreamEvents RPC until the
// client cancels the context or the server closes the stream.
func (s *Server) runStream(ctx context.Context) error {
	stream, err := s.client.StreamEvents(ctx, &pb.StreamTelemetryEventsRequest{
		AuthToken: s.config.AuthToken,
	})
	if err != nil {
		return err
	}
	for {
		ev, err := stream.Recv()
		if err != nil {
			return err
		}
		s.hub.TranslateAndBroadcast(ev)
	}
}

// Stop tears the dashboard down: cancel the gRPC stream pump, stop the
// WebSocket hub, and gracefully shut the HTTP listener.
func (s *Server) Stop() error {
	if s.streamCancel != nil {
		s.streamCancel()
	}
	s.hub.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), s.config.ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Hub returns the WebSocket hub, primarily so tests can drive the
// broadcast path directly.
func (s *Server) Hub() *Hub { return s.hub }
