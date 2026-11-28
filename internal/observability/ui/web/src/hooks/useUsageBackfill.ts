import { useQuery } from '@tanstack/react-query'
import { useEffect } from 'react'
import { api } from '@/lib/api'
import { useRealtimeStore } from '@/store/useRealtimeStore'
import { useUsageStore } from '@/store/useUsageStore'

/**
 * Seeds the live CPU/memory history from the coordinator's backlog, so
 * charts are full on first paint instead of growing from empty. Refetched
 * whenever the socket (re)connects, to fill the gap events missed.
 */
export function useUsageBackfill(): void {
  const status = useRealtimeStore((state) => state.status)
  const { data, refetch } = useQuery({
    queryKey: ['workerUsage'],
    queryFn: api.workerUsage,
    staleTime: Infinity,
    retry: 1,
  })

  useEffect(() => {
    if (status === 'open') void refetch()
  }, [status, refetch])

  useEffect(() => {
    if (!data) return
    const { record } = useUsageStore.getState()
    for (const worker of data.workers) record(worker.worker_id, worker.samples)
  }, [data])
}
