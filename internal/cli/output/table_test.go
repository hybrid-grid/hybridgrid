package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w
	defer func() {
		os.Stdout = original
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close stdout pipe failed: %v", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stdout reader failed: %v", err)
	}
	return string(data)
}

func TestNewTable(t *testing.T) {
	var buf bytes.Buffer
	table := NewTableWithConfig([]string{"Col1", "Col2"}, TableConfig{Writer: &buf})
	if table == nil {
		t.Fatal("NewTableWithConfig returned nil")
	}

	table.Append([]string{"val1", "val2"})
	table.Render()

	output := buf.String()
	if !strings.Contains(output, "Col1") {
		t.Errorf("Output missing header Col1: %s", output)
	}
	if !strings.Contains(output, "val1") {
		t.Errorf("Output missing value val1: %s", output)
	}
}

func TestNewTableNoHeader(t *testing.T) {
	var buf bytes.Buffer
	table := NewTableWithConfig([]string{}, TableConfig{Writer: &buf, NoHeader: true})
	if table == nil {
		t.Fatal("NewTableWithConfig returned nil")
	}

	table.Append([]string{"val1", "val2"})
	table.Render()

	output := buf.String()
	if !strings.Contains(output, "val1") {
		t.Errorf("Output missing value val1: %s", output)
	}
}

func TestNewTableDefaultWriter(t *testing.T) {
	// Test that it doesn't panic with nil writer
	table := NewTable([]string{"Col1"})
	if table == nil {
		t.Fatal("NewTable returned nil")
	}
}

func TestTableAppendBulk(t *testing.T) {
	var buf bytes.Buffer
	table := NewTableWithConfig([]string{"A", "B"}, TableConfig{Writer: &buf})

	table.AppendBulk([][]string{
		{"1", "2"},
		{"3", "4"},
	})
	table.Render()

	output := buf.String()
	if !strings.Contains(output, "1") || !strings.Contains(output, "4") {
		t.Errorf("Output missing bulk values: %s", output)
	}
}

func TestTableSetColWidth(t *testing.T) {
	var buf bytes.Buffer
	table := NewTableWithConfig([]string{"Col1"}, TableConfig{Writer: &buf})
	// Just verify it doesn't panic
	table.SetColWidth(20)
}

func TestBuildStats(t *testing.T) {
	stats := BuildStats{
		Total:      100,
		Remote:     60,
		Local:      30,
		CacheHits:  10,
		Failed:     0,
		Duration:   time.Minute,
		TimeSaved:  30 * time.Second,
		BytesSaved: 1024 * 1024,
	}

	// Verify struct works
	if stats.Total != 100 {
		t.Errorf("Expected Total=100, got %d", stats.Total)
	}
	if stats.Remote != 60 {
		t.Errorf("Expected Remote=60, got %d", stats.Remote)
	}
}

func TestBuildStatsWithFailures(t *testing.T) {
	stats := BuildStats{
		Total:       10,
		Remote:      5,
		Local:       2,
		CacheHits:   1,
		Failed:      2,
		Duration:    10 * time.Second,
		TasksFailed: []string{"file1.c", "file2.c"},
	}

	if stats.Failed != 2 {
		t.Errorf("Expected Failed=2, got %d", stats.Failed)
	}
	if len(stats.TasksFailed) != 2 {
		t.Errorf("Expected 2 failed tasks, got %d", len(stats.TasksFailed))
	}
}

func TestWorkerInfo(t *testing.T) {
	workers := []WorkerInfo{
		{
			ID:           "worker-1",
			Arch:         "amd64",
			Cores:        8,
			MemoryGB:     16.0,
			ActiveTasks:  2,
			Status:       "active",
			CircuitState: "CLOSED",
		},
		{
			ID:           "worker-2",
			Arch:         "arm64",
			Cores:        16,
			MemoryGB:     32.0,
			ActiveTasks:  0,
			Status:       "idle",
			CircuitState: "OPEN",
		},
	}

	for _, w := range workers {
		if w.ID == "" {
			t.Error("Worker ID should not be empty")
		}
		if w.Cores <= 0 {
			t.Errorf("Worker %s should have positive cores", w.ID)
		}
	}
}

func TestStatusInfo(t *testing.T) {
	status := StatusInfo{
		Address:     "localhost:50051",
		Healthy:     true,
		ActiveTasks: 10,
		QueuedTasks: 5,
		Workers:     3,
		CacheHits:   100,
		Uptime:      time.Hour,
	}

	if !status.Healthy {
		t.Error("Expected status to be healthy")
	}
	if status.Workers != 3 {
		t.Errorf("Expected Workers=3, got %d", status.Workers)
	}
}

func TestCacheStats(t *testing.T) {
	stats := CacheStats{
		Directory:  "/tmp/cache",
		Entries:    1000,
		TotalSize:  50 * 1024 * 1024,
		MaxSize:    100 * 1024 * 1024,
		TotalHits:  8550,
		TotalMiss:  1450,
		HitRate:    85.5,
		OldestFile: time.Now().Add(-24 * time.Hour),
		NewestFile: time.Now(),
	}

	if stats.Entries != 1000 {
		t.Errorf("Expected Entries=1000, got %d", stats.Entries)
	}
	if stats.TotalHits != 8550 {
		t.Errorf("Expected TotalHits=8550, got %d", stats.TotalHits)
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		max  int
		want string
	}{
		{"short string", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"needs truncation", "hello world", 8, "hello..."},
		{"very short max", "hello", 4, "h..."},
		{"empty string", "", 5, ""},
		{"single char max", "hello", 3, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncateString(tt.s, tt.max)
			if got != tt.want {
				t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.max, got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		want     string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 5*time.Minute + 30*time.Second, "5m30s"},
		{"hours", 2*time.Hour + 15*time.Minute, "2h15m"},
		{"days", 48*time.Hour + 3*time.Hour, "2d3h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.duration)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.duration, got, tt.want)
			}
		})
	}
}

func TestTableConfigCenter(t *testing.T) {
	var buf bytes.Buffer
	table := NewTableWithConfig([]string{"Col1"}, TableConfig{Writer: &buf, Center: true})
	if table == nil {
		t.Fatal("NewTableWithConfig with Center returned nil")
	}
}

func TestPrintBuildSummary(t *testing.T) {
	DisableColors()
	defer EnableColors()

	output := captureStdout(t, func() {
		PrintBuildSummary(BuildStats{
			Total:       10,
			Remote:      4,
			CacheHits:   3,
			Local:       2,
			Failed:      1,
			Duration:    12*time.Second + 500*time.Millisecond,
			TimeSaved:   4 * time.Second,
			TasksFailed: []string{"a.c", "b.c"},
		})
	})

	checks := []string{"Build Summary", "Total Files", "Remote", "Cache Hits", "Local Fallback", "Failed", "Time Saved", "Failed files:", "a.c", "b.c"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}

func TestPrintWorkersTable(t *testing.T) {
	DisableColors()
	defer EnableColors()

	noWorkersOutput := captureStdout(t, func() {
		PrintWorkersTable(nil, 0, 0)
	})
	if !strings.Contains(noWorkersOutput, "No workers connected") {
		t.Fatalf("expected no-workers message, got %q", noWorkersOutput)
	}

	workersOutput := captureStdout(t, func() {
		PrintWorkersTable([]WorkerInfo{{
			ID:           "worker-identifier-1234567890",
			Arch:         "amd64",
			Cores:        8,
			MemoryGB:     16,
			ActiveTasks:  2,
			CircuitState: "OPEN",
		}}, 1, 0)
	})

	checks := []string{"Workers:", "amd64", "16.0 GB", "2", "OPEN", "worker-identifier"}
	for _, check := range checks {
		if !strings.Contains(workersOutput, check) {
			t.Fatalf("expected output to contain %q, got %q", check, workersOutput)
		}
	}
}

func TestPrintWorkersTableCompact(t *testing.T) {
	DisableColors()
	defer EnableColors()

	output := captureStdout(t, func() {
		PrintWorkersTableCompact([]WorkerInfo{{
			ID:           "compact-worker-identifier-1234567890",
			Arch:         "arm64",
			Cores:        12,
			CircuitState: "HALF_OPEN",
		}}, 1, 1)
	})

	checks := []string{"Workers:", "arm64", "12", "HALF_OPEN", "compact-worker-id"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Fatalf("expected output to contain %q, got %q", check, output)
		}
	}
}

func TestPrintStatus(t *testing.T) {
	DisableColors()
	defer EnableColors()

	fullOutput := captureStdout(t, func() {
		PrintStatus(StatusInfo{
			Address:     "localhost:9000",
			Healthy:     true,
			ActiveTasks: 3,
			QueuedTasks: 1,
			Workers:     2,
			CacheHits:   9,
			Uptime:      2*time.Hour + 5*time.Minute,
		})
	})

	checks := []string{"Coordinator Status", "localhost:9000", "healthy", "Active Tasks:", "Queued Tasks:", "Workers:", "Cache Hits:", "2h5m"}
	for _, check := range checks {
		if !strings.Contains(fullOutput, check) {
			t.Fatalf("expected output to contain %q, got %q", check, fullOutput)
		}
	}

	minimalOutput := captureStdout(t, func() {
		PrintStatus(StatusInfo{Address: "localhost:9000"})
	})
	if strings.Contains(minimalOutput, "Cache Hits:") || strings.Contains(minimalOutput, "Workers:") {
		t.Fatalf("did not expect optional rows in minimal output: %q", minimalOutput)
	}
}

func TestPrintCacheStats(t *testing.T) {
	DisableColors()
	defer EnableColors()

	highUsageOutput := captureStdout(t, func() {
		PrintCacheStats(CacheStats{
			Directory: "/tmp/cache",
			Entries:   5,
			TotalSize: 95 * 1024 * 1024,
			MaxSize:   100 * 1024 * 1024,
			TotalHits: 9,
			TotalMiss: 1,
		})
	})

	highChecks := []string{"Cache Statistics", "/tmp/cache", "Entries:", "95.0%", "Total Hits:", "Total Miss:", "90.0%"}
	for _, check := range highChecks {
		if !strings.Contains(highUsageOutput, check) {
			t.Fatalf("expected output to contain %q, got %q", check, highUsageOutput)
		}
	}

	zeroMissOutput := captureStdout(t, func() {
		PrintCacheStats(CacheStats{
			Directory: "/tmp/cache",
			Entries:   2,
			TotalSize: 10 * 1024 * 1024,
			MaxSize:   100 * 1024 * 1024,
			TotalHits: 0,
			TotalMiss: 0,
		})
	})

	if strings.Contains(zeroMissOutput, "Total Miss:") || strings.Contains(zeroMissOutput, "Hit Rate:") {
		t.Fatalf("did not expect miss or hit-rate rows in zero-hit output: %q", zeroMissOutput)
	}
}
