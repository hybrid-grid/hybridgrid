import type { WorkerInfo } from './types'

/** The silhouette a worker is drawn as. */
export type MachineKind = 'laptop' | 'desktop' | 'server'

/** Power-LED state, from health first and the circuit breaker second. */
export type PowerTone = 'success' | 'warning' | 'failed'

/** Most slots drawn individually; beyond this the drawing shows "+N". */
export const MAX_DRAWN_SLOTS = 64

/** macOS workers are drawn as laptops, Windows as desktops, the rest as rack servers. */
export function machineKind(os: string | undefined): MachineKind {
  const value = (os ?? '').toLowerCase()
  if (value === 'darwin' || value === 'macos') return 'laptop'
  if (value === 'windows') return 'desktop'
  return 'server'
}

export function osLabel(os: string | undefined): string {
  switch ((os ?? '').toLowerCase()) {
    case 'darwin':
    case 'macos':
      return 'macOS'
    case 'linux':
      return 'Linux'
    case 'windows':
      return 'Windows'
    default:
      return os || '—'
  }
}

/** "ARCH_ARM64" → "arm64", "ARCH_X86_64" → "x86_64"; unknown/unspecified → "". */
export function archLabel(arch: string | undefined): string {
  const value = (arch ?? '').replace(/^ARCH_/, '').toLowerCase()
  return value === 'unspecified' ? '' : value
}

export function powerTone(worker: Pick<WorkerInfo, 'healthy' | 'circuit_state'>): PowerTone {
  if (!worker.healthy) return 'failed'
  const circuit = (worker.circuit_state || 'CLOSED').toUpperCase()
  if (circuit === 'OPEN') return 'failed'
  if (circuit === 'HALF_OPEN') return 'warning'
  return 'success'
}

export interface SlotCell {
  x: number
  y: number
  size: number
  busy: boolean
}

/**
 * Lays `total` square cells out as a centered grid inside a w×h box,
 * choosing the column count that makes the cells largest. The first
 * `busy` cells are marked busy. At most MAX_DRAWN_SLOTS are returned.
 */
export function slotGrid(total: number, busy: number, box: { x: number; y: number; w: number; h: number }): SlotCell[] {
  const n = Math.min(Math.max(total, busy, 0), MAX_DRAWN_SLOTS)
  if (n === 0 || box.w <= 0 || box.h <= 0) return []

  let best = { cols: 1, rows: n, pitch: 0 }
  for (let cols = 1; cols <= n; cols++) {
    const rows = Math.ceil(n / cols)
    const pitch = Math.min(box.w / cols, box.h / rows)
    if (pitch > best.pitch) best = { cols, rows, pitch }
  }

  const { cols, rows, pitch } = best
  const size = pitch * 0.72
  const offsetX = box.x + (box.w - cols * pitch) / 2 + (pitch - size) / 2
  const offsetY = box.y + (box.h - rows * pitch) / 2 + (pitch - size) / 2
  return Array.from({ length: n }, (_, i) => ({
    x: offsetX + (i % cols) * pitch,
    y: offsetY + Math.floor(i / cols) * pitch,
    size,
    busy: i < busy,
  }))
}

/**
 * Splits slots across rack units: one unit per 8 slots, 1–4 units, the
 * remainder filled front to back. Returns the slot count per unit.
 */
export function rackUnits(total: number): number[] {
  const n = Math.min(Math.max(total, 0), MAX_DRAWN_SLOTS)
  const units = Math.min(4, Math.max(1, Math.ceil(n / 8)))
  const base = Math.floor(n / units)
  const extra = n % units
  return Array.from({ length: units }, (_, i) => base + (i < extra ? 1 : 0))
}
