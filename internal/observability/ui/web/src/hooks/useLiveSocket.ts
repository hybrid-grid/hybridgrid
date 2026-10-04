import { useQueryClient } from '@tanstack/react-query'
import { useEffect, useRef } from 'react'
import { wsURL } from '@/lib/api'
import { queryKeys } from '@/lib/queries'
import type { WsMessage } from '@/lib/types'
import { useRealtimeStore } from '@/store/useRealtimeStore'

const RECONNECT_DELAY_MS = 2000

/**
 * Owns the single WebSocket connection to the coordinator's telemetry
 * hub. Task lifecycle frames invalidate the matching React Query
 * caches so lists refresh the instant an event lands, instead of
 * waiting for the next poll interval.
 */
export function useLiveSocket(): void {
  const queryClient = useQueryClient()
  const setStatus = useRealtimeStore((state) => state.setStatus)
  const pushEvent = useRealtimeStore((state) => state.pushEvent)
  const socketRef = useRef<WebSocket | null>(null)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    let cancelled = false

    function connect(): void {
      if (cancelled) return
      setStatus('connecting')
      const socket = new WebSocket(wsURL())
      socketRef.current = socket

      socket.onopen = () => setStatus('open')

      socket.onmessage = (event: MessageEvent<string>) => {
        for (const line of event.data.split('\n')) {
          if (!line) continue
          let parsed: WsMessage
          try {
            parsed = JSON.parse(line) as WsMessage
          } catch {
            continue
          }
          if (parsed.type === 'pong') continue
          pushEvent(parsed)

          if (parsed.type === 'stats') {
            queryClient.invalidateQueries({ queryKey: queryKeys.stats })
          }
          if (parsed.type === 'task_started' || parsed.type === 'task_completed') {
            queryClient.invalidateQueries({ queryKey: ['tasks'] })
            queryClient.invalidateQueries({ queryKey: queryKeys.builds })
            queryClient.invalidateQueries({ queryKey: queryKeys.workers })
          }
        }
      }

      socket.onclose = () => {
        setStatus('closed')
        if (!cancelled) {
          reconnectTimer.current = setTimeout(connect, RECONNECT_DELAY_MS)
        }
      }

      socket.onerror = () => socket.close()
    }

    connect()

    return () => {
      cancelled = true
      if (reconnectTimer.current) clearTimeout(reconnectTimer.current)
      socketRef.current?.close()
    }
  }, [queryClient, setStatus, pushEvent])
}
