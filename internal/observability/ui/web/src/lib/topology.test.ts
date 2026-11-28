import { describe, expect, it } from 'vitest'
import { MAX_BUILD_NODES, NODE_SIZE, UNGROUPED, buildTopology, layoutTopology, orthoPath, routeLength, withLiveActivity } from './topology'
import type { TaskInfo, WorkerInfo } from './types'

function task(id: string, overrides: Partial<TaskInfo> = {}): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: 'running',
    worker_id: 'w1',
    started_at_ms: 100,
    queue_ms: 0,
    compile_ms: 0,
    from_cache: false,
    ...overrides,
  }
}

const worker = (id: string, host = id) => ({ id, host }) as WorkerInfo

describe('buildTopology', () => {
  it('groups running tasks by worker and by build, ignoring finished ones', () => {
    const topo = buildTopology(
      [worker('w2'), worker('w1'), worker('w3')],
      [
        task('a', { worker_id: 'w1', started_at_ms: 300 }),
        task('b', { worker_id: 'w1', started_at_ms: 200 }),
        task('c', { worker_id: 'w2', build_id: 'b2' }),
        task('d', { worker_id: 'w2', build_id: '' }),
        task('e', { status: 'completed', completed_at_ms: 400 }),
      ],
    )
    expect(topo.running).toBe(4)
    expect(topo.workers.map((w) => [w.id, w.running.map((t) => t.id)])).toEqual([
      ['w1', ['b', 'a']],
      ['w2', ['c', 'd']],
      ['w3', []],
    ])
    expect(topo.builds).toEqual([
      { id: 'b1', running: 2, workerIds: ['w1'] },
      { id: 'b2', running: 1, workerIds: ['w2'] },
      { id: UNGROUPED, running: 1, workerIds: ['w2'] },
    ])
  })

  it('keeps an unregistered worker that still holds tasks, and names shared hosts by id', () => {
    const topo = buildTopology([worker('a', 'mac'), worker('b', 'mac')], [task('x', { worker_id: 'gone' })])
    expect(topo.workers.map((w) => [w.id, w.name, Boolean(w.worker)])).toEqual([
      ['a', 'a', true],
      ['b', 'b', true],
      ['gone', 'gone', false],
    ])
  })

  it('collapses builds beyond the cap into one bucket', () => {
    const tasks = Array.from({ length: MAX_BUILD_NODES + 2 }, (_, i) => task(`t${i}`, { build_id: `b${i}`, worker_id: `w${i % 2}` }))
    const topo = buildTopology([], tasks)
    expect(topo.builds).toHaveLength(MAX_BUILD_NODES)
    expect(topo.otherBuilds.builds).toBe(2)
    expect(topo.otherBuilds.running).toBe(2)
  })
})

describe('layoutTopology', () => {
  const W = NODE_SIZE.worker

  it('puts builds on top, the coordinator below, and workers in a grid underneath', () => {
    const layout = layoutTopology(2, 10, 1100)
    expect(layout.cols).toBe(3)
    expect(layout.coordinator.y).toBeGreaterThan(layout.builds[0].y + NODE_SIZE.build.h)
    expect(Math.min(...layout.workers.map((p) => p.y))).toBeGreaterThan(layout.coordinator.y + NODE_SIZE.coordinator.h)
    expect(new Set(layout.workers.map((p) => p.y)).size).toBe(4) // 10 workers / 3 columns
    expect(layout.height).toBe(Math.max(...layout.workers.map((p) => p.y)) + W.h)
  })

  it('adapts the column count to the width (one column on phones)', () => {
    expect(layoutTopology(1, 10, 1600).cols).toBe(4)
    expect(layoutTopology(1, 10, 340).cols).toBe(1)
    expect(layoutTopology(1, 2, 1600).cols).toBe(2)
  })

  it('routes every worker edge orthogonally from the coordinator to the top of its card, through the gaps', () => {
    const layout = layoutTopology(1, 10, 1100)
    const coordBottom = { x: layout.coordinator.x + NODE_SIZE.coordinator.w / 2, y: layout.coordinator.y + NODE_SIZE.coordinator.h }
    layout.routes.workers.forEach((route, i) => {
      const card = layout.workers[i]
      expect(route[0]).toEqual(coordBottom)
      expect(route[route.length - 1]).toEqual({ x: card.x + W.w / 2, y: card.y })
      for (let k = 1; k < route.length; k++) {
        const [a, b] = [route[k - 1], route[k]]
        expect(a.x === b.x || a.y === b.y).toBe(true) // axis-aligned
        // No segment may pass through any worker card.
        for (const other of layout.workers) {
          const minX = Math.min(a.x, b.x), maxX = Math.max(a.x, b.x), minY = Math.min(a.y, b.y), maxY = Math.max(a.y, b.y)
          const crosses = maxX > other.x + 1 && minX < other.x + W.w - 1 && maxY > other.y + 1 && minY < other.y + W.h - 1
          expect(crosses).toBe(false)
        }
      }
    })
  })

  it('handles an empty fleet', () => {
    const layout = layoutTopology(0, 0, 1100)
    expect(layout.workers).toEqual([])
    expect(layout.height).toBe(layout.coordinator.y + NODE_SIZE.coordinator.h)
  })
})

describe('orthoPath', () => {
  it('rounds corners and drops collinear points', () => {
    expect(orthoPath([{ x: 0, y: 0 }, { x: 0, y: 20 }, { x: 0, y: 40 }])).toBe('M0 0 L0 40')
    expect(orthoPath([{ x: 0, y: 0 }, { x: 0, y: 40 }, { x: 40, y: 40 }], 8)).toBe('M0 0 L0 32 Q0 40 8 40 L40 40')
    expect(routeLength([{ x: 0, y: 0 }, { x: 0, y: 40 }, { x: 30, y: 40 }])).toBe(70)
  })
})

describe('withLiveActivity', () => {
  it('replaces the heartbeat counter with the compiles actually running', () => {
    const stale = [{ ...worker('w1'), active_tasks: 3 }, { ...worker('w2'), active_tasks: 0 }]
    const live = withLiveActivity(stale, [task('a', { worker_id: 'w2' }), task('b', { worker_id: 'w2' }), task('c', { worker_id: 'w1', status: 'completed', completed_at_ms: 9 })])
    expect(live.map((w) => [w.id, w.active_tasks])).toEqual([['w1', 0], ['w2', 2]])
  })
})
