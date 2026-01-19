package build

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.CacheMaxSize != 10*1024*1024*1024 {
		t.Errorf("expected 10GB cache size, got %d", cfg.CacheMaxSize)
	}
	if cfg.CacheTTLHours != 168 {
		t.Errorf("expected 168 hours TTL, got %d", cfg.CacheTTLHours)
	}
	if cfg.CoordinatorAddr != "localhost:9000" {
		t.Errorf("expected localhost:9000, got %s", cfg.CoordinatorAddr)
	}
	if !cfg.FallbackEnabled {
		t.Error("fallback should be enabled by default")
	}
	if cfg.Timeout != 5*time.Minute {
		t.Errorf("expected 5 minute timeout, got %v", cfg.Timeout)
	}
}

func TestNew_WithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	if svc.preprocessor == nil {
		t.Error("preprocessor should not be nil")
	}
	if svc.fallback == nil {
		t.Error("fallback should not be nil")
	}
}

func TestNew_CacheFailure(t *testing.T) {
	cfg := DefaultConfig()
	// Use invalid cache dir (path that can't be created)
	cfg.CacheDir = "/nonexistent/path/that/cannot/exist/hybridgrid-test"

	svc, err := New(cfg)
	// Should not fail, just continue without cache
	if err != nil {
		t.Fatalf("New should not fail even with invalid cache: %v", err)
	}
	defer svc.Close()

	if svc.cache != nil {
		t.Log("Note: cache initialization succeeded unexpectedly, may be running as root")
	}
}

func TestService_Build_LocalFallback(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Skip if gcc not available
	if _, err := os.Stat("/usr/bin/gcc"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/gcc"); os.IsNotExist(err) {
			t.Skip("gcc not available")
		}
	}

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = filepath.Join(tmpDir, "cache")
	cfg.FallbackEnabled = true
	cfg.Verbose = true

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	// Create a simple C file
	srcFile := filepath.Join(tmpDir, "test.c")
	err = os.WriteFile(srcFile, []byte(`
int main(void) {
    return 0;
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	args := &compiler.ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	req := &Request{
		TaskID:     "test-task-1",
		SourceFile: srcFile,
		OutputFile: filepath.Join(tmpDir, "test.o"),
		Args:       args,
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    30 * time.Second,
	}

	ctx := context.Background()
	result, err := svc.Build(ctx, req)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Should use local fallback since no coordinator is connected
	if !result.Fallback {
		t.Error("expected fallback to be true")
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d: %s", result.ExitCode, result.Stderr)
	}
	if len(result.ObjectFile) == 0 {
		t.Error("object file should not be empty")
	}
	if result.FallbackReason != "no coordinator connection" {
		t.Errorf("unexpected fallback reason: %s", result.FallbackReason)
	}
}

func TestService_Build_CacheHit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	// Skip if gcc not available
	if _, err := os.Stat("/usr/bin/gcc"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/gcc"); os.IsNotExist(err) {
			t.Skip("gcc not available")
		}
	}

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = filepath.Join(tmpDir, "cache")
	cfg.FallbackEnabled = true
	cfg.Verbose = true

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	// Create a simple C file
	srcFile := filepath.Join(tmpDir, "test.c")
	err = os.WriteFile(srcFile, []byte(`
int add(int a, int b) {
    return a + b;
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	args := &compiler.ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	req := &Request{
		TaskID:     "test-cache-1",
		SourceFile: srcFile,
		OutputFile: filepath.Join(tmpDir, "test.o"),
		Args:       args,
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    30 * time.Second,
	}

	ctx := context.Background()

	// First build - should be cache miss
	result1, err := svc.Build(ctx, req)
	if err != nil {
		t.Fatalf("First build failed: %v", err)
	}
	if result1.CacheHit {
		t.Error("first build should be cache miss")
	}

	// Second build with same source - should be cache hit
	req.TaskID = "test-cache-2"
	result2, err := svc.Build(ctx, req)
	if err != nil {
		t.Fatalf("Second build failed: %v", err)
	}
	if !result2.CacheHit {
		t.Error("second build should be cache hit")
	}

	// Cache hit should be faster
	if result2.Duration > result1.Duration {
		t.Log("Warning: cache hit was slower than cache miss")
	}
}

func TestService_Build_PreprocessError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = filepath.Join(tmpDir, "cache")
	cfg.FallbackEnabled = true

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	// Create a C file with missing include
	srcFile := filepath.Join(tmpDir, "bad.c")
	err = os.WriteFile(srcFile, []byte(`
#include "nonexistent_header.h"
int main() { return 0; }
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	args := &compiler.ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	req := &Request{
		TaskID:     "test-bad",
		SourceFile: srcFile,
		OutputFile: filepath.Join(tmpDir, "bad.o"),
		Args:       args,
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    30 * time.Second,
	}

	ctx := context.Background()
	_, err = svc.Build(ctx, req)

	// Should fail during preprocessing
	if err == nil {
		t.Fatal("expected error for missing header")
	}
}

func TestService_Close(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	// Close should not fail
	err = svc.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}

	// Close again should not panic
	err = svc.Close()
	if err != nil {
		t.Errorf("Second close failed: %v", err)
	}
}

func TestIsDistributable(t *testing.T) {
	tests := []struct {
		name     string
		args     *compiler.ParsedArgs
		expected bool
	}{
		{
			name: "compile only",
			args: &compiler.ParsedArgs{
				Compiler:      "gcc",
				IsCompileOnly: true,
				InputFiles:    []string{"test.c"},
			},
			expected: true,
		},
		{
			name: "linking",
			args: &compiler.ParsedArgs{
				Compiler:      "gcc",
				IsCompileOnly: false,
				IsLink:        true,
				InputFiles:    []string{"test.c"},
			},
			expected: false,
		},
		{
			name: "preprocess only",
			args: &compiler.ParsedArgs{
				Compiler:     "gcc",
				IsPreprocess: true,
				InputFiles:   []string{"test.c"},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDistributable(tt.args)
			if result != tt.expected {
				t.Errorf("IsDistributable() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGenerateCacheKey(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	args := &compiler.ParsedArgs{
		Compiler: "gcc",
		Flags:    []string{"-O2", "-Wall"},
		Defines:  []string{"DEBUG"},
	}

	req := &Request{
		Args:       args,
		TargetArch: pb.Architecture_ARCH_X86_64,
	}

	key1 := svc.generateCacheKey(req, []byte("source code 1"))
	key2 := svc.generateCacheKey(req, []byte("source code 2"))
	key3 := svc.generateCacheKey(req, []byte("source code 1"))

	// Different source should produce different keys
	if key1 == key2 {
		t.Error("different source should produce different keys")
	}

	// Same source should produce same key
	if key1 != key3 {
		t.Error("same source should produce same key")
	}
}
