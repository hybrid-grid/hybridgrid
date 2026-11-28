import type { ReactNode } from 'react'

interface PageHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  actions?: ReactNode
}

/**
 * Paperclip page heading: a quiet breadcrumb-like line above a 20px
 * semibold title. Hierarchy comes from size and weight, not color.
 */
export function PageHeader({ eyebrow, title, actions }: PageHeaderProps) {
  return (
    <div className="mb-2 flex items-end justify-between gap-4">
      <div className="min-w-0">
        {eyebrow && <div className="mb-1 truncate text-[12px] font-medium text-(--color-ink-muted)">{eyebrow}</div>}
        <h1 className="truncate text-[20px] font-semibold tracking-[-0.015em] text-(--color-ink)">{title}</h1>
      </div>
      {actions && <div className="flex shrink-0 items-center gap-2">{actions}</div>}
    </div>
  )
}
