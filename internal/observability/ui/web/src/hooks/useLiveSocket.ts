import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { wsURL } from '@/lib/api'
import { queryKeys } from '@/lib/queries'
import type { Stats, WsMessage } from '@/lib/types'
import { useRealtimeStore } from '@/store/useRealtimeStore'

const RECONNECT_BASE_MS = 1000
const RECONNECT_MAX_MS = 30_000
// A -j32 build emits dozens of task frames per second; refetching the
// task/build/worker lists on every one would hammer the dashboard API.
const INVALIDATE_THROTTLE_MS = 1000

/**
 * Owns the single WebSocket connection to the telemetry hub. Stats
 * frames are written straight into the React Query cache (no extra
 * request) and sampled into a short history for trend charts; task
 * frames feed the event log and trigger a throttled list refresh.
 */
export function useLiveSocket(): void {
  const queryClient = useQueryClient()
  const socketRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    const { setStatus, pushEvent, recordStats, markMessage } = useRealtimeStore.getState()
    let cancelled = false
    let attempts = 0
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null
    let invalidateTimer: ReturnType<typeof setTimeout> | null = null

    function scheduleListRefresh(): void {
      if (invalidateTimer) return
      invalidateTimer = setTimeout(() => {
        invalidateTimer = null
        queryClient.invalidateQueries({ queryKey: ['tasks'] })
        queryClient.invalidateQueries({ queryKey: queryKeys.builds })
        queryClient.invalidateQueries({ queryKey: queryKeys.workers })
      }, INVALIDATE_THROTTLE_MS)
    }

    function handle(message: WsMessage): void {
      markMessage(Date.now())
      if (message.type === 'stats' && message.data) {
        const stats = message.data as Stats
        queryClient.setQueryData(queryKeys.stats, stats)
        recordStats(stats, Date.now())
        return
      }
      if (message.type === 'task_started' || message.type === 'task_completed') {
        pushEvent(message)
        scheduleListRefresh()
      }
    }

    function connect(): void {
      if (cancelled) return
      setStatus('connecting')
      const socket = new WebSocket(wsURL())
      socketRef.current = socket

      socket.onopen = () => {
        attempts = 0
        setStatus('open')
      }

      socket.onmessage = (event: MessageEvent<string>) => {
        for (const line of event.data.split('\n')) {
          if (!line) continue
          try {
            handle(JSON.parse(line) as WsMessage)
          } catch {
            // Malformed frame: skip it, the next poll heals any gap.
          }
        }
      }

      socket.onclose = () => {
        setStatus('closed')
        if (cancelled) return
        const delay = Math.min(RECONNECT_BASE_MS * 2 ** attempts, RECONNECT_MAX_MS)
        attempts += 1
        reconnectTimer = setTimeout(connect, delay)
      }

      socket.onerror = () => socket.close()
    }

    connect()

    return () => {
      cancelled = true
      if (reconnectTimer) clearTimeout(reconnectTimer)
      if (invalidateTimer) clearTimeout(invalidateTimer)
      socketRef.current?.close()
    }
  }, [queryClient])
}
