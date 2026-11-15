package opshttp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	coordserver "github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
)

const testToken = "0123456789abcdef0123456789abcdef"

type fixedDispatchCount int64

func (c fixedDispatchCount) Dispatches() int64 { return int64(c) }

func TestServer_ExperimentEndpoints(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tasks.jsonl")
	coord := coordserver.New(coordserver.Config{HeartbeatTTL: time.Minute, TaskLogPath: path})
	defer coord.Stop()
	server := &Server{AuthToken: testToken, EventSink: coord, DispatchCounter: coord}

	cases := []struct {
		name, method, path, body, auth string
		want                           int
	}{
		{"count open", "GET", "/dispatch-count", "", "", 200},
		{"count method", "POST", "/dispatch-count", "", "", 405},
		{"event method", "GET", "/events", "", "", 405},
		{"event unauthorized", "POST", "/events", `{}`, "", 401},
		{"event valid", "POST", "/events", `{"kind":"drift_on","target":"worker-5:50051","dispatch_count":120,"detail":"started"}`, "Bearer " + testToken, 204},
		{"event raw token", "POST", "/events", `{"kind":"drift_off","target":"worker-5:50051","dispatch_count":130}`, testToken, 204},
		{"bad kind", "POST", "/events", `{"kind":"other","target":"worker-5"}`, testToken, 400},
		{"missing target", "POST", "/events", `{"kind":"drift_note"}`, testToken, 400},
		{"long target", "POST", "/events", `{"kind":"drift_note","target":"` + strings.Repeat("x", 129) + `"}`, testToken, 400},
		{"long detail", "POST", "/events", `{"kind":"drift_note","target":"x","detail":"` + strings.Repeat("x", 513) + `"}`, testToken, 400},
		{"long body", "POST", "/events", `{"kind":"drift_note","target":"x","detail":"` + strings.Repeat(" ", 4096) + `"}`, testToken, 400},
		{"negative count", "POST", "/events", `{"kind":"drift_note","target":"x","dispatch_count":-1}`, testToken, 400},
		{"unknown field", "POST", "/events", `{"kind":"drift_note","target":"x","other":1}`, testToken, 400},
		{"two objects", "POST", "/events", `{"kind":"drift_note","target":"x"}{}`, testToken, 400},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			if tc.auth != "" {
				req.Header.Set("Authorization", tc.auth)
			}
			w := httptest.NewRecorder()
			server.Handler().ServeHTTP(w, req)
			if w.Code != tc.want {
				t.Errorf("status=%d, want=%d, body=%s", w.Code, tc.want, w.Body.String())
			}
			if tc.name == "count open" && w.Body.String() != "{\"dispatches\":0}\n" {
				t.Errorf("count response=%s", w.Body.String())
			}
		})
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("event records=%d: %s", len(lines), data)
	}
	var event map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &event); err != nil {
		t.Fatal(err)
	}
	if event["event"] != "injected_event" || event["kind"] != "drift_on" ||
		event["target"] != "worker-5:50051" || event["dispatch_count"] != float64(120) || event["ts"] == "" {
		t.Errorf("logged event=%v", event)
	}
}

func TestServer_EventsUnavailable(t *testing.T) {
	server := &Server{}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("POST", "/events", strings.NewReader(`{"kind":"drift_on","target":"x"}`)))
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status=%d, want 503", w.Code)
	}
}

func TestServer_DispatchCount(t *testing.T) {
	server := &Server{DispatchCounter: fixedDispatchCount(7)}
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/dispatch-count", nil))
	if w.Code != http.StatusOK || w.Body.String() != "{\"dispatches\":7}\n" {
		t.Errorf("response=%d %s", w.Code, w.Body.String())
	}
}
