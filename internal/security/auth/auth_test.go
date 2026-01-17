package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestValidateToken(t *testing.T) {
	tests := []struct {
		name     string
		provided string
		expected string
		want     bool
	}{
		{
			name:     "matching tokens",
			provided: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expected: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			want:     true,
		},
		{
			name:     "different tokens",
			provided: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expected: "00000000000000000000000000000000",
			want:     false,
		},
		{
			name:     "provided too short",
			provided: "short",
			expected: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			want:     false,
		},
		{
			name:     "expected too short",
			provided: "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
			expected: "short",
			want:     false,
		},
		{
			name:     "both empty",
			provided: "",
			expected: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateToken(tt.provided, tt.expected); got != tt.want {
				t.Errorf("ValidateToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if len(token) != DefaultTokenLength {
		t.Errorf("Token length = %d, want %d", len(token), DefaultTokenLength)
	}

	// Generate another token and verify they're different
	token2, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == token2 {
		t.Error("Two generated tokens should be different")
	}
}

func TestGenerateTokenWithLength(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{"valid 32", 32, false},
		{"valid 64", 64, false},
		{"valid 128", 128, false},
		{"too short", 16, true},
		{"minimum", MinTokenLength, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := GenerateTokenWithLength(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTokenWithLength() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(token) != tt.length {
				t.Errorf("Token length = %d, want %d", len(token), tt.length)
			}
		})
	}
}

func TestParseBearerToken(t *testing.T) {
	tests := []struct {
		name      string
		auth      string
		wantToken string
		wantOK    bool
	}{
		{
			name:      "valid bearer",
			auth:      "Bearer mytoken123",
			wantToken: "mytoken123",
			wantOK:    true,
		},
		{
			name:      "missing prefix",
			auth:      "mytoken123",
			wantToken: "",
			wantOK:    false,
		},
		{
			name:      "wrong prefix",
			auth:      "Basic mytoken123",
			wantToken: "",
			wantOK:    false,
		},
		{
			name:      "empty token",
			auth:      "Bearer ",
			wantToken: "",
			wantOK:    false,
		},
		{
			name:      "empty string",
			auth:      "",
			wantToken: "",
			wantOK:    false,
		},
		{
			name:      "just bearer",
			auth:      "Bearer",
			wantToken: "",
			wantOK:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, ok := ParseBearerToken(tt.auth)
			if ok != tt.wantOK {
				t.Errorf("ParseBearerToken() ok = %v, want %v", ok, tt.wantOK)
			}
			if token != tt.wantToken {
				t.Errorf("ParseBearerToken() token = %q, want %q", token, tt.wantToken)
			}
		})
	}
}

func TestInterceptor_Disabled(t *testing.T) {
	cfg := Config{
		Enabled: false,
		Token:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
	}
	interceptor := NewInterceptor(cfg)

	// Create a test handler that returns success
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Test unary interceptor
	result, err := interceptor.UnaryServerInterceptor()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
		handler,
	)

	if err != nil {
		t.Errorf("Disabled interceptor returned error: %v", err)
	}
	if result != "success" {
		t.Errorf("Result = %v, want 'success'", result)
	}
}

func TestInterceptor_MissingToken(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Token:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
	}
	interceptor := NewInterceptor(cfg)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Call without metadata
	_, err := interceptor.UnaryServerInterceptor()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
		handler,
	)

	if err == nil {
		t.Error("Expected error for missing token")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("Error code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestInterceptor_InvalidToken(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Token:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
	}
	interceptor := NewInterceptor(cfg)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Create context with wrong token
	md := metadata.New(map[string]string{
		AuthorizationKey: "Bearer wrongtoken00000000000000000000",
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	_, err := interceptor.UnaryServerInterceptor()(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
		handler,
	)

	if err == nil {
		t.Error("Expected error for invalid token")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("Error code = %v, want Unauthenticated", status.Code(err))
	}
}

func TestInterceptor_ValidToken(t *testing.T) {
	token := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	cfg := Config{
		Enabled: true,
		Token:   token,
	}
	interceptor := NewInterceptor(cfg)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Create context with correct token
	md := metadata.New(map[string]string{
		AuthorizationKey: "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	result, err := interceptor.UnaryServerInterceptor()(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/test/Method"},
		handler,
	)

	if err != nil {
		t.Errorf("Valid token returned error: %v", err)
	}
	if result != "success" {
		t.Errorf("Result = %v, want 'success'", result)
	}
}

func TestInterceptor_SkipMethods(t *testing.T) {
	cfg := Config{
		Enabled:     true,
		Token:       "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
		SkipMethods: []string{"/health/Check"},
	}
	interceptor := NewInterceptor(cfg)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	// Call skipped method without token
	result, err := interceptor.UnaryServerInterceptor()(
		context.Background(),
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/health/Check"},
		handler,
	)

	if err != nil {
		t.Errorf("Skipped method returned error: %v", err)
	}
	if result != "success" {
		t.Errorf("Result = %v, want 'success'", result)
	}
}

func TestContextWithToken(t *testing.T) {
	token := "mytoken123"
	ctx := ContextWithToken(context.Background(), token)

	// Extract outgoing metadata
	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		t.Fatal("No outgoing metadata")
	}

	values := md.Get(AuthorizationKey)
	if len(values) == 0 {
		t.Fatal("No authorization header")
	}

	expected := "Bearer " + token
	if values[0] != expected {
		t.Errorf("Authorization = %q, want %q", values[0], expected)
	}
}

// Mock server stream for testing
type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func TestStreamInterceptor_ValidToken(t *testing.T) {
	token := "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4"
	cfg := Config{
		Enabled: true,
		Token:   token,
	}
	interceptor := NewInterceptor(cfg)

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	// Create context with correct token
	md := metadata.New(map[string]string{
		AuthorizationKey: "Bearer " + token,
	})
	ctx := metadata.NewIncomingContext(context.Background(), md)

	err := interceptor.StreamServerInterceptor()(
		nil,
		&mockServerStream{ctx: ctx},
		&grpc.StreamServerInfo{FullMethod: "/test/Stream"},
		handler,
	)

	if err != nil {
		t.Errorf("Valid token returned error: %v", err)
	}
}

func TestStreamInterceptor_InvalidToken(t *testing.T) {
	cfg := Config{
		Enabled: true,
		Token:   "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
	}
	interceptor := NewInterceptor(cfg)

	handler := func(srv interface{}, stream grpc.ServerStream) error {
		return nil
	}

	err := interceptor.StreamServerInterceptor()(
		nil,
		&mockServerStream{ctx: context.Background()},
		&grpc.StreamServerInfo{FullMethod: "/test/Stream"},
		handler,
	)

	if err == nil {
		t.Error("Expected error for missing token")
	}
	if status.Code(err) != codes.Unauthenticated {
		t.Errorf("Error code = %v, want Unauthenticated", status.Code(err))
	}
}
