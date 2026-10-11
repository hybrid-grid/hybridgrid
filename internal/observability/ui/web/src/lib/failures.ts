import { normalizeStatus } from './status'
import type { TaskInfo } from './types'

// "src/io/fs_watch.cc:142:18: " — the location prefix compilers put in front of a diagnostic.
const LOCATION_PREFIX = /^\s*((?:[A-Za-z]:)?[^\s:]+):(\d+)(?::\d+)?:\s*/

export interface FailureLocation {
  file: string
  line: number
}

export interface FailureGroup {
  signature: string
  sample: TaskInfo
  count: number
  buildIds: string[]
  /** Failures per worker, most affected first. */
  workers: { workerId: string; count: number }[]
  locations: FailureLocation[]
  lastAtMs: number
}

export function isFailedTask(task: TaskInfo): boolean {
  const tone = normalizeStatus(task.status, task.exit_code)
  return tone === 'failed' || tone === 'timeout'
}

export function parseLocation(message: string): FailureLocation | undefined {
  const match = LOCATION_PREFIX.exec(message)
  return match ? { file: match[1], line: Number(match[2]) } : undefined
}

/**
 * Collapses one failure message into a grouping key: the location
 * prefix is dropped (the same missing header fails in forty files) and
 * whitespace is normalized. Failures without a message group by exit
 * code, so "killed by OOM (137)" still clusters.
 */
export function errorSignature(task: TaskInfo): string {
  const message = task.error_message?.split('\n')[0]?.trim()
  if (!message) return `exit code ${task.exit_code ?? 'unknown'}`
  return message.replace(LOCATION_PREFIX, '').replace(/\s+/g, ' ').trim()
}

function taskTime(task: TaskInfo): number {
  return task.completed_at_ms || task.started_at_ms || 0
}

export function groupFailures(tasks: TaskInfo[]): FailureGroup[] {
  const groups = new Map<string, TaskInfo[]>()
  for (const task of tasks.filter(isFailedTask)) {
    const signature = errorSignature(task)
    groups.set(signature, [...(groups.get(signature) ?? []), task])
  }

  return [...groups.entries()]
    .map(([signature, members]): FailureGroup => {
      const newest = members.reduce((a, b) => (taskTime(b) > taskTime(a) ? b : a))
      const workerCounts = new Map<string, number>()
      for (const task of members) {
        const id = task.worker_id || ''
        workerCounts.set(id, (workerCounts.get(id) ?? 0) + 1)
      }
      const locations = members
        .map((task) => (task.error_message ? parseLocation(task.error_message) : undefined))
        .filter((loc): loc is FailureLocation => loc !== undefined)
      const uniqueLocations = [...new Map(locations.map((loc) => [`${loc.file}:${loc.line}`, loc])).values()]

      return {
        signature,
        sample: newest,
        count: members.length,
        buildIds: [...new Set(members.map((task) => task.build_id).filter(Boolean))],
        workers: [...workerCounts.entries()]
          .map(([workerId, count]) => ({ workerId, count }))
          .sort((a, b) => b.count - a.count),
        locations: uniqueLocations,
        lastAtMs: taskTime(newest),
      }
    })
    .sort((a, b) => b.count - a.count || b.lastAtMs - a.lastAtMs)
}

// Two failures on one box is coincidence often enough; three starts to be a pattern.
const WORKER_SPECIFIC_MIN_COUNT = 3

/**
 * True when every failure in the group landed on one worker while the
 * fleet has more than one — the "is it the machine or the code" tell.
 */
export function isWorkerSpecific(group: FailureGroup, workerCount: number): boolean {
  return (
    workerCount > 1 &&
    group.count >= WORKER_SPECIFIC_MIN_COUNT &&
    group.workers.length === 1 &&
    group.workers[0].workerId !== ''
  )
}
