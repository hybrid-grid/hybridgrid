import type { ReactNode } from 'react'
import { cn } from '@/lib/utils'

export interface StatCell {
  label: string
  value: ReactNode
  hint?: ReactNode
  /** CSS color for the value; omit for neutral ink. */
  color?: string
  /** 0..1 — renders a hairline meter under the value. */
  meter?: number
  /** Marks the strip's lead figure: larger, first to be read. */
  lead?: boolean
}

interface StatStripProps {
  cells: StatCell[]
  className?: string
}

/**
 * One continuous instrument panel instead of a row of identical boxed
 * cards: cells share a single border and are split by 1px hairlines
 * (a gap over a hairline-colored backdrop), so the strip reads as one
 * readout and the lead cell can be wider and louder than the rest.
 */
export function StatStrip({ cells, className }: StatStripProps) {
  return (
    <div
      className={cn(
        'grid grid-cols-2 gap-px overflow-hidden rounded-(--radius-lg) border border-(--color-hairline) bg-(--color-hairline) sm:grid-cols-3 xl:grid-flow-col xl:auto-cols-fr xl:grid-cols-none',
        className,
      )}
    >
      {cells.map((cell) => (
        <div key={cell.label} className={cn('flex flex-col gap-1.5 bg-(--color-surface) px-4 py-3.5', cell.lead && 'xl:col-span-2')}>
          <span className="text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{cell.label}</span>
          <span
            className={cn('font-tabular font-semibold leading-none tracking-[-0.01em]', cell.lead ? 'text-[30px]' : 'text-[22px]')}
            style={{ color: cell.color ?? 'var(--color-ink)' }}
          >
            {cell.value}
          </span>
          {cell.meter !== undefined && (
            <span className="mt-0.5 h-1 w-full overflow-hidden rounded-full bg-(--color-surface-inset)">
              <span
                className="block h-full rounded-full transition-[width] duration-500 ease-out"
                style={{ width: `${Math.min(100, Math.max(0, cell.meter * 100))}%`, backgroundColor: cell.color ?? 'var(--color-signal)' }}
              />
            </span>
          )}
          {cell.hint && <span className="truncate font-tabular text-[11px] text-(--color-ink-muted)">{cell.hint}</span>}
        </div>
      ))}
    </div>
  )
}
