import { createFileRoute, Link } from '@tanstack/react-router'
import { Cpu, MemoryStick } from 'lucide-react'
import { useMemo } from 'react'
import { AddWorkerDialog } from '@/components/AddWorkerDialog'
import { ExecutorSlots } from '@/components/ExecutorSlots'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { CircuitBadge } from '@/components/StatusBadge'
import { WorkerHeartbeat } from '@/components/WorkerHeartbeat'
import { useNow } from '@/hooks/useNow'
import { useT, useRelativeTime } from '@/lib/i18n'
import { fleetCapacity } from '@/lib/metrics'
import { useTasks, useWorkers } from '@/lib/queries'
import { normalizeStatus } from '@/lib/status'
import type { TaskInfo, WorkerInfo } from '@/lib/types'
import { cn, formatBytes, formatPercent } from '@/lib/utils'

export const Route = createFileRoute('/workers/')({
  component: Workers,
})

const TASK_WINDOW = 200

function Workers() {
  const t = useT()
  const now = useNow(1000)
  const { data, isLoading } = useWorkers()
  const { data: tasksData } = useTasks(TASK_WINDOW)
  const workers = useMemo(() => data?.workers ?? [], [data])

  const runningByWorker = useMemo(() => {
    const map = new Map<string, TaskInfo[]>()
    for (const task of tasksData?.tasks ?? []) {
      if (!task.worker_id || normalizeStatus(task.status) !== 'running') continue
      map.set(task.worker_id, [...(map.get(task.worker_id) ?? []), task])
    }
    return map
  }, [tasksData])

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('workers.eyebrow')} title={t('workers.title')} actions={<AddWorkerDialog />} />

      {workers.length > 0 && <StatStrip cells={fleetCells(t, workers)} />}

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('workers.loading')}</div>}

      {!isLoading && workers.length === 0 && (
        <Panel>
          <div className="py-6 text-center text-[12px] text-(--color-ink-muted)">{t('workers.empty')}</div>
        </Panel>
      )}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 2xl:grid-cols-3">
        {workers.map((worker) => (
          <WorkerCard key={worker.id} worker={worker} running={runningByWorker.get(worker.id) ?? []} now={now} />
        ))}
      </div>
    </div>
  )
}

function fleetCells(t: ReturnType<typeof useT>, workers: WorkerInfo[]): StatCell[] {
  const healthy = workers.filter((worker) => worker.healthy).length
  const capacity = fleetCapacity(workers)
  const cores = workers.reduce((sum, worker) => sum + (worker.cpu_cores || 0), 0)
  const memory = workers.reduce((sum, worker) => sum + (worker.memory_gb || 0), 0)

  return [
    { label: t('workers.summaryWorkers'), value: workers.length },
    {
      label: t('workers.summaryHealthy'),
      value: `${healthy}/${workers.length}`,
      color: healthy < workers.length ? 'var(--color-status-failed)' : 'var(--color-signal)',
    },
    {
      label: t('workers.summaryCapacity'),
      value: `${capacity.active}/${capacity.max}`,
      meter: capacity.ratio,
      hint: formatPercent(capacity.ratio),
    },
    { label: t('workers.summaryCores'), value: cores, hint: formatBytes(memory) },
  ]
}

interface WorkerCardProps {
  worker: WorkerInfo
  running: TaskInfo[]
  now: number
}

function WorkerCard({ worker, running, now }: WorkerCardProps) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const loadRatio = worker.max_parallel_tasks > 0 ? worker.active_tasks / worker.max_parallel_tasks : 0
  const capabilities = workerCapabilities(worker)

  return (
    <Panel padded={false} className={cn('flex flex-col', !worker.healthy && 'border-[color-mix(in_oklch,var(--color-status-failed)_35%,transparent)]')}>
      <div className="flex items-start justify-between gap-3 px-4 pb-2 pt-3">
        <div className="min-w-0">
          <div className="flex items-center gap-2">
            <span
              className="h-2 w-2 shrink-0 rounded-full"
              style={{ backgroundColor: worker.healthy ? 'var(--color-signal)' : 'var(--color-status-failed)' }}
              title={worker.healthy ? t('workers.healthy') : t('workers.unhealthy')}
            />
            <Link
              to="/workers/$workerId"
              params={{ workerId: worker.id }}
              className="truncate font-mono text-[13px] font-semibold text-(--color-ink) hover:text-(--color-signal)"
            >
              {worker.host || worker.id}
            </Link>
          </div>
          {worker.host && worker.host !== worker.id && (
            <div className="truncate pl-4 font-mono text-[10px] text-(--color-ink-faint)">{worker.id}</div>
          )}
          <div className="pl-4 text-[11px] text-(--color-ink-muted)">
            {worker.os} · {worker.architecture}
          </div>
        </div>
        <div className="flex shrink-0 flex-col items-end gap-1">
          <CircuitBadge state={worker.circuit_state} className="px-0" />
          {!worker.healthy && (
            <span className="text-[10px] uppercase tracking-[0.06em] text-(--color-status-failed)">{t('workers.unhealthy')}</span>
          )}
        </div>
      </div>

      <WorkerHeartbeat workerId={worker.id} circuitState={worker.circuit_state} loadRatio={loadRatio} className="px-4" />

      <div className="border-t border-(--color-hairline) px-4 py-3">
        <ExecutorSlots active={worker.active_tasks} max={worker.max_parallel_tasks} running={running} now={now} />
      </div>

      <dl className="grid grid-cols-3 gap-3 border-t border-(--color-hairline) px-4 py-3">
        <Figure label={t('workers.successLabel')} value={worker.total_tasks > 0 ? formatPercent(worker.success_rate) : '—'} />
        <Figure label={t('workers.latencyLabel')} value={`${Math.round(worker.avg_latency_ms)}ms`} />
        <Figure label={t('workers.tasksLabel')} value={worker.total_tasks} />
      </dl>

      <div className="flex flex-wrap items-center gap-x-4 gap-y-1 border-t border-(--color-hairline) px-4 py-2.5 text-[11px] text-(--color-ink-muted)">
        <span className="flex items-center gap-1.5">
          <Cpu className="h-3 w-3" /> {t('workers.cores', { n: worker.cpu_cores })}
        </span>
        <span className="flex items-center gap-1.5">
          <MemoryStick className="h-3 w-3" /> {formatBytes(worker.memory_gb)}
        </span>
        {worker.version && <span className="font-mono">v{worker.version.replace(/^v/, '')}</span>}
        <span className="ml-auto">{t('workers.seenLabel', { t: relativeTime(worker.last_seen) })}</span>
      </div>

      {capabilities.length > 0 && (
        <div className="mt-auto flex flex-wrap gap-1 border-t border-(--color-hairline) px-4 py-2.5">
          {capabilities.map((cap) => (
            <span
              key={cap}
              className="rounded-(--radius-xs) bg-(--color-surface-inset) px-1.5 py-0.5 font-mono text-[10px] text-(--color-ink-muted)"
            >
              {cap}
            </span>
          ))}
        </div>
      )}

      <div className="flex items-center justify-between gap-3 border-t border-(--color-hairline) px-4 py-2 font-mono text-[10px] text-(--color-ink-faint)">
        <span className="truncate" title={t('workers.addressLabel')}>
          {worker.address || '—'}
        </span>
        {worker.discovery_source && (
          <span className="shrink-0" title={t('workers.discoveryLabel')}>
            {worker.discovery_source}
          </span>
        )}
      </div>
    </Panel>
  )
}

function workerCapabilities(worker: WorkerInfo): string[] {
  return [
    ...(worker.compilers ?? []),
    ...(worker.docker_available ? ['docker'] : []),
    ...(worker.flutter_available ? [`flutter${worker.flutter_sdk_version ? ` ${worker.flutter_sdk_version}` : ''}`] : []),
    ...(worker.unity_available ? (worker.unity_versions ?? ['unity']).map((v) => (v === 'unity' ? v : `unity ${v}`)) : []),
  ]
}

function Figure({ label, value }: { label: string; value: string | number }) {
  return (
    <div className="min-w-0">
      <dt className="text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{label}</dt>
      <dd className="mt-0.5 truncate font-tabular text-[15px] font-semibold text-(--color-ink)">{value}</dd>
    </div>
  )
}
