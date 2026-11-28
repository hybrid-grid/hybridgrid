import { uniqueWorkerNames } from './dispatch'
import { normalizeStatus } from './status'
import { taskRunEnd } from './timeline'
import type { TaskInfo, WorkerInfo } from './types'

export const UTILIZATION_BUCKETS = 60

export interface UtilizationRow {
  id: string
  name: string
  worker?: WorkerInfo
  /** Slots the worker advertises (0 when it has left the registry). */
  max: number
  /** Average busy slots per bucket. */
  busy: number[]
  /** busy / max per bucket, 0..1. */
  ratio: number[]
  /** Failed tasks that finished in each bucket. */
  failures: number[]
  /** Tasks of this worker that overlap the window. */
  taskCount: number
  /** Tasks compiling on it right now. */
  runningNow: number
}

export interface Utilization {
  start: number
  end: number
  bucketMs: number
  /** Average busy slots across the cluster per bucket. */
  total: number[]
  /** Sum of advertised slots of registered workers. */
  capacity: number
  /** Highest total in any bucket, rounded to whole slots. */
  peak: number
  rows: UtilizationRow[]
  taskCount: number
}

function isRunning(task: TaskInfo): boolean {
  return !task.completed_at_ms && normalizeStatus(task.status, task.exit_code) === 'running'
}

/**
 * How busy each worker (and the cluster) was over [start, end], as
 * average busy slots per bucket — a utilization strip rather than one
 * bar per task, so it stays readable for many machines and thousands of
 * compiles. Every registered worker gets a row (idle ones show spare
 * capacity); workers that ran tasks but left the registry keep theirs.
 */
export function utilization(
  tasks: TaskInfo[],
  workers: WorkerInfo[],
  window: { start: number; end: number },
  now: number,
  buckets = UTILIZATION_BUCKETS,
): Utilization {
  const start = window.start
  const end = Math.max(window.end, start + buckets) // never a zero-width window
  const bucketMs = (end - start) / buckets
  const registered = new Map(workers.map((w) => [w.id, w]))

  const ids = new Set(workers.map((w) => w.id))
  for (const task of tasks) if (task.worker_id) ids.add(task.worker_id)
  const names = uniqueWorkerNames([...ids].map((id) => ({ id, host: registered.get(id)?.host })))

  const rows = new Map<string, UtilizationRow>()
  for (const id of [...ids].sort()) {
    const worker = registered.get(id)
    rows.set(id, {
      id,
      name: names.get(id) ?? id,
      worker,
      max: worker?.max_parallel_tasks ?? 0,
      busy: new Array(buckets).fill(0),
      ratio: new Array(buckets).fill(0),
      failures: new Array(buckets).fill(0),
      taskCount: 0,
      runningNow: 0,
    })
  }

  let taskCount = 0
  for (const task of tasks) {
    const row = task.worker_id ? rows.get(task.worker_id) : undefined
    if (!row || !task.started_at_ms) continue
    if (isRunning(task)) row.runningNow++
    const s = Math.max(task.started_at_ms, start)
    const e = Math.min(taskRunEnd(task, now), end)
    if (e <= s) continue
    row.taskCount++
    taskCount++
    const first = Math.min(buckets - 1, Math.floor((s - start) / bucketMs))
    const last = Math.min(buckets - 1, Math.floor((e - start) / bucketMs))
    for (let b = first; b <= last; b++) {
      const overlap = Math.min(e, start + (b + 1) * bucketMs) - Math.max(s, start + b * bucketMs)
      if (overlap > 0) row.busy[b] += overlap / bucketMs
    }
    const tone = normalizeStatus(task.status, task.exit_code)
    if ((tone === 'failed' || tone === 'timeout') && task.completed_at_ms && task.completed_at_ms >= start && task.completed_at_ms <= end) {
      row.failures[Math.min(buckets - 1, Math.floor((task.completed_at_ms - start) / bucketMs))]++
    }
  }

  const total = new Array(buckets).fill(0)
  for (const row of rows.values()) {
    for (let b = 0; b < buckets; b++) {
      total[b] += row.busy[b]
      row.ratio[b] = row.max > 0 ? Math.min(1, row.busy[b] / row.max) : row.busy[b] > 0 ? 1 : 0
    }
  }

  return {
    start,
    end,
    bucketMs,
    total,
    capacity: workers.reduce((sum, w) => sum + Math.max(0, w.max_parallel_tasks || 0), 0),
    peak: Math.round(Math.max(0, ...total)),
    rows: [...rows.values()],
    taskCount,
  }
}
