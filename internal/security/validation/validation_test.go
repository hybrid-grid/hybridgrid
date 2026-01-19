package validation

import (
	"runtime"
	"strings"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestValidateBuildRequest_Valid(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "build-123",
		SourceHash:     "abc123def456",
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
		Config:         &pb.BuildRequest_CppConfig{CppConfig: &pb.CppConfig{Compiler: "gcc"}},
	}

	err := ValidateBuildRequest(req)
	if err != nil {
		t.Errorf("ValidateBuildRequest failed for valid request: %v", err)
	}
}

func TestValidateBuildRequest_MissingTaskID(t *testing.T) {
	req := &pb.BuildRequest{
		SourceHash:     "abc123def456",
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
		Config:         &pb.BuildRequest_CppConfig{CppConfig: &pb.CppConfig{Compiler: "gcc"}},
	}

	err := ValidateBuildRequest(req)
	if err == nil {
		t.Error("Expected error for missing task_id")
	}
	if !strings.Contains(err.Error(), "task_id") {
		t.Errorf("Error should mention task_id: %v", err)
	}
}

func TestValidateBuildRequest_InvalidTaskID(t *testing.T) {
	tests := []struct {
		name   string
		taskID string
	}{
		{"with space", "task 123"},
		{"with special chars", "task@123"},
		{"with semicolon", "task;123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &pb.BuildRequest{
				TaskId:         tt.taskID,
				SourceHash:     "abc123def456",
				BuildType:      pb.BuildType_BUILD_TYPE_CPP,
				TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
				Config:         &pb.BuildRequest_CppConfig{CppConfig: &pb.CppConfig{}},
			}

			err := ValidateBuildRequest(req)
			if err == nil {
				t.Errorf("Expected error for task_id %q", tt.taskID)
			}
		})
	}
}

func TestValidateBuildRequest_InvalidSourceHash(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "build-123",
		SourceHash:     "not-a-hex-string!",
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
		Config:         &pb.BuildRequest_CppConfig{CppConfig: &pb.CppConfig{}},
	}

	err := ValidateBuildRequest(req)
	if err == nil {
		t.Error("Expected error for invalid source_hash")
	}
	if !strings.Contains(err.Error(), "source_hash") {
		t.Errorf("Error should mention source_hash: %v", err)
	}
}

func TestValidateBuildRequest_MissingBuildType(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "build-123",
		SourceHash:     "abc123def456",
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
	}

	err := ValidateBuildRequest(req)
	if err == nil {
		t.Error("Expected error for missing build_type")
	}
}

func TestValidateBuildRequest_MismatchedConfig(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "build-123",
		SourceHash:     "abc123def456",
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
		// Missing CppConfig
	}

	err := ValidateBuildRequest(req)
	if err == nil {
		t.Error("Expected error for mismatched config")
	}
	if !strings.Contains(err.Error(), "config") {
		t.Errorf("Error should mention config: %v", err)
	}
}

func TestValidateBuildRequest_InvalidPriority(t *testing.T) {
	req := &pb.BuildRequest{
		TaskId:         "build-123",
		SourceHash:     "abc123def456",
		BuildType:      pb.BuildType_BUILD_TYPE_CPP,
		TargetPlatform: pb.TargetPlatform_PLATFORM_LINUX,
		Config:         &pb.BuildRequest_CppConfig{CppConfig: &pb.CppConfig{}},
		Priority:       150, // Invalid: > 100
	}

	err := ValidateBuildRequest(req)
	if err == nil {
		t.Error("Expected error for invalid priority")
	}
}

func TestValidateCompileRequest_Valid(t *testing.T) {
	req := &pb.CompileRequest{
		TaskId:             "compile-123",
		SourceHash:         "abc123def456",
		Compiler:           "gcc",
		PreprocessedSource: []byte("int main() { return 0; }"),
		CompilerArgs:       []string{"-O2", "-Wall"},
	}

	err := ValidateCompileRequest(req)
	if err != nil {
		t.Errorf("ValidateCompileRequest failed for valid request: %v", err)
	}
}

func TestValidateCompileRequest_InvalidCompiler(t *testing.T) {
	req := &pb.CompileRequest{
		TaskId:             "compile-123",
		SourceHash:         "abc123def456",
		Compiler:           "rm -rf /",
		PreprocessedSource: []byte("int main() { return 0; }"),
	}

	err := ValidateCompileRequest(req)
	if err == nil {
		t.Error("Expected error for invalid compiler")
	}
}

func TestValidateHandshakeRequest_Valid(t *testing.T) {
	req := &pb.HandshakeRequest{
		Capabilities: &pb.WorkerCapabilities{
			WorkerId:    "worker-1",
			CpuCores:    4,
			MemoryBytes: 8 * 1024 * 1024 * 1024,
			NativeArch:  pb.Architecture_ARCH_X86_64,
		},
		WorkerAddress: "192.168.1.100:50051",
	}

	err := ValidateHandshakeRequest(req)
	if err != nil {
		t.Errorf("ValidateHandshakeRequest failed for valid request: %v", err)
	}
}

func TestValidateHandshakeRequest_MissingCapabilities(t *testing.T) {
	req := &pb.HandshakeRequest{
		WorkerAddress: "192.168.1.100:50051",
	}

	err := ValidateHandshakeRequest(req)
	if err == nil {
		t.Error("Expected error for missing capabilities")
	}
}

func TestSanitizeCompilerArgs_RemovesDangerousFlags(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantLen int
		wantRem int
	}{
		{
			name:    "removes --plugin",
			args:    []string{"-O2", "--plugin", "malicious.so", "-Wall"},
			wantLen: 2, // -O2 and -Wall
			wantRem: 2, // --plugin and malicious.so
		},
		{
			name:    "removes -fplugin=",
			args:    []string{"-O2", "-fplugin=/path/to/plugin.so"},
			wantLen: 1,
			wantRem: 1,
		},
		{
			name:    "removes -B toolchain",
			args:    []string{"-O2", "-B", "/malicious/toolchain"},
			wantLen: 1,
			wantRem: 2,
		},
		{
			name:    "removes shell metacharacters",
			args:    []string{"-O2", "-DFOO=`id`", "-Wall"},
			wantLen: 2,
			wantRem: 1,
		},
		{
			name:    "removes command injection",
			args:    []string{"-O2", "-DBAR=$(whoami)", "-c"},
			wantLen: 2,
			wantRem: 1,
		},
		{
			name:    "keeps safe args",
			args:    []string{"-O2", "-Wall", "-Werror", "-I/usr/include", "-c", "-o", "output.o"},
			wantLen: 7,
			wantRem: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized, removed := SanitizeCompilerArgs(tt.args)
			if len(sanitized) != tt.wantLen {
				t.Errorf("sanitized len = %d, want %d. sanitized: %v", len(sanitized), tt.wantLen, sanitized)
			}
			if len(removed) != tt.wantRem {
				t.Errorf("removed len = %d, want %d. removed: %v", len(removed), tt.wantRem, removed)
			}
		})
	}
}

func TestSanitizeCompilerArgs_RemovesPathTraversal(t *testing.T) {
	var args []string
	var expectedSanitized, expectedRemoved int

	if runtime.GOOS == "windows" {
		// Windows-style path traversal
		args = []string{"-O2", "-I..\\..\\..\\Windows\\System32", "-IC:\\include"}
		expectedSanitized = 2
		expectedRemoved = 1
	} else {
		// Unix-style path traversal
		args = []string{"-O2", "-I../../../etc/passwd", "-I/usr/include"}
		expectedSanitized = 2
		expectedRemoved = 1
	}

	sanitized, removed := SanitizeCompilerArgs(args)

	if len(removed) != expectedRemoved {
		t.Errorf("Expected %d removed arg, got %d: %v", expectedRemoved, len(removed), removed)
	}
	if len(sanitized) != expectedSanitized {
		t.Errorf("Expected %d sanitized args, got %d: %v", expectedSanitized, len(sanitized), sanitized)
	}
}

func TestSanitizePath(t *testing.T) {
	type testCase struct {
		name     string
		basePath string
		path     string
		want     string
	}

	var tests []testCase

	if runtime.GOOS == "windows" {
		tests = []testCase{
			{
				name:     "valid relative path",
				basePath: "C:\\workspace",
				path:     "src\\main.c",
				want:     "C:\\workspace\\src\\main.c",
			},
			{
				name:     "blocks path traversal",
				basePath: "C:\\workspace",
				path:     "..\\..\\..\\Windows\\System32",
				want:     "",
			},
			{
				name:     "blocks absolute escape",
				basePath: "C:\\workspace",
				path:     "C:\\Windows\\System32",
				want:     "",
			},
			{
				name:     "allows subpath of base",
				basePath: "C:\\workspace",
				path:     "C:\\workspace\\src\\main.c",
				want:     "C:\\workspace\\src\\main.c",
			},
			{
				name:     "blocks reserved names",
				basePath: "C:\\workspace",
				path:     "CON",
				want:     "",
			},
			{
				name:     "blocks reserved names with extension",
				basePath: "C:\\workspace",
				path:     "NUL.txt",
				want:     "",
			},
			{
				name:     "empty path",
				basePath: "C:\\workspace",
				path:     "",
				want:     "",
			},
		}
	} else {
		tests = []testCase{
			{
				name:     "valid relative path",
				basePath: "/workspace",
				path:     "src/main.c",
				want:     "/workspace/src/main.c",
			},
			{
				name:     "blocks path traversal",
				basePath: "/workspace",
				path:     "../../../etc/passwd",
				want:     "",
			},
			{
				name:     "blocks absolute escape",
				basePath: "/workspace",
				path:     "/etc/passwd",
				want:     "",
			},
			{
				name:     "allows subpath of base",
				basePath: "/workspace",
				path:     "/workspace/src/main.c",
				want:     "/workspace/src/main.c",
			},
			{
				name:     "empty path",
				basePath: "/workspace",
				path:     "",
				want:     "",
			},
		}
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizePath(tt.basePath, tt.path)
			if got != tt.want {
				t.Errorf("SanitizePath(%q, %q) = %q, want %q", tt.basePath, tt.path, got, tt.want)
			}
		})
	}
}

func TestWindowsPathValidation(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		valid bool
	}{
		{"valid path", "foo/bar.txt", true},
		{"reserved name CON", "CON", false},
		{"reserved name PRN", "PRN", false},
		{"reserved name with ext", "NUL.txt", false},
		{"reserved name COM1", "COM1", false},
		{"invalid char <", "foo<bar", false},
		{"invalid char >", "foo>bar", false},
		{"invalid char : in filename", "foo:bar", false},
		{"invalid char |", "foo|bar", false},
		{"invalid char ?", "foo?bar", false},
		{"invalid char *", "foo*bar", false},
		{"valid with numbers", "abc123", true},
		{"valid drive letter C:", "C:\\folder\\file.txt", true},
		{"valid drive letter D:", "D:\\test", true},
		{"invalid colon after drive", "C:\\foo:bar", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := ValidatePathForWindows(tt.path)
			isValid := errMsg == ""
			if isValid != tt.valid {
				t.Errorf("ValidatePathForWindows(%q) = %q, want valid=%v", tt.path, errMsg, tt.valid)
			}
		})
	}
}

func TestValidateDockerImage(t *testing.T) {
	tests := []struct {
		name  string
		image string
		valid bool
	}{
		{"empty is valid", "", true},
		{"simple name", "ubuntu", true},
		{"with tag", "ubuntu:20.04", true},
		{"with registry", "docker.io/library/ubuntu:20.04", true},
		{"with digest", "ubuntu@sha256:abc123", true},
		{"shell injection", "ubuntu;rm -rf /", false},
		{"command substitution", "$(whoami)/image", false},
		{"pipe", "image|cat", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateDockerImage(tt.image)
			if got != tt.valid {
				t.Errorf("ValidateDockerImage(%q) = %v, want %v", tt.image, got, tt.valid)
			}
		})
	}
}

func TestIsHexString(t *testing.T) {
	tests := []struct {
		s    string
		want bool
	}{
		{"abc123", true},
		{"ABC123", true},
		{"abc123def456", true},
		{"", false},
		{"abc", false}, // Odd length
		{"ghijkl", false},
		{"abc 123", false},
	}

	for _, tt := range tests {
		got := isHexString(tt.s)
		if got != tt.want {
			t.Errorf("isHexString(%q) = %v, want %v", tt.s, got, tt.want)
		}
	}
}

func TestMultiError(t *testing.T) {
	errs := &MultiError{}

	if errs.HasErrors() {
		t.Error("Empty MultiError should not have errors")
	}
	if errs.ToError() != nil {
		t.Error("Empty MultiError.ToError() should return nil")
	}

	errs.Add("field1", "error1")
	if !errs.HasErrors() {
		t.Error("MultiError with errors should report HasErrors")
	}
	if errs.ToError() == nil {
		t.Error("MultiError.ToError() should return error")
	}
	if !strings.Contains(errs.Error(), "field1") {
		t.Error("Error should contain field name")
	}

	errs.Add("field2", "error2")
	if !strings.Contains(errs.Error(), "and 1 more") {
		t.Errorf("Error should mention additional errors: %v", errs.Error())
	}
}

func TestIsValidCompiler(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"gcc", true},
		{"g++", true},
		{"clang", true},
		{"clang++", true},
		{"/usr/bin/gcc", true},
		{"/usr/local/bin/clang-12", true},
		{"rm -rf /", false},
		{"gcc; whoami", false},
		{"gcc`id`", false},
	}

	for _, tt := range tests {
		got := isValidCompiler(tt.name)
		if got != tt.valid {
			t.Errorf("isValidCompiler(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}
