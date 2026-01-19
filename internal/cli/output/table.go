package output

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/olekukonko/tablewriter"
)

// Table wraps tablewriter with build-specific functionality.
type Table struct {
	table *tablewriter.Table
}

// TableConfig holds table configuration options.
type TableConfig struct {
	Writer   io.Writer
	NoHeader bool
	Center   bool
}

// NewTable creates a new table with the given headers.
func NewTable(headers []string) *Table {
	return NewTableWithConfig(headers, TableConfig{})
}

// NewTableWithConfig creates a table with custom configuration.
func NewTableWithConfig(headers []string, cfg TableConfig) *Table {
	writer := cfg.Writer
	if writer == nil {
		writer = os.Stdout
	}

	t := tablewriter.NewWriter(writer)

	if !cfg.NoHeader && len(headers) > 0 {
		t.SetHeader(headers)
	}

	// Default styling
	t.SetBorder(false)
	t.SetHeaderLine(true)
	t.SetColumnSeparator(" ")
	t.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	t.SetAlignment(tablewriter.ALIGN_LEFT)
	t.SetAutoWrapText(false)
	t.SetAutoFormatHeaders(false)

	if cfg.Center {
		t.SetAlignment(tablewriter.ALIGN_CENTER)
	}

	return &Table{table: t}
}

// Append adds a row to the table.
func (t *Table) Append(row []string) {
	t.table.Append(row)
}

// AppendBulk adds multiple rows to the table.
func (t *Table) AppendBulk(rows [][]string) {
	t.table.AppendBulk(rows)
}

// Render outputs the table.
func (t *Table) Render() {
	t.table.Render()
}

// SetColWidth sets the column width for a specific column.
func (t *Table) SetColWidth(width int) {
	t.table.SetColWidth(width)
}

// BuildStats holds build statistics for the summary table.
type BuildStats struct {
	Total       int
	Remote      int
	CacheHits   int
	Local       int
	Failed      int
	Duration    time.Duration
	TimeSaved   time.Duration
	BytesSaved  int64
	TasksFailed []string
}

// PrintBuildSummary prints a colored build summary table.
func PrintBuildSummary(stats BuildStats) {
	fmt.Println()
	fmt.Println(Bold("Build Summary"))
	fmt.Println("─────────────")

	table := NewTable([]string{"Metric", "Value"})
	table.table.SetBorder(false)

	// Add rows with colors
	table.Append([]string{"Total Files", fmt.Sprintf("%d", stats.Total)})

	if stats.Remote > 0 {
		table.Append([]string{"Remote", Success(fmt.Sprintf("%d", stats.Remote))})
	}

	if stats.CacheHits > 0 {
		table.Append([]string{"Cache Hits", Success(fmt.Sprintf("%d", stats.CacheHits))})
	}

	if stats.Local > 0 {
		table.Append([]string{"Local Fallback", Warning(fmt.Sprintf("%d", stats.Local))})
	}

	if stats.Failed > 0 {
		table.Append([]string{"Failed", Error(fmt.Sprintf("%d", stats.Failed))})
	}

	table.Append([]string{"Duration", fmt.Sprintf("%.2fs", stats.Duration.Seconds())})

	if stats.TimeSaved > 0 {
		table.Append([]string{"Time Saved", Success(fmt.Sprintf("%.1fs", stats.TimeSaved.Seconds()))})
	}

	table.Render()

	// Print failed tasks if any
	if len(stats.TasksFailed) > 0 && len(stats.TasksFailed) <= 5 {
		fmt.Println()
		fmt.Println(Error("Failed files:"))
		for _, f := range stats.TasksFailed {
			fmt.Printf("  • %s\n", f)
		}
	}

	fmt.Println()
}

// WorkerInfo holds worker information for the workers table.
type WorkerInfo struct {
	ID           string
	Arch         string
	Cores        int
	MemoryGB     float64
	ActiveTasks  int
	Status       string
	CircuitState string
}

// PrintWorkersTable prints a colored workers table.
func PrintWorkersTable(workers []WorkerInfo, totalWorkers, healthyWorkers int) {
	if len(workers) == 0 {
		fmt.Println(Warning("No workers connected"))
		return
	}

	fmt.Printf("Workers: %s total, %s healthy\n\n",
		Bold(fmt.Sprintf("%d", totalWorkers)),
		Success(fmt.Sprintf("%d", healthyWorkers)))

	table := NewTable([]string{"ID", "ARCH", "CORES", "MEMORY", "TASKS", "STATUS"})

	for _, w := range workers {
		status := WorkerStatus(w.CircuitState)

		table.Append([]string{
			truncateString(w.ID, 20),
			w.Arch,
			fmt.Sprintf("%d", w.Cores),
			fmt.Sprintf("%.1f GB", w.MemoryGB),
			fmt.Sprintf("%d", w.ActiveTasks),
			status,
		})
	}

	table.Render()
}

// PrintWorkersTableCompact prints a compact workers table (no memory/tasks columns).
func PrintWorkersTableCompact(workers []WorkerInfo, totalWorkers, healthyWorkers int) {
	if len(workers) == 0 {
		fmt.Println(Warning("No workers connected"))
		return
	}

	fmt.Printf("Workers: %s total, %s healthy\n\n",
		Bold(fmt.Sprintf("%d", totalWorkers)),
		Success(fmt.Sprintf("%d", healthyWorkers)))

	table := NewTable([]string{"ID", "ARCH", "CORES", "STATUS"})

	for _, w := range workers {
		status := WorkerStatus(w.CircuitState)

		table.Append([]string{
			truncateString(w.ID, 20),
			w.Arch,
			fmt.Sprintf("%d", w.Cores),
			status,
		})
	}

	table.Render()
}

// StatusInfo holds coordinator status information.
type StatusInfo struct {
	Address     string
	Healthy     bool
	ActiveTasks int
	QueuedTasks int
	Workers     int
	CacheHits   int64
	Uptime      time.Duration
}

// PrintStatus prints a colored status summary.
func PrintStatus(status StatusInfo) {
	fmt.Println(Bold("Coordinator Status"))
	fmt.Println("──────────────────")

	table := NewTable([]string{})
	table.table.SetHeader(nil)

	table.Append([]string{"Address:", Info(status.Address)})
	table.Append([]string{"Status:", Healthy(status.Healthy)})
	table.Append([]string{"Active Tasks:", fmt.Sprintf("%d", status.ActiveTasks)})
	table.Append([]string{"Queued Tasks:", fmt.Sprintf("%d", status.QueuedTasks)})

	if status.Workers > 0 {
		table.Append([]string{"Workers:", fmt.Sprintf("%d", status.Workers)})
	}

	if status.CacheHits > 0 {
		table.Append([]string{"Cache Hits:", Success(fmt.Sprintf("%d", status.CacheHits))})
	}

	if status.Uptime > 0 {
		table.Append([]string{"Uptime:", formatDuration(status.Uptime)})
	}

	table.Render()
}

// CacheStats holds cache statistics.
type CacheStats struct {
	Directory  string
	Entries    int
	TotalSize  int64
	MaxSize    int64
	TotalHits  int64
	TotalMiss  int64
	HitRate    float64
	OldestFile time.Time
	NewestFile time.Time
}

// PrintCacheStats prints a colored cache statistics table.
func PrintCacheStats(stats CacheStats) {
	fmt.Println(Bold("Cache Statistics"))
	fmt.Println("────────────────")

	table := NewTable([]string{})
	table.table.SetHeader(nil)

	table.Append([]string{"Directory:", Info(stats.Directory)})
	table.Append([]string{"Entries:", fmt.Sprintf("%d", stats.Entries)})

	// Size with percentage
	usedPercent := float64(stats.TotalSize) / float64(stats.MaxSize) * 100
	sizeStr := fmt.Sprintf("%.2f MB / %.0f MB (%.1f%%)",
		float64(stats.TotalSize)/(1024*1024),
		float64(stats.MaxSize)/(1024*1024),
		usedPercent)
	if usedPercent > 90 {
		sizeStr = Warning(sizeStr)
	}
	table.Append([]string{"Size:", sizeStr})

	table.Append([]string{"Total Hits:", Success(fmt.Sprintf("%d", stats.TotalHits))})

	if stats.TotalMiss > 0 {
		table.Append([]string{"Total Miss:", fmt.Sprintf("%d", stats.TotalMiss)})
	}

	if stats.TotalHits+stats.TotalMiss > 0 {
		hitRate := float64(stats.TotalHits) / float64(stats.TotalHits+stats.TotalMiss) * 100
		hitRateStr := fmt.Sprintf("%.1f%%", hitRate)
		if hitRate >= 80 {
			hitRateStr = Success(hitRateStr)
		} else if hitRate >= 50 {
			hitRateStr = Warning(hitRateStr)
		} else {
			hitRateStr = Error(hitRateStr)
		}
		table.Append([]string{"Hit Rate:", hitRateStr})
	}

	table.Render()
}

// truncateString truncates a string to the given max length.
func truncateString(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	} else if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm%ds", mins, secs)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh%dm", hours, mins)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd%dh", days, hours)
}
