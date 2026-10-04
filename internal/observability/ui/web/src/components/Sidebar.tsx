import { Link } from '@tanstack/react-router'
import { Boxes, GitBranch, LayoutGrid, Radio } from 'lucide-react'
import { useT } from '@/lib/i18n'

const NAV_ITEMS = [
  { to: '/', labelKey: 'nav.overview', icon: LayoutGrid },
  { to: '/builds', labelKey: 'nav.builds', icon: GitBranch },
  { to: '/workers', labelKey: 'nav.workers', icon: Boxes },
] as const

export function Sidebar() {
  const t = useT()

  return (
    <aside className="flex h-full w-[208px] shrink-0 flex-col border-r border-(--color-hairline) bg-(--color-canvas)">
      <div className="flex items-center gap-2 px-4 py-4">
        <Radio className="h-4 w-4 text-(--color-signal)" strokeWidth={2.25} />
        <span className="font-mono text-[12px] font-semibold uppercase tracking-[0.08em] text-(--color-ink)">
          Hybrid-Grid
        </span>
      </div>

      <nav className="flex flex-col gap-0.5 px-2">
        {NAV_ITEMS.map(({ to, labelKey, icon: Icon }) => (
          <Link
            key={to}
            to={to}
            activeOptions={{ exact: to === '/' }}
            className="relative flex items-center gap-2.5 rounded-(--radius-sm) px-2.5 py-1.5 text-[13px] font-medium text-(--color-ink-muted) transition-colors before:absolute before:-left-2 before:h-4 before:w-[2px] before:rounded-full before:bg-(--color-signal) before:opacity-0 before:transition-opacity hover:bg-(--color-canvas-raised) hover:text-(--color-ink)"
            activeProps={{
              className: 'bg-(--color-canvas-raised) text-(--color-ink) before:opacity-100',
            }}
          >
            <Icon className="h-3.5 w-3.5" strokeWidth={2} />
            {t(labelKey)}
          </Link>
        ))}
      </nav>

      <div className="mt-auto px-4 py-3 text-[10px] uppercase tracking-[0.08em] text-(--color-ink-faint)">
        {t('nav.brandSubtitle')}
      </div>
    </aside>
  )
}
