package capability

import (
	"fmt"
	"os"
	"os/exec"
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

	return caps
}

func detectArch() pb.Architecture {
	switch runtime.GOARCH {
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
	switch runtime.GOOS {
	case "linux":
		return detectMemoryLinux()
	case "darwin":
		return detectMemoryDarwin()
	case "windows":
		return detectMemoryWindows()
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
	cap := &pb.CppCapability{
		Compilers: make([]string, 0),
	}

	compilers := []string{"gcc", "g++", "clang", "clang++"}
	for _, c := range compilers {
		if _, err := exec.LookPath(c); err == nil {
			cap.Compilers = append(cap.Compilers, c)
		}
	}

	// Check for cross-compile toolchains
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
		if _, err := exec.LookPath(c); err == nil {
			cap.CrossCompile = true
			break
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

	// Check for Android SDK
	if os.Getenv("ANDROID_HOME") != "" || os.Getenv("ANDROID_SDK_ROOT") != "" {
		cap.AndroidSdk = true
		cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_ANDROID)
	}

	// Check for Xcode (macOS)
	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("xcodebuild"); err == nil {
			cap.XcodeAvailable = true
			cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_IOS)
			cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_MACOS)
		}
	}

	// Web is always available
	cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_WEB)

	// Linux desktop
	if runtime.GOOS == "linux" {
		cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_LINUX)
	}

	// Windows desktop
	if runtime.GOOS == "windows" {
		cap.Platforms = append(cap.Platforms, pb.TargetPlatform_PLATFORM_WINDOWS)
	}

	return cap
}
