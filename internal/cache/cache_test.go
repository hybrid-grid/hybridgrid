package cache

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestKeyBuilder(t *testing.T) {
	kb := NewKeyBuilder()
	kb.AddString("gcc")
	kb.AddString("12.0")
	key1 := kb.Sum()

	kb.Reset()
	kb.AddString("gcc")
	kb.AddString("12.0")
	key2 := kb.Sum()

	if key1 != key2 {
		t.Error("Same inputs should produce same hash")
	}

	kb.Reset()
	kb.AddString("gcc")
	kb.AddString("13.0")
	key3 := kb.Sum()

	if key1 == key3 {
		t.Error("Different inputs should produce different hash")
	}
}

func TestHashFile(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())

	f.WriteString("test content")
	f.Close()

	hash1, err := HashFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	hash2, err := HashFile(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	if hash1 != hash2 {
		t.Error("Same file should produce same hash")
	}

	if len(hash1) != 16 { // 64-bit hash = 16 hex chars
		t.Errorf("Expected 16 char hash, got %d", len(hash1))
	}
}

func TestCompilationKey(t *testing.T) {
	ck := &CompilationKey{
		Compiler:    "gcc",
		CompilerVer: "12.0",
		TargetArch:  "x86_64",
		Flags:       []string{"-O2", "-Wall"},
		Defines:     []string{"DEBUG", "VERSION=1"},
		SourceHash:  "abc123",
	}

	key1 := ck.Build()
	key2 := ck.Build()

	if key1 != key2 {
		t.Error("Same compilation key should produce same hash")
	}

	// Different flags order should still produce same key (sorted)
	ck2 := &CompilationKey{
		Compiler:    "gcc",
		CompilerVer: "12.0",
		TargetArch:  "x86_64",
		Flags:       []string{"-Wall", "-O2"},
		Defines:     []string{"VERSION=1", "DEBUG"},
		SourceHash:  "abc123",
	}

	if ck.Build() != ck2.Build() {
		t.Error("Same flags in different order should produce same key")
	}
}

func TestStore(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24) // 10MB, 24 hours
	if err != nil {
		t.Fatal(err)
	}

	// Test Put/Get
	err = store.PutBytes("key1", []byte("value1"))
	if err != nil {
		t.Fatal(err)
	}

	data, ok := store.GetBytes("key1")
	if !ok {
		t.Error("Expected to find key1")
	}
	if string(data) != "value1" {
		t.Errorf("Expected 'value1', got '%s'", string(data))
	}

	// Test Stats
	stats := store.Stats()
	if stats.Entries != 1 {
		t.Errorf("Expected 1 entry, got %d", stats.Entries)
	}

	// Test Delete
	store.Delete("key1")
	_, ok = store.GetBytes("key1")
	if ok {
		t.Error("Expected key1 to be deleted")
	}

	// Test Clear
	store.PutBytes("key2", []byte("value2"))
	store.PutBytes("key3", []byte("value3"))
	store.Clear()
	stats = store.Stats()
	if stats.Entries != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", stats.Entries)
	}
}

func TestStoreEviction(t *testing.T) {
	dir := t.TempDir()
	// 1KB max size
	store, err := NewStore(dir, 0, 24)
	if err != nil {
		t.Fatal(err)
	}
	store.maxSize = 1024 // 1KB

	// Add entries that exceed max size
	for i := 0; i < 10; i++ {
		data := make([]byte, 200) // 200 bytes each
		store.PutBytes(string(rune('a'+i)), data)
	}

	// Should have evicted some entries
	stats := store.Stats()
	if stats.TotalSize > store.maxSize {
		t.Errorf("Total size %d exceeds max %d after eviction", stats.TotalSize, store.maxSize)
	}
}

func TestStorePersistence(t *testing.T) {
	dir := t.TempDir()

	// Create store and add data
	store1, _ := NewStore(dir, 10, 24)
	store1.PutBytes("persist-key", []byte("persist-value"))

	// Create new store with same directory
	store2, _ := NewStore(dir, 10, 24)
	data, ok := store2.GetBytes("persist-key")
	if !ok {
		t.Error("Expected persisted key to be found")
	}
	if string(data) != "persist-value" {
		t.Errorf("Expected 'persist-value', got '%s'", string(data))
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"./foo/bar.c", "foo/bar.c"},
		{"foo\\bar.c", "foo/bar.c"},
		{"./foo\\bar.c", "foo/bar.c"},
		{"foo/bar.c", "foo/bar.c"},
	}

	for _, tt := range tests {
		got := NormalizePath(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizePath(%s) = %s, want %s", tt.input, got, tt.expected)
		}
	}
}

func TestKeyPath(t *testing.T) {
	dir := t.TempDir()
	store, _ := NewStore(dir, 10, 24)

	// Keys with 2+ chars should use subdirectory
	path := store.keyPath("abcdef123")
	expected := filepath.Join(dir, "ab", "abcdef123")
	if path != expected {
		t.Errorf("keyPath = %s, want %s", path, expected)
	}

	// Short keys should not use subdirectory
	path = store.keyPath("a")
	expected = filepath.Join(dir, "a")
	if path != expected {
		t.Errorf("keyPath = %s, want %s", path, expected)
	}
}

func TestHashBytes(t *testing.T) {
	hash1 := HashBytes([]byte("test data"))
	hash2 := HashBytes([]byte("test data"))
	hash3 := HashBytes([]byte("different data"))

	if hash1 != hash2 {
		t.Error("Same data should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("Different data should produce different hash")
	}
	if len(hash1) != 16 {
		t.Errorf("Expected 16 char hash, got %d", len(hash1))
	}
}

func TestHashString(t *testing.T) {
	hash1 := HashString("test string")
	hash2 := HashString("test string")
	hash3 := HashString("different string")

	if hash1 != hash2 {
		t.Error("Same string should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("Different strings should produce different hash")
	}
}

func TestHashStrings(t *testing.T) {
	hash1 := HashStrings("a", "b", "c")
	hash2 := HashStrings("a", "b", "c")
	hash3 := HashStrings("a", "c", "b") // different order

	if hash1 != hash2 {
		t.Error("Same strings should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("Different order should produce different hash")
	}
}

func TestKeyBuilder_AddBytes(t *testing.T) {
	kb := NewKeyBuilder()
	kb.AddBytes([]byte{1, 2, 3})
	key1 := kb.Sum()

	kb.Reset()
	kb.AddBytes([]byte{1, 2, 3})
	key2 := kb.Sum()

	if key1 != key2 {
		t.Error("Same bytes should produce same hash")
	}

	kb.Reset()
	kb.AddBytes([]byte{3, 2, 1})
	key3 := kb.Sum()

	if key1 == key3 {
		t.Error("Different bytes should produce different hash")
	}
}

func TestKeyBuilder_AddStrings(t *testing.T) {
	kb := NewKeyBuilder()
	kb.AddStrings([]string{"a", "b", "c"})
	key1 := kb.Sum()

	kb.Reset()
	kb.AddStrings([]string{"a", "b", "c"})
	key2 := kb.Sum()

	if key1 != key2 {
		t.Error("Same strings should produce same hash")
	}
}

func TestKeyBuilder_AddSortedStrings(t *testing.T) {
	kb := NewKeyBuilder()
	kb.AddSortedStrings([]string{"c", "a", "b"})
	key1 := kb.Sum()

	kb.Reset()
	kb.AddSortedStrings([]string{"a", "b", "c"})
	key2 := kb.Sum()

	// Sorted strings should produce same hash regardless of input order
	if key1 != key2 {
		t.Error("Sorted strings should produce same hash regardless of input order")
	}
}

func TestKeyBuilder_AddFile(t *testing.T) {
	// Create temp file
	f, err := os.CreateTemp("", "kb-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("file content")
	f.Close()

	kb := NewKeyBuilder()
	if err := kb.AddFile(f.Name()); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}
	key1 := kb.Sum()

	kb.Reset()
	if err := kb.AddFile(f.Name()); err != nil {
		t.Fatalf("AddFile failed: %v", err)
	}
	key2 := kb.Sum()

	if key1 != key2 {
		t.Error("Same file should produce same hash")
	}

	// Test non-existent file
	kb.Reset()
	err = kb.AddFile("/nonexistent/file")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestKeyBuilder_SumUint64(t *testing.T) {
	kb := NewKeyBuilder()
	kb.AddString("test")
	sum := kb.SumUint64()

	kb.Reset()
	kb.AddString("test")
	sum2 := kb.SumUint64()

	if sum != sum2 {
		t.Error("Same input should produce same uint64 sum")
	}

	if sum == 0 {
		t.Error("Sum should not be 0")
	}
}

func TestStore_GetWithReader(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Put data
	store.PutBytes("reader-key", []byte("reader content"))

	// Get using Get() which returns io.ReadCloser
	rc, ok := store.Get("reader-key")
	if !ok {
		t.Fatal("Expected to find reader-key")
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if string(data) != "reader content" {
		t.Errorf("Expected 'reader content', got '%s'", string(data))
	}
}

func TestStore_PutWithReader(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Create a temp file to use as reader
	f, err := os.CreateTemp("", "put-reader-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("put reader content")
	f.Close()

	// Open for reading
	rf, err := os.Open(f.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer rf.Close()

	// Put using io.Reader
	err = store.Put("put-reader-key", rf)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	// Verify
	data, ok := store.GetBytes("put-reader-key")
	if !ok {
		t.Fatal("Expected to find put-reader-key")
	}
	if string(data) != "put reader content" {
		t.Errorf("Expected 'put reader content', got '%s'", string(data))
	}
}

func TestStore_TTLExpiration(t *testing.T) {
	dir := t.TempDir()
	// Create store with very short TTL (1 hour)
	store, err := NewStore(dir, 10, 1)
	if err != nil {
		t.Fatal(err)
	}

	// Add entry
	store.PutBytes("ttl-key", []byte("ttl-value"))

	// Manually set entry's created time to be expired
	store.mu.Lock()
	if entry, ok := store.entries["ttl-key"]; ok {
		entry.CreatedAt = entry.CreatedAt.Add(-2 * store.ttl) // 2 hours ago
	}
	store.mu.Unlock()

	// Get should return false because entry is expired
	_, ok := store.GetBytes("ttl-key")
	if ok {
		t.Error("Expected expired entry to not be found")
	}

	// Entry should be deleted
	store.mu.RLock()
	_, exists := store.entries["ttl-key"]
	store.mu.RUnlock()
	if exists {
		t.Error("Expected expired entry to be deleted from entries map")
	}
}

func TestStore_GetMissingFile(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Add entry to map but don't create file
	store.mu.Lock()
	store.entries["ghost-key"] = &Entry{
		Key:  "ghost-key",
		Size: 100,
	}
	store.mu.Unlock()

	// Get should return false because file doesn't exist
	_, ok := store.GetBytes("ghost-key")
	if ok {
		t.Error("Expected missing file to return false")
	}
}

func TestStore_StatsWithHits(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	store.PutBytes("hit-key", []byte("hit-value"))

	// Access multiple times
	for i := 0; i < 5; i++ {
		store.GetBytes("hit-key")
	}

	stats := store.Stats()
	if stats.TotalHits != 5 {
		t.Errorf("Expected 5 hits, got %d", stats.TotalHits)
	}
}

func TestStore_UpdateExistingEntry(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Put initial value
	store.PutBytes("update-key", []byte("initial"))
	stats1 := store.Stats()

	// Update with new value
	store.PutBytes("update-key", []byte("updated value longer"))
	stats2 := store.Stats()

	// Should still have 1 entry
	if stats2.Entries != 1 {
		t.Errorf("Expected 1 entry, got %d", stats2.Entries)
	}

	// Size should be updated
	if stats2.TotalSize == stats1.TotalSize {
		t.Error("Total size should change after update")
	}

	// Value should be updated
	data, ok := store.GetBytes("update-key")
	if !ok {
		t.Fatal("Expected to find update-key")
	}
	if string(data) != "updated value longer" {
		t.Errorf("Expected 'updated value longer', got '%s'", string(data))
	}
}

func TestHashFile_NonExistent(t *testing.T) {
	_, err := HashFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestCompilationKey_DifferentInputs(t *testing.T) {
	base := &CompilationKey{
		Compiler:    "gcc",
		CompilerVer: "12.0",
		TargetArch:  "x86_64",
		Flags:       []string{"-O2"},
		Defines:     []string{"DEBUG"},
		SourceHash:  "abc123",
	}

	tests := []struct {
		name   string
		modify func(*CompilationKey)
	}{
		{"different compiler", func(c *CompilationKey) { c.Compiler = "clang" }},
		{"different version", func(c *CompilationKey) { c.CompilerVer = "13.0" }},
		{"different arch", func(c *CompilationKey) { c.TargetArch = "arm64" }},
		{"different flags", func(c *CompilationKey) { c.Flags = []string{"-O3"} }},
		{"different defines", func(c *CompilationKey) { c.Defines = []string{"RELEASE"} }},
		{"different source", func(c *CompilationKey) { c.SourceHash = "def456" }},
	}

	baseKey := base.Build()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modified := &CompilationKey{
				Compiler:    base.Compiler,
				CompilerVer: base.CompilerVer,
				TargetArch:  base.TargetArch,
				Flags:       append([]string{}, base.Flags...),
				Defines:     append([]string{}, base.Defines...),
				SourceHash:  base.SourceHash,
			}
			tt.modify(modified)

			if modified.Build() == baseKey {
				t.Error("Different inputs should produce different key")
			}
		})
	}
}

func TestNewStore_CreateDirError(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping invalid path test on Windows (path validation differs)")
	}
	// Try to create store in a path that can't be created (Unix-specific)
	_, err := NewStore("/dev/null/impossible", 10, 24)
	if err == nil {
		t.Error("Expected error when creating store in invalid path")
	}
}

func TestStore_DeleteNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Delete non-existent key should not error
	err = store.Delete("nonexistent")
	if err != nil {
		t.Errorf("Delete non-existent key should not error: %v", err)
	}
}

func TestStore_ClearEmpty(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	// Clear empty store should not error
	err = store.Clear()
	if err != nil {
		t.Errorf("Clear empty store should not error: %v", err)
	}
}

func TestStore_GetNonExistent(t *testing.T) {
	dir := t.TempDir()
	store, err := NewStore(dir, 10, 24)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := store.Get("nonexistent")
	if ok {
		t.Error("Get non-existent key should return false")
	}

	_, ok = store.GetBytes("nonexistent")
	if ok {
		t.Error("GetBytes non-existent key should return false")
	}
}
