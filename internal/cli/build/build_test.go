package build

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/compiler"
	"github.com/h3nr1-d14z/hybridgrid/internal/grpc/client"
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

func TestSetClient(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	if svc.client != nil {
		t.Error("client should be nil initially")
	}

	svc.SetClient(&client.Client{})

	if svc.client == nil {
		t.Error("client should be set")
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"nil error", nil, false},
		{"connection error", fmt.Errorf("connection refused"), true},
		{"timeout error", fmt.Errorf("context deadline exceeded"), true},
		{"unavailable", fmt.Errorf("service unavailable"), true},
		{"reset error", fmt.Errorf("connection reset by peer"), true},
		{"EOF error", fmt.Errorf("unexpected EOF"), true},
		{"other error", fmt.Errorf("invalid argument"), false},
		{"compilation error", fmt.Errorf("compilation failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetryableError(tt.err)
			if got != tt.want {
				t.Errorf("isRetryableError(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

func TestBuildRemoteArgs(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	tests := []struct {
		name     string
		args     *compiler.ParsedArgs
		wantArgs []string
	}{
		{
			name: "basic flags",
			args: &compiler.ParsedArgs{
				Compiler: "gcc",
				Flags:    []string{"-O2", "-Wall", "-g"},
				Standard: "c11",
			},
			wantArgs: []string{"-c", "-O2", "-Wall", "-g", "-std=c11"},
		},
		{
			name: "filters -I and -D",
			args: &compiler.ParsedArgs{
				Compiler: "gcc",
				Flags:    []string{"-O2", "-I/usr/include", "-DDEBUG", "-Wall"},
				Standard: "c++17",
			},
			wantArgs: []string{"-c", "-O2", "-Wall", "-std=c++17"},
		},
		{
			name: "no standard",
			args: &compiler.ParsedArgs{
				Compiler: "clang",
				Flags:    []string{"-O3"},
			},
			wantArgs: []string{"-c", "-O3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.buildRemoteArgs(tt.args)
			if len(got) != len(tt.wantArgs) {
				t.Errorf("buildRemoteArgs() len = %d, want %d", len(got), len(tt.wantArgs))
			}
			for i, arg := range got {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("buildRemoteArgs()[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestBuildRemoteArgsForRaw(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	tests := []struct {
		name     string
		args     *compiler.ParsedArgs
		wantArgs []string
	}{
		{
			name: "keeps -I and -D",
			args: &compiler.ParsedArgs{
				Compiler: "gcc",
				Flags:    []string{"-O2", "-I/usr/include", "-DDEBUG", "-Wall"},
				Standard: "c11",
			},
			wantArgs: []string{"-c", "-O2", "-I/usr/include", "-DDEBUG", "-Wall", "-std=c11"},
		},
		{
			name: "all flags preserved",
			args: &compiler.ParsedArgs{
				Compiler: "clang",
				Flags:    []string{"-O3", "-Iinclude", "-DNDEBUG"},
			},
			wantArgs: []string{"-c", "-O3", "-Iinclude", "-DNDEBUG"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.buildRemoteArgsForRaw(tt.args)
			if len(got) != len(tt.wantArgs) {
				t.Errorf("buildRemoteArgsForRaw() len = %d, want %d", len(got), len(tt.wantArgs))
			}
			for i, arg := range got {
				if i < len(tt.wantArgs) && arg != tt.wantArgs[i] {
					t.Errorf("buildRemoteArgsForRaw()[%d] = %q, want %q", i, arg, tt.wantArgs[i])
				}
			}
		})
	}
}

func TestGetClientOS(t *testing.T) {
	os := getClientOS()
	if os == "" {
		t.Error("getClientOS should not return empty string")
	}
	if os != "darwin" && os != "linux" && os != "windows" {
		t.Logf("getClientOS returned: %s (expected darwin/linux/windows)", os)
	}
}

func TestGetClientArch(t *testing.T) {
	arch := getClientArch()
	validArchs := []pb.Architecture{
		pb.Architecture_ARCH_X86_64,
		pb.Architecture_ARCH_ARM64,
		pb.Architecture_ARCH_ARMV7,
		pb.Architecture_ARCH_UNSPECIFIED,
	}
	valid := false
	for _, a := range validArchs {
		if arch == a {
			valid = true
			break
		}
	}
	if !valid {
		t.Errorf("getClientArch returned unexpected value: %v", arch)
	}
}

func TestCollectIncludeFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	includeDir := filepath.Join(tmpDir, "include")
	if err := os.MkdirAll(includeDir, 0755); err != nil {
		t.Fatal(err)
	}

	header1 := filepath.Join(includeDir, "test.h")
	if err := os.WriteFile(header1, []byte("#define TEST 1\n"), 0644); err != nil {
		t.Fatal(err)
	}

	header2 := filepath.Join(includeDir, "utils.hpp")
	if err := os.WriteFile(header2, []byte("void util();\n"), 0644); err != nil {
		t.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("failed to restore working directory: %v", chdirErr)
		}
	}()

	tests := []struct {
		name      string
		args      *compiler.ParsedArgs
		wantFiles int
	}{
		{
			name: "collects local includes",
			args: &compiler.ParsedArgs{
				Flags: []string{"-Iinclude", "-O2"},
			},
			wantFiles: 2,
		},
		{
			name: "skips system includes",
			args: &compiler.ParsedArgs{
				Flags: []string{"-I/usr/include", "-O2"},
			},
			wantFiles: 0,
		},
		{
			name: "no include flags",
			args: &compiler.ParsedArgs{
				Flags: []string{"-O2", "-Wall"},
			},
			wantFiles: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := svc.collectIncludeFiles(tt.args)
			if err != nil {
				t.Fatalf("collectIncludeFiles failed: %v", err)
			}
			if len(files) != tt.wantFiles {
				t.Errorf("collectIncludeFiles() collected %d files, want %d", len(files), tt.wantFiles)
			}
		})
	}
}

func TestGenerateCacheKeyRaw(t *testing.T) {
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
		Flags:    []string{"-O2"},
		Defines:  []string{"DEBUG"},
	}

	req := &Request{
		Args:       args,
		TargetArch: pb.Architecture_ARCH_X86_64,
	}

	key1 := svc.generateCacheKeyRaw(req, []byte("int main() { return 0; }"))
	key2 := svc.generateCacheKeyRaw(req, []byte("int main() { return 1; }"))

	if key1 == key2 {
		t.Error("different source should produce different cache keys")
	}

	if key1 == "" {
		t.Error("cache key should not be empty")
	}
}

func TestService_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir
	cfg.Verbose = true

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	if !svc.verbose {
		t.Error("verbose mode should be enabled")
	}

	cfg.Verbose = false
	svc2, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc2.Close()

	if svc2.verbose {
		t.Error("verbose mode should be disabled")
	}
}

func TestService_MaxRetries(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir
	cfg.MaxRetries = 5
	cfg.RetryDelay = 50 * time.Millisecond

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	if svc.maxRetries != 5 {
		t.Errorf("maxRetries = %d, want 5", svc.maxRetries)
	}

	if svc.retryDelay != 50*time.Millisecond {
		t.Errorf("retryDelay = %v, want 50ms", svc.retryDelay)
	}
}

func TestService_Build_ReadSourceError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	req := &Request{
		TaskID:     "test",
		SourceFile: filepath.Join(tmpDir, "nonexistent.c"),
		OutputFile: filepath.Join(tmpDir, "test.o"),
		Args: &compiler.ParsedArgs{
			Compiler:      "gcc",
			IsCompileOnly: true,
		},
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	_, err = svc.Build(ctx, req)

	if err == nil {
		t.Error("expected error for nonexistent source file")
	}
}

func TestService_Build_FallbackDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir
	cfg.FallbackEnabled = false

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	srcFile := filepath.Join(tmpDir, "test.c")
	err = os.WriteFile(srcFile, []byte("int main() { return 0; }"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &Request{
		TaskID:     "test",
		SourceFile: srcFile,
		OutputFile: filepath.Join(tmpDir, "test.o"),
		Args: &compiler.ParsedArgs{
			Compiler:      "gcc",
			IsCompileOnly: true,
			InputFiles:    []string{srcFile},
		},
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    5 * time.Second,
	}

	ctx := context.Background()
	_, err = svc.Build(ctx, req)

	if err == nil {
		t.Error("expected error when fallback is disabled and no coordinator")
	}
}

func TestCompileRemotePreprocessed_RetryableError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	grpcClient, err := client.New(client.Config{
		Address:  "localhost:1",
		Insecure: true,
		Timeout:  20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("client.New failed: %v", err)
	}
	defer grpcClient.Close()

	svc.SetClient(grpcClient)
	svc.maxRetries = 1
	svc.retryDelay = time.Millisecond

	req := &Request{
		TaskID:     "remote-preprocessed-error",
		SourceFile: "main.c",
		Args: &compiler.ParsedArgs{
			Compiler: "gcc",
			Flags:    []string{"-O2", "-Iinclude", "-DDEBUG"},
			Standard: "c11",
		},
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    20 * time.Millisecond,
	}

	_, err = svc.compileRemotePreprocessed(context.Background(), req, []byte("int main(void){return 0;}"))
	if err == nil {
		t.Fatal("expected remote preprocessed compilation to fail")
	}
	if !strings.Contains(err.Error(), "remote compilation failed after 1 attempts") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCompileRemoteRaw_RetryableError(t *testing.T) {
	tmpDir := t.TempDir()
	cfg := DefaultConfig()
	cfg.CacheDir = tmpDir

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer svc.Close()

	grpcClient, err := client.New(client.Config{
		Address:  "localhost:1",
		Insecure: true,
		Timeout:  20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("client.New failed: %v", err)
	}
	defer grpcClient.Close()

	svc.SetClient(grpcClient)
	svc.maxRetries = 1
	svc.retryDelay = time.Millisecond

	req := &Request{
		TaskID:     "remote-raw-error",
		SourceFile: "main.c",
		Args: &compiler.ParsedArgs{
			Compiler: "gcc",
			Flags:    []string{"-O2", "-Iinclude", "-DDEBUG"},
			Standard: "c11",
		},
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    20 * time.Millisecond,
	}

	_, err = svc.compileRemoteRaw(context.Background(), req, []byte("int main(void){return 0;}"), map[string][]byte{"header.h": []byte("#define X 1")})
	if err == nil {
		t.Fatal("expected remote raw compilation to fail")
	}
	if !strings.Contains(err.Error(), "remote compilation failed after 1 attempts") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestService_Build_RemoteFailureFallsBack(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	if _, err := os.Stat("/usr/bin/gcc"); os.IsNotExist(err) {
		if _, err := os.Stat("/usr/local/bin/gcc"); os.IsNotExist(err) {
			t.Skip("gcc not available")
		}
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

	grpcClient, err := client.New(client.Config{
		Address:  "localhost:1",
		Insecure: true,
		Timeout:  20 * time.Millisecond,
	})
	if err != nil {
		t.Fatalf("client.New failed: %v", err)
	}
	defer grpcClient.Close()
	svc.SetClient(grpcClient)
	svc.maxRetries = 1
	svc.retryDelay = time.Millisecond

	srcFile := filepath.Join(tmpDir, "remote-fallback.c")
	err = os.WriteFile(srcFile, []byte("int main(void) { return 0; }"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	req := &Request{
		TaskID:     "remote-fallback",
		SourceFile: srcFile,
		OutputFile: filepath.Join(tmpDir, "remote-fallback.o"),
		Args: &compiler.ParsedArgs{
			Compiler:      "gcc",
			IsCompileOnly: true,
			InputFiles:    []string{srcFile},
		},
		TargetArch: pb.Architecture_ARCH_X86_64,
		Timeout:    30 * time.Second,
	}

	result, err := svc.Build(context.Background(), req)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}
	if !result.Fallback {
		t.Fatal("expected fallback result")
	}
	if !strings.Contains(result.FallbackReason, "remote error") {
		t.Fatalf("unexpected fallback reason: %q", result.FallbackReason)
	}
	if result.ExitCode != 0 {
		t.Fatalf("unexpected exit code %d: %s", result.ExitCode, result.Stderr)
	}
}
