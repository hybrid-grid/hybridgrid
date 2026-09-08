package flutter

import (
	"strings"
)

// shouldExclude reports whether relPath should be left out of the source
// archive: build outputs, tool caches, and VCS state never affect the
// remote build result.
func shouldExclude(relPath string, isDir bool) bool {
	relPath = strings.TrimPrefix(relPath, "./")
	if relPath == "" {
		return false
	}

	parts := strings.Split(relPath, "/")
	if len(parts) > 0 {
		switch parts[0] {
		case "build", ".dart_tool", ".git":
			return true
		}
	}

	if strings.HasPrefix(relPath, "android/.gradle") {
		return true
	}
	if strings.HasPrefix(relPath, "android/app/build/") {
		return true
	}
	if strings.HasPrefix(relPath, "ios/Pods") {
		return true
	}

	if !isDir && strings.HasSuffix(relPath, ".DS_Store") {
		return true
	}

	return false
}
