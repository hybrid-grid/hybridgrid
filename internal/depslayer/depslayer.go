// Package depslayer packs a project's installed dependencies (every
// node_modules tree) into one archive that is shipped to a worker once
// and reused by every later task, instead of riding along with each
// task's source archive.
//
// The layer is keyed by a hash of file metadata (path, mode, size,
// modification time, symlink target), which is cheap to compute on every
// run without reading file contents. Symlinks are kept, because package
// manager workspaces (npm, yarn, pnpm) link packages into node_modules
// with them; links that point outside the layer's root are rejected on
// extraction.
package depslayer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MissingMessage is the gRPC FailedPrecondition message a coordinator or
// worker returns when a request names a layer by hash without carrying
// it and the layer is not cached there; the sender then retries with
// the archive attached.
const MissingMessage = "dependency layer not cached"

const dirName = "node_modules"

// ErrNoDeps means the project has no node_modules tree.
var ErrNoDeps = errors.New("no node_modules found")

// entry is one filesystem object of the layer.
type entry struct {
	rel  string // slash-separated, relative to the project root
	info fs.FileInfo
	link string // symlink target, for symlinks
}

// Roots returns the project-relative paths of every node_modules
// directory under root, skipping nested ones (node_modules inside
// node_modules belong to their parent tree) and VCS directories.
func Roots(root string) ([]string, error) {
	var roots []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() || path == root {
			return nil
		}
		switch d.Name() {
		case ".git", ".hg", ".svn":
			return filepath.SkipDir
		case dirName:
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			roots = append(roots, filepath.ToSlash(rel))
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find node_modules: %w", err)
	}
	sort.Strings(roots)
	return roots, nil
}

func collect(root string, roots []string) ([]entry, error) {
	var out []entry
	for _, r := range roots {
		base := filepath.Join(root, filepath.FromSlash(r))
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			e := entry{rel: filepath.ToSlash(rel), info: info}
			if d.Type()&fs.ModeSymlink != 0 {
				if e.link, err = os.Readlink(path); err != nil {
					return err
				}
			}
			out = append(out, e)
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", r, err)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out, nil
}

// Layer describes the dependency trees of a project.
type Layer struct {
	Root    string
	Roots   []string
	Hash    string
	entries []entry
}

// Scan finds the node_modules trees under root and hashes their
// metadata. It returns ErrNoDeps when there are none.
func Scan(root string) (*Layer, error) {
	roots, err := Roots(root)
	if err != nil {
		return nil, err
	}
	if len(roots) == 0 {
		return nil, ErrNoDeps
	}
	entries, err := collect(root, roots)
	if err != nil {
		return nil, err
	}
	h := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(h, "%s\x00%o\x00%d\x00%d\x00%s\n",
			e.rel, e.info.Mode(), e.info.Size(), e.info.ModTime().UnixNano(), e.link)
	}
	return &Layer{Root: root, Roots: roots, Hash: hex.EncodeToString(h.Sum(nil)), entries: entries}, nil
}

// Archive returns the layer as a gzip-compressed tar.
func (l *Layer) Archive() ([]byte, error) {
	var buf bytes.Buffer
	if err := l.write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (l *Layer) write(w io.Writer) error {
	gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
	if err != nil {
		return err
	}
	tw := tar.NewWriter(gz)
	for _, e := range l.entries {
		hdr, err := tar.FileInfoHeader(e.info, e.link)
		if err != nil {
			return fmt.Errorf("failed to archive %s: %w", e.rel, err)
		}
		hdr.Name = e.rel
		if e.info.IsDir() {
			hdr.Name += "/"
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if e.info.Mode().IsRegular() {
			if err := copyFile(tw, filepath.Join(l.Root, filepath.FromSlash(e.rel))); err != nil {
				return fmt.Errorf("failed to archive %s: %w", e.rel, err)
			}
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func copyFile(w io.Writer, path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

// ExtractFile extracts a layer archive stored at path into dest.
func ExtractFile(path, dest string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open dependency layer: %w", err)
	}
	defer f.Close()
	return Extract(f, dest)
}

// Extract unpacks a layer archive into dest. Entries and symlink
// targets must stay inside dest.
func Extract(r io.Reader, dest string) error {
	dest, err := filepath.Abs(dest)
	if err != nil {
		return err
	}
	gz, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("failed to open dependency layer: %w", err)
	}
	defer gz.Close()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read dependency layer: %w", err)
		}
		target, err := within(dest, hdr.Name)
		if err != nil {
			return err
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := writeFile(target, tr, hdr.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if filepath.IsAbs(hdr.Linkname) {
				return fmt.Errorf("dependency layer: absolute symlink %s -> %s", hdr.Name, hdr.Linkname)
			}
			resolved := filepath.Join(filepath.Dir(target), filepath.FromSlash(hdr.Linkname))
			if !inside(dest, resolved) {
				return fmt.Errorf("dependency layer: symlink %s escapes the project", hdr.Name)
			}
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
}

func writeFile(path string, r io.Reader, perm fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm|0o600)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// within resolves an archive entry name under dest, rejecting escapes.
func within(dest, name string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("dependency layer: invalid path %q", name)
	}
	target := filepath.Join(dest, clean)
	if !inside(dest, target) {
		return "", fmt.Errorf("dependency layer: invalid path %q", name)
	}
	return target, nil
}

func inside(base, path string) bool {
	rel, err := filepath.Rel(base, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}
