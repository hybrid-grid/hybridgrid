package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"

	"github.com/h3nr1-d14z/hybridgrid/internal/observability/tracing"
)

// Config holds the application configuration.
type Config struct {
	// Coordinator settings
	Coordinator CoordinatorConfig `mapstructure:"coordinator"`

	// Worker settings
	Worker WorkerConfig `mapstructure:"worker"`

	// Client settings
	Client ClientConfig `mapstructure:"client"`

	// Cache settings
	Cache CacheConfig `mapstructure:"cache"`

	// Logging settings
	Log LogConfig `mapstructure:"log"`

	// TLS settings
	TLS TLSConfig `mapstructure:"tls"`

	// Tracing settings
	Tracing TracingConfig `mapstructure:"tracing"`
}

// TLSConfig holds TLS/mTLS settings.
type TLSConfig struct {
	Enabled            bool   `mapstructure:"enabled"`
	CertFile           string `mapstructure:"cert_file"`
	KeyFile            string `mapstructure:"key_file"`
	ClientCA           string `mapstructure:"client_ca"`
	RequireClientCert  bool   `mapstructure:"require_client_cert"`
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

// TracingConfig holds OpenTelemetry tracing settings.
type TracingConfig struct {
	Enable      bool              `mapstructure:"enable"`
	Endpoint    string            `mapstructure:"endpoint"`
	ServiceName string            `mapstructure:"service_name"`
	SampleRate  float64           `mapstructure:"sample_rate"`
	Insecure    bool              `mapstructure:"insecure"`
	Headers     map[string]string `mapstructure:"headers"`
	Timeout     time.Duration     `mapstructure:"timeout"`
	BatchSize   int               `mapstructure:"batch_size"`
}

// LogRotationConfig holds log rotation settings.
type LogRotationConfig struct {
	MaxSizeMB  int  `mapstructure:"max_size_mb"`
	MaxBackups int  `mapstructure:"max_backups"`
	MaxAgeDays int  `mapstructure:"max_age_days"`
	Compress   bool `mapstructure:"compress"`
}

// CoordinatorConfig holds coordinator-specific settings.
type CoordinatorConfig struct {
	GRPCPort   int    `mapstructure:"grpc_port"`
	HTTPPort   int    `mapstructure:"http_port"`
	AuthToken  string `mapstructure:"auth_token"`
	TLSCert    string `mapstructure:"tls_cert"`
	TLSKey     string `mapstructure:"tls_key"`
	MDNSEnable bool   `mapstructure:"mdns_enable"`
}

// WorkerConfig holds worker-specific settings.
type WorkerConfig struct {
	Port            int           `mapstructure:"port"`
	CoordinatorAddr string        `mapstructure:"coordinator_addr"`
	AuthToken       string        `mapstructure:"auth_token"`
	MaxParallel     int           `mapstructure:"max_parallel"`
	WorkDir         string        `mapstructure:"work_dir"`
	Timeout         time.Duration `mapstructure:"timeout"`
	HeartbeatSec    int           `mapstructure:"heartbeat_sec"`
}

// ClientConfig holds client-specific settings.
type ClientConfig struct {
	CoordinatorAddr string        `mapstructure:"coordinator_addr"`
	AuthToken       string        `mapstructure:"auth_token"`
	Timeout         time.Duration `mapstructure:"timeout"`
	Fallback        bool          `mapstructure:"fallback"`
}

// CacheConfig holds cache settings.
type CacheConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Dir      string `mapstructure:"dir"`
	MaxSize  int64  `mapstructure:"max_size_mb"`
	TTLHours int    `mapstructure:"ttl_hours"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level    string            `mapstructure:"level"`
	Format   string            `mapstructure:"format"`
	File     string            `mapstructure:"file"`
	Rotation LogRotationConfig `mapstructure:"rotation"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	cacheDir, _ := os.UserCacheDir()
	return &Config{
		Coordinator: CoordinatorConfig{
			GRPCPort:   9000,
			HTTPPort:   8080,
			MDNSEnable: true,
		},
		Worker: WorkerConfig{
			Port:         9001,
			MaxParallel:  runtime.NumCPU(),
			WorkDir:      filepath.Join(os.TempDir(), "hybridgrid-worker"),
			Timeout:      5 * time.Minute,
			HeartbeatSec: 30,
		},
		Client: ClientConfig{
			Timeout:  30 * time.Second,
			Fallback: true,
		},
		Cache: CacheConfig{
			Enable:   true,
			Dir:      filepath.Join(cacheDir, "hybridgrid"),
			MaxSize:  1024, // 1GB
			TTLHours: 168,  // 7 days
		},
		Log: LogConfig{
			Level:  "info",
			Format: "console",
			Rotation: LogRotationConfig{
				MaxSizeMB:  100,
				MaxBackups: 3,
				MaxAgeDays: 28,
				Compress:   true,
			},
		},
		Tracing: TracingConfig{
			Enable:      false,
			Endpoint:    "localhost:4317",
			ServiceName: "hybridgrid",
			SampleRate:  0.1,
			Insecure:    true,
			Headers:     make(map[string]string),
			Timeout:     10 * time.Second,
			BatchSize:   512,
		},
	}
}

// Load loads configuration from file and environment.
func Load(configPath string) (*Config, error) {
	cfg := DefaultConfig()

	v := viper.New()
	v.SetConfigType("yaml")

	// Set defaults
	setDefaults(v, cfg)

	// Config file locations
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("hybridgrid")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME/.config/hybridgrid")
		v.AddConfigPath("/etc/hybridgrid")
	}

	// Environment variables
	v.SetEnvPrefix("HG")
	v.AutomaticEnv()

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}
		// Config file not found is OK, use defaults
	}

	// Unmarshal
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper, cfg *Config) {
	v.SetDefault("coordinator.grpc_port", cfg.Coordinator.GRPCPort)
	v.SetDefault("coordinator.http_port", cfg.Coordinator.HTTPPort)
	v.SetDefault("coordinator.mdns_enable", cfg.Coordinator.MDNSEnable)

	v.SetDefault("worker.port", cfg.Worker.Port)
	v.SetDefault("worker.max_parallel", cfg.Worker.MaxParallel)
	v.SetDefault("worker.work_dir", cfg.Worker.WorkDir)
	v.SetDefault("worker.timeout", cfg.Worker.Timeout)
	v.SetDefault("worker.heartbeat_sec", cfg.Worker.HeartbeatSec)

	v.SetDefault("client.timeout", cfg.Client.Timeout)
	v.SetDefault("client.fallback", cfg.Client.Fallback)

	v.SetDefault("cache.enable", cfg.Cache.Enable)
	v.SetDefault("cache.dir", cfg.Cache.Dir)
	v.SetDefault("cache.max_size_mb", cfg.Cache.MaxSize)
	v.SetDefault("cache.ttl_hours", cfg.Cache.TTLHours)

	v.SetDefault("log.level", cfg.Log.Level)
	v.SetDefault("log.format", cfg.Log.Format)
	v.SetDefault("log.rotation.max_size_mb", cfg.Log.Rotation.MaxSizeMB)
	v.SetDefault("log.rotation.max_backups", cfg.Log.Rotation.MaxBackups)
	v.SetDefault("log.rotation.max_age_days", cfg.Log.Rotation.MaxAgeDays)
	v.SetDefault("log.rotation.compress", cfg.Log.Rotation.Compress)

	v.SetDefault("tracing.enable", cfg.Tracing.Enable)
	v.SetDefault("tracing.endpoint", cfg.Tracing.Endpoint)
	v.SetDefault("tracing.service_name", cfg.Tracing.ServiceName)
	v.SetDefault("tracing.sample_rate", cfg.Tracing.SampleRate)
	v.SetDefault("tracing.insecure", cfg.Tracing.Insecure)
	v.SetDefault("tracing.headers", cfg.Tracing.Headers)
	v.SetDefault("tracing.timeout", cfg.Tracing.Timeout)
	v.SetDefault("tracing.batch_size", cfg.Tracing.BatchSize)
}

// WriteExample writes an example config file.
func WriteExample(path string) error {
	example := `# Hybrid-Grid Build Configuration

coordinator:
  grpc_port: 9000
  http_port: 8080
  auth_token: ""
  mdns_enable: true
  # tls_cert: /path/to/cert.pem
  # tls_key: /path/to/key.pem

worker:
  port: 9001
  coordinator_addr: ""  # Empty for auto-discovery
  auth_token: ""
  max_parallel: 0       # 0 = auto (number of CPUs)
  work_dir: /tmp/hybridgrid-worker
  timeout: 5m
  heartbeat_sec: 30

client:
  coordinator_addr: ""  # Empty for auto-discovery
  auth_token: ""
  timeout: 30s
  fallback: true        # Fall back to local build if remote fails

cache:
  enable: true
  dir: ~/.cache/hybridgrid
  max_size_mb: 1024     # 1GB
  ttl_hours: 168        # 7 days

log:
  level: info           # debug, info, warn, error
  format: console       # console, json
  # file: /var/log/hybridgrid.log
  rotation:
    max_size_mb: 100    # Max size in MB before rotation
    max_backups: 3      # Max number of backup files to keep
    max_age_days: 28    # Max age in days before deletion
    compress: true      # Compress rotated logs

# TLS / mTLS configuration (optional)
tls:
  enabled: false
  # cert_file: /etc/hybridgrid/tls/server.crt
  # key_file: /etc/hybridgrid/tls/server.key
  # client_ca: /etc/hybridgrid/tls/ca.crt        # CA for verifying client certs (mTLS)
  # require_client_cert: false                     # Enable mTLS
  # insecure_skip_verify: false                    # Skip server cert verification (testing only)

# OpenTelemetry tracing (optional)
tracing:
  enable: false
  endpoint: "localhost:4317"    # OTLP gRPC endpoint
  service_name: "hybridgrid"    # Service name in traces
  sample_rate: 0.1              # 0.0 to 1.0 (10% default)
  insecure: true                # Disable TLS for OTLP connection
  timeout: 10s                  # Timeout for OTLP exports
  batch_size: 512               # Max spans to batch before export
  # headers:                      # Additional OTLP headers (optional)
  #   header_name: header_value
`
	return os.WriteFile(path, []byte(example), 0644)
}

// TracingToLibConfig converts config.TracingConfig to tracing.Config for use with tracing.Init().
func TracingToLibConfig(tc TracingConfig) tracing.Config {
	return tracing.Config{
		Enable:      tc.Enable,
		Endpoint:    tc.Endpoint,
		ServiceName: tc.ServiceName,
		SampleRate:  tc.SampleRate,
		Insecure:    tc.Insecure,
		Headers:     tc.Headers,
		Timeout:     tc.Timeout,
		BatchSize:   tc.BatchSize,
	}
}

// Validate validates all configuration fields and returns the first error found.
func (c *Config) Validate() error {
	// Validate coordinator
	if err := c.Coordinator.Validate(); err != nil {
		return err
	}

	// Validate worker
	if err := c.Worker.Validate(); err != nil {
		return err
	}

	// Validate client
	if err := c.Client.Validate(); err != nil {
		return err
	}

	// Validate cache
	if err := c.Cache.Validate(); err != nil {
		return err
	}

	// Validate log
	if err := c.Log.Validate(); err != nil {
		return err
	}

	// Validate TLS
	if err := c.TLS.Validate(); err != nil {
		return err
	}

	// Validate tracing
	if err := c.Tracing.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates the coordinator configuration.
func (c *CoordinatorConfig) Validate() error {
	if c.GRPCPort < 1 || c.GRPCPort > 65535 {
		return fmt.Errorf("config: coordinator.grpc_port must be 1-65535, got %d", c.GRPCPort)
	}

	if c.HTTPPort < 1 || c.HTTPPort > 65535 {
		return fmt.Errorf("config: coordinator.http_port must be 1-65535, got %d", c.HTTPPort)
	}

	if c.GRPCPort == c.HTTPPort {
		return fmt.Errorf("config: coordinator.grpc_port and coordinator.http_port must be different, got %d for both", c.GRPCPort)
	}

	return nil
}

// Validate validates the worker configuration.
func (c *WorkerConfig) Validate() error {
	if c.Port < 1 || c.Port > 65535 {
		return fmt.Errorf("config: worker.port must be 1-65535, got %d", c.Port)
	}

	if c.MaxParallel < 0 {
		return fmt.Errorf("config: worker.max_parallel must be >= 0, got %d", c.MaxParallel)
	}

	if c.Timeout > 0 && c.Timeout < time.Second {
		return fmt.Errorf("config: worker.timeout must be > 0s, got %v", c.Timeout)
	}

	if c.HeartbeatSec > 0 && c.HeartbeatSec < 1 {
		return fmt.Errorf("config: worker.heartbeat_sec must be > 0, got %d", c.HeartbeatSec)
	}

	return nil
}

// Validate validates the client configuration.
func (c *ClientConfig) Validate() error {
	if c.Timeout > 0 && c.Timeout < time.Second {
		return fmt.Errorf("config: client.timeout must be > 0s or 0 (disabled), got %v", c.Timeout)
	}

	return nil
}

// Validate validates the cache configuration.
func (c *CacheConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.MaxSize <= 0 {
		return fmt.Errorf("config: cache.max_size_mb must be > 0 when cache is enabled, got %d", c.MaxSize)
	}

	if c.TTLHours <= 0 {
		return fmt.Errorf("config: cache.ttl_hours must be > 0 when cache is enabled, got %d", c.TTLHours)
	}

	if c.Dir == "" {
		return errors.New("config: cache.dir must not be empty when cache is enabled")
	}

	return nil
}

// Validate validates the log configuration.
func (c *LogConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}

	if !validLevels[c.Level] {
		return fmt.Errorf("config: log.level must be one of {debug,info,warn,error,fatal}, got %s", c.Level)
	}

	validFormats := map[string]bool{
		"console": true,
		"json":    true,
	}

	if !validFormats[c.Format] {
		return fmt.Errorf("config: log.format must be one of {console,json}, got %s", c.Format)
	}

	// Validate rotation settings only if File is set
	if c.File != "" {
		if err := c.Rotation.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate validates the log rotation configuration.
func (c *LogRotationConfig) Validate() error {
	if c.MaxSizeMB <= 0 {
		return fmt.Errorf("config: log.rotation.max_size_mb must be > 0, got %d", c.MaxSizeMB)
	}

	if c.MaxBackups < 0 {
		return fmt.Errorf("config: log.rotation.max_backups must be >= 0, got %d", c.MaxBackups)
	}

	if c.MaxAgeDays <= 0 {
		return fmt.Errorf("config: log.rotation.max_age_days must be > 0, got %d", c.MaxAgeDays)
	}

	return nil
}

// Validate validates the TLS configuration.
func (c *TLSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}

	// InsecureSkipVerify allows TLS without certificates (for testing)
	if !c.InsecureSkipVerify {
		if c.CertFile == "" {
			return errors.New("config: tls.cert_file is required when TLS is enabled")
		}
		if c.KeyFile == "" {
			return errors.New("config: tls.key_file is required when TLS is enabled")
		}
	}

	if c.RequireClientCert && c.ClientCA == "" {
		return errors.New("config: tls.client_ca is required when tls.require_client_cert is true")
	}

	return nil
}

// Validate validates the tracing configuration.
func (c *TracingConfig) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.Endpoint == "" {
		return errors.New("config: tracing.endpoint is required when tracing is enabled")
	}

	if c.SampleRate < 0 || c.SampleRate > 1 {
		return fmt.Errorf("config: tracing.sample_rate must be 0-1, got %f", c.SampleRate)
	}

	return nil
}
