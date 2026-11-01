import { createFileRoute, Link } from '@tanstack/react-router'
import { Cpu, MemoryStick } from 'lucide-react'
import { useMemo } from 'react'
import { ExecutorSlots } from '@/components/ExecutorSlots'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { CircuitBadge } from '@/components/StatusBadge'
import { TaskTable } from '@/components/TaskTable'
import { WorkerHeartbeat } from '@/components/WorkerHeartbeat'
import { WorkerThroughputChart } from '@/components/WorkerThroughputChart'
import { Button } from '@/components/ui/button'
import { useNow } from '@/hooks/useNow'
import { groupFailures } from '@/lib/failures'
import { useT, useRelativeTime } from '@/lib/i18n'
import { taskWindowStats } from '@/lib/metrics'
import { useTaskHistory, useWorkers } from '@/lib/queries'
import { normalizeStatus } from '@/lib/status'
import type { TaskInfo, WorkerInfo } from '@/lib/types'
import { formatBytes, formatDuration, formatPercent } from '@/lib/utils'

export const Route = createFileRoute('/workers/$workerId')({
  component: WorkerDetail,
})

const RECENT_TASKS = 25
const TOP_FAILURES = 5

function WorkerDetail() {
  const { workerId } = Route.useParams()
  const t = useT()
  const now = useNow(1000)
  const { data: workersData, isLoading } = useWorkers()
  const { data: tasksData } = useTaskHistory()
  const worker = workersData?.workers.find((w) => w.id === workerId)
  const tasks = useMemo(() => (tasksData?.tasks ?? []).filter((task) => task.worker_id === workerId), [tasksData, workerId])
  const running = useMemo(() => tasks.filter((task) => normalizeStatus(task.status) === 'running'), [tasks])
  const failures = useMemo(() => groupFailures(tasks).slice(0, TOP_FAILURES), [tasks])

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow={
          <Link to="/workers" className="hover:text-(--color-ink-secondary)">
            ← {t('workerDetail.backLabel')}
          </Link>
        }
        title={<span className="font-mono">{worker?.host || workerId}</span>}
        actions={worker && <CircuitBadge state={worker.circuit_state} />}
      />

      {!isLoading && !worker && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('workerDetail.notFound')}</div>
        </Panel>
      )}

      <StatStrip cells={summaryCells(t, tasks)} />

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel eyebrow={t('workerDetail.specsTitle')} title={<span className="font-mono text-[12px]">{workerId}</span>} padded={false}>
          {worker ? <MachinePanel worker={worker} /> : <div className="p-4 text-(--color-ink-muted)">—</div>}
        </Panel>
        <Panel eyebrow={t('overview.workerFleetTitle')} className="xl:col-span-2">
          {worker && (
            <WorkerHeartbeat workerId={worker.id} circuitState={worker.circuit_state} loadRatio={loadRatio(worker)} className="mb-3" />
          )}
          <ExecutorSlots active={worker?.active_tasks ?? running.length} max={worker?.max_parallel_tasks ?? 0} running={running} now={now} />
        </Panel>
      </div>

      <div className="grid grid-cols-1 gap-4 xl:grid-cols-3">
        <Panel className="xl:col-span-2" eyebrow={t('workerDetail.throughputTitle')} padded={false}>
          <WorkerThroughputChart tasks={tasks} workers={worker ? [worker] : []} />
        </Panel>
        <Panel eyebrow={t('workerDetail.failuresTitle')} padded={false}>
          <FailureList failures={failures} />
        </Panel>
      </div>

      <Panel
        eyebrow={t('workerDetail.recentTitle')}
        action={
          <Button variant="ghost" size="sm" asChild>
            <Link to="/tasks" search={{ worker: workerId }}>
              {t('nav.tasks')} →
            </Link>
          </Button>
        }
        padded={false}
        className="overflow-hidden"
      >
        <div className="overflow-x-auto">
          <TaskTable tasks={tasks.slice(0, RECENT_TASKS)} showBuild hideWorker />
        </div>
      </Panel>
    </div>
  )
}

function loadRatio(worker: WorkerInfo): number {
  return worker.max_parallel_tasks > 0 ? worker.active_tasks / worker.max_parallel_tasks : 0
}

function summaryCells(t: ReturnType<typeof useT>, tasks: TaskInfo[]): StatCell[] {
  const stats = taskWindowStats(tasks)
  return [
    { label: t('workerDetail.summaryTasks'), value: stats.total },
    {
      label: t('workerDetail.summarySuccess'),
      value: formatPercent(stats.successRate, 1),
      color: stats.failed > 0 ? 'var(--color-status-timeout)' : 'var(--color-signal)',
      hint: stats.failed > 0 ? `${stats.failed} ${t('status.failed').toLowerCase()}` : undefined,
    },
    { label: t('workerDetail.summaryAvgCompile'), value: stats.avgCompileMs !== undefined ? formatDuration(stats.avgCompileMs) : '—' },
    { label: t('workerDetail.summaryAvgQueue'), value: stats.avgQueueMs !== undefined ? formatDuration(stats.avgQueueMs) : '—' },
    { label: t('workerDetail.summaryCache'), value: stats.cacheHits },
  ]
}

function MachinePanel({ worker }: { worker: WorkerInfo }) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const rows: [string, string][] = [
    ['OS', `${worker.os} · ${worker.architecture}`],
    [t('workers.addressLabel'), worker.address || '—'],
    [t('workers.discoveryLabel'), worker.discovery_source || '—'],
    [t('workers.versionLabel'), worker.version || '—'],
    [t('workers.latencyLabel'), `${Math.round(worker.avg_latency_ms)}ms`],
    [t('workers.tasksLabel'), String(worker.total_tasks)],
  ]
  const capabilities = [...(worker.compilers ?? []), ...(worker.docker_available ? ['docker'] : []), ...(worker.flutter_available ? ['flutter'] : []), ...(worker.unity_available ? ['unity'] : [])]

  return (
    <div className="flex flex-col">
      <div className="flex items-center gap-4 px-4 py-3 text-[12px] text-(--color-ink-secondary)">
        <span className="flex items-center gap-1.5">
          <Cpu className="h-3.5 w-3.5 text-(--color-ink-faint)" /> {t('workers.cores', { n: worker.cpu_cores })}
        </span>
        <span className="flex items-center gap-1.5">
          <MemoryStick className="h-3.5 w-3.5 text-(--color-ink-faint)" /> {formatBytes(worker.memory_gb)}
        </span>
        <span className="ml-auto text-[11px] text-(--color-ink-faint)">{t('workers.seenLabel', { t: relativeTime(worker.last_seen) })}</span>
      </div>
      <dl className="divide-y divide-(--color-hairline) border-t border-(--color-hairline)">
        {rows.map(([label, value]) => (
          <div key={label} className="flex items-center justify-between gap-3 px-4 py-2 text-[12px]">
            <dt className="text-(--color-ink-faint)">{label}</dt>
            <dd className="truncate font-mono text-[11px] text-(--color-ink-secondary)">{value}</dd>
          </div>
        ))}
      </dl>
      {capabilities.length > 0 && (
        <div className="flex flex-wrap gap-1 border-t border-(--color-hairline) px-4 py-2.5">
          {capabilities.map((cap) => (
            <span key={cap} className="rounded-(--radius-xs) bg-(--color-surface-inset) px-1.5 py-0.5 font-mono text-[10px] text-(--color-ink-muted)">
              {cap}
            </span>
          ))}
        </div>
      )}
    </div>
  )
}

function FailureList({ failures }: { failures: ReturnType<typeof groupFailures> }) {
  const t = useT()
  if (failures.length === 0) {
    return <div className="px-4 py-8 text-center text-[12px] text-(--color-ink-muted)">{t('workerDetail.noFailures')}</div>
  }
  return (
    <ul className="divide-y divide-(--color-hairline)">
      {failures.map((group) => (
        <li key={group.signature} className="flex items-start gap-3 px-4 py-2.5">
          <span className="font-tabular text-[15px] font-semibold text-(--color-status-failed)">{group.count}</span>
          <div className="min-w-0">
            <p className="line-clamp-2 font-mono text-[11px] text-(--color-ink-secondary)">{group.signature}</p>
            <Link to="/failures" className="text-[10px] text-(--color-ink-faint) hover:text-(--color-signal)">
              {t('failures.occurrences', { n: group.count })} · {t('failures.builds', { n: group.buildIds.length })}
            </Link>
          </div>
        </li>
      ))}
    </ul>
  )
}
