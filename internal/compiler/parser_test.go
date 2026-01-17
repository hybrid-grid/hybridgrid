package compiler

import (
	"reflect"
	"testing"
)

func TestParse_SimpleCompile(t *testing.T) {
	args := []string{"gcc", "-c", "-O2", "foo.c", "-o", "foo.o"}
	p := Parse(args)

	if p.Compiler != "gcc" {
		t.Errorf("Expected compiler 'gcc', got '%s'", p.Compiler)
	}
	if p.CompilerType != CompilerGCC {
		t.Errorf("Expected CompilerGCC, got %v", p.CompilerType)
	}
	if !p.IsCompileOnly {
		t.Error("Expected IsCompileOnly to be true")
	}
	if p.Optimization != "2" {
		t.Errorf("Expected optimization '2', got '%s'", p.Optimization)
	}
	if len(p.InputFiles) != 1 || p.InputFiles[0] != "foo.c" {
		t.Errorf("Expected input file 'foo.c', got %v", p.InputFiles)
	}
	if p.OutputFile != "foo.o" {
		t.Errorf("Expected output 'foo.o', got '%s'", p.OutputFile)
	}
}

func TestParse_WithIncludes(t *testing.T) {
	args := []string{"g++", "-c", "-I/usr/include", "-I", "/opt/include", "main.cpp"}
	p := Parse(args)

	expected := []string{"/usr/include", "/opt/include"}
	if !reflect.DeepEqual(p.IncludeDirs, expected) {
		t.Errorf("Expected includes %v, got %v", expected, p.IncludeDirs)
	}
}

func TestParse_WithDefines(t *testing.T) {
	args := []string{"clang", "-c", "-DDEBUG", "-D", "VERSION=1.0", "test.c"}
	p := Parse(args)

	expected := []string{"DEBUG", "VERSION=1.0"}
	if !reflect.DeepEqual(p.Defines, expected) {
		t.Errorf("Expected defines %v, got %v", expected, p.Defines)
	}
}

func TestParse_WithStandard(t *testing.T) {
	args := []string{"g++", "-c", "-std=c++17", "main.cpp"}
	p := Parse(args)

	if p.Standard != "c++17" {
		t.Errorf("Expected standard 'c++17', got '%s'", p.Standard)
	}
	if p.Language != "c++" {
		t.Errorf("Expected language 'c++', got '%s'", p.Language)
	}
}

func TestParse_Preprocessor(t *testing.T) {
	args := []string{"gcc", "-E", "foo.c"}
	p := Parse(args)

	if !p.IsPreprocess {
		t.Error("Expected IsPreprocess to be true")
	}
	if p.IsDistributable() {
		t.Error("Preprocessing should not be distributable")
	}
}

func TestParse_Linking(t *testing.T) {
	args := []string{"gcc", "foo.o", "bar.o", "-o", "program"}
	p := Parse(args)

	if !p.IsLink {
		t.Error("Expected IsLink to be true")
	}
	if p.IsDistributable() {
		t.Error("Linking should not be distributable")
	}
}

func TestIsDistributable(t *testing.T) {
	tests := []struct {
		args   []string
		expect bool
	}{
		{[]string{"gcc", "-c", "foo.c", "-o", "foo.o"}, true},
		{[]string{"g++", "-c", "-O2", "main.cpp"}, true},
		{[]string{"gcc", "foo.o", "bar.o", "-o", "program"}, false}, // linking
		{[]string{"gcc", "-E", "foo.c"}, false},                     // preprocessing
		{[]string{"gcc", "-c", "foo.c", "bar.c"}, false},            // multiple inputs
	}

	for _, tt := range tests {
		p := Parse(tt.args)
		if p.IsDistributable() != tt.expect {
			t.Errorf("Args %v: expected distributable=%v, got %v", tt.args, tt.expect, p.IsDistributable())
		}
	}
}

func TestDetectCompilerType(t *testing.T) {
	tests := []struct {
		compiler string
		expected CompilerType
	}{
		{"/usr/bin/gcc", CompilerGCC},
		{"gcc-12", CompilerGCC},
		{"/usr/bin/g++", CompilerGPP},
		{"clang", CompilerClang},
		{"/opt/llvm/bin/clang++", CompilerClangPP},
		{"unknown-compiler", CompilerUnknown},
	}

	for _, tt := range tests {
		got := detectCompilerType(tt.compiler)
		if got != tt.expected {
			t.Errorf("detectCompilerType(%s) = %v, want %v", tt.compiler, got, tt.expected)
		}
	}
}

func TestToArgs(t *testing.T) {
	original := []string{"gcc", "-c", "-I/include", "-DDEBUG", "-O2", "foo.c", "-o", "foo.o"}
	p := Parse(original)
	reconstructed := p.ToArgs()

	// The order may differ, but key elements should be present
	if reconstructed[0] != "gcc" {
		t.Errorf("Expected first arg 'gcc', got '%s'", reconstructed[0])
	}

	hasInclude := false
	hasDefine := false
	hasInput := false
	hasOutput := false

	for i, arg := range reconstructed {
		if arg == "-I/include" {
			hasInclude = true
		}
		if arg == "-DDEBUG" {
			hasDefine = true
		}
		if arg == "foo.c" {
			hasInput = true
		}
		if arg == "-o" && i+1 < len(reconstructed) && reconstructed[i+1] == "foo.o" {
			hasOutput = true
		}
	}

	if !hasInclude || !hasDefine || !hasInput || !hasOutput {
		t.Errorf("Reconstructed args missing elements: %v", reconstructed)
	}
}
