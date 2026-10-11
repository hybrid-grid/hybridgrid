import { create } from 'zustand'
import type { Stats, WsMessage } from '@/lib/types'

export type ConnectionStatus = 'connecting' | 'open' | 'closed'

export interface RealtimeEvent {
  id: string
  message: WsMessage
}

/** One stats frame, kept so trend charts have a history the REST API never stores. */
export interface StatsSample {
  at: number
  cacheHitRate: number
  active: number
  queued: number
}

interface RealtimeState {
  status: ConnectionStatus
  events: RealtimeEvent[]
  statsHistory: StatsSample[]
  lastMessageAt: number | null
  setStatus: (status: ConnectionStatus) => void
  pushEvent: (message: WsMessage) => void
  recordStats: (stats: Stats, at: number) => void
  markMessage: (at: number) => void
  clearEvents: () => void
}

const MAX_EVENTS = 100
// The coordinator broadcasts stats every 2s, so 90 samples ≈ 3 minutes.
const MAX_STATS_SAMPLES = 90

let eventSeq = 0

/**
 * The hub replays its recent-event ring on every (re)connect, so the
 * same lifecycle frame can arrive more than once. A task starts and
 * completes once each, so type + task id identifies a frame.
 */
export function eventKey(message: WsMessage): string | undefined {
  const id = (message.data as { id?: unknown } | undefined)?.id
  return typeof id === 'string' ? `${message.type}:${id}` : undefined
}

export const useRealtimeStore = create<RealtimeState>((set) => ({
  status: 'connecting',
  events: [],
  statsHistory: [],
  lastMessageAt: null,
  setStatus: (status) => set({ status }),
  pushEvent: (message) => {
    const key = eventKey(message)
    set((state) => {
      if (key && state.events.some((event) => eventKey(event.message) === key)) return state
      eventSeq += 1
      return {
        events: [{ id: `${message.timestamp}-${eventSeq}`, message }, ...state.events].slice(0, MAX_EVENTS),
      }
    })
  },
  recordStats: (stats, at) =>
    set((state) => ({
      statsHistory: [
        ...state.statsHistory,
        { at, cacheHitRate: stats.cache_hit_rate, active: stats.active_tasks, queued: stats.queued_tasks },
      ].slice(-MAX_STATS_SAMPLES),
    })),
  markMessage: (at) => set({ lastMessageAt: at }),
  clearEvents: () => set({ events: [] }),
}))
