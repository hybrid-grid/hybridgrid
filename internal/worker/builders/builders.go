// Package builders implements worker-side plugin builders: external
// hg-exec-<id> executables that implement the builderv1alpha1 exec
// contract (see docs/plugins.md). Discovery scans a directory, the
// handshake probes each binary's capability JSON, and Run executes one
// build task with the task JSON on stdin and the result JSON on
// stdout (console text on stderr).
package builders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// APIVersion is the builder plugin contract version. Breaking changes
// bump this string; workers refuse plugins that do not match exactly.
const APIVersion = "builderv1alpha1"

// capabilityTimeout bounds the `hg-exec-<id> capability` probe.
const capabilityTimeout = 5 * time.Second

// Capability is the JSON a plugin prints (one object, stdout) for the
// `capability` verb.
type Capability struct {
	ID          string `json:"id"`
	API         string `json:"api"`
	Description string `json:"description,omitempty"`
}

// Task is the JSON piped to the plugin's stdin for the `execute` verb.
type Task struct {
	TaskID    string            `json:"task_id"`
	BuildID   string            `json:"build_id,omitempty"`
	BuilderID string            `json:"builder_id"`
	WorkDir   string            `json:"work_dir"`
	Args      []string          `json:"args,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	TimeoutMs int64             `json:"timeout_ms,omitempty"`
}

// Result is the JSON the plugin writes to stdout when it exits 0.
// Outputs are WorkDir-relative paths to archive back as artifacts.
type Result struct {
	Success bool     `json:"success"`
	Outputs []string `json:"outputs,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Builder is one discovered plugin.
type Builder struct {
	ID       string
	Path     string // absolute path to the hg-exec-<id> binary
	Desc     string
	ProbedAt time.Time
}

// Registry holds the builders discovered on this worker.
type Registry struct {
	mu    sync.RWMutex
	byID  map[string]Builder
	order []string
}

// NewRegistry returns an empty registry. The binary hook (optional,
// tests only) overrides how a builder id resolves to an executable.
func NewRegistry() *Registry {
	return &Registry{byID: map[string]Builder{}}
}

// IDs returns the discovered builder ids in discovery order.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]string(nil), r.order...)
}

// Get looks a builder up by id.
func (r *Registry) Get(id string) (Builder, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.byID[id]
	return b, ok
}

// add registers a probed builder.
func (r *Registry) add(b Builder) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byID[b.ID]; !exists {
		r.order = append(r.order, b.ID)
	}
	r.byID[b.ID] = b
}

// Discover scans dir for executables named hg-exec-<id>, probes each
// with the capability handshake, and registers the ones that answer
// with a matching API version. Ids are validated to [a-z0-9-] so a
// plugin can never impersonate path-like or core names. Errors are
// per-plugin (skipped with a message), not fatal: one broken plugin
// must not take the worker's other builders down.
func (r *Registry) Discover(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var registered []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "hg-exec-") {
			continue
		}
		id := strings.TrimPrefix(name, "hg-exec-")
		if !validID(id) {
			continue
		}
		path := dir + string(os.PathSeparator) + name
		info, err := os.Stat(path)
		if err != nil || info.IsDir() || info.Mode()&0111 == 0 {
			continue
		}
		cap, err := probe(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "builders: skip %s: %v\n", name, err)
			continue
		}
		if cap.ID != id {
			fmt.Fprintf(os.Stderr, "builders: skip %s: capability id %q does not match filename\n", name, cap.ID)
			continue
		}
		r.add(Builder{ID: id, Path: path, Desc: cap.Description, ProbedAt: time.Now()})
		registered = append(registered, id)
	}
	return registered
}

// validID enforces the plugin id charset.
func validID(id string) bool {
	if id == "" {
		return false
	}
	for _, c := range id {
		switch {
		case c >= 'a' && c <= 'z', c >= '0' && c <= '9', c == '-':
		default:
			return false
		}
	}
	return true
}

// probe runs the capability handshake for one plugin binary.
func probe(path string) (Capability, error) {
	ctx, cancel := context.WithTimeout(context.Background(), capabilityTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, "capability")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return Capability{}, fmt.Errorf("capability probe failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var cap Capability
	if err := json.Unmarshal(stdout.Bytes(), &cap); err != nil {
		return Capability{}, fmt.Errorf("capability JSON parse failed: %w", err)
	}
	if cap.API != APIVersion {
		return Capability{}, fmt.Errorf("api %q does not match worker contract %q", cap.API, APIVersion)
	}
	if !validID(cap.ID) {
		return Capability{}, fmt.Errorf("invalid capability id %q", cap.ID)
	}
	return cap, nil
}

// Run executes one build task: the task JSON goes to the plugin's
// stdin for the `execute` verb; the plugin writes the result JSON to
// stdout and console text to stderr. A non-zero exit or an unparseable
// stdout is a task failure (the stderr tail becomes the error text).
// The caller owns the deadline via ctx.
func (b Builder) Run(ctx context.Context, task Task) (Result, string, error) {
	taskJSON, err := json.Marshal(task)
	if err != nil {
		return Result{}, "", fmt.Errorf("failed to encode task: %w", err)
	}

	cmd := exec.CommandContext(ctx, b.Path, "execute")
	cmd.Stdin = bytes.NewReader(taskJSON)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	console := stderr.String()

	if runErr != nil {
		return Result{Success: false, Error: runErr.Error()}, console, nil
	}

	var result Result
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		tail := console
		if len(tail) > 512 {
			tail = tail[len(tail)-512:]
		}
		return Result{Success: false, Error: fmt.Sprintf("invalid result JSON from builder: %v; console tail: %s", err, strings.TrimSpace(tail))}, console, nil
	}
	return result, console, nil
}
