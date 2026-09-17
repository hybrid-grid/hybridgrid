import { useT } from '@/lib/i18n'
import type { CircuitState, TaskStatus } from '@/lib/types'
import { cn } from '@/lib/utils'

const STATUS_COLOR: Record<string, string> = {
  STATUS_QUEUED: 'var(--color-status-queued)',
  STATUS_RUNNING: 'var(--color-status-running)',
  STATUS_COMPLETED: 'var(--color-status-success)',
  STATUS_FAILED: 'var(--color-status-failed)',
  STATUS_TIMEOUT: 'var(--color-status-timeout)',
}

interface StatusBadgeProps {
  status: TaskStatus
  className?: string
}

export function StatusBadge({ status, className }: StatusBadgeProps) {
  const t = useT()
  const color = STATUS_COLOR[status] ?? 'var(--color-ink-faint)'
  const label = t(`status.${status}`) === `status.${status}` ? status.replace('STATUS_', '') : t(`status.${status}`)
  const isRunning = status === 'STATUS_RUNNING'

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-(--radius-xs) border px-1.5 py-0.5 text-[11px] font-medium',
        className,
      )}
      style={{
        color,
        borderColor: `color-mix(in oklch, ${color} 35%, transparent)`,
        backgroundColor: `color-mix(in oklch, ${color} 12%, transparent)`,
      }}
    >
      <span
        className={cn('h-1.5 w-1.5 rounded-full', isRunning && 'animate-(--animate-heartbeat)')}
        style={{ backgroundColor: color }}
      />
      {label}
    </span>
  )
}

const CIRCUIT_COLOR: Record<CircuitState, string> = {
  CLOSED: 'var(--color-circuit-closed)',
  HALF_OPEN: 'var(--color-circuit-half-open)',
  OPEN: 'var(--color-circuit-open)',
}

interface CircuitBadgeProps {
  state: string
  className?: string
}

export function CircuitBadge({ state, className }: CircuitBadgeProps) {
  const t = useT()
  const normalized = (state || 'CLOSED') as CircuitState
  const color = CIRCUIT_COLOR[normalized] ?? 'var(--color-ink-faint)'
  const label = t(`circuit.${normalized}`)

  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-(--radius-xs) px-1.5 py-0.5 text-[11px] font-mono font-medium uppercase tracking-wide',
        className,
      )}
      style={{ color }}
      title={`Circuit breaker: ${label}`}
    >
      <span className="relative flex h-1.5 w-1.5">
        {normalized === 'OPEN' && (
          <span
            className="absolute inline-flex h-full w-full animate-(--animate-pulse-ring) rounded-full"
            style={{ backgroundColor: color }}
          />
        )}
        <span className="relative inline-flex h-1.5 w-1.5 rounded-full" style={{ backgroundColor: color }} />
      </span>
      {label}
    </span>
  )
}
