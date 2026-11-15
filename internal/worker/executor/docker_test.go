package executor

import (
	"context"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	dockerclient "github.com/docker/docker/client"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// TestDockerExecutor_Name tests the Name method
func TestDockerExecutor_Name(t *testing.T) {
	e := &DockerExecutor{}
	if got := e.Name(); got != "docker" {
		t.Errorf("Name() = %q, want %q", got, "docker")
	}
}

// TestDockerExecutor_CanExecute tests the CanExecute method
func TestDockerExecutor_CanExecute(t *testing.T) {
	e := &DockerExecutor{
		images: defaultImages,
	}

	tests := []struct {
		name       string
		targetArch pb.Architecture
		nativeArch pb.Architecture
		want       bool
	}{
		{
			name:       "x86_64 supported",
			targetArch: pb.Architecture_ARCH_X86_64,
			nativeArch: pb.Architecture_ARCH_ARM64,
			want:       true,
		},
		{
			name:       "arm64 supported",
			targetArch: pb.Architecture_ARCH_ARM64,
			nativeArch: pb.Architecture_ARCH_X86_64,
			want:       true,
		},
		{
			name:       "armv7 supported",
			targetArch: pb.Architecture_ARCH_ARMV7,
			nativeArch: pb.Architecture_ARCH_X86_64,
			want:       true,
		},
		{
			name:       "unsupported architecture",
			targetArch: pb.Architecture_ARCH_UNSPECIFIED,
			nativeArch: pb.Architecture_ARCH_X86_64,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.CanExecute(tt.targetArch, tt.nativeArch)
			if got != tt.want {
				t.Errorf("CanExecute(%v, %v) = %v, want %v", tt.targetArch, tt.nativeArch, got, tt.want)
			}
		})
	}
}

// TestDockerExecutor_SetImage tests the SetImage method
func TestDockerExecutor_SetImage(t *testing.T) {
	e := &DockerExecutor{
		images: make(map[pb.Architecture]string),
	}

	e.SetImage(pb.Architecture_ARCH_X86_64, "custom/image:latest")

	got := e.images[pb.Architecture_ARCH_X86_64]
	want := "custom/image:latest"
	if got != want {
		t.Errorf("SetImage() failed: got %q, want %q", got, want)
	}
}

// TestDockerExecutor_selectImage tests the selectImage method
func TestDockerExecutor_selectImage(t *testing.T) {
	e := &DockerExecutor{
		images: defaultImages,
	}

	tests := []struct {
		name string
		arch pb.Architecture
		want string
	}{
		{
			name: "x86_64",
			arch: pb.Architecture_ARCH_X86_64,
			want: "dockcross/linux-x64",
		},
		{
			name: "arm64",
			arch: pb.Architecture_ARCH_ARM64,
			want: "dockcross/linux-arm64",
		},
		{
			name: "unknown",
			arch: pb.Architecture_ARCH_UNSPECIFIED,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.selectImage(tt.arch)
			if got != tt.want {
				t.Errorf("selectImage(%v) = %q, want %q", tt.arch, got, tt.want)
			}
		})
	}
}

// TestDockerExecutor_buildCommand tests the buildCommand method
func TestDockerExecutor_buildCommand(t *testing.T) {
	e := &DockerExecutor{}

	tests := []struct {
		name     string
		compiler string
		args     []string
		srcFile  string
		outFile  string
		wantHas  []string
		wantNot  []string
	}{
		{
			name:     "basic compile",
			compiler: "gcc",
			args:     []string{"-O2", "-Wall"},
			srcFile:  "source.i",
			outFile:  "output.o",
			wantHas:  []string{"gcc", "-c", "-O2", "-Wall", "source.i", "-o", "output.o"},
			wantNot:  []string{},
		},
		{
			name:     "filter -c and -o",
			compiler: "gcc",
			args:     []string{"-c", "-o", "old.o", "-O2"},
			srcFile:  "source.i",
			outFile:  "output.o",
			wantHas:  []string{"gcc", "-c", "-O2", "source.i", "-o", "output.o"},
			wantNot:  []string{"old.o"},
		},
		{
			name:     "filter input files",
			compiler: "gcc",
			args:     []string{"-O2", "original.c", "-Wall"},
			srcFile:  "source.i",
			outFile:  "output.o",
			wantHas:  []string{"gcc", "-c", "-O2", "-Wall", "source.i", "-o", "output.o"},
			wantNot:  []string{"original.c"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.buildCommand(tt.compiler, tt.args, tt.srcFile, tt.outFile)

			for _, want := range tt.wantHas {
				found := false
				for _, arg := range got {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildCommand() missing expected arg %q in %v", want, got)
				}
			}

			for _, notWant := range tt.wantNot {
				for _, arg := range got {
					if arg == notWant {
						t.Errorf("buildCommand() has unexpected arg %q in %v", notWant, got)
					}
				}
			}
		})
	}
}

// TestDockerExecutor_buildRawCommand tests the buildRawCommand method
func TestDockerExecutor_buildRawCommand(t *testing.T) {
	e := &DockerExecutor{}

	tests := []struct {
		name         string
		compiler     string
		args         []string
		srcFile      string
		outFile      string
		includePaths []string
		wantHas      []string
		wantNot      []string
	}{
		{
			name:         "with include paths",
			compiler:     "gcc",
			args:         []string{"-O2", "-Wall"},
			srcFile:      "main.cpp",
			outFile:      "output.o",
			includePaths: []string{"include", "src/headers"},
			wantHas:      []string{"gcc", "-c", "-I/work/include", "-I/work/src/headers", "-O2", "-Wall", "main.cpp", "-o", "output.o"},
			wantNot:      []string{},
		},
		{
			name:         "filter -I flags",
			compiler:     "g++",
			args:         []string{"-I", "/old/path", "-I/another/old", "-O2"},
			srcFile:      "source.cpp",
			outFile:      "output.o",
			includePaths: []string{"new/path"},
			wantHas:      []string{"g++", "-c", "-I/work/new/path", "-O2", "source.cpp", "-o", "output.o"},
			wantNot:      []string{"/old/path", "/another/old"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.buildRawCommand(tt.compiler, tt.args, tt.srcFile, tt.outFile, tt.includePaths)

			for _, want := range tt.wantHas {
				found := false
				for _, arg := range got {
					if arg == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("buildRawCommand() missing expected arg %q in %v", want, got)
				}
			}

			for _, notWant := range tt.wantNot {
				for _, arg := range got {
					if arg == notWant {
						t.Errorf("buildRawCommand() has unexpected arg %q in %v", notWant, got)
					}
				}
			}
		})
	}
}

// TestDefaultDockerResourceLimits tests the default resource limits
func TestDefaultDockerResourceLimits(t *testing.T) {
	limits := DefaultDockerResourceLimits()

	if limits.MemoryBytes != 512*1024*1024 {
		t.Errorf("MemoryBytes = %d, want %d", limits.MemoryBytes, 512*1024*1024)
	}

	if limits.NanoCPUs != 1_000_000_000 {
		t.Errorf("NanoCPUs = %d, want %d", limits.NanoCPUs, 1_000_000_000)
	}

	if limits.PidsLimit != 100 {
		t.Errorf("PidsLimit = %d, want %d", limits.PidsLimit, 100)
	}
}

func TestIsWSL2DockerAvailable_NonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-Windows assertion only")
	}

	if IsWSL2DockerAvailable() {
		t.Fatal("IsWSL2DockerAvailable() = true, want false on non-Windows")
	}
}

func TestDockerExecutor_EnsureImage_ImageAlreadyPresent(t *testing.T) {
	client := newTestDockerAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/images/json"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"RepoTags":["dockcross/linux-x64:latest"]}]`))
		default:
			http.NotFound(w, r)
		}
	})
	defer client.Close()

	e := &DockerExecutor{client: client}
	if err := e.ensureImage(context.Background(), "dockcross/linux-x64", pb.Architecture_ARCH_X86_64); err != nil {
		t.Fatalf("ensureImage() error = %v", err)
	}
}

func TestDockerExecutor_EnsureImage_PullsWhenMissing(t *testing.T) {
	pullCalled := false
	client := newTestDockerAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/images/json"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[]`))
		case strings.Contains(r.URL.Path, "/images/create"):
			pullCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte("{}\n"))
		default:
			http.NotFound(w, r)
		}
	})
	defer client.Close()

	e := &DockerExecutor{client: client}
	if err := e.ensureImage(context.Background(), "dockcross/linux-x64", pb.Architecture_ARCH_X86_64); err != nil {
		t.Fatalf("ensureImage() error = %v", err)
	}
	if !pullCalled {
		t.Fatal("ensureImage() did not pull missing image")
	}
}

func TestDockerExecutor_GetLogs(t *testing.T) {
	body := appendDockerLogFrame(1, "hello stdout\n")
	body = append(body, appendDockerLogFrame(2, "hello stderr\n")...)

	client := newTestDockerAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/containers/test-container/logs") {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
			return
		}
		http.NotFound(w, r)
	})
	defer client.Close()

	e := &DockerExecutor{client: client}
	stdout, stderr, err := e.getLogs(context.Background(), "test-container")
	if err != nil {
		t.Fatalf("getLogs() error = %v", err)
	}
	if stdout != "hello stdout\n" {
		t.Fatalf("stdout = %q, want %q", stdout, "hello stdout\n")
	}
	if stderr != "hello stderr\n" {
		t.Fatalf("stderr = %q, want %q", stderr, "hello stderr\n")
	}
}

func TestDockerExecutor_Close(t *testing.T) {
	client := newTestDockerAPIClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	e := &DockerExecutor{client: client}
	if err := e.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}

func newTestDockerAPIClient(t *testing.T, handler http.HandlerFunc) *dockerclient.Client {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client, err := dockerclient.NewClientWithOpts(
		dockerclient.WithHost(server.URL),
		dockerclient.WithVersion("1.48"),
		dockerclient.WithHTTPClient(server.Client()),
	)
	if err != nil {
		t.Fatalf("NewClientWithOpts() error = %v", err)
	}

	return client
}

func appendDockerLogFrame(stream byte, payload string) []byte {
	frame := make([]byte, 8+len(payload))
	frame[0] = stream
	binary.BigEndian.PutUint32(frame[4:8], uint32(len(payload)))
	copy(frame[8:], payload)
	return frame
}

func TestArchPlatform(t *testing.T) {
	cases := map[pb.Architecture]string{
		pb.Architecture_ARCH_X86_64:      "linux/amd64",
		pb.Architecture_ARCH_ARM64:       "linux/arm64",
		pb.Architecture_ARCH_ARMV7:       "linux/arm/v7",
		pb.Architecture_ARCH_UNSPECIFIED: "",
	}
	for arch, want := range cases {
		if got := archPlatform(arch); got != want {
			t.Fatalf("archPlatform(%v) = %q, want %q", arch, got, want)
		}
	}
}

func TestTripleMatches(t *testing.T) {
	if !tripleMatches("aarch64-linux-gnu", pb.Architecture_ARCH_ARM64) {
		t.Fatal("aarch64 triple must match ARM64")
	}
	if tripleMatches("x86_64-linux-gnu", pb.Architecture_ARCH_ARM64) {
		t.Fatal("x86_64 triple must NOT match ARM64 (the silent-wrong-arch bug)")
	}
	if !tripleMatches("x86_64-linux-gnu", pb.Architecture_ARCH_X86_64) {
		t.Fatal("x86_64 triple must match X86_64")
	}
	if !tripleMatches("arm-linux-gnueabihf", pb.Architecture_ARCH_ARMV7) {
		t.Fatal("arm triple must match ARMV7")
	}
	if !tripleMatches("", pb.Architecture_ARCH_UNSPECIFIED) {
		t.Fatal("unspecified target accepts anything")
	}
}

func TestToolchainFor(t *testing.T) {
	e := &DockerExecutor{toolchains: map[pb.Architecture]string{
		pb.Architecture_ARCH_ARM64: "aarch64-linux-gnu-gcc",
	}}
	if got := e.toolchainFor(pb.Architecture_ARCH_ARM64, "gcc"); got != "aarch64-linux-gnu-gcc" {
		t.Fatalf("toolchainFor override = %q", got)
	}
	if got := e.toolchainFor(pb.Architecture_ARCH_X86_64, "gcc"); got != "gcc" {
		t.Fatalf("toolchainFor fallback = %q, want client compiler", got)
	}
	if got := e.toolchainFor(pb.Architecture_ARCH_ARM64, "g++"); got != "aarch64-linux-gnu-gcc" {
		t.Fatalf("toolchainFor empty override must fall back, got %q", got)
	}
}
