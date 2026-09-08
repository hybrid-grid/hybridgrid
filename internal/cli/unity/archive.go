package unity

import (
	"strings"
)

// shouldExclude reports whether relPath should be left out of the source
// archive: Unity's generated/cache directories and IDE project files never
// affect the remote build result.
func shouldExclude(relPath string, isDir bool) bool {
	relPath = strings.TrimPrefix(relPath, "./")
	if relPath == "" {
		return false
	}

	parts := strings.Split(relPath, "/")
	for _, part := range parts {
		switch part {
		case "Library", "Temp", "Build", "Logs", ".git", "obj":
			return true
		}
	}

	if isDir {
		return false
	}

	if strings.HasSuffix(relPath, ".csproj") {
		return true
	}
	if strings.HasSuffix(relPath, ".sln") {
		return true
	}
	if strings.HasSuffix(relPath, ".unitypackage") {
		return true
	}

	return false
}
