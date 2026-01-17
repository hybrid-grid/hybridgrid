package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Entry represents a cached item's metadata.
type Entry struct {
	Key        string    `json:"key"`
	Size       int64     `json:"size"`
	CreatedAt  time.Time `json:"created_at"`
	AccessedAt time.Time `json:"accessed_at"`
	Hits       int64     `json:"hits"`
}

// Store is a local file-based cache.
type Store struct {
	dir     string
	maxSize int64 // max size in bytes
	ttl     time.Duration

	mu        sync.RWMutex
	entries   map[string]*Entry
	totalSize int64
}

// NewStore creates a new cache store.
func NewStore(dir string, maxSizeMB int64, ttlHours int) (*Store, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache dir: %w", err)
	}

	s := &Store{
		dir:     dir,
		maxSize: maxSizeMB * 1024 * 1024,
		ttl:     time.Duration(ttlHours) * time.Hour,
		entries: make(map[string]*Entry),
	}

	// Load existing entries
	if err := s.loadIndex(); err != nil {
		// Index corruption, start fresh
		s.entries = make(map[string]*Entry)
	}

	return s, nil
}

// Get retrieves a cached item.
func (s *Store) Get(key string) (io.ReadCloser, bool) {
	s.mu.RLock()
	entry, ok := s.entries[key]
	s.mu.RUnlock()

	if !ok {
		return nil, false
	}

	// Check TTL
	if time.Since(entry.CreatedAt) > s.ttl {
		s.Delete(key)
		return nil, false
	}

	path := s.keyPath(key)
	f, err := os.Open(path)
	if err != nil {
		s.Delete(key)
		return nil, false
	}

	// Update access time and hits
	s.mu.Lock()
	entry.AccessedAt = time.Now()
	entry.Hits++
	s.mu.Unlock()

	return f, true
}

// Put stores an item in the cache.
func (s *Store) Put(key string, r io.Reader) error {
	path := s.keyPath(key)

	// Create parent directories
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create cache subdir: %w", err)
	}

	// Write to temp file first
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	size, err := io.Copy(f, r)
	f.Close()
	if err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to write cache: %w", err)
	}

	// Rename to final path
	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	now := time.Now()
	entry := &Entry{
		Key:        key,
		Size:       size,
		CreatedAt:  now,
		AccessedAt: now,
		Hits:       0,
	}

	s.mu.Lock()
	// Remove old entry if exists
	if old, ok := s.entries[key]; ok {
		s.totalSize -= old.Size
	}
	s.entries[key] = entry
	s.totalSize += size
	s.mu.Unlock()

	// Evict if over size limit
	s.evictIfNeeded()

	return s.saveIndex()
}

// PutBytes stores bytes in the cache.
func (s *Store) PutBytes(key string, data []byte) error {
	path := s.keyPath(key)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return err
	}

	now := time.Now()
	entry := &Entry{
		Key:        key,
		Size:       int64(len(data)),
		CreatedAt:  now,
		AccessedAt: now,
	}

	s.mu.Lock()
	if old, ok := s.entries[key]; ok {
		s.totalSize -= old.Size
	}
	s.entries[key] = entry
	s.totalSize += int64(len(data))
	s.mu.Unlock()

	s.evictIfNeeded()
	return s.saveIndex()
}

// GetBytes retrieves bytes from cache.
func (s *Store) GetBytes(key string) ([]byte, bool) {
	rc, ok := s.Get(key)
	if !ok {
		return nil, false
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, false
	}
	return data, true
}

// Delete removes an item from the cache.
func (s *Store) Delete(key string) error {
	s.mu.Lock()
	entry, ok := s.entries[key]
	if ok {
		delete(s.entries, key)
		s.totalSize -= entry.Size
	}
	s.mu.Unlock()

	path := s.keyPath(key)
	os.Remove(path)
	return s.saveIndex()
}

// Clear removes all items from the cache.
func (s *Store) Clear() error {
	s.mu.Lock()
	s.entries = make(map[string]*Entry)
	s.totalSize = 0
	s.mu.Unlock()

	// Remove all files in cache directory
	entries, _ := os.ReadDir(s.dir)
	for _, e := range entries {
		if e.Name() != "index.json" {
			os.RemoveAll(filepath.Join(s.dir, e.Name()))
		}
	}

	return s.saveIndex()
}

// Stats returns cache statistics.
type Stats struct {
	Entries   int   `json:"entries"`
	TotalSize int64 `json:"total_size"`
	MaxSize   int64 `json:"max_size"`
	TotalHits int64 `json:"total_hits"`
}

// Stats returns cache statistics.
func (s *Store) Stats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalHits int64
	for _, e := range s.entries {
		totalHits += e.Hits
	}

	return Stats{
		Entries:   len(s.entries),
		TotalSize: s.totalSize,
		MaxSize:   s.maxSize,
		TotalHits: totalHits,
	}
}

func (s *Store) keyPath(key string) string {
	// Use first 2 chars as subdirectory for better filesystem performance
	if len(key) < 2 {
		return filepath.Join(s.dir, key)
	}
	return filepath.Join(s.dir, key[:2], key)
}

func (s *Store) evictIfNeeded() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.totalSize <= s.maxSize {
		return
	}

	// Find oldest entries to evict (LRU)
	type kv struct {
		key   string
		entry *Entry
	}
	var sorted []kv
	for k, v := range s.entries {
		sorted = append(sorted, kv{k, v})
	}

	// Sort by access time (oldest first)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i].entry.AccessedAt.After(sorted[j].entry.AccessedAt) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Evict until under limit
	for _, kv := range sorted {
		if s.totalSize <= s.maxSize*8/10 { // Evict to 80%
			break
		}
		delete(s.entries, kv.key)
		s.totalSize -= kv.entry.Size
		os.Remove(s.keyPath(kv.key))
	}
}

func (s *Store) loadIndex() error {
	path := filepath.Join(s.dir, "index.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var entries []*Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range entries {
		s.entries[e.Key] = e
		s.totalSize += e.Size
	}

	return nil
}

func (s *Store) saveIndex() error {
	s.mu.RLock()
	entries := make([]*Entry, 0, len(s.entries))
	for _, e := range s.entries {
		entries = append(entries, e)
	}
	s.mu.RUnlock()

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, "index.json")
	return os.WriteFile(path, data, 0644)
}
