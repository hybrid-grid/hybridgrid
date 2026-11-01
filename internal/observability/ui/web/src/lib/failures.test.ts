import { describe, expect, it } from 'vitest'
import { errorSignature, groupFailures, isWorkerSpecific, parseLocation } from './failures'
import type { TaskInfo } from './types'

function failed(id: string, overrides: Partial<TaskInfo> = {}): TaskInfo {
  return {
    id,
    build_type: 'cpp',
    build_id: 'b1',
    status: 'failed',
    worker_id: 'w1',
    started_at_ms: 100,
    completed_at_ms: 200,
    queue_ms: 0,
    compile_ms: 100,
    exit_code: 1,
    from_cache: false,
    ...overrides,
  }
}

describe('parseLocation', () => {
  it('extracts file and line from a compiler diagnostic', () => {
    expect(parseLocation("src/a.cc:142:18: error: no member named 'span'")).toEqual({ file: 'src/a.cc', line: 142 })
  })

  it('strips the worker temp directory from raw-source paths', () => {
    const message = "/var/folders/yv/T/hg-worker-task-7bbe3ad0e9ec9efa-6000-2910362472/codec_09.c:3:10: fatal error: 'common.h' file not found"
    expect(parseLocation(message)).toEqual({ file: 'codec_09.c', line: 3 })
  })

  it('returns undefined for messages without a location', () => {
    expect(parseLocation('ld: undefined reference to `foo`')).toBeUndefined()
  })
})

describe('errorSignature', () => {
  it('drops the location so the same error in different files groups together', () => {
    const a = failed('a', { error_message: "src/a.cc:1:2: error: no member named 'span'" })
    const b = failed('b', { error_message: "src/b.cc:9:4:  error: no member named 'span'" })
    expect(errorSignature(a)).toBe(errorSignature(b))
  })

  it('falls back to the exit code when there is no message', () => {
    expect(errorSignature(failed('a', { exit_code: 137 }))).toBe('exit code 137')
  })
})

describe('groupFailures', () => {
  it('ignores successes, counts members, and orders by frequency', () => {
    const tasks = [
      failed('a', { error_message: 'src/a.cc:1:1: error: X', build_id: 'b1' }),
      failed('b', { error_message: 'src/b.cc:2:1: error: X', build_id: 'b2', completed_at_ms: 900 }),
      failed('c', { error_message: 'ld: missing Y' }),
      failed('ok', { status: 'completed', exit_code: undefined }),
    ]
    const groups = groupFailures(tasks)

    expect(groups.map((g) => [g.signature, g.count])).toEqual([
      ['error: X', 2],
      ['ld: missing Y', 1],
    ])
    expect(groups[0].buildIds).toEqual(['b1', 'b2'])
    expect(groups[0].sample.id).toBe('b')
    expect(groups[0].locations.map((l) => l.file)).toEqual(['src/a.cc', 'src/b.cc'])
  })
})

describe('isWorkerSpecific', () => {
  it('flags repeated failures confined to one worker in a multi-worker fleet', () => {
    const [group] = groupFailures([failed('a'), failed('b'), failed('c')])
    expect(isWorkerSpecific(group, 3)).toBe(true)
    expect(isWorkerSpecific(group, 1)).toBe(false)
  })

  it('does not call two failures a pattern', () => {
    const [group] = groupFailures([failed('a'), failed('b')])
    expect(isWorkerSpecific(group, 3)).toBe(false)
  })
})
