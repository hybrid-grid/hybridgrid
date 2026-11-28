package blobcache

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newTestCache(t *testing.T, cfg Config) *Cache {
	t.Helper()
	if cfg.Dir == "" {
		cfg.Dir = t.TempDir()
	}
	c, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func do(t *testing.T, h http.Handler, method, key, body string, setup func(*http.Request)) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, GradlePrefix+key, strings.NewReader(body))
	if setup != nil {
		setup(req)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestCache_GradleProtocolRoundTrip(t *testing.T) {
	c := newTestCache(t, Config{})
	if rec := do(t, c, http.MethodGet, "abc123", "", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("GET missing = %d, want 404", rec.Code)
	}
	if rec := do(t, c, http.MethodPut, "abc123", "task-output", nil); rec.Code != http.StatusCreated {
		t.Fatalf("PUT = %d, want 201", rec.Code)
	}
	rec := do(t, c, http.MethodGet, "abc123", "", nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "task-output" {
		t.Fatalf("GET = %d %q", rec.Code, rec.Body.String())
	}
	if rec := do(t, c, http.MethodHead, "abc123", "", nil); rec.Code != http.StatusOK || rec.Body.Len() != 0 {
		t.Fatalf("HEAD = %d, body %d bytes", rec.Code, rec.Body.Len())
	}
	if rec := do(t, c, http.MethodDelete, "abc123", "", nil); rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE = %d, want 405", rec.Code)
	}
	if rec := do(t, c, http.MethodGet, "", "", nil); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"stores":1`) {
		t.Fatalf("stats endpoint = %d %s", rec.Code, rec.Body.String())
	}
	st := c.Stats()
	if st.Hits != 2 || st.Misses != 1 || st.Stores != 1 || st.Entries != 1 || st.TotalBytes != int64(len("task-output")) {
		t.Fatalf("stats = %+v", st)
	}
}

func TestCache_RejectsUnsafeKeys(t *testing.T) {
	c := newTestCache(t, Config{})
	for _, key := range []string{"", "../escape", "a/b", "a.b", strings.Repeat("a", maxKeyLen+1)} {
		if rec := do(t, c, http.MethodPut, key, "x", nil); rec.Code != http.StatusBadRequest {
			t.Errorf("PUT %q = %d, want 400", key, rec.Code)
		}
	}
	if entries, _ := os.ReadDir(c.cfg.Dir); len(entries) != 0 {
		t.Fatalf("unsafe keys left files: %v", entries)
	}
}

func TestCache_EntrySizeLimit(t *testing.T) {
	c := newTestCache(t, Config{MaxEntryBytes: 4})
	if rec := do(t, c, http.MethodPut, "big", "12345", nil); rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("oversized PUT = %d, want 413", rec.Code)
	}
	// Chunked upload without Content-Length is caught while copying.
	req := httptest.NewRequest(http.MethodPut, GradlePrefix+"big", io.MultiReader(strings.NewReader("123"), strings.NewReader("45")))
	req.ContentLength = -1
	rec := httptest.NewRecorder()
	c.ServeHTTP(rec, req)
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("chunked oversized PUT = %d, want 413", rec.Code)
	}
	if entries, _ := os.ReadDir(c.cfg.Dir); len(entries) != 0 {
		t.Fatalf("oversized upload left files: %v", entries)
	}
}

func TestCache_EvictsLeastRecentlyUsed(t *testing.T) {
	c := newTestCache(t, Config{MaxTotalBytes: 25})
	clock := time.Unix(1_700_000_000, 0)
	c.now = func() time.Time { return clock }
	put := func(key string) {
		t.Helper()
		if err := c.Put(key, bytes.NewReader(make([]byte, 10))); err != nil {
			t.Fatal(err)
		}
		p := filepath.Join(c.cfg.Dir, key)
		_ = os.Chtimes(p, clock, clock)
		clock = clock.Add(time.Minute)
	}
	put("a")
	put("b")
	f, _, err := c.Open("a") // a is now newer than b
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	clock = clock.Add(time.Minute)
	put("c") // 30 bytes > 25: evict down to <= 22, oldest (b) first

	if _, _, err := c.Open("b"); err != ErrNotFound {
		t.Fatalf("b should be evicted, got %v", err)
	}
	for _, key := range []string{"a", "c"} {
		f, _, err := c.Open(key)
		if err != nil {
			t.Fatalf("%s evicted: %v", key, err)
		}
		f.Close()
	}
	if st := c.Stats(); st.Evictions != 1 || st.TotalBytes != 20 || st.Entries != 2 {
		t.Fatalf("stats = %+v", st)
	}
}

func TestCache_ReplaceKeepsAccounting(t *testing.T) {
	c := newTestCache(t, Config{})
	_ = c.Put("k", strings.NewReader("12345"))
	_ = c.Put("k", strings.NewReader("12"))
	if st := c.Stats(); st.Entries != 1 || st.TotalBytes != 2 {
		t.Fatalf("stats = %+v", st)
	}
	if !c.Has("k") || c.Has("missing") || c.Has("../k") {
		t.Fatal("Has disagrees with the store")
	}
	reopened := newTestCache(t, Config{Dir: c.cfg.Dir})
	if st := reopened.Stats(); st.Entries != 1 || st.TotalBytes != 2 {
		t.Fatalf("reopened stats = %+v", st)
	}
}

func TestCache_TokenAuth(t *testing.T) {
	const token = "0123456789abcdef0123456789abcdef"
	c := newTestCache(t, Config{Token: token})
	tests := []struct {
		name  string
		setup func(*http.Request)
		want  int
	}{
		{"none", nil, http.StatusUnauthorized},
		{"wrong basic", func(r *http.Request) { r.SetBasicAuth("hg", "nope") }, http.StatusUnauthorized},
		{"basic", func(r *http.Request) { r.SetBasicAuth("hg", token) }, http.StatusNotFound},
		{"bearer", func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+token) }, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if rec := do(t, c, http.MethodGet, "k", "", tt.setup); rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}
