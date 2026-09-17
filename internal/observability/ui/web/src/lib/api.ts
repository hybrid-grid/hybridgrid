import type {
  BuildDetail,
  BuildsResponse,
  ConsoleOutput,
  Stats,
  TasksResponse,
  WorkersResponse,
} from './types'

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(path, { headers: { Accept: 'application/json' } })
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new Error(`${res.status} ${res.statusText}${body ? `: ${body}` : ''}`)
  }
  return (await res.json()) as T
}

export const api = {
  stats: (): Promise<Stats> => getJSON<Stats>('/api/v1/stats'),
  workers: (): Promise<WorkersResponse> => getJSON<WorkersResponse>('/api/v1/workers'),
  tasks: (limit?: number): Promise<TasksResponse> =>
    getJSON<TasksResponse>(`/api/v1/tasks${limit ? `?limit=${limit}` : ''}`),
  builds: (): Promise<BuildsResponse> => getJSON<BuildsResponse>('/api/v1/builds'),
  build: (id: string): Promise<BuildDetail> =>
    getJSON<BuildDetail>(`/api/v1/builds/${encodeURIComponent(id)}`),
  console: (taskId: string): Promise<ConsoleOutput> =>
    getJSON<ConsoleOutput>(`/api/v1/tasks/${encodeURIComponent(taskId)}/console`),
}

export function wsURL(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/ws`
}
