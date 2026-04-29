package server

import (
	"encoding/json"
	"io"
	"os"
	"sync"
	"time"
)

// TaskLogger emits one JSON Lines record per completed task.
//
// Records are intended for offline analysis and machine-learning training:
// the schema is stable and matches the one consumed by the experimentation
// notebooks under docs/thesis/. Writes are serialized through a mutex so
// the logger is safe for concurrent use from gRPC handler goroutines.
//
// The logger swallows write errors on purpose. The Compile() handler is on
// the request hot path; logging must never abort a build.
type TaskLogger struct {
	mu     sync.Mutex
	w      io.Writer
	closer io.Closer
}

// TaskLogRecord captures the per-task feature set used for scheduler
// evaluation and bandit/RL training. Field names use snake_case so pandas
// can ingest the file with pd.read_json(path, lines=True) without renaming.
type TaskLogRecord struct {
	TS                          time.Time `json:"ts"`
	Event                       string    `json:"event"`
	TaskID                      string    `json:"task_id"`
	BuildType                   string    `json:"build_type"`
	Scheduler                   string    `json:"scheduler"`
	WorkerID                    string    `json:"worker_id"`
	WorkerArch                  string    `json:"worker_arch"`
	WorkerNativeArch            string    `json:"worker_native_arch"`
	WorkerCPUCores              int32     `json:"worker_cpu_cores"`
	WorkerMemBytes              int64     `json:"worker_mem_bytes"`
	WorkerActiveTasksAtDispatch int32     `json:"worker_active_tasks_at_dispatch"`
	WorkerMaxParallel           int32     `json:"worker_max_parallel"`
	WorkerDiscoverySource       string    `json:"worker_discovery_source"`
	TargetArch                  string    `json:"target_arch"`
	ClientOS                    string    `json:"client_os"`
	SourceSizeBytes             int       `json:"source_size_bytes"`
	PreprocessedSizeBytes       int       `json:"preprocessed_size_bytes"`
	RawSourceSizeBytes          int       `json:"raw_source_size_bytes"`
	QueueTimeMs                 int64     `json:"queue_time_ms"`
	CompileTimeMs               int64     `json:"compile_time_ms"`
	WorkerRPCLatencyMs          int64     `json:"worker_rpc_latency_ms"`
	TotalDurationMs             int64     `json:"total_duration_ms"`
	Success                     bool      `json:"success"`
	ExitCode                    int32     `json:"exit_code"`
	FromCache                   bool      `json:"from_cache"`

	// Learner introspection (populated only by LearningScheduler
	// implementations; zero/false otherwise).
	QValueAtDispatch float64 `json:"q_value_at_dispatch"`
	WasExploration   bool    `json:"was_exploration"`
}

// NewTaskLogger opens a JSON Lines log file. An empty path or "stdout"
// directs output to os.Stdout.
func NewTaskLogger(path string) (*TaskLogger, error) {
	if path == "" || path == "stdout" {
		return &TaskLogger{w: os.Stdout}, nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return &TaskLogger{w: f, closer: f}, nil
}

// Log writes a single record. Safe for concurrent use. A nil receiver is a
// no-op so callers may use the zero value to disable logging.
func (l *TaskLogger) Log(r *TaskLogRecord) {
	if l == nil || l.w == nil || r == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	_ = json.NewEncoder(l.w).Encode(r)
}

// Close releases the underlying file when one was opened. Idempotent.
func (l *TaskLogger) Close() error {
	if l == nil || l.closer == nil {
		return nil
	}
	return l.closer.Close()
}
