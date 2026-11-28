package flutter

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// tarOf builds a plain tar of name→content entries.
func tarOf(t *testing.T, entries map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range entries {
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

func TestNormalizeABIs(t *testing.T) {
	tests := []struct {
		in      []string
		want    string
		wantErr bool
	}{
		{nil, "android-arm64,android-arm,android-x64", false},
		{[]string{"arm64", "android-x64", "ARM64"}, "android-arm64,android-x64", false},
		{[]string{"mips"}, "", true},
		{[]string{" "}, "", true},
	}
	for _, tt := range tests {
		got, err := normalizeABIs(tt.in)
		if (err != nil) != tt.wantErr || strings.Join(got, ",") != tt.want {
			t.Errorf("normalizeABIs(%q) = %q, %v; want %q, err=%v", tt.in, got, err, tt.want, tt.wantErr)
		}
	}
}

func TestAbiRequestsShareBuildButNotTask(t *testing.T) {
	base := &pb.BuildRequest{
		TaskId: "base", BuildId: "b1", SourceHash: "h",
		Config: &pb.BuildRequest_FlutterConfig{FlutterConfig: &pb.FlutterConfig{BuildMode: "apk-release"}},
	}
	reqs := abiRequests(base, []string{"android-arm64", "android-x64"})
	if len(reqs) != 2 || reqs[0].TaskId == reqs[1].TaskId || reqs[0].TaskId == "base" {
		t.Fatalf("task ids not unique: %v", []string{reqs[0].TaskId, reqs[1].TaskId})
	}
	for i, abi := range []string{"android-arm64", "android-x64"} {
		cfg := reqs[i].GetFlutterConfig()
		if reqs[i].BuildId != "b1" || cfg.GetBuildMode() != "apk-release" || !cfg.GetSplitPerAbi() ||
			len(cfg.GetTargetPlatforms()) != 1 || cfg.GetTargetPlatforms()[0] != abi {
			t.Fatalf("request %d = %+v", i, reqs[i])
		}
	}
	if len(base.GetFlutterConfig().GetTargetPlatforms()) != 0 {
		t.Fatal("base request was mutated")
	}
}

// concurrencyClient records how many builds were in flight at once.
type concurrencyClient struct {
	mu       sync.Mutex
	inFlight int
	peak     int
}

func (c *concurrencyClient) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	c.mu.Lock()
	c.inFlight++
	if c.inFlight > c.peak {
		c.peak = c.inFlight
	}
	c.mu.Unlock()
	time.Sleep(30 * time.Millisecond)
	c.mu.Lock()
	c.inFlight--
	c.mu.Unlock()
	abi := req.GetFlutterConfig().GetTargetPlatforms()[0]
	return &pb.BuildResponse{
		Status:    pb.TaskStatus_STATUS_COMPLETED,
		WorkerId:  "w-" + abi,
		Artifacts: nil,
	}, nil
}

func (c *concurrencyClient) Close() error { return nil }

func TestRunDistributedRunsInParallelAndKeepsOrder(t *testing.T) {
	base := &pb.BuildRequest{Config: &pb.BuildRequest_FlutterConfig{FlutterConfig: &pb.FlutterConfig{}}}
	client := &concurrencyClient{}
	results := runDistributed(context.Background(), client, abiRequests(base, []string{"android-x64", "android-arm64", "android-arm"}))
	if client.peak != 3 {
		t.Fatalf("peak concurrency = %d, want 3", client.peak)
	}
	var abis []string
	for _, r := range results {
		abis = append(abis, r.ABI)
		if r.Err != nil || r.Resp.GetWorkerId() != "w-"+r.ABI {
			t.Fatalf("result %+v", r)
		}
	}
	if strings.Join(abis, ",") != "android-arm64,android-arm,android-x64" {
		t.Fatalf("results order = %v", abis)
	}
}

func TestExtractOutputs(t *testing.T) {
	archive := tarOf(t, map[string]string{
		"build/app/outputs/flutter-apk/app-arm64-v8a-release.apk": "apk-bytes",
		"build/app/outputs/flutter-apk/app.apk.sha1":              "ignored",
		"../../escape.apk": "contained",
	})
	dir := t.TempDir()
	paths, err := extractOutputs(archive, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %v", paths)
	}
	for _, p := range paths {
		if filepath.Dir(p) != dir {
			t.Fatalf("%s written outside %s", p, dir)
		}
	}
	data, _ := os.ReadFile(filepath.Join(dir, "app-arm64-v8a-release.apk"))
	if string(data) != "apk-bytes" {
		t.Fatalf("apk content = %q", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "app.apk.sha1")); !os.IsNotExist(err) {
		t.Fatal("non-output file was extracted")
	}

	// gzip-wrapped archives are accepted too; zstd is reported, not mangled.
	var gz bytes.Buffer
	zw := gzip.NewWriter(&gz)
	_, _ = zw.Write(tarOf(t, map[string]string{"app.aab": "aab"}))
	_ = zw.Close()
	if paths, err := extractOutputs(gz.Bytes(), t.TempDir()); err != nil || len(paths) != 1 {
		t.Fatalf("gzip extract = %v, %v", paths, err)
	}
	if _, err := extractOutputs([]byte{0x28, 0xb5, 0x2f, 0xfd, 0}, t.TempDir()); err == nil {
		t.Fatal("expected an error for zstd archives")
	}
}

// outputClient returns one APK per requested ABI.
type outputClient struct{ t *testing.T }

func (c *outputClient) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	abi := strings.TrimPrefix(req.GetFlutterConfig().GetTargetPlatforms()[0], "android-")
	name := "app-" + abi + "-release.apk"
	return &pb.BuildResponse{
		Status:       pb.TaskStatus_STATUS_COMPLETED,
		WorkerId:     "w-" + abi,
		Artifacts:    tarOf(c.t, map[string]string{"build/" + name: abi}),
		ArtifactList: []*pb.ArtifactInfo{{Name: name, Path: "build/" + name, SizeBytes: int64(len(abi))}},
	}, nil
}

func (c *outputClient) Close() error { return nil }

func newTestCommand(t *testing.T, client BuildClient) (*bytes.Buffer, func(args ...string) error) {
	t.Helper()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "pubspec.yaml"), []byte("name: demo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	return &out, func(args ...string) error {
		cmd := NewCommand(Dependencies{
			CoordinatorAddr: func() string { return "coordinator:9000" },
			NewClient:       func(string, time.Duration) (BuildClient, error) { return client, nil },
			RequestTimeout:  time.Second,
		})
		cmd.SetOut(&out)
		cmd.SetErr(io.Discard)
		cmd.SetArgs(append([]string{"build"}, append(args, "--project", project)...))
		return cmd.Execute()
	}
}

func TestFlutterCLI_DistributeABIsSavesEveryAPK(t *testing.T) {
	out, run := newTestCommand(t, &outputClient{t: t})
	dir := t.TempDir()
	if err := run("apk", "--distribute-abis", "--abis", "arm64,x64", "--output", dir); err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	for _, name := range []string{"app-arm64-release.apk", "app-x64-release.apk"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Fatalf("%s not saved: %v\n%s", name, err, out)
		}
	}
	if !strings.Contains(out.String(), "Building 2 ABIs in parallel") || !strings.Contains(out.String(), "w-arm64") {
		t.Fatalf("output missing per-ABI report:\n%s", out)
	}
}

func TestFlutterCLI_DistributeRejectsAppBundle(t *testing.T) {
	_, run := newTestCommand(t, &outputClient{t: t})
	if err := run("appbundle", "--distribute-abis"); err == nil {
		t.Fatal("appbundle must not accept --distribute-abis")
	}
}
