package logging

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"

	"github.com/h3nr1-d14z/hybridgrid/internal/config"
)

func TestSetupFileWriter_CreatesFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	writer, err := SetupFileWriter(filePath)
	if err != nil {
		t.Fatalf("SetupFileWriter failed: %v", err)
	}
	defer writer.Close()

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatal("file was not created")
	}

	if _, err := io.WriteString(writer, "test log\n"); err != nil {
		t.Fatalf("failed to write to file: %v", err)
	}
}

func TestSetupFileWriter_AppendsToExisting(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	writer1, err := SetupFileWriter(filePath)
	if err != nil {
		t.Fatalf("first SetupFileWriter failed: %v", err)
	}
	io.WriteString(writer1, "line1\n")
	writer1.Close()

	writer2, err := SetupFileWriter(filePath)
	if err != nil {
		t.Fatalf("second SetupFileWriter failed: %v", err)
	}
	io.WriteString(writer2, "line2\n")
	writer2.Close()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	contentStr := string(content)
	if contentStr != "line1\nline2\n" {
		t.Fatalf("expected 'line1\\nline2\\n', got %q", contentStr)
	}
}

func TestSetupLogger_NoFile_DefaultsToConsole(t *testing.T) {
	cfg := config.LogConfig{
		Level:  "info",
		Format: "console",
		File:   "",
	}

	logger, closer, err := SetupLogger(cfg)
	if err != nil {
		t.Fatalf("SetupLogger failed: %v", err)
	}
	defer closer.Close()

	if logger.GetLevel() != zerolog.InfoLevel {
		t.Fatalf("expected level InfoLevel, got %v", logger.GetLevel())
	}
}

func TestSetupLogger_WithFile_WritesJSON(t *testing.T) {
	// Ensure global level is set to allow info-level messages
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogConfig{
		Level:  "info",
		Format: "json",
		File:   filePath,
	}

	logger, closer, err := SetupLogger(cfg)
	if err != nil {
		t.Fatalf("SetupLogger failed: %v", err)
	}
	defer closer.Close()

	logger.Info().Str("test", "value").Msg("test message")
	logger.Debug().Str("debug", "nope").Msg("should not appear")

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	contentStr := string(content)
	if len(contentStr) == 0 {
		t.Fatal("log file is empty")
	}

	if !contains(contentStr, "{") || !contains(contentStr, "}") {
		t.Fatalf("log output doesn't look like JSON: %s", contentStr)
	}

	if !contains(contentStr, "test") || !contains(contentStr, "value") {
		t.Fatalf("log doesn't contain expected fields: %s", contentStr)
	}

	if contains(contentStr, "debug") {
		t.Fatalf("debug message should not appear at info level: %s", contentStr)
	}
}

func TestSetupLogger_LevelParsing(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		expected  zerolog.Level
		shouldErr bool
	}{
		{"debug", "debug", zerolog.DebugLevel, false},
		{"info", "info", zerolog.InfoLevel, false},
		{"warn", "warn", zerolog.WarnLevel, false},
		{"error", "error", zerolog.ErrorLevel, false},
		{"invalid defaults to info", "invalid", zerolog.InfoLevel, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.LogConfig{
				Level:  tt.level,
				Format: "console",
				File:   "",
			}

			logger, closer, err := SetupLogger(cfg)
			defer closer.Close()

			if tt.shouldErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !tt.shouldErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if logger.GetLevel() != tt.expected {
				t.Fatalf("expected level %v, got %v", tt.expected, logger.GetLevel())
			}
		})
	}
}

func TestSetupLogger_MultiOutput(t *testing.T) {
	// Ensure global level is set to allow info-level messages
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.log")

	cfg := config.LogConfig{
		Level:  "info",
		Format: "console",
		File:   filePath,
	}

	logger, closer, err := SetupLogger(cfg)
	if err != nil {
		t.Fatalf("SetupLogger failed: %v", err)
	}
	defer closer.Close()

	logger.Info().Str("test", "multioutput").Msg("test message")

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if len(content) == 0 {
		t.Fatal("log file is empty despite multi-output")
	}
}

func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
