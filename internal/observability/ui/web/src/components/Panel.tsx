import type { HTMLAttributes, ReactNode } from 'react'
import { cn } from '@/lib/utils'

interface PanelProps extends Omit<HTMLAttributes<HTMLDivElement>, 'title'> {
  title?: ReactNode
  eyebrow?: ReactNode
  action?: ReactNode
  padded?: boolean
}

/**
 * Paperclip card: hairline border, flat surface, 8px radius, no shadow.
 * The header reads title-first with the eyebrow demoted to a muted
 * caption underneath (Paperclip's "Run Activity / Last 14 days"); a
 * panel with only an eyebrow promotes it to the title. The header
 * divider is kept only for flush content (tables, consoles), where the
 * rows need a top edge — padded content flows straight from the header.
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
  const heading = title ?? eyebrow
  const caption = title ? eyebrow : undefined
  return (
    <div
      className={cn('rounded-(--radius-lg) border border-(--color-hairline) bg-(--color-surface)', className)}
      {...props}
    >
      {(heading || action) && (
        <div
          className={cn(
            'flex items-start justify-between gap-3 px-4 pt-3.5',
            padded ? 'pb-0' : 'border-b border-(--color-hairline) pb-3',
          )}
        >
          <div className="min-w-0">
            {heading && <div className="truncate text-[14px] font-medium text-(--color-ink)">{heading}</div>}
            {caption && <div className="mt-0.5 truncate text-[12px] text-(--color-ink-muted)">{caption}</div>}
          </div>
          {action && <div className="-my-1 shrink-0">{action}</div>}
        </div>
      )}
      <div className={padded ? 'p-4' : ''}>{children}</div>
    </div>
  )
}
