import { describe, expect, it } from 'vitest'
import { MAX_DRAWN_SLOTS, archLabel, machineKind, osLabel, powerTone, rackUnits, slotGrid } from './machine'

describe('machineKind / osLabel / archLabel', () => {
  it.each([
    ['darwin', 'laptop', 'macOS'],
    ['windows', 'desktop', 'Windows'],
    ['linux', 'server', 'Linux'],
    ['', 'server', '—'],
  ] as const)('%s → %s (%s)', (os, kind, label) => {
    expect(machineKind(os)).toBe(kind)
    expect(osLabel(os)).toBe(label)
  })

  it('shortens architecture enums', () => {
    expect(archLabel('ARCH_ARM64')).toBe('arm64')
    expect(archLabel('ARCH_X86_64')).toBe('x86_64')
    expect(archLabel('ARCH_UNSPECIFIED')).toBe('')
    expect(archLabel(undefined)).toBe('')
  })
})

describe('powerTone', () => {
  it('puts health ahead of the circuit state', () => {
    expect(powerTone({ healthy: true, circuit_state: 'CLOSED' })).toBe('success')
    expect(powerTone({ healthy: true, circuit_state: '' })).toBe('success')
    expect(powerTone({ healthy: true, circuit_state: 'HALF_OPEN' })).toBe('warning')
    expect(powerTone({ healthy: true, circuit_state: 'OPEN' })).toBe('failed')
    expect(powerTone({ healthy: false, circuit_state: 'CLOSED' })).toBe('failed')
  })
})

describe('slotGrid', () => {
  const box = { x: 0, y: 0, w: 120, h: 60 }

  it('returns one cell per slot, busy ones first, inside the box', () => {
    const cells = slotGrid(12, 5, box)
    expect(cells).toHaveLength(12)
    expect(cells.filter((c) => c.busy)).toHaveLength(5)
    expect(cells.slice(0, 5).every((c) => c.busy)).toBe(true)
    for (const c of cells) {
      expect(c.x).toBeGreaterThanOrEqual(0)
      expect(c.y).toBeGreaterThanOrEqual(0)
      expect(c.x + c.size).toBeLessThanOrEqual(120)
      expect(c.y + c.size).toBeLessThanOrEqual(60)
    }
  })

  it('draws busy slots even when they exceed the advertised max', () => {
    expect(slotGrid(2, 3, box)).toHaveLength(3)
  })

  it('caps the drawing and handles empty input', () => {
    expect(slotGrid(500, 0, box)).toHaveLength(MAX_DRAWN_SLOTS)
    expect(slotGrid(0, 0, box)).toEqual([])
  })
})

describe('rackUnits', () => {
  it('spreads slots over 1–4 units', () => {
    expect(rackUnits(6)).toEqual([6])
    expect(rackUnits(12)).toEqual([6, 6])
    expect(rackUnits(17)).toEqual([6, 6, 5])
    expect(rackUnits(64)).toEqual([16, 16, 16, 16])
    expect(rackUnits(0)).toEqual([0])
  })
})
