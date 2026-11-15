package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/h3nr1-d14z/hybridgrid/internal/coordinator/server"
)

func TestWebhook_DeliversStartedAndCompleted(t *testing.T) {
	var mu sync.Mutex
	var received []event
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var ev event
		if err := json.NewDecoder(r.Body).Decode(&ev); err != nil {
			t.Errorf("decode: %v", err)
		}
		mu.Lock()
		received = append(received, ev)
		mu.Unlock()
		w.WriteHeader(200)
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 0, Buffer: 8})
	defer n.Close()

	n.NotifyTaskStarted(&server.TaskEvent{ID: "t1", BuildType: "cpp", Status: "running"})
	n.NotifyTaskCompleted(&server.TaskEvent{ID: "t1", BuildType: "cpp", Status: "completed", DurationMs: 42})

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		count := len(received)
		mu.Unlock()
		if count == 2 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
	if received[0].Type != "task_started" || received[0].ID != "t1" {
		t.Errorf("started event wrong: %+v", received[0])
	}
	if received[1].Type != "task_completed" || received[1].DurationMs != 42 {
		t.Errorf("completed event wrong: %+v", received[1])
	}
}

func TestWebhook_NotifyNeverBlocks(t *testing.T) {
	release := make(chan struct{})
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		<-release
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 0, Buffer: 4})

	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			n.NotifyTaskStarted(&server.TaskEvent{ID: "flood"})
		}
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("NotifyTaskStarted blocked on a stuck endpoint")
	}
	close(release) // unblock the in-flight POST before Close waits for it
	n.Close()
}

func TestWebhook_RetriesThenSucceeds(t *testing.T) {
	var attempts atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(500)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 3, Buffer: 8})
	defer n.Close()

	n.NotifyTaskStarted(&server.TaskEvent{ID: "retry-me"})

	deadline := time.Now().Add(3 * time.Second)
	for attempts.Load() < 3 {
		if time.Now().After(deadline) {
			t.Fatalf("retry did not succeed in time (attempts=%d)", attempts.Load())
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestWebhook_OverflowDropsWithoutBlocking(t *testing.T) {
	// Endpoint that accepts instantly but we hammer past the buffer.
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 0, Buffer: 4})
	defer n.Close()

	start := time.Now()
	for i := 0; i < 1000; i++ {
		n.NotifyTaskCompleted(&server.TaskEvent{ID: "drop-me"})
	}
	if elapsed := time.Since(start); elapsed > time.Second {
		t.Fatalf("enqueue of 1000 events over a full buffer blocked: %v", elapsed)
	}
}

func TestWebhook_SignsDeliveries(t *testing.T) {
	const secret = "s3cret"
	var mu sync.Mutex
	var sigs []string
	var bodies []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		mu.Lock()
		sigs = append(sigs, r.Header.Get("X-Hybridgrid-Signature"))
		bodies = append(bodies, string(b))
		mu.Unlock()
		w.WriteHeader(200)
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 0, Buffer: 8, Secret: secret})
	defer n.Close()

	n.NotifyTaskStarted(&server.TaskEvent{ID: "t9", BuildType: "cpp", Status: "running"})

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		count := len(sigs)
		mu.Unlock()
		if count == 1 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(sigs) != 1 {
		t.Fatalf("expected 1 signed delivery, got %d", len(sigs))
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(bodies[0]))
	want := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if sigs[0] != want {
		t.Fatalf("signature mismatch: got %q want %q", sigs[0], want)
	}
}

func TestWebhook_NoSignatureHeaderWithoutSecret(t *testing.T) {
	var mu sync.Mutex
	var got string
	var count int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		got = r.Header.Get("X-Hybridgrid-Signature")
		count++
		mu.Unlock()
		w.WriteHeader(200)
	}))
	defer srv.Close()

	n := New(Config{Endpoints: []string{srv.URL}, Retries: 0, Buffer: 8})
	defer n.Close()

	n.NotifyTaskStarted(&server.TaskEvent{ID: "t10", BuildType: "cpp", Status: "running"})

	deadline := time.Now().Add(2 * time.Second)
	for {
		mu.Lock()
		c := count
		mu.Unlock()
		if c == 1 || time.Now().After(deadline) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	mu.Lock()
	defer mu.Unlock()
	if count != 1 {
		t.Fatalf("expected 1 delivery, got %d", count)
	}
	if got != "" {
		t.Fatalf("expected no signature header without secret, got %q", got)
	}
}
