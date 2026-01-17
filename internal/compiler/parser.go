package compiler

import (
	"path/filepath"
	"strings"
)

// CompilerType represents the type of compiler.
type CompilerType int

const (
	CompilerUnknown CompilerType = iota
	CompilerGCC
	CompilerClang
	CompilerGPP     // g++
	CompilerClangPP // clang++
)

// ParsedArgs holds parsed compiler arguments.
type ParsedArgs struct {
	Compiler      string
	CompilerType  CompilerType
	InputFiles    []string
	OutputFile    string
	IncludeDirs   []string
	Defines       []string
	Flags         []string
	IsCompileOnly bool // -c flag present
	IsPreprocess  bool // -E flag present
	IsLink        bool // linking (no -c, -E, -S)
	TargetArch    string
	Language      string // c, c++, objective-c, etc.
	Standard      string // c11, c++17, etc.
	Optimization  string // O0, O1, O2, O3, Os, Ofast
}

// Parse parses compiler command line arguments.
func Parse(args []string) *ParsedArgs {
	if len(args) == 0 {
		return nil
	}

	p := &ParsedArgs{
		Compiler:     args[0],
		CompilerType: detectCompilerType(args[0]),
		Flags:        make([]string, 0),
		InputFiles:   make([]string, 0),
		IncludeDirs:  make([]string, 0),
		Defines:      make([]string, 0),
	}

	i := 1
	for i < len(args) {
		arg := args[i]

		switch {
		case arg == "-c":
			p.IsCompileOnly = true
			p.Flags = append(p.Flags, arg)

		case arg == "-E":
			p.IsPreprocess = true
			p.Flags = append(p.Flags, arg)

		case arg == "-S":
			p.Flags = append(p.Flags, arg)

		case arg == "-o":
			if i+1 < len(args) {
				i++
				p.OutputFile = args[i]
			}

		case strings.HasPrefix(arg, "-o"):
			p.OutputFile = arg[2:]

		case arg == "-I":
			if i+1 < len(args) {
				i++
				p.IncludeDirs = append(p.IncludeDirs, args[i])
			}

		case strings.HasPrefix(arg, "-I"):
			p.IncludeDirs = append(p.IncludeDirs, arg[2:])

		case arg == "-D":
			if i+1 < len(args) {
				i++
				p.Defines = append(p.Defines, args[i])
			}

		case strings.HasPrefix(arg, "-D"):
			p.Defines = append(p.Defines, arg[2:])

		case arg == "-x":
			if i+1 < len(args) {
				i++
				p.Language = args[i]
			}

		case strings.HasPrefix(arg, "-x"):
			p.Language = arg[2:]

		case strings.HasPrefix(arg, "-std="):
			p.Standard = arg[5:]
			p.Flags = append(p.Flags, arg)

		case arg == "-march" || arg == "-mtune":
			if i+1 < len(args) {
				i++
				p.TargetArch = args[i]
				p.Flags = append(p.Flags, arg, args[i])
			}

		case strings.HasPrefix(arg, "-march="):
			p.TargetArch = arg[7:]
			p.Flags = append(p.Flags, arg)

		case strings.HasPrefix(arg, "-O"):
			p.Optimization = arg[2:]
			p.Flags = append(p.Flags, arg)

		case strings.HasPrefix(arg, "-"):
			p.Flags = append(p.Flags, arg)

		default:
			// Input file
			if isSourceFile(arg) || isObjectFile(arg) {
				p.InputFiles = append(p.InputFiles, arg)
			} else {
				p.Flags = append(p.Flags, arg)
			}
		}
		i++
	}

	// Determine if linking
	p.IsLink = !p.IsCompileOnly && !p.IsPreprocess && len(p.InputFiles) > 0

	// Detect language from input files if not specified
	if p.Language == "" && len(p.InputFiles) > 0 {
		p.Language = detectLanguage(p.InputFiles[0])
	}

	return p
}

// IsDistributable returns true if this compilation can be distributed.
func (p *ParsedArgs) IsDistributable() bool {
	// Can distribute compile-only operations
	if !p.IsCompileOnly {
		return false
	}
	// Must have exactly one input file
	if len(p.InputFiles) != 1 {
		return false
	}
	// Must be a source file (not object file)
	if !isSourceFile(p.InputFiles[0]) {
		return false
	}
	return true
}

// ToArgs reconstructs the command line arguments.
func (p *ParsedArgs) ToArgs() []string {
	args := []string{p.Compiler}

	for _, inc := range p.IncludeDirs {
		args = append(args, "-I"+inc)
	}
	for _, def := range p.Defines {
		args = append(args, "-D"+def)
	}
	if p.Language != "" {
		args = append(args, "-x", p.Language)
	}

	args = append(args, p.Flags...)
	args = append(args, p.InputFiles...)

	if p.OutputFile != "" {
		args = append(args, "-o", p.OutputFile)
	}

	return args
}

func detectCompilerType(compiler string) CompilerType {
	base := filepath.Base(compiler)
	switch {
	case strings.Contains(base, "clang++"):
		return CompilerClangPP
	case strings.Contains(base, "clang"):
		return CompilerClang
	case strings.Contains(base, "g++"):
		return CompilerGPP
	case strings.Contains(base, "gcc"):
		return CompilerGCC
	default:
		return CompilerUnknown
	}
}

func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".c", ".cc", ".cpp", ".cxx", ".c++", ".m", ".mm", ".s", ".S":
		return true
	default:
		return false
	}
}

func isObjectFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".o" || ext == ".obj"
}

func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".c":
		return "c"
	case ".cc", ".cpp", ".cxx", ".c++":
		return "c++"
	case ".m":
		return "objective-c"
	case ".mm":
		return "objective-c++"
	case ".s", ".S":
		return "assembler"
	default:
		return ""
	}
}
