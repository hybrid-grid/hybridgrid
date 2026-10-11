import { describe, expect, it } from 'vitest'
import { layoutCluster, layoutTimeline, peakConcurrentTasks, taskToneClass } from './timeline'
import type { TaskInfo, WorkerInfo } from './types'

function task(id: string, startedAtMs: number, completedAtMs?: number, durationMs?: number): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'build-1',
    status: 'success',
    worker_id: 'worker-1',
    started_at_ms: startedAtMs,
    completed_at_ms: completedAtMs,
    duration_ms: durationMs,
    queue_ms: 0,
    compile_ms: 0,
    from_cache: false,
  }
}

describe('peakConcurrentTasks', () => {
  it('returns zero for an empty timeline', () => {
    expect(peakConcurrentTasks([])).toBe(0)
  })

  it('does not count sequential tasks as concurrent', () => {
    expect(peakConcurrentTasks([task('one', 100, 200), task('two', 200, 300)])).toBe(1)
  })

  it('counts overlapping task intervals', () => {
    expect(peakConcurrentTasks([task('one', 100, 300), task('two', 150, 250), task('three', 200, 400)])).toBe(3)
  })

  it('uses duration when completion time is unavailable', () => {
    expect(peakConcurrentTasks([task('one', 100, undefined, 100), task('two', 150, undefined, 100)])).toBe(2)
  })

  it('ignores tasks without a start and treats a running task as a minimal interval', () => {
    expect(peakConcurrentTasks([task('missing', 0), task('running', 100)])).toBe(1)
  })

  it('clamps malformed completion times to a minimal interval', () => {
    expect(peakConcurrentTasks([task('malformed', 100, 50), task('next', 101, 200)])).toBe(1)
  })
})

describe('taskToneClass', () => {
  it.each([
    ['STATUS_COMPLETED', 'success'],
    ['success', 'success'],
    ['completed', 'success'],
    ['STATUS_FAILED', 'failed'],
    ['STATUS_TIMEOUT', 'failed'],
    ['failed', 'failed'],
    ['timeout', 'failed'],
    ['STATUS_RUNNING', 'running'],
    ['running', 'running'],
    ['STATUS_QUEUED', 'queued'],
    ['queued', 'queued'],
    ['unexpected', 'queued'],
  ] as const)('maps %s to %s', (status, tone) => {
    expect(taskToneClass(status)).toBe(tone)
  })
})

describe('layoutTimeline', () => {
  it('extends a running task to now when now is supplied', () => {
    const running = { ...task('live', 1_500), status: 'running' }
    const layout = layoutTimeline([task('done', 1_000, 1_500), running], 2_000)

    expect(layout.windowEnd).toBe(2_000)
    const liveRow = layout.rows.find((row) => row.task.id === 'live')
    expect(liveRow?.runLeft).toBe(50)
    expect(liveRow?.runWidth).toBe(50)
  })

  it('orders rows by compile start', () => {
    const layout = layoutTimeline([task('late', 300, 400), task('early', 100, 200)])
    expect(layout.rows.map((row) => row.task.id)).toEqual(['early', 'late'])
  })
})

describe('layoutCluster', () => {
  const worker = (id: string, host = ''): WorkerInfo => ({ id, host, healthy: true }) as WorkerInfo

  it('keeps idle registered workers as empty lanes and appends history-only workers', () => {
    const tasks = [{ ...task('a', 100, 200), worker_id: 'ghost' }]
    const layout = layoutCluster(tasks, [worker('w1', 'box-1')], 1_000)

    expect(layout.lanes.map((lane) => [lane.id, lane.host, lane.known, lane.taskCount])).toEqual([
      ['w1', 'box-1', true, 0],
      ['ghost', 'ghost', false, 1],
    ])
  })

  it('packs overlapping tasks into separate sub-rows and reuses freed rows', () => {
    const tasks = [task('a', 100, 300), task('b', 150, 250), task('c', 300, 400)]
    const lane = layoutCluster(tasks, [worker('worker-1')], 1_000).lanes[0]

    expect(lane.slotRows.map((rows) => rows.map((row) => row.task.id))).toEqual([['a', 'c'], ['b']])
  })

  it('reserves a sub-row from queue entry so a queued task is not hidden', () => {
    const queued = { ...task('b', 300, 400), queue_ms: 150 }
    const lane = layoutCluster([task('a', 100, 300), queued], [worker('worker-1')], 1_000).lanes[0]

    expect(lane.slotRows).toHaveLength(2)
  })

  it('counts undispatched tasks and emits five axis ticks', () => {
    const layout = layoutCluster([task('a', 100, 200), task('pending', 0)], [], 1_000)

    expect(layout.undispatched).toBe(1)
    expect(layout.ticks).toHaveLength(5)
    expect(layout.ticks[0]).toBe(layout.windowStart)
    expect(layout.ticks.at(-1)).toBe(layout.windowEnd)
  })
})
