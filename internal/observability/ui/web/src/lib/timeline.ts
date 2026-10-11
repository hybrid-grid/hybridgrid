import { normalizeStatus } from './status'
import type { TaskInfo, WorkerInfo } from './types'

export interface TimelineRow {
  task: TaskInfo
  queueLeft: number
  queueWidth: number
  runLeft: number
  runWidth: number
}

export interface TimelineLayout {
  windowStart: number
  windowEnd: number
  rows: TimelineRow[]
}

const MIN_RUN_WIDTH_PCT = 0.6

/**
 * Lays out a set of tasks on a shared 0-100% time axis: a queue
 * segment (coordinator dispatch wait) followed by a run segment
 * (actual compile), mirroring the wire fields queue_ms/duration_ms.
 * Tasks with no timestamps are dropped — nothing to plot. When `now`
 * is given, still-running tasks extend to it so live bars grow.
 */
export function layoutTimeline(tasks: TaskInfo[], now?: number): TimelineLayout {
  const withTimes = tasks
    .filter((task) => task.started_at_ms > 0)
    .toSorted((a, b) => a.started_at_ms - b.started_at_ms)
  if (withTimes.length === 0) {
    return { windowStart: 0, windowEnd: 0, rows: [] }
  }

  const windowStart = Math.min(...withTimes.map(taskQueueStart))
  const windowEnd = Math.max(windowStart + 1, ...withTimes.map((task) => taskRunEnd(task, now)))
  const span = windowEnd - windowStart

  const rows: TimelineRow[] = withTimes.map((task) => {
    const queueLeft = ((taskQueueStart(task) - windowStart) / span) * 100
    const runLeft = ((task.started_at_ms - windowStart) / span) * 100
    const runRight = ((taskRunEnd(task, now) - windowStart) / span) * 100

    return {
      task,
      queueLeft: clampPercent(queueLeft),
      queueWidth: clampPercent(runLeft - queueLeft),
      runLeft: clampPercent(runLeft),
      runWidth: Math.max(MIN_RUN_WIDTH_PCT, clampPercent(runRight - runLeft)),
    }
  })

  return { windowStart, windowEnd, rows }
}

function taskQueueStart(task: TaskInfo): number {
  return task.started_at_ms - Math.max(0, task.queue_ms)
}

/**
 * End of a task's compile phase. A running task with no completion
 * extends to `now` when the caller supplies it, otherwise collapses to
 * a minimal interval so it is still plotted.
 */
export function taskRunEnd(task: TaskInfo, now?: number): number {
  if (!task.completed_at_ms && now !== undefined && normalizeStatus(task.status) === 'running') {
    return Math.max(task.started_at_ms + 1, now)
  }
  const fallbackEnd = task.started_at_ms + Math.max(1, task.duration_ms || 0)
  return Math.max(task.started_at_ms + 1, task.completed_at_ms || fallbackEnd)
}

export function peakConcurrentTasks(tasks: TaskInfo[], now?: number): number {
  const events = tasks
    .filter((task) => task.started_at_ms > 0)
    .flatMap((task) => [
      { at: task.started_at_ms, delta: 1 },
      { at: taskRunEnd(task, now), delta: -1 },
    ])
    .sort((a, b) => a.at - b.at || a.delta - b.delta)

  let active = 0
  let peak = 0
  for (const event of events) {
    active += event.delta
    peak = Math.max(peak, active)
  }
  return peak
}

function clampPercent(value: number): number {
  return Math.min(100, Math.max(0, value))
}

export function taskToneClass(status: string, exitCode?: number): 'success' | 'failed' | 'running' | 'queued' {
  const tone = normalizeStatus(status, exitCode)
  return tone === 'timeout' ? 'failed' : tone
}

export const TONE_COLOR: Record<ReturnType<typeof taskToneClass>, string> = {
  success: 'var(--color-status-success)',
  failed: 'var(--color-status-failed)',
  running: 'var(--color-status-running)',
  queued: 'var(--color-status-queued)',
}

export interface ClusterLane {
  id: string
  host: string
  known: boolean
  healthy: boolean
  taskCount: number
  /** One array per occupied executor sub-row. */
  slotRows: TimelineRow[][]
}

export interface ClusterLayout {
  lanes: ClusterLane[]
  windowStart: number
  windowEnd: number
  peak: number
  undispatched: number
  ticks: number[]
}

const AXIS_TICKS = 5

/**
 * Fleet-wide swimlanes: one lane per registered worker (idle ones
 * included, so a silent worker is visible), then any worker only known
 * from task history. Inside a lane, tasks are packed greedily into
 * sub-rows reserving from queue entry rather than compile start, so a
 * queued task never hides under the previous task's compile bar — the
 * dispatch wait is exactly what this view exists to show.
 */
export function layoutCluster(tasks: TaskInfo[], workers: WorkerInfo[], now: number): ClusterLayout {
  const dispatched = tasks.filter((task) => task.started_at_ms > 0 && task.worker_id)
  const undispatched = tasks.filter((task) => !task.started_at_ms).length
  const layout = layoutTimeline(dispatched, now)

  const byWorker = new Map<string, TimelineRow[]>()
  for (const row of layout.rows) {
    byWorker.set(row.task.worker_id, [...(byWorker.get(row.task.worker_id) ?? []), row])
  }

  const registeredIds = new Set(workers.map((worker) => worker.id))
  const historicalIds = [...byWorker.keys()].filter((id) => !registeredIds.has(id)).sort()
  const workerById = new Map(workers.map((worker) => [worker.id, worker]))

  const lanes = [...workers.map((worker) => worker.id), ...historicalIds].map((id): ClusterLane => {
    const rows = byWorker.get(id) ?? []
    const worker = workerById.get(id)
    return {
      id,
      host: worker?.host || id,
      known: Boolean(worker),
      healthy: worker?.healthy ?? false,
      taskCount: rows.length,
      slotRows: packSlots(rows, now),
    }
  })

  const span = layout.windowEnd - layout.windowStart
  const ticks =
    layout.rows.length === 0
      ? []
      : Array.from({ length: AXIS_TICKS }, (_, i) => layout.windowStart + (span * i) / (AXIS_TICKS - 1))

  return {
    lanes,
    windowStart: layout.windowStart,
    windowEnd: layout.windowEnd,
    peak: peakConcurrentTasks(dispatched, now),
    undispatched,
    ticks,
  }
}

function packSlots(rows: TimelineRow[], now: number): TimelineRow[][] {
  const slotEnds: number[] = []
  const slots: TimelineRow[][] = []
  for (const row of rows) {
    const queuedAt = taskQueueStart(row.task)
    const index = slotEnds.findIndex((end) => end <= queuedAt)
    const slot = index === -1 ? slotEnds.length : index
    slotEnds[slot] = taskRunEnd(row.task, now)
    slots[slot] = [...(slots[slot] ?? []), row]
  }
  return slots
}
