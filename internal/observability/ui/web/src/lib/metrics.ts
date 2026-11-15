import { normalizeStatus } from './status'
import type { BuildInfo, Stats, TaskInfo, WorkerInfo } from './types'

/** Share of finished tasks that succeeded, or undefined before any finish. */
export function passRate(stats: Pick<Stats, 'success_tasks' | 'failed_tasks'>): number | undefined {
  const finished = stats.success_tasks + stats.failed_tasks
  return finished > 0 ? stats.success_tasks / finished : undefined
}

export interface FleetCapacity {
  active: number
  max: number
  ratio: number | undefined
}

export function fleetCapacity(workers: WorkerInfo[]): FleetCapacity {
  const max = workers.reduce((sum, worker) => sum + Math.max(0, worker.max_parallel_tasks || 0), 0)
  const active = workers.reduce((sum, worker) => sum + Math.max(0, worker.active_tasks || 0), 0)
  return { active, max, ratio: max > 0 ? Math.min(1, active / max) : undefined }
}

/**
 * The one build-duration predicate every surface shares (table column,
 * average tile, durations chart), so two numbers on one screen cannot
 * disagree about the same build. Same-millisecond builds are a real 0;
 * missing or inverted timestamps mean "no duration".
 */
export function buildDurationMs(build: BuildInfo): number | null {
  if (!build.first_task_at_ms || !build.last_task_at_ms) return null
  if (build.last_task_at_ms < build.first_task_at_ms) return null
  return build.last_task_at_ms - build.first_task_at_ms
}

export function buildTone(build: BuildInfo) {
  if (build.running_tasks > 0) return 'running' as const
  if (build.failed_tasks > 0) return 'failed' as const
  return normalizeStatus(build.status)
}

/** Average span of builds that have settled — a running build's span is still growing. */
export function averageFinishedBuildMs(builds: BuildInfo[]): number | undefined {
  const durations = builds
    .filter((build) => buildTone(build) !== 'running')
    .map(buildDurationMs)
    .filter((ms): ms is number => ms !== null)
  if (durations.length === 0) return undefined
  return durations.reduce((sum, ms) => sum + ms, 0) / durations.length
}

export function buildProgress(build: BuildInfo): number {
  if (!build.total_tasks) return 0
  return Math.min(1, (build.completed_tasks + build.failed_tasks) / build.total_tasks)
}

export interface ThroughputSeries {
  workerId: string
  counts: number[]
}

export interface Throughput {
  bucketMs: number
  bucketStarts: number[]
  series: ThroughputSeries[]
  totals: number[]
  max: number
}

const MIN_BUCKET_MS = 1000

/**
 * Finished tasks per worker, bucketed from the oldest finish in the
 * window up to `now`. Failed finishes count too: a worker failing fast
 * is still activity, and that is exactly what this view should expose.
 */
export function workerThroughput(tasks: TaskInfo[], now: number, bucketCount = 12): Throughput {
  const done = tasks.filter((task) => task.completed_at_ms && task.worker_id)
  if (done.length === 0) {
    return { bucketMs: 0, bucketStarts: [], series: [], totals: [], max: 0 }
  }

  const finishes = done.map((task) => task.completed_at_ms as number)
  const newest = Math.max(now, ...finishes)
  const oldest = Math.min(...finishes)
  const bucketMs = Math.max(Math.ceil((newest - oldest + 1) / bucketCount), MIN_BUCKET_MS)
  const start = newest - bucketMs * bucketCount
  const bucketStarts = Array.from({ length: bucketCount }, (_, i) => start + i * bucketMs)

  const workerIds = [...new Set(done.map((task) => task.worker_id))].sort()
  const series = workerIds.map((workerId) => {
    const counts = new Array<number>(bucketCount).fill(0)
    for (const task of done) {
      if (task.worker_id !== workerId) continue
      const index = Math.min(bucketCount - 1, Math.floor(((task.completed_at_ms as number) - start) / bucketMs))
      if (index >= 0) counts[index] += 1
    }
    return { workerId, counts }
  })

  const totals = bucketStarts.map((_, i) => series.reduce((sum, s) => sum + s.counts[i], 0))
  return { bucketMs, bucketStarts, series, totals, max: Math.max(0, ...totals) }
}

/** Rounds an axis maximum up to a 1/2/5 × 10ⁿ step so gridlines land on readable values. */
export function niceCeil(value: number): number {
  if (value <= 0) return 1
  const magnitude = 10 ** Math.floor(Math.log10(value))
  const normalized = value / magnitude
  const step = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10
  return step * magnitude
}

export interface TaskWindowStats {
  total: number
  finished: number
  failed: number
  successRate: number | undefined
  avgCompileMs: number | undefined
  avgQueueMs: number | undefined
  cacheHits: number
}

function average(values: number[]): number | undefined {
  return values.length > 0 ? values.reduce((sum, v) => sum + v, 0) / values.length : undefined
}

/** Aggregates over whatever task window the caller passes (a worker's tasks, a build's, the fleet's). */
export function taskWindowStats(tasks: TaskInfo[]): TaskWindowStats {
  const finished = tasks.filter((task) => task.completed_at_ms)
  const failed = finished.filter((task) => {
    const tone = normalizeStatus(task.status, task.exit_code)
    return tone === 'failed' || tone === 'timeout'
  }).length
  return {
    total: tasks.length,
    finished: finished.length,
    failed,
    successRate: finished.length > 0 ? (finished.length - failed) / finished.length : undefined,
    avgCompileMs: average(finished.filter((task) => !task.from_cache && task.compile_ms > 0).map((task) => task.compile_ms)),
    avgQueueMs: average(tasks.filter((task) => task.started_at_ms > 0).map((task) => task.queue_ms)),
    cacheHits: tasks.filter((task) => task.from_cache).length,
  }
}

export interface CacheSavings {
  cachedCount: number
  avgCompileMs: number
  savedMs: number
}

/**
 * Rough compile time the cache avoided: cache hits × the average
 * compile time of real (uncached) compiles in the same window. An
 * estimate by construction — the UI labels it as one.
 */
export function estimateCacheSavings(tasks: TaskInfo[]): CacheSavings | undefined {
  const { cacheHits, avgCompileMs } = taskWindowStats(tasks)
  if (cacheHits === 0 || avgCompileMs === undefined) return undefined
  return { cachedCount: cacheHits, avgCompileMs, savedMs: cacheHits * avgCompileMs }
}

export interface BuildCacheShare {
  build: BuildInfo
  share: number
}

/** Cache share per build, oldest first so a chart reads left-to-right in time. */
export function buildCacheShares(builds: BuildInfo[]): BuildCacheShare[] {
  return builds
    .filter((build) => build.total_tasks > 0)
    .map((build) => ({ build, share: build.from_cache_count / build.total_tasks }))
    .toSorted((a, b) => a.build.first_task_at_ms - b.build.first_task_at_ms)
}
