package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// Stats represents cluster statistics.
type Stats struct {
	TotalTasks          int64   `json:"total_tasks"`
	SuccessTasks        int64   `json:"success_tasks"`
	FailedTasks         int64   `json:"failed_tasks"`
	ActiveTasks         int64   `json:"active_tasks"`
	QueuedTasks         int64   `json:"queued_tasks"`
	CacheHits           int64   `json:"cache_hits"`
	CacheMisses         int64   `json:"cache_misses"`
	CacheHitRate        float64 `json:"cache_hit_rate"`
	FlutterBuilds       int64   `json:"flutter_builds"`
	FlutterCacheHits    int64   `json:"flutter_cache_hits"`
	FlutterCacheMisses  int64   `json:"flutter_cache_misses"`
	FlutterCacheHitRate float64 `json:"flutter_cache_hit_rate"`
	UnityBuilds         int64   `json:"unity_builds"`
	UnityCacheHits      int64   `json:"unity_cache_hits"`
	UnityCacheMisses    int64   `json:"unity_cache_misses"`
	UnityCacheHitRate   float64 `json:"unity_cache_hit_rate"`
	TotalWorkers        int     `json:"total_workers"`
	HealthyWorkers      int     `json:"healthy_workers"`
	UptimeSeconds       int64   `json:"uptime_seconds"`
	Timestamp           int64   `json:"timestamp"`
}

// WorkerInfo represents worker information for the dashboard.
type WorkerInfo struct {
	ID                string   `json:"id"`
	Host              string   `json:"host"`
	Address           string   `json:"address"`
	OS                string   `json:"os"`
	Architecture      string   `json:"architecture"`
	Architectures     []string `json:"architectures"`
	CPUCores          int32    `json:"cpu_cores"`
	MemoryGB          float64  `json:"memory_gb"`
	MaxParallelTasks  int32    `json:"max_parallel_tasks"`
	ActiveTasks       int32    `json:"active_tasks"`
	TotalTasks        int64    `json:"total_tasks"`
	SuccessRate       float64  `json:"success_rate"`
	AvgLatencyMs      float64  `json:"avg_latency_ms"`
	CircuitState      string   `json:"circuit_state"`
	DiscoverySource   string   `json:"discovery_source"`
	Version           string   `json:"version"`
	DockerAvailable   bool     `json:"docker_available"`
	FlutterAvailable  bool     `json:"flutter_available"`
	FlutterSDKVersion string   `json:"flutter_sdk_version"`
	FlutterPlatforms  []string `json:"flutter_platforms"`
	UnityAvailable    bool     `json:"unity_available"`
	UnityVersions     []string `json:"unity_versions"`
	UnityPlatforms    []string `json:"unity_platforms"`
	Compilers         []string `json:"compilers"`
	BuildTypes        []string `json:"build_types"`
	Healthy           bool     `json:"healthy"`
	LastSeen          int64    `json:"last_seen"`
}

// TaskInfo represents task information for the dashboard.
type TaskInfo struct {
	ID            string `json:"id"`
	BuildType     string `json:"build_type"`
	BuildID       string `json:"build_id"`
	Status        string `json:"status"`
	WorkerID      string `json:"worker_id"`
	StartedAtMs   int64  `json:"started_at_ms"`
	CompletedAtMs int64  `json:"completed_at_ms,omitempty"`
	DurationMs    int64  `json:"duration_ms,omitempty"`
	QueueMs       int64  `json:"queue_ms"`
	CompileMs     int64  `json:"compile_ms"`
	ExitCode      int32  `json:"exit_code,omitempty"`
	FromCache     bool   `json:"from_cache"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// BuildInfo represents an aggregate logical build for the dashboard.
type BuildInfo struct {
	ID             string `json:"id"`
	BuildType      string `json:"build_type"`
	Status         string `json:"status"`
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	FailedTasks    int    `json:"failed_tasks"`
	RunningTasks   int    `json:"running_tasks"`
	FromCacheCount int    `json:"from_cache_count"`
	FirstTaskAtMs  int64  `json:"first_task_at_ms"`
	LastTaskAtMs   int64  `json:"last_task_at_ms"`
	Truncated      bool   `json:"truncated"`
}

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

// handleTasks returns recent task information.
func (s *Server) handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tasks := s.hub.GetTasks()

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
