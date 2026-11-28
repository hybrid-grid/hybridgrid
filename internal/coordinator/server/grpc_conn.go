package server

import (
	"context"
	"sync"
	"time"

	"google.golang.org/grpc"
)

// workerDialTimeout bounds how long connecting to a worker may take.
// The dial blocks until the connection is up, so without its own bound a
// worker that died between heartbeats would hold the build for the whole
// client deadline (tens of minutes for a Flutter build) instead of
// failing over to another worker.
const workerDialTimeout = 5 * time.Second

// connPool caches gRPC client connections to workers by address.
type connPool struct {
	mu          sync.Mutex
	conns       map[string]*grpc.ClientConn
	dialOpts    []grpc.DialOption
	dialTimeout time.Duration
}

func newConnPool(dialOpts []grpc.DialOption) *connPool {
	return &connPool{
		conns:       make(map[string]*grpc.ClientConn),
		dialOpts:    dialOpts,
		dialTimeout: workerDialTimeout,
	}
}

// get returns an existing connection or creates a new one.
func (p *connPool) get(ctx context.Context, addr string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	if conn, ok := p.conns[addr]; ok {
		p.mu.Unlock()
		return conn, nil
	}
	p.mu.Unlock()

	opts := make([]grpc.DialOption, len(p.dialOpts))
	copy(opts, p.dialOpts)
	// A refused connection means nothing listens there: fail now
	// instead of retrying it until the dial timeout.
	opts = append(opts, grpc.WithBlock(), grpc.FailOnNonTempDialError(true))

	dialCtx, cancel := context.WithTimeout(ctx, p.dialTimeout)
	defer cancel()
	conn, err := grpc.DialContext(dialCtx, addr, opts...)
	if err != nil {
		return nil, err
	}

	p.mu.Lock()
	defer p.mu.Unlock()
	if existing, ok := p.conns[addr]; ok {
		_ = conn.Close()
		return existing, nil
	}
	p.conns[addr] = conn
	return conn, nil
}

// closeAll closes all pooled connections.
func (p *connPool) closeAll() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for addr, conn := range p.conns {
		conn.Close()
		delete(p.conns, addr)
	}
}
