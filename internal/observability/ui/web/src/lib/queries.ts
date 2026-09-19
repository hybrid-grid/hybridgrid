import { useQuery } from '@tanstack/react-query'
import { api } from './api'

export const queryKeys = {
  stats: ['stats'] as const,
  workers: ['workers'] as const,
  tasks: (limit?: number) => ['tasks', limit ?? null] as const,
  builds: ['builds'] as const,
  build: (id: string) => ['builds', id] as const,
  console: (taskId: string) => ['console', taskId] as const,
}

export function useStats() {
  return useQuery({
    queryKey: queryKeys.stats,
    queryFn: api.stats,
    refetchInterval: 8000,
  })
}

export function useWorkers() {
  return useQuery({
    queryKey: queryKeys.workers,
    queryFn: api.workers,
    refetchInterval: 5000,
  })
}

export function useTasks(limit?: number) {
  return useQuery({
    queryKey: queryKeys.tasks(limit),
    queryFn: () => api.tasks(limit),
    refetchInterval: 6000,
  })
}

export function useBuilds() {
  return useQuery({
    queryKey: queryKeys.builds,
    queryFn: api.builds,
    refetchInterval: 6000,
  })
}

export function useBuildDetail(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.build(id ?? ''),
    queryFn: () => api.build(id as string),
    enabled: Boolean(id),
    refetchInterval: 5000,
  })
}

export function useConsole(taskId: string | undefined) {
  return useQuery({
    queryKey: queryKeys.console(taskId ?? ''),
    queryFn: () => api.console(taskId as string),
    enabled: Boolean(taskId),
    refetchInterval: 4000,
  })
}
