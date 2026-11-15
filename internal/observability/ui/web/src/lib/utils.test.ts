import { describe, expect, it } from 'vitest'
import { formatDuration } from './utils'

describe('formatDuration', () => {
  it.each([
    [0, '0ms'],
    [499, '499ms'],
    [1_500, '1.5s'],
    [59_400, '59.4s'],
    [59_900, '59.9s'],
    [60_000, '1m 0s'],
    [119_600, '2m 0s'],
    [120_000, '2m 0s'],
  ] as const)('formats %dms as %s', (milliseconds, expected) => {
    expect(formatDuration(milliseconds)).toBe(expected)
  })
})
