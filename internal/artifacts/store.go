// Package artifacts keeps build outputs on the coordinator's disk so the
// dashboard can list and serve them after a build finishes.
//
// Layout: <dir>/<build-id>/<name> plus <dir>/<build-id>/.manifest.json
// describing every artifact of that build. Names are reduced to a single
// safe path element, a per-file size cap bounds one upload, and a total
// cap evicts whole builds, oldest first.
package artifacts

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/h3nr1-d14z/hybridgrid/internal/security/validation"
)

const (
	manifestName = ".manifest.json"
	maxNameLen   = 200

	// Source values recorded on each artifact.
	SourceClient = "client"
	SourceWorker = "worker"
)

var (
	// ErrNotFound means the build or artifact is not in the store.
	ErrNotFound = errors.New("artifact not found")
	// ErrInvalidName means the artifact name is empty or not a plain file name.
	ErrInvalidName = errors.New("invalid artifact name")
	// ErrInvalidBuildID means the build id is empty or has unsafe characters.
	ErrInvalidBuildID = errors.New("invalid build id")
	// ErrTooLarge means an upload exceeded the per-file cap.
	ErrTooLarge = errors.New("artifact exceeds size limit")
)

// File is one entry inside an archive artifact (e.g. a Flutter APK inside
// the worker's tar.zst), listed for display only.
type File struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
}

// Artifact describes one stored file.
type Artifact struct {
	BuildID     string `json:"build_id"`
	Name        string `json:"name"`
	SizeBytes   int64  `json:"size_bytes"`
	SHA256      string `json:"sha256"`
	CreatedAtMs int64  `json:"created_at_ms"`
	Source      string `json:"source"`
	Files       []File `json:"files,omitempty"`
}

// Config bounds the store.
type Config struct {
	Dir           string
	MaxFileBytes  int64 // <= 0: unlimited
	MaxTotalBytes int64 // <= 0: unlimited
}

// Store is safe for concurrent use.
type Store struct {
	cfg Config
	mu  sync.Mutex
	now func() time.Time
}

// New creates the store directory if needed.
func New(cfg Config) (*Store, error) {
	if cfg.Dir == "" {
		return nil, errors.New("artifact dir required")
	}
	if err := os.MkdirAll(cfg.Dir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create artifact dir: %w", err)
	}
	return &Store{cfg: cfg, now: time.Now}, nil
}

// Dir returns the root directory.
func (s *Store) Dir() string { return s.cfg.Dir }

// SanitizeName reduces a client-supplied name to one safe path element:
// the base name, without control characters, never "." or "..".
func SanitizeName(name string) (string, error) {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || name == "/" || name == manifestName || len(name) > maxNameLen {
		return "", ErrInvalidName
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '/' {
			return "", ErrInvalidName
		}
	}
	return name, nil
}

func (s *Store) buildDir(buildID string) (string, error) {
	if validation.NormalizeBuildID(buildID) == "" {
		return "", ErrInvalidBuildID
	}
	return filepath.Join(s.cfg.Dir, buildID), nil
}

// Put stores r as buildID/name, replacing an artifact of the same name.
// The upload is written to a temp file and renamed into place, so a
// failed or oversized upload never leaves a partial artifact behind.
func (s *Store) Put(buildID, name, source string, r io.Reader, files []File) (*Artifact, error) {
	dir, err := s.buildDir(buildID)
	if err != nil {
		return nil, err
	}
	name, err = SanitizeName(name)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return nil, fmt.Errorf("failed to create build dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".upload-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename

	hash := sha256.New()
	src := r
	if s.cfg.MaxFileBytes > 0 {
		src = io.LimitReader(r, s.cfg.MaxFileBytes+1)
	}
	written, err := io.Copy(io.MultiWriter(tmp, hash), src)
	closeErr := tmp.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to write artifact: %w", err)
	}
	if closeErr != nil {
		return nil, fmt.Errorf("failed to close artifact: %w", closeErr)
	}
	if s.cfg.MaxFileBytes > 0 && written > s.cfg.MaxFileBytes {
		return nil, ErrTooLarge
	}

	art := &Artifact{
		BuildID:     buildID,
		Name:        name,
		SizeBytes:   written,
		SHA256:      hex.EncodeToString(hash.Sum(nil)),
		CreatedAtMs: s.now().UnixMilli(),
		Source:      source,
		Files:       files,
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := os.Rename(tmp.Name(), filepath.Join(dir, name)); err != nil {
		return nil, fmt.Errorf("failed to store artifact: %w", err)
	}
	manifest, _ := s.readManifest(dir)
	kept := manifest[:0]
	for _, a := range manifest {
		if a.Name != name {
			kept = append(kept, a)
		}
	}
	if err := s.writeManifest(dir, append(kept, *art)); err != nil {
		return nil, err
	}
	s.evictLocked(buildID)
	return art, nil
}

// List returns a build's artifacts, newest first. An unknown build has none.
func (s *Store) List(buildID string) ([]Artifact, error) {
	dir, err := s.buildDir(buildID)
	if err != nil {
		return nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.readManifest(dir)
	if err != nil {
		return nil, err
	}
	sort.Slice(list, func(i, j int) bool { return list[i].CreatedAtMs > list[j].CreatedAtMs })
	return list, nil
}

// Open returns the artifact's file for reading; the caller closes it.
func (s *Store) Open(buildID, name string) (*os.File, *Artifact, error) {
	dir, err := s.buildDir(buildID)
	if err != nil {
		return nil, nil, err
	}
	name, err = SanitizeName(name)
	if err != nil {
		return nil, nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	list, err := s.readManifest(dir)
	if err != nil {
		return nil, nil, err
	}
	for i := range list {
		if list[i].Name == name {
			f, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					return nil, nil, ErrNotFound
				}
				return nil, nil, fmt.Errorf("failed to open artifact: %w", err)
			}
			return f, &list[i], nil
		}
	}
	return nil, nil, ErrNotFound
}

func (s *Store) readManifest(dir string) ([]Artifact, error) {
	data, err := os.ReadFile(filepath.Join(dir, manifestName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}
	var list []Artifact
	if err := json.Unmarshal(data, &list); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}
	return list, nil
}

func (s *Store) writeManifest(dir string, list []Artifact) error {
	data, err := json.Marshal(list)
	if err != nil {
		return fmt.Errorf("failed to encode manifest: %w", err)
	}
	tmp := filepath.Join(dir, manifestName+".tmp")
	if err := os.WriteFile(tmp, data, 0o640); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}
	if err := os.Rename(tmp, filepath.Join(dir, manifestName)); err != nil {
		return fmt.Errorf("failed to replace manifest: %w", err)
	}
	return nil
}

// evictLocked removes whole builds, least recently updated first, until
// the store fits MaxTotalBytes. The build just written is never evicted,
// so one oversized build still keeps its newest outputs.
func (s *Store) evictLocked(keep string) {
	if s.cfg.MaxTotalBytes <= 0 {
		return
	}
	entries, err := os.ReadDir(s.cfg.Dir)
	if err != nil {
		return
	}
	type usage struct {
		id     string
		bytes  int64
		latest int64
	}
	var builds []usage
	var total int64
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		list, err := s.readManifest(filepath.Join(s.cfg.Dir, e.Name()))
		if err != nil {
			continue
		}
		u := usage{id: e.Name()}
		for _, a := range list {
			u.bytes += a.SizeBytes
			if a.CreatedAtMs > u.latest {
				u.latest = a.CreatedAtMs
			}
		}
		total += u.bytes
		builds = append(builds, u)
	}
	sort.Slice(builds, func(i, j int) bool { return builds[i].latest < builds[j].latest })
	for _, b := range builds {
		if total <= s.cfg.MaxTotalBytes {
			return
		}
		if b.id == keep {
			continue
		}
		if err := os.RemoveAll(filepath.Join(s.cfg.Dir, b.id)); err == nil {
			total -= b.bytes
		}
	}
}
