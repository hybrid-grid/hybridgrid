import { Link } from '@tanstack/react-router'
import { Boxes, Database, GitBranch, LayoutGrid, ListChecks, OctagonAlert, Radio } from 'lucide-react'
import { useT } from '@/lib/i18n'

const NAV_ITEMS = [
  { to: '/', labelKey: 'nav.overview', icon: LayoutGrid },
  { to: '/builds', labelKey: 'nav.builds', icon: GitBranch },
  { to: '/tasks', labelKey: 'nav.tasks', icon: ListChecks },
  { to: '/failures', labelKey: 'nav.failures', icon: OctagonAlert },
  { to: '/workers', labelKey: 'nav.workers', icon: Boxes },
  { to: '/cache', labelKey: 'nav.cache', icon: Database },
] as const

export function Sidebar() {
  const t = useT()

  return (
    <aside className="hidden h-full w-[208px] md:flex shrink-0 flex-col border-r border-(--color-hairline) bg-(--color-canvas)">
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

/** Below md the sidebar collapses into this icon row inside the header. */
export function MobileNav() {
  const t = useT()

  return (
    <nav className="flex items-center gap-0.5 md:hidden">
      <Radio className="mr-2 h-4 w-4 text-(--color-signal)" strokeWidth={2.25} aria-hidden />
      {NAV_ITEMS.map(({ to, labelKey, icon: Icon }) => (
        <Link
          key={to}
          to={to}
          activeOptions={{ exact: to === '/' }}
          aria-label={t(labelKey)}
          title={t(labelKey)}
          className="flex h-8 w-8 items-center justify-center rounded-(--radius-sm) text-(--color-ink-muted) transition-colors hover:bg-(--color-canvas-raised) hover:text-(--color-ink)"
          activeProps={{ className: 'bg-(--color-canvas-raised) text-(--color-signal)' }}
        >
          <Icon className="h-4 w-4" strokeWidth={2} />
        </Link>
      ))}
    </nav>
  )
}
