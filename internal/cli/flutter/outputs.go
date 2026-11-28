package flutter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// outputExts are the build outputs written to --output; anything else in
// the worker's archive (logs, mapping files) stays in the archive.
var outputExts = map[string]bool{".apk": true, ".aab": true}

// extractOutputs writes the APK/AAB files of a worker's artifact archive
// into dir and returns their paths. The archive is a plain tar (what
// workers send), or gzip-compressed tar. Entries are written by base name
// only, so a hostile archive cannot escape dir.
func extractOutputs(archive []byte, dir string) ([]string, error) {
	if len(archive) == 0 {
		return nil, nil
	}
	var r io.Reader = bytes.NewReader(archive)
	switch {
	case bytes.HasPrefix(archive, []byte{0x1f, 0x8b}):
		gz, err := gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("failed to open gzip artifacts: %w", err)
		}
		defer gz.Close()
		r = gz
	case bytes.HasPrefix(archive, []byte{0x28, 0xb5, 0x2f, 0xfd}):
		return nil, errors.New("zstd-compressed artifacts are not supported; download them from the dashboard")
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}
	var written []string
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return written, nil
		}
		if err != nil {
			return written, fmt.Errorf("failed to read artifacts: %w", err)
		}
		name := filepath.Base(filepath.FromSlash(hdr.Name))
		if hdr.Typeflag != tar.TypeReg || !outputExts[strings.ToLower(filepath.Ext(name))] {
			continue
		}
		dest := filepath.Join(dir, name)
		if err := writeOutput(dest, tr); err != nil {
			return written, err
		}
		written = append(written, dest)
	}
}

func writeOutput(dest string, r io.Reader) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", dest, err)
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return fmt.Errorf("failed to write %s: %w", dest, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close %s: %w", dest, err)
	}
	return nil
}
