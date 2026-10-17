package executor

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// Exported archive plumbing shared by the flutter executor and the
// plugin-builder execution path (internal/worker/builders). The wire
// format matches what the clients already speak: plain tar out, and
// tar or zstd-compressed tar in (zstd via the external binary, same
// as the flutter executor has always done).

// ExtractSourceArchive unpacks a source archive (tar, or zstd-compressed
// tar) into dest. Empty data is a no-op: a builder may legitimately
// receive no source.
func ExtractSourceArchive(dest string, data []byte) error {
	return extractSourceArchive(dest, data)
}

// ArchivePaths packs the given root-relative paths into a plain tar and
// returns the bytes plus per-artifact metadata (sha256, size) suitable
// for BuildResponse.ArtifactList. Paths are validated to stay inside
// root: a plugin's declared outputs can never exfiltrate files outside
// its work directory.
func ArchivePaths(root string, relPaths []string) ([]byte, []*pb.ArtifactInfo, error) {
	var (
		buf       bytes.Buffer
		tarWriter = tar.NewWriter(&buf)
		infos     []*pb.ArtifactInfo
	)
	defer tarWriter.Close()

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, nil, err
	}

	for _, rel := range relPaths {
		absPath := filepath.Join(absRoot, rel)
		if !hasRootPrefix(absPath, absRoot) {
			return nil, nil, fmt.Errorf("output path %q escapes the work directory", rel)
		}

		checksum, size, err := checksumFile(absPath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to hash output %q: %w", rel, err)
		}

		if err := writeTarEntry(tarWriter, absRoot, absPath); err != nil {
			return nil, nil, err
		}

		infos = append(infos, &pb.ArtifactInfo{
			Name:      filepath.ToSlash(rel),
			Path:      filepath.ToSlash(rel),
			SizeBytes: size,
			Checksum:  checksum,
		})
	}

	if err := tarWriter.Close(); err != nil {
		return nil, nil, err
	}
	return buf.Bytes(), infos, nil
}

// hasRootPrefix reports whether path is root or inside root,
// symlinks not followed (the extract path already rejects links).
func hasRootPrefix(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel != ".." && !filepath.IsAbs(rel) && !bytes.HasPrefix([]byte(rel), []byte(".."+string(os.PathSeparator)))
}

// writeTarEntry streams one file into the tar.
func writeTarEntry(tarWriter *tar.Writer, root, absPath string) error {
	file, err := os.Open(absPath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		// Directory outputs are packed implicitly by their files;
		// recurse via filepath.Walk by the caller listing files.
		return fmt.Errorf("output %q is a directory; list its files instead", absPath)
	}

	relPath, err := filepath.Rel(root, absPath)
	if err != nil {
		return err
	}

	header := &tar.Header{
		Name:    filepath.ToSlash(relPath),
		Mode:    int64(info.Mode().Perm()),
		Size:    info.Size(),
		ModTime: info.ModTime().UTC().Truncate(time.Second),
	}
	if err := tarWriter.WriteHeader(header); err != nil {
		return err
	}
	if _, err := io.Copy(tarWriter, file); err != nil {
		return err
	}
	return nil
}
