package mdns

import (
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestBuildTXTRecords(t *testing.T) {
	caps := &pb.WorkerCapabilities{
		WorkerId:         "test-worker-1",
		Hostname:         "testhost",
		CpuCores:         8,
		MemoryBytes:      16 * 1024 * 1024 * 1024, // 16GB
		NativeArch:       pb.Architecture_ARCH_X86_64,
		DockerAvailable:  true,
		DockerImages:     []string{"dockcross/linux-x64", "dockcross/linux-arm64"},
		MaxParallelTasks: 4,
		Version:          "1.0.0",
		Os:               "linux",
	}

	txt := buildTXTRecords(caps)

	// Convert to map for easier testing
	txtMap := ParseTXTRecords(txt)

	tests := []struct {
		key      string
		expected string
	}{
		{"id", "test-worker-1"},
		{"host", "testhost"},
		{"cpu", "8"},
		{"ram", "16G"},
		{"arch", "ARCH_X86_64"},
		{"docker", "true"},
		{"max_parallel", "4"},
		{"version", "1.0.0"},
		{"os", "linux"},
	}

	for _, tt := range tests {
		got := txtMap[tt.key]
		if got != tt.expected {
			t.Errorf("TXT[%s] = %q, want %q", tt.key, got, tt.expected)
		}
	}

	// Check images
	if txtMap["images"] != "dockcross/linux-x64,dockcross/linux-arm64" {
		t.Errorf("TXT[images] = %q, want comma-separated images", txtMap["images"])
	}
}

func TestParseTXTRecords(t *testing.T) {
	txt := []string{
		"id=worker-1",
		"cpu=4",
		"ram=8G",
		"arch=ARCH_ARM64",
		"docker=false",
	}

	result := ParseTXTRecords(txt)

	if result["id"] != "worker-1" {
		t.Errorf("id = %q, want 'worker-1'", result["id"])
	}
	if result["cpu"] != "4" {
		t.Errorf("cpu = %q, want '4'", result["cpu"])
	}
	if result["arch"] != "ARCH_ARM64" {
		t.Errorf("arch = %q, want 'ARCH_ARM64'", result["arch"])
	}
}

func TestParseCapsFromTXT(t *testing.T) {
	txt := map[string]string{
		"id":           "worker-1",
		"host":         "myhost",
		"cpu":          "16",
		"ram":          "32G",
		"arch":         "ARCH_X86_64",
		"docker":       "true",
		"images":       "img1,img2,img3",
		"max_parallel": "8",
		"version":      "2.0.0",
		"os":           "darwin",
	}

	caps := parseCapsFromTXT(txt, nil)

	if caps.WorkerId != "worker-1" {
		t.Errorf("WorkerId = %q, want 'worker-1'", caps.WorkerId)
	}
	if caps.Hostname != "myhost" {
		t.Errorf("Hostname = %q, want 'myhost'", caps.Hostname)
	}
	if caps.CpuCores != 16 {
		t.Errorf("CpuCores = %d, want 16", caps.CpuCores)
	}
	if caps.MemoryBytes != 32*1024*1024*1024 {
		t.Errorf("MemoryBytes = %d, want 32GB", caps.MemoryBytes)
	}
	if caps.NativeArch != pb.Architecture_ARCH_X86_64 {
		t.Errorf("NativeArch = %v, want ARCH_X86_64", caps.NativeArch)
	}
	if !caps.DockerAvailable {
		t.Error("DockerAvailable = false, want true")
	}
	if len(caps.DockerImages) != 3 {
		t.Errorf("DockerImages count = %d, want 3", len(caps.DockerImages))
	}
	if caps.MaxParallelTasks != 8 {
		t.Errorf("MaxParallelTasks = %d, want 8", caps.MaxParallelTasks)
	}
	if caps.Version != "2.0.0" {
		t.Errorf("Version = %q, want '2.0.0'", caps.Version)
	}
	if caps.Os != "darwin" {
		t.Errorf("Os = %q, want 'darwin'", caps.Os)
	}
}

func TestAnnouncer_NewAnnouncer(t *testing.T) {
	cfg := AnnouncerConfig{
		Instance: "test-instance",
		Port:     50052,
	}

	a := NewAnnouncer(cfg)

	if a.instance != "test-instance" {
		t.Errorf("instance = %q, want 'test-instance'", a.instance)
	}
	if a.port != 50052 {
		t.Errorf("port = %d, want 50052", a.port)
	}
}

func TestBrowser_NewBrowser(t *testing.T) {
	callback := func(w *DiscoveredWorker, event string) {
		// Callback for testing
	}

	cfg := BrowserConfig{TTL: 30 * time.Second}
	b := NewBrowser(cfg, callback)

	if b.ttl != 30*time.Second {
		t.Errorf("ttl = %v, want 30s", b.ttl)
	}
	if b.callback == nil {
		t.Error("callback is nil")
	}
}

func TestBrowser_DefaultConfig(t *testing.T) {
	cfg := DefaultBrowserConfig()

	if cfg.TTL != 60*time.Second {
		t.Errorf("default TTL = %v, want 60s", cfg.TTL)
	}
}

func TestBrowser_ListEmpty(t *testing.T) {
	b := NewBrowser(DefaultBrowserConfig(), nil)

	workers := b.List()
	if len(workers) != 0 {
		t.Errorf("List() returned %d workers, want 0", len(workers))
	}

	if b.Count() != 0 {
		t.Errorf("Count() = %d, want 0", b.Count())
	}
}

func TestBrowser_Get(t *testing.T) {
	b := NewBrowser(DefaultBrowserConfig(), nil)

	// Add a worker manually
	worker := &DiscoveredWorker{
		ID:           "test-worker",
		Address:      "192.168.1.100:9001",
		DiscoveredAt: time.Now(),
		Source:       "mdns",
		Capabilities: &pb.WorkerCapabilities{
			WorkerId: "test-worker",
		},
	}
	b.mu.Lock()
	b.workers["test-worker"] = worker
	b.mu.Unlock()

	// Test Get existing worker
	got, ok := b.Get("test-worker")
	if !ok {
		t.Error("Expected to find test-worker")
	}
	if got.ID != "test-worker" {
		t.Errorf("ID = %q, want 'test-worker'", got.ID)
	}
	if got.Address != "192.168.1.100:9001" {
		t.Errorf("Address = %q, want '192.168.1.100:9001'", got.Address)
	}

	// Test Get non-existent worker
	_, ok = b.Get("nonexistent")
	if ok {
		t.Error("Expected not to find nonexistent worker")
	}
}

func TestBrowser_List(t *testing.T) {
	b := NewBrowser(DefaultBrowserConfig(), nil)

	// Add multiple workers
	b.mu.Lock()
	b.workers["worker-1"] = &DiscoveredWorker{ID: "worker-1", DiscoveredAt: time.Now()}
	b.workers["worker-2"] = &DiscoveredWorker{ID: "worker-2", DiscoveredAt: time.Now()}
	b.workers["worker-3"] = &DiscoveredWorker{ID: "worker-3", DiscoveredAt: time.Now()}
	b.mu.Unlock()

	workers := b.List()
	if len(workers) != 3 {
		t.Errorf("List() returned %d workers, want 3", len(workers))
	}

	if b.Count() != 3 {
		t.Errorf("Count() = %d, want 3", b.Count())
	}
}

func TestBrowser_ZeroTTL(t *testing.T) {
	cfg := BrowserConfig{TTL: 0}
	b := NewBrowser(cfg, nil)

	// Default TTL should be applied
	if b.ttl != 60*time.Second {
		t.Errorf("ttl = %v, want 60s (default)", b.ttl)
	}
}

func TestBuildTXTRecords_NilCaps(t *testing.T) {
	txt := buildTXTRecords(nil)
	if len(txt) != 0 {
		t.Errorf("Expected empty TXT records for nil caps, got %d", len(txt))
	}
}

func TestBuildTXTRecords_EmptyCaps(t *testing.T) {
	caps := &pb.WorkerCapabilities{}
	txt := buildTXTRecords(caps)

	// Should still have cpu, ram, arch, docker, max_parallel entries
	if len(txt) < 5 {
		t.Errorf("Expected at least 5 TXT records, got %d", len(txt))
	}

	txtMap := ParseTXTRecords(txt)
	if txtMap["cpu"] != "0" {
		t.Errorf("cpu = %q, want '0'", txtMap["cpu"])
	}
	if txtMap["docker"] != "false" {
		t.Errorf("docker = %q, want 'false'", txtMap["docker"])
	}
}

func TestBuildTXTRecords_ManyDockerImages(t *testing.T) {
	caps := &pb.WorkerCapabilities{
		DockerImages: []string{"img1", "img2", "img3", "img4", "img5", "img6", "img7"},
	}
	txt := buildTXTRecords(caps)
	txtMap := ParseTXTRecords(txt)

	// Should only have first 5 images
	if txtMap["images"] != "img1,img2,img3,img4,img5" {
		t.Errorf("images = %q, want first 5 images only", txtMap["images"])
	}
}

func TestParseCapsFromTXT_DifferentArchs(t *testing.T) {
	tests := []struct {
		arch     string
		expected pb.Architecture
	}{
		{"ARCH_X86_64", pb.Architecture_ARCH_X86_64},
		{"ARCH_ARM64", pb.Architecture_ARCH_ARM64},
		{"ARCH_ARMV7", pb.Architecture_ARCH_ARMV7},
		{"UNKNOWN", pb.Architecture_ARCH_UNSPECIFIED},
	}

	for _, tt := range tests {
		txt := map[string]string{"arch": tt.arch}
		caps := parseCapsFromTXT(txt, nil)
		if caps.NativeArch != tt.expected {
			t.Errorf("arch=%s: NativeArch = %v, want %v", tt.arch, caps.NativeArch, tt.expected)
		}
	}
}

func TestParseCapsFromTXT_InvalidNumbers(t *testing.T) {
	txt := map[string]string{
		"cpu":          "invalid",
		"ram":          "invalidG",
		"max_parallel": "notanumber",
	}
	caps := parseCapsFromTXT(txt, nil)

	// Should default to 0 for invalid numbers
	if caps.CpuCores != 0 {
		t.Errorf("CpuCores = %d, want 0", caps.CpuCores)
	}
	if caps.MemoryBytes != 0 {
		t.Errorf("MemoryBytes = %d, want 0", caps.MemoryBytes)
	}
	if caps.MaxParallelTasks != 0 {
		t.Errorf("MaxParallelTasks = %d, want 0", caps.MaxParallelTasks)
	}
}

func TestParseCapsFromTXT_DockerFalse(t *testing.T) {
	txt := map[string]string{"docker": "false"}
	caps := parseCapsFromTXT(txt, nil)

	if caps.DockerAvailable {
		t.Error("DockerAvailable should be false")
	}
}

func TestParseCapsFromTXT_EmptyImages(t *testing.T) {
	txt := map[string]string{"images": ""}
	caps := parseCapsFromTXT(txt, nil)

	// Empty images string - the condition `images != ""` fails,
	// so DockerImages should remain nil/empty
	if len(caps.DockerImages) != 0 {
		t.Errorf("DockerImages = %v, want empty for empty images string", caps.DockerImages)
	}
}

func TestParseCapsFromTXT_NoImages(t *testing.T) {
	txt := map[string]string{}
	caps := parseCapsFromTXT(txt, nil)

	// No images key should result in nil/empty slice
	if len(caps.DockerImages) != 0 {
		t.Errorf("DockerImages = %v, want empty", caps.DockerImages)
	}
}

func TestParseTXTRecords_MalformedEntry(t *testing.T) {
	txt := []string{
		"validkey=value",
		"noequals",
		"emptyval=",
		"=nokey",
		"multiequals=value=with=equals",
	}

	result := ParseTXTRecords(txt)

	if result["validkey"] != "value" {
		t.Errorf("validkey = %q, want 'value'", result["validkey"])
	}
	if _, ok := result["noequals"]; ok {
		t.Error("Should not have parsed 'noequals'")
	}
	if result["emptyval"] != "" {
		t.Errorf("emptyval = %q, want empty string", result["emptyval"])
	}
	// "=nokey" splits to ["", "nokey"]
	if result[""] != "nokey" {
		t.Errorf("empty key = %q, want 'nokey'", result[""])
	}
	// "multiequals=value=with=equals" with SplitN(..., 2) gives ["multiequals", "value=with=equals"]
	if result["multiequals"] != "value=with=equals" {
		t.Errorf("multiequals = %q, want 'value=with=equals'", result["multiequals"])
	}
}

func TestAnnouncer_Stop_NotStarted(t *testing.T) {
	a := NewAnnouncer(AnnouncerConfig{Instance: "test", Port: 9001})

	// Stop without starting should not panic
	a.Stop()
}

func TestBrowser_Stop_NotStarted(t *testing.T) {
	b := NewBrowser(DefaultBrowserConfig(), nil)

	// Stop without starting should not panic
	b.Stop()
}

func TestBrowser_Cleanup(t *testing.T) {
	b := NewBrowser(BrowserConfig{TTL: 100 * time.Millisecond}, nil)

	// Add a worker with old discovery time
	b.mu.Lock()
	b.workers["old-worker"] = &DiscoveredWorker{
		ID:           "old-worker",
		DiscoveredAt: time.Now().Add(-1 * time.Second), // 1 second ago
	}
	b.workers["fresh-worker"] = &DiscoveredWorker{
		ID:           "fresh-worker",
		DiscoveredAt: time.Now(), // just now
	}
	b.mu.Unlock()

	// Run cleanup
	b.cleanup()

	// Old worker should be removed
	_, exists := b.Get("old-worker")
	if exists {
		t.Error("old-worker should have been cleaned up")
	}

	// Fresh worker should still exist
	_, exists = b.Get("fresh-worker")
	if !exists {
		t.Error("fresh-worker should still exist")
	}
}

func TestBrowser_CleanupCallback(t *testing.T) {
	lostWorkers := make([]string, 0)
	callback := func(w *DiscoveredWorker, event string) {
		if event == "lost" {
			lostWorkers = append(lostWorkers, w.ID)
		}
	}

	b := NewBrowser(BrowserConfig{TTL: 100 * time.Millisecond}, callback)

	// Add an old worker
	b.mu.Lock()
	b.workers["expired-worker"] = &DiscoveredWorker{
		ID:           "expired-worker",
		DiscoveredAt: time.Now().Add(-1 * time.Second),
	}
	b.mu.Unlock()

	// Run cleanup
	b.cleanup()

	// Callback should have been called
	if len(lostWorkers) != 1 || lostWorkers[0] != "expired-worker" {
		t.Errorf("Expected callback for 'expired-worker', got %v", lostWorkers)
	}
}

func TestConstants(t *testing.T) {
	if ServiceType != "_hybridgrid._tcp" {
		t.Errorf("ServiceType = %q, want '_hybridgrid._tcp'", ServiceType)
	}
	if Domain != "local." {
		t.Errorf("Domain = %q, want 'local.'", Domain)
	}
}

// Integration test - requires network
// NOTE: This test may panic due to zeroconf library issue with context cancellation
// Run manually with: go test -v -run TestMDNS_Integration ./internal/discovery/mdns/...
func TestMDNS_Integration(t *testing.T) {
	t.Skip("Skipping mDNS integration test - zeroconf library has channel closing issue")

	// Create announcer
	announcerCfg := AnnouncerConfig{
		Instance: "test-worker-integration",
		Port:     59999,
	}
	announcer := NewAnnouncer(announcerCfg)

	caps := &pb.WorkerCapabilities{
		WorkerId:         "integration-worker",
		Hostname:         "testhost",
		CpuCores:         4,
		MemoryBytes:      8 * 1024 * 1024 * 1024,
		NativeArch:       pb.Architecture_ARCH_X86_64,
		DockerAvailable:  true,
		MaxParallelTasks: 2,
	}

	// Start announcer
	if err := announcer.Start(caps); err != nil {
		t.Fatalf("Failed to start announcer: %v", err)
	}

	// Create browser with shorter TTL for testing
	found := make(chan *DiscoveredWorker, 1)
	callback := func(w *DiscoveredWorker, event string) {
		if event == "found" && w.ID == "integration-worker" {
			select {
			case found <- w:
			default:
			}
		}
	}

	browser := NewBrowser(BrowserConfig{TTL: 10 * time.Second}, callback)
	if err := browser.Start(); err != nil {
		announcer.Stop()
		t.Fatalf("Failed to start browser: %v", err)
	}

	// Wait for discovery with timeout
	discoveryTimeout := time.After(5 * time.Second)
	var discoveredWorker *DiscoveredWorker

	select {
	case worker := <-found:
		discoveredWorker = worker
		t.Logf("Discovered worker: %s at %s", worker.ID, worker.Address)
		if worker.Capabilities.CpuCores != 4 {
			t.Errorf("CpuCores = %d, want 4", worker.Capabilities.CpuCores)
		}
	case <-discoveryTimeout:
		t.Log("mDNS discovery timed out - this may be expected in some network environments")
	}

	// Stop browser first, then announcer (order matters for cleanup)
	browser.Stop()
	time.Sleep(50 * time.Millisecond) // Give browser time to cleanup
	announcer.Stop()

	if discoveredWorker != nil {
		t.Logf("Test passed - worker discovered successfully")
	}
}
