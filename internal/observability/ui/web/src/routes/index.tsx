import { createFileRoute, Link } from '@tanstack/react-router'
import { ArrowRight } from 'lucide-react'
import { BuildDurationsChart } from '@/components/BuildDurationsChart'
import { BuildsTable } from '@/components/BuildsTable'
import { CacheStatsPanel } from '@/components/CacheStatsPanel'
import { ClusterActivity } from '@/components/ClusterActivity'
import { EventLog } from '@/components/EventLog'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { CircuitBadge } from '@/components/StatusBadge'
import { WorkerThroughputChart } from '@/components/WorkerThroughputChart'
import { WorkerHeartbeat } from '@/components/WorkerHeartbeat'
import { Button } from '@/components/ui/button'
import { useT } from '@/lib/i18n'
import { averageFinishedBuildMs, fleetCapacity, passRate } from '@/lib/metrics'
import { useBuilds, useStats, useTasks, useWorkers } from '@/lib/queries'
import type { BuildInfo, Stats, WorkerInfo } from '@/lib/types'
import { cn, formatDuration, formatPercent, formatUptime } from '@/lib/utils'
import { useRealtimeStore } from '@/store/useRealtimeStore'

export const Route = createFileRoute('/')({
  component: Overview,
})

const ACTIVITY_WINDOW = 200
const RECENT_BUILDS = 6

function Overview() {
  const t = useT()
  const { data: stats } = useStats()
  const { data: workersData } = useWorkers()
  const { data: buildsData } = useBuilds()
  const { data: tasksData } = useTasks(ACTIVITY_WINDOW)
  const events = useRealtimeStore((state) => state.events)
  const clearEvents = useRealtimeStore((state) => state.clearEvents)
  const workers = workersData?.workers ?? []
  const builds = buildsData?.builds ?? []
  const tasks = tasksData?.tasks ?? []

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('overview.eyebrow')} title={t('overview.title')} />

      <StatStrip cells={overviewCells(t, stats, workers, builds)} />

      <Panel eyebrow={t('overview.clusterActivityEyebrow')} title={t('overview.clusterActivityTitle')} padded={false}>
        <ClusterActivity tasks={tasks} workers={workers} />
      </Panel>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel
          className="overflow-hidden xl:col-span-2"
          eyebrow={t('overview.recentBuildsEyebrow')}
          title={t('overview.recentBuildsTitle')}
          action={
            <Button variant="ghost" size="sm" asChild>
              <Link to="/builds">
                {t('overview.viewAllBuilds')}
                <ArrowRight className="h-3 w-3" />
              </Link>
            </Button>
          }
          padded={false}
        >
          <BuildsTable builds={builds.slice(0, RECENT_BUILDS)} loading={!buildsData} compact />
        </Panel>
        <Panel eyebrow={t('builds.durationsEyebrow')} title={t('builds.durationsTitle')} padded={false}>
          <BuildDurationsChart builds={builds} />
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel
          className="xl:col-span-2"
          eyebrow={t('overview.throughputEyebrow')}
          title={t('overview.throughputTitle')}
          padded={false}
        >
          <WorkerThroughputChart tasks={tasks} workers={workers} />
        </Panel>
        <Panel eyebrow={t('overview.cacheStatsEyebrow')} title={t('overview.cacheStatsTitle')} padded={false}>
          {stats ? <CacheStatsPanel stats={stats} /> : <div className="p-6 text-center text-(--color-ink-muted)">—</div>}
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel
          className="xl:col-span-2"
          eyebrow={t('overview.workerFleetEyebrow')}
          title={t('overview.workerFleetTitle')}
          action={
            <Button variant="ghost" size="sm" asChild>
              <Link to="/workers">
                {t('overview.viewAllWorkers')}
                <ArrowRight className="h-3 w-3" />
              </Link>
            </Button>
          }
          padded={false}
        >
          <FleetList workers={workers} />
        </Panel>

        <Panel
          eyebrow={t('overview.activityEyebrow')}
          title={t('overview.eventLogTitle')}
          action={
            events.length > 0 && (
              <Button variant="ghost" size="sm" onClick={clearEvents}>
                {t('common.clear')}
              </Button>
            )
          }
          padded={false}
        >
          <div className="max-h-[360px] overflow-auto">
            <EventLog events={events} />
          </div>
        </Panel>
      </div>
    </div>
  )
}

type Translate = ReturnType<typeof useT>

const RELIABILITY_GOOD = 0.98
const RELIABILITY_WARN = 0.9

function reliabilityColor(rate: number | undefined): string | undefined {
  if (rate === undefined) return undefined
  if (rate >= RELIABILITY_GOOD) return 'var(--color-signal)'
  if (rate >= RELIABILITY_WARN) return 'var(--color-status-timeout)'
  return 'var(--color-status-failed)'
}

function overviewCells(t: Translate, stats: Stats | undefined, workers: WorkerInfo[], builds: BuildInfo[]): StatCell[] {
  const reliability = stats ? passRate(stats) : undefined
  const capacity = fleetCapacity(workers)
  const avgBuild = averageFinishedBuildMs(builds)
  const settled = builds.filter((build) => build.running_tasks === 0).length

  return [
    {
      label: t('overview.metricReliability'),
      value: formatPercent(reliability, 1),
      color: reliabilityColor(reliability),
      hint: stats
        ? t('overview.hintPassed', { ok: stats.success_tasks, total: stats.success_tasks + stats.failed_tasks })
        : undefined,
      lead: true,
    },
    {
      label: t('overview.metricInFlight'),
      value: stats?.active_tasks ?? '—',
      color: stats?.active_tasks ? 'var(--color-status-running)' : undefined,
      hint: stats ? t('overview.hintQueued', { n: stats.queued_tasks }) : undefined,
    },
    {
      label: t('overview.metricExecutors'),
      value: stats ? `${stats.healthy_workers}/${stats.total_workers}` : '—',
      meter: capacity.ratio,
      hint: capacity.max > 0 ? t('overview.hintUtilization', { active: capacity.active, max: capacity.max }) : undefined,
    },
    {
      label: t('overview.metricCacheHitRate'),
      value: stats ? formatPercent(stats.cache_hit_rate, 1) : '—',
      meter: stats?.cache_hit_rate,
      hint: stats ? t('overview.hintHitsMisses', { hits: stats.cache_hits, misses: stats.cache_misses }) : undefined,
    },
    {
      label: t('overview.metricAvgBuildDuration'),
      value: avgBuild !== undefined ? formatDuration(avgBuild) : '—',
      hint: t('overview.hintSettledBuilds', { n: settled }),
    },
    {
      label: t('overview.metricUptime'),
      value: stats ? formatUptime(stats.uptime_seconds) : '—',
    },
  ]
}

function FleetList({ workers }: { workers: WorkerInfo[] }) {
  const t = useT()

  if (workers.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('overview.noWorkers')}</div>
  }

  return (
    <div className="divide-y divide-(--color-hairline)">
      {workers.map((worker) => {
        const loadRatio = worker.max_parallel_tasks > 0 ? worker.active_tasks / worker.max_parallel_tasks : 0
        return (
          <div
            key={worker.id}
            className={cn('grid grid-cols-[minmax(0,1fr)_auto] items-center gap-4 px-4 py-2.5 sm:grid-cols-[minmax(0,180px)_minmax(0,1fr)_auto]', !worker.healthy && 'opacity-60')}
          >
            <div className="min-w-0">
              <Link
                to="/workers/$workerId"
                params={{ workerId: worker.id }}
                className="block truncate font-mono text-[12px] font-medium text-(--color-ink) hover:text-(--color-signal)"
                title={worker.id}
              >
                {worker.host || worker.id}
              </Link>
              <div className="flex items-center gap-2">
                <CircuitBadge state={worker.circuit_state} className="px-0" />
                {!worker.healthy && (
                  <span className="text-[10px] uppercase tracking-[0.06em] text-(--color-status-failed)">
                    {t('overview.fleetUnhealthy')}
                  </span>
                )}
              </div>
            </div>
            <WorkerHeartbeat
              workerId={worker.id}
              circuitState={worker.circuit_state}
              loadRatio={loadRatio}
              className="hidden sm:block"
            />
            <div className="text-right">
              <div className="font-tabular text-[13px] font-medium text-(--color-ink-secondary)">
                {worker.active_tasks}/{worker.max_parallel_tasks}
              </div>
              <div className="font-tabular text-[10px] text-(--color-ink-faint)">
                {Math.round(worker.avg_latency_ms)}ms · {worker.total_tasks > 0 ? formatPercent(worker.success_rate) : '—'}
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
