package depslayer

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// workspace builds a small npm-workspaces layout:
//
//	node_modules/typescript/bin/tsc
//	node_modules/@demo/lib -> ../../packages/lib   (workspace link)
//	packages/lib/package.json
//	packages/app/node_modules/left-pad/index.js     (nested tree)
func workspace(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	files := map[string]string{
		"node_modules/typescript/bin/tsc":             "#!/usr/bin/env node\n",
		"packages/lib/package.json":                   `{"name":"@demo/lib"}`,
		"packages/app/node_modules/left-pad/index.js": "module.exports = 1",
		"packages/app/src/index.ts":                   "export {}",
	}
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(root, "node_modules", "@demo"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("../../packages/lib", filepath.Join(root, "node_modules", "@demo", "lib")); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestScanFindsEveryTreeAndHashesMetadata(t *testing.T) {
	root := workspace(t)
	l, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(l.Roots, ",") != "node_modules,packages/app/node_modules" {
		t.Fatalf("roots = %v", l.Roots)
	}
	again, _ := Scan(root)
	if again.Hash != l.Hash {
		t.Fatal("hash not stable")
	}
	later := time.Now().Add(time.Hour)
	if err := os.Chtimes(filepath.Join(root, "node_modules/typescript/bin/tsc"), later, later); err != nil {
		t.Fatal(err)
	}
	if changed, _ := Scan(root); changed.Hash == l.Hash {
		t.Fatal("hash ignores a modified file")
	}
	if _, err := Scan(t.TempDir()); !errors.Is(err, ErrNoDeps) {
		t.Fatalf("empty project: %v", err)
	}
}

func TestArchiveRoundTripKeepsWorkspaceLinks(t *testing.T) {
	root := workspace(t)
	l, err := Scan(root)
	if err != nil {
		t.Fatal(err)
	}
	data, err := l.Archive()
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	// The source tree is extracted next to the layer, as on a worker.
	if err := os.MkdirAll(filepath.Join(dest, "packages/lib"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dest, "packages/lib/package.json"), []byte("lib"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := Extract(bytes.NewReader(data), dest); err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(filepath.Join(dest, "node_modules/@demo/lib/package.json")); string(got) != "lib" {
		t.Fatalf("workspace link not resolved: %q", got)
	}
	if _, err := os.Stat(filepath.Join(dest, "packages/app/node_modules/left-pad/index.js")); err != nil {
		t.Fatal("nested tree missing")
	}
	if _, err := os.Stat(filepath.Join(dest, "packages/app/src/index.ts")); err == nil {
		t.Fatal("source files leaked into the layer")
	}
}

func layerOf(t *testing.T, hdrs ...*tar.Header) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, h := range hdrs {
		if err := tw.WriteHeader(h); err != nil {
			t.Fatal(err)
		}
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

func TestExtractRejectsEscapes(t *testing.T) {
	tests := map[string]*tar.Header{
		"dotdot path":   {Name: "../evil", Typeflag: tar.TypeReg, Mode: 0o644},
		"absolute path": {Name: "/tmp/evil", Typeflag: tar.TypeReg, Mode: 0o644},
		"absolute link": {Name: "node_modules/x", Typeflag: tar.TypeSymlink, Linkname: "/etc"},
		"escaping link": {Name: "node_modules/x", Typeflag: tar.TypeSymlink, Linkname: "../../.."},
	}
	for name, hdr := range tests {
		t.Run(name, func(t *testing.T) {
			if err := Extract(bytes.NewReader(layerOf(t, hdr)), t.TempDir()); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}
