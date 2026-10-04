import type { TaskInfo } from './types'

export interface TimelineRow {
  task: TaskInfo
  queueLeft: number
  queueWidth: number
  runLeft: number
  runWidth: number
}

export interface TimelineLayout {
  windowStart: number
  windowEnd: number
  rows: TimelineRow[]
}

/**
 * Lays out a set of tasks on a shared 0-100% time axis: a queue
 * segment (coordinator dispatch wait) followed by a run segment
 * (actual compile), mirroring the wire fields queue_ms/duration_ms.
 * Tasks with no timestamps are dropped — nothing to plot.
 */
export function layoutTimeline(tasks: TaskInfo[]): TimelineLayout {
  const withTimes = tasks.filter((task) => task.started_at_ms > 0)
  if (withTimes.length === 0) {
    return { windowStart: 0, windowEnd: 0, rows: [] }
  }

  const queueStarts = withTimes.map((task) => task.started_at_ms - Math.max(0, task.queue_ms))
  const ends = withTimes.map((task) => task.completed_at_ms || task.started_at_ms + (task.duration_ms || 0))

  const windowStart = Math.min(...queueStarts)
  const windowEnd = Math.max(windowStart + 1, ...ends)
  const span = windowEnd - windowStart

  const rows: TimelineRow[] = withTimes.map((task) => {
    const queueStart = task.started_at_ms - Math.max(0, task.queue_ms)
    const runEnd = task.completed_at_ms || task.started_at_ms + (task.duration_ms || 0)

    const queueLeft = ((queueStart - windowStart) / span) * 100
    const runLeft = ((task.started_at_ms - windowStart) / span) * 100
    const runRight = ((runEnd - windowStart) / span) * 100

    return {
      task,
      queueLeft: clampPercent(queueLeft),
      queueWidth: clampPercent(runLeft - queueLeft),
      runLeft: clampPercent(runLeft),
      runWidth: Math.max(0.6, clampPercent(runRight - runLeft)),
    }
  })

  return { windowStart, windowEnd, rows }
}

function clampPercent(value: number): number {
  return Math.min(100, Math.max(0, value))
}

export function taskToneClass(status: string): 'success' | 'failed' | 'running' | 'queued' {
  switch (status) {
    case 'STATUS_COMPLETED':
      return 'success'
    case 'STATUS_FAILED':
    case 'STATUS_TIMEOUT':
      return 'failed'
    case 'STATUS_RUNNING':
      return 'running'
    default:
      return 'queued'
  }
}

export const TONE_COLOR: Record<ReturnType<typeof taskToneClass>, string> = {
  success: 'var(--color-status-success)',
  failed: 'var(--color-status-failed)',
  running: 'var(--color-status-running)',
  queued: 'var(--color-status-queued)',
}
