package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
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
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	File   string `mapstructure:"file"`
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
`
	return os.WriteFile(path, []byte(example), 0644)
}
