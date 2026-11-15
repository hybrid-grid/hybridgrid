import { describe, expect, it } from 'vitest'
import { normalizeStatus } from './status'

describe('normalizeStatus', () => {
  it.each([
    ['running', undefined, 'running'],
    ['STATUS_RUNNING', undefined, 'running'],
    ['completed', undefined, 'success'],
    ['completed', 0, 'success'],
    ['completed', 2, 'failed'],
    ['failed', undefined, 'failed'],
    ['STATUS_TIMEOUT', undefined, 'timeout'],
    ['pending', undefined, 'queued'],
    ['', undefined, 'queued'],
    [undefined, 1, 'failed'],
  ] as const)('maps %s (exit %s) to %s', (status, exitCode, tone) => {
    expect(normalizeStatus(status, exitCode)).toBe(tone)
  })
})
