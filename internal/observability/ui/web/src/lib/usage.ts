import type { UsageSample } from './types'

/** Samples kept per worker: five minutes at the workers' 2-second interval. */
export const USAGE_HISTORY = 150
/** A reading older than this is not shown as live (5 missed 2-second reports). */
export const USAGE_STALE_MS = 10_000

/**
 * Appends samples to a worker's history in time order, dropping duplicates
 * (the backfill and live events overlap) and keeping the newest USAGE_HISTORY.
 */
export function mergeUsage(history: UsageSample[], incoming: UsageSample[]): UsageSample[] {
  if (incoming.length === 0) return history
  const byTime = new Map(history.map((s) => [s.at_ms, s]))
  for (const s of incoming) byTime.set(s.at_ms, s)
  const merged = [...byTime.values()].sort((a, b) => a.at_ms - b.at_ms)
  return merged.length > USAGE_HISTORY ? merged.slice(merged.length - USAGE_HISTORY) : merged
}

export function memoryPercent(s: Pick<UsageSample, 'memory_used_bytes' | 'memory_total_bytes'>): number {
  return s.memory_total_bytes > 0 ? Math.min(100, (s.memory_used_bytes / s.memory_total_bytes) * 100) : 0
}

/** True when there is a reading recent enough to call live. */
export function isFresh(atMs: number | undefined, now: number): boolean {
  return Boolean(atMs) && now - (atMs as number) <= USAGE_STALE_MS
}

export type PressureTone = 'ok' | 'warn' | 'hot'

/** How loaded a resource is: amber from 70%, red from 90%. */
export function pressureTone(percent: number): PressureTone {
  if (percent >= 90) return 'hot'
  if (percent >= 70) return 'warn'
  return 'ok'
}
