package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreprocessLocalExpandsIncludes(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "t.c")
	if err := os.WriteFile(file, []byte("#include <stdio.h>\nint main(void){return 0;}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got := preprocessLocal(file, "gcc", nil)
	if got == nil {
		t.Fatal("preprocessLocal returned nil")
	}
	if strings.Contains(string(got), "#include") {
		t.Fatalf("output still contains #include directives: %q", got[:80])
	}
	if !strings.Contains(string(got), "int main") {
		t.Fatal("output lost the translation unit")
	}
}

func TestPreprocessLocalFallsBackOnMissingFile(t *testing.T) {
	if got := preprocessLocal("/nonexistent/x.c", "gcc", nil); got != nil {
		t.Fatalf("expected nil fallback for missing file, got %d bytes", len(got))
	}
}
