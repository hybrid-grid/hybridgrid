package interceptors

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryRequestIDInterceptor_PreservesID(t *testing.T) {
	interceptor := UnaryRequestIDInterceptor()
	expectedID := "test-request-id-12345"

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(requestIDMetadataKey, expectedID),
	)

	called := false
	var capturedCtx context.Context

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}

	retrievedID := RequestIDFromContext(capturedCtx)
	if retrievedID != expectedID {
		t.Errorf("expected request ID %q, got %q", expectedID, retrievedID)
	}
}

func TestUnaryRequestIDInterceptor_GeneratesID(t *testing.T) {
	interceptor := UnaryRequestIDInterceptor()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())

	called := false
	var capturedCtx context.Context

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		called = true
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}

	retrievedID := RequestIDFromContext(capturedCtx)
	if retrievedID == "" {
		t.Fatal("request ID was empty")
	}

	if len(retrievedID) < 32 {
		t.Errorf("expected request ID length >= 32, got %d", len(retrievedID))
	}

	for _, c := range retrievedID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Errorf("request ID contains non-hex character: %c", c)
		}
	}
}

func TestUnaryRequestIDInterceptor_UniquenessOf100IDs(t *testing.T) {
	interceptor := UnaryRequestIDInterceptor()
	idSet := make(map[string]bool)

	for i := 0; i < 100; i++ {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())

		var capturedCtx context.Context
		handler := func(ctx context.Context, req interface{}) (interface{}, error) {
			capturedCtx = ctx
			return nil, nil
		}

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test"}, handler)
		if err != nil {
			t.Fatalf("iteration %d: interceptor returned error: %v", i, err)
		}

		id := RequestIDFromContext(capturedCtx)
		if id == "" {
			t.Fatalf("iteration %d: request ID was empty", i)
		}

		if idSet[id] {
			t.Fatalf("iteration %d: duplicate ID generated: %s", i, id)
		}
		idSet[id] = true
	}

	if len(idSet) != 100 {
		t.Errorf("expected 100 unique IDs, got %d", len(idSet))
	}
}

func TestRequestIDFromContext_ReturnsEmptyStringWhenMissing(t *testing.T) {
	ctx := context.Background()
	id := RequestIDFromContext(ctx)
	if id != "" {
		t.Errorf("expected empty string, got %q", id)
	}
}

func TestRequestIDFromContext_ReturnsCorrectValue(t *testing.T) {
	expectedID := "my-request-id"
	ctx := context.WithValue(context.Background(), contextKey, expectedID)
	id := RequestIDFromContext(ctx)
	if id != expectedID {
		t.Errorf("expected %q, got %q", expectedID, id)
	}
}

func TestStreamRequestIDInterceptor_PreservesID(t *testing.T) {
	interceptor := StreamRequestIDInterceptor()
	expectedID := "stream-request-id-12345"

	ctx := metadata.NewIncomingContext(
		context.Background(),
		metadata.Pairs(requestIDMetadataKey, expectedID),
	)

	mockStream := &mockServerStream{ctx: ctx}

	called := false
	handler := func(srv interface{}, ss grpc.ServerStream) error {
		called = true
		retrievedID := RequestIDFromContext(ss.Context())
		if retrievedID != expectedID {
			t.Errorf("expected request ID %q, got %q", expectedID, retrievedID)
		}
		return nil
	}

	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{FullMethod: "/test"}, handler)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}
}

func TestStreamRequestIDInterceptor_GeneratesID(t *testing.T) {
	interceptor := StreamRequestIDInterceptor()

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())
	mockStream := &mockServerStream{ctx: ctx}

	called := false
	var retrievedID string

	handler := func(srv interface{}, ss grpc.ServerStream) error {
		called = true
		retrievedID = RequestIDFromContext(ss.Context())
		return nil
	}

	err := interceptor(nil, mockStream, &grpc.StreamServerInfo{FullMethod: "/test"}, handler)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("handler was not called")
	}

	if retrievedID == "" {
		t.Fatal("request ID was empty")
	}

	if len(retrievedID) < 32 {
		t.Errorf("expected request ID length >= 32, got %d", len(retrievedID))
	}
}

func TestUnaryRequestIDClientInterceptor_PropagatesToOutgoingMetadata(t *testing.T) {
	interceptor := UnaryRequestIDClientInterceptor()
	requestID := "test-request-id-client"

	ctx := context.WithValue(context.Background(), contextKey, requestID)

	called := false
	var capturedCtx context.Context

	invoker := func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		opts ...grpc.CallOption,
	) error {
		called = true
		capturedCtx = ctx
		return nil
	}

	err := interceptor(ctx, "/test", nil, nil, nil, invoker)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("invoker was not called")
	}

	md, ok := metadata.FromOutgoingContext(capturedCtx)
	if !ok {
		t.Fatal("no outgoing metadata in context")
	}

	values := md.Get(requestIDMetadataKey)
	if len(values) == 0 {
		t.Fatal("request ID not found in outgoing metadata")
	}

	if values[0] != requestID {
		t.Errorf("expected request ID %q in metadata, got %q", requestID, values[0])
	}
}

func TestStreamRequestIDClientInterceptor_PropagatesToOutgoingMetadata(t *testing.T) {
	interceptor := StreamRequestIDClientInterceptor()
	requestID := "test-request-id-stream-client"

	ctx := context.WithValue(context.Background(), contextKey, requestID)

	called := false
	var capturedCtx context.Context

	streamer := func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		called = true
		capturedCtx = ctx
		return nil, nil
	}

	_, err := interceptor(ctx, &grpc.StreamDesc{}, nil, "/test", streamer)
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}

	if !called {
		t.Fatal("streamer was not called")
	}

	md, ok := metadata.FromOutgoingContext(capturedCtx)
	if !ok {
		t.Fatal("no outgoing metadata in context")
	}

	values := md.Get(requestIDMetadataKey)
	if len(values) == 0 {
		t.Fatal("request ID not found in outgoing metadata")
	}

	if values[0] != requestID {
		t.Errorf("expected request ID %q in metadata, got %q", requestID, values[0])
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) SendHeader(metadata.MD) error {
	return status.Error(codes.Unimplemented, "mock")
}

func (m *mockServerStream) SetHeader(metadata.MD) error {
	return status.Error(codes.Unimplemented, "mock")
}

func (m *mockServerStream) SetTrailer(metadata.MD) {
}

func (m *mockServerStream) SendMsg(msg interface{}) error {
	return status.Error(codes.Unimplemented, "mock")
}

func (m *mockServerStream) RecvMsg(msg interface{}) error {
	return status.Error(codes.Unimplemented, "mock")
}
