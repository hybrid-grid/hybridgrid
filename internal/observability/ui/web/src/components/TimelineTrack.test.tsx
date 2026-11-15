import { renderToStaticMarkup } from 'react-dom/server'
import { describe, expect, it } from 'vitest'
import { TimelineTrack } from './TimelineTrack'
import type { TimelineRow } from '@/lib/timeline'

const row: TimelineRow = {
  task: {
    id: 'task-42',
    build_type: 'cpp',
    build_id: 'build-1',
    status: 'failed',
    worker_id: 'worker-1',
    started_at_ms: 100,
    completed_at_ms: 200,
    duration_ms: 100,
    queue_ms: 0,
    compile_ms: 100,
    from_cache: false,
  },
  queueLeft: 0,
  queueWidth: 0,
  runLeft: 0,
  runWidth: 100,
}

describe('TimelineTrack', () => {
  it('gives each task button an accessible name', () => {
    const markup = renderToStaticMarkup(<TimelineTrack rows={[row]} />)

    expect(markup).toContain('aria-label="Task task-42 · failed"')
  })
})
