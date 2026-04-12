package capability

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// Detect detects the capabilities of the current system.
func Detect() *pb.WorkerCapabilities {
	hostname, _ := os.Hostname()

	caps := &pb.WorkerCapabilities{
		Hostname:        hostname,
		CpuCores:        int32(runtime.NumCPU()),
		MemoryBytes:     detectMemory(),
		NativeArch:      detectArch(),
		Os:              runtime.GOOS,
		DockerAvailable: detectDocker(),
	}

	// Detect C/C++ capabilities
	caps.Cpp = detectCpp()

	// Detect Go capabilities
	caps.Go = detectGo()

	// Detect Rust capabilities
	caps.Rust = detectRust()

	// Detect Node.js capabilities
	caps.Nodejs = detectNode()

	// Detect Flutter capabilities
	caps.Flutter = detectFlutter()

	// Detect Unity capabilities
	caps.Unity = detectUnity()

	return caps
}

func detectArch() pb.Architecture {
	return detectArchForGOARCH(runtime.GOARCH)
}

func detectArchForGOARCH(goarch string) pb.Architecture {
	switch goarch {
	case "amd64":
		return pb.Architecture_ARCH_X86_64
	case "arm64":
		return pb.Architecture_ARCH_ARM64
	case "arm":
		return pb.Architecture_ARCH_ARMV7
	default:
		return pb.Architecture_ARCH_UNSPECIFIED
	}
}

func detectMemory() int64 {
	return detectMemoryForGOOS(runtime.GOOS, detectMemoryLinux, detectMemoryDarwin, detectMemoryWindows)
}

func detectMemoryForGOOS(goos string, linux, darwin, windows func() int64) int64 {
	switch goos {
	case "linux":
		return linux()
	case "darwin":
		return darwin()
	case "windows":
		return windows()
	default:
		return 0
	}
}

func detectMemoryLinux() int64 {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0
	}

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MemTotal:") {
			// Format: "MemTotal:       16384000 kB"
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				valueStr := strings.TrimSpace(parts[1])
				fields := strings.Fields(valueStr)
				if len(fields) >= 1 {
					var value int64
					if _, err := fmt.Sscanf(fields[0], "%d", &value); err == nil {
						// Check unit (usually kB)
						if len(fields) >= 2 && strings.ToLower(fields[1]) == "kb" {
							return value * 1024 // kB to bytes
						}
						// Assume kB if no unit specified
						return value * 1024
					}
				}
			}
		}
	}
	return 0
}

func detectMemoryDarwin() int64 {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	var bytes int64
	_, err = fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &bytes)
	if err != nil {
		return 0
	}
	return bytes
}

func detectMemoryWindows() int64 {
	// Try PowerShell first (more reliable)
	psCmd := exec.Command("powershell", "-Command",
		"(Get-CimInstance Win32_ComputerSystem).TotalPhysicalMemory")
	if out, err := psCmd.Output(); err == nil {
		outStr := strings.TrimSpace(string(out))
		// Remove any BOM or special characters
		outStr = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, outStr)
		var bytes int64
		if _, err := fmt.Sscanf(outStr, "%d", &bytes); err == nil && bytes > 0 {
			return bytes
		}
	}

	// Fallback to wmic
	cmd := exec.Command("wmic", "ComputerSystem", "get", "TotalPhysicalMemory", "/value")
	out, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Handle Windows line endings and encoding
	outStr := string(out)
	outStr = strings.ReplaceAll(outStr, "\r\n", "\n")
	outStr = strings.ReplaceAll(outStr, "\r", "\n")

	for _, line := range strings.Split(outStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "TotalPhysicalMemory=") {
			valStr := strings.TrimPrefix(line, "TotalPhysicalMemory=")
			// Clean the value string
			valStr = strings.Map(func(r rune) rune {
				if r >= '0' && r <= '9' {
					return r
				}
				return -1
			}, valStr)
			var bytes int64
			if _, err := fmt.Sscanf(valStr, "%d", &bytes); err == nil {
				return bytes
			}
		}
	}
	return 0
}

func detectDocker() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	return cmd.Run() == nil
}

func detectCpp() *pb.CppCapability {
	return detectCppForGOOS(runtime.GOOS, exec.LookPath, DetectMSVC)
}

func detectCppForGOOS(goos string, lookPath func(string) (string, error), detectMSVCFn func() *MSVCInfo) *pb.CppCapability {
	cap := &pb.CppCapability{
		Compilers: make([]string, 0),
	}

	compilers := []string{"gcc", "g++", "clang", "clang++"}
	for _, c := range compilers {
		if _, err := lookPath(c); err == nil {
			cap.Compilers = append(cap.Compilers, c)
		}
	}

	// On Windows, detect MSVC
	if goos == "windows" {
		msvc := detectMSVCFn()
		if msvc != nil && msvc.Available {
			cap.Compilers = append(cap.Compilers, "cl.exe")
			// Set MSVC-specific fields
			cap.MsvcVersion = msvc.Version
			cap.MsvcArchitectures = msvc.Architectures
			if msvc.WindowsSDK != "" {
				cap.HasWindowsSdk = true
			}
		}

		// Check for MinGW compilers on Windows
		mingwCompilers := []string{
			"x86_64-w64-mingw32-gcc",
			"i686-w64-mingw32-gcc",
		}
		for _, c := range mingwCompilers {
			if _, err := lookPath(c); err == nil {
				cap.Compilers = append(cap.Compilers, c)
				cap.CrossCompile = true
			}
		}
	}

	// Check for cross-compile toolchains (Linux/macOS)
	if goos != "windows" {
		crossCompilers := []string{
			// Linux cross-compilers
			"aarch64-linux-gnu-gcc",
			"arm-linux-gnueabihf-gcc",
			"x86_64-linux-gnu-gcc",
			"x86_64-w64-mingw32-gcc",
			// Windows cross-compilers (MSYS2/MinGW)
			"aarch64-w64-mingw32-gcc",
			// macOS cross-compilers (via Homebrew)
			"aarch64-unknown-linux-gnu-gcc",
			"x86_64-unknown-linux-gnu-gcc",
		}
		for _, c := range crossCompilers {
			if _, err := lookPath(c); err == nil {
				cap.CrossCompile = true
				break
			}
		}
	}

	return cap
}

func detectGo() *pb.GoCapability {
	cmd := exec.Command("go", "version")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Parse "go version go1.22.0 darwin/arm64"
	parts := strings.Fields(string(out))
	if len(parts) >= 3 {
		version := strings.TrimPrefix(parts[2], "go")
		return &pb.GoCapability{
			Version:      version,
			CrossCompile: true, // Go always supports cross-compile
		}
	}

	return nil
}

func detectRust() *pb.RustCapability {
	cmd := exec.Command("rustc", "--version")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	cap := &pb.RustCapability{
		Toolchains: make([]string, 0),
		Targets:    make([]string, 0),
	}

	// Get installed toolchains
	rustupCmd := exec.Command("rustup", "toolchain", "list")
	if toolchains, err := rustupCmd.Output(); err == nil {
		for _, line := range strings.Split(string(toolchains), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				// Remove "(default)" suffix
				parts := strings.Fields(line)
				if len(parts) > 0 {
					cap.Toolchains = append(cap.Toolchains, parts[0])
				}
			}
		}
	}

	// Get installed targets
	targetsCmd := exec.Command("rustup", "target", "list", "--installed")
	if targets, err := targetsCmd.Output(); err == nil {
		for _, line := range strings.Split(string(targets), "\n") {
			line = strings.TrimSpace(line)
			if line != "" {
				cap.Targets = append(cap.Targets, line)
			}
		}
	}

	// Fall back if rustup not available
	if len(cap.Toolchains) == 0 {
		// Parse "rustc 1.75.0 (hash date)"
		parts := strings.Fields(string(out))
		if len(parts) >= 2 {
			cap.Toolchains = append(cap.Toolchains, "stable-"+parts[1])
		}
	}

	return cap
}

func detectNode() *pb.NodeCapability {
	cmd := exec.Command("node", "--version")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	version := strings.TrimSpace(strings.TrimPrefix(string(out), "v"))

	cap := &pb.NodeCapability{
		Versions:        []string{version},
		PackageManagers: make([]string, 0),
	}

	// Check package managers
	managers := []string{"npm", "yarn", "pnpm", "bun"}
	for _, m := range managers {
		if _, err := exec.LookPath(m); err == nil {
			cap.PackageManagers = append(cap.PackageManagers, m)
		}
	}

	return cap
}

func detectFlutter() *pb.FlutterCapability {
	cmd := exec.Command("flutter", "--version", "--machine")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}

	// Basic detection - just check if Flutter is available
	cap := &pb.FlutterCapability{
		Platforms: make([]pb.TargetPlatform, 0),
	}

	// Parse version from output
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		if strings.Contains(line, "frameworkVersion") {
			// Simple extraction
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				ver := strings.Trim(parts[1], ` ",`)
				cap.SdkVersion = ver
			}
		}
	}

	// Check for Android SDK — Flutter capability is Android-only for v0.4.0
	if os.Getenv("ANDROID_HOME") == "" && os.Getenv("ANDROID_SDK_ROOT") == "" {
		return nil
	}

	cap.AndroidSdk = true
	cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_ANDROID)

	return cap
}

func detectUnity() *pb.UnityCapability {
	versions := detectUnityVersions()
	if len(versions) == 0 {
		return nil
	}

	cap := &pb.UnityCapability{
		Versions:     versions,
		BuildTargets: make([]pb.TargetPlatform, 0),
	}

	// Detect build targets for the first (latest) Unity installation
	latestVersion := versions[0]
	editorPath := getUnityEditorPath(latestVersion)
	if editorPath != "" {
		cap.BuildTargets = detectUnityBuildTargets(editorPath)
	}

	return cap
}

func detectUnityVersions() []string {
	var patterns []string

	switch runtime.GOOS {
	case "darwin":
		patterns = []string{
			"/Applications/Unity/Hub/Editor/*/Unity.app",
			"/Applications/Unity*.app",
		}
	case "linux":
		home := os.Getenv("HOME")
		patterns = []string{
			home + "/Unity/Hub/Editor/*/Editor/Unity",
			home + "/Unity/Editor/Unity",
		}
	case "windows":
		patterns = []string{
			`C:\Program Files\Unity\Hub\Editor\*\Editor\Unity.exe`,
			`C:\Program Files\Unity\Editor\Unity.exe`,
		}
	}

	versions := make([]string, 0)
	seen := make(map[string]bool)

	for _, pattern := range patterns {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		for _, match := range matches {
			version := extractUnityVersion(match)
			if version != "" && !seen[version] {
				versions = append(versions, version)
				seen[version] = true
			}
		}
	}

	return versions
}

func getUnityEditorPath(version string) string {
	var pattern string

	switch runtime.GOOS {
	case "darwin":
		pattern = "/Applications/Unity/Hub/Editor/" + version + "/Unity.app"
	case "linux":
		home := os.Getenv("HOME")
		pattern = home + "/Unity/Hub/Editor/" + version + "/Editor/Unity"
	case "windows":
		pattern = `C:\Program Files\Unity\Hub\Editor\` + version + `\Editor\Unity.exe`
	}

	if _, err := os.Stat(pattern); err == nil {
		return pattern
	}

	// Try standalone Unity path
	switch runtime.GOOS {
	case "darwin":
		pattern = "/Applications/Unity.app"
		if _, err := os.Stat(pattern); err == nil {
			return pattern
		}
	}

	return ""
}

func extractUnityVersion(path string) string {
	// Extract version from path like:
	// /Applications/Unity/Hub/Editor/2022.3.10f1/Unity.app
	// ~/Unity/Hub/Editor/2022.3.10f1/Editor/Unity
	// C:\Program Files\Unity\Hub\Editor\2022.3.10f1\Editor\Unity.exe

	// For macOS Unity.app, the version is the parent dir of Contents
	// e.g., /Applications/Unity/Hub/Editor/2022.3.10f1/Unity.app
	if strings.HasSuffix(path, ".app") {
		dir := filepath.Dir(path)
		base := filepath.Base(dir)
		if isUnityVersion(base) {
			return base
		}
	}

	// For Linux/Windows paths like .../Editor/<version>/Editor/Unity
	// Split on "Editor" to find the version directory
	// Need to handle both / and \ separators
	sep := string(filepath.Separator)
	for _, candidate := range []string{sep + "Editor" + sep, "\\Editor\\"} {
		parts := strings.Split(path, candidate)
		if len(parts) >= 2 {
			rest := parts[1]
			version := strings.Split(rest, sep)[0]
			if version == "" {
				version = strings.Split(rest, "\\")[0]
			}
			if isUnityVersion(version) {
				return version
			}
		}
	}

	// Fallback: get base name of parent directory
	base := filepath.Base(filepath.Dir(path))
	if isUnityVersion(base) {
		return base
	}

	return ""
}

func isUnityVersion(s string) bool {
	if len(s) < 6 || strings.Count(s, ".") < 2 {
		return false
	}
	// Verify it starts with a year (4 digits)
	for i := 0; i < 4 && i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func detectUnityBuildTargets(editorPath string) []pb.TargetPlatform {
	targets := make([]pb.TargetPlatform, 0)
	seen := make(map[pb.TargetPlatform]bool)

	// Get the Unity.app Contents or the Editor directory
	editorDir := filepath.Dir(editorPath)
	if runtime.GOOS == "darwin" {
		// For macOS Unity.app, look inside Contents/PlaybackEngines
		editorDir = filepath.Join(editorPath, "Contents", "PlaybackEngines")
	} else {
		// For Linux/Windows, PlaybackEngines is alongside Editor
		editorDir = filepath.Join(editorDir, "PlaybackEngines")
	}

	// Check for Android
	androidDir := filepath.Join(editorDir, "AndroidPlayer")
	if _, err := os.Stat(androidDir); err == nil {
		if !seen[pb.TargetPlatform_PLATFORM_ANDROID] {
			targets = append(targets, pb.TargetPlatform_PLATFORM_ANDROID)
			seen[pb.TargetPlatform_PLATFORM_ANDROID] = true
		}
	}

	// Check for iOS (macOS only)
	if runtime.GOOS == "darwin" {
		iosDir := filepath.Join(editorDir, "iOSSupport")
		if _, err := os.Stat(iosDir); err == nil {
			if !seen[pb.TargetPlatform_PLATFORM_IOS] {
				targets = append(targets, pb.TargetPlatform_PLATFORM_IOS)
				seen[pb.TargetPlatform_PLATFORM_IOS] = true
			}
		}
	}

	// Check for Windows
	windowsDir := filepath.Join(editorDir, "WindowsStandaloneSupport")
	if _, err := os.Stat(windowsDir); err == nil {
		if !seen[pb.TargetPlatform_PLATFORM_WINDOWS] {
			targets = append(targets, pb.TargetPlatform_PLATFORM_WINDOWS)
			seen[pb.TargetPlatform_PLATFORM_WINDOWS] = true
		}
	}

	// Check for Linux
	linuxDir := filepath.Join(editorDir, "LinuxStandaloneSupport")
	if _, err := os.Stat(linuxDir); err == nil {
		if !seen[pb.TargetPlatform_PLATFORM_LINUX] {
			targets = append(targets, pb.TargetPlatform_PLATFORM_LINUX)
			seen[pb.TargetPlatform_PLATFORM_LINUX] = true
		}
	}

	// Check for macOS (native platform - check if Unity.app exists)
	if runtime.GOOS == "darwin" {
		if _, err := os.Stat(editorPath); err == nil {
			if !seen[pb.TargetPlatform_PLATFORM_MACOS] {
				targets = append(targets, pb.TargetPlatform_PLATFORM_MACOS)
				seen[pb.TargetPlatform_PLATFORM_MACOS] = true
			}
		}
	}

	// Check for WebGL
	webglDir := filepath.Join(editorDir, "WebGLSupport")
	if _, err := os.Stat(webglDir); err == nil {
		if !seen[pb.TargetPlatform_PLATFORM_WEBGL] {
			targets = append(targets, pb.TargetPlatform_PLATFORM_WEBGL)
			seen[pb.TargetPlatform_PLATFORM_WEBGL] = true
		}
	}

	return targets
}
