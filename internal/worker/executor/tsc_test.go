package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func tarFiles(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// fakeTsc is a tsc stand-in: `tsc -p <project>` writes dist/index.js
// next to the project, or fails with a diagnostic when the source
// contains "ERROR".
func fakeTsc(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tsc")
	script := `#!/bin/sh
project="$2"
dir=$(dirname "$project")
if grep -q ERROR "$dir/src/index.ts"; then echo "$dir/src/index.ts(1,1): error TS2322: nope"; exit 2; fi
mkdir -p "$dir/dist" && echo "compiled" > "$dir/dist/index.js"
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func archiveNames(t *testing.T, data []byte) []string {
	t.Helper()
	var names []string
	tr := tar.NewReader(bytes.NewReader(data))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return names
		}
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, hdr.Name)
	}
}

func TestTscExecutor_ReturnsOnlyEmittedFiles(t *testing.T) {
	e := &TscExecutor{node: "node", tsc: fakeTsc(t)}
	res, err := e.Execute(context.Background(), &Request{
		TaskID:     "t1",
		NodeConfig: &pb.NodeConfig{TscProject: "packages/lib/tsconfig.json"},
		SourceArchive: tarFiles(t, map[string]string{
			"packages/lib/tsconfig.json": "{}",
			"packages/lib/src/index.ts":  "export {}",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Fatalf("failed: %s %s", res.Stdout, res.Stderr)
	}
	if got := strings.Join(archiveNames(t, res.ArtifactArchive), ","); got != "packages/lib/dist/index.js" {
		t.Fatalf("outputs = %s", got)
	}
	if len(res.Artifacts) != 1 || res.Artifacts[0].Checksum == "" {
		t.Fatalf("artifact list = %+v", res.Artifacts)
	}
}

func TestTscExecutor_TypeErrorsFailWithDiagnostics(t *testing.T) {
	e := &TscExecutor{node: "node", tsc: fakeTsc(t)}
	res, err := e.Execute(context.Background(), &Request{
		TaskID:     "t2",
		NodeConfig: &pb.NodeConfig{TscProject: "lib/tsconfig.json"},
		SourceArchive: tarFiles(t, map[string]string{
			"lib/tsconfig.json": "{}",
			"lib/src/index.ts":  "ERROR",
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Success || res.ExitCode != 2 || !strings.Contains(res.Stdout, "TS2322") || len(res.ArtifactArchive) != 0 {
		t.Fatalf("result = %+v", res)
	}
}

func TestTscExecutor_RejectsBadProjects(t *testing.T) {
	e := NewTscExecutor()
	for _, p := range []string{"", "../outside/tsconfig.json", "/abs/tsconfig.json", "missing/tsconfig.json"} {
		_, err := e.Execute(context.Background(), &Request{TaskID: "t", NodeConfig: &pb.NodeConfig{TscProject: p}})
		if err == nil {
			t.Errorf("project %q accepted", p)
		}
	}
}

func TestTscExecutor_PrefersProjectTypeScript(t *testing.T) {
	work := t.TempDir()
	e := &TscExecutor{node: "node", tsc: "tsc"}
	if got := strings.Join(e.command(work, "packages/app/tsconfig.json"), " "); got != "tsc -p packages/app/tsconfig.json" {
		t.Fatalf("without node_modules: %s", got)
	}
	bin := filepath.Join(work, "node_modules", "typescript", "bin", "tsc")
	if err := os.MkdirAll(filepath.Dir(bin), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bin, []byte(""), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(work, "packages/app"), 0o755); err != nil {
		t.Fatal(err)
	}
	want := "node " + bin + " -p packages/app/tsconfig.json"
	if got := strings.Join(e.command(work, "packages/app/tsconfig.json"), " "); got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
}
