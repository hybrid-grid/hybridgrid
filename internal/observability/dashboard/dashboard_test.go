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
	if !strings.Contains(body, "alpinejs") {
		t.Error("Response should contain Alpine.js")
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
		ID:           "task-1",
		BuildType:    "cpp",
		Status:       "completed",
		WorkerID:     "worker-1",
		StartedAt:    1234567890,
		CompletedAt:  1234567895,
		DurationMs:   5000,
		ExitCode:     0,
		FromCache:    true,
		ErrorMessage: "",
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
}

func TestTaskInfo_JSON_WithError(t *testing.T) {
	task := &TaskInfo{
		ID:           "task-2",
		BuildType:    "rust",
		Status:       "failed",
		WorkerID:     "worker-2",
		StartedAt:    1234567890,
		CompletedAt:  1234567900,
		DurationMs:   10000,
		ExitCode:     1,
		FromCache:    false,
		ErrorMessage: "compilation error",
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
