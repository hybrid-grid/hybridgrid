package telemetry

import "sync"

// Store retains the latest state of every observed task, derives
// logical build aggregates from it, and keeps a bounded ring of
// recent task events for late-joining consumers. Retention is
// bounded on three axes: builds (maxBuilds), tasks per build
// (maxTasksPerBuild) and tasks in total (maxTasksTotal); eviction on
// any axis marks the affected build truncated so consumers never
// mistake a shrunken retained set for a complete one.
type Store struct {
	mu sync.RWMutex

	tasks          map[string]*TaskInfo
	taskOrder      []string
	buildTasks     map[string][]string
	buildOrder     []string
	buildTruncated map[string]bool
	recentEvents   []*TaskInfo

	maxBuilds        int
	maxTasksTotal    int
	maxTasksPerBuild int
	maxEvents        int
}

// NewStore creates a store with the retention limits the embedded
// dashboard has always used.
func NewStore() *Store {
	return &Store{
		tasks:            make(map[string]*TaskInfo),
		buildTasks:       make(map[string][]string),
		buildTruncated:   make(map[string]bool),
		recentEvents:     make([]*TaskInfo, 0, 100),
		maxBuilds:        100,
		maxTasksTotal:    20000,
		maxTasksPerBuild: 5000,
		maxEvents:        100,
	}
}

// RecordTask records the latest state of one task and updates build
// membership. Called from the coordinator's task event hooks.
func (s *Store) RecordTask(task *TaskInfo) {
	if task == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.recentEvents = append(s.recentEvents, task)
	if len(s.recentEvents) > s.maxEvents {
		s.recentEvents = s.recentEvents[len(s.recentEvents)-s.maxEvents:]
	}

	updated := *task
	previous, exists := s.tasks[task.ID]
	if !exists {
		s.taskOrder = append(s.taskOrder, task.ID)
	} else if previous.BuildID != task.BuildID {
		s.removeTaskFromBuild(previous.BuildID, task.ID)
	}
	s.tasks[task.ID] = &updated

	if task.BuildID != "" {
		s.ensureBuild(task.BuildID)
		if ids := s.buildTasks[task.BuildID]; len(ids) < s.maxTasksPerBuild && !containsString(ids, task.ID) {
			s.buildTasks[task.BuildID] = append(ids, task.ID)
		}
		if len(s.buildTasks[task.BuildID]) >= s.maxTasksPerBuild {
			s.buildTruncated[task.BuildID] = true
		}
	}

	for len(s.tasks) > s.maxTasksTotal {
		s.evictOldestTask()
	}
}

// Tasks returns the latest state of each task, newest first.
func (s *Store) Tasks() []*TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*TaskInfo, 0, len(s.taskOrder))
	for i := len(s.taskOrder) - 1; i >= 0; i-- {
		task, ok := s.tasks[s.taskOrder[i]]
		if !ok {
			continue
		}
		cp := *task
		tasks = append(tasks, &cp)
	}
	return tasks
}

// RecentTasks returns the bounded event ring, oldest first.
func (s *Store) RecentTasks() []*TaskInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]*TaskInfo, 0, len(s.recentEvents))
	for _, task := range s.recentEvents {
		cp := *task
		out = append(out, &cp)
	}
	return out
}

// Builds returns derived logical build aggregates, newest first.
func (s *Store) Builds() []*BuildInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	builds := make([]*BuildInfo, 0, len(s.buildOrder))
	for i := len(s.buildOrder) - 1; i >= 0; i-- {
		builds = append(builds, s.buildInfo(s.buildOrder[i]))
	}
	return builds
}

// BuildDetail returns a logical build and its retained tasks, newest
// first. The boolean reports whether the build is known.
func (s *Store) BuildDetail(buildID string) (*BuildInfo, []*TaskInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.buildTasks[buildID]; !ok {
		return nil, nil, false
	}
	tasks := make([]*TaskInfo, 0, len(s.buildTasks[buildID]))
	for i := len(s.buildTasks[buildID]) - 1; i >= 0; i-- {
		task, ok := s.tasks[s.buildTasks[buildID][i]]
		if !ok {
			continue
		}
		cp := *task
		tasks = append(tasks, &cp)
	}
	return s.buildInfo(buildID), tasks, true
}

// buildInfo derives a logical build aggregate. The caller must hold
// the read lock.
func (s *Store) buildInfo(buildID string) *BuildInfo {
	build := &BuildInfo{ID: buildID, Status: "completed", Truncated: s.buildTruncated[buildID]}
	for _, task := range s.tasks {
		if task.BuildID != buildID {
			continue
		}
		if build.BuildType == "" {
			build.BuildType = task.BuildType
		}
		build.TotalTasks++
		switch task.Status {
		case "running":
			build.RunningTasks++
		case "failed":
			build.FailedTasks++
		default:
			build.CompletedTasks++
		}
		if task.FromCache {
			build.FromCacheCount++
		}
		if task.StartedAtMs != 0 && (build.FirstTaskAtMs == 0 || task.StartedAtMs < build.FirstTaskAtMs) {
			build.FirstTaskAtMs = task.StartedAtMs
		}
		lastAt := task.CompletedAtMs
		if lastAt == 0 {
			lastAt = task.StartedAtMs
		}
		if lastAt > build.LastTaskAtMs {
			build.LastTaskAtMs = lastAt
		}
	}
	if build.RunningTasks > 0 {
		build.Status = "running"
	} else if build.FailedTasks > 0 {
		build.Status = "failed"
	}
	return build
}

func (s *Store) ensureBuild(buildID string) {
	if _, exists := s.buildTasks[buildID]; exists {
		return
	}
	if s.maxBuilds <= 0 {
		return
	}
	for len(s.buildOrder) >= s.maxBuilds {
		s.evictOldestBuild()
	}
	s.buildTasks[buildID] = make([]string, 0)
	s.buildOrder = append(s.buildOrder, buildID)
}

func (s *Store) evictOldestBuild() {
	if len(s.buildOrder) == 0 {
		return
	}
	buildID := s.buildOrder[0]
	s.buildOrder = s.buildOrder[1:]
	delete(s.buildTasks, buildID)
	delete(s.buildTruncated, buildID)
	for id, task := range s.tasks {
		if task.BuildID == buildID {
			delete(s.tasks, id)
		}
	}
	s.removeMissingTasksFromOrder()
}

func (s *Store) evictOldestTask() {
	for len(s.taskOrder) > 0 {
		id := s.taskOrder[0]
		s.taskOrder = s.taskOrder[1:]
		if task, ok := s.tasks[id]; ok {
			delete(s.tasks, id)
			s.removeTaskFromBuild(task.BuildID, id)
			// The build this task belonged to now holds a shrunken
			// subset of its real task set; mark it so Builds reports
			// the counts as truncated rather than complete
			// (total-task eviction can otherwise drain an old build
			// to a phantom 0/0 "completed" row).
			if task.BuildID != "" {
				s.buildTruncated[task.BuildID] = true
			}
			return
		}
	}
}

func (s *Store) removeMissingTasksFromOrder() {
	order := s.taskOrder[:0]
	for _, id := range s.taskOrder {
		if _, ok := s.tasks[id]; ok {
			order = append(order, id)
		}
	}
	s.taskOrder = order
}

func (s *Store) removeTaskFromBuild(buildID, taskID string) {
	if buildID == "" {
		return
	}
	ids := s.buildTasks[buildID]
	for i, id := range ids {
		if id == taskID {
			s.buildTasks[buildID] = append(ids[:i], ids[i+1:]...)
			return
		}
	}
}

func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
