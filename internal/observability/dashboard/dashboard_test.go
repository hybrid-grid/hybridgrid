package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

// mockProvider implements StatsProvider for testing.
type mockProvider struct {
	stats   *Stats
	workers []*WorkerInfo
}

func (m *mockProvider) GetStats() *Stats {
	if m.stats != nil {
		return m.stats
	}
	return &Stats{
		TotalTasks:     100,
		SuccessTasks:   90,
		FailedTasks:    10,
		ActiveTasks:    5,
		CacheHits:      80,
		CacheMisses:    20,
		TotalWorkers:   3,
		HealthyWorkers: 2,
		Timestamp:      time.Now().Unix(),
	}
}

func (m *mockProvider) GetWorkers() []*WorkerInfo {
	if m.workers != nil {
		return m.workers
	}
	return []*WorkerInfo{
		{
			ID:           "worker-1",
			Host:         "host1.local",
			Architecture: "x86_64",
			CPUCores:     8,
			MemoryGB:     16,
			ActiveTasks:  2,
			Healthy:      true,
		},
		{
			ID:           "worker-2",
			Host:         "host2.local",
			Architecture: "arm64",
			CPUCores:     4,
			MemoryGB:     8,
			ActiveTasks:  1,
			Healthy:      true,
		},
	}
}

func TestServer_New(t *testing.T) {
	cfg := DefaultConfig()
	provider := &mockProvider{}
	s := New(cfg, provider)

	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.hub == nil {
		t.Error("Hub is nil")
	}
}

func TestServer_HandleStats(t *testing.T) {
	cfg := DefaultConfig()
	provider := &mockProvider{}
	s := New(cfg, provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()

	s.handleStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	var stats Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.TotalTasks != 100 {
		t.Errorf("TotalTasks = %d, want 100", stats.TotalTasks)
	}
	if stats.HealthyWorkers != 2 {
		t.Errorf("HealthyWorkers = %d, want 2", stats.HealthyWorkers)
	}
}

func TestServer_HandleStats_MethodNotAllowed(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()

	s.handleStats(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", rec.Code)
	}
}

func TestServer_HandleWorkers(t *testing.T) {
	cfg := DefaultConfig()
	provider := &mockProvider{}
	s := New(cfg, provider)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	rec := httptest.NewRecorder()

	s.handleWorkers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	workers, ok := response["workers"].([]interface{})
	if !ok {
		t.Fatal("Response missing workers array")
	}
	if len(workers) != 2 {
		t.Errorf("Workers count = %d, want 2", len(workers))
	}

	firstWorker, ok := workers[0].(map[string]interface{})
	if !ok {
		t.Fatal("first worker should decode as object")
	}

	if _, ok := firstWorker["architectures"]; !ok {
		t.Fatal("worker response missing architectures field")
	}
	if _, ok := firstWorker["compilers"]; !ok {
		t.Fatal("worker response missing compilers field")
	}
	if _, ok := firstWorker["build_types"]; !ok {
		t.Fatal("worker response missing build_types field")
	}
	if _, ok := firstWorker["docker_available"]; !ok {
		t.Fatal("worker response missing docker_available field")
	}
}

func TestServer_HandleWorkers_NilProvider(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers", nil)
	rec := httptest.NewRecorder()

	s.handleWorkers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}
}

func TestHub_NewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}
	if hub.clients == nil {
		t.Error("clients map is nil")
	}
	if hub.broadcast == nil {
		t.Error("broadcast channel is nil")
	}
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	if hub.ClientCount() != 0 {
		t.Error("Initial client count should be 0")
	}
}

func TestHub_Broadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Give hub time to start
	time.Sleep(10 * time.Millisecond)

	// Broadcast should not panic with no clients
	hub.BroadcastStats(&Stats{TotalTasks: 10})
	hub.BroadcastWorkerAdded(&WorkerInfo{ID: "test"})
	hub.BroadcastWorkerRemoved("test")
	hub.BroadcastTaskStarted(&TaskInfo{ID: "task-1"})
	hub.BroadcastTaskCompleted(&TaskInfo{ID: "task-1"})
}

func TestMessage_Types(t *testing.T) {
	tests := []struct {
		msgType MessageType
		want    string
	}{
		{MessageTypeStats, "stats"},
		{MessageTypeWorkerAdded, "worker_added"},
		{MessageTypeWorkerRemove, "worker_removed"},
		{MessageTypeTaskStarted, "task_started"},
		{MessageTypeTaskComplete, "task_completed"},
	}

	for _, tt := range tests {
		if string(tt.msgType) != tt.want {
			t.Errorf("MessageType %v = %s, want %s", tt.msgType, tt.msgType, tt.want)
		}
	}
}

func TestServer_StaticAssets(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	// Test that index.html is served
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Hybrid-Grid") {
		t.Error("Response should contain 'Hybrid-Grid'")
	}
	if !strings.Contains(body, "chart.umd.js") {
		t.Error("Response should reference the vendored Chart.js bundle")
	}
}

func TestServer_ChartAsset(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	// The vendored Chart.js bundle must be served — the dashboard is
	// offline-capable and must not fall back to a CDN.
	req := httptest.NewRequest(http.MethodGet, "/chart.umd.js", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("chart.umd.js status = %d, want 200", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Chart") {
		t.Error("chart.umd.js body should contain Chart.js code")
	}
	// The bundle must be loadable as a classic script. Chart.js ships
	// both a UMD and an ESM build; the ESM one (top-level export{...})
	// throws at parse time under a plain <script defer> tag and never
	// defines window.Chart — a regression that leaves the dashboard
	// chartless without any network error.
	if strings.Contains(body, "export{") {
		t.Error("chart.umd.js looks like the ESM build (top-level export); serve the UMD bundle instead")
	}
	if !strings.Contains(body, "!function(") {
		t.Error("chart.umd.js should start with the UMD IIFE wrapper")
	}
	if n := rec.Body.Len(); n < 100_000 {
		t.Errorf("chart.umd.js len = %d, want at least 100000 (suspiciously small bundle)", n)
	}
}

func TestServer_MetricsEndpoint(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()

	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	// Prometheus metrics should be present
	body := rec.Body.String()
	if !strings.Contains(body, "go_") {
		t.Error("Response should contain Go metrics")
	}
}

func TestServer_WebSocketUpgrade(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	// Start hub
	go s.hub.Run()
	defer s.hub.Stop()

	// Create test server
	ts := httptest.NewServer(s.server.Handler)
	defer ts.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// Connect WebSocket
	ws, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket dial failed: %v", err)
	}
	defer ws.Close()

	// Give time for connection to register
	time.Sleep(50 * time.Millisecond)

	if s.hub.ClientCount() != 1 {
		t.Errorf("Client count = %d, want 1", s.hub.ClientCount())
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}
	if cfg.ReadTimeout != 15*time.Second {
		t.Errorf("ReadTimeout = %v, want 15s", cfg.ReadTimeout)
	}
	if cfg.WriteTimeout != 15*time.Second {
		t.Errorf("WriteTimeout = %v, want 15s", cfg.WriteTimeout)
	}
}

func TestStats_JSON(t *testing.T) {
	stats := &Stats{
		TotalTasks:     100,
		SuccessTasks:   90,
		FailedTasks:    10,
		ActiveTasks:    5,
		CacheHits:      80,
		CacheMisses:    20,
		CacheHitRate:   0.8,
		TotalWorkers:   3,
		HealthyWorkers: 2,
		UptimeSeconds:  3600,
		Timestamp:      1234567890,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Failed to marshal stats: %v", err)
	}

	var decoded Stats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal stats: %v", err)
	}

	if decoded.TotalTasks != stats.TotalTasks {
		t.Errorf("TotalTasks = %d, want %d", decoded.TotalTasks, stats.TotalTasks)
	}
}

func TestWorkerInfo_JSON(t *testing.T) {
	worker := &WorkerInfo{
		ID:              "worker-1",
		Host:            "host.local",
		Address:         "192.168.1.1:50052",
		Architecture:    "x86_64",
		CPUCores:        8,
		MemoryGB:        16.0,
		ActiveTasks:     2,
		TotalTasks:      100,
		SuccessRate:     0.95,
		AvgLatencyMs:    25.5,
		CircuitState:    "CLOSED",
		DiscoverySource: "mdns",
		Healthy:         true,
		LastSeen:        1234567890,
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Failed to marshal worker: %v", err)
	}

	var decoded WorkerInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal worker: %v", err)
	}

	if decoded.ID != worker.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, worker.ID)
	}
	if decoded.MemoryGB != worker.MemoryGB {
		t.Errorf("MemoryGB = %f, want %f", decoded.MemoryGB, worker.MemoryGB)
	}
}

func TestTaskInfo_JSON(t *testing.T) {
	task := &TaskInfo{
		ID:            "task-1",
		BuildType:     "cpp",
		BuildID:       "build-1",
		Status:        "completed",
		WorkerID:      "worker-1",
		StartedAtMs:   1234567890,
		CompletedAtMs: 1234567895,
		DurationMs:    5000,
		ExitCode:      0,
		FromCache:     true,
		ErrorMessage:  "",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	var decoded TaskInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}

	if decoded.ID != task.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, task.ID)
	}
	if decoded.DurationMs != task.DurationMs {
		t.Errorf("DurationMs = %d, want %d", decoded.DurationMs, task.DurationMs)
	}
	if decoded.FromCache != task.FromCache {
		t.Errorf("FromCache = %v, want %v", decoded.FromCache, task.FromCache)
	}
	if decoded.BuildID != task.BuildID {
		t.Errorf("BuildID = %s, want %s", decoded.BuildID, task.BuildID)
	}
}

func TestTaskInfo_JSON_WithError(t *testing.T) {
	task := &TaskInfo{
		ID:            "task-2",
		BuildType:     "rust",
		Status:        "failed",
		WorkerID:      "worker-2",
		StartedAtMs:   1234567890,
		CompletedAtMs: 1234567900,
		DurationMs:    10000,
		ExitCode:      1,
		FromCache:     false,
		ErrorMessage:  "compilation error",
	}

	data, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Failed to marshal task: %v", err)
	}

	var decoded TaskInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal task: %v", err)
	}

	if decoded.ErrorMessage != "compilation error" {
		t.Errorf("ErrorMessage = %s, want 'compilation error'", decoded.ErrorMessage)
	}
}

func TestServer_HandleStats_NilProvider(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/stats", nil)
	rec := httptest.NewRecorder()

	s.handleStats(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	var stats Stats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Nil provider returns empty stats with timestamp
	if stats.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

func TestServer_HandleWorkers_MethodNotAllowed(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workers", nil)
	rec := httptest.NewRecorder()

	s.handleWorkers(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", rec.Code)
	}
}

func TestServer_Hub(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	hub := s.Hub()
	if hub == nil {
		t.Fatal("Hub() returned nil")
	}
	if hub != s.hub {
		t.Error("Hub() should return the server's hub")
	}
}

func TestHub_RegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create mock client
	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	// Register
	hub.register <- client
	time.Sleep(10 * time.Millisecond)
	if hub.ClientCount() != 1 {
		t.Errorf("ClientCount = %d, want 1", hub.ClientCount())
	}

	// Unregister
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)
	if hub.ClientCount() != 0 {
		t.Errorf("ClientCount = %d, want 0", hub.ClientCount())
	}
}

func TestHub_BroadcastWithClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Create mock client
	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	// Broadcast stats
	hub.BroadcastStats(&Stats{TotalTasks: 100})
	time.Sleep(10 * time.Millisecond)

	// Check client received message
	select {
	case msg := <-client.send:
		if !strings.Contains(string(msg), "stats") {
			t.Error("Message should contain 'stats'")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Client should have received message")
	}
}

func TestHub_BroadcastAllTypes(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	time.Sleep(10 * time.Millisecond)

	// These should not panic with no clients
	hub.BroadcastStats(&Stats{TotalTasks: 1})
	hub.BroadcastWorkerAdded(&WorkerInfo{ID: "w1"})
	hub.BroadcastWorkerRemoved("w1")
	hub.BroadcastTaskStarted(&TaskInfo{ID: "t1"})
	hub.BroadcastTaskCompleted(&TaskInfo{ID: "t1"})
}

func TestMessage_JSON(t *testing.T) {
	msg := &Message{
		Type:      MessageTypeStats,
		Timestamp: 1234567890,
		Data:      map[string]int{"count": 5},
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	var decoded Message
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if decoded.Type != MessageTypeStats {
		t.Errorf("Type = %s, want %s", decoded.Type, MessageTypeStats)
	}
}

func TestMessageType_AllTypes(t *testing.T) {
	types := []struct {
		msgType MessageType
		want    string
	}{
		{MessageTypeStats, "stats"},
		{MessageTypeWorkerAdded, "worker_added"},
		{MessageTypeWorkerRemove, "worker_removed"},
		{MessageTypeTaskStarted, "task_started"},
		{MessageTypeTaskComplete, "task_completed"},
		{MessageTypePing, "ping"},
		{MessageTypePong, "pong"},
	}

	for _, tt := range types {
		if string(tt.msgType) != tt.want {
			t.Errorf("MessageType %v = %s, want %s", tt.msgType, tt.msgType, tt.want)
		}
	}
}

func TestConfig_Custom(t *testing.T) {
	cfg := Config{
		Port:            9999,
		ReadTimeout:     30 * time.Second,
		WriteTimeout:    30 * time.Second,
		ShutdownTimeout: 20 * time.Second,
	}

	if cfg.Port != 9999 {
		t.Errorf("Port = %d, want 9999", cfg.Port)
	}
	if cfg.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want 30s", cfg.ReadTimeout)
	}
	if cfg.ShutdownTimeout != 20*time.Second {
		t.Errorf("ShutdownTimeout = %v, want 20s", cfg.ShutdownTimeout)
	}
}

func TestServer_NewWithCustomConfig(t *testing.T) {
	cfg := Config{
		Port:            9000,
		ReadTimeout:     20 * time.Second,
		WriteTimeout:    20 * time.Second,
		ShutdownTimeout: 15 * time.Second,
	}
	s := New(cfg, &mockProvider{})

	if s.config.Port != 9000 {
		t.Errorf("Port = %d, want 9000", s.config.Port)
	}
}

func TestMockProvider_CustomStats(t *testing.T) {
	provider := &mockProvider{
		stats: &Stats{
			TotalTasks:     200,
			SuccessTasks:   180,
			FailedTasks:    20,
			CacheHits:      150,
			CacheMisses:    50,
			CacheHitRate:   0.75,
			TotalWorkers:   5,
			HealthyWorkers: 4,
		},
	}

	stats := provider.GetStats()
	if stats.TotalTasks != 200 {
		t.Errorf("TotalTasks = %d, want 200", stats.TotalTasks)
	}
	if stats.CacheHitRate != 0.75 {
		t.Errorf("CacheHitRate = %f, want 0.75", stats.CacheHitRate)
	}
}

func TestMockProvider_CustomWorkers(t *testing.T) {
	provider := &mockProvider{
		workers: []*WorkerInfo{
			{ID: "w1", Host: "host1"},
			{ID: "w2", Host: "host2"},
			{ID: "w3", Host: "host3"},
		},
	}

	workers := provider.GetWorkers()
	if len(workers) != 3 {
		t.Errorf("Workers count = %d, want 3", len(workers))
	}
}

func TestServer_APIEndpoints(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	tests := []struct {
		path       string
		method     string
		wantStatus int
	}{
		{"/api/v1/stats", http.MethodGet, http.StatusOK},
		{"/api/v1/stats", http.MethodPost, http.StatusMethodNotAllowed},
		{"/api/v1/workers", http.MethodGet, http.StatusOK},
		{"/api/v1/workers", http.MethodDelete, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			s.server.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestStats_AllFields(t *testing.T) {
	stats := &Stats{
		TotalTasks:     100,
		SuccessTasks:   80,
		FailedTasks:    20,
		ActiveTasks:    5,
		QueuedTasks:    10,
		CacheHits:      60,
		CacheMisses:    40,
		CacheHitRate:   0.6,
		TotalWorkers:   4,
		HealthyWorkers: 3,
		UptimeSeconds:  7200,
		Timestamp:      1234567890,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	// Check all fields present in JSON
	str := string(data)
	fields := []string{
		"total_tasks", "success_tasks", "failed_tasks",
		"active_tasks", "queued_tasks", "cache_hits",
		"cache_misses", "cache_hit_rate", "total_workers",
		"healthy_workers", "uptime_seconds", "timestamp",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing field: %s", f)
		}
	}
}

func TestWorkerInfo_AllFields(t *testing.T) {
	worker := &WorkerInfo{
		ID:              "w1",
		Host:            "host",
		Address:         "addr",
		Architecture:    "x86_64",
		CPUCores:        4,
		MemoryGB:        8.0,
		ActiveTasks:     1,
		TotalTasks:      50,
		SuccessRate:     0.9,
		AvgLatencyMs:    15.0,
		CircuitState:    "CLOSED",
		DiscoverySource: "mdns",
		Healthy:         true,
		LastSeen:        123456,
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"id", "host", "address", "architecture",
		"cpu_cores", "memory_gb", "active_tasks",
		"total_tasks", "success_rate", "avg_latency_ms",
		"circuit_state", "discovery_source", "healthy", "last_seen",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing field: %s", f)
		}
	}
}

func TestStats_FlutterFields(t *testing.T) {
	stats := &Stats{
		TotalTasks:          100,
		SuccessTasks:        80,
		FailedTasks:         20,
		ActiveTasks:         5,
		QueuedTasks:         10,
		CacheHits:           60,
		CacheMisses:         40,
		CacheHitRate:        0.6,
		FlutterBuilds:       30,
		FlutterCacheHits:    20,
		FlutterCacheMisses:  10,
		FlutterCacheHitRate: 0.67,
		TotalWorkers:        4,
		HealthyWorkers:      3,
		UptimeSeconds:       7200,
		Timestamp:           1234567890,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"flutter_builds", "flutter_cache_hits",
		"flutter_cache_misses", "flutter_cache_hit_rate",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing Flutter field: %s", f)
		}
	}

	var decoded Stats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.FlutterBuilds != 30 {
		t.Errorf("FlutterBuilds = %d, want 30", decoded.FlutterBuilds)
	}
	if decoded.FlutterCacheHits != 20 {
		t.Errorf("FlutterCacheHits = %d, want 20", decoded.FlutterCacheHits)
	}
	if decoded.FlutterCacheMisses != 10 {
		t.Errorf("FlutterCacheMisses = %d, want 10", decoded.FlutterCacheMisses)
	}
	if decoded.FlutterCacheHitRate != 0.67 {
		t.Errorf("FlutterCacheHitRate = %f, want 0.67", decoded.FlutterCacheHitRate)
	}
}

func TestWorkerInfo_FlutterFields(t *testing.T) {
	worker := &WorkerInfo{
		ID:                "w1",
		Host:              "host",
		Address:           "addr",
		Architecture:      "arm64",
		CPUCores:          4,
		MemoryGB:          8.0,
		ActiveTasks:       1,
		TotalTasks:        50,
		SuccessRate:       0.9,
		AvgLatencyMs:      15.0,
		CircuitState:      "CLOSED",
		DiscoverySource:   "mdns",
		Healthy:           true,
		LastSeen:          123456,
		FlutterAvailable:  true,
		FlutterSDKVersion: "3.19.6",
		FlutterPlatforms:  []string{"TARGET_PLATFORM_ANDROID", "TARGET_PLATFORM_LINUX"},
		Compilers:         []string{"gcc", "g++"},
		BuildTypes:        []string{"BUILD_TYPE_CPP", "BUILD_TYPE_FLUTTER"},
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"flutter_available", "flutter_sdk_version", "flutter_platforms",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing Flutter field: %s", f)
		}
	}

	var decoded WorkerInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !decoded.FlutterAvailable {
		t.Error("FlutterAvailable = false, want true")
	}
	if decoded.FlutterSDKVersion != "3.19.6" {
		t.Errorf("FlutterSDKVersion = %s, want 3.19.6", decoded.FlutterSDKVersion)
	}
	if len(decoded.FlutterPlatforms) != 2 {
		t.Errorf("FlutterPlatforms len = %d, want 2", len(decoded.FlutterPlatforms))
	}
}

func TestServer_HandleTasks(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	go s.hub.Run()
	defer s.hub.Stop()

	s.hub.BroadcastTaskStarted(&TaskInfo{
		ID:          "task-flutter-1",
		BuildType:   "flutter",
		Status:      "running",
		WorkerID:    "worker-1",
		StartedAtMs: 1234567890,
	})
	s.hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "task-flutter-2",
		BuildType:     "flutter",
		Status:        "completed",
		WorkerID:      "worker-1",
		StartedAtMs:   1234567891,
		CompletedAtMs: 1234567900,
		DurationMs:    9000,
		ExitCode:      0,
		FromCache:     false,
	})

	time.Sleep(20 * time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()

	s.handleTasks(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Status = %d, want 200", rec.Code)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	tasks, ok := response["tasks"].([]interface{})
	if !ok {
		t.Fatal("Response missing tasks array")
	}
	if len(tasks) != 2 {
		t.Errorf("Tasks count = %d, want 2", len(tasks))
	}

	firstTask, ok := tasks[0].(map[string]interface{})
	if !ok {
		t.Fatal("first task should decode as object")
	}
	if firstTask["build_type"] != "flutter" {
		t.Errorf("build_type = %v, want flutter", firstTask["build_type"])
	}
}

func TestServer_HandleTasks_MethodNotAllowed(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()

	s.handleTasks(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", rec.Code)
	}
}

func TestHub_GetTasks(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	hub.BroadcastTaskStarted(&TaskInfo{
		ID:          "task-cpp-1",
		BuildType:   "cpp",
		Status:      "running",
		WorkerID:    "w1",
		StartedAtMs: 1000000,
	})
	hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "task-flutter-1",
		BuildType:     "flutter",
		Status:        "completed",
		WorkerID:      "w2",
		StartedAtMs:   1000001,
		CompletedAtMs: 1000010,
		DurationMs:    9000,
		ExitCode:      0,
		FromCache:     false,
	})
	hub.BroadcastStats(&Stats{TotalTasks: 10})

	time.Sleep(20 * time.Millisecond)

	tasks := hub.GetTasks()
	if len(tasks) != 2 {
		t.Errorf("GetTasks len = %d, want 2", len(tasks))
	}

	var hasFlutter bool
	for _, task := range tasks {
		if task.BuildType == "flutter" {
			hasFlutter = true
			if task.Status != "completed" {
				t.Errorf("Flutter task status = %s, want completed", task.Status)
			}
			if task.DurationMs != 9000 {
				t.Errorf("Flutter task DurationMs = %d, want 9000", task.DurationMs)
			}
		}
	}
	if !hasFlutter {
		t.Error("Expected at least one Flutter task")
	}
}

func TestHub_GetTasks_DeduplicatesLatestState(t *testing.T) {
	hub := NewHub()
	hub.BroadcastTaskStarted(&TaskInfo{
		ID:          "task-1",
		BuildID:     "build-1",
		BuildType:   "cpp",
		Status:      "running",
		StartedAtMs: 100,
	})
	hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "task-1",
		BuildID:       "build-1",
		BuildType:     "cpp",
		Status:        "completed",
		StartedAtMs:   100,
		CompletedAtMs: 125,
		DurationMs:    25,
	})

	tasks := hub.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("GetTasks len = %d, want 1", len(tasks))
	}
	if tasks[0].ID != "task-1" {
		t.Errorf("GetTasks ID = %q, want task-1", tasks[0].ID)
	}
	if tasks[0].Status != "completed" {
		t.Errorf("GetTasks Status = %q, want completed", tasks[0].Status)
	}
	if tasks[0].CompletedAtMs != 125 {
		t.Errorf("GetTasks CompletedAt = %d, want 125", tasks[0].CompletedAtMs)
	}
}

func TestHub_GetBuilds_Grouping(t *testing.T) {
	hub := NewHub()
	hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "old-completed",
		BuildID:       "build-old",
		BuildType:     "cpp",
		Status:        "completed",
		StartedAtMs:   10,
		CompletedAtMs: 20,
		FromCache:     true,
	})
	hub.BroadcastTaskStarted(&TaskInfo{
		ID:          "new-running",
		BuildID:     "build-new",
		BuildType:   "cpp",
		Status:      "running",
		StartedAtMs: 30,
	})
	hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "new-failed",
		BuildID:       "build-new",
		BuildType:     "cpp",
		Status:        "failed",
		StartedAtMs:   35,
		CompletedAtMs: 50,
	})
	hub.BroadcastTaskCompleted(&TaskInfo{
		ID:            "ungrouped",
		BuildType:     "cpp",
		Status:        "completed",
		StartedAtMs:   60,
		CompletedAtMs: 70,
	})

	builds := hub.GetBuilds()
	if len(builds) != 2 {
		t.Fatalf("GetBuilds len = %d, want 2", len(builds))
	}
	if builds[0].ID != "build-new" {
		t.Errorf("newest build ID = %q, want build-new", builds[0].ID)
	}
	if builds[0].Status != "running" {
		t.Errorf("build-new Status = %q, want running", builds[0].Status)
	}
	if builds[0].TotalTasks != 2 {
		t.Errorf("build-new TotalTasks = %d, want 2", builds[0].TotalTasks)
	}
	if builds[0].RunningTasks != 1 {
		t.Errorf("build-new RunningTasks = %d, want 1", builds[0].RunningTasks)
	}
	if builds[0].FailedTasks != 1 {
		t.Errorf("build-new FailedTasks = %d, want 1", builds[0].FailedTasks)
	}
	if builds[0].CompletedTasks != 0 {
		t.Errorf("build-new CompletedTasks = %d, want 0", builds[0].CompletedTasks)
	}
	if builds[0].FirstTaskAtMs != 30 {
		t.Errorf("build-new FirstTaskAt = %d, want 30", builds[0].FirstTaskAtMs)
	}
	if builds[0].LastTaskAtMs != 50 {
		t.Errorf("build-new LastTaskAt = %d, want 50", builds[0].LastTaskAtMs)
	}

	if builds[1].ID != "build-old" {
		t.Errorf("oldest build ID = %q, want build-old", builds[1].ID)
	}
	if builds[1].Status != "completed" {
		t.Errorf("build-old Status = %q, want completed", builds[1].Status)
	}
	if builds[1].TotalTasks != 1 {
		t.Errorf("build-old TotalTasks = %d, want 1", builds[1].TotalTasks)
	}
	if builds[1].CompletedTasks != 1 {
		t.Errorf("build-old CompletedTasks = %d, want 1", builds[1].CompletedTasks)
	}
	if builds[1].FromCacheCount != 1 {
		t.Errorf("build-old FromCacheCount = %d, want 1", builds[1].FromCacheCount)
	}
}

func TestHub_GetBuilds_EvictsOldestBuild(t *testing.T) {
	hub := NewHub()
	hub.maxBuilds = 1
	hub.BroadcastTaskStarted(&TaskInfo{ID: "old-task", BuildID: "old-build", Status: "running"})
	hub.BroadcastTaskStarted(&TaskInfo{ID: "new-task", BuildID: "new-build", Status: "running"})

	builds := hub.GetBuilds()
	if len(builds) != 1 {
		t.Fatalf("GetBuilds len = %d, want 1", len(builds))
	}
	if builds[0].ID != "new-build" {
		t.Errorf("GetBuilds ID = %q, want new-build", builds[0].ID)
	}
	tasks := hub.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("GetTasks len = %d, want 1", len(tasks))
	}
	if tasks[0].ID != "new-task" {
		t.Errorf("GetTasks ID = %q, want new-task", tasks[0].ID)
	}
}

func TestServer_HandleBuilds(t *testing.T) {
	s := New(DefaultConfig(), &mockProvider{})
	s.hub.BroadcastTaskStarted(&TaskInfo{
		ID:          "task-1",
		BuildID:     "build-1",
		BuildType:   "cpp",
		Status:      "running",
		StartedAtMs: 100,
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/builds", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, want 200", rec.Code)
	}
	var response struct {
		Builds    []*BuildInfo `json:"builds"`
		Count     int          `json:"count"`
		Timestamp int64        `json:"timestamp"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response.Count != 1 {
		t.Errorf("Count = %d, want 1", response.Count)
	}
	if len(response.Builds) != 1 {
		t.Fatalf("Builds len = %d, want 1", len(response.Builds))
	}
	if response.Builds[0].ID != "build-1" {
		t.Errorf("Build ID = %q, want build-1", response.Builds[0].ID)
	}
	if response.Timestamp == 0 {
		t.Error("Timestamp should be set")
	}
}

func TestServer_HandleBuilds_MethodNotAllowed(t *testing.T) {
	s := New(DefaultConfig(), &mockProvider{})
	req := httptest.NewRequest(http.MethodPost, "/api/v1/builds", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", rec.Code)
	}
}

func TestServer_CreateEventNotifier_ForwardsBuildID(t *testing.T) {
	s := New(DefaultConfig(), &mockProvider{})
	onStart, onComplete := s.CreateEventNotifier()
	onStart("task-1", "build-1", "cpp", "running", "worker-1", 100)
	onComplete("task-1", "build-1", "cpp", "completed", "worker-1", 100, 120, 20, 5, 15, 0, "")

	tasks := s.hub.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("GetTasks len = %d, want 1", len(tasks))
	}
	if tasks[0].BuildID != "build-1" {
		t.Errorf("BuildID = %q, want build-1", tasks[0].BuildID)
	}
	if tasks[0].Status != "completed" {
		t.Errorf("Status = %q, want completed", tasks[0].Status)
	}
}

func TestHub_GetTasks_Empty(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	tasks := hub.GetTasks()
	if tasks == nil {
		t.Error("GetTasks should return empty slice, not nil")
	}
	if len(tasks) != 0 {
		t.Errorf("GetTasks len = %d, want 0", len(tasks))
	}
}

func TestServer_APIEndpoints_Tasks(t *testing.T) {
	cfg := DefaultConfig()
	s := New(cfg, &mockProvider{})

	go s.hub.Run()
	defer s.hub.Stop()

	tests := []struct {
		path       string
		method     string
		wantStatus int
	}{
		{"/api/v1/tasks", http.MethodGet, http.StatusOK},
		{"/api/v1/tasks", http.MethodPost, http.StatusMethodNotAllowed},
		{"/api/v1/tasks", http.MethodDelete, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method+" "+tt.path, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			s.server.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Status = %d, want %d", rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestWorkerInfo_AllFields_Flutter(t *testing.T) {
	worker := &WorkerInfo{
		ID:                "w1",
		Host:              "host",
		Address:           "addr",
		Architecture:      "x86_64",
		CPUCores:          4,
		MemoryGB:          8.0,
		ActiveTasks:       1,
		TotalTasks:        50,
		SuccessRate:       0.9,
		AvgLatencyMs:      15.0,
		CircuitState:      "CLOSED",
		DiscoverySource:   "mdns",
		Healthy:           true,
		LastSeen:          123456,
		FlutterAvailable:  true,
		FlutterSDKVersion: "3.19.6",
		FlutterPlatforms:  []string{"TARGET_PLATFORM_ANDROID"},
		Compilers:         []string{"gcc", "g++"},
		BuildTypes:        []string{"BUILD_TYPE_CPP", "BUILD_TYPE_FLUTTER"},
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"flutter_available", "flutter_sdk_version", "flutter_platforms",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing Flutter field: %s", f)
		}
	}
}

func TestStats_UnityFields(t *testing.T) {
	stats := &Stats{
		TotalTasks:          100,
		SuccessTasks:        80,
		FailedTasks:         20,
		ActiveTasks:         5,
		QueuedTasks:         10,
		CacheHits:           60,
		CacheMisses:         40,
		CacheHitRate:        0.6,
		FlutterBuilds:       30,
		FlutterCacheHits:    20,
		FlutterCacheMisses:  10,
		FlutterCacheHitRate: 0.67,
		UnityBuilds:         25,
		UnityCacheHits:      15,
		UnityCacheMisses:    10,
		UnityCacheHitRate:   0.6,
		TotalWorkers:        4,
		HealthyWorkers:      3,
		UptimeSeconds:       7200,
		Timestamp:           1234567890,
	}

	data, err := json.Marshal(stats)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"unity_builds", "unity_cache_hits",
		"unity_cache_misses", "unity_cache_hit_rate",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing Unity field: %s", f)
		}
	}

	var decoded Stats
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if decoded.UnityBuilds != 25 {
		t.Errorf("UnityBuilds = %d, want 25", decoded.UnityBuilds)
	}
	if decoded.UnityCacheHits != 15 {
		t.Errorf("UnityCacheHits = %d, want 15", decoded.UnityCacheHits)
	}
	if decoded.UnityCacheMisses != 10 {
		t.Errorf("UnityCacheMisses = %d, want 10", decoded.UnityCacheMisses)
	}
	if decoded.UnityCacheHitRate != 0.6 {
		t.Errorf("UnityCacheHitRate = %f, want 0.6", decoded.UnityCacheHitRate)
	}
}

func TestWorkerInfo_UnityFields(t *testing.T) {
	worker := &WorkerInfo{
		ID:                "w1",
		Host:              "host",
		Address:           "addr",
		Architecture:      "x86_64",
		CPUCores:          8,
		MemoryGB:          32.0,
		ActiveTasks:       1,
		TotalTasks:        50,
		SuccessRate:       0.9,
		AvgLatencyMs:      15.0,
		CircuitState:      "CLOSED",
		DiscoverySource:   "mdns",
		Healthy:           true,
		LastSeen:          123456,
		FlutterAvailable:  true,
		FlutterSDKVersion: "3.19.6",
		FlutterPlatforms:  []string{"TARGET_PLATFORM_ANDROID"},
		UnityAvailable:    true,
		UnityVersions:     []string{"2022.3.10f1", "2023.2.1f1"},
		UnityPlatforms:    []string{"UNITY_BUILD_TARGET_ANDROID", "UNITY_BUILD_TARGET_STANDALONE_WINDOWS64"},
		Compilers:         []string{"gcc", "g++"},
		BuildTypes:        []string{"BUILD_TYPE_CPP", "BUILD_TYPE_FLUTTER", "BUILD_TYPE_UNITY"},
	}

	data, err := json.Marshal(worker)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	str := string(data)
	fields := []string{
		"unity_available", "unity_versions", "unity_platforms",
	}
	for _, f := range fields {
		if !strings.Contains(str, f) {
			t.Errorf("JSON missing Unity field: %s", f)
		}
	}

	var decoded WorkerInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if !decoded.UnityAvailable {
		t.Error("UnityAvailable = false, want true")
	}
	if len(decoded.UnityVersions) != 2 {
		t.Errorf("UnityVersions len = %d, want 2", len(decoded.UnityVersions))
	}
	if decoded.UnityVersions[0] != "2022.3.10f1" {
		t.Errorf("UnityVersions[0] = %s, want 2022.3.10f1", decoded.UnityVersions[0])
	}
	if len(decoded.UnityPlatforms) != 2 {
		t.Errorf("UnityPlatforms len = %d, want 2", len(decoded.UnityPlatforms))
	}
}

func TestServer_HandleBuildByID(t *testing.T) {
	s := New(DefaultConfig(), &mockProvider{})
	s.hub.BroadcastTaskCompleted(&TaskInfo{ID: "old", BuildID: "build-1", BuildType: "cpp", Status: "completed", QueueMs: 10, CompileMs: 15})
	s.hub.BroadcastTaskCompleted(&TaskInfo{ID: "new", BuildID: "build-1", BuildType: "cpp", Status: "completed", QueueMs: 20, CompileMs: 30})

	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/builds/build-1", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, want 200", rec.Code)
	}
	var response struct {
		Build *BuildInfo  `json:"build"`
		Tasks []*TaskInfo `json:"tasks"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Build == nil || response.Build.ID != "build-1" {
		t.Fatalf("build = %#v, want build-1", response.Build)
	}
	if len(response.Tasks) != 2 {
		t.Fatalf("tasks len = %d, want 2", len(response.Tasks))
	}
	if response.Tasks[0].ID != "new" || response.Tasks[0].QueueMs != 20 || response.Tasks[0].CompileMs != 30 {
		t.Errorf("newest task = %#v, want propagated timing fields", response.Tasks[0])
	}

	rec = httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/builds/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("missing Status = %d, want 404", rec.Code)
	}
	rec = httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/builds/build-1", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST Status = %d, want 405", rec.Code)
	}
}

// consoleMockProvider extends mockProvider with console retention.
type consoleMockProvider struct {
	mockProvider
	logs map[string]string
}

func (m *consoleMockProvider) GetConsole(taskID string) (stdout, stderr string, truncated bool, ok bool) {
	out, exists := m.logs[taskID]
	return out, "", false, exists
}

func TestServer_HandleTaskConsole(t *testing.T) {
	s := New(DefaultConfig(), &consoleMockProvider{
		logs: map[string]string{"task-1": "gcc: compiled main.c"},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/task-1/console", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Status = %d, want 200", rec.Code)
	}
	var response struct {
		TaskID    string `json:"task_id"`
		Stdout    string `json:"stdout"`
		Truncated bool   `json:"truncated"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	if response.TaskID != "task-1" {
		t.Errorf("task_id = %q, want task-1", response.TaskID)
	}
	if response.Stdout != "gcc: compiled main.c" {
		t.Errorf("stdout = %q, want gcc: compiled main.c", response.Stdout)
	}
	if response.Truncated {
		t.Error("truncated = true, want false")
	}
}

func TestServer_HandleTaskConsole_NotFound(t *testing.T) {
	s := New(DefaultConfig(), &consoleMockProvider{logs: map[string]string{}})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/unknown/console", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Status = %d, want 404", rec.Code)
	}
}

func TestServer_HandleTaskConsole_MethodNotAllowed(t *testing.T) {
	s := New(DefaultConfig(), &consoleMockProvider{logs: map[string]string{}})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks/task-1/console", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Status = %d, want 405", rec.Code)
	}
}

func TestServer_HandleTaskConsole_NoProvider(t *testing.T) {
	// A plain StatsProvider without console support must yield 503, not
	// an empty 200 that would read as "task produced no output".
	s := New(DefaultConfig(), &mockProvider{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/task-1/console", nil)
	rec := httptest.NewRecorder()
	s.server.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("Status = %d, want 503", rec.Code)
	}
}
