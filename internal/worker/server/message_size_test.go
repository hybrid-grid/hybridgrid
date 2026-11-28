package server

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "github.com/h3nr1-d14z/hybridgrid/gen/go/hybridgrid/v1"
)

// TestServer_AcceptsLargeSourceArchives: a build request carries the
// whole source archive, so the server must accept more than gRPC's
// 4 MiB default. The request reaches the handler and is rejected there
// (no task id), not by the transport with ResourceExhausted.
func TestServer_AcceptsLargeSourceArchives(t *testing.T) {
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer(messageSizeOptions()...)
	pb.RegisterBuildServiceServer(srv, New(DefaultConfig()))
	go func() { _ = srv.Serve(lis) }()
	defer srv.Stop()

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallSendMsgSize(maxGRPCMessageSize), grpc.MaxCallRecvMsgSize(maxGRPCMessageSize)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err = pb.NewBuildServiceClient(conn).Build(ctx, &pb.BuildRequest{SourceArchive: make([]byte, 6<<20)})
	if status.Code(err) == codes.ResourceExhausted {
		t.Fatalf("6 MiB request rejected by the transport: %v", err)
	}
}
