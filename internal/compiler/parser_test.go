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
		{[]string{"gcc", "-c", "trampoline.S", "-o", "t.o"}, false}, // assembly: local only
		{[]string{"gcc", "-c", "lowercase.s", "-o", "l.o"}, false},  // assembly: local only
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

func TestParse_DriverOperands(t *testing.T) {
	cmake := []string{"/usr/bin/c++", "-DFOO", "-I/src/x", "-isystem", "/usr/include/y", "-O2", "-std=c++17", "-MD", "-MT", "absl/base/CMakeFiles/base.dir/internal/cycleclock.cc.o", "-MF", "absl/base/CMakeFiles/base.dir/internal/cycleclock.cc.o.d", "-o", "absl/base/CMakeFiles/base.dir/internal/cycleclock.cc.o", "-c", "/src/absl/base/internal/cycleclock.cc"}
	tests := []struct {
		name          string
		args          []string
		inputs        []string
		flags         []string
		depFlags      []string
		distributable bool
	}{
		{
			name: "cmake ninja", args: cmake,
			inputs:        []string{"/src/absl/base/internal/cycleclock.cc"},
			flags:         []string{"-isystem", "/usr/include/y", "-O2", "-std=c++17", "-c"},
			depFlags:      []string{"-MD", "-MT", "absl/base/CMakeFiles/base.dir/internal/cycleclock.cc.o", "-MF", "absl/base/CMakeFiles/base.dir/internal/cycleclock.cc.o.d"},
			distributable: true,
		},
		{
			name: "forced includes", args: []string{"gcc", "-c", "-include", "config.h", "-imacros", "macros.h", "a.c"},
			inputs: []string{"a.c"}, flags: []string{"-c", "-include", "config.h", "-imacros", "macros.h"}, distributable: true,
		},
		{
			name: "separate dependencies", args: []string{"gcc", "-c", "-MMD", "-MP", "-MF", "x.d", "x.c"},
			inputs: []string{"x.c"}, flags: []string{"-c"}, depFlags: []string{"-MMD", "-MP", "-MF", "x.d"}, distributable: true,
		},
		{
			name: "joined dependencies and include", args: []string{"gcc", "-c", "-MFx.d", "-isystem/dir", "x.c"},
			inputs: []string{"x.c"}, flags: []string{"-c", "-isystem/dir"}, depFlags: []string{"-MFx.d"}, distributable: true,
		},
		{
			name: "cpython", args: []string{"gcc", "-c", "-o", "build/python.o", "-I", "Include", "-DPy_BUILD_CORE", "-std=c11", "-O3", "-march=x86-64", "-x", "c", "Python/python.c"},
			inputs: []string{"Python/python.c"}, flags: []string{"-c", "-std=c11", "-O3", "-march=x86-64"}, distributable: true,
		},
		{
			name: "link", args: []string{"g++", "one.o", "two.o", "-o", "app"},
			inputs: []string{"one.o", "two.o"}, flags: []string{}, distributable: false,
		},
		{
			name: "explicit header language", args: []string{"g++", "-c", "-x", "c++", "file.h", "-o", "file.o"},
			inputs: []string{"file.h"}, flags: []string{"-c"}, distributable: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.args)
			if !reflect.DeepEqual(got.InputFiles, tt.inputs) {
				t.Errorf("InputFiles = %v, want %v", got.InputFiles, tt.inputs)
			}
			if !reflect.DeepEqual(got.Flags, tt.flags) {
				t.Errorf("Flags = %v, want %v", got.Flags, tt.flags)
			}
			if !reflect.DeepEqual(got.DepFlags, tt.depFlags) {
				t.Errorf("DepFlags = %v, want %v", got.DepFlags, tt.depFlags)
			}
			if got.IsDistributable() != tt.distributable {
				t.Errorf("IsDistributable = %v, want %v", got.IsDistributable(), tt.distributable)
			}
			if tt.name == "cpython" {
				if got.OutputFile != "build/python.o" || !reflect.DeepEqual(got.IncludeDirs, []string{"Include"}) || !reflect.DeepEqual(got.Defines, []string{"Py_BUILD_CORE"}) || got.Standard != "c11" || got.Optimization != "3" || got.TargetArch != "x86-64" || got.Language != "c" {
					t.Errorf("CPython fields changed: %+v", got)
				}
			}
		})
	}
}

func TestParse_OptionOperandsNeverBecomeInputs(t *testing.T) {
	options := []string{
		"-MF", "-MT", "-MQ", "-isystem", "-idirafter", "-iquote",
		"-iprefix", "-iwithprefix", "-iwithprefixbefore", "-isysroot",
		"-iframework", "-imultilib", "-include", "-imacros", "-U",
		"-Xpreprocessor", "-Xassembler", "-Xlinker", "--sysroot",
		"-target", "-arch", "-aux-info", "-dumpbase", "-dumpdir",
		"-Wa", "-Wp",
	}
	for _, option := range options {
		t.Run(option, func(t *testing.T) {
			parsed := Parse([]string{"gcc", "-c", option, "value.o", "src.c"})
			if !reflect.DeepEqual(parsed.InputFiles, []string{"src.c"}) {
				t.Errorf("InputFiles = %v, want [src.c]", parsed.InputFiles)
			}
			values := parsed.Flags
			if option == "-MF" || option == "-MT" || option == "-MQ" {
				values = parsed.DepFlags
			}
			found := false
			for i := 0; i+1 < len(values); i++ {
				if values[i] == option && values[i+1] == "value.o" {
					found = true
				}
			}
			if !found {
				t.Errorf("option and operand not adjacent in %v", values)
			}
		})
	}
}

func TestParse_DependencyOnlyModesStayLocal(t *testing.T) {
	for _, mode := range []string{"-M", "-MM"} {
		t.Run(mode, func(t *testing.T) {
			parsed := Parse([]string{"gcc", "-c", mode, "src.c"})
			if !reflect.DeepEqual(parsed.DepFlags, []string{mode}) || parsed.IsDistributable() {
				t.Errorf("dependency-only mode parsed incorrectly: %+v", parsed)
			}
		})
	}
}

func TestParse_PreprocessingOrder(t *testing.T) {
	tests := []struct {
		name    string
		options []string
	}{
		{"undefine then define", []string{"-U", "FOO", "-D", "FOO"}},
		{"define then undefine", []string{"-D", "FOO", "-U", "FOO"}},
		{"include family", []string{"-I", "inc", "-iquote", "quoted", "-isystem", "system", "-idirafter", "after", "-include", "config.h", "-imacros", "macros.h", "-nostdinc"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"gcc", "-c"}, tt.options...)
			args = append(args, "source.c")
			parsed := Parse(args)
			if !reflect.DeepEqual(parsed.PreprocessFlags, tt.options) {
				t.Errorf("PreprocessFlags = %v, want %v", parsed.PreprocessFlags, tt.options)
			}
		})
	}
}

func TestParse_IndirectDependencies(t *testing.T) {
	tests := []struct {
		name    string
		options []string
		want    []string
	}{
		{"wp md", []string{"-Wp,-MD,wp.d"}, []string{"-MD", "-MF", "wp.d"}},
		{"wp mmd", []string{"-Wp,-MMD,wp.d"}, []string{"-MMD", "-MF", "wp.d"}},
		{"wp mmd with mp", []string{"-Wp,-MMD,wp.d,-MP"}, []string{"-MMD", "-MF", "wp.d", "-MP"}},
		{"wp md with target", []string{"-Wp,-MD,wp.d,-MT,obj.o"}, []string{"-MD", "-MF", "wp.d", "-MT", "obj.o"}},
		{"xpreprocessor", []string{"-Xpreprocessor", "-MD", "-Xpreprocessor", "-MF", "-Xpreprocessor", "xp.d"}, []string{"-MD", "-MF", "xp.d"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{"gcc", "-c"}, tt.options...)
			args = append(args, "source.c")
			parsed := Parse(args)
			if !reflect.DeepEqual(parsed.DepFlags, tt.want) || !parsed.GeneratesDependencies() {
				t.Errorf("dependency flags = %v, want %v", parsed.DepFlags, tt.want)
			}
			if !parsed.IsDistributable() || len(parsed.InputFiles) != 1 {
				t.Errorf("indirect dependency command parsed incorrectly: %+v", parsed)
			}
		})
	}
}

func TestParse_ClangIncludeOperands(t *testing.T) {
	options := []string{"-include-pch", "prefix.pch", "-isystem-after", "later", "-isystem-afterjoined", "-iwithsysroot", "sysinc", "-iframework", "frameworks", "-iframeworkwithsysroot", "sysframeworks", "-isysroot", "sdk", "-iwithprefixbefore", "before", "-iwithprefix", "with", "-iprefix", "prefix", "-imultilib", "multi"}
	args := append([]string{"clang", "-c"}, options...)
	args = append(args, "source.c")
	parsed := Parse(args)
	if !reflect.DeepEqual(parsed.PreprocessFlags, options) {
		t.Errorf("PreprocessFlags = %v, want %v", parsed.PreprocessFlags, options)
	}
	if !reflect.DeepEqual(parsed.InputFiles, []string{"source.c"}) {
		t.Errorf("InputFiles = %v", parsed.InputFiles)
	}
	if len(parsed.CompileFlags()) != 0 {
		t.Errorf("preprocessed compile flags = %v, want none", parsed.CompileFlags())
	}
}

func TestParse_SideOutputsStayLocal(t *testing.T) {
	tests := [][]string{
		{"-aux-info", "info.txt"},
		{"-dumpbase", "unit.c"},
		{"-dumpdir", "dumps/"},
		{"-save-temps"},
		{"-save-temps=obj"},
		{"-fdump-tree-all"},
		{"-gsplit-dwarf"},
		{"-ftest-coverage"},
		{"--coverage"},
	}
	for _, options := range tests {
		t.Run(options[0], func(t *testing.T) {
			args := append([]string{"gcc", "-c"}, options...)
			args = append(args, "source.c")
			parsed := Parse(args)
			if parsed.IsDistributable() {
				t.Errorf("side output command should stay local: %v", args)
			}
			if !reflect.DeepEqual(parsed.InputFiles, []string{"source.c"}) {
				t.Errorf("InputFiles = %v", parsed.InputFiles)
			}
		})
	}
	if parsed := Parse([]string{"gcc", "-c", "-ftime-report", "source.c"}); !parsed.IsDistributable() {
		t.Error("-ftime-report only writes diagnostics and should remain distributable")
	}
}
