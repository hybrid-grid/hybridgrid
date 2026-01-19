//go:build !windows

package executor

import (
	"context"
	"fmt"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// MSVCExecutor is a stub for non-Windows platforms.
type MSVCExecutor struct{}

// NewMSVCExecutor returns an error on non-Windows platforms.
func NewMSVCExecutor() (*MSVCExecutor, error) {
	return nil, fmt.Errorf("MSVC executor only available on Windows")
}

// Name returns the executor name.
func (e *MSVCExecutor) Name() string {
	return "msvc-stub"
}

// CanExecute always returns false on non-Windows platforms.
func (e *MSVCExecutor) CanExecute(targetArch pb.Architecture, nativeArch pb.Architecture) bool {
	return false
}

// Execute is not supported on non-Windows platforms.
func (e *MSVCExecutor) Execute(ctx context.Context, req *Request) (*Result, error) {
	return nil, fmt.Errorf("MSVC executor not available on this platform")
}
