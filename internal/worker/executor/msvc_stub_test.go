//go:build !windows

package executor

import (
	"context"
	"testing"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

func TestMSVCStub_Name(t *testing.T) {
	e := &MSVCExecutor{}
	if got := e.Name(); got != "msvc-stub" {
		t.Fatalf("Name() = %q, want %q", got, "msvc-stub")
	}
}

func TestMSVCStub_CanExecute(t *testing.T) {
	e := &MSVCExecutor{}
	if e.CanExecute(pb.Architecture_ARCH_X86_64, pb.Architecture_ARCH_X86_64) {
		t.Fatal("CanExecute() = true, want false")
	}
}

func TestMSVCStub_Execute(t *testing.T) {
	e := &MSVCExecutor{}
	result, err := e.Execute(context.Background(), &Request{TaskID: "msvc-stub"})
	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if result != nil {
		t.Fatalf("Execute() result = %#v, want nil", result)
	}
}
