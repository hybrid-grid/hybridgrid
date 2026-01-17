package capability

import (
	"runtime"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestDetect(t *testing.T) {
	caps := Detect()

	if caps == nil {
		t.Fatal("Detect() returned nil")
	}

	// Basic sanity checks
	if caps.CpuCores <= 0 {
		t.Errorf("CpuCores should be > 0, got %d", caps.CpuCores)
	}

	if caps.Os == "" {
		t.Error("Os should not be empty")
	}

	if caps.Os != runtime.GOOS {
		t.Errorf("Os = %s, want %s", caps.Os, runtime.GOOS)
	}

	// Architecture should be detected
	if caps.NativeArch == pb.Architecture_ARCH_UNSPECIFIED {
		// Only fail if we're on a known arch
		switch runtime.GOARCH {
		case "amd64", "arm64", "arm":
			t.Errorf("NativeArch should be detected for %s", runtime.GOARCH)
		}
	}
}

func TestDetectArch(t *testing.T) {
	arch := detectArch()

	// Test that arch matches runtime.GOARCH
	switch runtime.GOARCH {
	case "amd64":
		if arch != pb.Architecture_ARCH_X86_64 {
			t.Errorf("detectArch() = %v for amd64, want ARCH_X86_64", arch)
		}
	case "arm64":
		if arch != pb.Architecture_ARCH_ARM64 {
			t.Errorf("detectArch() = %v for arm64, want ARCH_ARM64", arch)
		}
	case "arm":
		if arch != pb.Architecture_ARCH_ARMV7 {
			t.Errorf("detectArch() = %v for arm, want ARCH_ARMV7", arch)
		}
	default:
		if arch != pb.Architecture_ARCH_UNSPECIFIED {
			t.Errorf("detectArch() = %v for unknown arch, want ARCH_UNSPECIFIED", arch)
		}
	}
}

func TestDetectDocker(t *testing.T) {
	// Just ensure it doesn't panic
	result := detectDocker()
	t.Logf("Docker available: %v", result)
}

func TestDetectCpp(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp() returned nil")
	}

	if caps.Compilers == nil {
		t.Error("Compilers slice should not be nil")
	}

	t.Logf("C++ compilers found: %v", caps.Compilers)
	t.Logf("Cross-compile available: %v", caps.CrossCompile)
}

func TestDetectGo(t *testing.T) {
	caps := detectGo()

	// Go should be available since we're running Go tests
	if caps == nil {
		t.Fatal("detectGo() returned nil - Go should be available")
	}

	if caps.Version == "" {
		t.Error("Go version should not be empty")
	}

	if !caps.CrossCompile {
		t.Error("Go should support cross-compile")
	}

	t.Logf("Go version: %s", caps.Version)
}

func TestDetectRust(t *testing.T) {
	caps := detectRust()

	// Rust may or may not be installed
	if caps != nil {
		t.Logf("Rust toolchains: %v", caps.Toolchains)
		t.Logf("Rust targets: %v", caps.Targets)
	} else {
		t.Log("Rust not installed (this is OK)")
	}
}

func TestDetectNode(t *testing.T) {
	caps := detectNode()

	// Node may or may not be installed
	if caps != nil {
		if len(caps.Versions) == 0 {
			t.Error("Node versions should not be empty if detected")
		}
		t.Logf("Node versions: %v", caps.Versions)
		t.Logf("Package managers: %v", caps.PackageManagers)
	} else {
		t.Log("Node not installed (this is OK)")
	}
}

func TestDetectFlutter(t *testing.T) {
	caps := detectFlutter()

	// Flutter may or may not be installed
	if caps != nil {
		t.Logf("Flutter SDK version: %s", caps.SdkVersion)
		t.Logf("Flutter platforms: %v", caps.Platforms)
		t.Logf("Android SDK: %v", caps.AndroidSdk)
		t.Logf("Xcode: %v", caps.XcodeAvailable)
	} else {
		t.Log("Flutter not installed (this is OK)")
	}
}

func TestDetect_Hostname(t *testing.T) {
	caps := Detect()

	// Hostname should be set (may be empty on some systems but shouldn't panic)
	t.Logf("Hostname: %s", caps.Hostname)
}

func TestDetect_AllCapabilities(t *testing.T) {
	caps := Detect()

	// Log all capabilities for debugging
	t.Logf("Full capabilities:")
	t.Logf("  Hostname: %s", caps.Hostname)
	t.Logf("  OS: %s", caps.Os)
	t.Logf("  Arch: %v", caps.NativeArch)
	t.Logf("  CPUs: %d", caps.CpuCores)
	t.Logf("  Docker: %v", caps.DockerAvailable)

	if caps.Cpp != nil {
		t.Logf("  C++: compilers=%v, cross=%v", caps.Cpp.Compilers, caps.Cpp.CrossCompile)
	}
	if caps.Go != nil {
		t.Logf("  Go: version=%s", caps.Go.Version)
	}
	if caps.Rust != nil {
		t.Logf("  Rust: toolchains=%v", caps.Rust.Toolchains)
	}
	if caps.Nodejs != nil {
		t.Logf("  Node: versions=%v", caps.Nodejs.Versions)
	}
	if caps.Flutter != nil {
		t.Logf("  Flutter: sdk=%s", caps.Flutter.SdkVersion)
	}
}
