import { Link } from '@tanstack/react-router'
import { CodeBlock } from '@/components/CodeBlock'
import { StatusDot } from '@/components/StatusBadge'
import { useT, useRelativeTime } from '@/lib/i18n'
import { buildDurationMs, buildProgress, buildTone } from '@/lib/metrics'
import { STATUS_COLOR } from '@/lib/status'
import type { BuildInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'

const GROUP_BUILD_COMMAND = 'HG_BUILD_ID=$(uuidgen) hgbuild make -j8'

interface BuildsTableProps {
  builds: BuildInfo[]
  loading?: boolean
  /** Overview variant: drops the type and start-time columns. */
  compact?: boolean
}

/** Build history table shared by the builds page and the overview's "latest builds" panel. */
export function BuildsTable({ builds, loading = false, compact = false }: BuildsTableProps) {
  const t = useT()
  const columnCount = compact ? 6 : 8

  return (
    <div className="overflow-x-auto">
      <table className={cn('w-full border-collapse text-left text-[12px]', compact ? 'min-w-[600px]' : 'min-w-[880px]')}>
        <thead>
          <tr className="border-b border-(--color-hairline) text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">
            <th className="w-8 py-2 pl-4" aria-label={t('builds.colStatus')} />
            <th className="px-3 py-2 font-medium">{t('builds.colBuild')}</th>
            {!compact && <th className="px-3 py-2 font-medium">{t('builds.colType')}</th>}
            <th className="px-3 py-2 font-medium">{t('builds.colProgress')}</th>
            <th className="px-3 py-2 text-right font-medium">{t('builds.colCache')}</th>
            <th className="px-3 py-2 text-right font-medium">{t('builds.colDuration')}</th>
            {!compact && <th className="px-3 py-2 font-medium">{t('builds.colStarted')}</th>}
            <th className="py-2 pl-3 pr-4 font-medium">{t('builds.colLast')}</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-(--color-hairline)">
          {loading && (
            <tr>
              <td colSpan={columnCount} className="px-4 py-8 text-center text-(--color-ink-muted)">
                {t('builds.loading')}
              </td>
            </tr>
          )}
          {!loading && builds.length === 0 && (
            <tr>
              <td colSpan={columnCount} className="px-4 py-8">
                <EmptyBuilds />
              </td>
            </tr>
          )}
          {builds.map((build) => (
            <BuildRow key={build.id} build={build} compact={compact} />
          ))}
        </tbody>
      </table>
    </div>
  )
}

function BuildRow({ build, compact }: { build: BuildInfo; compact: boolean }) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const tone = buildTone(build)
  const duration = buildDurationMs(build)
  const progress = buildProgress(build)

  return (
    <tr className="transition-colors hover:bg-(--color-canvas-raised)">
      <td className="py-2.5 pl-4">
        <StatusDot tone={tone} />
      </td>
      <td className="max-w-[260px] px-3 py-2.5">
        <Link
          to="/builds/$buildId"
          params={{ buildId: build.id }}
          className="block truncate font-mono text-(--color-ink) hover:text-(--color-signal)"
          title={build.id}
        >
          {build.id}
        </Link>
      </td>
      {!compact && <td className="px-3 py-2.5 text-(--color-ink-muted)">{build.build_type || 'native'}</td>}
      <td className="px-3 py-2.5">
        <div className="flex items-center gap-2.5">
          <span className="w-16 font-tabular text-(--color-ink-secondary)">
            {build.completed_tasks + build.failed_tasks}/{build.total_tasks}
          </span>
          <span className="h-1 w-24 overflow-hidden rounded-full bg-(--color-surface-inset)">
            <span className="block h-full rounded-full" style={{ width: `${progress * 100}%`, backgroundColor: STATUS_COLOR[tone] }} />
          </span>
          {build.failed_tasks > 0 && (
            <span className="font-tabular text-(--color-status-failed)">
              {t('builds.failedSuffix', { n: build.failed_tasks })}
            </span>
          )}
        </div>
      </td>
      <td className="px-3 py-2.5 text-right font-tabular text-(--color-ink-secondary)">{build.from_cache_count}</td>
      <td className="px-3 py-2.5 text-right font-tabular text-(--color-ink-secondary)">
        {duration !== null ? formatDuration(duration) : '—'}
      </td>
      {!compact && (
        <td className="px-3 py-2.5 font-mono text-[11px] text-(--color-ink-muted)">{formatClock(build.first_task_at_ms)}</td>
      )}
      <td className="py-2.5 pl-3 pr-4 text-(--color-ink-muted)" title={formatClock(build.last_task_at_ms)}>
        {relativeTime(Math.round(build.last_task_at_ms / 1000))}
      </td>
    </tr>
  )
}

function EmptyBuilds() {
  const t = useT()
  return (
    <div className="mx-auto flex max-w-lg flex-col gap-3 text-center">
      <p className="text-[12px] text-(--color-ink-muted)">{t('builds.empty')}</p>
      <p className="text-[12px] text-(--color-ink-secondary)">{t('builds.emptyHint')}</p>
      <CodeBlock code={GROUP_BUILD_COMMAND} className="text-left" />
    </div>
  )
}
