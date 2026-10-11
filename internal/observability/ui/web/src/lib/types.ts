// Wire types mirror internal/telemetry/types.go byte-for-byte. Do not
// rename fields without updating the Go struct tags first.

export type CircuitState = 'CLOSED' | 'HALF_OPEN' | 'OPEN'

export type TaskStatus =
  | 'STATUS_QUEUED'
  | 'STATUS_RUNNING'
  | 'STATUS_COMPLETED'
  | 'STATUS_FAILED'
  | 'STATUS_TIMEOUT'
  | string

export interface Stats {
  total_tasks: number
  success_tasks: number
  failed_tasks: number
  active_tasks: number
  queued_tasks: number
  cache_hits: number
  cache_misses: number
  cache_hit_rate: number
  flutter_builds: number
  flutter_cache_hits: number
  flutter_cache_misses: number
  flutter_cache_hit_rate: number
  unity_builds: number
  unity_cache_hits: number
  unity_cache_misses: number
  unity_cache_hit_rate: number
  total_workers: number
  healthy_workers: number
  uptime_seconds: number
  timestamp: number
}

export interface WorkerInfo {
  id: string
  host: string
  address: string
  os: string
  architecture: string
  architectures: string[] | null
  cpu_cores: number
  memory_gb: number
  max_parallel_tasks: number
  active_tasks: number
  total_tasks: number
  success_rate: number
  avg_latency_ms: number
  circuit_state: CircuitState | string
  discovery_source: string
  version: string
  docker_available: boolean
  flutter_available: boolean
  flutter_sdk_version: string
  flutter_platforms: string[] | null
  unity_available: boolean
  unity_versions: string[] | null
  unity_platforms: string[] | null
  compilers: string[] | null
  build_types: string[] | null
  healthy: boolean
  last_seen: number
}

export interface TaskInfo {
  id: string
  build_type: string
  build_id: string
  status: TaskStatus
  worker_id: string
  started_at_ms: number
  completed_at_ms?: number
  duration_ms?: number
  queue_ms: number
  compile_ms: number
  exit_code?: number
  from_cache: boolean
  error_message?: string
}

export interface BuildInfo {
  id: string
  build_type: string
  status: TaskStatus
  total_tasks: number
  completed_tasks: number
  failed_tasks: number
  running_tasks: number
  from_cache_count: number
  first_task_at_ms: number
  last_task_at_ms: number
  truncated: boolean
}

export interface BuildDetail {
  build: BuildInfo
  tasks: TaskInfo[]
}

export interface ConsoleOutput {
  task_id: string
  stdout: string
  stderr: string
  truncated: boolean
}

interface ListEnvelope {
  count: number
  timestamp: number
}

export interface WorkersResponse extends ListEnvelope {
  workers: WorkerInfo[]
}

export interface TasksResponse extends ListEnvelope {
  tasks: TaskInfo[]
}

export interface BuildsResponse extends ListEnvelope {
  builds: BuildInfo[]
}

export type WsMessageType = 'stats' | 'task_started' | 'task_completed' | 'ping' | 'pong'

export interface WsMessage<T = unknown> {
  type: WsMessageType
  timestamp: number
  data?: T
}
