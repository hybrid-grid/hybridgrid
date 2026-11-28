import { describe, expect, it } from 'vitest'
import type { TaskInfo, WorkerInfo } from './types'
import { utilization } from './utilization'

/** Times in tests are offsets from T0: real timestamps are never 0, and 0 means "not started". */
const T0 = 1_000_000

function task(id: string, worker: string, start: number, end?: number, overrides: Partial<TaskInfo> = {}): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: end === undefined ? 'running' : 'completed',
    worker_id: worker,
    started_at_ms: T0 + start,
    completed_at_ms: end === undefined ? undefined : T0 + end,
    queue_ms: 0,
    compile_ms: 0,
    from_cache: false,
    ...overrides,
  }
}

const worker = (id: string, max: number, host = id) => ({ id, host, max_parallel_tasks: max }) as WorkerInfo

describe('utilization', () => {
  const window = { start: T0, end: T0 + 1000 }
  const NOW = T0 + 1000

  it('averages busy slots per bucket and sums the cluster', () => {
    // w1 (2 slots): one task over the first half, another over the whole window.
    const u = utilization([task('a', 'w1', 0, 500), task('b', 'w1', 0, 1000)], [worker('w1', 2), worker('w2', 2)], window, NOW, 10)
    const w1 = u.rows.find((r) => r.id === 'w1')!
    expect(w1.busy.slice(0, 5)).toEqual([2, 2, 2, 2, 2])
    expect(w1.busy.slice(5)).toEqual([1, 1, 1, 1, 1])
    expect(w1.ratio[0]).toBe(1)
    expect(w1.ratio[9]).toBe(0.5)
    expect(u.total[0]).toBe(2)
    expect(u.capacity).toBe(4)
    expect(u.peak).toBe(2)
    expect(u.rows.find((r) => r.id === 'w2')!.taskCount).toBe(0)
  })

  it('splits a task across the buckets it partly covers', () => {
    const u = utilization([task('a', 'w1', 50, 150)], [worker('w1', 1)], window, NOW, 10)
    expect(u.rows[0].busy[0]).toBeCloseTo(0.5)
    expect(u.rows[0].busy[1]).toBeCloseTo(0.5)
  })

  it('extends running tasks to now and counts them', () => {
    const u = utilization([task('a', 'w1', 500)], [worker('w1', 1)], window, NOW, 10)
    expect(u.rows[0].runningNow).toBe(1)
    expect(u.rows[0].busy[9]).toBeCloseTo(1)
  })

  it('marks failures in the bucket they finished, and keeps unregistered workers', () => {
    const u = utilization([task('a', 'gone', 0, 250, { status: 'failed', exit_code: 1 })], [], window, NOW, 10)
    expect(u.rows.map((r) => r.id)).toEqual(['gone'])
    expect(u.rows[0].failures[2]).toBe(1)
    expect(u.rows[0].ratio[0]).toBe(1) // unknown capacity: busy shows as full
  })

  it('names workers that share a host by id', () => {
    const u = utilization([], [worker('a', 1, 'mac'), worker('b', 1, 'mac')], window, NOW, 10)
    expect(u.rows.map((r) => r.name)).toEqual(['a', 'b'])
  })
})
