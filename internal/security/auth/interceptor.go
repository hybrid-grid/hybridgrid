package auth

import (
	"context"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	// AuthorizationKey is the metadata key for the authorization token.
	AuthorizationKey = "authorization"
)

// Config holds authentication configuration.
type Config struct {
	// Enabled determines if authentication is required
	Enabled bool `yaml:"enabled" json:"enabled"`

	// Token is the expected authentication token
	Token string `yaml:"token" json:"token"`

	// SkipMethods lists method names that skip authentication
	SkipMethods []string `yaml:"skip_methods" json:"skip_methods"`
}

// DefaultConfig returns default auth configuration.
func DefaultConfig() Config {
	return Config{
		Enabled:     false,
		SkipMethods: []string{"/grpc.health.v1.Health/Check"},
	}
}

// Interceptor provides gRPC authentication interceptors.
type Interceptor struct {
	enabled     bool
	token       string
	skipMethods map[string]bool
}

// NewInterceptor creates a new authentication interceptor.
func NewInterceptor(cfg Config) *Interceptor {
	skipMethods := make(map[string]bool)
	for _, m := range cfg.SkipMethods {
		skipMethods[m] = true
	}

	return &Interceptor{
		enabled:     cfg.Enabled,
		token:       cfg.Token,
		skipMethods: skipMethods,
	}
}

// UnaryServerInterceptor returns a unary server interceptor that validates tokens.
func (i *Interceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !i.enabled {
			return handler(ctx, req)
		}

		if i.skipMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		if err := i.validateContext(ctx, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a stream server interceptor that validates tokens.
func (i *Interceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !i.enabled {
			return handler(srv, ss)
		}

		if i.skipMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		if err := i.validateContext(ss.Context(), info.FullMethod); err != nil {
			return err
		}

		return handler(srv, ss)
	}
}

// validateContext validates the authentication token from context metadata.
func (i *Interceptor) validateContext(ctx context.Context, method string) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Warn().Str("method", method).Msg("Auth failed: no metadata")
		return status.Error(codes.Unauthenticated, "no metadata provided")
	}

	values := md.Get(AuthorizationKey)
	if len(values) == 0 {
		log.Warn().Str("method", method).Msg("Auth failed: no authorization header")
		return status.Error(codes.Unauthenticated, "authorization token required")
	}

	token, ok := ParseBearerToken(values[0])
	if !ok {
		log.Warn().Str("method", method).Msg("Auth failed: invalid bearer format")
		return status.Error(codes.Unauthenticated, "invalid authorization format")
	}

	if !ValidateToken(token, i.token) {
		log.Warn().Str("method", method).Msg("Auth failed: invalid token")
		return status.Error(codes.Unauthenticated, "invalid token")
	}

	return nil
}

// ContextWithToken adds an authentication token to an outgoing context.
func ContextWithToken(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, AuthorizationKey, "Bearer "+token)
}

// UnaryClientInterceptor returns a unary client interceptor that adds the token.
func UnaryClientInterceptor(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if token != "" {
			ctx = ContextWithToken(ctx, token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// StreamClientInterceptor returns a stream client interceptor that adds the token.
func StreamClientInterceptor(token string) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if token != "" {
			ctx = ContextWithToken(ctx, token)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
