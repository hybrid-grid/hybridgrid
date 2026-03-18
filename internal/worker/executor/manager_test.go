package executor

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"time"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

type fakeExecutor struct {
	name       string
	canExecute bool
	execute    func(context.Context, *Request) (*Result, error)
	closeErr   error
}

func (f *fakeExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	if f.execute != nil {
		return f.execute(ctx, req)
	}
	return &Result{Success: true}, nil
}

func (f *fakeExecutor) Name() string {
	return f.name
}

func (f *fakeExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	return f.canExecute
}

func (f *fakeExecutor) Close() error {
	return f.closeErr
}

// TestManager_NewManager tests Manager initialization
func TestManager_NewManager(t *testing.T) {
	tests := []struct {
		name            string
		nativeArch      pb.Architecture
		dockerAvailable bool
		wantNative      bool
		wantDocker      bool
	}{
		{
			name:            "without docker",
			nativeArch:      pb.Architecture_ARCH_X86_64,
			dockerAvailable: false,
			wantNative:      true,
			wantDocker:      false,
		},
		{
			name:            "with docker",
			nativeArch:      pb.Architecture_ARCH_ARM64,
			dockerAvailable: true,
			wantNative:      true,
			wantDocker:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.nativeArch, tt.dockerAvailable)

			if m == nil {
				t.Fatal("NewManager returned nil")
			}

			if m.nativeArch != tt.nativeArch {
				t.Errorf("nativeArch = %v, want %v", m.nativeArch, tt.nativeArch)
			}

			if (m.native != nil) != tt.wantNative {
				t.Errorf("native executor present = %v, want %v", m.native != nil, tt.wantNative)
			}

			if tt.dockerAvailable && m.docker == nil {
				t.Log("Docker executor not initialized (likely Docker daemon unavailable)")
			}
		})
	}
}

// TestManager_SelectArchitectures tests executor selection by architecture
func TestManager_SelectArchitectures(t *testing.T) {
	tests := []struct {
		name            string
		nativeArch      pb.Architecture
		targetArch      pb.Architecture
		dockerAvailable bool
		wantExecutor    string
	}{
		{
			name:            "native match",
			nativeArch:      pb.Architecture_ARCH_X86_64,
			targetArch:      pb.Architecture_ARCH_X86_64,
			dockerAvailable: false,
			wantExecutor:    "native",
		},
		{
			name:            "unspecified uses native",
			nativeArch:      pb.Architecture_ARCH_ARM64,
			targetArch:      pb.Architecture_ARCH_UNSPECIFIED,
			dockerAvailable: false,
			wantExecutor:    "native",
		},
		{
			name:            "cross-compile without docker falls back to native",
			nativeArch:      pb.Architecture_ARCH_X86_64,
			targetArch:      pb.Architecture_ARCH_ARM64,
			dockerAvailable: false,
			wantExecutor:    "native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.nativeArch, tt.dockerAvailable)
			executor := m.Select(tt.targetArch)

			if executor == nil {
				t.Fatal("Select returned nil executor")
			}

			if executor.Name() != tt.wantExecutor {
				t.Errorf("Select() executor = %s, want %s", executor.Name(), tt.wantExecutor)
			}
		})
	}
}

func TestManager_Select_UsesDockerWhenAvailable(t *testing.T) {
	m := &Manager{
		native:     &fakeExecutor{name: "native"},
		docker:     &fakeExecutor{name: "docker", canExecute: true},
		nativeArch: pb.Architecture_ARCH_X86_64,
	}

	got := m.Select(pb.Architecture_ARCH_ARM64)
	if got.Name() != "docker" {
		t.Fatalf("Select() executor = %s, want docker", got.Name())
	}
}

// TestManager_SelectForRequest tests executor selection considering client OS
func TestManager_SelectForRequest(t *testing.T) {
	currentOS := runtime.GOOS

	tests := []struct {
		name         string
		req          *Request
		nativeArch   pb.Architecture
		wantExecutor string
	}{
		{
			name: "preprocessed source uses arch-based selection",
			req: &Request{
				PreprocessedSource: []byte("int main() {}"),
				TargetArch:         pb.Architecture_ARCH_X86_64,
			},
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
		{
			name: "raw source same OS uses arch-based selection",
			req: &Request{
				RawSource:  []byte("int main() {}"),
				ClientOs:   currentOS,
				TargetArch: pb.Architecture_ARCH_X86_64,
			},
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
		{
			name: "raw source different OS without docker falls back",
			req: &Request{
				RawSource:  []byte("int main() {}"),
				ClientOs:   "different-os",
				TargetArch: pb.Architecture_ARCH_X86_64,
			},
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.nativeArch, false)
			executor := m.SelectForRequest(tt.req)

			if executor == nil {
				t.Fatal("SelectForRequest returned nil executor")
			}

			if executor.Name() != tt.wantExecutor {
				t.Errorf("SelectForRequest() executor = %s, want %s", executor.Name(), tt.wantExecutor)
			}
		})
	}
}

func TestManager_SelectForRequest_UsesDockerForCrossOSRawSource(t *testing.T) {
	m := &Manager{
		native:     &fakeExecutor{name: "native"},
		docker:     &fakeExecutor{name: "docker", canExecute: true},
		nativeArch: pb.Architecture_ARCH_X86_64,
	}

	req := &Request{
		RawSource:  []byte("int main() {}"),
		ClientOs:   "different-os",
		TargetArch: pb.Architecture_ARCH_ARM64,
	}

	got := m.SelectForRequest(req)
	if got.Name() != "docker" {
		t.Fatalf("SelectForRequest() executor = %s, want docker", got.Name())
	}
}

// TestManager_SelectForCompiler tests compiler-based executor selection
func TestManager_SelectForCompiler(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("MSVC executor is always present on Windows; cl.exe cases require a non-MSVC environment")
	}

	tests := []struct {
		name         string
		compiler     string
		targetArch   pb.Architecture
		nativeArch   pb.Architecture
		wantExecutor string
	}{
		{
			name:         "gcc uses standard selection",
			compiler:     "gcc",
			targetArch:   pb.Architecture_ARCH_X86_64,
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
		{
			name:         "g++ uses standard selection",
			compiler:     "g++",
			targetArch:   pb.Architecture_ARCH_ARM64,
			nativeArch:   pb.Architecture_ARCH_ARM64,
			wantExecutor: "native",
		},
		{
			name:         "cl.exe without msvc executor uses standard selection",
			compiler:     "cl.exe",
			targetArch:   pb.Architecture_ARCH_X86_64,
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
		{
			name:         "cl without msvc executor uses standard selection",
			compiler:     "cl",
			targetArch:   pb.Architecture_ARCH_X86_64,
			nativeArch:   pb.Architecture_ARCH_X86_64,
			wantExecutor: "native",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.nativeArch, false)
			executor := m.SelectForCompiler(tt.compiler, tt.targetArch)

			if executor == nil {
				t.Fatal("SelectForCompiler returned nil executor")
			}

			if executor.Name() != tt.wantExecutor {
				t.Errorf("SelectForCompiler() executor = %s, want %s", executor.Name(), tt.wantExecutor)
			}
		})
	}
}

func TestManager_SelectForCompiler_UsesMSVCExecutorWhenAvailable(t *testing.T) {
	m := &Manager{
		native:     &fakeExecutor{name: "native"},
		msvc:       &fakeExecutor{name: "msvc", canExecute: true},
		nativeArch: pb.Architecture_ARCH_X86_64,
	}

	got := m.SelectForCompiler("cl.exe", pb.Architecture_ARCH_X86_64)
	if got.Name() != "msvc" {
		t.Fatalf("SelectForCompiler() executor = %s, want msvc", got.Name())
	}
}

// TestManager_isMSVCCompiler tests MSVC compiler detection
func TestManager_isMSVCCompiler(t *testing.T) {
	tests := []struct {
		compiler string
		want     bool
	}{
		{"cl.exe", true},
		{"cl", true},
		{"CL.EXE", true},
		{"CL", true},
		{"gcc", false},
		{"g++", false},
		{"clang", false},
		{"clang++", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.compiler, func(t *testing.T) {
			got := isMSVCCompiler(tt.compiler)
			if got != tt.want {
				t.Errorf("isMSVCCompiler(%q) = %v, want %v", tt.compiler, got, tt.want)
			}
		})
	}
}

// TestManager_Close tests resource cleanup
func TestManager_Close(t *testing.T) {
	tests := []struct {
		name            string
		dockerAvailable bool
	}{
		{
			name:            "without docker",
			dockerAvailable: false,
		},
		{
			name:            "with docker (may fail if daemon unavailable)",
			dockerAvailable: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(pb.Architecture_ARCH_X86_64, tt.dockerAvailable)

			err := m.Close()
			if tt.dockerAvailable && m.docker != nil && err != nil {
				t.Logf("Close() error = %v (Docker daemon may be unavailable)", err)
			}
		})
	}
}

func TestManager_Close_PropagatesDockerCloserError(t *testing.T) {
	wantErr := errors.New("close failed")
	m := &Manager{
		docker: &fakeExecutor{name: "docker", closeErr: wantErr},
	}

	err := m.Close()
	if !errors.Is(err, wantErr) {
		t.Fatalf("Close() error = %v, want %v", err, wantErr)
	}
}

// TestManager_GetMSVC tests MSVC executor retrieval
func TestManager_GetMSVC(t *testing.T) {
	m := NewManager(pb.Architecture_ARCH_X86_64, false)
	msvc := m.GetMSVC()

	if runtime.GOOS == "windows" {
		if msvc == nil {
			t.Log("MSVC executor not available on Windows (expected if MSVC not installed)")
		}
	} else {
		if msvc != nil {
			t.Error("GetMSVC() returned non-nil on non-Windows platform")
		}
	}
}

// TestManager_Execute tests the Manager's Execute method with tracing
func TestManager_Execute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	m := NewManager(pb.Architecture_ARCH_X86_64, false)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := &Request{
		TaskID:             "manager-exec-001",
		Compiler:           "gcc",
		Args:               []string{"-c", "-O2"},
		PreprocessedSource: []byte("int main() { return 0; }"),
		TargetArch:         pb.Architecture_ARCH_X86_64,
		SourceFilename:     "test.c",
		Timeout:            10 * time.Second,
	}

	result, err := m.Execute(ctx, req)

	if err != nil {
		t.Skipf("Execute failed (likely gcc not available): %v", err)
	}

	if !result.Success {
		t.Errorf("Execute() failed: stderr=%s", result.Stderr)
	}

	if result.ExitCode != 0 {
		t.Errorf("Execute() exit code = %d, want 0", result.ExitCode)
	}

	if len(result.ObjectCode) == 0 {
		t.Error("Execute() returned empty object code")
	}
}

func TestManager_Execute_ErrorFromExecutor(t *testing.T) {
	wantErr := errors.New("executor failed")
	m := &Manager{
		native:     &fakeExecutor{name: "native", execute: func(context.Context, *Request) (*Result, error) { return nil, wantErr }},
		nativeArch: pb.Architecture_ARCH_X86_64,
	}

	_, err := m.Execute(context.Background(), &Request{TaskID: "task-err", Compiler: "gcc", SourceFilename: "main.c"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("Execute() error = %v, want %v", err, wantErr)
	}
}

func TestManager_Execute_FailedResultWithoutError(t *testing.T) {
	m := &Manager{
		native: &fakeExecutor{name: "native", execute: func(context.Context, *Request) (*Result, error) {
			return &Result{Success: false, ExitCode: 2, Stderr: "compile failed"}, nil
		}},
		nativeArch: pb.Architecture_ARCH_X86_64,
	}

	result, err := m.Execute(context.Background(), &Request{TaskID: "task-fail", Compiler: "gcc", SourceFilename: "main.c"})
	if err != nil {
		t.Fatalf("Execute() unexpected error = %v", err)
	}
	if result == nil || result.Success {
		t.Fatalf("Execute() result = %#v, want failed result", result)
	}
}
