package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStderr(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stderr = w
	defer func() {
		os.Stderr = original
	}()

	fn()
	if err := w.Close(); err != nil {
		t.Fatalf("close stderr pipe failed: %v", err)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("ReadAll failed: %v", err)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("close stderr reader failed: %v", err)
	}
	return string(data)
}

func TestNewProgressAndMethods(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgress(ProgressConfig{
		Total:       2,
		Description: "Compiling",
		Verbose:     true,
		Writer:      &buf,
	})

	if bar == nil {
		t.Fatal("NewProgress returned nil")
	}
	if !bar.IsVerbose() {
		t.Fatal("expected verbose progress bar")
	}
	bar.SetDescription("Linking")
	if err := bar.Add(1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := bar.Increment(); err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Linking") {
		t.Fatalf("expected output to contain updated description, got %q", output)
	}
}

func TestProgressClear(t *testing.T) {
	var buf bytes.Buffer
	bar := NewProgress(ProgressConfig{
		Total:       1,
		Description: "Clearing",
		Writer:      &buf,
	})

	if err := bar.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}
	if bar.IsVerbose() {
		t.Fatal("expected non-verbose progress bar")
	}
}

func TestNewBuildProgress(t *testing.T) {
	output := captureStderr(t, func() {
		bar := NewBuildProgress(1)
		if bar == nil {
			t.Fatal("NewBuildProgress returned nil")
		}
		if err := bar.Increment(); err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
		if err := bar.Finish(); err != nil {
			t.Fatalf("Finish failed: %v", err)
		}
	})

	if !strings.Contains(output, "Compiling") {
		t.Fatalf("expected stderr output to contain default description, got %q", output)
	}
}

func TestCompileProgress(t *testing.T) {
	DisableColors()
	defer EnableColors()

	output := captureStderr(t, func() {
		bar := CompileProgress(1, "Compile")
		if bar == nil {
			t.Fatal("CompileProgress returned nil")
		}
		if err := bar.Increment(); err != nil {
			t.Fatalf("Increment failed: %v", err)
		}
		if err := bar.Finish(); err != nil {
			t.Fatalf("Finish failed: %v", err)
		}
	})

	if !strings.Contains(output, "Compile") {
		t.Fatalf("expected stderr output to contain description, got %q", output)
	}
}

func TestSpinnerProgress(t *testing.T) {
	output := captureStderr(t, func() {
		bar := SpinnerProgress("Waiting")
		if bar == nil {
			t.Fatal("SpinnerProgress returned nil")
		}
		if err := bar.Add(1); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		if err := bar.Clear(); err != nil {
			t.Fatalf("Clear failed: %v", err)
		}
	})

	if !strings.Contains(output, "Waiting") {
		t.Fatalf("expected stderr output to contain description, got %q", output)
	}
}

func TestDownloadProgress(t *testing.T) {
	DisableColors()
	defer EnableColors()

	output := captureStderr(t, func() {
		bar := DownloadProgress(1024, "Downloading")
		if bar == nil {
			t.Fatal("DownloadProgress returned nil")
		}
		if err := bar.Add(512); err != nil {
			t.Fatalf("Add failed: %v", err)
		}
		if err := bar.Finish(); err != nil {
			t.Fatalf("Finish failed: %v", err)
		}
	})

	if !strings.Contains(output, "Downloading") {
		t.Fatalf("expected stderr output to contain description, got %q", output)
	}
}
