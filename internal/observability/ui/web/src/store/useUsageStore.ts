import { create } from 'zustand'
import type { UsageSample } from '@/lib/types'
import { mergeUsage } from '@/lib/usage'

interface UsageState {
  /** Recent readings per worker id, oldest first. */
  history: Record<string, UsageSample[]>
  record: (workerId: string, samples: UsageSample[]) => void
}

/**
 * Live CPU/memory history per worker: backfilled from
 * /api/v1/workers/usage, then extended by "worker_usage" socket events.
 */
export const useUsageStore = create<UsageState>((set) => ({
  history: {},
  record: (workerId, samples) =>
    set((state) => {
      const current = state.history[workerId] ?? []
      const next = mergeUsage(current, samples)
      return next === current ? state : { history: { ...state.history, [workerId]: next } }
    }),
}))
