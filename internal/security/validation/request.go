package validation

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"unicode"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const (
	// MaxSourceArchiveBytes is the maximum allowed source archive size (100MB)
	MaxSourceArchiveBytes = 100 * 1024 * 1024

	// MaxTaskIDLength is the maximum length of task ID
	MaxTaskIDLength = 128

	// MaxCompilerArgsCount is the maximum number of compiler arguments
	MaxCompilerArgsCount = 256

	// MaxTimeoutSeconds is the maximum allowed timeout
	MaxTimeoutSeconds = 3600 // 1 hour
)

var (
	// taskIDRegex validates task IDs (alphanumeric, dash, underscore)
	taskIDRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
)

// Error represents a validation error.
type Error struct {
	Field   string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// MultiError collects multiple validation errors.
type MultiError struct {
	Errors []*Error
}

func (m *MultiError) Error() string {
	if len(m.Errors) == 0 {
		return "no errors"
	}
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	return fmt.Sprintf("%s (and %d more errors)", m.Errors[0].Error(), len(m.Errors)-1)
}

func (m *MultiError) Add(field, message string) {
	m.Errors = append(m.Errors, &Error{Field: field, Message: message})
}

func (m *MultiError) HasErrors() bool {
	return len(m.Errors) > 0
}

func (m *MultiError) ToError() error {
	if !m.HasErrors() {
		return nil
	}
	return m
}

// ValidateBuildRequest validates a BuildRequest message.
func ValidateBuildRequest(req *pb.BuildRequest) error {
	errs := &MultiError{}

	// Required fields
	if req.TaskId == "" {
		errs.Add("task_id", "required")
	} else {
		if len(req.TaskId) > MaxTaskIDLength {
			errs.Add("task_id", fmt.Sprintf("must be <= %d characters", MaxTaskIDLength))
		}
		if !taskIDRegex.MatchString(req.TaskId) {
			errs.Add("task_id", "must contain only alphanumeric, dash, or underscore")
		}
	}

	if req.SourceHash == "" {
		errs.Add("source_hash", "required")
	} else if !isHexString(req.SourceHash) {
		errs.Add("source_hash", "must be a hex string")
	}

	// Build type
	if req.BuildType == pb.BuildType_BUILD_TYPE_UNSPECIFIED {
		errs.Add("build_type", "required")
	}

	// Target platform
	if req.TargetPlatform == pb.TargetPlatform_PLATFORM_UNSPECIFIED {
		errs.Add("target_platform", "required")
	}

	// Source archive size
	if len(req.SourceArchive) > MaxSourceArchiveBytes {
		errs.Add("source_archive", fmt.Sprintf("must be <= %d bytes", MaxSourceArchiveBytes))
	}

	// Timeout
	if req.TimeoutSeconds < 0 {
		errs.Add("timeout_seconds", "must be >= 0")
	}
	if req.TimeoutSeconds > MaxTimeoutSeconds {
		errs.Add("timeout_seconds", fmt.Sprintf("must be <= %d", MaxTimeoutSeconds))
	}

	// Priority
	if req.Priority < 0 || req.Priority > 100 {
		errs.Add("priority", "must be between 0 and 100")
	}

	// Validate config matches build type
	if err := validateBuildConfig(req); err != nil {
		errs.Add("config", err.Error())
	}

	return errs.ToError()
}

// ValidateCompileRequest validates a legacy CompileRequest message.
func ValidateCompileRequest(req *pb.CompileRequest) error {
	errs := &MultiError{}

	// Task ID
	if req.TaskId == "" {
		errs.Add("task_id", "required")
	} else {
		if len(req.TaskId) > MaxTaskIDLength {
			errs.Add("task_id", fmt.Sprintf("must be <= %d characters", MaxTaskIDLength))
		}
		if !taskIDRegex.MatchString(req.TaskId) {
			errs.Add("task_id", "must contain only alphanumeric, dash, or underscore")
		}
	}

	// Source hash
	if req.SourceHash == "" {
		errs.Add("source_hash", "required")
	} else if !isHexString(req.SourceHash) {
		errs.Add("source_hash", "must be a hex string")
	}

	// Compiler
	if req.Compiler == "" {
		errs.Add("compiler", "required")
	} else if !isValidCompiler(req.Compiler) {
		errs.Add("compiler", "invalid compiler name")
	}

	// Preprocessed source
	if len(req.PreprocessedSource) == 0 {
		errs.Add("preprocessed_source", "required")
	}

	// Compiler args
	if len(req.CompilerArgs) > MaxCompilerArgsCount {
		errs.Add("compiler_args", fmt.Sprintf("must have <= %d arguments", MaxCompilerArgsCount))
	}

	// Timeout
	if req.TimeoutSeconds < 0 {
		errs.Add("timeout_seconds", "must be >= 0")
	}
	if req.TimeoutSeconds > MaxTimeoutSeconds {
		errs.Add("timeout_seconds", fmt.Sprintf("must be <= %d", MaxTimeoutSeconds))
	}

	return errs.ToError()
}

// ValidateHandshakeRequest validates a HandshakeRequest message.
func ValidateHandshakeRequest(req *pb.HandshakeRequest) error {
	errs := &MultiError{}

	if req.Capabilities == nil {
		errs.Add("capabilities", "required")
		return errs.ToError()
	}

	caps := req.Capabilities

	// Worker ID
	if caps.WorkerId == "" {
		errs.Add("capabilities.worker_id", "required")
	} else {
		if len(caps.WorkerId) > MaxTaskIDLength {
			errs.Add("capabilities.worker_id", fmt.Sprintf("must be <= %d characters", MaxTaskIDLength))
		}
		if !taskIDRegex.MatchString(caps.WorkerId) {
			errs.Add("capabilities.worker_id", "must contain only alphanumeric, dash, or underscore")
		}
	}

	// CPU cores
	if caps.CpuCores <= 0 {
		errs.Add("capabilities.cpu_cores", "must be > 0")
	}

	// Memory
	if caps.MemoryBytes <= 0 {
		errs.Add("capabilities.memory_bytes", "must be > 0")
	}

	// Architecture
	if caps.NativeArch == pb.Architecture_ARCH_UNSPECIFIED {
		errs.Add("capabilities.native_arch", "required")
	}

	// Worker address
	if req.WorkerAddress == "" {
		errs.Add("worker_address", "required")
	}

	return errs.ToError()
}

func validateBuildConfig(req *pb.BuildRequest) error {
	switch req.BuildType {
	case pb.BuildType_BUILD_TYPE_CPP:
		if req.GetCppConfig() == nil {
			return fmt.Errorf("cpp_config required for BUILD_TYPE_CPP")
		}
	case pb.BuildType_BUILD_TYPE_FLUTTER:
		if req.GetFlutterConfig() == nil {
			return fmt.Errorf("flutter_config required for BUILD_TYPE_FLUTTER")
		}
	case pb.BuildType_BUILD_TYPE_UNITY:
		if req.GetUnityConfig() == nil {
			return fmt.Errorf("unity_config required for BUILD_TYPE_UNITY")
		}
	case pb.BuildType_BUILD_TYPE_COCOS:
		if req.GetCocosConfig() == nil {
			return fmt.Errorf("cocos_config required for BUILD_TYPE_COCOS")
		}
	case pb.BuildType_BUILD_TYPE_RUST:
		if req.GetRustConfig() == nil {
			return fmt.Errorf("rust_config required for BUILD_TYPE_RUST")
		}
	case pb.BuildType_BUILD_TYPE_GO:
		if req.GetGoConfig() == nil {
			return fmt.Errorf("go_config required for BUILD_TYPE_GO")
		}
	case pb.BuildType_BUILD_TYPE_NODEJS:
		if req.GetNodeConfig() == nil {
			return fmt.Errorf("node_config required for BUILD_TYPE_NODEJS")
		}
	}
	return nil
}

func isHexString(s string) bool {
	if len(s) == 0 || len(s)%2 != 0 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

func isValidCompiler(name string) bool {
	// Allow common compiler names
	valid := map[string]bool{
		"gcc":     true,
		"g++":     true,
		"clang":   true,
		"clang++": true,
		"cc":      true,
		"c++":     true,
	}
	if valid[name] {
		return true
	}
	// Also allow paths like /usr/bin/gcc
	for _, char := range name {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '/' && char != '-' && char != '_' && char != '+' && char != '.' {
			return false
		}
	}
	return true
}
