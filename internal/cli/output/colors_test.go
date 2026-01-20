package output

import (
	"testing"
)

func TestStatusLabel(t *testing.T) {
	tests := []struct {
		name   string
		status string
		want   string
	}{
		{"remote status", "remote", "[remote]"},
		{"cache status", "cache", "[cache]"},
		{"local status", "local", "[local]"},
		{"failed status", "failed", "[failed]"},
		{"healthy status", "healthy", "[healthy]"},
		{"unhealthy status", "unhealthy", "[unhealthy]"},
		{"pending status", "pending", "[pending]"},
		{"running status", "running", "[running]"},
		{"success status", "success", "[success]"},
		{"unknown status", "unknown", "[unknown]"},
		{"empty status", "", "[]"},
	}

	// Disable colors for consistent testing
	DisableColors()
	defer EnableColors()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusLabel(tt.status)
			if got != tt.want {
				t.Errorf("StatusLabel(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestStatusIcon(t *testing.T) {
	DisableColors()
	defer EnableColors()

	tests := []struct {
		name string
		ok   bool
		want string
	}{
		{"ok true", true, "✓"},
		{"ok false", false, "✗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StatusIcon(tt.ok)
			if got != tt.want {
				t.Errorf("StatusIcon(%v) = %q, want %q", tt.ok, got, tt.want)
			}
		})
	}
}

func TestHealthy(t *testing.T) {
	DisableColors()
	defer EnableColors()

	tests := []struct {
		name    string
		healthy bool
		want    string
	}{
		{"healthy true", true, "healthy ✓"},
		{"healthy false", false, "unhealthy ✗"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Healthy(tt.healthy)
			if got != tt.want {
				t.Errorf("Healthy(%v) = %q, want %q", tt.healthy, got, tt.want)
			}
		})
	}
}

func TestWorkerStatus(t *testing.T) {
	DisableColors()
	defer EnableColors()

	tests := []struct {
		name         string
		circuitState string
		want         string
	}{
		{"empty state", "", "healthy"},
		{"closed state", "CLOSED", "healthy"},
		{"open state", "OPEN", "OPEN"},
		{"half open state", "HALF_OPEN", "HALF_OPEN"},
		{"other state", "OTHER", "OTHER"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WorkerStatus(tt.circuitState)
			if got != tt.want {
				t.Errorf("WorkerStatus(%q) = %q, want %q", tt.circuitState, got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	DisableColors()
	defer EnableColors()

	tests := []struct {
		name  string
		value float64
		want  string
	}{
		{"high percentage", 90.5, "90.5%"},
		{"medium percentage", 65.0, "65.0%"},
		{"low percentage", 30.0, "30.0%"},
		{"zero percentage", 0.0, "0.0%"},
		{"exactly 80", 80.0, "80.0%"},
		{"exactly 50", 50.0, "50.0%"},
		{"hundred percent", 100.0, "100.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Percent(tt.value)
			if got != tt.want {
				t.Errorf("Percent(%v) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestByteSize(t *testing.T) {
	DisableColors()
	defer EnableColors()

	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 2048, "2.00 KB"},
		{"megabytes", 5 * 1024 * 1024, "5.00 MB"},
		{"gigabytes", 3 * 1024 * 1024 * 1024, "3.00 GB"},
		{"zero bytes", 0, "0 B"},
		{"one byte", 1, "1 B"},
		{"exactly 1KB", 1024, "1.00 KB"},
		{"exactly 1MB", 1024 * 1024, "1.00 MB"},
		{"exactly 1GB", 1024 * 1024 * 1024, "1.00 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ByteSize(tt.bytes)
			if got != tt.want {
				t.Errorf("ByteSize(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestColorFunctions(t *testing.T) {
	DisableColors()
	defer EnableColors()

	// Test that color functions return the input string when colors are disabled
	tests := []struct {
		name string
		fn   func(a ...interface{}) string
		arg  string
	}{
		{"Success", Success, "test"},
		{"Error", Error, "test"},
		{"Warning", Warning, "test"},
		{"Info", Info, "test"},
		{"Bold", Bold, "test"},
		{"Dim", Dim, "test"},
		{"SuccessBold", SuccessBold, "test"},
		{"ErrorBold", ErrorBold, "test"},
		{"WarningBold", WarningBold, "test"},
		{"InfoBold", InfoBold, "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.arg)
			if got != tt.arg {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.arg, got, tt.arg)
			}
		})
	}
}

func TestAutoDetectColors(t *testing.T) {
	// Just test that it doesn't panic
	AutoDetectColors()
}

func TestEnableDisableColors(t *testing.T) {
	// Test enable/disable cycle
	EnableColors()
	DisableColors()
	EnableColors()
	// No assertion needed, just verifying no panic
}
