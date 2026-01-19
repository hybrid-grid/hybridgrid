package capability

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// MSVCInfo holds information about an MSVC installation.
type MSVCInfo struct {
	Available     bool     `json:"available"`
	Version       string   `json:"version"`       // e.g., "2022", "2019"
	InstallDir    string   `json:"install_dir"`   // e.g., "C:\Program Files\Microsoft Visual Studio\2022\Community"
	VCToolsPath   string   `json:"vc_tools_path"` // e.g., "...\VC\Tools\MSVC\14.39.33519"
	Architectures []string `json:"architectures"` // e.g., ["x64", "x86", "arm64"]
	ClExePath     string   `json:"cl_exe_path"`   // Full path to cl.exe
	WindowsSDK    string   `json:"windows_sdk"`   // e.g., "10.0.22621.0"
	ProductID     string   `json:"product_id"`    // e.g., "Microsoft.VisualStudio.Product.Community"
}

// vswhereFindResult represents a single VS installation from vswhere JSON output.
type vswhereFindResult struct {
	InstallationPath string `json:"installationPath"`
	InstallationVer  string `json:"installationVersion"`
	ProductID        string `json:"productId"`
	DisplayName      string `json:"displayName"`
}

// DetectMSVC detects MSVC installations on Windows.
// Returns nil if not on Windows or MSVC is not found.
func DetectMSVC() *MSVCInfo {
	if runtime.GOOS != "windows" {
		return nil
	}

	// 1. Try environment variable (fast path for VS Developer Command Prompt)
	if info := detectMSVCFromEnv(); info != nil && info.Available {
		return info
	}

	// 2. Try vswhere (most reliable)
	if info := detectMSVCFromVSWhere(); info != nil && info.Available {
		return info
	}

	// 3. Fall back to known paths
	return detectMSVCFromKnownPaths()
}

// detectMSVCFromEnv detects MSVC from environment variables.
// This works when running from a VS Developer Command Prompt.
func detectMSVCFromEnv() *MSVCInfo {
	vcInstallDir := os.Getenv("VCINSTALLDIR")
	if vcInstallDir == "" {
		return nil
	}

	info := &MSVCInfo{
		Available:  true,
		InstallDir: filepath.Dir(vcInstallDir), // VCINSTALLDIR points to VC folder
	}

	// Get tools path
	vcToolsVersion := os.Getenv("VCToolsVersion")
	if vcToolsVersion != "" {
		info.VCToolsPath = filepath.Join(vcInstallDir, "Tools", "MSVC", vcToolsVersion)
	}

	// Get version from path or env
	if strings.Contains(vcInstallDir, "2022") {
		info.Version = "2022"
	} else if strings.Contains(vcInstallDir, "2019") {
		info.Version = "2019"
	} else if strings.Contains(vcInstallDir, "2017") {
		info.Version = "2017"
	}

	// Get Windows SDK version
	info.WindowsSDK = os.Getenv("WindowsSDKVersion")
	if strings.HasSuffix(info.WindowsSDK, "\\") {
		info.WindowsSDK = strings.TrimSuffix(info.WindowsSDK, "\\")
	}

	// Find cl.exe
	info.ClExePath = findClExe(info.VCToolsPath)

	// Detect architectures
	info.Architectures = detectMSVCArchitectures(info.VCToolsPath)

	return info
}

// detectMSVCFromVSWhere uses vswhere.exe to find VS installations.
func detectMSVCFromVSWhere() *MSVCInfo {
	vswhereExe := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`
	if _, err := os.Stat(vswhereExe); err != nil {
		return nil
	}

	// Run vswhere to get installations as JSON
	cmd := exec.Command(vswhereExe,
		"-latest",
		"-requires", "Microsoft.VisualStudio.Component.VC.Tools.x86.x64",
		"-format", "json",
		"-utf8",
	)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var results []vswhereFindResult
	if err := json.Unmarshal(output, &results); err != nil || len(results) == 0 {
		return nil
	}

	vs := results[0]
	info := &MSVCInfo{
		Available:  true,
		InstallDir: vs.InstallationPath,
		ProductID:  vs.ProductID,
	}

	// Parse version from installation version (e.g., "17.9.34607.119" -> "2022")
	if strings.HasPrefix(vs.InstallationVer, "17.") {
		info.Version = "2022"
	} else if strings.HasPrefix(vs.InstallationVer, "16.") {
		info.Version = "2019"
	} else if strings.HasPrefix(vs.InstallationVer, "15.") {
		info.Version = "2017"
	}

	// Find VC Tools path
	vcPath := filepath.Join(vs.InstallationPath, "VC", "Tools", "MSVC")
	if entries, err := os.ReadDir(vcPath); err == nil && len(entries) > 0 {
		// Use latest version
		info.VCToolsPath = filepath.Join(vcPath, entries[len(entries)-1].Name())
	}

	// Find cl.exe
	info.ClExePath = findClExe(info.VCToolsPath)

	// Detect architectures
	info.Architectures = detectMSVCArchitectures(info.VCToolsPath)

	// Find Windows SDK
	info.WindowsSDK = detectWindowsSDK()

	return info
}

// detectMSVCFromKnownPaths searches common VS installation paths.
func detectMSVCFromKnownPaths() *MSVCInfo {
	knownPaths := []struct {
		path    string
		version string
	}{
		{`C:\Program Files\Microsoft Visual Studio\2022\Enterprise`, "2022"},
		{`C:\Program Files\Microsoft Visual Studio\2022\Professional`, "2022"},
		{`C:\Program Files\Microsoft Visual Studio\2022\Community`, "2022"},
		{`C:\Program Files\Microsoft Visual Studio\2022\BuildTools`, "2022"},
		{`C:\Program Files (x86)\Microsoft Visual Studio\2019\Enterprise`, "2019"},
		{`C:\Program Files (x86)\Microsoft Visual Studio\2019\Professional`, "2019"},
		{`C:\Program Files (x86)\Microsoft Visual Studio\2019\Community`, "2019"},
		{`C:\Program Files (x86)\Microsoft Visual Studio\2019\BuildTools`, "2019"},
	}

	for _, kp := range knownPaths {
		if _, err := os.Stat(kp.path); err == nil {
			// Found VS installation
			info := &MSVCInfo{
				Available:  true,
				InstallDir: kp.path,
				Version:    kp.version,
			}

			// Find VC Tools path
			vcPath := filepath.Join(kp.path, "VC", "Tools", "MSVC")
			if entries, err := os.ReadDir(vcPath); err == nil && len(entries) > 0 {
				info.VCToolsPath = filepath.Join(vcPath, entries[len(entries)-1].Name())
			}

			// Find cl.exe
			info.ClExePath = findClExe(info.VCToolsPath)

			// Detect architectures
			info.Architectures = detectMSVCArchitectures(info.VCToolsPath)

			// Find Windows SDK
			info.WindowsSDK = detectWindowsSDK()

			if info.ClExePath != "" {
				return info
			}
		}
	}

	return nil
}

// findClExe finds the cl.exe compiler for x64.
func findClExe(vcToolsPath string) string {
	if vcToolsPath == "" {
		return ""
	}

	// Prefer x64 hosted x64 target
	clPath := filepath.Join(vcToolsPath, "bin", "Hostx64", "x64", "cl.exe")
	if _, err := os.Stat(clPath); err == nil {
		return clPath
	}

	// Fall back to x86 hosted x64 target
	clPath = filepath.Join(vcToolsPath, "bin", "Hostx86", "x64", "cl.exe")
	if _, err := os.Stat(clPath); err == nil {
		return clPath
	}

	return ""
}

// detectMSVCArchitectures detects available MSVC target architectures.
func detectMSVCArchitectures(vcToolsPath string) []string {
	if vcToolsPath == "" {
		return nil
	}

	var archs []string
	hostArch := "Hostx64"
	if runtime.GOARCH == "386" {
		hostArch = "Hostx86"
	}

	binPath := filepath.Join(vcToolsPath, "bin", hostArch)
	entries, err := os.ReadDir(binPath)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if entry.IsDir() {
			clPath := filepath.Join(binPath, entry.Name(), "cl.exe")
			if _, err := os.Stat(clPath); err == nil {
				archs = append(archs, entry.Name())
			}
		}
	}

	return archs
}

// detectWindowsSDK finds the installed Windows SDK version.
func detectWindowsSDK() string {
	sdkRoot := os.Getenv("WindowsSdkDir")
	if sdkRoot == "" {
		// Try common paths
		sdkRoot = `C:\Program Files (x86)\Windows Kits\10`
	}

	includePath := filepath.Join(sdkRoot, "Include")
	entries, err := os.ReadDir(includePath)
	if err != nil {
		return ""
	}

	// Find latest SDK version
	for i := len(entries) - 1; i >= 0; i-- {
		name := entries[i].Name()
		if strings.HasPrefix(name, "10.0.") && entries[i].IsDir() {
			// Verify ucrt exists
			ucrtPath := filepath.Join(includePath, name, "ucrt")
			if _, err := os.Stat(ucrtPath); err == nil {
				return name
			}
		}
	}

	return ""
}

// GetVCVarsPath returns the path to vcvars64.bat for environment setup.
func (m *MSVCInfo) GetVCVarsPath() string {
	if m == nil || !m.Available {
		return ""
	}

	vcvarsPath := filepath.Join(m.InstallDir, "VC", "Auxiliary", "Build", "vcvars64.bat")
	if _, err := os.Stat(vcvarsPath); err == nil {
		return vcvarsPath
	}

	return ""
}

// GetClExeForArch returns the path to cl.exe for a specific target architecture.
func (m *MSVCInfo) GetClExeForArch(targetArch string) string {
	if m == nil || m.VCToolsPath == "" {
		return ""
	}

	hostArch := "Hostx64"
	if runtime.GOARCH == "386" {
		hostArch = "Hostx86"
	}

	clPath := filepath.Join(m.VCToolsPath, "bin", hostArch, targetArch, "cl.exe")
	if _, err := os.Stat(clPath); err == nil {
		return clPath
	}

	return ""
}
