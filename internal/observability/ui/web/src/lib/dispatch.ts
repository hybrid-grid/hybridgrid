import { normalizeStatus } from './status'
import type { BuildInfo, TaskInfo, WorkerInfo, WsMessage } from './types'

/** Completed tasks shown per worker lane, newest first. */
export const RECENT_PER_LANE = 6

function isSettled(task: TaskInfo): boolean {
  return Boolean(task.completed_at_ms) || normalizeStatus(task.status, task.exit_code) !== 'running'
}

/**
 * Overlays live WebSocket task events (for one build, or for every build
 * when buildId is omitted) onto the polled task list, so a dispatch shows up the moment it happens instead of on the
 * next poll. An event never downgrades a task: a late "started" frame
 * for a task already known to be finished is ignored.
 */
export function mergeLiveTasks(tasks: TaskInfo[], messages: WsMessage[], buildId?: string): TaskInfo[] {
  const byId = new Map(tasks.map((task) => [task.id, task]))
  // Oldest first so a task's completion lands after its start.
  for (const message of [...messages].reverse()) {
    if (message.type !== 'task_started' && message.type !== 'task_completed') continue
    const live = message.data as TaskInfo | undefined
    if (!live?.id || (buildId !== undefined && live.build_id !== buildId)) continue
    const known = byId.get(live.id)
    if (known && isSettled(known) && !isSettled(live)) continue
    byId.set(live.id, { ...known, ...live })
  }
  return [...byId.values()]
}

export interface DispatchLane {
  workerId: string
  worker?: WorkerInfo
  /** Tasks of this build compiling on the worker right now, oldest first. */
  running: TaskInfo[]
  /** Finished tasks of this build on the worker, newest first, capped. */
  recent: TaskInfo[]
  succeeded: number
  failed: number
  total: number
}

export interface DispatchView {
  lanes: DispatchLane[]
  running: number
  settled: number
  total: number
  /** Tasks known to the build but not (yet) assigned a worker. */
  unassigned: number
}

/**
 * Groups a build's tasks by worker. Every registered worker gets a lane
 * (idle ones show the build is not using them); workers that ran tasks
 * but have since left the registry keep theirs. Lanes keep a stable
 * order — by name, with workers the build never used last — so rows do
 * not jump around while the build is being watched.
 */
export function dispatchView(tasks: TaskInfo[], workers: WorkerInfo[]): DispatchView {
  const lanes = new Map<string, DispatchLane>()
  const lane = (workerId: string, worker?: WorkerInfo): DispatchLane => {
    let entry = lanes.get(workerId)
    if (!entry) {
      entry = { workerId, worker, running: [], recent: [], succeeded: 0, failed: 0, total: 0 }
      lanes.set(workerId, entry)
    }
    return entry
  }
  for (const worker of workers) lane(worker.id, worker)

  let running = 0
  let unassigned = 0
  for (const task of tasks) {
    if (!task.worker_id) {
      unassigned++
      continue
    }
    const entry = lane(task.worker_id)
    entry.total++
    const tone = normalizeStatus(task.status, task.exit_code)
    if (tone === 'running' && !task.completed_at_ms) {
      entry.running.push(task)
      running++
    } else {
      entry.recent.push(task)
      if (tone === 'failed' || tone === 'timeout') entry.failed++
      else entry.succeeded++
    }
  }

  for (const entry of lanes.values()) {
    entry.running.sort((a, b) => a.started_at_ms - b.started_at_ms)
    entry.recent.sort((a, b) => (b.completed_at_ms ?? b.started_at_ms) - (a.completed_at_ms ?? a.started_at_ms))
    entry.recent = entry.recent.slice(0, RECENT_PER_LANE)
  }

  const used = (lane: DispatchLane) => (lane.total > 0 ? 0 : 1)
  const sorted = [...lanes.values()].sort((a, b) => used(a) - used(b) || a.workerId.localeCompare(b.workerId))

  return {
    lanes: sorted,
    running,
    settled: tasks.length - running - unassigned,
    total: tasks.length,
    unassigned,
  }
}

/**
 * A lane's display name: the host, unless several lanes share it (e.g.
 * multiple workers on one machine), in which case the unique worker id.
 */
export function laneNames(lanes: DispatchLane[]): Map<string, string> {
  return uniqueWorkerNames(lanes.map((lane) => ({ id: lane.workerId, host: lane.worker?.host })))
}

/** Host when it identifies one worker, else the worker id (several workers on one machine). */
export function uniqueWorkerNames(items: { id: string; host?: string }[]): Map<string, string> {
  const hostCount = new Map<string, number>()
  for (const { host } of items) if (host) hostCount.set(host, (hostCount.get(host) ?? 0) + 1)
  return new Map(items.map(({ id, host }) => [id, host && hostCount.get(host) === 1 ? host : id]))
}

/**
 * Re-derives a build's counters from its (live-merged) task list with
 * the coordinator's own rules (telemetry Store.buildInfo), so the header
 * badge, stats and donut move in step with the dispatch view instead of
 * lagging a poll behind. A truncated build's task list is a subset, so
 * its server-side counters are kept.
 */
export function liveBuild(build: BuildInfo, tasks: TaskInfo[]): BuildInfo {
  if (build.truncated || tasks.length === 0) return build
  let running = 0
  let failed = 0
  let completed = 0
  let fromCache = 0
  let first = 0
  let last = 0
  for (const task of tasks) {
    if (task.status === 'running') running++
    else if (task.status === 'failed') failed++
    else completed++
    if (task.from_cache) fromCache++
    if (task.started_at_ms && (!first || task.started_at_ms < first)) first = task.started_at_ms
    const at = task.completed_at_ms || task.started_at_ms
    if (at > last) last = at
  }
  return {
    ...build,
    status: running > 0 ? 'running' : failed > 0 ? 'failed' : 'completed',
    total_tasks: tasks.length,
    running_tasks: running,
    failed_tasks: failed,
    completed_tasks: completed,
    from_cache_count: fromCache,
    first_task_at_ms: first,
    last_task_at_ms: last,
  }
}
