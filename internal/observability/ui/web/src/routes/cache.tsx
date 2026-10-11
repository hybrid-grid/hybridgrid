import { createFileRoute, Link, useNavigate } from '@tanstack/react-router'
import { useMemo } from 'react'
import { HitRateTrend } from '@/components/HitRateTrend'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { useT } from '@/lib/i18n'
import { buildCacheShares, estimateCacheSavings, taskWindowStats, type BuildCacheShare } from '@/lib/metrics'
import { useBuilds, useStats, useTaskHistory } from '@/lib/queries'
import type { Stats, TaskInfo } from '@/lib/types'
import { cn, formatClock, formatDuration, formatPercent } from '@/lib/utils'
import { useRealtimeStore } from '@/store/useRealtimeStore'

export const Route = createFileRoute('/cache')({
  component: CachePage,
})

const MAX_SHARE_BARS = 40
const LOWEST_COUNT = 5

function CachePage() {
  const t = useT()
  const { data: stats } = useStats()
  const { data: buildsData } = useBuilds()
  const { data: tasksData } = useTaskHistory()
  const history = useRealtimeStore((state) => state.statsHistory)
  const tasks = useMemo(() => tasksData?.tasks ?? [], [tasksData])
  const shares = useMemo(() => buildCacheShares(buildsData?.builds ?? []), [buildsData])

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('cache.eyebrow')} title={t('cache.title')} />

      <StatStrip cells={summaryCells(t, stats, tasks)} />

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel className="xl:col-span-2" eyebrow={t('cache.trendNote')} title={t('cache.trendTitle')}>
          <HitRateTrend samples={history} tall />
        </Panel>
        <Panel eyebrow={t('cache.eyebrow')} title={t('cache.byTypeTitle')}>
          {stats ? <ByType stats={stats} /> : <div className="text-(--color-ink-muted)">—</div>}
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel className="xl:col-span-2" eyebrow={t('cache.perBuildNote')} title={t('cache.perBuildTitle')} padded={false}>
          <ShareChart shares={shares.slice(-MAX_SHARE_BARS)} />
        </Panel>
        <Panel eyebrow={t('cache.lowestNote')} title={t('cache.lowestTitle')} padded={false}>
          <LowestShares shares={shares} />
        </Panel>
      </div>
    </div>
  )
}

function summaryCells(t: ReturnType<typeof useT>, stats: Stats | undefined, tasks: TaskInfo[]): StatCell[] {
  const windowStats = taskWindowStats(tasks)
  const savings = estimateCacheSavings(tasks)
  return [
    {
      label: t('cache.summaryHitRate'),
      value: stats ? formatPercent(stats.cache_hit_rate, 1) : '—',
      color: 'var(--color-signal)',
      meter: stats?.cache_hit_rate,
      hint: stats ? t('overview.hintHitsMisses', { hits: stats.cache_hits, misses: stats.cache_misses }) : undefined,
      lead: true,
    },
    {
      label: t('cache.summaryWindowHits'),
      value: `${windowStats.cacheHits}/${windowStats.total}`,
      hint: windowStats.total > 0 ? formatPercent(windowStats.cacheHits / windowStats.total) : undefined,
    },
    {
      label: t('cache.summarySaved'),
      value: savings ? formatDuration(savings.savedMs) : '—',
      hint: savings ? t('cache.summarySavedHint', { n: savings.cachedCount, avg: formatDuration(savings.avgCompileMs) }) : undefined,
    },
  ]
}

function ByType({ stats }: { stats: Stats }) {
  const t = useT()
  const rows = [
    { label: t('cache.typeCpp'), hits: stats.cache_hits, misses: stats.cache_misses, rate: stats.cache_hit_rate },
    { label: t('cache.typeFlutter'), hits: stats.flutter_cache_hits, misses: stats.flutter_cache_misses, rate: stats.flutter_cache_hit_rate },
    { label: t('cache.typeUnity'), hits: stats.unity_cache_hits, misses: stats.unity_cache_misses, rate: stats.unity_cache_hit_rate },
  ]
  return (
    <div className="flex flex-col gap-4">
      {rows.map((row) => {
        const total = row.hits + row.misses
        return (
          <div key={row.label} className="flex flex-col gap-1.5">
            <div className="flex items-baseline justify-between gap-2">
              <span className="text-[12px] text-(--color-ink-secondary)">{row.label}</span>
              <span className={cn('font-tabular text-[15px] font-semibold', total > 0 ? 'text-(--color-ink)' : 'text-(--color-ink-faint)')}>
                {total > 0 ? formatPercent(row.rate) : '—'}
              </span>
            </div>
            <span className="h-1 overflow-hidden rounded-full bg-(--color-surface-inset)">
              <span className="block h-full rounded-full bg-(--color-signal)" style={{ width: `${Math.min(100, row.rate * 100)}%` }} />
            </span>
            <span className="font-tabular text-[10px] text-(--color-ink-faint)">
              {t('overview.hintHitsMisses', { hits: row.hits, misses: row.misses })}
            </span>
          </div>
        )
      })}
    </div>
  )
}

function ShareChart({ shares }: { shares: BuildCacheShare[] }) {
  const t = useT()
  const navigate = useNavigate()

  if (shares.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('cache.perBuildEmpty')}</div>
  }

  return (
    <div className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-2 p-4">
      <div className="flex h-40 flex-col justify-between text-right font-tabular text-[10px] text-(--color-ink-faint)">
        <span>100%</span>
        <span>50%</span>
        <span>0</span>
      </div>
      <div className="relative h-40">
        <div className="pointer-events-none absolute inset-0 flex flex-col justify-between">
          <span className="h-px bg-(--color-hairline)" />
          <span className="h-px bg-(--color-hairline)" />
          <span className="h-px bg-(--color-hairline-strong)" />
        </div>
        <div className="absolute inset-0 flex items-end gap-1">
          {shares.map(({ build, share }, i) => (
            <button
              key={build.id}
              type="button"
              onClick={() => navigate({ to: '/builds/$buildId', params: { buildId: build.id } })}
              aria-label={`${build.id} · ${formatPercent(share)}`}
              className="group relative flex h-full min-w-[4px] max-w-10 flex-1 cursor-pointer flex-col justify-end"
            >
              <span
                className="w-full rounded-t-[3px] bg-(--color-signal) transition-[filter] group-hover:brightness-125"
                style={{ height: `${Math.max(1.5, share * 100)}%`, opacity: share > 0 ? 1 : 0.35 }}
              />
              <span
                role="tooltip"
                className={cn(
                  'pointer-events-none invisible absolute bottom-full z-20 mb-2 w-48 rounded-(--radius-md) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) p-2.5 text-left opacity-0 shadow-[0_8px_24px_-8px_oklch(0_0_0/50%)] transition-opacity duration-150 group-hover:visible group-hover:opacity-100',
                  i > shares.length / 2 ? 'right-0' : 'left-0',
                )}
              >
                <span className="block truncate font-mono text-[11px] text-(--color-ink)">{build.id}</span>
                <span className="block font-mono text-[10px] text-(--color-ink-faint)">{formatClock(build.first_task_at_ms)}</span>
                <span className="mt-1.5 block font-tabular text-[15px] font-semibold text-(--color-signal)">{formatPercent(share)}</span>
                <span className="block font-tabular text-[11px] text-(--color-ink-muted)">
                  {build.from_cache_count}/{build.total_tasks} {t('common.cache')}
                </span>
              </span>
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}

function LowestShares({ shares }: { shares: BuildCacheShare[] }) {
  const t = useT()
  const lowest = shares.toSorted((a, b) => a.share - b.share || b.build.total_tasks - a.build.total_tasks).slice(0, LOWEST_COUNT)

  if (lowest.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('cache.perBuildEmpty')}</div>
  }

  return (
    <ul className="divide-y divide-(--color-hairline)">
      {lowest.map(({ build, share }) => (
        <li key={build.id} className="flex items-center justify-between gap-3 px-4 py-2.5">
          <div className="min-w-0">
            <Link
              to="/builds/$buildId"
              params={{ buildId: build.id }}
              className="block truncate font-mono text-[12px] text-(--color-ink) hover:text-(--color-signal)"
            >
              {build.id}
            </Link>
            <span className="font-tabular text-[10px] text-(--color-ink-faint)">
              {build.from_cache_count}/{build.total_tasks} {t('common.cache')}
            </span>
          </div>
          <span className="font-tabular text-[13px] font-semibold text-(--color-ink-secondary)">{formatPercent(share)}</span>
        </li>
      ))}
    </ul>
  )
}
