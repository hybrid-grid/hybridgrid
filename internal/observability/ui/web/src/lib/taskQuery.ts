import { normalizeStatus, type StatusTone } from './status'
import type { TaskInfo } from './types'

export type TaskStatusFilter = 'all' | 'running' | 'success' | 'failed'
export type TaskCacheFilter = 'all' | 'hit' | 'miss'
export type TaskSortKey = 'started' | 'duration' | 'queue' | 'compile'
export type SortDir = 'asc' | 'desc'

export interface TaskQuery {
  q?: string
  status?: TaskStatusFilter
  worker?: string
  cache?: TaskCacheFilter
  sort?: TaskSortKey
  dir?: SortDir
}

const FAILED_TONES: StatusTone[] = ['failed', 'timeout']

function matchesStatus(task: TaskInfo, status: TaskStatusFilter): boolean {
  if (status === 'all') return true
  const tone = normalizeStatus(task.status, task.exit_code)
  return status === 'failed' ? FAILED_TONES.includes(tone) : tone === status
}

function matchesQuery(task: TaskInfo, q: string): boolean {
  const needle = q.trim().toLowerCase()
  if (!needle) return true
  return [task.id, task.build_id, task.worker_id, task.error_message ?? ''].some((field) =>
    field.toLowerCase().includes(needle),
  )
}

function sortValue(task: TaskInfo, key: TaskSortKey): number {
  switch (key) {
    case 'duration':
      return task.duration_ms ?? -1
    case 'queue':
      return task.queue_ms
    case 'compile':
      return task.compile_ms
    default:
      return task.started_at_ms
  }
}

/** Filters then sorts without mutating the input — the query cache owns that array. */
export function applyTaskQuery(tasks: TaskInfo[], query: TaskQuery): TaskInfo[] {
  const { q = '', status = 'all', worker, cache = 'all', sort = 'started', dir = 'desc' } = query
  const sign = dir === 'asc' ? 1 : -1
  return tasks
    .filter(
      (task) =>
        matchesStatus(task, status) &&
        matchesQuery(task, q) &&
        (!worker || task.worker_id === worker) &&
        (cache === 'all' || task.from_cache === (cache === 'hit')),
    )
    .toSorted((a, b) => sign * (sortValue(a, sort) - sortValue(b, sort)))
}

export function paginate<T>(items: T[], page: number, pageSize: number): { rows: T[]; page: number; pageCount: number } {
  const pageCount = Math.max(1, Math.ceil(items.length / pageSize))
  const current = Math.min(Math.max(1, page), pageCount)
  return { rows: items.slice((current - 1) * pageSize, current * pageSize), page: current, pageCount }
}
