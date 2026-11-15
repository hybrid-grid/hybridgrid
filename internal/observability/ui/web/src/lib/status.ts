/**
 * The coordinator emits lowercase task/build states ("running",
 * "completed", "failed"), older fixtures and the proto enum use the
 * STATUS_* spelling, and exit_code is omitempty (absent == success).
 * Every status-colored surface goes through this one normalizer so a
 * badge, a timeline bar and a chart bar can never disagree.
 */
export type StatusTone = 'success' | 'failed' | 'running' | 'queued' | 'timeout'

export function normalizeStatus(status: string | undefined, exitCode?: number): StatusTone {
  switch ((status ?? '').toLowerCase()) {
    case 'running':
    case 'status_running':
      return 'running'
    case 'queued':
    case 'pending':
    case 'status_queued':
      return 'queued'
    case 'timeout':
    case 'status_timeout':
      return 'timeout'
    case 'failed':
    case 'status_failed':
      return 'failed'
    case 'completed':
    case 'success':
    case 'status_completed':
      return exitCode ? 'failed' : 'success'
    default:
      return exitCode ? 'failed' : 'queued'
  }
}

export const STATUS_COLOR: Record<StatusTone, string> = {
  success: 'var(--color-status-success)',
  failed: 'var(--color-status-failed)',
  running: 'var(--color-status-running)',
  queued: 'var(--color-status-queued)',
  timeout: 'var(--color-status-timeout)',
}
