package flutter

import (
	"archive/tar"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

func createSourceArchive(projectPath string) ([]byte, string, error) {
	root, err := filepath.Abs(projectPath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to resolve project path: %w", err)
	}

	info, err := os.Stat(root)
	if err != nil {
		return nil, "", fmt.Errorf("project not found: %w", err)
	}
	if !info.IsDir() {
		return nil, "", fmt.Errorf("project must be a directory")
	}

	var buf bytes.Buffer
	tarWriter := tar.NewWriter(&buf)

	walkErr := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)

		if shouldExclude(rel, entry.IsDir()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if entry.IsDir() {
			return nil
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		info, err := file.Stat()
		if err != nil {
			_ = file.Close()
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			_ = file.Close()
			return err
		}
		header.Name = rel

		if err := tarWriter.WriteHeader(header); err != nil {
			_ = file.Close()
			return err
		}

		if _, err := io.Copy(tarWriter, file); err != nil {
			_ = file.Close()
			return err
		}

		if err := file.Close(); err != nil {
			return err
		}

		return nil
	})

	if walkErr != nil {
		_ = tarWriter.Close()
		return nil, "", fmt.Errorf("failed to archive project: %w", walkErr)
	}

	if err := tarWriter.Close(); err != nil {
		return nil, "", fmt.Errorf("failed to finalize archive: %w", err)
	}

	data := buf.Bytes()
	if len(data) > validation.MaxSourceArchiveBytes {
		return nil, "", fmt.Errorf("source archive too large: %d bytes", len(data))
	}

	sum := sha256.Sum256(data)
	return data, hex.EncodeToString(sum[:]), nil
}

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
