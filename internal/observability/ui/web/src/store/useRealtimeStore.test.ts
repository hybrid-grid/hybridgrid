import { beforeEach, describe, expect, it } from 'vitest'
import type { WsMessage } from '@/lib/types'
import { useRealtimeStore } from './useRealtimeStore'

function frame(type: WsMessage['type'], id: string): WsMessage {
  return { type, timestamp: 1, data: { id } }
}

describe('useRealtimeStore.pushEvent', () => {
  beforeEach(() => useRealtimeStore.getState().clearEvents())

  it('drops a lifecycle frame replayed after reconnect', () => {
    const { pushEvent } = useRealtimeStore.getState()
    pushEvent(frame('task_started', 't1'))
    pushEvent(frame('task_started', 't1'))

    expect(useRealtimeStore.getState().events).toHaveLength(1)
  })

  it('keeps start and completion of the same task as separate events', () => {
    const { pushEvent } = useRealtimeStore.getState()
    pushEvent(frame('task_started', 't1'))
    pushEvent(frame('task_completed', 't1'))

    expect(useRealtimeStore.getState().events.map((event) => event.message.type)).toEqual([
      'task_completed',
      'task_started',
    ])
  })
})
