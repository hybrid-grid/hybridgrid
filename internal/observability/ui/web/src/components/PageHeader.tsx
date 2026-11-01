import type { ReactNode } from 'react'

interface PageHeaderProps {
  eyebrow?: ReactNode
  title: ReactNode
  actions?: ReactNode
}

export function PageHeader({ eyebrow, title, actions }: PageHeaderProps) {
  return (
    <div className="mb-5 flex items-center justify-between">
      <div>
        {eyebrow && (
          <div className="text-[10px] font-semibold uppercase tracking-[0.08em] text-(--color-ink-faint)">
            {eyebrow}
          </div>
        )}
        <h1 className="text-[18px] font-semibold tracking-[-0.01em] text-(--color-ink)">{title}</h1>
      </div>
      {actions && <div className="flex items-center gap-2">{actions}</div>}
    </div>
  )
}
