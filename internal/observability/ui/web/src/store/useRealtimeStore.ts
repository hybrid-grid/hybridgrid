import { create } from 'zustand'
import type { WsMessage } from '@/lib/types'

export type ConnectionStatus = 'connecting' | 'open' | 'closed'

export interface RealtimeEvent {
  id: string
  message: WsMessage
}

interface RealtimeState {
  status: ConnectionStatus
  events: RealtimeEvent[]
  setStatus: (status: ConnectionStatus) => void
  pushEvent: (message: WsMessage) => void
  clearEvents: () => void
}

const MAX_EVENTS = 40

export const useRealtimeStore = create<RealtimeState>((set) => ({
  status: 'connecting',
  events: [],
  setStatus: (status) => set({ status }),
  pushEvent: (message) =>
    set((state) => ({
      events: [
        { id: `${message.timestamp}-${state.events.length}-${Math.random().toString(36).slice(2, 8)}`, message },
        ...state.events,
      ].slice(0, MAX_EVENTS),
    })),
  clearEvents: () => set({ events: [] }),
}))
