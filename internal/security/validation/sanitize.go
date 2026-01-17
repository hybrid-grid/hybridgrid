package validation

import (
	"path/filepath"
	"strings"
)

// DangerousFlags are compiler flags that should be blocked.
var DangerousFlags = []string{
	"--plugin",
	"-fplugin",
	"-B",                 // Specify toolchain directory
	"-specs",             // Spec files
	"--sysroot",          // System root override
	"-Xlinker",           // Linker passthrough (limited)
	"-Wl,--wrap",         // Function wrapping
	"-Wl,--defsym",       // Symbol definition
	"-fprofile-generate", // Profile generation can write files
	"-fprofile-use",      // Profile use reads external files
	"-frepo",             // Repository files
	"-save-temps",        // Save intermediate files
	"@",                  // Response files can include arbitrary content
}

// DangerousFlagPrefixes are flag prefixes that should be blocked.
var DangerousFlagPrefixes = []string{
	"-fplugin=",
	"-fplugin-arg-",
	"-specs=",
	"--sysroot=",
	"-B/", "-B./", "-B..", // Toolchain directory
}

// ShellMetaCharacters that could indicate injection attempts.
var ShellMetaCharacters = []byte{
	';', '|', '&', '$', '`', '(', ')', '{', '}', '[', ']',
	'<', '>', '\n', '\r',
}

// SanitizeCompilerArgs filters and sanitizes compiler arguments.
// Returns sanitized args and a list of removed dangerous args.
func SanitizeCompilerArgs(args []string) (sanitized []string, removed []string) {
	for i := 0; i < len(args); i++ {
		arg := args[i]

		// Check for dangerous flags
		if isDangerousFlag(arg) {
			removed = append(removed, arg)
			// Some flags take a value as next arg
			if needsValueArg(arg) && i+1 < len(args) {
				removed = append(removed, args[i+1])
				i++
			}
			continue
		}

		// Check for shell metacharacters
		if hasShellMetaChars(arg) {
			removed = append(removed, arg)
			continue
		}

		// Check for path traversal in include paths
		if strings.HasPrefix(arg, "-I") {
			path := strings.TrimPrefix(arg, "-I")
			if containsPathTraversal(path) {
				removed = append(removed, arg)
				continue
			}
		}

		// Check for output redirection attempts
		if strings.Contains(arg, ">>") || strings.Contains(arg, "> ") {
			removed = append(removed, arg)
			continue
		}

		sanitized = append(sanitized, arg)
	}

	return sanitized, removed
}

// SanitizePath validates and sanitizes a file path.
// Returns empty string if the path is invalid or attempts traversal.
func SanitizePath(basePath, path string) string {
	if path == "" {
		return ""
	}

	// Clean the path
	cleaned := filepath.Clean(path)

	// Check for path traversal
	if containsPathTraversal(cleaned) {
		return ""
	}

	// Handle absolute paths
	if filepath.IsAbs(cleaned) {
		// If path is already absolute, verify it's within basePath
		if basePath != "" && !strings.HasPrefix(cleaned, basePath) {
			return ""
		}
		return cleaned
	}

	// Resolve relative path against base
	if basePath != "" {
		abs := filepath.Join(basePath, cleaned)
		abs = filepath.Clean(abs)
		if !strings.HasPrefix(abs, basePath) {
			return ""
		}
		return abs
	}

	return cleaned
}

// ValidateDockerImage checks if a Docker image name is valid.
func ValidateDockerImage(image string) bool {
	if image == "" {
		return true // Optional field
	}

	// Basic validation - no shell metacharacters
	if hasShellMetaChars(image) {
		return false
	}

	// Check for valid image name format
	// Allow: registry/repo:tag, repo:tag, repo
	validChars := func(r rune) bool {
		return (r >= 'a' && r <= 'z') ||
			(r >= 'A' && r <= 'Z') ||
			(r >= '0' && r <= '9') ||
			r == '.' || r == '-' || r == '_' || r == '/' || r == ':' || r == '@'
	}

	for _, r := range image {
		if !validChars(r) {
			return false
		}
	}

	return true
}

func isDangerousFlag(arg string) bool {
	// Check exact matches
	for _, dangerous := range DangerousFlags {
		if arg == dangerous {
			return true
		}
	}

	// Check prefixes
	for _, prefix := range DangerousFlagPrefixes {
		if strings.HasPrefix(arg, prefix) {
			return true
		}
	}

	return false
}

func needsValueArg(arg string) bool {
	// Flags that take a separate value argument
	needsNext := []string{
		"--plugin",
		"-B",
		"-specs",
		"--sysroot",
	}
	for _, flag := range needsNext {
		if arg == flag {
			return true
		}
	}
	return false
}

func hasShellMetaChars(s string) bool {
	for _, c := range ShellMetaCharacters {
		if strings.ContainsRune(s, rune(c)) {
			return true
		}
	}
	return false
}

func containsPathTraversal(path string) bool {
	// Check for .. segments
	parts := strings.Split(path, string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}

	// Also check for URL-encoded traversal
	if strings.Contains(path, "%2e%2e") || strings.Contains(path, "%2E%2E") {
		return true
	}

	return false
}
