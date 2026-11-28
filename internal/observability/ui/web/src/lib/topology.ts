import { uniqueWorkerNames } from './dispatch'
import { normalizeStatus } from './status'
import type { TaskInfo, WorkerInfo } from './types'

/** Build nodes drawn individually; the rest collapse into one "+N" node. */
export const MAX_BUILD_NODES = 5
/** Key for running tasks that belong to no build (ungrouped compiles). */
export const UNGROUPED = '__ungrouped__'

export interface TopologyBuild {
  /** build id, or UNGROUPED */
  id: string
  running: number
  /** Workers this build's running tasks are on. */
  workerIds: string[]
}

export interface TopologyWorker {
  id: string
  name: string
  worker?: WorkerInfo
  /** Tasks compiling on this worker right now, oldest first. */
  running: TaskInfo[]
}

export interface Topology {
  builds: TopologyBuild[]
  /** Running tasks of builds beyond MAX_BUILD_NODES, and how many builds. */
  otherBuilds: { builds: number; running: number; workerIds: string[] }
  workers: TopologyWorker[]
  running: number
}

function isRunning(task: TaskInfo): boolean {
  return !task.completed_at_ms && normalizeStatus(task.status, task.exit_code) === 'running'
}

/**
 * The fleet as a flow: running builds → coordinator → workers. Every
 * registered worker appears (idle ones show spare capacity); a worker
 * holding tasks but missing from the registry is kept so in-flight work
 * never disappears. Builds sort by how much they are running.
 */
export function buildTopology(workers: WorkerInfo[], tasks: TaskInfo[]): Topology {
  const running = tasks.filter(isRunning).sort((a, b) => a.started_at_ms - b.started_at_ms)

  const byWorker = new Map<string, TaskInfo[]>()
  for (const worker of workers) byWorker.set(worker.id, [])
  for (const task of running) {
    if (!task.worker_id) continue
    byWorker.set(task.worker_id, [...(byWorker.get(task.worker_id) ?? []), task])
  }
  const registered = new Map(workers.map((w) => [w.id, w]))
  const names = uniqueWorkerNames([...byWorker.keys()].map((id) => ({ id, host: registered.get(id)?.host })))
  const topoWorkers = [...byWorker.entries()]
    .map(([id, tasksOnWorker]) => ({ id, name: names.get(id) ?? id, worker: registered.get(id), running: tasksOnWorker }))
    .sort((a, b) => a.id.localeCompare(b.id))

  const byBuild = new Map<string, { running: number; workers: Set<string> }>()
  for (const task of running) {
    const key = task.build_id || UNGROUPED
    const entry = byBuild.get(key) ?? { running: 0, workers: new Set<string>() }
    entry.running++
    if (task.worker_id) entry.workers.add(task.worker_id)
    byBuild.set(key, entry)
  }
  const builds = [...byBuild.entries()]
    .map(([id, entry]) => ({ id, running: entry.running, workerIds: [...entry.workers].sort() }))
    // Busiest first; loose (ungrouped) tasks after named builds on ties.
    .sort((a, b) => b.running - a.running || Number(a.id === UNGROUPED) - Number(b.id === UNGROUPED) || a.id.localeCompare(b.id))
  const shown = builds.slice(0, MAX_BUILD_NODES)
  const rest = builds.slice(MAX_BUILD_NODES)

  return {
    builds: shown,
    otherBuilds: {
      builds: rest.length,
      running: rest.reduce((sum, b) => sum + b.running, 0),
      workerIds: [...new Set(rest.flatMap((b) => b.workerIds))].sort(),
    },
    workers: topoWorkers,
    running: running.length,
  }
}

export interface XY {
  x: number
  y: number
}

export const NODE_SIZE = {
  build: { w: 200, h: 64 },
  coordinator: { w: 240, h: 150 },
  worker: { w: 320, h: 104 },
}

const BUILD_GAP = 16
const COL_GAP = 16
/** Vertical room between worker rows, where each row's bus line runs. */
const ROW_GAP = 40
/** Space between the build row, the coordinator and the first worker row. */
const TIER_GAP = 56
/** Left channel the trunk runs down to reach rows below the first. */
const CHANNEL = 28
const BUS_OFFSET = ROW_GAP / 2

/**
 * Top-down "org chart" layout: running builds in a row on top, the
 * coordinator under them, and the workers in a grid below that grows
 * downward — as many columns as availableWidth allows (1 on phones).
 *
 * Edges are orthogonal routes, never straight lines through the grid:
 * from the coordinator a line drops to a bus above the first row, then
 * branches into each worker; rows further down are reached through a
 * trunk in the left channel and a bus above their own row. Every
 * segment runs in the gaps, so no edge crosses a node.
 */
export function layoutTopology(buildCount: number, workerCount: number, availableWidth: number) {
  const card = NODE_SIZE.worker
  const fit = Math.floor((availableWidth - CHANNEL + COL_GAP) / (card.w + COL_GAP))
  const cols = Math.max(1, Math.min(fit, Math.max(workerCount, 1)))
  const gridW = cols * card.w + (cols - 1) * COL_GAP
  const n = Math.max(buildCount, 1)
  const buildsW = n * NODE_SIZE.build.w + (n - 1) * BUILD_GAP
  const width = Math.max(CHANNEL + gridW, buildsW, NODE_SIZE.coordinator.w)
  const gridLeft = CHANNEL + (width - CHANNEL - gridW) / 2
  const centerX = gridLeft + gridW / 2
  const trunkX = gridLeft - CHANNEL / 2

  const builds: XY[] = Array.from({ length: buildCount }, (_, i) => ({
    x: centerX - buildsW / 2 + i * (NODE_SIZE.build.w + BUILD_GAP),
    y: 0,
  }))
  const coordinator = { x: centerX - NODE_SIZE.coordinator.w / 2, y: NODE_SIZE.build.h + TIER_GAP }
  const coordBottom = coordinator.y + NODE_SIZE.coordinator.h
  const gridTop = coordBottom + TIER_GAP
  const firstBus = gridTop - BUS_OFFSET

  const workers: XY[] = Array.from({ length: workerCount }, (_, i) => ({
    x: gridLeft + (i % cols) * (card.w + COL_GAP),
    y: gridTop + Math.floor(i / cols) * (card.h + ROW_GAP),
  }))

  const workerRoutes: XY[][] = workers.map((p) => {
    const tx = p.x + card.w / 2
    const start = [
      { x: centerX, y: coordBottom },
      { x: centerX, y: firstBus },
    ]
    if (p.y === gridTop) return [...start, { x: tx, y: firstBus }, { x: tx, y: p.y }]
    const bus = p.y - BUS_OFFSET
    return [...start, { x: trunkX, y: firstBus }, { x: trunkX, y: bus }, { x: tx, y: bus }, { x: tx, y: p.y }]
  })
  const buildBus = NODE_SIZE.build.h + TIER_GAP / 2
  const buildRoutes: XY[][] = builds.map((p) => {
    const bx = p.x + NODE_SIZE.build.w / 2
    return [
      { x: bx, y: NODE_SIZE.build.h },
      { x: bx, y: buildBus },
      { x: centerX, y: buildBus },
      { x: centerX, y: coordinator.y },
    ]
  })

  const rows = Math.ceil(workerCount / cols)
  const height = workerCount > 0 ? gridTop + rows * (card.h + ROW_GAP) - ROW_GAP : coordBottom
  return { builds, coordinator, workers, cols, width, height, routes: { builds: buildRoutes, workers: workerRoutes } }
}

/** Drops repeated points and points that do not turn the route. */
function simplify(points: XY[]): XY[] {
  const out: XY[] = []
  for (const p of points) {
    const last = out[out.length - 1]
    if (last && last.x === p.x && last.y === p.y) continue
    const prev = out[out.length - 2]
    if (prev && last && ((prev.x === last.x && last.x === p.x) || (prev.y === last.y && last.y === p.y))) out.pop()
    out.push(p)
  }
  return out
}

/** SVG path through an orthogonal route, with corners rounded by radius. */
export function orthoPath(points: XY[], radius = 8): string {
  const pts = simplify(points)
  if (pts.length === 0) return ''
  let d = `M${pts[0].x} ${pts[0].y}`
  for (let i = 1; i < pts.length; i++) {
    const p = pts[i]
    const next = pts[i + 1]
    if (!next) {
      d += ` L${p.x} ${p.y}`
      break
    }
    const prev = pts[i - 1]
    const r = Math.min(radius, Math.hypot(p.x - prev.x, p.y - prev.y) / 2, Math.hypot(next.x - p.x, next.y - p.y) / 2)
    const into = { x: p.x - Math.sign(p.x - prev.x) * r, y: p.y - Math.sign(p.y - prev.y) * r }
    const out = { x: p.x + Math.sign(next.x - p.x) * r, y: p.y + Math.sign(next.y - p.y) * r }
    d += ` L${into.x} ${into.y} Q${p.x} ${p.y} ${out.x} ${out.y}`
  }
  return d
}

export function routeLength(points: XY[]): number {
  let length = 0
  for (let i = 1; i < points.length; i++) length += Math.abs(points[i].x - points[i - 1].x) + Math.abs(points[i].y - points[i - 1].y)
  return length
}

/**
 * Workers with active_tasks replaced by the compiles actually running on
 * them in the (live-merged) task list. The registry's own counter only
 * moves with heartbeats, so on its own it lags the dispatch map by
 * seconds; deriving both from one source keeps every number on the page
 * in step.
 */
export function withLiveActivity(workers: WorkerInfo[], tasks: TaskInfo[]): WorkerInfo[] {
  const running = new Map<string, number>()
  for (const task of tasks) {
    if (task.worker_id && isRunning(task)) running.set(task.worker_id, (running.get(task.worker_id) ?? 0) + 1)
  }
  return workers.map((worker) => ({ ...worker, active_tasks: running.get(worker.id) ?? 0 }))
}
