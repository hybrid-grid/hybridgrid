package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

// Stats represents cluster statistics.
type Stats struct {
	TotalTasks     int64   `json:"total_tasks"`
	SuccessTasks   int64   `json:"success_tasks"`
	FailedTasks    int64   `json:"failed_tasks"`
	ActiveTasks    int64   `json:"active_tasks"`
	QueuedTasks    int64   `json:"queued_tasks"`
	CacheHits      int64   `json:"cache_hits"`
	CacheMisses    int64   `json:"cache_misses"`
	CacheHitRate   float64 `json:"cache_hit_rate"`
	TotalWorkers   int     `json:"total_workers"`
	HealthyWorkers int     `json:"healthy_workers"`
	UptimeSeconds  int64   `json:"uptime_seconds"`
	Timestamp      int64   `json:"timestamp"`
}

// WorkerInfo represents worker information for the dashboard.
type WorkerInfo struct {
	ID              string  `json:"id"`
	Host            string  `json:"host"`
	Address         string  `json:"address"`
	Architecture    string  `json:"architecture"`
	CPUCores        int32   `json:"cpu_cores"`
	MemoryGB        float64 `json:"memory_gb"`
	ActiveTasks     int32   `json:"active_tasks"`
	TotalTasks      int64   `json:"total_tasks"`
	SuccessRate     float64 `json:"success_rate"`
	AvgLatencyMs    float64 `json:"avg_latency_ms"`
	CircuitState    string  `json:"circuit_state"`
	DiscoverySource string  `json:"discovery_source"`
	Healthy         bool    `json:"healthy"`
	LastSeen        int64   `json:"last_seen"`
}

// TaskInfo represents task information for the dashboard.
type TaskInfo struct {
	ID           string `json:"id"`
	BuildType    string `json:"build_type"`
	Status       string `json:"status"`
	WorkerID     string `json:"worker_id"`
	StartedAt    int64  `json:"started_at"`
	CompletedAt  int64  `json:"completed_at,omitempty"`
	DurationMs   int64  `json:"duration_ms,omitempty"`
	ExitCode     int32  `json:"exit_code,omitempty"`
	FromCache    bool   `json:"from_cache"`
	ErrorMessage string `json:"error_message,omitempty"`
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
