import { describe, expect, it } from 'vitest'
import type { UsageSample } from './types'
import { USAGE_HISTORY, USAGE_STALE_MS, isFresh, memoryPercent, mergeUsage, pressureTone } from './usage'

const s = (at: number, cpu = 0): UsageSample => ({ at_ms: at, cpu_percent: cpu, memory_used_bytes: 0, memory_total_bytes: 0, load1: 0 })

describe('mergeUsage', () => {
  it('merges in time order and lets a later copy of a sample win', () => {
    const merged = mergeUsage([s(1), s(3)], [s(2), s(3, 50)])
    expect(merged.map((x) => x.at_ms)).toEqual([1, 2, 3])
    expect(merged[2].cpu_percent).toBe(50)
  })

  it('keeps only the newest USAGE_HISTORY samples', () => {
    const many = Array.from({ length: USAGE_HISTORY + 5 }, (_, i) => s(i))
    const merged = mergeUsage([], many)
    expect(merged).toHaveLength(USAGE_HISTORY)
    expect(merged[0].at_ms).toBe(5)
  })

  it('returns the same array when nothing arrives', () => {
    const history = [s(1)]
    expect(mergeUsage(history, [])).toBe(history)
  })
})

describe('memoryPercent / isFresh / pressureTone', () => {
  it('computes memory share and guards a zero total', () => {
    expect(memoryPercent({ memory_used_bytes: 4, memory_total_bytes: 16 })).toBe(25)
    expect(memoryPercent({ memory_used_bytes: 4, memory_total_bytes: 0 })).toBe(0)
  })

  it('treats missing or old readings as not live', () => {
    expect(isFresh(0, 1000)).toBe(false)
    expect(isFresh(undefined, 1000)).toBe(false)
    expect(isFresh(1000, 1000 + USAGE_STALE_MS)).toBe(true)
    expect(isFresh(1000, 1001 + USAGE_STALE_MS)).toBe(false)
  })

  it.each([
    [10, 'ok'],
    [70, 'warn'],
    [89.9, 'warn'],
    [90, 'hot'],
  ] as const)('%d%% → %s', (pct, tone) => {
    expect(pressureTone(pct)).toBe(tone)
  })
})
