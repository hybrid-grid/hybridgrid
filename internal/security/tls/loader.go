package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/credentials"
)

// LoadServerTLS creates a TLS configuration for the server.
func LoadServerTLS(cfg Config) (*tls.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return nil, nil
	}

	// Load server certificate
	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   cfg.MinVersion,
	}

	// Load client CA if mTLS is enabled
	if cfg.RequireClientCert && cfg.ClientCA != "" {
		caCert, err := os.ReadFile(cfg.ClientCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read client CA: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse client CA certificate")
		}

		tlsConfig.ClientCAs = caPool
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}

	log.Info().
		Str("cert", cfg.CertFile).
		Bool("mtls", cfg.RequireClientCert).
		Str("min_version", cfg.MinVersionName()).
		Msg("Loaded server TLS configuration")

	return tlsConfig, nil
}

// LoadClientTLS creates a TLS configuration for the client.
func LoadClientTLS(cfg Config) (*tls.Config, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	if !cfg.Enabled {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		MinVersion:         cfg.MinVersion,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
	}

	// Load client certificate if mTLS is enabled
	if cfg.CertFile != "" && cfg.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Load CA certificate to verify server
	if cfg.ClientCA != "" {
		caCert, err := os.ReadFile(cfg.ClientCA)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA certificate: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caPool
	}

	log.Debug().
		Bool("mtls", cfg.CertFile != "").
		Bool("skip_verify", cfg.InsecureSkipVerify).
		Msg("Loaded client TLS configuration")

	return tlsConfig, nil
}

// ServerCredentials returns gRPC server credentials from the config.
func ServerCredentials(cfg Config) (credentials.TransportCredentials, error) {
	tlsConfig, err := LoadServerTLS(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig == nil {
		return nil, nil // TLS not enabled
	}
	return credentials.NewTLS(tlsConfig), nil
}

// ClientCredentials returns gRPC client credentials from the config.
func ClientCredentials(cfg Config) (credentials.TransportCredentials, error) {
	tlsConfig, err := LoadClientTLS(cfg)
	if err != nil {
		return nil, err
	}
	if tlsConfig == nil {
		return nil, nil // TLS not enabled
	}
	return credentials.NewTLS(tlsConfig), nil
}

// MustLoadServerTLS loads server TLS config, panics on error.
func MustLoadServerTLS(cfg Config) *tls.Config {
	tlsConfig, err := LoadServerTLS(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to load server TLS: %v", err))
	}
	return tlsConfig
}

// MustLoadClientTLS loads client TLS config, panics on error.
func MustLoadClientTLS(cfg Config) *tls.Config {
	tlsConfig, err := LoadClientTLS(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to load client TLS: %v", err))
	}
	return tlsConfig
}
