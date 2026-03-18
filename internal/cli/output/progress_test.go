package output

import (
	"bytes"
	"strings"
	"testing"
)

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
	var buf bytes.Buffer
	bar := NewBuildProgressWithWriter(1, &buf)
	if bar == nil {
		t.Fatal("NewBuildProgressWithWriter returned nil")
	}
	if err := bar.Increment(); err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Compiling") {
		t.Fatalf("expected output to contain default description, got %q", buf.String())
	}
}

func TestCompileProgress(t *testing.T) {
	var buf bytes.Buffer
	bar := CompileProgressWithWriter(1, "Compile", &buf)
	if bar == nil {
		t.Fatal("CompileProgressWithWriter returned nil")
	}
	if err := bar.Increment(); err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Compile") {
		t.Fatalf("expected output to contain description, got %q", buf.String())
	}
}

func TestSpinnerProgress(t *testing.T) {
	var buf bytes.Buffer
	bar := SpinnerProgressWithWriter("Waiting", &buf)
	if bar == nil {
		t.Fatal("SpinnerProgressWithWriter returned nil")
	}
	if err := bar.Add(1); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := bar.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Waiting") {
		t.Fatalf("expected output to contain description, got %q", buf.String())
	}
}

func TestDownloadProgress(t *testing.T) {
	var buf bytes.Buffer
	bar := DownloadProgressWithWriter(1024, "Downloading", &buf)
	if bar == nil {
		t.Fatal("DownloadProgressWithWriter returned nil")
	}
	if err := bar.Add(512); err != nil {
		t.Fatalf("Add failed: %v", err)
	}
	if err := bar.Finish(); err != nil {
		t.Fatalf("Finish failed: %v", err)
	}

	if !strings.Contains(buf.String(), "Downloading") {
		t.Fatalf("expected output to contain description, got %q", buf.String())
	}
}
