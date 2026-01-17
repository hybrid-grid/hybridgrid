package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

const (
	// MinTokenLength is the minimum required token length.
	MinTokenLength = 32

	// DefaultTokenLength is the default length for generated tokens.
	DefaultTokenLength = 64
)

// ValidateToken validates an authentication token.
// Uses constant-time comparison to prevent timing attacks.
func ValidateToken(provided, expected string) bool {
	if len(provided) < MinTokenLength {
		return false
	}
	if len(expected) < MinTokenLength {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}

// GenerateToken generates a cryptographically secure random token.
func GenerateToken() (string, error) {
	return GenerateTokenWithLength(DefaultTokenLength)
}

// GenerateTokenWithLength generates a token of the specified length.
func GenerateTokenWithLength(length int) (string, error) {
	if length < MinTokenLength {
		return "", fmt.Errorf("token length must be at least %d", MinTokenLength)
	}

	// Generate random bytes (half the length since hex encoding doubles it)
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// ParseBearerToken extracts the token from a Bearer authorization string.
// Returns the token and true if valid, empty string and false otherwise.
func ParseBearerToken(auth string) (string, bool) {
	const prefix = "Bearer "
	if len(auth) <= len(prefix) {
		return "", false
	}
	if auth[:len(prefix)] != prefix {
		return "", false
	}
	return auth[len(prefix):], true
}
