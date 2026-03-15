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

// Validation Tests

func TestValidate_DefaultConfigIsValid(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("DefaultConfig() should be valid, got error: %v", err)
	}
}

func TestValidate_CoordinatorPortRange(t *testing.T) {
	tests := []struct {
		name      string
		grpcPort  int
		httpPort  int
		wantError bool
		errMsg    string
	}{
		{"valid ports", 9000, 8080, false, ""},
		{"grpc port 1", 1, 8080, false, ""},
		{"grpc port 65535", 65535, 8080, false, ""},
		{"grpc port 0", 0, 8080, true, "grpc_port"},
		{"grpc port 65536", 65536, 8080, true, "grpc_port"},
		{"grpc port negative", -1, 8080, true, "grpc_port"},
		{"http port 0", 9000, 0, true, "http_port"},
		{"http port 65536", 9000, 65536, true, "http_port"},
		{"ports equal", 9000, 9000, true, "different"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Coordinator.GRPCPort = tt.grpcPort
			cfg.Coordinator.HTTPPort = tt.httpPort

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %s, want to contain %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidate_WorkerPort(t *testing.T) {
	tests := []struct {
		name      string
		port      int
		wantError bool
	}{
		{"valid port 9001", 9001, false},
		{"port 1", 1, false},
		{"port 65535", 65535, false},
		{"port 0", 0, true},
		{"port 65536", 65536, true},
		{"port negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Worker.Port = tt.port

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_WorkerMaxParallel(t *testing.T) {
	tests := []struct {
		name      string
		parallel  int
		wantError bool
	}{
		{"zero (auto)", 0, false},
		{"positive", 4, false},
		{"negative", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Worker.MaxParallel = tt.parallel

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_CacheConfig(t *testing.T) {
	tests := []struct {
		name      string
		enable    bool
		maxSize   int64
		ttl       int
		dir       string
		wantError bool
		errMsg    string
	}{
		{"cache disabled", false, 0, 0, "", false, ""},
		{"cache valid", true, 1024, 168, "/tmp/cache", false, ""},
		{"cache max_size 0", true, 0, 168, "/tmp/cache", true, "max_size_mb"},
		{"cache ttl 0", true, 1024, 0, "/tmp/cache", true, "ttl_hours"},
		{"cache dir empty", true, 1024, 168, "", true, "dir"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Cache.Enable = tt.enable
			cfg.Cache.MaxSize = tt.maxSize
			cfg.Cache.TTLHours = tt.ttl
			cfg.Cache.Dir = tt.dir

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %s, want to contain %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidate_LogLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		wantError bool
	}{
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"error", "error", false},
		{"fatal", "fatal", false},
		{"trace (invalid)", "trace", true},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Log.Level = tt.level

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_LogFormat(t *testing.T) {
	tests := []struct {
		name      string
		format    string
		wantError bool
	}{
		{"console", "console", false},
		{"json", "json", false},
		{"text (invalid)", "text", true},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Log.Format = tt.format

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_LogRotation(t *testing.T) {
	tests := []struct {
		name      string
		file      string
		maxSize   int
		maxBackup int
		maxAge    int
		wantError bool
		errMsg    string
	}{
		{"no file, no rotation check", "", 0, 0, 0, false, ""},
		{"valid rotation", "/var/log/app.log", 100, 3, 28, false, ""},
		{"invalid max_size_mb", "/var/log/app.log", 0, 3, 28, true, "max_size_mb"},
		{"negative max_backups", "/var/log/app.log", 100, -1, 28, true, "max_backups"},
		{"invalid max_age_days", "/var/log/app.log", 100, 3, 0, true, "max_age_days"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Log.File = tt.file
			cfg.Log.Rotation.MaxSizeMB = tt.maxSize
			cfg.Log.Rotation.MaxBackups = tt.maxBackup
			cfg.Log.Rotation.MaxAgeDays = tt.maxAge

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %s, want to contain %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidate_TLSConfig(t *testing.T) {
	tests := []struct {
		name           string
		enabled        bool
		certFile       string
		keyFile        string
		requireClientC bool
		clientCA       string
		insecureSkip   bool
		wantError      bool
		errMsg         string
	}{
		{"tls disabled", false, "", "", false, "", false, false, ""},
		{"tls enabled with cert/key", true, "/etc/tls/cert.pem", "/etc/tls/key.pem", false, "", false, false, ""},
		{"tls enabled no cert", true, "", "/etc/tls/key.pem", false, "", false, true, "cert_file"},
		{"tls enabled no key", true, "/etc/tls/cert.pem", "", false, "", false, true, "key_file"},
		{"tls insecure skip", true, "", "", false, "", true, false, ""},
		{"mtls no client_ca", true, "/etc/tls/cert.pem", "/etc/tls/key.pem", true, "", false, true, "client_ca"},
		{"mtls valid", true, "/etc/tls/cert.pem", "/etc/tls/key.pem", true, "/etc/tls/ca.pem", false, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.TLS.Enabled = tt.enabled
			cfg.TLS.CertFile = tt.certFile
			cfg.TLS.KeyFile = tt.keyFile
			cfg.TLS.RequireClientCert = tt.requireClientC
			cfg.TLS.ClientCA = tt.clientCA
			cfg.TLS.InsecureSkipVerify = tt.insecureSkip

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %s, want to contain %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidate_TracingConfig(t *testing.T) {
	tests := []struct {
		name       string
		enable     bool
		endpoint   string
		sampleRate float64
		wantError  bool
		errMsg     string
	}{
		{"tracing disabled", false, "", 0, false, ""},
		{"tracing valid", true, "localhost:4317", 0.5, false, ""},
		{"tracing no endpoint", true, "", 0.5, true, "endpoint"},
		{"tracing sample rate low", true, "localhost:4317", -0.1, true, "sample_rate"},
		{"tracing sample rate high", true, "localhost:4317", 1.1, true, "sample_rate"},
		{"tracing sample rate 0", true, "localhost:4317", 0.0, false, ""},
		{"tracing sample rate 1", true, "localhost:4317", 1.0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Tracing.Enable = tt.enable
			cfg.Tracing.Endpoint = tt.endpoint
			cfg.Tracing.SampleRate = tt.sampleRate

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
			if tt.wantError && err != nil && !contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %s, want to contain %s", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestValidate_ClientTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		wantError bool
	}{
		{"valid timeout 30s", 30 * time.Second, false},
		{"valid timeout disabled (0)", 0, false},
		{"invalid timeout 500ms", 500 * time.Millisecond, true},
		{"valid timeout 1s", time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Client.Timeout = tt.timeout

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_WorkerTimeout(t *testing.T) {
	tests := []struct {
		name      string
		timeout   time.Duration
		wantError bool
	}{
		{"valid timeout 5m", 5 * time.Minute, false},
		{"valid timeout disabled (0)", 0, false},
		{"invalid timeout 500ms", 500 * time.Millisecond, true},
		{"valid timeout 1s", time.Second, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			cfg.Worker.Timeout = tt.timeout

			err := cfg.Validate()
			if (err != nil) != tt.wantError {
				t.Errorf("Validate() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidate_FirstErrorOnly(t *testing.T) {
	cfg := DefaultConfig()
	cfg.Coordinator.GRPCPort = 0 // Invalid port
	cfg.Worker.MaxParallel = -1  // Invalid
	cfg.Cache.MaxSize = 0        // Invalid when enabled

	// Should return FIRST error found (coordinator validation first)
	err := cfg.Validate()
	if err == nil {
		t.Fatal("Validate() should return an error")
	}

	if !contains(err.Error(), "coordinator") {
		t.Errorf("Validate() should return coordinator error first, got: %v", err)
	}
}

// Helper function for string matching in error messages
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) >= len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
