package logging

import (
	"io"
	"os"
	"sync"

	"github.com/rs/zerolog"

	"github.com/h3nr1-d14z/hybridgrid/internal/config"
)

type syncedFileWriter struct {
	file *os.File
	mu   sync.Mutex
}

func (w *syncedFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	n, err := w.file.Write(p)
	if err == nil {
		w.file.Sync()
	}
	return n, err
}

func (w *syncedFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// SetupFileWriter opens a file for appending and returns it as an io.WriteCloser.
// If the file doesn't exist, it creates it with mode 0644.
func SetupFileWriter(filePath string) (io.WriteCloser, error) {
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	return &syncedFileWriter{file: file}, nil
}

// SetupLogger configures and returns a zerolog.Logger based on the LogConfig.
// It also returns an io.Closer for the underlying file writer (if any), which the caller
// should defer close() to ensure buffered writes are flushed.
//
// Behavior:
// - If cfg.File is empty: uses ConsoleWriter to os.Stderr (current behavior preserved)
// - If cfg.File is set: writes JSON logs to the file
// - If cfg.Format == "json": no ConsoleWriter wrapper, raw JSON output
// - If cfg.Format == "console": uses ConsoleWriter (default)
// - Supports multi-output (console + file) using MultiLevelWriter
// - Sets log level from cfg.Level using zerolog.ParseLevel()
func SetupLogger(cfg config.LogConfig) (zerolog.Logger, io.Closer, error) {
	// Parse log level
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		// Default to info if parsing fails
		level = zerolog.InfoLevel
	}

	var output io.Writer
	var closer io.Closer

	// Determine output writers
	if cfg.File == "" {
		// No file: use console writer to stderr (preserve current behavior)
		output = zerolog.ConsoleWriter{Out: os.Stderr}
		closer = noOpCloser{}
	} else {
		// File is set: open it
		fileWriter, err := SetupFileWriter(cfg.File)
		if err != nil {
			return zerolog.Logger{}, nil, err
		}

		// Check format preference
		if cfg.Format == "json" {
			// JSON-only output to file (no console wrapper)
			output = fileWriter
			closer = fileWriter
		} else {
			// Default to console format: multi-output (console + file)
			consoleWriter := zerolog.ConsoleWriter{Out: os.Stderr}
			output = zerolog.MultiLevelWriter(consoleWriter, fileWriter)
			closer = fileWriter
		}
	}

	// Create and configure logger
	logger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger, closer, nil
}

// noOpCloser is a no-op closer for when no file is used
type noOpCloser struct{}

func (noOpCloser) Close() error {
	return nil
}
