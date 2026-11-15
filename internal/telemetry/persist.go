package telemetry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// buildStamp pins the scheduling policy that produced one build.
type buildStamp struct {
	Scheduler string `json:"scheduler"`
	Params    string `json:"scheduler_params,omitempty"`
}

// snapshot is the on-disk shape of the store's retained state. Tasks
// replay through RecordTask on load, so caps and derived aggregates
// behave exactly as they would have in the original process.
type snapshot struct {
	Version   int                    `json:"version"`
	Scheduler string                 `json:"scheduler"`
	Params    string                 `json:"scheduler_params,omitempty"`
	Stamps    map[string]*buildStamp `json:"build_schedulers"`
	Tasks     []*TaskInfo            `json:"tasks"`
}

const snapshotVersion = 1

// markDirty flags the store for the next persistence write. Callers
// must hold the write lock.
func (s *Store) markDirty() {
	if s.historyPath != "" {
		s.dirty = true
	}
}

// EnablePersistence loads a previously saved history file (if any)
// and keeps the file updated roughly every 2 seconds while the store
// changes. It also raises retention so a multi-scheduler campaign of
// several hundred builds stays addressable. A load failure is
// returned; the caller decides whether to continue without history.
func (s *Store) EnablePersistence(ctx context.Context, path string) error {
	s.mu.Lock()
	// Raise retention before replaying so a large snapshot is not
	// truncated by the in-memory defaults (100 builds / 20k tasks).
	s.maxBuilds = 500
	s.maxTasksTotal = 200000
	s.historyPath = path
	s.mu.Unlock()

	if data, err := os.ReadFile(path); err == nil {
		var snap snapshot
		if err := json.Unmarshal(data, &snap); err != nil {
			s.disablePersistence()
			return fmt.Errorf("corrupt build history %s: %w", path, err)
		}
		if snap.Version != snapshotVersion {
			s.disablePersistence()
			return fmt.Errorf("unsupported build history version %d in %s", snap.Version, path)
		}
		// Deliberately do NOT restore snap.Scheduler as the current
		// label: the caller sets the live scheduler right after, and
		// replayed builds recover their own labels from snap.Stamps.
		for _, task := range snap.Tasks {
			s.RecordTask(task)
		}
		// Replay already stamped every build with the restored
		// scheduler; now restore per-build labels for builds recorded
		// under earlier coordinators, dropping stamps whose build did
		// not survive retention.
		s.mu.Lock()
		for id, stamp := range snap.Stamps {
			if _, ok := s.buildTasks[id]; ok {
				s.stamps[id] = *stamp
			}
		}
		s.mu.Unlock()
	} else if !os.IsNotExist(err) {
		s.disablePersistence()
		return fmt.Errorf("cannot read build history %s: %w", path, err)
	}

	s.persistWG.Add(1)
	go s.persistLoop(ctx)
	return nil
}

// WaitPersistLoop blocks until the persistence loop exits. Tests
// cancel the context then wait here so the final flush cannot race
// TempDir cleanup (Windows cannot delete open/recreated files).
func (s *Store) WaitPersistLoop() { s.persistWG.Wait() }

func (s *Store) disablePersistence() {
	s.mu.Lock()
	s.historyPath = ""
	s.mu.Unlock()
}

// persistLoop writes the snapshot when the store changed and flushes
// once more on shutdown.
func (s *Store) persistLoop(ctx context.Context) {
	defer s.persistWG.Done()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.flushOnce()
			return
		case <-ticker.C:
			s.flushOnce()
		}
	}
}

func (s *Store) flushOnce() {
	s.mu.Lock()
	dirty := s.dirty
	s.dirty = false
	path := s.historyPath
	s.mu.Unlock()
	if dirty {
		if err := s.writeSnapshot(path); err != nil {
			logPersistError(err)
		}
	}
}

func (s *Store) writeSnapshot(path string) error {
	s.mu.RLock()
	snap := snapshot{
		Version:   snapshotVersion,
		Scheduler: s.scheduler,
		Params:    s.schedulerParams,
		Stamps:    make(map[string]*buildStamp, len(s.stamps)),
		Tasks:     make([]*TaskInfo, 0, len(s.taskOrder)),
	}
	for id, stamp := range s.stamps {
		cp := stamp
		snap.Stamps[id] = &cp
	}
	for _, id := range s.taskOrder {
		if task, ok := s.tasks[id]; ok {
			snap.Tasks = append(snap.Tasks, task)
		}
	}
	s.mu.RUnlock()

	data, err := json.Marshal(&snap)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func logPersistError(err error) {
	fmt.Fprintf(os.Stderr, "build history: %v\n", err)
}
