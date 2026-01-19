package tracing

import (
	"time"
)

// Config holds tracing configuration.
type Config struct {
	// Enable enables or disables tracing
	Enable bool `mapstructure:"enable" yaml:"enable"`

	// Endpoint is the OTLP gRPC endpoint (e.g., "localhost:4317")
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`

	// ServiceName is the name of this service in traces
	ServiceName string `mapstructure:"service_name" yaml:"service_name"`

	// SampleRate is the sampling rate (0.0 to 1.0)
	SampleRate float64 `mapstructure:"sample_rate" yaml:"sample_rate"`

	// Insecure disables TLS for the OTLP connection
	Insecure bool `mapstructure:"insecure" yaml:"insecure"`

	// Headers are additional headers to send with OTLP requests
	Headers map[string]string `mapstructure:"headers" yaml:"headers"`

	// Timeout is the timeout for OTLP exports
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`

	// BatchSize is the maximum number of spans to batch before export
	BatchSize int `mapstructure:"batch_size" yaml:"batch_size"`
}

// DefaultConfig returns sensible default tracing configuration.
func DefaultConfig() Config {
	return Config{
		Enable:      false,
		Endpoint:    "localhost:4317",
		ServiceName: "hybridgrid",
		SampleRate:  0.1, // 10% sampling by default
		Insecure:    true,
		Timeout:     10 * time.Second,
		BatchSize:   512,
	}
}

// CoordinatorConfig returns default config for coordinator.
func CoordinatorConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = "hg-coord"
	return cfg
}

// WorkerConfig returns default config for worker.
func WorkerConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = "hg-worker"
	return cfg
}

// ClientConfig returns default config for CLI client.
func ClientConfig() Config {
	cfg := DefaultConfig()
	cfg.ServiceName = "hgbuild"
	cfg.SampleRate = 0.01 // Lower sample rate for client
	return cfg
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if !c.Enable {
		return nil
	}

	if c.Endpoint == "" {
		return ErrEndpointRequired
	}

	if c.SampleRate < 0 || c.SampleRate > 1 {
		return ErrInvalidSampleRate
	}

	if c.ServiceName == "" {
		c.ServiceName = "hybridgrid"
	}

	return nil
}
