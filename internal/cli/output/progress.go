package output

import (
	"io"
	"os"

	"github.com/schollz/progressbar/v3"
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

// NewBuildProgress creates a progress bar for build operations.
func NewBuildProgress(total int) *ProgressBar {
	return NewProgress(ProgressConfig{
		Total:       total,
		Description: "Compiling",
	})
}

// NewProgress creates a configurable progress bar.
func NewProgress(cfg ProgressConfig) *ProgressBar {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stderr
	}

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
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(100),
		progressbar.OptionUseANSICodes(true),
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

// CompileProgress creates a specialized progress bar for compilation.
func CompileProgress(total int, description string) *ProgressBar {
	bar := progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionOnCompletion(func() {
			_, _ = io.WriteString(os.Stderr, "\n")
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
	)

	return &ProgressBar{bar: bar}
}

// SpinnerProgress creates a spinner for indeterminate progress.
func SpinnerProgress(description string) *ProgressBar {
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionEnableColorCodes(true),
	)

	return &ProgressBar{bar: bar}
}

// DownloadProgress creates a progress bar for file downloads.
func DownloadProgress(totalBytes int64, description string) *ProgressBar {
	bar := progressbar.NewOptions64(totalBytes,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionOnCompletion(func() {
			_, _ = io.WriteString(os.Stderr, "\n")
		}),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        Info("█"),
			SaucerHead:    "▓",
			SaucerPadding: "░",
			BarStart:      "[",
			BarEnd:        "]",
		}),
		progressbar.OptionSetWidth(35),
	)

	return &ProgressBar{bar: bar}
}
