package validation

import (
	"path/filepath"
	"runtime"
	"strings"
)

// WindowsReservedNames are device names that cannot be used as filenames on Windows.
var WindowsReservedNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

// WindowsInvalidChars are characters that cannot be used in Windows filenames.
var WindowsInvalidChars = []byte{'<', '>', ':', '"', '|', '?', '*'}

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

	// On Windows, validate for reserved names and invalid characters
	if runtime.GOOS == "windows" {
		if errMsg := ValidatePathForWindows(cleaned); errMsg != "" {
			return ""
		}
	}

	// Handle absolute paths
	if filepath.IsAbs(cleaned) {
		// If path is already absolute, verify it's within basePath
		if basePath != "" && !pathStartsWithBase(cleaned, basePath) {
			return ""
		}
		return cleaned
	}

	// Resolve relative path against base
	if basePath != "" {
		abs := filepath.Join(basePath, cleaned)
		abs = filepath.Clean(abs)
		if !pathStartsWithBase(abs, basePath) {
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
	// Normalize path separators for cross-platform support
	// Convert backslashes to forward slashes before checking
	normalizedPath := filepath.ToSlash(path)

	// Check for .. segments using both separators
	parts := strings.Split(normalizedPath, "/")
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

// isWindowsReservedName checks if the given name is a Windows reserved device name.
func isWindowsReservedName(name string) bool {
	// Strip extension if present (CON.txt is still reserved)
	base := strings.ToUpper(name)
	if idx := strings.LastIndex(base, "."); idx != -1 {
		base = base[:idx]
	}

	for _, reserved := range WindowsReservedNames {
		if base == reserved {
			return true
		}
	}
	return false
}

// hasWindowsInvalidChars checks if the path contains characters invalid on Windows.
// Note: Colons are allowed as part of drive letters (e.g., "C:\") but not elsewhere.
func hasWindowsInvalidChars(path string) bool {
	for i, r := range path {
		// Allow colon only at position 1 (drive letter, e.g., "C:")
		if r == ':' {
			if i != 1 {
				return true
			}
			continue
		}
		for _, c := range WindowsInvalidChars {
			if r == rune(c) {
				return true
			}
		}
	}
	return false
}

// ValidatePathForWindows checks if a path is valid on Windows.
// Returns an error message if invalid, empty string if valid.
func ValidatePathForWindows(path string) string {
	if path == "" {
		return ""
	}

	// Check for invalid characters
	if hasWindowsInvalidChars(path) {
		return "path contains invalid Windows characters"
	}

	// Check each component for reserved names
	normalizedPath := filepath.ToSlash(path)
	parts := strings.Split(normalizedPath, "/")
	for _, part := range parts {
		if part == "" {
			continue
		}
		if isWindowsReservedName(part) {
			return "path contains Windows reserved name: " + part
		}
	}

	return ""
}

// pathStartsWithBase checks if fullPath starts with basePath, using case-insensitive
// comparison on Windows.
func pathStartsWithBase(fullPath, basePath string) bool {
	fullPath = filepath.Clean(fullPath)
	basePath = filepath.Clean(basePath)

	if runtime.GOOS == "windows" {
		// Windows paths are case-insensitive
		return strings.HasPrefix(strings.ToLower(fullPath), strings.ToLower(basePath))
	}
	return strings.HasPrefix(fullPath, basePath)
}
