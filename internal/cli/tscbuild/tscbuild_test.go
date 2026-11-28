package tscbuild

import (
	"archive/tar"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/depslayer"
)

func write(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		p := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// monorepo: app -> (lib, util), lib -> util; other is unrelated.
func monorepo(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	write(t, root, map[string]string{
		"tsconfig.json": `{
			// solution file
			"files": [],
			"references": [{ "path": "packages/app" }, { "path": "packages/other/tsconfig.json" },],
		}`,
		"tsconfig.base.json":             `{"compilerOptions": {"composite": true}}`,
		"packages/app/tsconfig.json":     `{"references": [{"path": "../lib"}, {"path": "../util"}], "include": ["src"]}`,
		"packages/app/src/main.ts":       "export {}",
		"packages/lib/tsconfig.json":     `{"references": [{"path": "../util"}], "include": ["src"]}`,
		"packages/lib/src/index.ts":      "export {}",
		"packages/util/tsconfig.json":    `{"include": ["src"]}`,
		"packages/util/src/index.ts":     "export {}",
		"packages/other/tsconfig.json":   `{"include": ["src"]}`,
		"packages/other/src/index.ts":    "export {}",
		"packages/app/dist/old.js":       "stale",
		"node_modules/typescript/tsc.js": "compiler",
	})
	return root
}

func TestStripJSONC(t *testing.T) {
	in := `{
		// line comment with "quotes"
		"a": "http://x/*not a comment*/", /* block */
		"b": [1, 2,],
		"c": "escaped \" quote, }",
	}`
	var got map[string]any
	if err := json.Unmarshal(stripJSONC([]byte(in)), &got); err != nil {
		t.Fatalf("%v\n%s", err, stripJSONC([]byte(in)))
	}
	if got["a"] != "http://x/*not a comment*/" || got["c"] != `escaped " quote, }` || len(got["b"].([]any)) != 2 {
		t.Fatalf("got %v", got)
	}
}

func configs(g *Graph) []string {
	var out []string
	for _, p := range g.Projects {
		out = append(out, p.Config)
	}
	return out
}

func TestLoadFollowsReferencesInBuildOrder(t *testing.T) {
	root := monorepo(t)
	g, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.Join(configs(g), " ")
	want := "packages/util/tsconfig.json packages/lib/tsconfig.json packages/app/tsconfig.json packages/other/tsconfig.json"
	if got != want {
		t.Fatalf("order = %s\nwant    %s", got, want)
	}
	app := g.Projects[2]
	if strings.Join(app.Refs, ",") != "packages/lib/tsconfig.json,packages/util/tsconfig.json" {
		t.Fatalf("app refs = %v", app.Refs)
	}

	// One project of the workspace: it is built with what it references,
	// relative to the workspace root (found by its node_modules).
	g, err = Load(filepath.Join(root, "packages/lib"))
	if err != nil {
		t.Fatal(err)
	}
	if g.Root != root || strings.Join(configs(g), " ") != "packages/util/tsconfig.json packages/lib/tsconfig.json" {
		t.Fatalf("lib graph root %s = %v", g.Root, configs(g))
	}
}

func TestLoadRejectsCycles(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"a/tsconfig.json": `{"references": [{"path": "../b"}]}`,
		"b/tsconfig.json": `{"references": [{"path": "../a"}]}`,
	})
	if _, err := Load(filepath.Join(root, "a")); err == nil || !strings.Contains(err.Error(), "cycle") {
		t.Fatalf("cycle: %v", err)
	}
}

func TestWorkspaceRootWidensToContainEveryProject(t *testing.T) {
	root := t.TempDir()
	write(t, root, map[string]string{
		"apps/web/tsconfig.json": `{"references": [{"path": "../../libs/ui"}], "include": ["src"]}`,
		"libs/ui/tsconfig.json":  `{"include": ["src"]}`,
		"apps/web/package.json":  `{}`,
	})
	// No marker anywhere: the root is the smallest directory holding both.
	g, err := Load(filepath.Join(root, "apps/web"))
	if err != nil {
		t.Fatal(err)
	}
	if g.Root != root {
		t.Fatalf("root = %s, want %s", g.Root, root)
	}
}

func TestExcludeKeepsOnlyTheClosure(t *testing.T) {
	g, err := Load(monorepo(t))
	if err != nil {
		t.Fatal(err)
	}
	lib := g.Projects[1]
	ex := g.excludeFor(lib)
	tests := []struct {
		rel   string
		isDir bool
		want  bool
	}{
		{"packages/lib", true, false},
		{"packages/util", true, false},
		{"packages/app", true, true},
		{"packages/other", true, true},
		{"packages", true, false},
		{"node_modules", true, true},
		{".git", true, true},
		{"tsconfig.base.json", false, false},
	}
	for _, tt := range tests {
		if got := ex(tt.rel, tt.isDir); got != tt.want {
			t.Errorf("exclude(%q) = %v, want %v", tt.rel, got, tt.want)
		}
	}
}

// sent is what one Build call carried (the client reuses the request
// when it retries, so it is copied out at call time).
type sent struct {
	buildType pb.BuildType
	depsHash  string
	withLayer bool
}

// fakeClient builds projects instantly, recording order and requests.
type fakeClient struct {
	mu       sync.Mutex
	order    []string
	requests []sent
	fail     string // project config that fails type-checking
	haveDeps bool   // coordinator already holds the layer
}

func outputs(t *testing.T, files map[string]string) []byte {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range files {
		_ = tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content)), Typeflag: tar.TypeReg})
		_, _ = tw.Write([]byte(content))
	}
	_ = tw.Close()
	return buf.Bytes()
}

func (c *fakeClient) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	project := req.GetNodeConfig().GetTscProject()
	c.mu.Lock()
	c.requests = append(c.requests, sent{req.BuildType, req.DepsHash, len(req.DepsArchive) > 0})
	if req.DepsHash != "" && len(req.DepsArchive) == 0 && !c.haveDeps {
		c.mu.Unlock()
		return nil, status.Error(codes.FailedPrecondition, depslayer.MissingMessage)
	}
	if len(req.DepsArchive) > 0 {
		c.haveDeps = true
	}
	c.order = append(c.order, project)
	c.mu.Unlock()
	time.Sleep(5 * time.Millisecond)
	if project == c.fail {
		return &pb.BuildResponse{Status: pb.TaskStatus_STATUS_FAILED, ExitCode: 2, Stdout: "error TS2322"}, nil
	}
	dir := filepath.ToSlash(filepath.Dir(project))
	return &pb.BuildResponse{
		Status:    pb.TaskStatus_STATUS_COMPLETED,
		WorkerId:  "w-" + filepath.Base(dir),
		Artifacts: outputs(nil, map[string]string{dir + "/dist/index.js": "built " + dir}),
	}, nil
}

func indexOf(list []string, s string) int {
	for i, v := range list {
		if v == s {
			return i
		}
	}
	return -1
}

func TestRunRespectsReferencesAndWritesOutputs(t *testing.T) {
	root := monorepo(t)
	g, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	c := &fakeClient{}
	results, err := Run(context.Background(), c, g, Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range results {
		if r.Err != nil || r.Skipped || r.Files != 1 {
			t.Fatalf("result %s: %+v", r.Project.Config, r)
		}
	}
	util, lib, app := indexOf(c.order, "packages/util/tsconfig.json"), indexOf(c.order, "packages/lib/tsconfig.json"), indexOf(c.order, "packages/app/tsconfig.json")
	if !(util < lib && lib < app) {
		t.Fatalf("build order %v", c.order)
	}
	if got, _ := os.ReadFile(filepath.Join(root, "packages/app/dist/index.js")); string(got) != "built packages/app" {
		t.Fatalf("app output = %q", got)
	}

	// The layer was attached once, after the coordinator reported it
	// missing; every request named it by hash.
	attached := 0
	for _, req := range c.requests {
		if req.depsHash == "" || req.buildType != pb.BuildType_BUILD_TYPE_NODEJS {
			t.Fatalf("request %+v", req)
		}
		if req.withLayer {
			attached++
		}
	}
	if attached < 1 || attached > len(g.Projects) || len(c.requests) != len(g.Projects)+attached {
		t.Fatalf("%d requests, %d with the layer", len(c.requests), attached)
	}
}

func TestRunSkipsDependentsOfAFailure(t *testing.T) {
	g, err := Load(monorepo(t))
	if err != nil {
		t.Fatal(err)
	}
	c := &fakeClient{fail: "packages/lib/tsconfig.json", haveDeps: true}
	results, err := Run(context.Background(), c, g, Options{})
	if err != nil {
		t.Fatal(err)
	}
	got := map[string]string{}
	for _, r := range results {
		switch {
		case r.Skipped:
			got[r.Project.Config] = "skipped"
		case r.Err != nil:
			got[r.Project.Config] = "failed: " + r.Output
		default:
			got[r.Project.Config] = "ok"
		}
	}
	want := map[string]string{
		"packages/util/tsconfig.json":  "ok",
		"packages/lib/tsconfig.json":   "failed: error TS2322",
		"packages/app/tsconfig.json":   "skipped",
		"packages/other/tsconfig.json": "ok",
	}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("%s = %q, want %q", k, got[k], v)
		}
	}
}

func TestExtractRejectsEscapingOutputs(t *testing.T) {
	root := t.TempDir()
	if _, err := extract(outputs(t, map[string]string{"../evil.js": "x"}), root); err == nil {
		t.Fatal("expected an error")
	}
	if n, err := extract(outputs(t, map[string]string{"a/dist/x.js": "x"}), root); err != nil || n != 1 {
		t.Fatalf("extract = %d, %v", n, err)
	}
}

func TestUnreachableOnlyWhenNothingReachedTheCoordinator(t *testing.T) {
	down := status.Error(codes.Unavailable, "down")
	if unreachable([]Result{{Err: down}, {Skipped: true}}) == nil {
		t.Fatal("all unavailable must fall back")
	}
	if unreachable([]Result{{Err: down}, {}}) != nil {
		t.Fatal("a project reached the coordinator; no fallback")
	}
}
