package client

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

const (
	defaultTimeout   = 30 * time.Second
	defaultChunkSize = 64 * 1024 // 64KB chunks for streaming
)

// Config holds the gRPC client configuration.
type Config struct {
	Address   string
	AuthToken string
	Timeout   time.Duration
	Insecure  bool
}

// Client wraps the BuildService gRPC client.
type Client struct {
	config Config
	conn   *grpc.ClientConn
	client pb.BuildServiceClient
}

// New creates a new gRPC client connected to the specified address.
func New(cfg Config) (*Client, error) {
	if cfg.Timeout == 0 {
		cfg.Timeout = defaultTimeout
	}

	opts := []grpc.DialOption{}
	if cfg.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.NewClient(cfg.Address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &Client{
		config: cfg,
		conn:   conn,
		client: pb.NewBuildServiceClient(conn),
	}, nil
}

// Close closes the gRPC connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Handshake registers a worker with the coordinator using a full request.
func (c *Client) Handshake(ctx context.Context, req *pb.HandshakeRequest) (*pb.HandshakeResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	// Use configured auth token if not provided in request
	if req.AuthToken == "" {
		req.AuthToken = c.config.AuthToken
	}

	return c.client.Handshake(ctx, req)
}

// HandshakeWithCaps is a convenience method for simple handshakes.
func (c *Client) HandshakeWithCaps(ctx context.Context, caps *pb.WorkerCapabilities) (*pb.HandshakeResponse, error) {
	return c.Handshake(ctx, &pb.HandshakeRequest{
		Capabilities: caps,
		AuthToken:    c.config.AuthToken,
	})
}

// Build sends a build request to the coordinator.
func (c *Client) Build(ctx context.Context, req *pb.BuildRequest) (*pb.BuildResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	return c.client.Build(ctx, req)
}

// StreamBuild sends a streaming build request for large projects.
func (c *Client) StreamBuild(ctx context.Context, metadata *pb.BuildMetadata, source io.Reader) (*pb.BuildResponse, error) {
	stream, err := c.client.StreamBuild(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to open stream: %w", err)
	}

	// Send metadata first
	if err := stream.Send(&pb.BuildChunk{
		Payload: &pb.BuildChunk_Metadata{Metadata: metadata},
	}); err != nil {
		return nil, fmt.Errorf("failed to send metadata: %w", err)
	}

	// Stream source data in chunks
	buf := make([]byte, defaultChunkSize)
	for {
		n, err := source.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read source: %w", err)
		}

		if err := stream.Send(&pb.BuildChunk{
			Payload: &pb.BuildChunk_SourceChunk{SourceChunk: buf[:n]},
		}); err != nil {
			return nil, fmt.Errorf("failed to send chunk: %w", err)
		}
	}

	// Close and receive response
	return stream.CloseAndRecv()
}

// Compile sends a legacy C/C++ compilation request.
func (c *Client) Compile(ctx context.Context, req *pb.CompileRequest) (*pb.CompileResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	return c.client.Compile(ctx, req)
}

// HealthCheck checks the health of the coordinator.
func (c *Client) HealthCheck(ctx context.Context) (*pb.HealthResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	return c.client.HealthCheck(ctx, &pb.HealthRequest{})
}

// GetWorkerStatus retrieves the status of all workers.
func (c *Client) GetWorkerStatus(ctx context.Context) (*pb.WorkerStatusResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	return c.client.GetWorkerStatus(ctx, &pb.WorkerStatusRequest{})
}

// GetWorkersForBuild retrieves workers capable of handling a specific build.
func (c *Client) GetWorkersForBuild(ctx context.Context, buildType pb.BuildType, platform pb.TargetPlatform) (*pb.WorkersForBuildResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.config.Timeout)
	defer cancel()

	return c.client.GetWorkersForBuild(ctx, &pb.WorkersForBuildRequest{
		BuildType:      buildType,
		TargetPlatform: platform,
	})
}
