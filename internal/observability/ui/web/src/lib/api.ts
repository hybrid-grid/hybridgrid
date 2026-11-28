import type {
  ArtifactsResponse,
  BuildDetail,
  BuildsResponse,
  WorkerUsageResponse,
  ConsoleOutput,
  Stats,
  TasksResponse,
  WorkersResponse,
} from './types'

/** A non-2xx API response; status lets callers tell "disabled" (503) from failure. */
export class ApiError extends Error {
  readonly status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function getJSON<T>(path: string): Promise<T> {
  const res = await fetch(path, { headers: { Accept: 'application/json' } })
  if (!res.ok) {
    const body = await res.text().catch(() => '')
    throw new ApiError(res.status, `${res.status} ${res.statusText}${body ? `: ${body}` : ''}`)
  }
  return (await res.json()) as T
}

export const api = {
  stats: (): Promise<Stats> => getJSON<Stats>('/api/v1/stats'),
  workers: (): Promise<WorkersResponse> => getJSON<WorkersResponse>('/api/v1/workers'),
  workerUsage: (): Promise<WorkerUsageResponse> => getJSON<WorkerUsageResponse>('/api/v1/workers/usage'),
  tasks: (limit?: number): Promise<TasksResponse> =>
    getJSON<TasksResponse>(`/api/v1/tasks${limit ? `?limit=${limit}` : ''}`),
  builds: (): Promise<BuildsResponse> => getJSON<BuildsResponse>('/api/v1/builds'),
  build: (id: string): Promise<BuildDetail> =>
    getJSON<BuildDetail>(`/api/v1/builds/${encodeURIComponent(id)}`),
  console: (taskId: string): Promise<ConsoleOutput> =>
    getJSON<ConsoleOutput>(`/api/v1/tasks/${encodeURIComponent(taskId)}/console`),
  artifacts: (buildId: string): Promise<ArtifactsResponse> =>
    getJSON<ArtifactsResponse>(`/api/v1/builds/${encodeURIComponent(buildId)}/artifacts`),
}

/** Direct download URL for one stored artifact (served as an attachment). */
export function artifactURL(buildId: string, name: string): string {
  return `/api/v1/builds/${encodeURIComponent(buildId)}/artifacts/${encodeURIComponent(name)}`
}

export function wsURL(): string {
  const proto = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  return `${proto}//${window.location.host}/ws`
}
