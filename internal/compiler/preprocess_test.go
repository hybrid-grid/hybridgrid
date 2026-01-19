package compiler

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewPreprocessor(t *testing.T) {
	p := NewPreprocessor(DefaultPreprocessorConfig())
	if p == nil {
		t.Fatal("NewPreprocessor returned nil")
	}
	if p.defaultTimeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", p.defaultTimeout)
	}
}

func TestNewPreprocessor_CustomTimeout(t *testing.T) {
	p := NewPreprocessor(PreprocessorConfig{Timeout: 30 * time.Second})
	if p.defaultTimeout != 30*time.Second {
		t.Errorf("expected timeout 30s, got %v", p.defaultTimeout)
	}
}

func TestPreprocessor_SimpleFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	// Check if gcc is available
	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	// Create a simple C file
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "simple.c")
	err := os.WriteFile(srcFile, []byte(`
int main() {
    return 0;
}
`), 0644)
	if err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	p := NewPreprocessor(DefaultPreprocessorConfig())
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	ctx := context.Background()
	result, err := p.Preprocess(ctx, args, srcFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	if len(result.PreprocessedSource) == 0 {
		t.Error("preprocessed source is empty")
	}

	// Should contain main function
	if !strings.Contains(string(result.PreprocessedSource), "main") {
		t.Error("preprocessed source should contain 'main'")
	}
}

func TestPreprocessor_WithIncludes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmpDir := t.TempDir()

	// Create a header file
	incDir := filepath.Join(tmpDir, "include")
	if err := os.MkdirAll(incDir, 0755); err != nil {
		t.Fatal(err)
	}

	headerFile := filepath.Join(incDir, "myheader.h")
	err := os.WriteFile(headerFile, []byte(`
#define MY_CONSTANT 42
int my_function(void);
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Create source file that includes the header
	srcFile := filepath.Join(tmpDir, "with_include.c")
	err = os.WriteFile(srcFile, []byte(`
#include "myheader.h"

int main() {
    return MY_CONSTANT;
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPreprocessor(DefaultPreprocessorConfig())
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
		IncludeDirs:   []string{incDir},
	}

	ctx := context.Background()
	result, err := p.Preprocess(ctx, args, srcFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	preprocessed := string(result.PreprocessedSource)

	// MY_CONSTANT should be expanded to 42
	if !strings.Contains(preprocessed, "42") {
		t.Error("macro MY_CONSTANT should be expanded to 42")
	}

	// my_function declaration should be present
	if !strings.Contains(preprocessed, "my_function") {
		t.Error("my_function from header should be present")
	}
}

func TestPreprocessor_WithDefines(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "with_defines.c")
	err := os.WriteFile(srcFile, []byte(`
#ifdef DEBUG
int debug_mode = 1;
#else
int debug_mode = 0;
#endif

int value = VALUE;
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPreprocessor(DefaultPreprocessorConfig())
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
		Defines:       []string{"DEBUG", "VALUE=100"},
	}

	ctx := context.Background()
	result, err := p.Preprocess(ctx, args, srcFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	preprocessed := string(result.PreprocessedSource)

	// DEBUG branch should be taken
	if !strings.Contains(preprocessed, "debug_mode = 1") {
		t.Error("DEBUG branch should be taken")
	}

	// VALUE should be expanded to 100
	if !strings.Contains(preprocessed, "value = 100") {
		t.Error("VALUE should be expanded to 100")
	}
}

func TestPreprocessor_WithSystemHeaders(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "with_stdio.c")
	err := os.WriteFile(srcFile, []byte(`
#include <stdio.h>

int main() {
    printf("hello\n");
    return 0;
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPreprocessor(DefaultPreprocessorConfig())
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	ctx := context.Background()
	result, err := p.Preprocess(ctx, args, srcFile)
	if err != nil {
		t.Fatalf("Preprocess failed: %v", err)
	}

	// Preprocessed output should be significantly larger due to stdio.h
	if len(result.PreprocessedSource) < 1000 {
		t.Error("preprocessed source with stdio.h should be large")
	}
}

func TestPreprocessor_MissingHeader(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "missing_header.c")
	err := os.WriteFile(srcFile, []byte(`
#include "nonexistent_header.h"

int main() {
    return 0;
}
`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPreprocessor(DefaultPreprocessorConfig())
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	ctx := context.Background()
	_, err = p.Preprocess(ctx, args, srcFile)

	if err == nil {
		t.Fatal("expected error for missing header")
	}

	// Should be a PreprocessError
	prepErr, ok := err.(*PreprocessError)
	if !ok {
		t.Fatalf("expected PreprocessError, got %T", err)
	}

	// Error message should mention the missing file
	errMsg := prepErr.Error()
	if !strings.Contains(errMsg, "nonexistent_header.h") && !strings.Contains(errMsg, "file not found") {
		t.Errorf("error should mention missing header: %s", errMsg)
	}
}

func TestPreprocessor_NilArgs(t *testing.T) {
	p := NewPreprocessor(DefaultPreprocessorConfig())
	ctx := context.Background()

	_, err := p.Preprocess(ctx, nil, "test.c")
	if err == nil {
		t.Error("expected error for nil args")
	}
}

func TestPreprocessor_ContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping preprocessor test in short mode")
	}

	if _, err := exec.LookPath("gcc"); err != nil {
		t.Skip("gcc not available")
	}

	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "cancel.c")
	err := os.WriteFile(srcFile, []byte(`int main() { return 0; }`), 0644)
	if err != nil {
		t.Fatal(err)
	}

	p := NewPreprocessor(PreprocessorConfig{Timeout: 100 * time.Millisecond})
	args := &ParsedArgs{
		Compiler:      "gcc",
		IsCompileOnly: true,
		InputFiles:    []string{srcFile},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _ = p.Preprocess(ctx, args, srcFile)
	// Should either fail due to cancellation or succeed quickly
	// This test mainly ensures no panic
}

func TestBuildPreprocessArgs(t *testing.T) {
	p := NewPreprocessor(DefaultPreprocessorConfig())

	args := &ParsedArgs{
		Compiler:    "gcc",
		IncludeDirs: []string{"/usr/include", "/opt/include"},
		Defines:     []string{"DEBUG", "VERSION=1"},
		Standard:    "c11",
		Language:    "c",
		Flags:       []string{"-Wall", "-nostdinc"},
	}

	result := p.buildPreprocessArgs(args, "test.c")

	// Should contain -E
	found := false
	for _, arg := range result {
		if arg == "-E" {
			found = true
			break
		}
	}
	if !found {
		t.Error("should contain -E flag")
	}

	// Should contain include dirs
	hasInc1 := false
	hasInc2 := false
	for _, arg := range result {
		if arg == "-I/usr/include" {
			hasInc1 = true
		}
		if arg == "-I/opt/include" {
			hasInc2 = true
		}
	}
	if !hasInc1 || !hasInc2 {
		t.Error("should contain include directories")
	}

	// Should contain defines
	hasDef1 := false
	hasDef2 := false
	for _, arg := range result {
		if arg == "-DDEBUG" {
			hasDef1 = true
		}
		if arg == "-DVERSION=1" {
			hasDef2 = true
		}
	}
	if !hasDef1 || !hasDef2 {
		t.Error("should contain defines")
	}

	// Should contain -std=c11
	hasStd := false
	for _, arg := range result {
		if arg == "-std=c11" {
			hasStd = true
			break
		}
	}
	if !hasStd {
		t.Error("should contain -std=c11")
	}

	// Should contain -nostdinc (preprocessing flag)
	hasNostdinc := false
	for _, arg := range result {
		if arg == "-nostdinc" {
			hasNostdinc = true
			break
		}
	}
	if !hasNostdinc {
		t.Error("should contain -nostdinc")
	}

	// Should NOT contain -Wall (not a preprocessing flag)
	for _, arg := range result {
		if arg == "-Wall" {
			t.Error("should not contain -Wall")
		}
	}

	// Should contain source file at end
	if result[len(result)-1] != "test.c" {
		t.Error("source file should be last argument")
	}
}

func TestIsPreprocessingFlag(t *testing.T) {
	tests := []struct {
		flag     string
		expected bool
	}{
		{"-nostdinc", true},
		{"-nostdinc++", true},
		{"-Ufoo", true},
		{"-include", true},
		{"-isystem/path", true},
		{"-Wall", false},
		{"-O2", false},
		{"-c", false},
		{"-o", false},
		{"-fno-exceptions", true},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			result := isPreprocessingFlag(tt.flag)
			if result != tt.expected {
				t.Errorf("isPreprocessingFlag(%q) = %v, want %v", tt.flag, result, tt.expected)
			}
		})
	}
}

func TestExtractWarnings(t *testing.T) {
	stderr := `test.c:1:1: warning: unused variable 'x'
test.c:2:1: error: unknown type
test.c:3:1: warning: implicit declaration`

	warnings := extractWarnings(stderr)

	if !strings.Contains(warnings, "unused variable") {
		t.Error("should contain first warning")
	}
	if !strings.Contains(warnings, "implicit declaration") {
		t.Error("should contain second warning")
	}
	if strings.Contains(warnings, "error:") {
		t.Error("should not contain errors")
	}
}
