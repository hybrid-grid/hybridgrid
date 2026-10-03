package ui

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// handleStats returns cluster statistics. The envelope is the Stats
// object itself, byte-identical to the embedded dashboard's /stats.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.client.GetStats(r.Context(), &pb.GetTelemetryStatsRequest{AuthToken: s.config.AuthToken})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(statsFromProto(resp.GetStats()))
}

// handleWorkers returns the registered workers with the count +
// timestamp wrapper the SPA expects.
func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.client.ListWorkers(r.Context(), &pb.ListTelemetryWorkersRequest{AuthToken: s.config.AuthToken})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	var workers []*WorkerInfo
	for _, w := range resp.GetWorkers() {
		workers = append(workers, workerFromProto(w))
	}
	encodeList(w, "workers", workers)
}

// handleEvents returns recent task events stored on the WS hub as the
// SPA consumes them: the translated {"type","timestamp","data"} frames
// the dashboard already broadcasts over the socket.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	events := s.hub.GetRecentEvents()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"events":    events,
		"count":     len(events),
		"timestamp": time.Now().Unix(),
	})
}

// handleTasks returns recent task information, newest first, with an
// optional ?limit=N bound. The coordinator-side store already returns
// newest-first and applies the same truncation semantics the embedded
// dashboard enforced locally.
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var limit int32
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	resp, err := s.client.ListTasks(r.Context(), &pb.ListTelemetryTasksRequest{
		AuthToken: s.config.AuthToken,
		Limit:     limit,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	var tasks []*TaskInfo
	for _, t := range resp.GetTasks() {
		tasks = append(tasks, taskFromProto(t))
	}
	encodeList(w, "tasks", tasks)
}

// handleBuilds returns logical build aggregates, newest first.
func (s *Server) handleBuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	resp, err := s.client.ListBuilds(r.Context(), &pb.ListTelemetryBuildsRequest{AuthToken: s.config.AuthToken})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	var builds []*BuildInfo
	for _, b := range resp.GetBuilds() {
		builds = append(builds, buildFromProto(b))
	}
	encodeList(w, "builds", builds)
}

// handleTaskConsole returns retained console output for one task. The
// coordinator's GetConsole can report Unavailable when the provider
// lacks console retention, or NotFound for an unknown task — both map
// to the same HTTP statuses the embedded dashboard exposed.
func (s *Server) handleTaskConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	taskID := r.PathValue("id")
	resp, err := s.client.GetConsole(r.Context(), &pb.GetTelemetryConsoleRequest{
		AuthToken: s.config.AuthToken,
		TaskId:    taskID,
	})
	if err != nil {
		respondGRPCError(w, err)
		return
	}
	console := resp.GetConsole()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":   taskID,
		"stdout":    console.GetStdout(),
		"stderr":    console.GetStderr(),
		"truncated": console.GetTruncated(),
	})
}

// handleBuildByID returns a logical build and its retained tasks. The
// contract pins the 404 body byte-for-byte as
// {"error":"build not found"}: the embedded dashboard served exactly
// that string regardless of the requested id, and the SPA keys off it,
// so the gRPC layer's id-containing status text must NOT leak through.
// Any non-NotFound error falls back to respondGRPCError's generic 5xx
// envelope.
func (s *Server) handleBuildByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	buildID := r.PathValue("id")
	resp, err := s.client.GetBuild(r.Context(), &pb.GetTelemetryBuildRequest{
		AuthToken: s.config.AuthToken,
		BuildId:   buildID,
	})
	if err != nil {
		if statusFromGRPC(err) == codes.NotFound {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "build not found"})
			return
		}
		respondGRPCError(w, err)
		return
	}
	build := buildFromProto(resp.GetBuild())
	var tasks []*TaskInfo
	for _, t := range resp.GetTasks() {
		tasks = append(tasks, taskFromProto(t))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"build": build, "tasks": tasks})
}

// encodeList writes the "{key: list, count: N, timestamp: T}" envelope
// the SPA polls for workers/tasks/builds, regardless of whether the
// slice header stayed nil (the embedded dashboard always encoded an
// explicit non-nil list in that slot).
func encodeList(w http.ResponseWriter, key string, items interface{}) {
	count := 0
	switch v := items.(type) {
	case []*WorkerInfo:
		count = len(v)
	case []*TaskInfo:
		count = len(v)
	case []*BuildInfo:
		count = len(v)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		key:         items,
		"count":     count,
		"timestamp": time.Now().Unix(),
	})
}

// respondGRPCError maps gRPC error codes back onto the only HTTP
// tristates the embedded dashboard exposed: 404 for NotFound, 503 for
// Unavailable/Unimplemented, and a generic 500 (with the gRPC message
// body) for anything else. Build-not-found and console-not-available
// are exactly the two endpoints that deliberately distinguish these.
func respondGRPCError(w http.ResponseWriter, err error) {
	switch statusFromGRPC(err) {
	case codes.NotFound:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": errBody(err)})
	case codes.Unavailable, codes.Unimplemented:
		// Console-not-available mirrors the embedded dashboard's
		// 503 short-text body.
		http.Error(w, errBody(err), http.StatusServiceUnavailable)
		return
	case codes.Unauthenticated:
		http.Error(w, errBody(err), http.StatusUnauthorized)
		return
	default:
		http.Error(w, errBody(err), http.StatusInternalServerError)
	}
}

// statusFromGRPC is status.FromError's nil-safe form, kept local so
// tests can hit the default branch deterministically.
func statusFromGRPC(err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if st, ok := status.FromError(err); ok {
		return st.Code()
	}
	return codes.Unknown
}

// errBody writes a short human text body from a gRPC error, preferring
// the status message and falling back to the raw error.
func errBody(err error) string {
	if st, ok := status.FromError(err); ok && st.Message() != "" {
		return st.Message()
	}
	return err.Error()
}
