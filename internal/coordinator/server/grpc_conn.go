package server

import (
	"context"
	"sync"

	"google.golang.org/grpc"
)

// connPool caches gRPC client connections to workers by address.
type connPool struct {
	mu       sync.Mutex
	conns    map[string]*grpc.ClientConn
	dialOpts []grpc.DialOption
}

func newConnPool(dialOpts []grpc.DialOption) *connPool {
	return &connPool{
		conns:    make(map[string]*grpc.ClientConn),
		dialOpts: dialOpts,
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
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.DialContext(ctx, addr, opts...)
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
