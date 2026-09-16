package dashboard

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/h3nr1-d14z/hybridgrid/internal/telemetry"
)

// The dashboard's wire types live in internal/telemetry so the gRPC
// telemetry contract (consumed by the standalone dashboard binary
// and any third-party UI) and the browser-facing REST API cannot
// drift apart.
type (
	Stats      = telemetry.Stats
	WorkerInfo = telemetry.WorkerInfo
	TaskInfo   = telemetry.TaskInfo
	BuildInfo  = telemetry.BuildInfo
)

// handleStats returns cluster statistics.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var stats *Stats
	if s.provider != nil {
		stats = s.provider.GetStats()
	} else {
		stats = &Stats{
			Timestamp: time.Now().Unix(),
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// handleWorkers returns worker list.
func (s *Server) handleWorkers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var workers []*WorkerInfo
	if s.provider != nil {
		workers = s.provider.GetWorkers()
	} else {
		workers = []*WorkerInfo{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"workers":   workers,
		"count":     len(workers),
		"timestamp": time.Now().Unix(),
	})
}

// handleEvents returns recent events.
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	events := s.hub.GetRecentEvents()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"events":    events,
		"count":     len(events),
		"timestamp": time.Now().Unix(),
	})
}

// handleTasks returns recent task information. An optional ?limit=N bounds
// the response to the newest N tasks — the cluster activity pane polls a
// bounded window rather than the full retained set (up to 20k tasks).
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := s.hub.GetTasks()
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if limit, err := strconv.Atoi(raw); err == nil && limit > 0 && len(tasks) > limit {
			tasks = tasks[:limit]
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"tasks":     tasks,
		"count":     len(tasks),
		"timestamp": time.Now().Unix(),
	})
}

// handleBuilds returns logical build aggregates.
func (s *Server) handleBuilds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	builds := s.hub.GetBuilds()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"builds":    builds,
		"count":     len(builds),
		"timestamp": time.Now().Unix(),
	})
}

// handleTaskConsole returns retained console output for one task.
// Console data lives in the coordinator (responses are unary, so
// output arrives complete at task completion); the dashboard reaches it
// through the optional ConsoleProvider. Without such a provider the
// endpoint reports 503 rather than pretending no output exists.
func (s *Server) handleTaskConsole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	provider, ok := s.provider.(ConsoleProvider)
	if !ok {
		http.Error(w, "console output not available", http.StatusServiceUnavailable)
		return
	}
	taskID := r.PathValue("id")
	stdout, stderr, truncated, found := provider.GetConsole(taskID)
	if !found {
		http.Error(w, "task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"task_id":   taskID,
		"stdout":    stdout,
		"stderr":    stderr,
		"truncated": truncated,
	})
}

// handleBuildByID returns a logical build and its tasks.
func (s *Server) handleBuildByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	build, tasks, ok := s.hub.GetBuildDetail(r.PathValue("id"))
	w.Header().Set("Content-Type", "application/json")
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "build not found"})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"build": build, "tasks": tasks})
}
