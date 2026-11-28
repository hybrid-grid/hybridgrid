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
 * A row of Paperclip metric cards sharing one bordered container, split
 * by hairlines (a 1px gap over a hairline-colored backdrop). Each cell
 * reads value-first, then label, then hint — Paperclip's MetricCard
 * order — and the lead cell can span wider.
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
        <div key={cell.label} className={cn('flex flex-col bg-(--color-surface) px-5 py-4', cell.lead && 'xl:col-span-2')}>
          <span
            className={cn('font-tabular font-semibold leading-none tracking-[-0.02em]', cell.lead ? 'text-[30px]' : 'text-[24px]')}
            style={{ color: cell.color ?? 'var(--color-ink)' }}
          >
            {cell.value}
          </span>
          <span className="mt-2 text-[13px] font-medium text-(--color-ink-muted)">{cell.label}</span>
          {cell.meter !== undefined && (
            <span className="mt-2 h-1 w-full overflow-hidden rounded-full bg-(--color-surface-inset)">
              <span
                className="block h-full rounded-full transition-[width] duration-500 ease-out"
                style={{ width: `${Math.min(100, Math.max(0, cell.meter * 100))}%`, backgroundColor: cell.color ?? 'var(--color-signal)' }}
              />
            </span>
          )}
          {cell.hint && <span className="mt-1.5 truncate font-tabular text-[12px] text-(--color-ink-faint)">{cell.hint}</span>}
        </div>
      ))}
    </div>
  )
}
