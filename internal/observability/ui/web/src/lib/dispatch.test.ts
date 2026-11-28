import { describe, expect, it } from 'vitest'
import { RECENT_PER_LANE, dispatchView, laneNames, liveBuild, mergeLiveTasks } from './dispatch'
import type { BuildInfo, TaskInfo, WorkerInfo, WsMessage } from './types'

function task(id: string, overrides: Partial<TaskInfo> = {}): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: 'completed',
    worker_id: 'w1',
    started_at_ms: 100,
    completed_at_ms: 200,
    queue_ms: 0,
    compile_ms: 100,
    from_cache: false,
    ...overrides,
  }
}

function worker(id: string, host = id): WorkerInfo {
  return { id, host } as WorkerInfo
}

const running = { status: 'running', completed_at_ms: undefined }

function msg(type: WsMessage['type'], data: TaskInfo): WsMessage {
  return { type, timestamp: 1, data }
}

describe('mergeLiveTasks', () => {
  it('adds a freshly dispatched task of this build and ignores other builds', () => {
    const merged = mergeLiveTasks(
      [task('a')],
      [msg('task_started', task('b', running)), msg('task_started', task('x', { ...running, build_id: 'other' }))],
      'b1',
    )
    expect(merged.map((t) => t.id).sort()).toEqual(['a', 'b'])
  })

  it('applies events oldest-first (store keeps newest first)', () => {
    const merged = mergeLiveTasks(
      [],
      [msg('task_completed', task('a', { status: 'failed', exit_code: 1 })), msg('task_started', task('a', running))],
      'b1',
    )
    expect(merged).toHaveLength(1)
    expect(merged[0].status).toBe('failed')
  })

  it('never downgrades a finished task back to running', () => {
    const merged = mergeLiveTasks([task('a')], [msg('task_started', task('a', running))], 'b1')
    expect(merged[0].status).toBe('completed')
  })

  it('merges every build when no build id is given', () => {
    const merged = mergeLiveTasks([], [msg('task_started', task('x', { ...running, build_id: 'other' }))])
    expect(merged.map((t) => t.id)).toEqual(['x'])
  })

  it('ignores non-task frames', () => {
    expect(mergeLiveTasks([], [{ type: 'stats', timestamp: 1, data: {} }], 'b1')).toEqual([])
  })
})

describe('dispatchView', () => {
  it('gives every registered worker a lane in a stable order, unused workers last', () => {
    const view = dispatchView(
      [
        task('a', { worker_id: 'w1' }),
        task('b', { ...running, worker_id: 'w2', started_at_ms: 300 }),
        task('c', { ...running, worker_id: 'w2', started_at_ms: 200 }),
        task('d', { status: 'failed', exit_code: 1, worker_id: 'w1' }),
      ],
      [worker('w3'), worker('w2'), worker('w1')],
    )
    expect(view.lanes.map((l) => l.workerId)).toEqual(['w1', 'w2', 'w3'])
    expect(view.lanes[1].running.map((t) => t.id)).toEqual(['c', 'b'])
    expect(view.lanes[0]).toMatchObject({ succeeded: 1, failed: 1, total: 2 })
    expect(view.lanes[2].total).toBe(0)
    expect(view).toMatchObject({ running: 2, settled: 2, total: 4, unassigned: 0 })
  })

  it('keeps lanes for workers that left the registry and counts unassigned tasks', () => {
    const view = dispatchView([task('a', { worker_id: 'gone' }), task('b', { worker_id: '' })], [])
    expect(view.lanes.map((l) => l.workerId)).toEqual(['gone'])
    expect(view.lanes[0].worker).toBeUndefined()
    expect(view.unassigned).toBe(1)
  })

  it('caps the recent list, newest first', () => {
    const tasks = Array.from({ length: RECENT_PER_LANE + 3 }, (_, i) => task(`t${i}`, { completed_at_ms: 1000 + i }))
    const [lane] = dispatchView(tasks, [worker('w1')]).lanes
    expect(lane.recent).toHaveLength(RECENT_PER_LANE)
    expect(lane.recent[0].id).toBe(`t${RECENT_PER_LANE + 2}`)
    expect(lane.total).toBe(RECENT_PER_LANE + 3)
  })
})

describe('laneNames', () => {
  it('uses the host unless several workers share it', () => {
    const view = dispatchView([], [worker('a', 'mac'), worker('b', 'mac'), worker('c', 'linux-1')])
    const names = laneNames(view.lanes)
    expect(names.get('a')).toBe('a')
    expect(names.get('b')).toBe('b')
    expect(names.get('c')).toBe('linux-1')
  })
})

describe('liveBuild', () => {
  const build: BuildInfo = {
    id: 'b1',
    build_type: 'cpp',
    status: 'running',
    total_tasks: 2,
    completed_tasks: 0,
    failed_tasks: 0,
    running_tasks: 2,
    from_cache_count: 0,
    first_task_at_ms: 100,
    last_task_at_ms: 100,
    truncated: false,
    scheduler: 'p2c',
  }

  it('recomputes counters with the coordinator rules', () => {
    const live = liveBuild(build, [
      task('a', { started_at_ms: 100, completed_at_ms: 400 }),
      task('b', { status: 'failed', exit_code: 1, started_at_ms: 120, completed_at_ms: 500 }),
      task('c', { ...running, started_at_ms: 450 }),
    ])
    expect(live).toMatchObject({
      status: 'running',
      total_tasks: 3,
      running_tasks: 1,
      failed_tasks: 1,
      completed_tasks: 1,
      first_task_at_ms: 100,
      last_task_at_ms: 500,
      scheduler: 'p2c',
    })
  })

  it('keeps server counters for truncated builds or an empty list', () => {
    expect(liveBuild({ ...build, truncated: true }, [task('a')])).toEqual({ ...build, truncated: true })
    expect(liveBuild(build, [])).toBe(build)
  })
})
