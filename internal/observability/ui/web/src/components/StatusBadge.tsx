import { useT } from '@/lib/i18n'
import { normalizeStatus, STATUS_COLOR, type StatusTone } from '@/lib/status'
import type { CircuitState } from '@/lib/types'
import { cn } from '@/lib/utils'

interface StatusBadgeProps {
  status: string
  exitCode?: number
  /** Overrides the derived tone, e.g. a build that is running because of in-flight tasks. */
  tone?: StatusTone
  className?: string
}

export function StatusBadge({ status, exitCode, tone, className }: StatusBadgeProps) {
  const t = useT()
  const resolved = tone ?? normalizeStatus(status, exitCode)
  const color = STATUS_COLOR[resolved]

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
        className={cn('h-1.5 w-1.5 rounded-full', resolved === 'running' && 'animate-(--animate-heartbeat)')}
        style={{ backgroundColor: color }}
      />
      {t(`status.${resolved}`)}
    </span>
  )
}

/** Bare status dot for dense rows where a full badge would be noise. */
export function StatusDot({ tone, className }: { tone: StatusTone; className?: string }) {
  const t = useT()
  return (
    <span
      role="img"
      aria-label={t(`status.${tone}`)}
      title={t(`status.${tone}`)}
      className={cn('inline-block h-2 w-2 shrink-0 rounded-full', tone === 'running' && 'animate-(--animate-heartbeat)', className)}
      style={{ backgroundColor: STATUS_COLOR[tone] }}
    />
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
