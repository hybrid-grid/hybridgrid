package logging

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rs/zerolog"

	"github.com/h3nr1-d14z/hybridgrid/internal/config"
)

func TestNewRotatingWriter_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogRotationConfig{
		MaxSizeMB:  100,
		MaxBackups: 3,
		MaxAgeDays: 28,
		Compress:   true,
	}

	writer := NewRotatingWriter(cfg, filePath)
	defer writer.Close()

	// lumberjack creates file on first write, not on NewRotatingWriter()
	if _, err := io.WriteString(writer, "test\n"); err != nil {
		t.Fatalf("failed to write: %v", err)
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("rotating writer did not create file after write")
	}
}

func TestNewRotatingWriter_WritesSuccessfully(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogRotationConfig{
		MaxSizeMB:  100,
		MaxBackups: 3,
		MaxAgeDays: 28,
		Compress:   true,
	}

	writer := NewRotatingWriter(cfg, filePath)
	defer writer.Close()

	n, err := io.WriteString(writer, "test log entry\n")
	if err != nil {
		t.Fatalf("failed to write to rotating writer: %v", err)
	}
	if n == 0 {
		t.Fatal("no bytes written")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if string(content) != "test log entry\n" {
		t.Fatalf("expected 'test log entry\\n', got %q", string(content))
	}
}

func TestNewRotatingWriter_ConfigPassthrough(t *testing.T) {
	tests := []struct {
		name string
		cfg  config.LogRotationConfig
	}{
		{
			name: "default config",
			cfg: config.LogRotationConfig{
				MaxSizeMB:  100,
				MaxBackups: 3,
				MaxAgeDays: 28,
				Compress:   true,
			},
		},
		{
			name: "custom config",
			cfg: config.LogRotationConfig{
				MaxSizeMB:  50,
				MaxBackups: 5,
				MaxAgeDays: 14,
				Compress:   false,
			},
		},
		{
			name: "minimal config",
			cfg: config.LogRotationConfig{
				MaxSizeMB:  10,
				MaxBackups: 1,
				MaxAgeDays: 7,
				Compress:   false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			filePath := filepath.Join(tmpDir, "test.log")

			writer := NewRotatingWriter(tt.cfg, filePath)
			defer writer.Close()

			if _, err := io.WriteString(writer, "test\n"); err != nil {
				t.Fatalf("failed to write: %v", err)
			}

			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Fatal("file was not created")
			}
		})
	}
}

func TestNewRotatingWriter_ImplementsWriteCloser(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogRotationConfig{
		MaxSizeMB:  100,
		MaxBackups: 3,
		MaxAgeDays: 28,
		Compress:   true,
	}

	var w io.WriteCloser = NewRotatingWriter(cfg, filePath)
	defer w.Close()

	if w == nil {
		t.Fatal("NewRotatingWriter returned nil")
	}
}

func TestSetupLogger_WithRotation_CreatesBackup(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "rotate.log")

	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
		File:   filePath,
		Rotation: config.LogRotationConfig{
			MaxSizeMB:  1,
			MaxBackups: 2,
			MaxAgeDays: 1,
		},
	}

	logger, closer, err := SetupLogger(cfg)
	if err != nil {
		t.Fatalf("SetupLogger() error = %v", err)
	}

	payload := strings.Repeat("x", 16*1024)
	for i := 0; i < 96; i++ {
		logger.Info().Str("payload", payload).Int("index", i).Msg("rotation-test")
	}

	if err := closer.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	if len(entries) < 2 {
		t.Fatalf("expected rotated backup files, found %d entries", len(entries))
	}
}
