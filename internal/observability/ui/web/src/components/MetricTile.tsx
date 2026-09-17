import type { ComponentType, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface MetricTileProps {
  label: string
  value: ReactNode
  hint?: ReactNode
  accent?: 'signal' | 'failed' | 'running' | 'neutral'
  icon?: ComponentType<{ className?: string }>
  className?: string
}

const ACCENT_COLOR: Record<NonNullable<MetricTileProps['accent']>, string> = {
  signal: 'var(--color-signal)',
  failed: 'var(--color-status-failed)',
  running: 'var(--color-status-running)',
  neutral: 'var(--color-ink)',
}

/**
 * The hero-number pattern: label demoted (11px/500/muted/tracked),
 * value wins (28px/600, one accent, tabular-nums), hint drops to a
 * lower tier. Same data as a flat "label + value" pair, but decided.
 */
export function MetricTile({ label, value, hint, accent = 'neutral', icon: Icon, className }: MetricTileProps) {
  return (
    <div className={cn('flex flex-col gap-1.5', className)}>
      <span className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">
        {Icon && <Icon className="h-3 w-3" />}
        {label}
      </span>
      <span
        className="font-tabular text-[26px] font-semibold leading-none"
        style={{ color: ACCENT_COLOR[accent] }}
      >
        {value}
      </span>
      {hint && <span className="text-[11px] text-(--color-ink-muted)">{hint}</span>}
    </div>
  )
}
