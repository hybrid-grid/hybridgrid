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
	Compiler     string
	CompilerType CompilerType
	InputFiles   []string
	OutputFile   string
	IncludeDirs  []string
	Defines      []string
	Flags        []string
	// PreprocessFlags retain preprocessing options in command-line order.
	PreprocessFlags []string
	// DepFlags are applied locally beside preprocessing, never on the worker.
	DepFlags      []string
	SideOutput    bool // Compilation would produce an artifact the protocol cannot return.
	IsCompileOnly bool // -c flag present
	IsPreprocess  bool // -E flag present
	IsLink        bool // linking (no -c, -E, -S)
	TargetArch    string
	Language      string // c, c++, objective-c, etc.
	Standard      string // c11, c++17, etc.
	Optimization  string // O0, O1, O2, O3, Os, Ofast
}

// operandOptions records the accepted operand forms of each driver option.
type operandOption struct {
	dependency bool
	preprocess bool
	localOnly  bool
	sideOutput bool
	forms      optionForm
}

type optionForm uint8

const (
	separateForm optionForm = 1 << iota
	joinedForm
	equalsForm
	commaForm
)

var operandOptions = map[string]operandOption{
	"-MF":                    {dependency: true, forms: separateForm | joinedForm},
	"-MT":                    {dependency: true, forms: separateForm | joinedForm},
	"-MQ":                    {dependency: true, forms: separateForm | joinedForm},
	"-isystem":               {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-isystem-after":         {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iwithsysroot":          {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-idirafter":             {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iquote":                {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iprefix":               {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iwithprefix":           {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iwithprefixbefore":     {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-isysroot":              {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iframework":            {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-iframeworkwithsysroot": {preprocess: true, localOnly: true, forms: separateForm},
	"-imultilib":             {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-include":               {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-include-pch":           {preprocess: true, localOnly: true, forms: separateForm},
	"-imacros":               {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-U":                     {preprocess: true, localOnly: true, forms: separateForm | joinedForm},
	"-Xpreprocessor":         {preprocess: true, localOnly: true, forms: separateForm},
	"-Xassembler":            {forms: separateForm},
	"-Xlinker":               {localOnly: true, forms: separateForm},
	"--sysroot":              {preprocess: true, localOnly: true, forms: separateForm | equalsForm},
	"-target":                {preprocess: true, forms: separateForm},
	"-arch":                  {preprocess: true, forms: separateForm},
	"-aux-info":              {sideOutput: true, forms: separateForm | joinedForm},
	"-dumpbase":              {sideOutput: true, forms: separateForm | equalsForm},
	"-dumpbase-ext":          {sideOutput: true, forms: separateForm | equalsForm},
	"-dumpdir":               {sideOutput: true, forms: separateForm | equalsForm},
	"-Wa":                    {forms: separateForm | commaForm},
	"-Wp":                    {preprocess: true, localOnly: true, forms: separateForm | commaForm},
}

var dependencyOptions = map[string]bool{
	"-MD": true, "-MMD": true, "-MP": true, "-M": true, "-MM": true,
}

// compilerOption checks exact names first, then only the joined forms declared
// for each option. A dash suffix cannot be a joined operand.
func compilerOption(arg string) (operandOption, bool, bool) {
	if spec, ok := operandOptions[arg]; ok {
		return spec, spec.forms&separateForm != 0, true
	}
	name := ""
	for option := range operandOptions {
		spec := operandOptions[option]
		if !strings.HasPrefix(arg, option) || len(arg) <= len(option) {
			continue
		}
		suffix := arg[len(option):]
		valid := (spec.forms&joinedForm != 0 && suffix[0] != '-' && suffix[0] != '=') ||
			(spec.forms&equalsForm != 0 && suffix[0] == '=') ||
			(spec.forms&commaForm != 0 && suffix[0] == ',')
		if valid && len(option) > len(name) {
			name = option
		}
	}
	if name != "" {
		return operandOptions[name], false, true
	}
	return operandOption{}, false, false
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

		case dependencyOptions[arg]:
			p.DepFlags = append(p.DepFlags, arg)
			if arg == "-M" || arg == "-MM" {
				p.IsPreprocess = true // These modes produce dependencies, not an object.
			}

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
				p.PreprocessFlags = append(p.PreprocessFlags, arg, args[i])
			}

		case strings.HasPrefix(arg, "-I"):
			p.IncludeDirs = append(p.IncludeDirs, arg[2:])
			p.PreprocessFlags = append(p.PreprocessFlags, arg)

		case arg == "-D":
			if i+1 < len(args) {
				i++
				p.Defines = append(p.Defines, args[i])
				p.PreprocessFlags = append(p.PreprocessFlags, arg, args[i])
			}

		case strings.HasPrefix(arg, "-D"):
			p.Defines = append(p.Defines, arg[2:])
			p.PreprocessFlags = append(p.PreprocessFlags, arg)

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
			// These request extra compiler outputs (.dwo, .gcno, dumps) that the
			// remote protocol does not return, so the compile must stay local.
			if strings.HasPrefix(arg, "-save-temps") || strings.HasPrefix(arg, "-fdump-") ||
				arg == "-gsplit-dwarf" || arg == "-ftest-coverage" || arg == "--coverage" {
				p.SideOutput = true
			}
			spec, separate, ok := compilerOption(arg)
			if !ok {
				p.Flags = append(p.Flags, arg)
				if isPreprocessingFlag(arg) {
					p.PreprocessFlags = append(p.PreprocessFlags, arg)
				}
				break
			}
			values := []string{arg}
			if separate && i+1 < len(args) {
				i++
				values = append(values, args[i])
			}
			if spec.dependency {
				p.DepFlags = append(p.DepFlags, values...)
			} else if arg == "-Wp" || strings.HasPrefix(arg, "-Wp,") {
				p.recordWp(values)
			} else if arg == "-Xpreprocessor" && len(values) == 2 && isDependencyFlag(values[1]) {
				p.DepFlags = append(p.DepFlags, values[1])
			} else if arg == "-Xpreprocessor" && len(values) == 2 && len(p.DepFlags) > 0 &&
				(p.DepFlags[len(p.DepFlags)-1] == "-MF" || p.DepFlags[len(p.DepFlags)-1] == "-MT" || p.DepFlags[len(p.DepFlags)-1] == "-MQ") {
				p.DepFlags = append(p.DepFlags, values[1])
			} else {
				p.Flags = append(p.Flags, values...)
				if spec.preprocess {
					p.PreprocessFlags = append(p.PreprocessFlags, values...)
				}
			}
			p.SideOutput = p.SideOutput || spec.sideOutput

		default:
			// Input file
			if isSourceFile(arg) || isObjectFile(arg) || (p.Language != "" && strings.EqualFold(filepath.Ext(arg), ".h")) {
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

func isDependencyFlag(arg string) bool {
	return dependencyOptions[arg] || arg == "-MF" || arg == "-MT" || arg == "-MQ"
}

// recordWp normalizes the dependency forms accepted by the preprocessor into
// ordinary driver flags, so local -E can regenerate the same depfile.
func (p *ParsedArgs) recordWp(values []string) {
	if len(values) != 1 || !strings.HasPrefix(values[0], "-Wp,") {
		p.Flags = append(p.Flags, values...)
		p.PreprocessFlags = append(p.PreprocessFlags, values...)
		return
	}
	// GCC splits the -Wp argument at commas: "-Wp,-MMD,f.d,-MP" passes
	// -MMD f.d -MP to the preprocessor, so only the component after -MD/-MMD
	// is the depfile; the rest are separate dependency options.
	parts := strings.Split(values[0][4:], ",")
	if len(parts) >= 2 && (parts[0] == "-MD" || parts[0] == "-MMD") {
		dep := []string{parts[0], "-MF", parts[1]}
		var rest []string
		for j := 2; j < len(parts); j++ {
			switch part := parts[j]; {
			case (part == "-MT" || part == "-MQ" || part == "-MF") && j+1 < len(parts):
				dep = append(dep, part, parts[j+1])
				j++
			case isDependencyFlag(part):
				dep = append(dep, part)
			default:
				rest = append(rest, part)
			}
		}
		p.DepFlags = append(p.DepFlags, dep...)
		if len(rest) > 0 {
			other := "-Wp," + strings.Join(rest, ",")
			p.Flags = append(p.Flags, other)
			p.PreprocessFlags = append(p.PreprocessFlags, other)
		}
		return
	}
	p.Flags = append(p.Flags, values...)
	p.PreprocessFlags = append(p.PreprocessFlags, values...)
}

// IsDistributable returns true if this compilation can be distributed.
func (p *ParsedArgs) IsDistributable() bool {
	// Can distribute compile-only operations
	if !p.IsCompileOnly || p.IsPreprocess || p.SideOutput {
		return false
	}
	// Must have exactly one input file
	if len(p.InputFiles) != 1 {
		return false
	}
	// Must be a source file (not object file)
	if !isSourceFile(p.InputFiles[0]) && !(strings.EqualFold(filepath.Ext(p.InputFiles[0]), ".h") && p.Language != "") {
		return false
	}
	// Assembly units (.s/.S) must compile locally: the preprocess-and-
	// ship pipeline produces objects missing their symbol definitions
	// (observed on CPython's asm_trampoline_x86_64.S, whose absence
	// only surfaces later at link time), and distributing them buys
	// nothing — assembling is not the expensive part of a build.
	if ext := strings.ToLower(filepath.Ext(p.InputFiles[0])); ext == ".s" {
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
	args = append(args, p.DepFlags...)
	args = append(args, p.InputFiles...)

	if p.OutputFile != "" {
		args = append(args, "-o", p.OutputFile)
	}

	return args
}

// CompileFlags returns flags safe to use when compiling preprocessed input.
// Dependency files and client-side include paths must never reach the worker.
func (p *ParsedArgs) CompileFlags() []string {
	return p.compileFlags(true)
}

// RawCompileFlags retains preprocessing options for raw source compilation.
func (p *ParsedArgs) RawCompileFlags() []string {
	return p.compileFlags(false)
}

func (p *ParsedArgs) compileFlags(preprocessed bool) []string {
	flags := make([]string, 0, len(p.Flags))
	for i := 0; i < len(p.Flags); i++ {
		flag := p.Flags[i]
		if flag == "-c" || flag == "-E" || flag == "-S" || dependencyOptions[flag] {
			continue
		}
		if p.Standard != "" && flag == "-std="+p.Standard {
			continue
		}
		if preprocessed && (strings.HasPrefix(flag, "-I") || strings.HasPrefix(flag, "-D")) {
			if (flag == "-I" || flag == "-D") && i+1 < len(p.Flags) {
				i++
			}
			continue
		}
		spec, separate, ok := compilerOption(flag)
		if ok && (spec.dependency || (preprocessed && spec.localOnly)) {
			if separate && i+1 < len(p.Flags) {
				i++
			}
			continue
		}
		flags = append(flags, flag)
		if ok && separate && i+1 < len(p.Flags) {
			i++
			flags = append(flags, p.Flags[i])
		}
	}
	return flags
}

// GeneratesDependencies reports whether preprocessing must write a depfile.
func (p *ParsedArgs) GeneratesDependencies() bool {
	for _, flag := range p.DepFlags {
		if flag == "-MD" || flag == "-MMD" {
			return true
		}
	}
	return false
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
