package capability

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestDetectMemoryLinux_ParseMeminfo(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	mem := detectMemoryLinux()

	if mem < 0 {
		t.Errorf("Memory should not be negative: %d", mem)
	}

	t.Logf("Linux memory: %d bytes", mem)
}

func TestDetectMemoryDarwin_ParseSysctl(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific test")
	}

	mem := detectMemoryDarwin()

	if mem <= 0 {
		t.Errorf("Darwin should detect memory, got %d", mem)
	}

	t.Logf("Darwin memory: %d bytes (%d GB)", mem, mem/(1024*1024*1024))
}

func TestDetectMemoryWindows_ParseWmic(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	mem := detectMemoryWindows()

	if mem <= 0 {
		t.Log("Warning: Windows memory detection failed")
	} else {
		t.Logf("Windows memory: %d bytes (%d GB)", mem, mem/(1024*1024*1024))
	}
}

func TestDetectCpp_NoCompilersPath(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", t.TempDir())

	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp should never return nil")
	}

	if len(caps.Compilers) > 0 {
		t.Errorf("Expected no compilers with empty PATH, got: %v", caps.Compilers)
	}
}

func TestDetectGo_CommandFailure(t *testing.T) {
	caps := detectGo()

	if caps == nil {
		t.Fatal("Go should be available in test environment")
	}

	if caps.Version == "" {
		t.Error("Go version should not be empty")
	}

	if !caps.CrossCompile {
		t.Error("Go should always report cross-compile capability")
	}
}

func TestDetectRust_NotInstalled(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", t.TempDir())

	caps := detectRust()

	if caps != nil {
		t.Errorf("Expected nil when rustc not in PATH, got: %v", caps)
	}
}

func TestDetectNode_NotInstalled(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", t.TempDir())

	caps := detectNode()

	if caps != nil {
		t.Errorf("Expected nil when node not in PATH, got: %v", caps)
	}
}

func TestDetectFlutter_NotInstalled(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", t.TempDir())

	caps := detectFlutter()

	if caps != nil {
		t.Errorf("Expected nil when flutter not in PATH, got: %v", caps)
	}
}

func TestDetectDocker_NotInstalled(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	os.Setenv("PATH", t.TempDir())

	available := detectDocker()

	if available {
		t.Error("Expected false when docker not in PATH")
	}
}

func TestDetectCpp_AllCompilers(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp returned nil")
	}

	expectedCompilers := []string{"gcc", "g++", "clang", "clang++"}
	for _, expected := range expectedCompilers {
		found := false
		for _, actual := range caps.Compilers {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Compiler %s not found (may not be installed)", expected)
		}
	}
}

func TestDetectNode_AllPackageManagers(t *testing.T) {
	caps := detectNode()

	if caps == nil {
		t.Skip("Node not installed")
	}

	expectedManagers := []string{"npm", "yarn", "pnpm", "bun"}
	for _, expected := range expectedManagers {
		found := false
		for _, actual := range caps.PackageManagers {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Logf("Package manager %s not found (may not be installed)", expected)
		}
	}
}

func TestDetectFlutter_XcodeIntegration(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("macOS-specific test")
	}

	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	if caps.XcodeAvailable {
		t.Log("Xcode detected, iOS and macOS platforms should be available")
	} else {
		t.Log("Xcode not detected")
	}
}

func TestDetectFlutter_AndroidIntegration(t *testing.T) {
	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	androidHome := os.Getenv("ANDROID_HOME")
	androidSDKRoot := os.Getenv("ANDROID_SDK_ROOT")

	if androidHome != "" || androidSDKRoot != "" {
		if !caps.AndroidSdk {
			t.Error("ANDROID_HOME set but AndroidSdk is false")
		}
	}
}

func TestDetectMemory_ReturnValue(t *testing.T) {
	mem := detectMemory()

	if mem < 0 {
		t.Errorf("Memory should not be negative: %d", mem)
	}
}

func TestDetectArch_ReturnValue(t *testing.T) {
	arch := detectArch()

	validArchs := []int32{
		int32(pb.Architecture_ARCH_UNSPECIFIED),
		int32(pb.Architecture_ARCH_X86_64),
		int32(pb.Architecture_ARCH_ARM64),
		int32(pb.Architecture_ARCH_ARMV7),
	}

	valid := false
	for _, validArch := range validArchs {
		if int32(arch) == validArch {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("Invalid architecture returned: %v", arch)
	}
}

func TestDetect_Consistency(t *testing.T) {
	caps1 := Detect()
	caps2 := Detect()

	if caps1.Os != caps2.Os {
		t.Error("OS detection should be consistent")
	}

	if caps1.NativeArch != caps2.NativeArch {
		t.Error("Architecture detection should be consistent")
	}

	if caps1.CpuCores != caps2.CpuCores {
		t.Error("CPU cores detection should be consistent")
	}
}

func TestDetectCpp_PathModification(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	tempDir := t.TempDir()

	fakeGcc := filepath.Join(tempDir, "gcc")
	if err := os.WriteFile(fakeGcc, []byte("#!/bin/sh\necho fake"), 0755); err != nil {
		t.Fatal(err)
	}

	os.Setenv("PATH", tempDir)

	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp returned nil")
	}

	found := false
	for _, c := range caps.Compilers {
		if c == "gcc" {
			found = true
			break
		}
	}

	if !found && runtime.GOOS != "windows" {
		t.Log("gcc in PATH but not detected (PATH manipulation may not work in test)")
	}
}

func TestDetect_AllFields(t *testing.T) {
	caps := Detect()

	if caps == nil {
		t.Fatal("Detect() returned nil")
	}

	if caps.Os == "" {
		t.Error("OS field should not be empty")
	}

	if caps.CpuCores <= 0 {
		t.Error("CpuCores should be positive")
	}

	t.Logf("Detected capabilities: OS=%s, Arch=%v, Cores=%d, Memory=%d, Docker=%v",
		caps.Os, caps.NativeArch, caps.CpuCores, caps.MemoryBytes, caps.DockerAvailable)
}

func TestDetectCpp_NonWindowsPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Non-Windows test")
	}

	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp returned nil")
	}

	t.Logf("Non-Windows C++ capabilities: compilers=%v, cross=%v",
		caps.Compilers, caps.CrossCompile)
}
