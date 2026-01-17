package tls

import (
	"crypto/tls"
	"fmt"
)

// Config holds TLS configuration options.
type Config struct {
	// Enabled determines whether TLS is active
	Enabled bool `yaml:"enabled" json:"enabled"`

	// CertFile path to the server certificate (PEM format)
	CertFile string `yaml:"cert_file" json:"cert_file"`

	// KeyFile path to the server private key (PEM format)
	KeyFile string `yaml:"key_file" json:"key_file"`

	// ClientCA path to the CA certificate for client verification (mTLS)
	ClientCA string `yaml:"client_ca" json:"client_ca"`

	// InsecureSkipVerify disables server certificate verification (testing only)
	InsecureSkipVerify bool `yaml:"insecure_skip_verify" json:"insecure_skip_verify"`

	// MinVersion specifies the minimum TLS version (default: TLS 1.2)
	MinVersion uint16 `yaml:"min_version" json:"min_version"`

	// RequireClientCert requires clients to present certificates (mTLS)
	RequireClientCert bool `yaml:"require_client_cert" json:"require_client_cert"`
}

// DefaultConfig returns a sensible default TLS configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:            false,
		MinVersion:         tls.VersionTLS12,
		InsecureSkipVerify: false,
		RequireClientCert:  false,
	}
}

// Validate checks if the configuration is valid.
func (c *Config) Validate() error {
	if !c.Enabled {
		return nil
	}

	// InsecureSkipVerify allows TLS without certificates (for testing)
	if !c.InsecureSkipVerify {
		if c.CertFile == "" {
			return fmt.Errorf("tls: cert_file is required when TLS is enabled")
		}
		if c.KeyFile == "" {
			return fmt.Errorf("tls: key_file is required when TLS is enabled")
		}
	}
	if c.RequireClientCert && c.ClientCA == "" {
		return fmt.Errorf("tls: client_ca is required when require_client_cert is true")
	}
	if c.MinVersion != 0 && c.MinVersion < tls.VersionTLS12 {
		return fmt.Errorf("tls: min_version must be at least TLS 1.2")
	}

	return nil
}

// IsEnabled returns true if TLS is enabled.
func (c *Config) IsEnabled() bool {
	return c.Enabled
}

// IsMTLS returns true if mutual TLS is enabled.
func (c *Config) IsMTLS() bool {
	return c.Enabled && c.RequireClientCert
}

// MinVersionName returns the human-readable name of the minimum TLS version.
func (c *Config) MinVersionName() string {
	switch c.MinVersion {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return fmt.Sprintf("unknown (0x%04x)", c.MinVersion)
	}
}
