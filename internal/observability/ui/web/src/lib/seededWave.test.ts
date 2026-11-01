import { describe, expect, it } from 'vitest'
import { buildWavePoints, toPeriodicSvgPath } from './seededWave'

function coords(path: string): [number, number][] {
  return [...path.matchAll(/[ML](-?[\d.]+),(-?[\d.]+)/g)].map((m) => [Number(m[1]), Number(m[2])])
}

describe('toPeriodicSvgPath', () => {
  it('ends at the same height it starts so tiled copies join without a seam', () => {
    const points = coords(toPeriodicSvgPath(buildWavePoints('worker-a', 40, 0.7), 240, 40))
    expect(points[0][0]).toBe(0)
    expect(points.at(-1)?.[0]).toBe(240)
    expect(points.at(-1)?.[1]).toBe(points[0][1])
  })

  it('returns an empty path for no samples', () => {
    expect(toPeriodicSvgPath([], 240, 40)).toBe('')
  })
})
