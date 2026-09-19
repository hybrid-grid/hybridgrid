function mulberry32(seed: number): () => number {
  let a = seed
  return () => {
    a |= 0
    a = (a + 0x6d2b79f5) | 0
    let t = Math.imul(a ^ (a >>> 15), 1 | a)
    t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296
  }
}

function hashString(value: string): number {
  let h = 0
  for (let i = 0; i < value.length; i += 1) {
    h = (Math.imul(31, h) + value.charCodeAt(i)) | 0
  }
  return h
}

/**
 * Builds one period of an oscilloscope-style trace for `seed`, stable
 * across re-renders so each worker keeps its own waveform "identity"
 * instead of jittering randomly on every poll. `intensity` (0..1) sets
 * how energetic the trace looks, driven by the worker's load ratio.
 */
export function buildWavePoints(seed: string, points = 48, intensity = 0.4): number[] {
  const rand = mulberry32(hashString(seed))
  const baseAmplitude = 0.15 + intensity * 0.75
  const spikeChance = 0.12 + intensity * 0.25
  const values: number[] = []
  let current = 0

  for (let i = 0; i < points; i += 1) {
    const drift = (rand() - 0.5) * baseAmplitude
    const spike = rand() < spikeChance ? (rand() - 0.5) * baseAmplitude * 2.2 : 0
    current = current * 0.55 + (drift + spike) * 0.9
    values.push(Math.max(-1, Math.min(1, current)))
  }
  return values
}

export function toSvgPath(values: number[], width: number, height: number): string {
  if (values.length === 0) return ''
  const midY = height / 2
  const step = width / (values.length - 1)
  return values
    .map((v, i) => {
      const x = i * step
      const y = midY - v * (height / 2 - 2)
      return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`
    })
    .join(' ')
}
