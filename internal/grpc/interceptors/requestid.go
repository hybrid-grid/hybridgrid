package interceptors

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	// requestIDMetadataKey is the gRPC metadata key for request ID.
	requestIDMetadataKey = "x-request-id"
)

// requestIDContextKey is an unexported type used for context keys.
type requestIDContextKey struct{}

// contextKey is the unexported key for storing request ID in context.
var contextKey requestIDContextKey

// generateRequestID generates a random 32-character hexadecimal request ID.
func generateRequestID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate request ID: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// RequestIDFromContext extracts the request ID from a context.
func RequestIDFromContext(ctx context.Context) string {
	id, ok := ctx.Value(contextKey).(string)
	if !ok {
		return ""
	}
	return id
}

// UnaryRequestIDInterceptor returns a unary server interceptor that extracts or generates
// a request ID from incoming metadata, adds it to context and outgoing metadata.
func UnaryRequestIDInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract request ID from incoming metadata
		md, _ := metadata.FromIncomingContext(ctx)
		var requestID string

		if values := md.Get(requestIDMetadataKey); len(values) > 0 {
			requestID = values[0]
		} else {
			// Generate new request ID if not provided
			var err error
			requestID, err = generateRequestID()
			if err != nil {
				return nil, err
			}
		}

		log.Info().Str("request_id", requestID).Msg("Request ID extracted/generated")

		// Add request ID to context
		ctx = context.WithValue(ctx, contextKey, requestID)

		// Add request ID to outgoing metadata
		ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)

		return handler(ctx, req)
	}
}

// StreamRequestIDInterceptor returns a stream server interceptor that extracts or generates
// a request ID from incoming metadata, adds it to context and outgoing metadata.
func StreamRequestIDInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()

		// Extract request ID from incoming metadata
		md, _ := metadata.FromIncomingContext(ctx)
		var requestID string

		if values := md.Get(requestIDMetadataKey); len(values) > 0 {
			requestID = values[0]
		} else {
			// Generate new request ID if not provided
			var err error
			requestID, err = generateRequestID()
			if err != nil {
				return err
			}
		}

		log.Info().Str("request_id", requestID).Msg("Request ID extracted/generated")

		// Add request ID to context
		ctx = context.WithValue(ctx, contextKey, requestID)

		// Add request ID to outgoing metadata
		ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)

		// Wrap the stream with the new context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		return handler(srv, wrappedStream)
	}
}

// wrappedServerStream wraps a gRPC ServerStream with a custom context.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the wrapped context.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// UnaryRequestIDClientInterceptor returns a unary client interceptor that propagates
// the request ID from context to outgoing metadata.
func UnaryRequestIDClientInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		// Extract request ID from context if present
		if requestID := RequestIDFromContext(ctx); requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)
		}

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamRequestIDClientInterceptor returns a stream client interceptor that propagates
// the request ID from context to outgoing metadata.
func StreamRequestIDClientInterceptor() grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		// Extract request ID from context if present
		if requestID := RequestIDFromContext(ctx); requestID != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, requestIDMetadataKey, requestID)
		}

		return streamer(ctx, desc, cc, method, opts...)
	}
}

// TODO(v0.3.0): Add auth interceptor
