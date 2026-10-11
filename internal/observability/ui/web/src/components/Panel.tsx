import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface PanelProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: ReactNode
  eyebrow?: ReactNode
  action?: ReactNode
  padded?: boolean
}

/**
 * Borders-only elevation, per the control-room direction: shadows read
 * as almost nothing on a dark canvas, so structure comes from a
 * hairline + a one-step-lighter surface instead.
 */
export function Panel({
  title,
  eyebrow,
  action,
  padded = true,
  className,
  children,
  ...props
}: PanelProps) {
  return (
    <div
      className={cn(
        'rounded-(--radius-lg) border border-(--color-hairline) bg-(--color-surface)',
        className,
      )}
      {...props}
    >
      {(title || eyebrow || action) && (
        <div className="flex items-center justify-between border-b border-(--color-hairline) px-4 py-2.5">
          <div>
            {eyebrow && (
              <div className="text-[10px] font-semibold uppercase tracking-[0.08em] text-(--color-ink-faint)">
                {eyebrow}
              </div>
            )}
            {title && <div className="text-[13px] font-semibold text-(--color-ink)">{title}</div>}
          </div>
          {action}
        </div>
      )}
      <div className={padded ? 'p-4' : ''}>{children}</div>
    </div>
  )
}
