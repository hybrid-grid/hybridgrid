import { describe, expect, it } from 'vitest'
import { applyTaskQuery, paginate } from './taskQuery'
import type { TaskInfo } from './types'

function task(id: string, overrides: Partial<TaskInfo> = {}): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: 'completed',
    worker_id: 'w1',
    started_at_ms: 100,
    completed_at_ms: 200,
    duration_ms: 100,
    queue_ms: 0,
    compile_ms: 100,
    from_cache: false,
    ...overrides,
  }
}

const tasks = [
  task('a', { started_at_ms: 1, duration_ms: 50 }),
  task('b', { started_at_ms: 2, status: 'failed', exit_code: 1, error_message: 'boom', worker_id: 'w2' }),
  task('c', { started_at_ms: 3, status: 'running', duration_ms: undefined, from_cache: true }),
]

describe('applyTaskQuery', () => {
  it('defaults to newest first', () => {
    expect(applyTaskQuery(tasks, {}).map((t) => t.id)).toEqual(['c', 'b', 'a'])
  })

  it('filters by status, worker, cache and free text', () => {
    expect(applyTaskQuery(tasks, { status: 'failed' }).map((t) => t.id)).toEqual(['b'])
    expect(applyTaskQuery(tasks, { worker: 'w2' }).map((t) => t.id)).toEqual(['b'])
    expect(applyTaskQuery(tasks, { cache: 'hit' }).map((t) => t.id)).toEqual(['c'])
    expect(applyTaskQuery(tasks, { q: 'BOOM' }).map((t) => t.id)).toEqual(['b'])
  })

  it('sorts by duration ascending with unfinished tasks first', () => {
    expect(applyTaskQuery(tasks, { sort: 'duration', dir: 'asc' }).map((t) => t.id)).toEqual(['c', 'a', 'b'])
  })

  it('does not mutate the input array', () => {
    const copy = [...tasks]
    applyTaskQuery(tasks, { sort: 'duration', dir: 'asc' })
    expect(tasks).toEqual(copy)
  })
})

describe('paginate', () => {
  it('clamps the page into range', () => {
    const items = Array.from({ length: 5 }, (_, i) => i)
    expect(paginate(items, 9, 2)).toEqual({ rows: [4], page: 3, pageCount: 3 })
    expect(paginate([], 1, 2)).toEqual({ rows: [], page: 1, pageCount: 1 })
  })
})
