import { Link } from '@tanstack/react-router'
import { Boxes, Database, GitBranch, LayoutGrid, ListChecks, OctagonAlert, Radio } from 'lucide-react'
import { useT } from '@/lib/i18n'
import { cn } from '@/lib/utils'

type NavItem = { to: string; labelKey: string; icon: typeof LayoutGrid }

/** Paperclip-style grouping: a top-level home, then labelled sections. */
const NAV_GROUPS: { labelKey?: string; items: NavItem[] }[] = [
  { items: [{ to: '/', labelKey: 'nav.overview', icon: LayoutGrid }] },
  {
    labelKey: 'nav.groupBuilds',
    items: [
      { to: '/builds', labelKey: 'nav.builds', icon: GitBranch },
      { to: '/tasks', labelKey: 'nav.tasks', icon: ListChecks },
      { to: '/failures', labelKey: 'nav.failures', icon: OctagonAlert },
    ],
  },
  {
    labelKey: 'nav.groupFleet',
    items: [
      { to: '/workers', labelKey: 'nav.workers', icon: Boxes },
      { to: '/cache', labelKey: 'nav.cache', icon: Database },
    ],
  },
]

const NAV_ITEMS = NAV_GROUPS.flatMap((group) => group.items)

function BrandMark({ className = 'h-6 w-6' }: { className?: string }) {
  return (
    <span
      className={cn(
        'flex shrink-0 items-center justify-center rounded-(--radius-md) bg-(--color-signal) text-(--color-on-signal)',
        className,
      )}
      aria-hidden
    >
      <Radio className="h-3.5 w-3.5" strokeWidth={2.25} />
    </span>
  )
}

export function Sidebar() {
  const t = useT()

  return (
    <aside className="hidden h-full w-[240px] shrink-0 flex-col border-r border-(--color-hairline) bg-(--color-sidebar) md:flex">
      <div className="flex h-12 items-center gap-2.5 px-4">
        <BrandMark />
        <span className="text-[14px] font-semibold tracking-[-0.01em] text-(--color-ink)">Hybrid-Grid</span>
      </div>

      <nav className="flex flex-col gap-5 px-3 pt-3">
        {NAV_GROUPS.map((group, index) => (
          <div key={group.labelKey ?? index} className="flex flex-col gap-0.5">
            {group.labelKey && (
              <div className="px-3 pb-1.5 font-mono text-[10px] font-medium uppercase tracking-widest text-(--color-ink-muted)">
                {t(group.labelKey)}
              </div>
            )}
            {group.items.map(({ to, labelKey, icon: Icon }) => (
              <Link
                key={to}
                to={to}
                activeOptions={{ exact: to === '/' }}
                className="flex items-center gap-2.5 rounded-(--radius-md) px-3 py-1.5 text-[14px] font-medium text-(--color-ink-secondary) transition-colors hover:bg-(--color-sidebar-accent)/60 hover:text-(--color-ink)"
                activeProps={{ className: 'bg-(--color-sidebar-accent) text-(--color-ink)' }}
              >
                <Icon className="h-4 w-4 shrink-0" strokeWidth={1.75} />
                {t(labelKey)}
              </Link>
            ))}
          </div>
        ))}
      </nav>

      <div className="mt-auto border-t border-(--color-hairline) px-4 py-3 text-[12px] text-(--color-ink-muted)">
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
      <BrandMark className="mr-2 hidden h-6 w-6 sm:flex" />
      {NAV_ITEMS.map(({ to, labelKey, icon: Icon }) => (
        <Link
          key={to}
          to={to}
          activeOptions={{ exact: to === '/' }}
          aria-label={t(labelKey)}
          title={t(labelKey)}
          className="flex h-8 w-8 items-center justify-center rounded-(--radius-md) text-(--color-ink-muted) transition-colors hover:bg-(--color-canvas-raised) hover:text-(--color-ink)"
          activeProps={{ className: 'bg-(--color-canvas-raised) text-(--color-ink)' }}
        >
          <Icon className="h-4 w-4" strokeWidth={1.75} />
        </Link>
      ))}
    </nav>
  )
}
