package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const namespace = "hybridgrid"

// Metrics contains all Prometheus metrics for Hybrid-Grid.
type Metrics struct {
	// Counters
	TasksTotal     *prometheus.CounterVec
	CacheHits      prometheus.Counter
	CacheMisses    prometheus.Counter
	FallbacksTotal *prometheus.CounterVec

	// Gauges
	WorkersTotal *prometheus.GaugeVec
	ActiveTasks  *prometheus.GaugeVec
	QueueDepth   prometheus.Gauge

	// Histograms
	TaskDuration    *prometheus.HistogramVec
	QueueTime       *prometheus.HistogramVec
	TransferBytes   *prometheus.HistogramVec
	WorkerLatencyMs *prometheus.HistogramVec

	// Circuit breaker states
	CircuitState *prometheus.GaugeVec
}

var (
	defaultMetrics *Metrics
	once           sync.Once
)

// Default returns the singleton metrics instance.
func Default() *Metrics {
	once.Do(func() {
		defaultMetrics = New()
		defaultMetrics.Register(prometheus.DefaultRegisterer)
	})
	return defaultMetrics
}

// New creates a new Metrics instance.
func New() *Metrics {
	return &Metrics{
		// Counters
		TasksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "tasks_total",
				Help:      "Total number of build tasks processed",
			},
			[]string{"status", "build_type", "worker"},
		),
		CacheHits: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_hits_total",
				Help:      "Total number of cache hits",
			},
		),
		CacheMisses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "cache_misses_total",
				Help:      "Total number of cache misses",
			},
		),
		FallbacksTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: namespace,
				Name:      "fallbacks_total",
				Help:      "Total number of local fallback compilations",
			},
			[]string{"reason"},
		),

		// Gauges
		WorkersTotal: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "workers_total",
				Help:      "Current number of workers by state and discovery source",
			},
			[]string{"state", "source"},
		),
		ActiveTasks: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "active_tasks",
				Help:      "Number of tasks currently being processed per worker",
			},
			[]string{"worker"},
		),
		QueueDepth: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "queue_depth",
				Help:      "Number of tasks waiting in queue",
			},
		),

		// Histograms
		TaskDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "task_duration_seconds",
				Help:      "Duration of task execution in seconds",
				Buckets:   []float64{.1, .5, 1, 5, 10, 30, 60, 120, 300},
			},
			[]string{"build_type", "status"},
		),
		QueueTime: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "queue_time_seconds",
				Help:      "Time spent waiting in queue before processing",
				Buckets:   []float64{.01, .05, .1, .25, .5, 1, 2.5, 5, 10},
			},
			[]string{"build_type"},
		),
		TransferBytes: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "network_transfer_bytes",
				Help:      "Size of data transferred to/from workers",
				Buckets:   prometheus.ExponentialBuckets(1024, 4, 10), // 1KB to ~1GB
			},
			[]string{"direction"}, // "upload" or "download"
		),
		WorkerLatencyMs: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: namespace,
				Name:      "worker_latency_ms",
				Help:      "gRPC round-trip latency to workers in milliseconds",
				Buckets:   []float64{1, 5, 10, 25, 50, 100, 250, 500, 1000},
			},
			[]string{"worker"},
		),

		// Circuit breaker
		CircuitState: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "circuit_state",
				Help:      "Circuit breaker state (0=closed, 1=half-open, 2=open)",
			},
			[]string{"worker"},
		),
	}
}

// Register registers all metrics with the given registerer.
func (m *Metrics) Register(reg prometheus.Registerer) {
	reg.MustRegister(
		m.TasksTotal,
		m.CacheHits,
		m.CacheMisses,
		m.FallbacksTotal,
		m.WorkersTotal,
		m.ActiveTasks,
		m.QueueDepth,
		m.TaskDuration,
		m.QueueTime,
		m.TransferBytes,
		m.WorkerLatencyMs,
		m.CircuitState,
	)
}

// Handler returns an HTTP handler for the /metrics endpoint.
func Handler() http.Handler {
	return promhttp.Handler()
}

// TaskStatus represents the outcome of a task.
type TaskStatus string

const (
	TaskStatusSuccess TaskStatus = "success"
	TaskStatusError   TaskStatus = "error"
	TaskStatusTimeout TaskStatus = "timeout"
)

// RecordTaskComplete records a completed task with its outcome.
func (m *Metrics) RecordTaskComplete(status TaskStatus, buildType, workerID string, durationSec float64) {
	m.TasksTotal.WithLabelValues(string(status), buildType, workerID).Inc()
	m.TaskDuration.WithLabelValues(buildType, string(status)).Observe(durationSec)
}

// RecordCacheHit records a cache hit.
func (m *Metrics) RecordCacheHit() {
	m.CacheHits.Inc()
}

// RecordCacheMiss records a cache miss.
func (m *Metrics) RecordCacheMiss() {
	m.CacheMisses.Inc()
}

// RecordFallback records a local fallback with reason.
func (m *Metrics) RecordFallback(reason string) {
	m.FallbacksTotal.WithLabelValues(reason).Inc()
}

// SetWorkerCount updates the worker count gauge.
func (m *Metrics) SetWorkerCount(state, source string, count float64) {
	m.WorkersTotal.WithLabelValues(state, source).Set(count)
}

// SetActiveTaskCount updates the active task count for a worker.
func (m *Metrics) SetActiveTaskCount(workerID string, count float64) {
	m.ActiveTasks.WithLabelValues(workerID).Set(count)
}

// SetQueueDepth updates the queue depth gauge.
func (m *Metrics) SetQueueDepth(depth float64) {
	m.QueueDepth.Set(depth)
}

// RecordQueueTime records time spent in queue.
func (m *Metrics) RecordQueueTime(buildType string, durationSec float64) {
	m.QueueTime.WithLabelValues(buildType).Observe(durationSec)
}

// RecordTransfer records bytes transferred.
func (m *Metrics) RecordTransfer(direction string, bytes float64) {
	m.TransferBytes.WithLabelValues(direction).Observe(bytes)
}

// RecordWorkerLatency records worker latency in milliseconds.
func (m *Metrics) RecordWorkerLatency(workerID string, latencyMs float64) {
	m.WorkerLatencyMs.WithLabelValues(workerID).Observe(latencyMs)
}

// CircuitStateValue represents circuit breaker states as numeric values.
type CircuitStateValue float64

const (
	CircuitStateClosed   CircuitStateValue = 0
	CircuitStateHalfOpen CircuitStateValue = 1
	CircuitStateOpen     CircuitStateValue = 2
)

// SetCircuitState updates the circuit breaker state for a worker.
func (m *Metrics) SetCircuitState(workerID string, state CircuitStateValue) {
	m.CircuitState.WithLabelValues(workerID).Set(float64(state))
}

// RemoveWorkerMetrics removes all metrics associated with a worker.
func (m *Metrics) RemoveWorkerMetrics(workerID string) {
	m.ActiveTasks.DeleteLabelValues(workerID)
	m.WorkerLatencyMs.DeleteLabelValues(workerID)
	m.CircuitState.DeleteLabelValues(workerID)
}
