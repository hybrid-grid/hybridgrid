package output

import (
	"fmt"
	"os"

	"github.com/fatih/color"
)

var (
	// Color functions for different status levels
	Success = color.New(color.FgGreen).SprintFunc()
	Error   = color.New(color.FgRed).SprintFunc()
	Warning = color.New(color.FgYellow).SprintFunc()
	Info    = color.New(color.FgCyan).SprintFunc()
	Bold    = color.New(color.Bold).SprintFunc()
	Dim     = color.New(color.Faint).SprintFunc()

	// Styled text formatters
	SuccessBold = color.New(color.FgGreen, color.Bold).SprintFunc()
	ErrorBold   = color.New(color.FgRed, color.Bold).SprintFunc()
	WarningBold = color.New(color.FgYellow, color.Bold).SprintFunc()
	InfoBold    = color.New(color.FgCyan, color.Bold).SprintFunc()
)

// DisableColors disables color output (for non-TTY environments).
func DisableColors() {
	color.NoColor = true
}

// EnableColors enables color output.
func EnableColors() {
	color.NoColor = false
}

// AutoDetectColors enables/disables colors based on terminal capability.
func AutoDetectColors() {
	if !isTerminal() {
		DisableColors()
	}
}

// isTerminal checks if stdout is connected to a terminal.
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// StatusLabel returns a colored status label.
func StatusLabel(status string) string {
	switch status {
	case "remote":
		return Info("[remote]")
	case "cache":
		return Success("[cache]")
	case "local":
		return Warning("[local]")
	case "failed":
		return Error("[failed]")
	case "healthy":
		return Success("[healthy]")
	case "unhealthy":
		return Error("[unhealthy]")
	case "pending":
		return Dim("[pending]")
	case "running":
		return Info("[running]")
	case "success":
		return Success("[success]")
	default:
		return "[" + status + "]"
	}
}

// StatusIcon returns a colored status icon.
func StatusIcon(ok bool) string {
	if ok {
		return Success("✓")
	}
	return Error("✗")
}

// Healthy returns a colored health status.
func Healthy(healthy bool) string {
	if healthy {
		return Success("healthy ✓")
	}
	return Error("unhealthy ✗")
}

// WorkerStatus returns a colored worker status.
func WorkerStatus(circuitState string) string {
	if circuitState == "" || circuitState == "CLOSED" {
		return Success("healthy")
	}
	switch circuitState {
	case "OPEN":
		return Error(circuitState)
	case "HALF_OPEN":
		return Warning(circuitState)
	default:
		return Warning(circuitState)
	}
}

// Percent returns a colored percentage (green if > 80, yellow if > 50, red otherwise).
func Percent(value float64) string {
	formatted := fmt.Sprintf("%.1f%%", value)
	if value >= 80 {
		return Success(formatted)
	} else if value >= 50 {
		return Warning(formatted)
	}
	return Error(formatted)
}

// ByteSize returns a formatted byte size.
func ByteSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return Bold(fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB)))
	case bytes >= MB:
		return Bold(fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB)))
	case bytes >= KB:
		return Bold(fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB)))
	default:
		return Bold(fmt.Sprintf("%d B", bytes))
	}
}
