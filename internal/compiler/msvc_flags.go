package compiler

import (
	"strings"
)

// MSVCCompilerType represents the MSVC compiler.
const (
	CompilerMSVC CompilerType = iota + 100
)

// GCCToMSVCFlags maps GCC/Clang flags to MSVC equivalents.
var GCCToMSVCFlags = map[string]string{
	// Compile-only
	"-c": "/c",

	// Output file (handled specially)
	"-o": "/Fo",

	// Optimization levels
	"-O0":    "/Od",
	"-O1":    "/O1",
	"-O2":    "/O2",
	"-O3":    "/Ox",
	"-Os":    "/O1", // Favor size
	"-Ofast": "/Ox /fp:fast",

	// Debugging
	"-g":  "/Zi",
	"-g0": "",
	"-g1": "/Zi",
	"-g2": "/Zi",
	"-g3": "/Zi",

	// Warnings
	"-Wall":       "/W4",
	"-Wextra":     "/W4",
	"-Werror":     "/WX",
	"-w":          "/W0",
	"-pedantic":   "", // No direct equivalent
	"-Wpedantic":  "", // No direct equivalent
	"-Wno-unused": "", // Handled via pragma in MSVC

	// C/C++ standards
	"-std=c89":   "", // MSVC defaults to C89
	"-std=c99":   "", // Limited support
	"-std=c11":   "/std:c11",
	"-std=c17":   "/std:c17",
	"-std=c++98": "",           // MSVC defaults
	"-std=c++03": "",           // MSVC defaults
	"-std=c++11": "/std:c++14", // MSVC 14 minimum
	"-std=c++14": "/std:c++14",
	"-std=c++17": "/std:c++17",
	"-std=c++20": "/std:c++20",
	"-std=c++23": "/std:c++latest",

	// GNU extensions (disable)
	"-std=gnu89":   "",
	"-std=gnu99":   "",
	"-std=gnu11":   "/std:c11",
	"-std=gnu17":   "/std:c17",
	"-std=gnu++11": "/std:c++14",
	"-std=gnu++14": "/std:c++14",
	"-std=gnu++17": "/std:c++17",
	"-std=gnu++20": "/std:c++20",

	// Preprocessor
	"-E": "/EP", // Preprocess only
	"-P": "/P",  // Preprocess to file

	// Exception handling
	"-fexceptions":    "/EHsc",
	"-fno-exceptions": "/EHs-c-",

	// RTTI
	"-frtti":    "/GR",
	"-fno-rtti": "/GR-",

	// Position independent code (Windows doesn't need it)
	"-fPIC":    "",
	"-fPIE":    "",
	"-fpic":    "",
	"-fpie":    "",
	"-fno-pic": "",

	// Inline functions
	"-finline-functions":    "/Ob2",
	"-fno-inline-functions": "/Ob0",
	"-fno-inline":           "/Ob0",

	// Stack protection
	"-fstack-protector":        "/GS",
	"-fstack-protector-all":    "/GS",
	"-fstack-protector-strong": "/GS",
	"-fno-stack-protector":     "/GS-",

	// Common flags to ignore (no MSVC equivalent needed)
	"-pipe":                "",
	"-pthread":             "",
	"-fdiagnostics-color":  "",
	"-fcolor-diagnostics":  "",
	"-fvisibility=hidden":  "",
	"-fvisibility=default": "",

	// Architecture (handled specially)
	"-m32":          "",
	"-m64":          "",
	"-march=native": "",
}

// TranslateToMSVC translates GCC/Clang compiler arguments to MSVC arguments.
func TranslateToMSVC(args []string) []string {
	var result []string

	i := 0
	for i < len(args) {
		arg := args[i]

		// Handle output file
		if arg == "-o" && i+1 < len(args) {
			i++
			outFile := args[i]
			// Determine if it's an object file output or executable
			if strings.HasSuffix(outFile, ".o") || strings.HasSuffix(outFile, ".obj") {
				result = append(result, "/Fo"+outFile)
			} else {
				result = append(result, "/Fe"+outFile)
			}
			i++
			continue
		}

		if strings.HasPrefix(arg, "-o") {
			outFile := arg[2:]
			if strings.HasSuffix(outFile, ".o") || strings.HasSuffix(outFile, ".obj") {
				result = append(result, "/Fo"+outFile)
			} else {
				result = append(result, "/Fe"+outFile)
			}
			i++
			continue
		}

		// Handle include directories
		if arg == "-I" && i+1 < len(args) {
			i++
			result = append(result, "/I"+args[i])
			i++
			continue
		}

		if strings.HasPrefix(arg, "-I") {
			result = append(result, "/I"+arg[2:])
			i++
			continue
		}

		// Handle defines
		if arg == "-D" && i+1 < len(args) {
			i++
			result = append(result, "/D"+args[i])
			i++
			continue
		}

		if strings.HasPrefix(arg, "-D") {
			result = append(result, "/D"+arg[2:])
			i++
			continue
		}

		// Handle undefines
		if arg == "-U" && i+1 < len(args) {
			i++
			result = append(result, "/U"+args[i])
			i++
			continue
		}

		if strings.HasPrefix(arg, "-U") {
			result = append(result, "/U"+arg[2:])
			i++
			continue
		}

		// Handle language selection
		if arg == "-x" && i+1 < len(args) {
			i++
			lang := args[i]
			switch lang {
			case "c":
				result = append(result, "/Tc")
			case "c++":
				result = append(result, "/Tp")
			}
			i++
			continue
		}

		// Handle specific warning flags
		if strings.HasPrefix(arg, "-Wno-") {
			// MSVC uses /wd<num> for disabling warnings, skip these
			i++
			continue
		}

		if strings.HasPrefix(arg, "-W") {
			// Handle generic -W flags
			if translated, ok := GCCToMSVCFlags[arg]; ok {
				if translated != "" {
					result = append(result, translated)
				}
			}
			i++
			continue
		}

		// Handle -f flags
		if strings.HasPrefix(arg, "-f") {
			if translated, ok := GCCToMSVCFlags[arg]; ok {
				if translated != "" {
					// Some translations have multiple flags
					result = append(result, strings.Fields(translated)...)
				}
			}
			i++
			continue
		}

		// Handle -std flags
		if strings.HasPrefix(arg, "-std=") {
			if translated, ok := GCCToMSVCFlags[arg]; ok {
				if translated != "" {
					result = append(result, translated)
				}
			}
			i++
			continue
		}

		// Handle -m flags (architecture)
		if strings.HasPrefix(arg, "-m") {
			// MSVC uses different toolchains for different architectures
			// Skip these flags as they're handled at the toolchain level
			i++
			continue
		}

		// Handle optimization flags
		if strings.HasPrefix(arg, "-O") {
			if translated, ok := GCCToMSVCFlags[arg]; ok {
				if translated != "" {
					result = append(result, strings.Fields(translated)...)
				}
			}
			i++
			continue
		}

		// Handle debug flags
		if strings.HasPrefix(arg, "-g") {
			if translated, ok := GCCToMSVCFlags[arg]; ok {
				if translated != "" {
					result = append(result, translated)
				}
			}
			i++
			continue
		}

		// Check direct mapping
		if translated, ok := GCCToMSVCFlags[arg]; ok {
			if translated != "" {
				result = append(result, strings.Fields(translated)...)
			}
			i++
			continue
		}

		// If it's a GCC/Clang specific flag we don't recognize, skip it
		if strings.HasPrefix(arg, "-") {
			i++
			continue
		}

		// Pass through source files and other arguments
		result = append(result, arg)
		i++
	}

	return result
}

// MSVCToCLFlags adds common MSVC-specific flags for compatibility.
func MSVCToCLFlags(existing []string) []string {
	// Add standard MSVC flags if not already present
	hasNologo := false
	hasEH := false
	hasPermissive := false

	for _, flag := range existing {
		if flag == "/nologo" {
			hasNologo = true
		}
		if strings.HasPrefix(flag, "/EH") {
			hasEH = true
		}
		if flag == "/permissive-" {
			hasPermissive = true
		}
	}

	result := existing

	// Suppress startup banner
	if !hasNologo {
		result = append(result, "/nologo")
	}

	// Enable C++ exceptions (standard mode)
	if !hasEH {
		result = append(result, "/EHsc")
	}

	// Disable non-standard extensions
	if !hasPermissive {
		result = append(result, "/permissive-")
	}

	return result
}

// IsMSVCCompiler returns true if the compiler is MSVC (cl.exe).
func IsMSVCCompiler(compiler string) bool {
	base := strings.ToLower(compiler)
	return strings.Contains(base, "cl.exe") || base == "cl"
}
