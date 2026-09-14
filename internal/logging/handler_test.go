package logging

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogLevelHandler_GET(t *testing.T) {
	// Set a known level for testing
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	handler := NewLogLevelHandler("")
	req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp LogLevelResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Level != "info" {
		t.Errorf("expected level 'info', got %q", resp.Level)
	}

	if resp.Previous != "" {
		t.Errorf("expected empty Previous field for GET, got %q", resp.Previous)
	}
}

func TestLogLevelHandler_PUT_ValidLevel(t *testing.T) {
	// Start with a known level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	handler := NewLogLevelHandler("")

	body := LogLevelRequest{Level: "debug"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/log-level", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp LogLevelResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Level != "debug" {
		t.Errorf("expected level 'debug', got %q", resp.Level)
	}

	if resp.Previous != "info" {
		t.Errorf("expected previous 'info', got %q", resp.Previous)
	}

	// Verify the global level was actually changed
	if zerolog.GlobalLevel() != zerolog.DebugLevel {
		t.Errorf("expected global level to be DebugLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestLogLevelHandler_POST_ValidLevel(t *testing.T) {
	// Start with a known level
	zerolog.SetGlobalLevel(zerolog.WarnLevel)

	handler := NewLogLevelHandler("")

	body := LogLevelRequest{Level: "error"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/log-level", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp LogLevelResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Level != "error" {
		t.Errorf("expected level 'error', got %q", resp.Level)
	}

	if resp.Previous != "warn" {
		t.Errorf("expected previous 'warn', got %q", resp.Previous)
	}

	// Verify the global level was actually changed
	if zerolog.GlobalLevel() != zerolog.ErrorLevel {
		t.Errorf("expected global level to be ErrorLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestLogLevelHandler_InvalidLevel(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	handler := NewLogLevelHandler("")

	body := LogLevelRequest{Level: "invalid"}
	bodyBytes, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/log-level", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatal("expected error message, got empty string")
	}

	// Verify the global level did NOT change
	if zerolog.GlobalLevel() != zerolog.InfoLevel {
		t.Errorf("expected global level to remain InfoLevel, got %v", zerolog.GlobalLevel())
	}
}

func TestLogLevelHandler_InvalidJSON(t *testing.T) {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	handler := NewLogLevelHandler("")

	req := httptest.NewRequest(http.MethodPut, "/log-level", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatal("expected error message, got empty string")
	}
}

func TestLogLevelHandler_AllValidLevels(t *testing.T) {
	handler := NewLogLevelHandler("")

	levels := []string{"trace", "debug", "info", "warn", "error", "fatal", "panic"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			body := LogLevelRequest{Level: level}
			bodyBytes, _ := json.Marshal(body)
			req := httptest.NewRequest(http.MethodPut, "/log-level", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			var resp LogLevelResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if resp.Level != level {
				t.Errorf("expected level %q, got %q", level, resp.Level)
			}
		})
	}
}

func TestLogLevelHandler_MethodNotAllowed(t *testing.T) {
	handler := NewLogLevelHandler("")

	req := httptest.NewRequest(http.MethodDelete, "/log-level", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status 405, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp.Error == "" {
		t.Fatal("expected error message, got empty string")
	}
}

func TestLogLevelHandler_ContentTypeHeader(t *testing.T) {
	handler := NewLogLevelHandler("")

	req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", contentType)
	}
}

// --- Authentication gating ---

const testToken = "0123456789abcdef0123456789abcdef" // 32 chars: auth.MinTokenLength

func TestLogLevelHandler_AuthGating(t *testing.T) {
	previous := zerolog.GlobalLevel()
	t.Cleanup(func() { zerolog.SetGlobalLevel(previous) })

	tests := []struct {
		name     string
		method   string
		authHdr  string
		body     string
		wantCode int
	}{
		{name: "get without token is rejected", method: http.MethodGet, authHdr: "", wantCode: http.StatusUnauthorized},
		{name: "get with wrong token is rejected", method: http.MethodGet, authHdr: "Bearer " + strings.Repeat("f", 32), wantCode: http.StatusUnauthorized},
		{name: "get with short token is rejected", method: http.MethodGet, authHdr: "Bearer short", wantCode: http.StatusUnauthorized},
		{name: "get with correct token passes", method: http.MethodGet, authHdr: "Bearer " + testToken, wantCode: http.StatusOK},
		{name: "get with correct raw token (no Bearer prefix) passes", method: http.MethodGet, authHdr: testToken, wantCode: http.StatusOK},
		{name: "set with correct token passes", method: http.MethodPut, authHdr: "Bearer " + testToken, body: `{"level":"debug"}`, wantCode: http.StatusOK},
		{name: "set without token is rejected", method: http.MethodPut, authHdr: "", body: `{"level":"debug"}`, wantCode: http.StatusUnauthorized},
		{name: "invalid level still validated after auth", method: http.MethodPut, authHdr: "Bearer " + testToken, body: `{"level":"nope"}`, wantCode: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewLogLevelHandler(testToken)
			var body io.Reader
			if tt.body != "" {
				body = strings.NewReader(tt.body)
			}
			req := httptest.NewRequest(tt.method, "/log-level", body)
			if tt.authHdr != "" {
				req.Header.Set("Authorization", tt.authHdr)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d; body: %s", rec.Code, tt.wantCode, rec.Body.String())
			}
		})
	}
}

func TestLogLevelHandler_NoTokenConfiguredStaysOpen(t *testing.T) {
	previous := zerolog.GlobalLevel()
	t.Cleanup(func() { zerolog.SetGlobalLevel(previous) })

	handler := NewLogLevelHandler("")
	req := httptest.NewRequest(http.MethodGet, "/log-level", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("unconfigured handler must stay open: status = %d, want %d", rec.Code, http.StatusOK)
	}
}
