package server

import (
	"sync"
	"time"
)

// Console retention bounds. Output arrives once, at task completion
// (Compile/Build are unary), so the retained blob is the whole log.
const (
	// maxConsoleBytesPerStream caps a single stdout or stderr capture.
	// Compiler output that exceeds it is truncated at the head-worst
	// boundary; the dashboard shows a truncation marker.
	maxConsoleBytesPerStream = 256 * 1024

	// maxConsoleTasks bounds how many tasks keep console output. Oldest
	// tasks are evicted whole, mirroring the dashboard hub's retention.
	maxConsoleTasks = 1000
)

// consoleLog is the retained output of one completed task.
type consoleLog struct {
	Stdout     string
	Stderr     string
	Truncated  bool
	ReceivedAt int64 // unix nanoseconds, for oldest-first eviction
}

// consoleStore retains per-task console output for the dashboard,
// bounded in count and per-stream size. The zero value is not usable;
// use newConsoleStore.
type consoleStore struct {
	mu       sync.Mutex
	logs     map[string]*consoleLog
	order    []string // insertion order, oldest first
	maxTasks int      // task cap, overridable in tests
}

func newConsoleStore() *consoleStore {
	return &consoleStore{logs: make(map[string]*consoleLog), maxTasks: maxConsoleTasks}
}

// Put retains the output of one task. Empty output is stored as an
// entry (the task ran and produced nothing — distinct from unknown).
func (c *consoleStore) Put(taskID, stdout, stderr string) {
	if taskID == "" {
		return
	}
	truncated := false
	if len(stdout) > maxConsoleBytesPerStream {
		stdout = stdout[:maxConsoleBytesPerStream]
		truncated = true
	}
	if len(stderr) > maxConsoleBytesPerStream {
		stderr = stderr[:maxConsoleBytesPerStream]
		truncated = true
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, exists := c.logs[taskID]; !exists {
		c.order = append(c.order, taskID)
	}
	c.logs[taskID] = &consoleLog{
		Stdout:     stdout,
		Stderr:     stderr,
		Truncated:  truncated,
		ReceivedAt: time.Now().UnixNano(),
	}
	for len(c.order) > c.maxTasks {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.logs, oldest)
	}
}

// Get returns the retained output for a task and whether it exists.
func (c *consoleStore) Get(taskID string) (stdout, stderr string, truncated bool, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	log, exists := c.logs[taskID]
	if !exists {
		return "", "", false, false
	}
	return log.Stdout, log.Stderr, log.Truncated, true
}
