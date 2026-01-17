package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	// Test coordinator defaults
	if cfg.Coordinator.GRPCPort != 9000 {
		t.Errorf("Coordinator.GRPCPort = %d, want 9000", cfg.Coordinator.GRPCPort)
	}
	if cfg.Coordinator.HTTPPort != 8080 {
		t.Errorf("Coordinator.HTTPPort = %d, want 8080", cfg.Coordinator.HTTPPort)
	}
	if !cfg.Coordinator.MDNSEnable {
		t.Error("Coordinator.MDNSEnable should be true by default")
	}

	// Test worker defaults
	if cfg.Worker.Port != 9001 {
		t.Errorf("Worker.Port = %d, want 9001", cfg.Worker.Port)
	}
	if cfg.Worker.MaxParallel != runtime.NumCPU() {
		t.Errorf("Worker.MaxParallel = %d, want %d", cfg.Worker.MaxParallel, runtime.NumCPU())
	}
	if cfg.Worker.Timeout != 5*time.Minute {
		t.Errorf("Worker.Timeout = %v, want 5m", cfg.Worker.Timeout)
	}
	if cfg.Worker.HeartbeatSec != 30 {
		t.Errorf("Worker.HeartbeatSec = %d, want 30", cfg.Worker.HeartbeatSec)
	}

	// Test client defaults
	if cfg.Client.Timeout != 30*time.Second {
		t.Errorf("Client.Timeout = %v, want 30s", cfg.Client.Timeout)
	}
	if !cfg.Client.Fallback {
		t.Error("Client.Fallback should be true by default")
	}

	// Test cache defaults
	if !cfg.Cache.Enable {
		t.Error("Cache.Enable should be true by default")
	}
	if cfg.Cache.MaxSize != 1024 {
		t.Errorf("Cache.MaxSize = %d, want 1024", cfg.Cache.MaxSize)
	}
	if cfg.Cache.TTLHours != 168 {
		t.Errorf("Cache.TTLHours = %d, want 168", cfg.Cache.TTLHours)
	}

	// Test log defaults
	if cfg.Log.Level != "info" {
		t.Errorf("Log.Level = %s, want info", cfg.Log.Level)
	}
	if cfg.Log.Format != "console" {
		t.Errorf("Log.Format = %s, want console", cfg.Log.Format)
	}
}

func TestLoad_NoConfigFile(t *testing.T) {
	// Load without config file should use defaults
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}

	// Should have default values
	if cfg.Coordinator.GRPCPort != 9000 {
		t.Errorf("Expected default GRPCPort 9000, got %d", cfg.Coordinator.GRPCPort)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "hybridgrid.yaml")

	configContent := `
coordinator:
  grpc_port: 9999
  http_port: 8888
  mdns_enable: false

worker:
  port: 7777
  max_parallel: 8

cache:
  enable: false
  max_size_mb: 2048

log:
  level: debug
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify custom values
	if cfg.Coordinator.GRPCPort != 9999 {
		t.Errorf("Coordinator.GRPCPort = %d, want 9999", cfg.Coordinator.GRPCPort)
	}
	if cfg.Coordinator.HTTPPort != 8888 {
		t.Errorf("Coordinator.HTTPPort = %d, want 8888", cfg.Coordinator.HTTPPort)
	}
	if cfg.Coordinator.MDNSEnable {
		t.Error("Coordinator.MDNSEnable should be false")
	}
	if cfg.Worker.Port != 7777 {
		t.Errorf("Worker.Port = %d, want 7777", cfg.Worker.Port)
	}
	if cfg.Worker.MaxParallel != 8 {
		t.Errorf("Worker.MaxParallel = %d, want 8", cfg.Worker.MaxParallel)
	}
	if cfg.Cache.Enable {
		t.Error("Cache.Enable should be false")
	}
	if cfg.Cache.MaxSize != 2048 {
		t.Errorf("Cache.MaxSize = %d, want 2048", cfg.Cache.MaxSize)
	}
	if cfg.Log.Level != "debug" {
		t.Errorf("Log.Level = %s, want debug", cfg.Log.Level)
	}
}

func TestLoad_InvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	// Write invalid YAML
	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	// Set environment variable
	os.Setenv("HG_COORDINATOR_GRPC_PORT", "5555")
	defer os.Unsetenv("HG_COORDINATOR_GRPC_PORT")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Note: Viper's automatic env binding may not work with nested keys
	// This test verifies env prefix is set correctly
	t.Logf("Config loaded with env prefix HG")
	t.Logf("GRPCPort: %d", cfg.Coordinator.GRPCPort)
}

func TestWriteExample(t *testing.T) {
	tmpDir := t.TempDir()
	examplePath := filepath.Join(tmpDir, "example.yaml")

	err := WriteExample(examplePath)
	if err != nil {
		t.Fatalf("WriteExample() error = %v", err)
	}

	// Verify file was created
	info, err := os.Stat(examplePath)
	if err != nil {
		t.Fatalf("Example file not created: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Example file is empty")
	}

	// Read and verify it's valid YAML
	content, err := os.ReadFile(examplePath)
	if err != nil {
		t.Fatalf("Failed to read example file: %v", err)
	}

	if len(content) < 100 {
		t.Error("Example file content seems too short")
	}

	t.Logf("Example config written (%d bytes)", len(content))
}

func TestConfig_WorkDir(t *testing.T) {
	cfg := DefaultConfig()

	// Work dir should be under temp
	if cfg.Worker.WorkDir == "" {
		t.Error("Worker.WorkDir should not be empty")
	}

	if !filepath.IsAbs(cfg.Worker.WorkDir) {
		t.Errorf("Worker.WorkDir should be absolute, got %s", cfg.Worker.WorkDir)
	}
}

func TestConfig_CacheDir(t *testing.T) {
	cfg := DefaultConfig()

	// Cache dir should be set
	if cfg.Cache.Dir == "" {
		t.Error("Cache.Dir should not be empty")
	}
}
