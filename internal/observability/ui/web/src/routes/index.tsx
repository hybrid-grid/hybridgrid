import { createFileRoute } from '@tanstack/react-router'
import {
  Activity,
  AlertTriangle,
  Boxes,
  CheckCircle2,
  Clock,
  Database,
  Gauge,
  Hourglass,
  Layers,
  ListTree,
  Timer,
  XCircle,
} from 'lucide-react'
import { useMemo } from 'react'
import { MetricTile } from '@/components/MetricTile'
import { Panel } from '@/components/Panel'
import { PageHeader } from '@/components/PageHeader'
import { WorkerHeartbeat } from '@/components/WorkerHeartbeat'
import { CircuitBadge } from '@/components/StatusBadge'
import { CacheStatsPanel } from '@/components/CacheStatsPanel'
import { ClusterActivity } from '@/components/ClusterActivity'
import { Button } from '@/components/ui/button'
import { useT, useRelativeTime } from '@/lib/i18n'
import { useBuilds, useStats, useTasks, useWorkers } from '@/lib/queries'
import { formatDuration, formatUptime } from '@/lib/utils'
import { useRealtimeStore } from '@/store/useRealtimeStore'

export const Route = createFileRoute('/')({
  component: Overview,
})

function Overview() {
  const t = useT()
  const relativeTime = useRelativeTime()
  const { data: stats } = useStats()
  const { data: workersData } = useWorkers()
  const { data: buildsData } = useBuilds()
  const { data: tasksData } = useTasks(150)
  const events = useRealtimeStore((state) => state.events)
  const clearEvents = useRealtimeStore((state) => state.clearEvents)
  const workers = workersData?.workers ?? []
  const builds = buildsData?.builds ?? []
  const tasks = tasksData?.tasks ?? []

  const avgBuildDuration = useMemo(() => {
    if (builds.length === 0) return undefined
    const total = builds.reduce((sum, build) => sum + Math.max(0, build.last_task_at_ms - build.first_task_at_ms), 0)
    return total / builds.length
  }, [builds])

  return (
    <div className="flex flex-col gap-5">
      <PageHeader eyebrow={t('overview.eyebrow')} title={t('overview.title')} />

      <div className="grid grid-cols-2 gap-3 sm:grid-cols-4 lg:grid-cols-8">
        <Panel className="col-span-1">
          <MetricTile
            icon={Activity}
            label={t('overview.metricActive')}
            value={stats?.active_tasks ?? '—'}
            accent="running"
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile icon={Hourglass} label={t('overview.metricQueued')} value={stats?.queued_tasks ?? '—'} />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={CheckCircle2}
            label={t('overview.metricSucceeded')}
            value={stats?.success_tasks ?? '—'}
            accent="signal"
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={XCircle}
            label={t('overview.metricFailed')}
            value={stats?.failed_tasks ?? '—'}
            accent="failed"
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={Database}
            label={t('overview.metricCacheHitRate')}
            value={stats ? `${Math.round(stats.cache_hit_rate * 100)}%` : '—'}
            hint={stats ? t('overview.hintHitsMisses', { hits: stats.cache_hits, misses: stats.cache_misses }) : undefined}
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={Boxes}
            label={t('overview.metricHealthyWorkers')}
            value={stats ? `${stats.healthy_workers}/${stats.total_workers}` : '—'}
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={Timer}
            label={t('overview.metricAvgBuildDuration')}
            value={avgBuildDuration !== undefined ? formatDuration(avgBuildDuration) : '—'}
            hint={t('overview.hintBuildsRetained', { n: builds.length })}
          />
        </Panel>
        <Panel className="col-span-1">
          <MetricTile
            icon={Clock}
            label={t('overview.metricUptime')}
            value={stats ? formatUptime(stats.uptime_seconds) : '—'}
            hint={stats ? t('overview.hintQueueActive', { queued: stats.queued_tasks, active: stats.active_tasks }) : undefined}
          />
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel
          className="xl:col-span-2"
          eyebrow={t('overview.workerFleetEyebrow')}
          title={
            <span className="flex items-center gap-1.5">
              <Gauge className="h-3.5 w-3.5" /> {t('overview.workerFleetTitle')}
            </span>
          }
          padded={false}
        >
          {workers.length === 0 ? (
            <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('overview.noWorkers')}</div>
          ) : (
            <div className="divide-y divide-(--color-hairline)">
              {workers.map((worker) => {
                const loadRatio =
                  worker.max_parallel_tasks > 0 ? worker.active_tasks / worker.max_parallel_tasks : 0
                return (
                  <div key={worker.id} className="flex items-center gap-4 px-4 py-3">
                    <div className="w-36 shrink-0">
                      <div className="truncate font-mono text-[12px] font-medium text-(--color-ink)">
                        {worker.id}
                      </div>
                      <CircuitBadge state={worker.circuit_state} className="mt-0.5" />
                    </div>
                    <WorkerHeartbeat
                      workerId={worker.id}
                      circuitState={worker.circuit_state}
                      loadRatio={loadRatio}
                      className="flex-1"
                    />
                    <div className="w-20 shrink-0 text-right font-tabular text-[12px] text-(--color-ink-muted)">
                      {worker.active_tasks}/{worker.max_parallel_tasks}
                    </div>
                  </div>
                )
              })}
            </div>
          )}
        </Panel>

        <Panel
          eyebrow={t('overview.cacheStatsEyebrow')}
          title={t('overview.cacheStatsTitle')}
          padded={false}
        >
          {stats ? (
            <CacheStatsPanel stats={stats} />
          ) : (
            <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">—</div>
          )}
        </Panel>
      </div>

      <Panel
        eyebrow={t('overview.clusterActivityEyebrow')}
        title={t('overview.clusterActivityTitle')}
        padded={false}
      >
        <ClusterActivity tasks={tasks} />
      </Panel>

      <Panel
        eyebrow={t('overview.activityEyebrow')}
        title={
          <span className="flex items-center gap-1.5">
            <ListTree className="h-3.5 w-3.5" /> {t('overview.eventLogTitle')}
          </span>
        }
        action={
          events.length > 0 && (
            <Button variant="ghost" size="sm" onClick={clearEvents}>
              {t('common.clear')}
            </Button>
          )
        }
        padded={false}
      >
        <div className="max-h-[320px] overflow-auto">
          {events.length === 0 ? (
            <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('overview.waitingEvents')}</div>
          ) : (
            <ul className="divide-y divide-(--color-hairline)">
              {events.map((event) => (
                <li key={event.id} className="flex items-start gap-2 px-4 py-2.5 text-[12px]">
                  {event.message.type === 'task_completed' ? (
                    <AlertTriangle className="mt-0.5 h-3 w-3 shrink-0 text-(--color-status-running)" />
                  ) : (
                    <Layers className="mt-0.5 h-3 w-3 shrink-0 text-(--color-ink-faint)" />
                  )}
                  <div className="min-w-0 flex-1">
                    <div className="font-mono text-(--color-ink-secondary)">{event.message.type}</div>
                    <div className="text-(--color-ink-faint)">{relativeTime(event.message.timestamp)}</div>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </Panel>
    </div>
  )
}
