// Package gradlecache is a Gradle remote build cache served by the
// coordinator, so a Gradle task output produced on one worker (Kotlin/Java
// compile, dexing, R8, resource merging, …) is reused by every other
// worker instead of being rebuilt.
//
// It speaks Gradle's HTTP build cache protocol: GET <prefix>/<key>
// returns the entry (200) or 404, PUT <prefix>/<key> stores it. Entries
// are opaque files on disk named by key; a total size cap evicts the
// least recently used entries.
package gradlecache

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/h3nr1-d14z/hybridgrid/internal/security/auth"
)

// Prefix is where the cache is mounted on the coordinator's HTTP server.
// A worker's Gradle points at http://<coordinator>:<http-port>/cache/gradle/.
const Prefix = "/cache/gradle/"

const maxKeyLen = 128

var (
	// ErrNotFound means no entry exists for the key.
	ErrNotFound = errors.New("cache entry not found")
	// ErrInvalidKey means the key is empty or not a plain token.
	ErrInvalidKey = errors.New("invalid cache key")
	// ErrTooLarge means an entry exceeded the per-entry cap.
	ErrTooLarge = errors.New("cache entry exceeds size limit")
)

// Config bounds the cache.
type Config struct {
	Dir           string
	MaxEntryBytes int64 // <= 0: unlimited
	MaxTotalBytes int64 // <= 0: unlimited
	// Token, when set, is required as the HTTP basic-auth password (any
	// user name) or as a bearer token.
	Token string
}

// Stats counts cache traffic since start.
type Stats struct {
	Hits       int64 `json:"hits"`
	Misses     int64 `json:"misses"`
	Stores     int64 `json:"stores"`
	Evictions  int64 `json:"evictions"`
	Entries    int64 `json:"entries"`
	TotalBytes int64 `json:"total_bytes"`
}

// Cache is safe for concurrent use.
type Cache struct {
	cfg Config

	mu      sync.Mutex
	total   int64
	entries int64

	hits, misses, stores, evictions atomic.Int64
	now                             func() time.Time
}

// New creates the cache directory and measures what is already in it.
func New(cfg Config) (*Cache, error) {
	if cfg.Dir == "" {
		return nil, errors.New("gradle cache dir required")
	}
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create gradle cache dir: %w", err)
	}
	c := &Cache{cfg: cfg, now: time.Now}
	files, err := c.list()
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		c.total += f.size
		c.entries++
	}
	return c, nil
}

func validKey(key string) bool {
	if key == "" || len(key) > maxKeyLen {
		return false
	}
	for _, r := range key {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
		default:
			return false
		}
	}
	return true
}

func (c *Cache) path(key string) (string, error) {
	if !validKey(key) {
		return "", ErrInvalidKey
	}
	return filepath.Join(c.cfg.Dir, key), nil
}

// Open returns the entry for key and its size, marking it recently used.
// The caller closes the file.
func (c *Cache) Open(key string) (*os.File, int64, error) {
	p, err := c.path(key)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(p)
	if errors.Is(err, os.ErrNotExist) {
		c.misses.Add(1)
		return nil, 0, ErrNotFound
	}
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open cache entry: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		f.Close()
		return nil, 0, fmt.Errorf("failed to stat cache entry: %w", err)
	}
	c.hits.Add(1)
	now := c.now()
	_ = os.Chtimes(p, now, now) // LRU: eviction drops the oldest mtime first
	return f, info.Size(), nil
}

// Put stores r under key, replacing any previous entry. The body goes to
// a temp file first, so an aborted or oversized upload leaves nothing.
func (c *Cache) Put(key string, r io.Reader) error {
	p, err := c.path(key)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(c.cfg.Dir, ".put-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	src := r
	if c.cfg.MaxEntryBytes > 0 {
		src = io.LimitReader(r, c.cfg.MaxEntryBytes+1)
	}
	n, err := io.Copy(tmp, src)
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("failed to write cache entry: %w", err)
	}
	if c.cfg.MaxEntryBytes > 0 && n > c.cfg.MaxEntryBytes {
		return ErrTooLarge
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	var old int64 = -1
	if info, err := os.Stat(p); err == nil {
		old = info.Size()
	}
	if err := os.Rename(tmp.Name(), p); err != nil {
		return fmt.Errorf("failed to store cache entry: %w", err)
	}
	if old >= 0 {
		c.total -= old
	} else {
		c.entries++
	}
	c.total += n
	c.stores.Add(1)
	c.evictLocked()
	return nil
}

type entry struct {
	name  string
	size  int64
	mtime time.Time
}

func (c *Cache) list() ([]entry, error) {
	des, err := os.ReadDir(c.cfg.Dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read gradle cache dir: %w", err)
	}
	out := make([]entry, 0, len(des))
	for _, de := range des {
		if !de.Type().IsRegular() || !validKey(de.Name()) {
			continue // temp files and foreign files are not entries
		}
		info, err := de.Info()
		if err != nil {
			continue
		}
		out = append(out, entry{name: de.Name(), size: info.Size(), mtime: info.ModTime()})
	}
	return out, nil
}

// evictLocked removes least recently used entries until the cache is at
// 90% of its cap, so a full cache does not rescan on every store.
func (c *Cache) evictLocked() {
	if c.cfg.MaxTotalBytes <= 0 || c.total <= c.cfg.MaxTotalBytes {
		return
	}
	files, err := c.list()
	if err != nil {
		log.Warn().Err(err).Msg("Gradle cache eviction skipped")
		return
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mtime.Before(files[j].mtime) })
	target := c.cfg.MaxTotalBytes * 9 / 10
	for _, f := range files {
		if c.total <= target {
			return
		}
		if err := os.Remove(filepath.Join(c.cfg.Dir, f.name)); err != nil {
			continue
		}
		c.total -= f.size
		c.entries--
		c.evictions.Add(1)
	}
}

// Stats returns counters and the current size.
func (c *Cache) Stats() Stats {
	c.mu.Lock()
	total, entries := c.total, c.entries
	c.mu.Unlock()
	return Stats{
		Hits: c.hits.Load(), Misses: c.misses.Load(), Stores: c.stores.Load(),
		Evictions: c.evictions.Load(), Entries: entries, TotalBytes: total,
	}
}

// transferWindow is how long one GET or PUT may take; the coordinator's
// HTTP server defaults to short timeouts meant for ops endpoints.
const transferWindow = 5 * time.Minute

// ServeHTTP implements Gradle's HTTP build cache protocol under Prefix.
// GET on Prefix itself returns Stats as JSON.
func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !c.authorized(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="hybridgrid gradle cache"`)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	key := strings.TrimPrefix(r.URL.Path, Prefix)
	if key == "" && r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(c.Stats())
		return
	}
	if !validKey(key) {
		http.Error(w, "invalid cache key", http.StatusBadRequest)
		return
	}
	rc := http.NewResponseController(w)
	deadline := time.Now().Add(transferWindow)
	_ = rc.SetReadDeadline(deadline)
	_ = rc.SetWriteDeadline(deadline)

	switch r.Method {
	case http.MethodGet, http.MethodHead:
		f, size, err := c.Open(key)
		if errors.Is(err, ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Warn().Err(err).Str("key", key).Msg("Gradle cache read failed")
			http.Error(w, "read failed", http.StatusInternalServerError)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "application/vnd.gradle.build-cache-artifact.v2")
		w.Header().Set("Content-Length", fmt.Sprint(size))
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodGet {
			_, _ = io.Copy(w, f)
		}
	case http.MethodPut:
		if c.cfg.MaxEntryBytes > 0 && r.ContentLength > c.cfg.MaxEntryBytes {
			// 413 tells Gradle to skip storing this entry, not to fail.
			http.Error(w, ErrTooLarge.Error(), http.StatusRequestEntityTooLarge)
			return
		}
		err := c.Put(key, r.Body)
		switch {
		case errors.Is(err, ErrTooLarge):
			http.Error(w, err.Error(), http.StatusRequestEntityTooLarge)
		case err != nil:
			log.Warn().Err(err).Str("key", key).Msg("Gradle cache store failed")
			http.Error(w, "store failed", http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (c *Cache) authorized(r *http.Request) bool {
	if c.cfg.Token == "" {
		return true
	}
	if _, pass, ok := r.BasicAuth(); ok {
		return auth.ValidateToken(pass, c.cfg.Token)
	}
	if tok, ok := auth.ParseBearerToken(r.Header.Get("Authorization")); ok {
		return auth.ValidateToken(tok, c.cfg.Token)
	}
	return false
}
