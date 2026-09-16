// Package telemetry owns the coordinator's observation plane: the
// aggregate state (tasks, builds, recent events) and the gRPC
// TelemetryService that exposes it to out-of-process consumers such
// as the standalone dashboard binary.
//
// The types here are the single definition of the dashboard's wire
// shapes; internal/observability/dashboard aliases them so the
// browser-facing REST API and the gRPC telemetry contract cannot
// drift apart.
package telemetry

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

// WorkerInfo represents worker information as the coordinator sees it.
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

// TaskInfo represents task information. ExitCode is omitempty in the
// JSON encoding — success serializes without it — so consumers must
// treat absent or zero as success.
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

// BuildInfo represents an aggregate logical build. All counts
// describe the RETAINED task set, which may be a subset of the
// build's real tasks: capped per build (maxTasksPerBuild) and/or
// shrunk when total-task eviction removes the oldest rows. Truncated
// marks both cases; totals are not the build's lifetime counts.
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

// StatsProvider provides live cluster statistics.
type StatsProvider interface {
	GetStats() *Stats
	GetWorkers() []*WorkerInfo
}

// ConsoleProvider is an optional StatsProvider extension exposing
// retained per-task console output.
type ConsoleProvider interface {
	GetConsole(taskID string) (stdout, stderr string, truncated bool, ok bool)
}
