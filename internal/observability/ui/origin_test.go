package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

func TestNewOriginChecker(t *testing.T) {
	tests := []struct {
		name    string
		allowed []string
		host    string
		origin  string // empty means no Origin header
		want    bool
	}{
		{name: "no origin header", host: "dash.lan:8081", origin: "", want: true},
		{name: "same origin", host: "dash.lan:8081", origin: "http://dash.lan:8081", want: true},
		{name: "same host different case", host: "Dash.LAN:8081", origin: "http://dash.lan:8081", want: true},
		{name: "same host https scheme", host: "dash.lan:8081", origin: "https://dash.lan:8081", want: true},
		{name: "different port", host: "dash.lan:8081", origin: "http://dash.lan:5173", want: false},
		{name: "different host", host: "dash.lan:8081", origin: "http://evil.example", want: false},
		{name: "malformed origin", host: "dash.lan:8081", origin: "://not a url", want: false},
		{name: "opaque null origin", host: "dash.lan:8081", origin: "null", want: false},
		{name: "allowlisted origin", allowed: []string{"http://localhost:5173"}, host: "dash.lan:8081", origin: "http://localhost:5173", want: true},
		{name: "allowlisted origin case and trailing slash", allowed: []string{"HTTP://Localhost:5173/"}, host: "dash.lan:8081", origin: "http://localhost:5173", want: true},
		{name: "allowlist does not match other origin", allowed: []string{"http://localhost:5173"}, host: "dash.lan:8081", origin: "http://localhost:3000", want: false},
		{name: "allowlist still allows same origin", allowed: []string{"http://localhost:5173"}, host: "dash.lan:8081", origin: "http://dash.lan:8081", want: true},
		{name: "wildcard allows foreign origin", allowed: []string{"*"}, host: "dash.lan:8081", origin: "http://evil.example", want: true},
		{name: "wildcard allows malformed origin", allowed: []string{"*"}, host: "dash.lan:8081", origin: "://not a url", want: true},
		{name: "blank allowlist entries ignored", allowed: []string{"", "  "}, host: "dash.lan:8081", origin: "http://evil.example", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := newOriginChecker(tt.allowed)
			r := httptest.NewRequest(http.MethodGet, "/ws", nil)
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if got := check(r); got != tt.want {
				t.Errorf("check(origin=%q, host=%q, allowed=%v) = %v, want %v", tt.origin, tt.host, tt.allowed, got, tt.want)
			}
		})
	}
}

// TestServer_WebSocketOriginEnforcement dials the real /ws route with
// foreign, same-origin and allowlisted Origin headers.
func TestServer_WebSocketOriginEnforcement(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AllowedOrigins = []string{"http://localhost:5173"}
	s := New(cfg, nil)
	go s.hub.Run()
	defer s.hub.Stop()

	ts := httptest.NewServer(s.routes())
	defer ts.Close()
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "no origin", origin: "", want: true},
		{name: "same origin", origin: ts.URL, want: true},
		{name: "allowlisted origin", origin: "http://localhost:5173", want: true},
		{name: "foreign origin", origin: "http://evil.example", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hdr := http.Header{}
			if tt.origin != "" {
				hdr.Set("Origin", tt.origin)
			}
			ws, resp, err := websocket.DefaultDialer.Dial(wsURL, hdr)
			if ws != nil {
				defer ws.Close()
			}
			if tt.want {
				if err != nil {
					t.Fatalf("dial with origin %q: %v", tt.origin, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("dial with origin %q succeeded, want rejection", tt.origin)
			}
			if resp == nil || resp.StatusCode != http.StatusForbidden {
				t.Errorf("rejected dial response = %v, want 403", resp)
			}
		})
	}
}
