package logging

import (
	"io"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/h3nr1-d14z/hybridgrid/internal/config"
)

// NewRotatingWriter creates a rotating file writer using lumberjack.Logger.
// It returns an io.WriteCloser that handles log rotation based on:
// - MaxSizeMB: maximum size of a log file before rotation (in megabytes)
// - MaxBackups: maximum number of backup files to keep
// - MaxAgeDays: maximum age of a log file in days before deletion
// - Compress: whether to compress rotated log files
func NewRotatingWriter(cfg config.LogRotationConfig, filePath string) io.WriteCloser {
	return &lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
		Compress:   cfg.Compress,
	}
}
