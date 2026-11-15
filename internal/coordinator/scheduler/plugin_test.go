package scheduler

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/registry"
)

// fakePlugin is an in-process scheduler plugin used as the external
// endpoint in tests.
type fakePlugin struct {
	pb.UnimplementedSchedulerPluginServer
	selectFn func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error)
	feedFn   func(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error)
}

func (f *fakePlugin) Select(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
	return f.selectFn(ctx, req)
}

func (f *fakePlugin) Feedback(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error) {
	return f.feedFn(ctx, req)
}

// startPlugin launches a fake plugin on a random port and returns its
// address plus a stop function.
func startPlugin(t *testing.T, p *fakePlugin) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	gs := grpc.NewServer()
	pb.RegisterSchedulerPluginServer(gs, p)
	go gs.Serve(lis)
	return lis.Addr().String(), gs.Stop
}

func TestPluginScheduler_SelectUsesPlugin(t *testing.T) {
	// Registry with two workers; the plugin will pick worker-b.
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-b", Address: "127.0.0.1:2",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 2,
	}); err != nil {
		t.Fatal(err)
	}

	var sawWorkers int32Shim
	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			sawWorkers.n = len(req.GetWorkers())
			return &pb.SelectResponse{WorkerId: "worker-b"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w.ID != "worker-b" {
		t.Fatalf("expected worker-b (plugin's answer), got %s", w.ID)
	}
	if sawWorkers.n != 2 {
		t.Fatalf("plugin should see 2 workers, saw %d", sawWorkers.n)
	}
}

func TestPluginScheduler_FallsBackWhenUnreachable(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	// Port 1 is never listening; the constructor must still succeed
	// fast (dial timeout) — actually it fails to dial: assert the
	// documented behavior.
	_, err := NewPluginScheduler(PluginConfig{
		Address: "127.0.0.1:1", Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err == nil {
		t.Fatal("expected dial error for unreachable plugin")
	}
}

func TestPluginScheduler_FallsBackOnTimeout(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			time.Sleep(300 * time.Millisecond) // far beyond the timeout
			return &pb.SelectResponse{WorkerId: "worker-a"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Timeout: 30 * time.Millisecond,
		Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	start := time.Now()
	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w == nil || w.ID != "worker-a" {
		t.Fatalf("fallback should pick worker-a, got %+v", w)
	}
	if elapsed := time.Since(start); elapsed > 250*time.Millisecond {
		t.Fatalf("timeout fallback took too long: %v", elapsed)
	}
}

func TestPluginScheduler_FallsBackOnUnknownWorker(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-a", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4,
	}); err != nil {
		t.Fatal(err)
	}

	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{WorkerId: "ghost-worker"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w == nil || w.ID != "worker-a" {
		t.Fatalf("unknown-worker answer must fall back to worker-a, got %+v", w)
	}
}

func TestPluginScheduler_RevalidatesArch(t *testing.T) {
	// worker-x86 is idle (the plugin's greedy pick), worker-arm is busy —
	// but worker-x86 cannot serve an arm64 task, so the coordinator must
	// reject the plugin's answer and fall back to the inner scheduler.
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-x86", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 0,
	}); err != nil {
		t.Fatal(err)
	}
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-arm", Address: "127.0.0.1:2",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_ARM64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 3,
	}); err != nil {
		t.Fatal(err)
	}

	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{WorkerId: "worker-x86"}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_ARM64, "")
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if w.ID != "worker-arm" {
		t.Fatalf("plugin answered worker-x86 for an arm64 task; expected fallback to worker-arm, got %s", w.ID)
	}
}

func TestPluginScheduler_FeedbackNeverBlocks(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	release := make(chan struct{})
	addr, stop := startPlugin(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{}, nil
		},
		feedFn: func(ctx context.Context, req *pb.FeedbackRequest) (*pb.FeedbackResponse, error) {
			<-release // block every feedback call
			return &pb.FeedbackResponse{}, nil
		},
	})
	defer stop()

	s, err := NewPluginScheduler(PluginConfig{
		Address: addr, Inner: NewLeastLoadedScheduler(reg), Registry: reg,
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler: %v", err)
	}
	defer s.Close()

	// RecordOutcome must return promptly even with the plugin stuck.
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < feedbackBuffer*2; i++ {
			s.RecordOutcome("worker-a", -10, true, TaskContext{})
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("RecordOutcome blocked on a stuck plugin")
	}
	close(release)
}

// int32Shim is a tiny mutable int for capturing plugin observations.
type int32Shim struct{ n int }

// newCABundle mints a throwaway CA plus one server and one client
// certificate for mTLS tests.
type caBundle struct {
	caDER      []byte
	caCert     *x509.Certificate
	serverCert tls.Certificate
	clientCert tls.Certificate
}

func newCABundle(t *testing.T) *caBundle {
	t.Helper()
	caKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	caCert, err := x509.ParseCertificate(caDER)
	if err != nil {
		t.Fatal(err)
	}
	mint := func(sn int64, cn string, dnsSANs []string) tls.Certificate {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		tmpl := &x509.Certificate{
			SerialNumber: big.NewInt(sn),
			Subject:      pkix.Name{CommonName: cn},
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     time.Now().Add(time.Hour),
			KeyUsage:     x509.KeyUsageDigitalSignature,
			ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
			DNSNames:     dnsSANs,
		}
		der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
		if err != nil {
			t.Fatal(err)
		}
		return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
	}
	return &caBundle{
		caDER:      caDER,
		caCert:     caCert,
		serverCert: mint(2, "plugin", []string{"localhost"}),
		clientCert: mint(3, "coordinator", nil),
	}
}

// startPluginTLS serves fakePlugin behind RequireAndVerifyClientCert.
func startPluginTLS(t *testing.T, p *fakePlugin, b *caBundle) (string, func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatal(err)
	}
	caPool := x509.NewCertPool()
	caPool.AddCert(b.caCert)
	gs := grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{b.serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    caPool,
		MinVersion:   tls.VersionTLS12,
	})))
	pb.RegisterSchedulerPluginServer(gs, p)
	go gs.Serve(lis)
	return lis.Addr().String(), gs.Stop
}

func TestPluginScheduler_MTLS(t *testing.T) {
	reg := registry.NewInMemoryRegistry(30 * time.Second)
	if err := reg.Add(&registry.WorkerInfo{
		ID: "worker-tls", Address: "127.0.0.1:1",
		Capabilities: &pb.WorkerCapabilities{NativeArch: pb.Architecture_ARCH_X86_64, Cpp: &pb.CppCapability{Compilers: []string{"gcc"}}},
		MaxParallel:  4, ActiveTasks: 0,
	}); err != nil {
		t.Fatal(err)
	}
	inner := NewLeastLoadedScheduler(reg)
	b := newCABundle(t)
	addr, stop := startPluginTLS(t, &fakePlugin{
		selectFn: func(ctx context.Context, req *pb.SelectRequest) (*pb.SelectResponse, error) {
			return &pb.SelectResponse{WorkerId: "worker-tls"}, nil
		},
	}, b)
	defer stop()

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: b.caDER})) {
		t.Fatal("failed to parse test CA")
	}

	s, err := NewPluginScheduler(PluginConfig{
		Address:   addr,
		Inner:     inner,
		Registry:  reg,
		TLSConfig: &tls.Config{Certificates: []tls.Certificate{b.clientCert}, RootCAs: caPool, ServerName: "localhost", MinVersion: tls.VersionTLS12},
	})
	if err != nil {
		t.Fatalf("NewPluginScheduler over mTLS: %v", err)
	}
	defer s.Close()

	w, err := s.Select(pb.BuildType_BUILD_TYPE_CPP, pb.Architecture_ARCH_X86_64, "")
	if err != nil {
		t.Fatalf("Select over mTLS: %v", err)
	}
	if w.ID != "worker-tls" {
		t.Fatalf("expected plugin answer worker-tls over mTLS, got %s", w.ID)
	}
}
