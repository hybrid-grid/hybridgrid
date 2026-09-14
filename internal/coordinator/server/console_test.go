package server

import (
	"strings"
	"testing"
)

func TestConsoleStore_PutGet(t *testing.T) {
	store := newConsoleStore()
	store.Put("task-1", "hello", "warning: something")

	stdout, stderr, truncated, ok := store.Get("task-1")
	if !ok {
		t.Fatal("Get ok = false, want true after Put")
	}
	if stdout != "hello" {
		t.Errorf("stdout = %q, want hello", stdout)
	}
	if stderr != "warning: something" {
		t.Errorf("stderr = %q, want warning: something", stderr)
	}
	if truncated {
		t.Error("truncated = true, want false for small output")
	}
}

func TestConsoleStore_UnknownTask(t *testing.T) {
	store := newConsoleStore()
	if _, _, _, ok := store.Get("nope"); ok {
		t.Error("ok = true, want false for unknown task")
	}
}

func TestConsoleStore_EmptyTaskIDIgnored(t *testing.T) {
	store := newConsoleStore()
	store.Put("", "out", "")
	if _, _, _, ok := store.Get(""); ok {
		t.Error("ok = true, want false; empty ID must not be stored")
	}
}

func TestConsoleStore_TruncatesOversizedStreams(t *testing.T) {
	store := newConsoleStore()
	store.Put("big", strings.Repeat("a", maxConsoleBytesPerStream+100), strings.Repeat("b", maxConsoleBytesPerStream+50))

	stdout, stderr, truncated, ok := store.Get("big")
	if !ok {
		t.Fatal("Get ok = false, want true")
	}
	if !truncated {
		t.Error("truncated = false, want true when a stream exceeded the cap")
	}
	if len(stdout) != maxConsoleBytesPerStream {
		t.Errorf("len(stdout) = %d, want %d", len(stdout), maxConsoleBytesPerStream)
	}
	if len(stderr) != maxConsoleBytesPerStream {
		t.Errorf("len(stderr) = %d, want %d", len(stderr), maxConsoleBytesPerStream)
	}
}

func TestConsoleStore_EvictsOldestTask(t *testing.T) {
	store := newConsoleStore()
	store.maxTasks = 2

	store.Put("t1", "one", "")
	store.Put("t2", "two", "")
	store.Put("t3", "three", "")

	if _, _, _, ok := store.Get("t1"); ok {
		t.Error("t1 retained, want evicted (oldest falls out beyond the cap)")
	}
	for _, id := range []string{"t2", "t3"} {
		if _, _, _, ok := store.Get(id); !ok {
			t.Errorf("%s evicted, want retained", id)
		}
	}
	if len(store.order) != 2 {
		t.Errorf("order len = %d, want 2", len(store.order))
	}
}

func TestConsoleStore_OverwriteKeepsSingleEntry(t *testing.T) {
	store := newConsoleStore()
	store.Put("task-1", "first", "")
	store.Put("task-1", "second", "")

	if len(store.order) != 1 {
		t.Errorf("order len = %d, want 1 (re-Put must not duplicate)", len(store.order))
	}
	stdout, _, _, ok := store.Get("task-1")
	if !ok {
		t.Fatal("Get ok = false, want true")
	}
	if stdout != "second" {
		t.Errorf("stdout = %q, want second (last write wins)", stdout)
	}
}
