import { createFileRoute, Link } from '@tanstack/react-router'
import { useMemo } from 'react'
import { AddWorkerDialog } from '@/components/AddWorkerDialog'
import { MachineIllustration } from '@/components/MachineIllustration'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { WorkerFlow } from '@/components/WorkerFlow'
import { UnityCapabilityBadge } from '@/components/UnityCapabilityBadge'
import { WorkerUsage } from '@/components/WorkerUsage'
import { useNow } from '@/hooks/useNow'
import { mergeLiveTasks, uniqueWorkerNames } from '@/lib/dispatch'
import { useT, useRelativeTime } from '@/lib/i18n'
import { archLabel, osLabel, powerTone } from '@/lib/machine'
import { fleetCapacity } from '@/lib/metrics'
import { useStats, useTasks, useWorkers } from '@/lib/queries'
import type { WorkerInfo } from '@/lib/types'
import { buildTopology, withLiveActivity } from '@/lib/topology'
import { cn, formatBytes, formatPercent } from '@/lib/utils'
import { useRealtimeStore } from '@/store/useRealtimeStore'

export const Route = createFileRoute('/workers/')({
  component: Workers,
})

const TASK_WINDOW = 200

function Workers() {
  const t = useT()
  const { data, isLoading } = useWorkers()
  const { data: tasksData } = useTasks(TASK_WINDOW)
  const { data: stats } = useStats()
  const events = useRealtimeStore((state) => state.events)
  const registry = useMemo(() => data?.workers ?? [], [data])

  // Polled recent tasks plus live WebSocket frames, so a dispatch shows up
  // on the map the moment it happens.
  const liveTasks = useMemo(
    () => mergeLiveTasks(tasksData?.tasks ?? [], events.map((event) => event.message)),
    [tasksData, events],
  )
  // One source for "busy": the strip, the cards and the map all count live tasks.
  const workers = useMemo(() => withLiveActivity(registry, liveTasks), [registry, liveTasks])
  const topology = useMemo(() => buildTopology(workers, liveTasks), [workers, liveTasks])
  const now = useNow(topology.running > 0 ? 250 : 1000)

  // Several workers on one machine share a host name; label them by id then.
  const names = useMemo(() => uniqueWorkerNames(workers.map((w) => ({ id: w.id, host: w.host }))), [workers])
  // The registry returns workers in no stable order; keep cards still between
  // polls, in the same order as the dispatch map above.
  const sortedWorkers = useMemo(() => [...workers].sort((a, b) => a.id.localeCompare(b.id)), [workers])

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('workers.eyebrow')} title={t('workers.title')} actions={<AddWorkerDialog />} />

      {workers.length > 0 && <StatStrip cells={fleetCells(t, workers)} />}

      {workers.length > 0 && (
        <Panel
          title={t('workers.flowTitle')}
          eyebrow={topology.running > 0 ? t('workers.flowLive') : t('workers.flowIdle')}
          action={
            topology.running > 0 && (
              <span className="flex items-center gap-1.5 rounded-full border border-[color-mix(in_oklch,var(--color-status-running)_35%,transparent)] px-2 py-0.5 text-[11px] font-medium text-(--color-status-running)">
                <span className="h-1.5 w-1.5 rounded-full bg-(--color-status-running) animate-(--animate-heartbeat)" />
                {t('workers.flowRunning', { n: topology.running })}
              </span>
            )
          }
          padded={false}
          className="overflow-hidden"
        >
          <WorkerFlow topology={topology} queued={stats?.queued_tasks ?? 0} now={now} />
        </Panel>
      )}

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('workers.loading')}</div>}

      {!isLoading && workers.length === 0 && (
        <Panel>
          <div className="py-6 text-center text-[12px] text-(--color-ink-muted)">{t('workers.empty')}</div>
        </Panel>
      )}

      <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
        {sortedWorkers.map((worker) => (
          <WorkerCard key={worker.id} worker={worker} name={names.get(worker.id) ?? worker.id} />
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
      color: healthy < workers.length ? 'var(--color-status-failed)' : 'var(--color-status-success)',
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
  /** Unique display name (host, or id when several workers share a host). */
  name: string
}

/**
 * Compact worker summary for the fleet grid (four per row on wide
 * screens): identity, live slots, CPU/RAM, and one line of history. The
 * full machine view — files compiling, capabilities, address — lives on
 * the worker page the whole card links to.
 */
function WorkerCard({ worker, name }: WorkerCardProps) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const power = powerTone(worker)
  // The machine drawing already shows the OS; the line keeps what it cannot.
  const platform = [osLabel(worker.os), archLabel(worker.architecture)].filter(Boolean).join(' · ')
  const specs = [archLabel(worker.architecture), t('workers.cores', { n: worker.cpu_cores }), formatBytes(worker.memory_gb)]
    .filter(Boolean)
    .join(' · ')
  const busy = worker.active_tasks

  return (
    <Link
      to="/workers/$workerId"
      params={{ workerId: worker.id }}
      title={`${worker.id} · ${worker.address || '—'}`}
      className={cn(
        'flex flex-col gap-2.5 rounded-(--radius-lg) border bg-(--color-surface) p-3 transition-colors',
        power === 'failed'
          ? 'border-[color-mix(in_oklch,var(--color-status-failed)_40%,transparent)]'
          : 'border-(--color-hairline) hover:border-(--color-hairline-strong)',
      )}
    >
      <div className="flex items-center gap-2.5">
        <MachineIllustration
          os={worker.os}
          slots={worker.max_parallel_tasks}
          busy={busy}
          power={power}
          caption=""
          label={name}
          className="w-11 shrink-0"
        />
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-1.5">
            <span className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: POWER_DOT[power] }} />
            <span className="truncate font-mono text-[12px] font-semibold text-(--color-ink)">{name}</span>
          </div>
          <div className="truncate text-[11px] text-(--color-ink-muted)" title={`${platform} · ${specs}`}>
            {specs}
          </div>
        </div>
        <span
          className={cn(
            'self-start font-tabular text-[12px] font-medium',
            busy > 0 ? 'text-(--color-status-running)' : 'text-(--color-ink-faint)',
          )}
          title={t('workers.utilizationLabel', { active: busy, max: worker.max_parallel_tasks })}
        >
          {busy}/{worker.max_parallel_tasks}
        </span>
      </div>

      <UnityCapabilityBadge worker={worker} />
      <WorkerUsage worker={worker} variant="compact" />

      <div className="flex items-center justify-between gap-2 border-t border-(--color-hairline) pt-2 text-[11px] text-(--color-ink-muted)">
        <span className="truncate font-tabular">
          {worker.total_tasks > 0
            ? t('workers.cardStats', { rate: formatPercent(worker.success_rate), n: worker.total_tasks })
            : t('workers.cardNoTasks')}
        </span>
        <span className="shrink-0 text-(--color-ink-faint)" title={t('workers.seenLabel', { t: relativeTime(worker.last_seen) })}>
          {power === 'failed' ? (
            <span className="text-(--color-status-failed)">{t('workers.unhealthy')}</span>
          ) : (
            relativeTime(worker.last_seen)
          )}
        </span>
      </div>
    </Link>
  )
}

const POWER_DOT: Record<ReturnType<typeof powerTone>, string> = {
  success: 'var(--color-status-success)',
  warning: 'var(--color-status-timeout)',
  failed: 'var(--color-status-failed)',
}
