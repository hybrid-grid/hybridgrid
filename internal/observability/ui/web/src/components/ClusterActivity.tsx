import { useNavigate } from '@tanstack/react-router'
import { useMemo } from 'react'
import { TimelineTrack } from '@/components/TimelineTrack'
import { useNow } from '@/hooks/useNow'
import { useT } from '@/lib/i18n'
import { layoutCluster, type ClusterLane } from '@/lib/timeline'
import type { TaskInfo, WorkerInfo } from '@/lib/types'
import { cn, formatClock } from '@/lib/utils'

interface ClusterActivityProps {
  tasks: TaskInfo[]
  workers: WorkerInfo[]
}

/**
 * The overview's focal view: where compile time is going right now.
 * One swimlane per worker, one sub-row per executor slot actually
 * used, on a shared wall-clock axis. Running bars keep growing between
 * polls so a stuck compile is visibly the longest bar on screen.
 */
export function ClusterActivity({ tasks, workers }: ClusterActivityProps) {
  const t = useT()
  const navigate = useNavigate()
  const now = useNow(1000)
  const layout = useMemo(() => layoutCluster(tasks, workers, now), [tasks, workers, now])
  const dispatchedCount = layout.lanes.reduce((sum, lane) => sum + lane.taskCount, 0)

  if (dispatchedCount === 0) {
    return (
      <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('overview.noClusterActivity')}</div>
    )
  }

  function openTask(task: TaskInfo): void {
    if (task.build_id) {
      navigate({ to: '/builds/$buildId', params: { buildId: task.build_id }, search: { task: task.id } })
    } else {
      navigate({ to: '/tasks/$taskId/console', params: { taskId: task.id } })
    }
  }

  return (
    <div className="flex flex-col px-4 pb-4 pt-3">
      <div className="mb-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-[11px] text-(--color-ink-muted)">
        <span className="font-tabular">{t('overview.clusterActivityNote', { peak: layout.peak, n: dispatchedCount })}</span>
        {layout.undispatched > 0 && (
          <span className="font-tabular text-(--color-status-timeout)">
            {t('overview.clusterUndispatched', { n: layout.undispatched })}
          </span>
        )}
        <Legend />
      </div>

      <div className="grid grid-cols-[minmax(0,88px)_minmax(0,1fr)] gap-x-3 sm:grid-cols-[minmax(0,160px)_minmax(0,1fr)] sm:gap-x-4">
        <div className="text-[10px] uppercase tracking-[0.08em] text-(--color-ink-faint)">{t('overview.clusterAxis')}</div>
        <Axis ticks={layout.ticks} />

        {layout.lanes.map((lane) => (
          <Lane key={lane.id} lane={lane} onSelect={openTask} />
        ))}
      </div>
    </div>
  )
}

function Axis({ ticks }: { ticks: number[] }) {
  return (
    <div className="relative mb-1 h-4">
      {ticks.map((tick, i) => (
        <span
          key={tick}
          className={cn(
            'absolute top-0 font-mono text-[10px] text-(--color-ink-faint)',
            i === 0 ? 'translate-x-0' : i === ticks.length - 1 ? '-translate-x-full' : '-translate-x-1/2',
            // Narrow screens keep only the ends and the midpoint so labels never collide.
            i % 2 === 1 && 'hidden sm:block',
          )}
          style={{ left: `${(i / (ticks.length - 1)) * 100}%` }}
        >
          {formatClock(tick)}
        </span>
      ))}
    </div>
  )
}

function Lane({ lane, onSelect }: { lane: ClusterLane; onSelect: (task: TaskInfo) => void }) {
  const t = useT()
  const idle = lane.taskCount === 0

  return (
    <>
      <div className="flex min-w-0 items-start gap-2 border-t border-(--color-hairline) py-2">
        <span
          className="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full"
          style={{
            backgroundColor: !lane.known
              ? 'var(--color-ink-faint)'
              : lane.healthy
                ? 'var(--color-signal)'
                : 'var(--color-status-failed)',
          }}
        />
        <div className="min-w-0">
          <div
            className={cn(
              'truncate font-mono text-[11px]',
              lane.known ? 'text-(--color-ink-secondary)' : 'text-(--color-ink-faint)',
            )}
            title={lane.id}
          >
            {lane.host}
          </div>
          <div className="text-[10px] text-(--color-ink-faint)">
            {idle ? t('overview.idle') : t('overview.taskCount', { n: lane.taskCount })}
          </div>
        </div>
      </div>
      <div className="flex min-w-0 flex-col justify-center gap-1 border-t border-(--color-hairline) py-2">
        {idle ? (
          <div className="h-3.5 rounded-(--radius-xs) border border-dashed border-(--color-hairline)" />
        ) : (
          lane.slotRows.map((rows, slot) => <TimelineTrack key={slot} rows={rows} onSelect={onSelect} size="sm" />)
        )}
      </div>
    </>
  )
}

function Legend() {
  const t = useT()
  const items = [
    { label: t('buildDetail.legendQueue'), className: 'h-1 bg-(--color-ink-faint) opacity-50' },
    { label: t('buildDetail.legendSuccess'), className: 'h-1.5 bg-(--color-status-success)' },
    { label: t('buildDetail.legendRunning'), className: 'h-1.5 bg-(--color-status-running)' },
    { label: t('buildDetail.legendFailed'), className: 'h-1.5 bg-(--color-status-failed)' },
  ]
  return (
    <span className="ml-auto flex flex-wrap items-center gap-3 text-[10px] text-(--color-ink-faint)">
      {items.map((item) => (
        <span key={item.label} className="flex items-center gap-1.5">
          <span className={cn('w-3 rounded-full', item.className)} />
          {item.label}
        </span>
      ))}
    </span>
  )
}
