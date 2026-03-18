package capability

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
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
	if runtime.GOOS == "windows" {
		t.Skip("Windows MSVC detection uses registry, not PATH")
	}

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

func TestDetectArchForGOARCH(t *testing.T) {
	tests := []struct {
		name   string
		goarch string
		want   pb.Architecture
	}{
		{name: "amd64", goarch: "amd64", want: pb.Architecture_ARCH_X86_64},
		{name: "arm64", goarch: "arm64", want: pb.Architecture_ARCH_ARM64},
		{name: "arm", goarch: "arm", want: pb.Architecture_ARCH_ARMV7},
		{name: "unknown", goarch: "riscv64", want: pb.Architecture_ARCH_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := detectArchForGOARCH(tt.goarch); got != tt.want {
				t.Fatalf("detectArchForGOARCH(%q) = %v, want %v", tt.goarch, got, tt.want)
			}
		})
	}
}

func TestDetectMemoryForGOOS(t *testing.T) {
	tests := []struct {
		name         string
		goos         string
		want         int64
		linuxCalls   int
		darwinCalls  int
		windowsCalls int
	}{
		{name: "linux", goos: "linux", want: 11, linuxCalls: 1},
		{name: "darwin", goos: "darwin", want: 22, darwinCalls: 1},
		{name: "windows", goos: "windows", want: 33, windowsCalls: 1},
		{name: "unknown", goos: "plan9", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			linuxCalls := 0
			darwinCalls := 0
			windowsCalls := 0

			got := detectMemoryForGOOS(
				tt.goos,
				func() int64 {
					linuxCalls++
					return 11
				},
				func() int64 {
					darwinCalls++
					return 22
				},
				func() int64 {
					windowsCalls++
					return 33
				},
			)

			if got != tt.want {
				t.Fatalf("detectMemoryForGOOS(%q) = %d, want %d", tt.goos, got, tt.want)
			}
			if linuxCalls != tt.linuxCalls {
				t.Fatalf("linux calls = %d, want %d", linuxCalls, tt.linuxCalls)
			}
			if darwinCalls != tt.darwinCalls {
				t.Fatalf("darwin calls = %d, want %d", darwinCalls, tt.darwinCalls)
			}
			if windowsCalls != tt.windowsCalls {
				t.Fatalf("windows calls = %d, want %d", windowsCalls, tt.windowsCalls)
			}
		})
	}
}

func TestDetectCppForGOOS_NonWindows(t *testing.T) {
	lookups := map[string]bool{
		"gcc":                   true,
		"clang":                 true,
		"aarch64-linux-gnu-gcc": true,
	}

	got := detectCppForGOOS(
		"darwin",
		func(name string) (string, error) {
			if lookups[name] {
				return "/tmp/" + name, nil
			}
			return "", errors.New("missing")
		},
		func() *MSVCInfo {
			t.Fatal("MSVC detection should not run on non-Windows")
			return nil
		},
	)

	if got == nil {
		t.Fatal("detectCppForGOOS() returned nil")
	}

	wantCompilers := []string{"gcc", "clang"}
	if len(got.Compilers) != len(wantCompilers) {
		t.Fatalf("compilers = %v, want %v", got.Compilers, wantCompilers)
	}
	for i, want := range wantCompilers {
		if got.Compilers[i] != want {
			t.Fatalf("compilers[%d] = %q, want %q", i, got.Compilers[i], want)
		}
	}
	if !got.CrossCompile {
		t.Fatal("CrossCompile = false, want true")
	}
}

func TestDetectCppForGOOS_WindowsWithMSVCAndMinGW(t *testing.T) {
	got := detectCppForGOOS(
		"windows",
		func(name string) (string, error) {
			switch name {
			case "clang++", "x86_64-w64-mingw32-gcc":
				return `C:\toolchains\` + name, nil
			default:
				return "", errors.New("missing")
			}
		},
		func() *MSVCInfo {
			return &MSVCInfo{
				Available:     true,
				Version:       "2022",
				Architectures: []string{"x64", "arm64"},
				WindowsSDK:    "10.0.22621.0",
			}
		},
	)

	if got == nil {
		t.Fatal("detectCppForGOOS() returned nil")
	}

	wantCompilers := []string{"clang++", "cl.exe", "x86_64-w64-mingw32-gcc"}
	if len(got.Compilers) != len(wantCompilers) {
		t.Fatalf("compilers = %v, want %v", got.Compilers, wantCompilers)
	}
	for i, want := range wantCompilers {
		if got.Compilers[i] != want {
			t.Fatalf("compilers[%d] = %q, want %q", i, got.Compilers[i], want)
		}
	}
	if got.MsvcVersion != "2022" {
		t.Fatalf("MsvcVersion = %q, want %q", got.MsvcVersion, "2022")
	}
	if !got.HasWindowsSdk {
		t.Fatal("HasWindowsSdk = false, want true")
	}
	if !got.CrossCompile {
		t.Fatal("CrossCompile = false, want true")
	}
	if len(got.MsvcArchitectures) != 2 {
		t.Fatalf("MsvcArchitectures = %v, want 2 entries", got.MsvcArchitectures)
	}
}

func TestDetectCppForGOOS_WindowsWithoutMSVC(t *testing.T) {
	got := detectCppForGOOS(
		"windows",
		func(string) (string, error) {
			return "", errors.New("missing")
		},
		func() *MSVCInfo {
			return nil
		},
	)

	if got == nil {
		t.Fatal("detectCppForGOOS() returned nil")
	}
	if len(got.Compilers) != 0 {
		t.Fatalf("compilers = %v, want empty", got.Compilers)
	}
	if got.CrossCompile {
		t.Fatal("CrossCompile = true, want false")
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

func TestDetectCpp_PathModification_DeduplicatedOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows MSVC detection uses registry, not PATH")
	}

	origPath := os.Getenv("PATH")
	defer os.Setenv("PATH", origPath)

	tempDir := t.TempDir()
	for _, name := range []string{"gcc", "clang", "x86_64-linux-gnu-gcc"} {
		path := filepath.Join(tempDir, name)
		if err := os.WriteFile(path, []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
			t.Fatal(err)
		}
	}

	os.Setenv("PATH", tempDir)

	caps := detectCpp()
	if caps == nil {
		t.Fatal("detectCpp returned nil")
	}

	compilers := append([]string(nil), caps.Compilers...)
	sort.Strings(compilers)
	want := []string{"clang", "gcc"}
	if len(compilers) != len(want) {
		t.Fatalf("compilers = %v, want %v", compilers, want)
	}
	for i, compiler := range want {
		if compilers[i] != compiler {
			t.Fatalf("compilers[%d] = %q, want %q", i, compilers[i], compiler)
		}
	}
	if !caps.CrossCompile {
		t.Fatal("CrossCompile = false, want true")
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

func TestFindClExe(t *testing.T) {
	t.Run("prefers hostx64", func(t *testing.T) {
		vcToolsPath := t.TempDir()
		clPath := filepath.Join(vcToolsPath, "bin", "Hostx64", "x64", "cl.exe")
		if err := os.MkdirAll(filepath.Dir(clPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(clPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		got := findClExe(vcToolsPath)
		if got != clPath {
			t.Fatalf("findClExe() = %q, want %q", got, clPath)
		}
	})

	t.Run("falls back to hostx86", func(t *testing.T) {
		vcToolsPath := t.TempDir()
		clPath := filepath.Join(vcToolsPath, "bin", "Hostx86", "x64", "cl.exe")
		if err := os.MkdirAll(filepath.Dir(clPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(clPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		got := findClExe(vcToolsPath)
		if got != clPath {
			t.Fatalf("findClExe() = %q, want %q", got, clPath)
		}
	})

	t.Run("missing", func(t *testing.T) {
		if got := findClExe(t.TempDir()); got != "" {
			t.Fatalf("findClExe() = %q, want empty", got)
		}
	})
}

func TestDetectMSVCArchitectures(t *testing.T) {
	vcToolsPath := t.TempDir()
	hostArch := "Hostx64"
	if runtime.GOARCH == "386" {
		hostArch = "Hostx86"
	}

	for _, arch := range []string{"x64", "arm64"} {
		clPath := filepath.Join(vcToolsPath, "bin", hostArch, arch, "cl.exe")
		if err := os.MkdirAll(filepath.Dir(clPath), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(clPath, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(vcToolsPath, "bin", hostArch, "x86"), 0755); err != nil {
		t.Fatal(err)
	}

	got := detectMSVCArchitectures(vcToolsPath)
	want := []string{"arm64", "x64"}
	sort.Strings(got)
	if len(got) != len(want) {
		t.Fatalf("detectMSVCArchitectures() = %v, want %v", got, want)
	}
	for i, arch := range want {
		if got[i] != arch {
			t.Fatalf("architectures[%d] = %q, want %q", i, got[i], arch)
		}
	}
}

func TestDetectWindowsSDK(t *testing.T) {
	sdkRoot := filepath.Join(t.TempDir(), "sdk")
	t.Setenv("WindowsSdkDir", sdkRoot)

	for _, version := range []string{"10.0.19041.0", "10.0.22621.0"} {
		ucrt := filepath.Join(sdkRoot, "Include", version, "ucrt")
		if err := os.MkdirAll(ucrt, 0755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(sdkRoot, "Include", "not-an-sdk"), 0755); err != nil {
		t.Fatal(err)
	}

	if got := detectWindowsSDK(); got != "10.0.22621.0" {
		t.Fatalf("detectWindowsSDK() = %q, want %q", got, "10.0.22621.0")
	}
}

func TestDetectMSVCFromEnv(t *testing.T) {
	rootDir := filepath.Join(t.TempDir(), "VS2022")
	vcInstallDir := filepath.Join(rootDir, "VC")
	vcToolsPath := filepath.Join(vcInstallDir, "Tools", "MSVC", "14.39.33519")
	clPath := filepath.Join(vcToolsPath, "bin", "Hostx64", "x64", "cl.exe")
	arm64ClPath := filepath.Join(vcToolsPath, "bin", "Hostx64", "arm64", "cl.exe")

	for _, file := range []string{clPath, arm64ClPath} {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
	}

	t.Setenv("VCINSTALLDIR", vcInstallDir)
	t.Setenv("VCToolsVersion", "14.39.33519")
	t.Setenv("WindowsSDKVersion", "10.0.22621.0\\")

	got := detectMSVCFromEnv()
	if got == nil {
		t.Fatal("detectMSVCFromEnv() returned nil")
	}
	if !got.Available {
		t.Fatal("Available = false, want true")
	}
	if got.Version != "2022" {
		t.Fatalf("Version = %q, want %q", got.Version, "2022")
	}
	if got.InstallDir != rootDir {
		t.Fatalf("InstallDir = %q, want %q", got.InstallDir, rootDir)
	}
	if got.VCToolsPath != vcToolsPath {
		t.Fatalf("VCToolsPath = %q, want %q", got.VCToolsPath, vcToolsPath)
	}
	if got.ClExePath != clPath {
		t.Fatalf("ClExePath = %q, want %q", got.ClExePath, clPath)
	}
	if got.WindowsSDK != "10.0.22621.0" {
		t.Fatalf("WindowsSDK = %q, want %q", got.WindowsSDK, "10.0.22621.0")
	}
	if len(got.Architectures) != 2 {
		t.Fatalf("Architectures = %v, want 2 entries", got.Architectures)
	}
}

func TestDetectMSVCForGOOS(t *testing.T) {
	t.Run("non-windows returns nil", func(t *testing.T) {
		got := detectMSVCForGOOS(
			"darwin",
			func() *MSVCInfo {
				t.Fatal("fromEnv should not be called")
				return nil
			},
			func() *MSVCInfo {
				t.Fatal("fromVSWhere should not be called")
				return nil
			},
			func() *MSVCInfo {
				t.Fatal("fromKnownPaths should not be called")
				return nil
			},
		)
		if got != nil {
			t.Fatalf("detectMSVCForGOOS() = %v, want nil", got)
		}
	})

	t.Run("prefers env then vswhere then known paths", func(t *testing.T) {
		envCalls := 0
		vsWhereCalls := 0
		knownCalls := 0
		envInfo := &MSVCInfo{Available: true, Version: "env"}
		vsWhereInfo := &MSVCInfo{Available: true, Version: "vswhere"}
		knownInfo := &MSVCInfo{Available: true, Version: "known"}

		got := detectMSVCForGOOS(
			"windows",
			func() *MSVCInfo {
				envCalls++
				return envInfo
			},
			func() *MSVCInfo {
				vsWhereCalls++
				return vsWhereInfo
			},
			func() *MSVCInfo {
				knownCalls++
				return knownInfo
			},
		)
		if got != envInfo {
			t.Fatalf("detectMSVCForGOOS() returned %v, want env result", got)
		}
		if envCalls != 1 || vsWhereCalls != 0 || knownCalls != 0 {
			t.Fatalf("call counts = env:%d vswhere:%d known:%d, want 1/0/0", envCalls, vsWhereCalls, knownCalls)
		}

		envCalls = 0
		vsWhereCalls = 0
		knownCalls = 0
		got = detectMSVCForGOOS(
			"windows",
			func() *MSVCInfo {
				envCalls++
				return &MSVCInfo{Available: false}
			},
			func() *MSVCInfo {
				vsWhereCalls++
				return vsWhereInfo
			},
			func() *MSVCInfo {
				knownCalls++
				return knownInfo
			},
		)
		if got != vsWhereInfo {
			t.Fatalf("detectMSVCForGOOS() returned %v, want vswhere result", got)
		}
		if envCalls != 1 || vsWhereCalls != 1 || knownCalls != 0 {
			t.Fatalf("call counts = env:%d vswhere:%d known:%d, want 1/1/0", envCalls, vsWhereCalls, knownCalls)
		}

		envCalls = 0
		vsWhereCalls = 0
		knownCalls = 0
		got = detectMSVCForGOOS(
			"windows",
			func() *MSVCInfo {
				envCalls++
				return nil
			},
			func() *MSVCInfo {
				vsWhereCalls++
				return &MSVCInfo{Available: false}
			},
			func() *MSVCInfo {
				knownCalls++
				return knownInfo
			},
		)
		if got != knownInfo {
			t.Fatalf("detectMSVCForGOOS() returned %v, want known-paths result", got)
		}
		if envCalls != 1 || vsWhereCalls != 1 || knownCalls != 1 {
			t.Fatalf("call counts = env:%d vswhere:%d known:%d, want 1/1/1", envCalls, vsWhereCalls, knownCalls)
		}
	})
}

func TestMSVCInfoPaths(t *testing.T) {
	installDir := filepath.Join(t.TempDir(), "vs")
	vcVarsPath := filepath.Join(installDir, "VC", "Auxiliary", "Build", "vcvars64.bat")
	if err := os.MkdirAll(filepath.Dir(vcVarsPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(vcVarsPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	vcToolsPath := filepath.Join(installDir, "VC", "Tools", "MSVC", "14.39.33519")
	hostArch := "Hostx64"
	if runtime.GOARCH == "386" {
		hostArch = "Hostx86"
	}
	clPath := filepath.Join(vcToolsPath, "bin", hostArch, "arm64", "cl.exe")
	if err := os.MkdirAll(filepath.Dir(clPath), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(clPath, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	info := &MSVCInfo{
		Available:   true,
		InstallDir:  installDir,
		VCToolsPath: vcToolsPath,
	}

	if got := info.GetVCVarsPath(); got != vcVarsPath {
		t.Fatalf("GetVCVarsPath() = %q, want %q", got, vcVarsPath)
	}
	if got := info.GetClExeForArch("arm64"); got != clPath {
		t.Fatalf("GetClExeForArch() = %q, want %q", got, clPath)
	}
}
