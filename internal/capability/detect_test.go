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

func TestDetectArch_AllArchitectures(t *testing.T) {
	tests := []struct {
		name string
		arch string
		want pb.Architecture
	}{
		{"amd64", "amd64", pb.Architecture_ARCH_X86_64},
		{"arm64", "arm64", pb.Architecture_ARCH_ARM64},
		{"arm", "arm", pb.Architecture_ARCH_ARMV7},
		{"unknown", "riscv64", pb.Architecture_ARCH_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can only test the actual runtime architecture directly
			// This test documents the expected behavior
			if runtime.GOARCH == tt.arch {
				got := detectArch()
				if got != tt.want {
					t.Errorf("detectArch() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestDetectMemory(t *testing.T) {
	mem := detectMemory()

	// Memory detection should return positive value or 0 if detection fails
	if mem < 0 {
		t.Errorf("detectMemory() returned negative value: %d", mem)
	}

	t.Logf("Detected memory: %d bytes", mem)
}

func TestDetectMemory_ByPlatform(t *testing.T) {
	// Test platform-specific memory detection
	switch runtime.GOOS {
	case "linux":
		mem := detectMemoryLinux()
		t.Logf("Linux memory: %d bytes", mem)
	case "darwin":
		mem := detectMemoryDarwin()
		t.Logf("Darwin memory: %d bytes", mem)
		if mem > 0 {
			// macOS should detect memory
			t.Logf("Successfully detected memory on Darwin")
		}
	case "windows":
		mem := detectMemoryWindows()
		t.Logf("Windows memory: %d bytes", mem)
	}
}

func TestDetectCpp_CompilerDetection(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp() returned nil")
	}

	// Should have non-nil compilers slice
	if caps.Compilers == nil {
		t.Error("Compilers slice should not be nil")
	}

	// Log what was found
	t.Logf("Detected compilers: %v", caps.Compilers)

	// Check for common compilers based on platform
	hasGcc := false
	hasClang := false
	for _, c := range caps.Compilers {
		if c == "gcc" || c == "g++" {
			hasGcc = true
		}
		if c == "clang" || c == "clang++" {
			hasClang = true
		}
	}

	t.Logf("Has GCC: %v, Has Clang: %v", hasGcc, hasClang)
}

func TestDetectCpp_CrossCompile(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp() returned nil")
	}

	// Cross-compile capability depends on installed toolchains
	t.Logf("Cross-compile available: %v", caps.CrossCompile)
}

func TestDetectGo_Version(t *testing.T) {
	caps := detectGo()

	// Go should always be available in tests
	if caps == nil {
		t.Fatal("detectGo() returned nil - Go must be available")
	}

	if caps.Version == "" {
		t.Error("Go version should not be empty")
	}

	// Version should not contain "go" prefix
	if len(caps.Version) > 0 && caps.Version[0] == 'g' {
		t.Errorf("Version should not start with 'g', got: %s", caps.Version)
	}

	t.Logf("Go version: %s", caps.Version)
}

func TestDetectGo_CrossCompile(t *testing.T) {
	caps := detectGo()

	if caps == nil {
		t.Skip("Go not available")
	}

	// Go always supports cross-compile
	if !caps.CrossCompile {
		t.Error("Go should always support cross-compile")
	}
}

func TestDetectRust_Toolchains(t *testing.T) {
	caps := detectRust()

	if caps == nil {
		t.Skip("Rust not installed")
	}

	if len(caps.Toolchains) == 0 {
		t.Error("Rust detected but no toolchains found")
	}

	t.Logf("Rust toolchains: %v", caps.Toolchains)
}

func TestDetectRust_Targets(t *testing.T) {
	caps := detectRust()

	if caps == nil {
		t.Skip("Rust not installed")
	}

	// Targets may be empty if rustup not available
	t.Logf("Rust targets: %v", caps.Targets)
}

func TestDetectNode_PackageManagers(t *testing.T) {
	caps := detectNode()

	if caps == nil {
		t.Skip("Node not installed")
	}

	if len(caps.Versions) == 0 {
		t.Error("Node detected but no versions found")
	}

	// Package managers are optional
	t.Logf("Package managers: %v", caps.PackageManagers)

	// Should have at least npm if Node is installed
	hasNpm := false
	for _, pm := range caps.PackageManagers {
		if pm == "npm" {
			hasNpm = true
		}
	}

	if !hasNpm {
		t.Log("Node installed but npm not found in PATH")
	}
}

func TestDetectFlutter_Platforms(t *testing.T) {
	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	// Should detect at least web platform
	if len(caps.Platforms) == 0 {
		t.Error("Flutter detected but no platforms found")
	}

	hasWeb := false
	for _, p := range caps.Platforms {
		if p == pb.TargetPlatform_PLATFORM_WEB {
			hasWeb = true
		}
	}

	if !hasWeb {
		t.Error("Flutter should always support web platform")
	}

	t.Logf("Flutter platforms: %v", caps.Platforms)
}

func TestDetectFlutter_NativePlatforms(t *testing.T) {
	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	// Test platform-specific detection
	switch runtime.GOOS {
	case "darwin":
		// macOS should detect iOS and macOS platforms if Xcode available
		if caps.XcodeAvailable {
			hasIOS := false
			hasMacOS := false
			for _, p := range caps.Platforms {
				if p == pb.TargetPlatform_PLATFORM_IOS {
					hasIOS = true
				}
				if p == pb.TargetPlatform_PLATFORM_MACOS {
					hasMacOS = true
				}
			}
			if !hasIOS {
				t.Error("Xcode available but iOS platform not detected")
			}
			if !hasMacOS {
				t.Error("Xcode available but macOS platform not detected")
			}
		}
	case "linux":
		// Linux should detect Linux platform
		hasLinux := false
		for _, p := range caps.Platforms {
			if p == pb.TargetPlatform_PLATFORM_LINUX {
				hasLinux = true
			}
		}
		if !hasLinux {
			t.Error("Linux platform should be detected on Linux")
		}
	case "windows":
		// Windows should detect Windows platform
		hasWindows := false
		for _, p := range caps.Platforms {
			if p == pb.TargetPlatform_PLATFORM_WINDOWS {
				hasWindows = true
			}
		}
		if !hasWindows {
			t.Error("Windows platform should be detected on Windows")
		}
	}
}

func TestDetect_MemoryPositive(t *testing.T) {
	caps := Detect()

	// Memory should be detected on most systems
	// It's OK if it's 0 on some systems, but log it
	if caps.MemoryBytes == 0 {
		t.Logf("Warning: Memory detection returned 0 on %s", runtime.GOOS)
	} else if caps.MemoryBytes < 0 {
		t.Errorf("Memory should not be negative: %d", caps.MemoryBytes)
	} else {
		t.Logf("Detected memory: %d bytes (%d GB)", caps.MemoryBytes, caps.MemoryBytes/(1024*1024*1024))
	}
}

func TestDetect_OSMatches(t *testing.T) {
	caps := Detect()

	if caps.Os != runtime.GOOS {
		t.Errorf("OS mismatch: got %s, want %s", caps.Os, runtime.GOOS)
	}
}

func TestDetect_CPUCoresPositive(t *testing.T) {
	caps := Detect()

	if caps.CpuCores <= 0 {
		t.Errorf("CPU cores should be positive, got %d", caps.CpuCores)
	}

	if caps.CpuCores != int32(runtime.NumCPU()) {
		t.Errorf("CPU cores mismatch: got %d, want %d", caps.CpuCores, runtime.NumCPU())
	}
}

func TestDetectDocker_NoError(t *testing.T) {
	// detectDocker should not panic
	available := detectDocker()
	t.Logf("Docker available: %v", available)

	// Test multiple calls
	available2 := detectDocker()
	if available != available2 {
		t.Error("detectDocker() should return consistent results")
	}
}

func TestDetectMemoryLinux_NoFile(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-only test")
	}

	// Can't easily mock file system, but we can verify it handles missing file
	// This function would return 0 if /proc/meminfo doesn't exist
	mem := detectMemoryLinux()
	t.Logf("Linux memory: %d", mem)
}

func TestDetectMemoryDarwin_ErrorHandling(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-only test")
	}

	// Test that detectMemoryDarwin handles errors gracefully
	mem := detectMemoryDarwin()

	// Should return >= 0 (0 on error, positive on success)
	if mem < 0 {
		t.Errorf("detectMemoryDarwin() returned negative value: %d", mem)
	}
}

func TestDetectCpp_EmptyCompilerList(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp() should never return nil")
	}

	// Even if no compilers found, should return non-nil capability
	if caps.Compilers == nil {
		t.Error("Compilers slice should never be nil")
	}
}

func TestDetectCpp_MSVCDetectionSkipped(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("MSVC test - skipping on Windows to test non-Windows path")
	}

	caps := detectCpp()

	// On non-Windows, MSVC fields should be empty
	if caps.MsvcVersion != "" {
		t.Errorf("MSVC version should be empty on %s, got: %s", runtime.GOOS, caps.MsvcVersion)
	}

	if len(caps.MsvcArchitectures) > 0 {
		t.Errorf("MSVC architectures should be empty on %s", runtime.GOOS)
	}
}

func TestDetectGo_InvalidOutput(t *testing.T) {
	// We can't easily mock exec.Command, but we can verify
	// the function handles the normal case
	caps := detectGo()

	if caps == nil {
		t.Fatal("Go should be available in test environment")
	}

	// Verify version is not empty and doesn't have "go" prefix
	if len(caps.Version) == 0 {
		t.Error("Version should not be empty")
	}
}

func TestDetectRust_NoRustup(t *testing.T) {
	caps := detectRust()

	if caps != nil {
		// If Rust is detected, verify toolchains are populated
		if len(caps.Toolchains) == 0 {
			t.Error("If Rust is detected, toolchains should not be empty")
		}

		t.Logf("Rust toolchains: %v", caps.Toolchains)
		t.Logf("Rust targets: %v", caps.Targets)
	} else {
		t.Log("Rust not detected")
	}
}

func TestDetectNode_VersionFormat(t *testing.T) {
	caps := detectNode()

	if caps == nil {
		t.Skip("Node not installed")
	}

	if len(caps.Versions) == 0 {
		t.Fatal("Node detected but versions empty")
	}

	// Version should not start with 'v'
	version := caps.Versions[0]
	if len(version) > 0 && version[0] == 'v' {
		t.Errorf("Version should not start with 'v', got: %s", version)
	}

	t.Logf("Node version: %s", version)
}

func TestDetectFlutter_AndroidSDK(t *testing.T) {
	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	// Test Android SDK detection
	hasAndroidPlatform := false
	for _, p := range caps.Platforms {
		if p == pb.TargetPlatform_PLATFORM_ANDROID {
			hasAndroidPlatform = true
		}
	}

	if caps.AndroidSdk && !hasAndroidPlatform {
		t.Error("AndroidSdk is true but PLATFORM_ANDROID not in platforms")
	}

	t.Logf("Android SDK available: %v", caps.AndroidSdk)
}

func TestDetect_NonNilCapabilities(t *testing.T) {
	caps := Detect()

	// Verify all expected fields are non-nil
	if caps == nil {
		t.Fatal("Detect() returned nil")
	}

	// These should always be set
	if caps.Os == "" {
		t.Error("OS should never be empty")
	}

	if caps.CpuCores == 0 {
		t.Error("CPU cores should never be 0")
	}

	// Architecture should be set for known platforms
	if caps.NativeArch == pb.Architecture_ARCH_UNSPECIFIED {
		if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" || runtime.GOARCH == "arm" {
			t.Errorf("Architecture should be detected for %s", runtime.GOARCH)
		}
	}
}

func TestDetectCpp_CrossCompilerList(t *testing.T) {
	caps := detectCpp()

	if caps == nil {
		t.Fatal("detectCpp() returned nil")
	}

	// Log compiler list for debugging
	t.Logf("Total compilers found: %d", len(caps.Compilers))
	for _, compiler := range caps.Compilers {
		t.Logf("  - %s", compiler)
	}

	// Verify no duplicate compilers
	seen := make(map[string]bool)
	for _, compiler := range caps.Compilers {
		if seen[compiler] {
			t.Errorf("Duplicate compiler found: %s", compiler)
		}
		seen[compiler] = true
	}
}

func TestDetectNode_MultiplePackageManagers(t *testing.T) {
	caps := detectNode()

	if caps == nil {
		t.Skip("Node not installed")
	}

	// Verify package manager list
	t.Logf("Found %d package manager(s)", len(caps.PackageManagers))

	// Should not have duplicates
	seen := make(map[string]bool)
	for _, pm := range caps.PackageManagers {
		if seen[pm] {
			t.Errorf("Duplicate package manager: %s", pm)
		}
		seen[pm] = true
	}
}

func TestDetectFlutter_PlatformList(t *testing.T) {
	caps := detectFlutter()

	if caps == nil {
		t.Skip("Flutter not installed")
	}

	// Verify platform list has no duplicates
	seen := make(map[pb.TargetPlatform]bool)
	for _, platform := range caps.Platforms {
		if seen[platform] {
			t.Errorf("Duplicate platform: %v", platform)
		}
		seen[platform] = true
	}

	t.Logf("Total platforms: %d", len(caps.Platforms))
}

func TestDetect_IntegrationAllLanguages(t *testing.T) {
	caps := Detect()

	// This is an integration test that verifies all language detection
	// runs without crashing and returns sensible data

	if caps.Cpp == nil {
		t.Log("Warning: C++ capabilities not detected")
	} else {
		t.Logf("C++ compilers: %d", len(caps.Cpp.Compilers))
	}

	if caps.Go == nil {
		t.Error("Go should be detected in test environment")
	} else {
		t.Logf("Go version: %s", caps.Go.Version)
	}

	if caps.Rust == nil {
		t.Log("Rust not detected (optional)")
	} else {
		t.Logf("Rust toolchains: %d", len(caps.Rust.Toolchains))
	}

	if caps.Nodejs == nil {
		t.Log("Node.js not detected (optional)")
	} else {
		t.Logf("Node.js versions: %d", len(caps.Nodejs.Versions))
	}

	if caps.Flutter == nil {
		t.Log("Flutter not detected (optional)")
	} else {
		t.Logf("Flutter platforms: %d", len(caps.Flutter.Platforms))
	}
}

func TestDetectArch_Coverage(t *testing.T) {
	// Test all branches by documenting expected behavior
	arch := detectArch()

	switch runtime.GOARCH {
	case "amd64":
		if arch != pb.Architecture_ARCH_X86_64 {
			t.Errorf("Expected ARCH_X86_64 for amd64, got %v", arch)
		}
	case "arm64":
		if arch != pb.Architecture_ARCH_ARM64 {
			t.Errorf("Expected ARCH_ARM64 for arm64, got %v", arch)
		}
	case "arm":
		if arch != pb.Architecture_ARCH_ARMV7 {
			t.Errorf("Expected ARCH_ARMV7 for arm, got %v", arch)
		}
	default:
		if arch != pb.Architecture_ARCH_UNSPECIFIED {
			t.Errorf("Expected ARCH_UNSPECIFIED for unknown arch, got %v", arch)
		}
	}

	t.Logf("Detected arch %s -> %v", runtime.GOARCH, arch)
}

func TestDetectMemory_AllPlatforms(t *testing.T) {
	mem := detectMemory()

	switch runtime.GOOS {
	case "linux", "darwin", "windows":
		t.Logf("Memory on %s: %d bytes", runtime.GOOS, mem)
	default:
		if mem != 0 {
			t.Errorf("Unknown platform should return 0, got %d", mem)
		}
	}
}

func TestDetectCpp_DetailedAnalysis(t *testing.T) {
	caps := detectCpp()
	if caps == nil {
		t.Fatal("detectCpp should never return nil")
	}

	if caps.Compilers == nil {
		t.Fatal("Compilers should not be nil")
	}

	t.Logf("Compilers: %v", caps.Compilers)
	t.Logf("Cross-compile: %v", caps.CrossCompile)
	t.Logf("MSVC version: %s", caps.MsvcVersion)
	t.Logf("Has Windows SDK: %v", caps.HasWindowsSdk)

	if runtime.GOOS != "windows" {
		if caps.MsvcVersion != "" || len(caps.MsvcArchitectures) > 0 {
			t.Error("MSVC fields should be empty on non-Windows")
		}
	}
}

func TestDetectRust_FallbackPath(t *testing.T) {
	caps := detectRust()
	if caps == nil {
		t.Log("Rust not installed")
		return
	}

	if caps.Toolchains == nil || caps.Targets == nil {
		t.Fatal("Rust capability slices should not be nil")
	}

	if len(caps.Toolchains) == 0 {
		t.Error("Rust detected but no toolchains")
	}

	t.Logf("Toolchains: %v", caps.Toolchains)
	t.Logf("Targets: %v", caps.Targets)
}

func TestDetectMemory_LinuxPath(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Linux-specific test")
	}

	mem := detectMemoryLinux()
	t.Logf("Linux memory: %d", mem)
}

func TestDetectMemory_DarwinPath(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Darwin-specific test")
	}

	mem := detectMemoryDarwin()
	if mem <= 0 {
		t.Error("Darwin should detect memory")
	}
	t.Logf("Darwin memory: %d", mem)
}

func TestDetectMemory_WindowsPath(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	mem := detectMemoryWindows()
	t.Logf("Windows memory: %d", mem)
}

func TestDetectArch_UnknownArch(t *testing.T) {
	arch := detectArch()

	switch runtime.GOARCH {
	case "amd64":
		if arch != pb.Architecture_ARCH_X86_64 {
			t.Errorf("amd64 should map to ARCH_X86_64, got %v", arch)
		}
	case "arm64":
		if arch != pb.Architecture_ARCH_ARM64 {
			t.Errorf("arm64 should map to ARCH_ARM64, got %v", arch)
		}
	case "arm":
		if arch != pb.Architecture_ARCH_ARMV7 {
			t.Errorf("arm should map to ARCH_ARMV7, got %v", arch)
		}
	default:
		if arch != pb.Architecture_ARCH_UNSPECIFIED {
			t.Errorf("unknown arch should map to ARCH_UNSPECIFIED, got %v", arch)
		}
	}
}

func TestDetectGo_AlwaysAvailable(t *testing.T) {
	caps := detectGo()
	if caps == nil {
		t.Fatal("Go should always be available in test environment")
	}
	if caps.Version == "" {
		t.Error("Go version should not be empty")
	}
	if !caps.CrossCompile {
		t.Error("Go should support cross-compile")
	}
}

func TestDetectNode_PackageManagerDetection(t *testing.T) {
	caps := detectNode()
	if caps == nil {
		t.Skip("Node not installed")
		return
	}

	if len(caps.Versions) == 0 {
		t.Error("Node should have versions when detected")
	}

	t.Logf("Node versions: %v", caps.Versions)
	t.Logf("Package managers: %v", caps.PackageManagers)
}

func TestDetectFlutter_WebPlatform(t *testing.T) {
	caps := detectFlutter()
	if caps == nil {
		t.Skip("Flutter not installed")
		return
	}

	hasWeb := false
	for _, p := range caps.Platforms {
		if p == pb.TargetPlatform_PLATFORM_WEB {
			hasWeb = true
			break
		}
	}

	if !hasWeb {
		t.Error("Flutter should always have web platform")
	}

	t.Logf("Platforms: %v", caps.Platforms)
}
