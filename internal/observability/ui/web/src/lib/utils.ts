import { type ClassValue, clsx } from 'clsx'
import { twMerge } from 'tailwind-merge'
import type { TaskInfo } from './types'

export function cn(...inputs: ClassValue[]): string {
  return twMerge(clsx(inputs))
}

export function formatDuration(ms: number): string {
  if (ms <= 0) return '0ms'
  if (ms < 1000) return `${Math.round(ms)}ms`
  const seconds = ms / 1000
  if (seconds < 60) return `${seconds.toFixed(1)}s`
  const totalSeconds = Math.round(seconds)
  const minutes = Math.floor(totalSeconds / 60)
  const remainder = totalSeconds % 60
  return `${minutes}m ${remainder}s`
}

export function formatBytes(gb: number): string {
  if (gb < 1) return `${Math.round(gb * 1024)} MB`
  return `${gb.toFixed(1)} GB`
}

export function formatUptime(seconds: number): string {
  if (!seconds || seconds < 60) return `${Math.max(0, Math.round(seconds || 0))}s`
  const minutes = Math.floor(seconds / 60)
  if (minutes < 60) return `${minutes}m`
  const hours = Math.floor(minutes / 60)
  const remMinutes = minutes % 60
  if (hours < 24) return `${hours}h ${remMinutes}m`
  const days = Math.floor(hours / 24)
  const remHours = hours % 24
  return `${days}d ${remHours}h`
}

/** Wall-clock "14:03:27" for a Unix-milliseconds timestamp. */
export function formatClock(ms: number | undefined): string {
  if (!ms) return '—'
  return new Date(ms).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false })
}

export function formatPercent(ratio: number | undefined, digits = 0): string {
  if (ratio === undefined || Number.isNaN(ratio)) return '—'
  return `${(ratio * 100).toFixed(digits)}%`
}

/**
 * The human name for a task: its source file when the coordinator
 * reported one, otherwise the opaque task id (Flutter/Unity tasks,
 * older coordinators).
 */
export function taskLabel(task: Pick<TaskInfo, 'id' | 'source_file'>): string {
  return task.source_file || task.id
}

/**
 * Shortens a label from the front ("engine058.cc" → "…ine058.cc"): in a
 * build's file names the distinguishing part — number, extension — is
 * at the end, while the prefix is usually shared.
 */
export function tailEllipsis(label: string, max: number): string {
  if (max < 2 || label.length <= max) return label
  return `…${label.slice(-(max - 1))}`
}

/** Byte count in binary units: "512 B", "1.5 KiB", "12.3 MiB". */
export function formatSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 1024) return `${Math.max(0, Math.round(bytes || 0))} B`
  const units = ['KiB', 'MiB', 'GiB', 'TiB']
  let value = bytes / 1024
  let unit = 0
  while (value >= 1024 && unit < units.length - 1) {
    value /= 1024
    unit++
  }
  return `${value.toFixed(value >= 100 ? 0 : 1)} ${units[unit]}`
}
