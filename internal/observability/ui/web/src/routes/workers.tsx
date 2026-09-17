import { createFileRoute } from '@tanstack/react-router'
import { Cpu, MemoryStick } from 'lucide-react'
import { AddWorkerDialog } from '@/components/AddWorkerDialog'
import { Panel } from '@/components/Panel'
import { PageHeader } from '@/components/PageHeader'
import { CircuitBadge } from '@/components/StatusBadge'
import { WorkerHeartbeat } from '@/components/WorkerHeartbeat'
import { ExecutorSlots } from '@/components/ExecutorSlots'
import { useT, useRelativeTime } from '@/lib/i18n'
import { useWorkers } from '@/lib/queries'
import { formatBytes } from '@/lib/utils'
import type { WorkerInfo } from '@/lib/types'

export const Route = createFileRoute('/workers')({
  component: Workers,
})

function Workers() {
  const t = useT()
  const { data, isLoading } = useWorkers()
  const workers = data?.workers ?? []

  return (
    <div className="flex flex-col gap-5">
      <PageHeader eyebrow={t('workers.eyebrow')} title={t('workers.title')} actions={<AddWorkerDialog />} />

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('workers.loading')}</div>}

      {!isLoading && workers.length === 0 && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('workers.empty')}</div>
        </Panel>
      )}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
        {workers.map((worker) => (
          <WorkerCard key={worker.id} worker={worker} />
        ))}
      </div>
    </div>
  )
}

function WorkerCard({ worker }: { worker: WorkerInfo }) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const loadRatio = worker.max_parallel_tasks > 0 ? worker.active_tasks / worker.max_parallel_tasks : 0

  return (
    <Panel padded={false}>
      <div className="flex items-start justify-between px-4 py-3">
        <div className="min-w-0">
          <div className="truncate font-mono text-[13px] font-semibold text-(--color-ink)">{worker.id}</div>
          <div className="text-[11px] text-(--color-ink-muted)">
            {worker.os} · {worker.architecture}
          </div>
        </div>
        <CircuitBadge state={worker.circuit_state} />
      </div>

      <WorkerHeartbeat
        workerId={worker.id}
        circuitState={worker.circuit_state}
        loadRatio={loadRatio}
        className="px-4"
      />

      <div className="border-t border-(--color-hairline) px-4 py-3">
        <ExecutorSlots active={worker.active_tasks} max={worker.max_parallel_tasks} />
      </div>

      <div className="grid grid-cols-2 gap-3 border-t border-(--color-hairline) px-4 py-3 text-[11px]">
        <div className="flex items-center gap-1.5 text-(--color-ink-muted)">
          <Cpu className="h-3 w-3" /> {t('workers.cores', { n: worker.cpu_cores })}
        </div>
        <div className="flex items-center gap-1.5 text-(--color-ink-muted)">
          <MemoryStick className="h-3 w-3" /> {formatBytes(worker.memory_gb)}
        </div>
        <div className="text-(--color-ink-muted)">
          {t('workers.successLabel')}{' '}
          <span className="font-tabular text-(--color-ink-secondary)">
            {Math.round(worker.success_rate * 100)}%
          </span>
        </div>
        <div className="text-(--color-ink-muted)">
          {t('workers.latencyLabel')}{' '}
          <span className="font-tabular text-(--color-ink-secondary)">
            {Math.round(worker.avg_latency_ms)}ms
          </span>
        </div>
        <div className="col-span-2 text-(--color-ink-muted)">
          {t('workers.seenLabel', { t: relativeTime(worker.last_seen) })}
        </div>
      </div>

      {(worker.compilers?.length ?? 0) > 0 && (
        <div className="flex flex-wrap gap-1 border-t border-(--color-hairline) px-4 py-2.5">
          {(worker.compilers ?? []).map((compiler) => (
            <span
              key={compiler}
              className="rounded-(--radius-xs) bg-(--color-surface-inset) px-1.5 py-0.5 font-mono text-[10px] text-(--color-ink-muted)"
            >
              {compiler}
            </span>
          ))}
        </div>
      )}
    </Panel>
  )
}
