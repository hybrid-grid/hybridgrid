package tls

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// generateTestCert creates a self-signed certificate for testing.
func generateTestCert(dir string) (certFile, keyFile string, err error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", "", err
	}

	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")

	certOut, err := os.Create(certFile)
	if err != nil {
		return "", "", err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	certOut.Close()

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return "", "", err
	}
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	keyOut.Close()

	return certFile, keyFile, nil
}

func TestConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Enabled {
		t.Error("Default config should have TLS disabled")
	}
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Error("Default min version should be TLS 1.2")
	}
	if cfg.InsecureSkipVerify {
		t.Error("InsecureSkipVerify should be false by default")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name:    "disabled is valid",
			cfg:     Config{Enabled: false},
			wantErr: false,
		},
		{
			name: "enabled without cert fails",
			cfg: Config{
				Enabled: true,
			},
			wantErr: true,
		},
		{
			name: "enabled without key fails",
			cfg: Config{
				Enabled:  true,
				CertFile: "/path/to/cert.pem",
			},
			wantErr: true,
		},
		{
			name: "mtls without client CA fails",
			cfg: Config{
				Enabled:           true,
				CertFile:          "/path/to/cert.pem",
				KeyFile:           "/path/to/key.pem",
				RequireClientCert: true,
			},
			wantErr: true,
		},
		{
			name: "complete config is valid",
			cfg: Config{
				Enabled:    true,
				CertFile:   "/path/to/cert.pem",
				KeyFile:    "/path/to/key.pem",
				MinVersion: tls.VersionTLS12,
			},
			wantErr: false,
		},
		{
			name: "mtls with client CA is valid",
			cfg: Config{
				Enabled:           true,
				CertFile:          "/path/to/cert.pem",
				KeyFile:           "/path/to/key.pem",
				ClientCA:          "/path/to/ca.pem",
				RequireClientCert: true,
				MinVersion:        tls.VersionTLS12,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_IsEnabled(t *testing.T) {
	cfg := Config{Enabled: true}
	if !cfg.IsEnabled() {
		t.Error("IsEnabled should return true")
	}

	cfg.Enabled = false
	if cfg.IsEnabled() {
		t.Error("IsEnabled should return false")
	}
}

func TestConfig_IsMTLS(t *testing.T) {
	cfg := Config{Enabled: true, RequireClientCert: true}
	if !cfg.IsMTLS() {
		t.Error("IsMTLS should return true")
	}

	cfg.RequireClientCert = false
	if cfg.IsMTLS() {
		t.Error("IsMTLS should return false")
	}

	cfg.Enabled = false
	cfg.RequireClientCert = true
	if cfg.IsMTLS() {
		t.Error("IsMTLS should return false when TLS disabled")
	}
}

func TestConfig_MinVersionName(t *testing.T) {
	tests := []struct {
		version uint16
		want    string
	}{
		{tls.VersionTLS10, "TLS 1.0"},
		{tls.VersionTLS11, "TLS 1.1"},
		{tls.VersionTLS12, "TLS 1.2"},
		{tls.VersionTLS13, "TLS 1.3"},
		{0x0999, "unknown (0x0999)"},
	}

	for _, tt := range tests {
		cfg := Config{MinVersion: tt.version}
		if got := cfg.MinVersionName(); got != tt.want {
			t.Errorf("MinVersionName(%d) = %s, want %s", tt.version, got, tt.want)
		}
	}
}

func TestLoadServerTLS_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	tlsConfig, err := LoadServerTLS(cfg)
	if err != nil {
		t.Fatalf("LoadServerTLS failed: %v", err)
	}
	if tlsConfig != nil {
		t.Error("Expected nil TLS config when disabled")
	}
}

func TestLoadServerTLS_Success(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile, err := generateTestCert(dir)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	cfg := Config{
		Enabled:    true,
		CertFile:   certFile,
		KeyFile:    keyFile,
		MinVersion: tls.VersionTLS12,
	}

	tlsConfig, err := LoadServerTLS(cfg)
	if err != nil {
		t.Fatalf("LoadServerTLS failed: %v", err)
	}
	if tlsConfig == nil {
		t.Fatal("Expected TLS config, got nil")
	}
	if len(tlsConfig.Certificates) != 1 {
		t.Errorf("Expected 1 certificate, got %d", len(tlsConfig.Certificates))
	}
	if tlsConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %d, want %d", tlsConfig.MinVersion, tls.VersionTLS12)
	}
}

func TestLoadServerTLS_InvalidCert(t *testing.T) {
	cfg := Config{
		Enabled:  true,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	}

	_, err := LoadServerTLS(cfg)
	if err == nil {
		t.Error("Expected error for invalid cert path")
	}
}

func TestLoadServerTLS_MTLS(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile, err := generateTestCert(dir)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	cfg := Config{
		Enabled:           true,
		CertFile:          certFile,
		KeyFile:           keyFile,
		ClientCA:          certFile, // Use same cert as CA for testing
		RequireClientCert: true,
		MinVersion:        tls.VersionTLS12,
	}

	tlsConfig, err := LoadServerTLS(cfg)
	if err != nil {
		t.Fatalf("LoadServerTLS failed: %v", err)
	}
	if tlsConfig.ClientAuth != tls.RequireAndVerifyClientCert {
		t.Error("Expected RequireAndVerifyClientCert for mTLS")
	}
	if tlsConfig.ClientCAs == nil {
		t.Error("Expected ClientCAs to be set for mTLS")
	}
}

func TestLoadClientTLS_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	tlsConfig, err := LoadClientTLS(cfg)
	if err != nil {
		t.Fatalf("LoadClientTLS failed: %v", err)
	}
	if tlsConfig != nil {
		t.Error("Expected nil TLS config when disabled")
	}
}

func TestLoadClientTLS_Success(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile, err := generateTestCert(dir)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	cfg := Config{
		Enabled:    true,
		CertFile:   certFile,
		KeyFile:    keyFile,
		ClientCA:   certFile, // Use same cert as CA
		MinVersion: tls.VersionTLS12,
	}

	tlsConfig, err := LoadClientTLS(cfg)
	if err != nil {
		t.Fatalf("LoadClientTLS failed: %v", err)
	}
	if tlsConfig == nil {
		t.Fatal("Expected TLS config, got nil")
	}
	if tlsConfig.RootCAs == nil {
		t.Error("Expected RootCAs to be set")
	}
}

func TestLoadClientTLS_InsecureSkipVerify(t *testing.T) {
	cfg := Config{
		Enabled:            true,
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	tlsConfig, err := LoadClientTLS(cfg)
	if err != nil {
		t.Fatalf("LoadClientTLS failed: %v", err)
	}
	if !tlsConfig.InsecureSkipVerify {
		t.Error("Expected InsecureSkipVerify to be true")
	}
}

func TestServerCredentials(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile, err := generateTestCert(dir)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	cfg := Config{
		Enabled:    true,
		CertFile:   certFile,
		KeyFile:    keyFile,
		MinVersion: tls.VersionTLS12,
	}

	creds, err := ServerCredentials(cfg)
	if err != nil {
		t.Fatalf("ServerCredentials failed: %v", err)
	}
	if creds == nil {
		t.Error("Expected credentials, got nil")
	}
}

func TestServerCredentials_Disabled(t *testing.T) {
	cfg := Config{Enabled: false}
	creds, err := ServerCredentials(cfg)
	if err != nil {
		t.Fatalf("ServerCredentials failed: %v", err)
	}
	if creds != nil {
		t.Error("Expected nil credentials when disabled")
	}
}

func TestClientCredentials(t *testing.T) {
	dir := t.TempDir()
	certFile, keyFile, err := generateTestCert(dir)
	if err != nil {
		t.Fatalf("Failed to generate test cert: %v", err)
	}

	cfg := Config{
		Enabled:    true,
		CertFile:   certFile,
		KeyFile:    keyFile,
		ClientCA:   certFile,
		MinVersion: tls.VersionTLS12,
	}

	creds, err := ClientCredentials(cfg)
	if err != nil {
		t.Fatalf("ClientCredentials failed: %v", err)
	}
	if creds == nil {
		t.Error("Expected credentials, got nil")
	}
}
