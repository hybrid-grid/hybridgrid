package cache

import (
	"encoding/hex"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/cespare/xxhash/v2"
)

// KeyBuilder builds cache keys from various inputs.
type KeyBuilder struct {
	hasher *xxhash.Digest
}

// NewKeyBuilder creates a new cache key builder.
func NewKeyBuilder() *KeyBuilder {
	return &KeyBuilder{
		hasher: xxhash.New(),
	}
}

// Reset resets the key builder for reuse.
func (k *KeyBuilder) Reset() {
	k.hasher.Reset()
}

// AddString adds a string to the hash.
func (k *KeyBuilder) AddString(s string) *KeyBuilder {
	k.hasher.WriteString(s)
	k.hasher.WriteString("\x00") // null separator
	return k
}

// AddStrings adds multiple strings to the hash.
func (k *KeyBuilder) AddStrings(ss []string) *KeyBuilder {
	for _, s := range ss {
		k.AddString(s)
	}
	return k
}

// AddSortedStrings adds strings in sorted order (for deterministic hashing).
func (k *KeyBuilder) AddSortedStrings(ss []string) *KeyBuilder {
	sorted := make([]string, len(ss))
	copy(sorted, ss)
	sort.Strings(sorted)
	return k.AddStrings(sorted)
}

// AddBytes adds bytes to the hash.
func (k *KeyBuilder) AddBytes(b []byte) *KeyBuilder {
	k.hasher.Write(b)
	return k
}

// AddFile adds a file's contents to the hash.
func (k *KeyBuilder) AddFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(k.hasher, f)
	return err
}

// Sum returns the final hash as a hex string.
func (k *KeyBuilder) Sum() string {
	sum := k.hasher.Sum64()
	buf := make([]byte, 8)
	buf[0] = byte(sum >> 56)
	buf[1] = byte(sum >> 48)
	buf[2] = byte(sum >> 40)
	buf[3] = byte(sum >> 32)
	buf[4] = byte(sum >> 24)
	buf[5] = byte(sum >> 16)
	buf[6] = byte(sum >> 8)
	buf[7] = byte(sum)
	return hex.EncodeToString(buf)
}

// SumUint64 returns the final hash as uint64.
func (k *KeyBuilder) SumUint64() uint64 {
	return k.hasher.Sum64()
}

// CompilationKey generates a cache key for a compilation task.
type CompilationKey struct {
	Compiler    string
	CompilerVer string
	TargetArch  string
	Flags       []string
	Defines     []string
	SourceHash  string
}

// Build generates the cache key.
func (c *CompilationKey) Build() string {
	kb := NewKeyBuilder()

	kb.AddString(c.Compiler)
	kb.AddString(c.CompilerVer)
	kb.AddString(c.TargetArch)
	kb.AddSortedStrings(c.Flags)
	kb.AddSortedStrings(c.Defines)
	kb.AddString(c.SourceHash)

	return kb.Sum()
}

// HashFile computes xxhash of a file.
func HashFile(path string) (string, error) {
	kb := NewKeyBuilder()
	if err := kb.AddFile(path); err != nil {
		return "", err
	}
	return kb.Sum(), nil
}

// HashBytes computes xxhash of bytes.
func HashBytes(data []byte) string {
	kb := NewKeyBuilder()
	kb.AddBytes(data)
	return kb.Sum()
}

// HashString computes xxhash of a string.
func HashString(s string) string {
	kb := NewKeyBuilder()
	kb.AddString(s)
	return kb.Sum()
}

// HashStrings computes xxhash of multiple strings.
func HashStrings(ss ...string) string {
	kb := NewKeyBuilder()
	kb.AddStrings(ss)
	return kb.Sum()
}

// NormalizePath normalizes a file path for consistent hashing.
func NormalizePath(path string) string {
	// Remove leading ./
	path = strings.TrimPrefix(path, "./")
	// Normalize separators
	path = strings.ReplaceAll(path, "\\", "/")
	return path
}
