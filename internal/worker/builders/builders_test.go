package builders

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// The tests use the helper-process pattern: the test binary re-invokes
// itself as the plugin (HG_BUILDER_FIXTURE=1), so the fixtures run on
// every OS CI covers without shipping shell scripts.

func TestMain(m *testing.M) {
	if os.Getenv("HG_BUILDER_FIXTURE") == "1" {
		fixtureMain()
		os.Exit(0)
	}
	os.Exit(m.Run())
}

// fixtureMain implements the builderv1alpha1 contract for the test
// binary. Behavior is selected with HG_BUILDER_MODE.
func fixtureMain() {
	mode := os.Getenv("HG_BUILDER_MODE")
	switch os.Args[1] {
	case "capability":
		switch mode {
		case "wrongapi":
			fmt.Println(`{"id":"fixture","api":"builderv9"}`)
		case "wrongid":
			fmt.Println(`{"id":"not-the-name","api":"builderv1alpha1"}`)
		case "broken":
			fmt.Println("this is not json")
		default:
			fmt.Println(`{"id":"fixture","api":"builderv1alpha1","description":"test fixture builder"}`)
		}
	case "execute":
		var task Task
		if err := json.NewDecoder(os.Stdin).Decode(&task); err != nil {
			fmt.Fprintln(os.Stderr, "fixture: bad task json:", err)
			os.Exit(2)
		}
		fmt.Fprintf(os.Stderr, "fixture console: task %s args %v\n", task.TaskID, task.Args)
		switch mode {
		case "fail":
			fmt.Println(`{"success":false,"error":"fixture deliberate failure"}`)
		case "badexit":
			fmt.Fprintln(os.Stderr, "fixture boom")
			os.Exit(3)
		case "sleep":
			time.Sleep(30 * time.Second)
		case "nojson":
			fmt.Fprintln(os.Stderr, "only console, no result json")
		default:
			out := filepath.Join(task.WorkDir, "hello.txt")
			if err := os.WriteFile(out, []byte("built by fixture\n"+strings.Join(task.Args, " ")+"\n"), 0o644); err != nil {
				fmt.Fprintln(os.Stderr, "fixture: write failed:", err)
				os.Exit(4)
			}
			for k, v := range task.Env {
				_ = k
				_ = v
			}
			fmt.Println(`{"success":true,"outputs":["hello.txt"]}`)
		}
	}
}

// fixtureDir materializes a temp directory containing an executable
// named hg-exec-fixture that is a copy of the test binary.
func fixtureDir(t *testing.T, mode string) string {
	t.Helper()
	dir := t.TempDir()
	src, err := os.ReadFile(os.Args[0])
	if err != nil {
		t.Fatalf("failed to read test binary: %v", err)
	}
	name := "hg-exec-fixture"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	dst := filepath.Join(dir, name)
	if err := os.WriteFile(dst, src, 0o755); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	t.Setenv("HG_BUILDER_FIXTURE", "1")
	t.Setenv("HG_BUILDER_MODE", mode)
	return dir
}

func TestDiscover_Handshake(t *testing.T) {
	dir := fixtureDir(t, "ok")
	r := NewRegistry()
	ids := r.Discover(dir)
	if len(ids) != 1 || ids[0] != "fixture" {
		t.Fatalf("expected [fixture], got %v", ids)
	}
	b, ok := r.Get("fixture")
	if !ok {
		t.Fatal("builder not registered")
	}
	if b.Desc != "test fixture builder" {
		t.Errorf("description not captured: %q", b.Desc)
	}
}

func TestDiscover_SkipsBrokenPlugins(t *testing.T) {
	for _, mode := range []string{"wrongapi", "wrongid", "broken"} {
		dir := fixtureDir(t, mode)
		r := NewRegistry()
		if ids := r.Discover(dir); len(ids) != 0 {
			t.Errorf("mode %s: expected no builders, got %v", mode, ids)
		}
	}
}

func TestDiscover_SkipsNonExecutableAndWrongName(t *testing.T) {
	dir := t.TempDir()
	// Not executable.
	if err := os.WriteFile(filepath.Join(dir, "hg-exec-noexec"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// Wrong prefix.
	if err := os.WriteFile(filepath.Join(dir, "other-tool"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	// Invalid id charset.
	if err := os.WriteFile(filepath.Join(dir, "hg-exec-..evil"), []byte("x"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := NewRegistry()
	if ids := r.Discover(dir); len(ids) != 0 {
		t.Fatalf("expected no builders, got %v", ids)
	}
	// Missing directory is not an error.
	if ids := r.Discover(filepath.Join(dir, "does-not-exist")); ids != nil {
		t.Fatalf("expected nil for missing dir, got %v", ids)
	}
}

func TestRun_Success(t *testing.T) {
	dir := fixtureDir(t, "ok")
	r := NewRegistry()
	if ids := r.Discover(dir); len(ids) != 1 {
		t.Fatalf("discover failed: %v", ids)
	}
	b, _ := r.Get("fixture")

	work := t.TempDir()
	res, console, err := b.Run(context.Background(), Task{
		TaskID: "t1", BuilderID: "fixture", WorkDir: work, Args: []string{"--release"},
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !res.Success || len(res.Outputs) != 1 || res.Outputs[0] != "hello.txt" {
		t.Fatalf("unexpected result: %+v", res)
	}
	if !strings.Contains(console, "fixture console: task t1") {
		t.Errorf("console (stderr) not captured: %q", console)
	}
	data, err := os.ReadFile(filepath.Join(work, "hello.txt"))
	if err != nil || !strings.Contains(string(data), "--release") {
		t.Errorf("plugin did not see args: %v %q", err, data)
	}
}

func TestRun_PluginReportsFailure(t *testing.T) {
	dir := fixtureDir(t, "fail")
	r := NewRegistry()
	r.Discover(dir)
	b, _ := r.Get("fixture")
	res, _, err := b.Run(context.Background(), Task{TaskID: "t2", BuilderID: "fixture", WorkDir: t.TempDir()})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure result")
	}
	if res.Error != "fixture deliberate failure" {
		t.Errorf("error text: %q", res.Error)
	}
}

func TestRun_NonZeroExit(t *testing.T) {
	dir := fixtureDir(t, "badexit")
	r := NewRegistry()
	r.Discover(dir)
	b, _ := r.Get("fixture")
	res, console, err := b.Run(context.Background(), Task{TaskID: "t3", BuilderID: "fixture", WorkDir: t.TempDir()})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure result")
	}
	if !strings.Contains(console, "fixture boom") {
		t.Errorf("stderr not captured: %q", console)
	}
}

func TestRun_TimeoutKillsPlugin(t *testing.T) {
	dir := fixtureDir(t, "sleep")
	r := NewRegistry()
	r.Discover(dir)
	b, _ := r.Get("fixture")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	start := time.Now()
	res, _, err := b.Run(ctx, Task{TaskID: "t4", BuilderID: "fixture", WorkDir: t.TempDir()})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Success {
		t.Fatal("sleeping plugin must fail on timeout")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("plugin was not killed promptly: %v", elapsed)
	}
}

func TestRun_InvalidResultJSON(t *testing.T) {
	dir := fixtureDir(t, "nojson")
	r := NewRegistry()
	r.Discover(dir)
	b, _ := r.Get("fixture")
	res, console, err := b.Run(context.Background(), Task{TaskID: "t5", BuilderID: "fixture", WorkDir: t.TempDir()})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if res.Success {
		t.Fatal("expected failure result")
	}
	if !strings.Contains(res.Error, "invalid result JSON") {
		t.Errorf("error text: %q", res.Error)
	}
	if !strings.Contains(console, "only console") {
		t.Errorf("console: %q", console)
	}
}
