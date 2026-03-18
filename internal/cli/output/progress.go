package output

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
)

// ProgressBar wraps a progress bar with build-specific functionality.
type ProgressBar struct {
	bar     *progressbar.ProgressBar
	verbose bool
}

// ProgressConfig holds progress bar configuration.
type ProgressConfig struct {
	Total       int
	Description string
	Verbose     bool
	Writer      io.Writer
}

// isTerminalWriter returns true when w is a real TTY.
// Only os.File descriptors can be checked; all other writers (bytes.Buffer,
// pipes, test writers) are treated as non-terminal.
func isTerminalWriter(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// NewBuildProgress creates a progress bar for build operations writing to stderr.
func NewBuildProgress(total int) *ProgressBar {
	return NewProgress(ProgressConfig{
		Total:       total,
		Description: "Compiling",
	})
}

// NewBuildProgressWithWriter creates a progress bar for build operations writing to w.
func NewBuildProgressWithWriter(total int, w io.Writer) *ProgressBar {
	return NewProgress(ProgressConfig{
		Total:       total,
		Description: "Compiling",
		Writer:      w,
	})
}

// NewProgress creates a configurable progress bar.
func NewProgress(cfg ProgressConfig) *ProgressBar {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

	useANSI := isTerminalWriter(writer)

	bar := progressbar.NewOptions(cfg.Total,
		progressbar.OptionSetDescription(cfg.Description),
		progressbar.OptionSetWriter(writer),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetItsString("files"),
		progressbar.OptionOnCompletion(func() {
			_, _ = io.WriteString(writer, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "█",
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionEnableColorCodes(useANSI),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100),
		progressbar.OptionUseANSICodes(useANSI),
	)

	return &ProgressBar{
		bar:     bar,
		verbose: cfg.Verbose,
	}
}

// Add increments the progress bar by the given amount.
func (p *ProgressBar) Add(n int) error {
	return p.bar.Add(n)
}

// Increment increments the progress bar by 1.
func (p *ProgressBar) Increment() error {
	return p.bar.Add(1)
}

// SetDescription updates the progress bar description.
func (p *ProgressBar) SetDescription(desc string) {
	p.bar.Describe(desc)
}

// Finish completes the progress bar.
func (p *ProgressBar) Finish() error {
	return p.bar.Finish()
}

// Clear removes the progress bar from display.
func (p *ProgressBar) Clear() error {
	return p.bar.Clear()
}

// IsVerbose returns whether verbose mode is enabled.
func (p *ProgressBar) IsVerbose() bool {
	return p.verbose
}

// CompileProgress creates a specialized progress bar for compilation writing to stderr.
func CompileProgress(total int, description string) *ProgressBar {
	return CompileProgressWithWriter(total, description, os.Stderr)
}

// CompileProgressWithWriter creates a specialized progress bar for compilation writing to w.
func CompileProgressWithWriter(total int, description string, w io.Writer) *ProgressBar {
	useANSI := isTerminalWriter(w)
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(w),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			_, _ = io.WriteString(w, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        Success("█"),
			SaucerHead:    Info("▓"),
			SaucerPadding: Dim("░"),
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetWidth(35),
		progressbar.OptionThrottle(100),
		progressbar.OptionUseANSICodes(useANSI),
	)

	return &ProgressBar{bar: bar}
}

// SpinnerProgress creates a spinner for indeterminate progress writing to stderr.
func SpinnerProgress(description string) *ProgressBar {
	return SpinnerProgressWithWriter(description, os.Stderr)
}

// SpinnerProgressWithWriter creates a spinner for indeterminate progress writing to w.
func SpinnerProgressWithWriter(description string, w io.Writer) *ProgressBar {
	useANSI := isTerminalWriter(w)
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(w),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(useANSI),
		progressbar.OptionUseANSICodes(useANSI),
	)

	return &ProgressBar{bar: bar}
}

// DownloadProgress creates a progress bar for file downloads writing to stderr.
func DownloadProgress(totalBytes int64, description string) *ProgressBar {
	return DownloadProgressWithWriter(totalBytes, description, os.Stderr)
}

// DownloadProgressWithWriter creates a progress bar for file downloads writing to w.
func DownloadProgressWithWriter(totalBytes int64, description string, w io.Writer) *ProgressBar {
	useANSI := isTerminalWriter(w)
	bar := progressbar.NewOptions64(totalBytes,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(w),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionOnCompletion(func() {
			_, _ = io.WriteString(w, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        Info("█"),
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetWidth(35),
		progressbar.OptionUseANSICodes(useANSI),
	)

	return &ProgressBar{bar: bar}
}
