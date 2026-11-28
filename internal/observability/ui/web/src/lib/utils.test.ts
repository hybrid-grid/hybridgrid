import { describe, expect, it } from 'vitest'
import { formatDuration, formatSize, tailEllipsis, taskLabel } from './utils'

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

describe('taskLabel', () => {
  it('prefers the source file and falls back to the task id', () => {
    expect(taskLabel({ id: 'task-1', source_file: 'main.c' })).toBe('main.c')
    expect(taskLabel({ id: 'task-2', source_file: '' })).toBe('task-2')
    expect(taskLabel({ id: 'task-3' })).toBe('task-3')
  })
})

describe('tailEllipsis', () => {
  it('keeps the distinguishing tail of long labels', () => {
    expect(tailEllipsis('engine058.cc', 10)).toBe('…ine058.cc')
    expect(tailEllipsis('main.c', 10)).toBe('main.c')
    expect(tailEllipsis('abcdef', 6)).toBe('abcdef')
    expect(tailEllipsis('abcdefg', 6)).toBe('…cdefg')
  })
})

describe('formatSize', () => {
  it.each([
    [0, '0 B'],
    [1023, '1023 B'],
    [1536, '1.5 KiB'],
    [5 * 1024 * 1024, '5.0 MiB'],
    [250 * 1024 * 1024, '250 MiB'],
    [3 * 1024 ** 3, '3.0 GiB'],
  ])('%d → %s', (bytes, want) => {
    expect(formatSize(bytes)).toBe(want)
  })
})
