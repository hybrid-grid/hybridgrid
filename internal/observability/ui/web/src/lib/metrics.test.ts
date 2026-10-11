import { describe, expect, it } from 'vitest'
import {
  averageFinishedBuildMs,
  buildCacheShares,
  estimateCacheSavings,
  taskWindowStats,
  buildDurationMs,
  buildProgress,
  buildTone,
  fleetCapacity,
  niceCeil,
  passRate,
  workerThroughput,
} from './metrics'
import type { BuildInfo, TaskInfo, WorkerInfo } from './types'

function build(overrides: Partial<BuildInfo> = {}): BuildInfo {
  return {
    id: 'b1',
    build_type: 'cpp',
    status: 'completed',
    total_tasks: 4,
    completed_tasks: 4,
    failed_tasks: 0,
    running_tasks: 0,
    from_cache_count: 0,
    first_task_at_ms: 1_000,
    last_task_at_ms: 3_000,
    truncated: false,
    ...overrides,
  }
}

function finished(id: string, workerId: string, completedAtMs: number): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: 'completed',
    worker_id: workerId,
    started_at_ms: completedAtMs - 100,
    completed_at_ms: completedAtMs,
    queue_ms: 0,
    compile_ms: 100,
    from_cache: false,
  }
}

describe('passRate', () => {
  it('is undefined before any task finishes', () => {
    expect(passRate({ success_tasks: 0, failed_tasks: 0 })).toBeUndefined()
  })

  it('divides successes by finished tasks', () => {
    expect(passRate({ success_tasks: 3, failed_tasks: 1 })).toBe(0.75)
  })
})

describe('fleetCapacity', () => {
  it('sums active and max slots across workers', () => {
    const workers = [
      { active_tasks: 2, max_parallel_tasks: 4 },
      { active_tasks: 1, max_parallel_tasks: 4 },
    ] as WorkerInfo[]
    expect(fleetCapacity(workers)).toEqual({ active: 3, max: 8, ratio: 0.375 })
  })

  it('reports no ratio without capacity', () => {
    expect(fleetCapacity([]).ratio).toBeUndefined()
  })
})

describe('buildDurationMs', () => {
  it('returns the wall-clock span', () => {
    expect(buildDurationMs(build())).toBe(2_000)
  })

  it('treats a same-millisecond build as a real zero', () => {
    expect(buildDurationMs(build({ first_task_at_ms: 5, last_task_at_ms: 5 }))).toBe(0)
  })

  it('returns null for missing or inverted timestamps', () => {
    expect(buildDurationMs(build({ first_task_at_ms: 0 }))).toBeNull()
    expect(buildDurationMs(build({ first_task_at_ms: 10, last_task_at_ms: 5 }))).toBeNull()
  })
})

describe('buildTone', () => {
  it('prefers running over failed while tasks are in flight', () => {
    expect(buildTone(build({ running_tasks: 1, failed_tasks: 1 }))).toBe('running')
  })

  it('marks settled builds with failures as failed', () => {
    expect(buildTone(build({ failed_tasks: 1 }))).toBe('failed')
  })

  it('maps a clean completed build to success', () => {
    expect(buildTone(build())).toBe('success')
  })
})

describe('averageFinishedBuildMs', () => {
  it('excludes running builds and builds without a duration', () => {
    const builds = [
      build({ first_task_at_ms: 0, last_task_at_ms: 1_000 }),
      build({ first_task_at_ms: 1_000, last_task_at_ms: 2_000 }),
      build({ first_task_at_ms: 1_000, last_task_at_ms: 4_000 }),
      build({ running_tasks: 1, first_task_at_ms: 1_000, last_task_at_ms: 99_000 }),
    ]
    expect(averageFinishedBuildMs(builds)).toBe(2_000)
  })

  it('is undefined with nothing settled', () => {
    expect(averageFinishedBuildMs([])).toBeUndefined()
  })
})

describe('buildProgress', () => {
  it('counts failed tasks as progressed and clamps at 1', () => {
    expect(buildProgress(build({ total_tasks: 4, completed_tasks: 1, failed_tasks: 1 }))).toBe(0.5)
    expect(buildProgress(build({ total_tasks: 0 }))).toBe(0)
  })
})

describe('workerThroughput', () => {
  it('returns an empty series without finished tasks', () => {
    expect(workerThroughput([], 10_000).series).toEqual([])
  })

  it('buckets finished tasks per worker and totals them', () => {
    const now = 12_000
    const tasks = [finished('a', 'w1', 1_000), finished('b', 'w1', 11_900), finished('c', 'w2', 11_950)]
    const result = workerThroughput(tasks, now, 4)

    expect(result.series.map((s) => s.workerId)).toEqual(['w1', 'w2'])
    expect(result.totals.reduce((sum, n) => sum + n, 0)).toBe(3)
    expect(result.totals.at(-1)).toBe(2)
    expect(result.max).toBe(2)
  })

  it('never shrinks buckets below one second', () => {
    const tasks = [finished('a', 'w1', 5_000), finished('b', 'w1', 5_001)]
    expect(workerThroughput(tasks, 5_001, 12).bucketMs).toBe(1_000)
  })
})

describe('niceCeil', () => {
  it.each([
    [0, 1],
    [0.7, 1],
    [3, 5],
    [12, 20],
    [51, 100],
  ] as const)('rounds %d up to %d', (value, expected) => {
    expect(niceCeil(value)).toBe(expected)
  })
})

describe('taskWindowStats', () => {
  it('derives success rate and averages over the window', () => {
    const tasks = [
      finished('a', 'w1', 1_000),
      { ...finished('b', 'w1', 2_000), status: 'failed', exit_code: 1 },
      { ...finished('c', 'w1', 3_000), from_cache: true },
    ]
    const stats = taskWindowStats(tasks)
    expect(stats.finished).toBe(3)
    expect(stats.failed).toBe(1)
    expect(stats.successRate).toBeCloseTo(2 / 3)
    expect(stats.cacheHits).toBe(1)
    expect(stats.avgCompileMs).toBe(100)
  })
})

describe('estimateCacheSavings', () => {
  it('multiplies cache hits by the average real compile time', () => {
    const tasks = [finished('a', 'w1', 1_000), { ...finished('b', 'w1', 2_000), from_cache: true }]
    expect(estimateCacheSavings(tasks)).toEqual({ cachedCount: 1, avgCompileMs: 100, savedMs: 100 })
  })

  it('is undefined without hits', () => {
    expect(estimateCacheSavings([finished('a', 'w1', 1_000)])).toBeUndefined()
  })
})

describe('buildCacheShares', () => {
  it('orders oldest first and skips empty builds', () => {
    const shares = buildCacheShares([
      build({ id: 'new', first_task_at_ms: 9, from_cache_count: 2, total_tasks: 4 }),
      build({ id: 'old', first_task_at_ms: 1, from_cache_count: 1, total_tasks: 4 }),
      build({ id: 'empty', total_tasks: 0 }),
    ])
    expect(shares.map((s) => [s.build.id, s.share])).toEqual([
      ['old', 0.25],
      ['new', 0.5],
    ])
  })
})
