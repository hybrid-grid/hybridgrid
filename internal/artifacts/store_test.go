package artifacts

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newStore(t *testing.T, cfg Config) *Store {
	t.Helper()
	cfg.Dir = t.TempDir()
	s, err := New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return s
}

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		in, want string
		wantErr  bool
	}{
		{"app", "app", false},
		{"build/out/app.apk", "app.apk", false},
		{`..\..\evil.exe`, "evil.exe", false},
		{"../../etc/passwd", "passwd", false},
		{"", "", true},
		{"..", "", true},
		{".", "", true},
		{"/", "", true},
		{".manifest.json", "", true},
		{"bad\x00name", "", true},
		{strings.Repeat("a", 201), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := SanitizeName(tt.in)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Fatalf("SanitizeName(%q) = %q, %v; want %q, err=%v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

func TestPutListOpen(t *testing.T) {
	s := newStore(t, Config{})
	art, err := s.Put("build-1", "out/app", SourceClient, strings.NewReader("binary"), nil)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if art.Name != "app" || art.SizeBytes != 6 || len(art.SHA256) != 64 || art.Source != SourceClient {
		t.Fatalf("unexpected artifact %+v", art)
	}

	list, err := s.List("build-1")
	if err != nil || len(list) != 1 || list[0].Name != "app" {
		t.Fatalf("List = %+v, %v", list, err)
	}

	f, meta, err := s.Open("build-1", "app")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	data, _ := io.ReadAll(f)
	if string(data) != "binary" || meta.SHA256 != art.SHA256 {
		t.Fatalf("Open returned %q / %+v", data, meta)
	}
}

func TestPutReplacesSameName(t *testing.T) {
	s := newStore(t, Config{})
	if _, err := s.Put("b", "app", SourceClient, strings.NewReader("v1"), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Put("b", "app", SourceClient, strings.NewReader("version-2"), nil); err != nil {
		t.Fatal(err)
	}
	list, _ := s.List("b")
	if len(list) != 1 || list[0].SizeBytes != int64(len("version-2")) {
		t.Fatalf("List after replace = %+v", list)
	}
}

func TestPutRejectsOversizedAndLeavesNothing(t *testing.T) {
	s := newStore(t, Config{MaxFileBytes: 4})
	if _, err := s.Put("b", "big", SourceClient, strings.NewReader("12345"), nil); !errors.Is(err, ErrTooLarge) {
		t.Fatalf("Put oversized err = %v, want ErrTooLarge", err)
	}
	entries, _ := os.ReadDir(filepath.Join(s.Dir(), "b"))
	if len(entries) != 0 {
		t.Fatalf("oversized upload left files behind: %v", entries)
	}
	if _, err := s.Put("b", "ok", SourceClient, strings.NewReader("1234"), nil); err != nil {
		t.Fatalf("Put at the limit: %v", err)
	}
}

func TestInvalidInputs(t *testing.T) {
	s := newStore(t, Config{})
	if _, err := s.Put("../escape", "app", SourceClient, strings.NewReader("x"), nil); !errors.Is(err, ErrInvalidBuildID) {
		t.Fatalf("bad build id err = %v", err)
	}
	if _, err := s.Put("b", "..", SourceClient, strings.NewReader("x"), nil); !errors.Is(err, ErrInvalidName) {
		t.Fatalf("bad name err = %v", err)
	}
	if _, _, err := s.Open("b", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Open missing err = %v", err)
	}
	if list, err := s.List("unknown"); err != nil || len(list) != 0 {
		t.Fatalf("List unknown = %v, %v", list, err)
	}
}

func TestEvictsOldestBuildsOverTotal(t *testing.T) {
	s := newStore(t, Config{MaxTotalBytes: 10})
	clock := time.UnixMilli(1_000)
	s.now = func() time.Time { clock = clock.Add(time.Second); return clock }

	for _, id := range []string{"old", "mid", "new"} {
		if _, err := s.Put(id, "a", SourceClient, strings.NewReader("12345"), nil); err != nil {
			t.Fatal(err)
		}
	}
	// 15 bytes > 10: the oldest build goes, the newest is always kept.
	if list, _ := s.List("old"); len(list) != 0 {
		t.Fatalf("oldest build not evicted: %v", list)
	}
	for _, id := range []string{"mid", "new"} {
		if list, _ := s.List(id); len(list) != 1 {
			t.Fatalf("build %s evicted unexpectedly", id)
		}
	}
}
