package compiler

import (
	"reflect"
	"testing"
)

// sliceEqual compares two string slices, treating nil and empty slice as equal
func sliceEqual(a, b []string) bool {
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	return reflect.DeepEqual(a, b)
}

func TestTranslateToMSVC_BasicFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Compile-only flag
		{"compile only", []string{"-c"}, []string{"/c"}},

		// Optimization flags
		{"O0 optimization", []string{"-O0"}, []string{"/Od"}},
		{"O1 optimization", []string{"-O1"}, []string{"/O1"}},
		{"O2 optimization", []string{"-O2"}, []string{"/O2"}},
		{"O3 to Ox", []string{"-O3"}, []string{"/Ox"}},
		{"Os optimization", []string{"-Os"}, []string{"/O1"}},
		{"Ofast with fp:fast", []string{"-Ofast"}, []string{"/Ox", "/fp:fast"}},

		// Debug flags
		{"debug g", []string{"-g"}, []string{"/Zi"}},
		{"debug g0", []string{"-g0"}, []string{}},
		{"debug g1", []string{"-g1"}, []string{"/Zi"}},
		{"debug g2", []string{"-g2"}, []string{"/Zi"}},
		{"debug g3", []string{"-g3"}, []string{"/Zi"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_WarningFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"Wall to W4", []string{"-Wall"}, []string{"/W4"}},
		{"Wextra to W4", []string{"-Wextra"}, []string{"/W4"}},
		{"Werror to WX", []string{"-Werror"}, []string{"/WX"}},
		{"w to W0", []string{"-w"}, []string{"/W0"}},
		{"pedantic ignored", []string{"-pedantic"}, []string{}},
		{"Wpedantic ignored", []string{"-Wpedantic"}, []string{}},
		{"Wno-unused ignored", []string{"-Wno-unused"}, []string{}},
		{"Wno-warning ignored", []string{"-Wno-error"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_IncludeDefine(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Include paths
		{"include path attached", []string{"-I/usr/include"}, []string{"/I/usr/include"}},
		{"include path separate", []string{"-I", "/opt/include"}, []string{"/I/opt/include"}},
		{"include local", []string{"-I./include"}, []string{"/I./include"}},

		// Defines
		{"define macro", []string{"-DDEBUG"}, []string{"/DDEBUG"}},
		{"define with value", []string{"-DDEBUG=1"}, []string{"/DDEBUG=1"}},
		{"define separate", []string{"-D", "VERSION"}, []string{"/DVERSION"}},
		{"define string value", []string{"-DNAME=\"foo\""}, []string{"/DNAME=\"foo\""}},

		// Undefines
		{"undef attached", []string{"-UDEBUG"}, []string{"/UDEBUG"}},
		{"undef separate", []string{"-U", "NDEBUG"}, []string{"/UNDEBUG"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_Standards(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// C standards
		{"C89 ignored", []string{"-std=c89"}, []string{}},
		{"C99 ignored", []string{"-std=c99"}, []string{}},
		{"C11 standard", []string{"-std=c11"}, []string{"/std:c11"}},
		{"C17 standard", []string{"-std=c17"}, []string{"/std:c17"}},

		// C++ standards
		{"C++98 ignored", []string{"-std=c++98"}, []string{}},
		{"C++03 ignored", []string{"-std=c++03"}, []string{}},
		{"C++11 to c++14", []string{"-std=c++11"}, []string{"/std:c++14"}},
		{"C++14 standard", []string{"-std=c++14"}, []string{"/std:c++14"}},
		{"C++17 standard", []string{"-std=c++17"}, []string{"/std:c++17"}},
		{"C++20 standard", []string{"-std=c++20"}, []string{"/std:c++20"}},
		{"C++23 to latest", []string{"-std=c++23"}, []string{"/std:c++latest"}},

		// GNU extensions
		{"GNU89 ignored", []string{"-std=gnu89"}, []string{}},
		{"GNU99 ignored", []string{"-std=gnu99"}, []string{}},
		{"GNU11 to c11", []string{"-std=gnu11"}, []string{"/std:c11"}},
		{"GNU17 to c17", []string{"-std=gnu17"}, []string{"/std:c17"}},
		{"GNU++11 to c++14", []string{"-std=gnu++11"}, []string{"/std:c++14"}},
		{"GNU++14 to c++14", []string{"-std=gnu++14"}, []string{"/std:c++14"}},
		{"GNU++17 to c++17", []string{"-std=gnu++17"}, []string{"/std:c++17"}},
		{"GNU++20 to c++20", []string{"-std=gnu++20"}, []string{"/std:c++20"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_OutputFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Object file output
		{"object output separate", []string{"-o", "main.o"}, []string{"/Fomain.o"}},
		{"object output attached", []string{"-omain.o"}, []string{"/Fomain.o"}},
		{"obj output", []string{"-o", "main.obj"}, []string{"/Fomain.obj"}},

		// Executable output
		{"exe output", []string{"-o", "main.exe"}, []string{"/Femain.exe"}},
		{"no extension", []string{"-o", "program"}, []string{"/Feprogram"}},
		{"exe attached", []string{"-omain"}, []string{"/Femain"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_FunctionFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Exception handling
		{"fexceptions", []string{"-fexceptions"}, []string{"/EHsc"}},
		{"fno-exceptions", []string{"-fno-exceptions"}, []string{"/EHs-c-"}},

		// RTTI
		{"frtti", []string{"-frtti"}, []string{"/GR"}},
		{"fno-rtti", []string{"-fno-rtti"}, []string{"/GR-"}},

		// Inline functions
		{"finline-functions", []string{"-finline-functions"}, []string{"/Ob2"}},
		{"fno-inline-functions", []string{"-fno-inline-functions"}, []string{"/Ob0"}},
		{"fno-inline", []string{"-fno-inline"}, []string{"/Ob0"}},

		// Stack protection
		{"fstack-protector", []string{"-fstack-protector"}, []string{"/GS"}},
		{"fstack-protector-all", []string{"-fstack-protector-all"}, []string{"/GS"}},
		{"fstack-protector-strong", []string{"-fstack-protector-strong"}, []string{"/GS"}},
		{"fno-stack-protector", []string{"-fno-stack-protector"}, []string{"/GS-"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_IgnoredFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		// Position independent code
		{"fPIC ignored", []string{"-fPIC"}, []string{}},
		{"fPIE ignored", []string{"-fPIE"}, []string{}},
		{"fpic ignored", []string{"-fpic"}, []string{}},
		{"fpie ignored", []string{"-fpie"}, []string{}},
		{"fno-pic ignored", []string{"-fno-pic"}, []string{}},

		// Common flags to ignore
		{"pipe ignored", []string{"-pipe"}, []string{}},
		{"pthread ignored", []string{"-pthread"}, []string{}},
		{"fdiagnostics-color ignored", []string{"-fdiagnostics-color"}, []string{}},
		{"fcolor-diagnostics ignored", []string{"-fcolor-diagnostics"}, []string{}},
		{"fvisibility=hidden ignored", []string{"-fvisibility=hidden"}, []string{}},
		{"fvisibility=default ignored", []string{"-fvisibility=default"}, []string{}},

		// Architecture flags
		{"m32 ignored", []string{"-m32"}, []string{}},
		{"m64 ignored", []string{"-m64"}, []string{}},
		{"march=native ignored", []string{"-march=native"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_Preprocessor(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"preprocess E", []string{"-E"}, []string{"/EP"}},
		{"preprocess P", []string{"-P"}, []string{"/P"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_LanguageSelection(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"language C", []string{"-x", "c"}, []string{"/Tc"}},
		{"language C++", []string{"-x", "c++"}, []string{"/Tp"}},
		{"language other", []string{"-x", "assembler"}, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_SourceFilePassthrough(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{"C source file", []string{"main.c"}, []string{"main.c"}},
		{"C++ source file", []string{"main.cpp"}, []string{"main.cpp"}},
		{"Header file", []string{"header.h"}, []string{"header.h"}},
		{"Multiple sources", []string{"a.c", "b.c"}, []string{"a.c", "b.c"}},
		{"path with spaces", []string{"/path/to/source file.c"}, []string{"/path/to/source file.c"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_ComplexCombinations(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "real world compile",
			input:    []string{"-c", "-O2", "-Wall", "-I./include", "-DNDEBUG", "main.c", "-o", "main.o"},
			expected: []string{"/c", "/O2", "/W4", "/I./include", "/DNDEBUG", "main.c", "/Fomain.o"},
		},
		{
			name:     "C++ with standard",
			input:    []string{"-c", "-std=c++17", "-O2", "-g", "-I/usr/include", "app.cpp"},
			expected: []string{"/c", "/std:c++17", "/O2", "/Zi", "/I/usr/include", "app.cpp"},
		},
		{
			name:     "debug build",
			input:    []string{"-c", "-O0", "-g", "-DDEBUG", "-Wall", "-Werror", "test.c"},
			expected: []string{"/c", "/Od", "/Zi", "/DDEBUG", "/W4", "/WX", "test.c"},
		},
		{
			name:     "release build",
			input:    []string{"-c", "-O3", "-DNDEBUG", "-fno-rtti", "-fno-exceptions", "perf.cpp"},
			expected: []string{"/c", "/Ox", "/DNDEBUG", "/GR-", "/EHs-c-", "perf.cpp"},
		},
		{
			name:     "mixed ignored and valid",
			input:    []string{"-c", "-fPIC", "-O2", "-pthread", "-Wall", "lib.c"},
			expected: []string{"/c", "/O2", "/W4", "lib.c"},
		},
		{
			name:     "multiple includes and defines",
			input:    []string{"-I/a", "-I", "/b", "-DA=1", "-D", "B=2", "x.c"},
			expected: []string{"/I/a", "/I/b", "/DA=1", "/DB=2", "x.c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TranslateToMSVC(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("TranslateToMSVC(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTranslateToMSVC_EmptyInput(t *testing.T) {
	got := TranslateToMSVC([]string{})
	if got == nil {
		got = []string{}
	}
	if len(got) != 0 {
		t.Errorf("TranslateToMSVC([]) = %v, want []", got)
	}
}

func TestTranslateToMSVC_UnknownFlagIgnored(t *testing.T) {
	input := []string{"-unknown-flag", "-O2", "file.c"}
	got := TranslateToMSVC(input)
	expected := []string{"/O2", "file.c"}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("TranslateToMSVC(%v) = %v, want %v", input, got, expected)
	}
}

func TestMSVCToCLFlags_AddsDefaults(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty input adds all defaults",
			input:    []string{},
			expected: []string{"/nologo", "/EHsc", "/permissive-"},
		},
		{
			name:     "with existing flags adds defaults",
			input:    []string{"/O2", "/W4"},
			expected: []string{"/O2", "/W4", "/nologo", "/EHsc", "/permissive-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MSVCToCLFlags(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("MSVCToCLFlags(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestMSVCToCLFlags_SkipsExisting(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "has nologo",
			input:    []string{"/nologo", "/O2"},
			expected: []string{"/nologo", "/O2", "/EHsc", "/permissive-"},
		},
		{
			name:     "has EHsc",
			input:    []string{"/EHsc", "/O2"},
			expected: []string{"/EHsc", "/O2", "/nologo", "/permissive-"},
		},
		{
			name:     "has EHa",
			input:    []string{"/EHa", "/O2"},
			expected: []string{"/EHa", "/O2", "/nologo", "/permissive-"},
		},
		{
			name:     "has permissive-",
			input:    []string{"/permissive-", "/O2"},
			expected: []string{"/permissive-", "/O2", "/nologo", "/EHsc"},
		},
		{
			name:     "has all defaults",
			input:    []string{"/nologo", "/EHsc", "/permissive-"},
			expected: []string{"/nologo", "/EHsc", "/permissive-"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MSVCToCLFlags(tt.input)
			if !sliceEqual(got, tt.expected) {
				t.Errorf("MSVCToCLFlags(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsMSVCCompiler(t *testing.T) {
	tests := []struct {
		compiler string
		expected bool
	}{
		// Positive cases
		{"cl.exe", true},
		{"cl", true},
		{"CL.EXE", true},
		{"CL", true},
		{"Cl.Exe", true},
		{"C:\\Program Files\\MSVC\\bin\\cl.exe", true},
		{"/path/to/cl.exe", true},
		{"x86_64-w64-mingw32-cl.exe", true},

		// Negative cases
		{"gcc", false},
		{"g++", false},
		{"clang", false},
		{"clang++", false},
		{"/usr/bin/gcc", false},
		{"/usr/bin/clang", false},
		{"cls.exe", false},
		{"clcompile", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.compiler, func(t *testing.T) {
			got := IsMSVCCompiler(tt.compiler)
			if got != tt.expected {
				t.Errorf("IsMSVCCompiler(%q) = %v, want %v", tt.compiler, got, tt.expected)
			}
		})
	}
}

func BenchmarkTranslateToMSVC(b *testing.B) {
	args := []string{"-c", "-O2", "-Wall", "-I./include", "-I/usr/include", "-DNDEBUG", "-DVERSION=1.0", "-std=c++17", "-fno-rtti", "main.cpp", "-o", "main.o"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		TranslateToMSVC(args)
	}
}

func BenchmarkMSVCToCLFlags(b *testing.B) {
	args := []string{"/O2", "/W4", "/I./include", "/DNDEBUG"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		MSVCToCLFlags(args)
	}
}

func BenchmarkIsMSVCCompiler(b *testing.B) {
	compilers := []string{"gcc", "cl.exe", "clang", "CL", "/path/to/cl.exe"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, c := range compilers {
			IsMSVCCompiler(c)
		}
	}
}
