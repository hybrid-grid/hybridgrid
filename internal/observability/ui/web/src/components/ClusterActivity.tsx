import { useNavigate } from '@tanstack/react-router'
import { useMemo, useState } from 'react'
import { Segmented } from '@/components/FilterControls'
import { useNow } from '@/hooks/useNow'
import { useT } from '@/lib/i18n'
import { powerTone } from '@/lib/machine'
import { activityBursts, tasksInWindow } from '@/lib/timeline'
import type { TaskInfo, WorkerInfo } from '@/lib/types'
import { utilization, type Utilization, type UtilizationRow } from '@/lib/utilization'
import { cn, formatClock, formatDuration, tailEllipsis } from '@/lib/utils'

interface ClusterActivityProps {
  tasks: TaskInfo[]
  workers: WorkerInfo[]
}

const LABEL_COL = 'grid grid-cols-[minmax(0,104px)_minmax(0,1fr)] gap-x-3 sm:grid-cols-[minmax(0,180px)_minmax(0,1fr)] sm:gap-x-4'

/**
 * How busy the cluster has been, at a glance: a cluster-wide chart of
 * threads in use against capacity, then one thin heat strip per worker
 * (darker = busier, red tick = a failed compile). Unlike one bar per
 * task, this stays readable for many machines and thousands of
 * compiles, and answers "is the farm saturated, which machine is idle,
 * when did things fail".
 */
export function ClusterActivity({ tasks, workers }: ClusterActivityProps) {
  const t = useT()
  const now = useNow(1000)
  const [scope, setScope] = useState<'latest' | 'all'>('latest')
  const bursts = useMemo(() => activityBursts(tasks, now), [tasks, now])

  const data = useMemo(() => {
    if (bursts.length === 0) return undefined
    const live = tasks.some((task) => !task.completed_at_ms && task.status === 'running')
    // Default to the newest burst: builds seconds long but minutes apart
    // would otherwise be squeezed into slivers at the edges of one axis.
    const latest = scope === 'latest' || bursts.length === 1
    const window = latest
      ? { start: bursts[0].start, end: live ? now : bursts[0].end }
      : { start: bursts[bursts.length - 1].start, end: live ? now : bursts[0].end }
    const visible = latest ? tasksInWindow(tasks, bursts[0], now) : tasks
    return utilization(visible, workers, window, now)
  }, [bursts, scope, tasks, workers, now])

  if (!data || data.taskCount === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('overview.noClusterActivity')}</div>
  }

  return (
    <div className="flex flex-col px-4 pb-4 pt-3">
      <div className="mb-3 flex flex-wrap items-center gap-x-3 gap-y-2 text-[12px] text-(--color-ink-muted)">
        <span className="font-tabular">
          {t('overview.clusterSummary', { peak: data.peak, capacity: data.capacity, n: data.taskCount })}
        </span>
        {bursts.length > 1 && (
          <Segmented
            label={t('overview.clusterActivityTitle')}
            value={scope}
            onChange={setScope}
            options={[
              { value: 'latest', label: t('overview.clusterScopeLatest') },
              { value: 'all', label: t('overview.clusterScopeAll', { n: bursts.length }) },
            ]}
          />
        )}
        <Legend />
      </div>

      <div className={LABEL_COL}>
        <ClusterRow data={data} />
        <div className="col-span-2 h-3" />
        {data.rows.map((row) => (
          <WorkerRow key={row.id} row={row} data={data} />
        ))}
        <div />
        <Axis data={data} now={now} />
      </div>
    </div>
  )
}

/** Cluster-wide threads in use, as an area against the capacity line. */
function ClusterRow({ data }: { data: Utilization }) {
  const t = useT()
  const n = data.total.length
  const top = Math.max(data.capacity, ...data.total, 1)
  const y = (v: number) => 100 - (v / top) * 100
  const area = `M0 100 ${data.total.map((v, i) => `L${i} ${y(v)} L${i + 1} ${y(v)}`).join(' ')} L${n} 100 Z`
  const current = data.total[n - 1] ?? 0

  return (
    <>
      <div className="flex min-w-0 flex-col justify-center">
        <div className="text-[12px] font-medium text-(--color-ink)">{t('overview.clusterTotal')}</div>
        <div className="font-tabular text-[11px] text-(--color-ink-muted)">
          {t('overview.clusterBusyNow', { busy: Math.round(current), max: data.capacity })}
        </div>
      </div>
      <div className="relative h-16">
        <svg viewBox={`0 0 ${n} 100`} preserveAspectRatio="none" className="absolute inset-0 h-full w-full" aria-hidden>
          <path d={area} fill="var(--color-status-running)" fillOpacity="0.16" />
          <path
            d={data.total.map((v, i) => `${i === 0 ? 'M' : 'L'}${i} ${y(v)} L${i + 1} ${y(v)}`).join(' ')}
            fill="none"
            stroke="var(--color-status-running)"
            strokeWidth="1.5"
            vectorEffect="non-scaling-stroke"
          />
          {data.capacity > 0 && (
            <line
              x1="0"
              x2={n}
              y1={y(data.capacity)}
              y2={y(data.capacity)}
              stroke="var(--color-ink-faint)"
              strokeDasharray="4 4"
              vectorEffect="non-scaling-stroke"
            />
          )}
        </svg>
        {data.capacity > 0 && (
          <span className="absolute right-1 top-1 rounded-(--radius-xs) bg-(--color-surface) px-1 font-tabular text-[10px] text-(--color-ink-faint)">
            {t('overview.clusterCapacity', { n: data.capacity })}
          </span>
        )}
      </div>
    </>
  )
}

function WorkerRow({ row, data }: { row: UtilizationRow; data: Utilization }) {
  const t = useT()
  const navigate = useNavigate()
  const tone = row.worker ? powerTone(row.worker) : 'failed'
  const dot = tone === 'success' ? 'var(--color-status-success)' : tone === 'warning' ? 'var(--color-status-timeout)' : 'var(--color-status-failed)'
  const open = () => navigate({ to: '/workers/$workerId', params: { workerId: row.id } })

  return (
    <>
      <button type="button" onClick={open} className="flex min-w-0 items-center gap-2 py-1 text-left" title={row.id}>
        <span className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: dot }} />
        <span className="min-w-0 flex-1 truncate font-mono text-[11px] text-(--color-ink-secondary) hover:text-(--color-ink)">
          {/* Narrow screens keep the name's tail (worker-node-04 → …node-04): the prefix is usually shared. */}
          <span className="sm:hidden">{tailEllipsis(row.name, 9)}</span>
          <span className="hidden sm:inline">{row.name}</span>
        </span>
        <span
          className={cn(
            'shrink-0 font-tabular text-[11px]',
            row.runningNow > 0 ? 'text-(--color-status-running)' : 'text-(--color-ink-faint)',
          )}
        >
          {row.runningNow}/{row.max || '—'}
        </span>
      </button>
      <button type="button" onClick={open} className="flex h-[18px] items-stretch gap-px self-center rounded-(--radius-xs)" aria-label={row.name}>
        {row.ratio.map((ratio, b) => (
          <span
            key={b}
            className="relative flex-1 overflow-hidden bg-(--color-surface-inset) first:rounded-l-(--radius-xs) last:rounded-r-(--radius-xs)"
            title={bucketTip(t, data, row, b)}
          >
            {/* Same blue at varying opacity: mixing colours across hues drifts through purple. */}
            {ratio > 0 && (
              <span className="absolute inset-0 bg-(--color-status-running)" style={{ opacity: 0.15 + ratio * 0.85 }} />
            )}
            {row.failures[b] > 0 && <span className="absolute inset-x-0 bottom-0 h-1 bg-(--color-status-failed)" />}
          </span>
        ))}
      </button>
    </>
  )
}

function bucketTip(t: ReturnType<typeof useT>, data: Utilization, row: UtilizationRow, b: number): string {
  const tip = t('overview.clusterTip', {
    time: formatClock(data.start + b * data.bucketMs),
    busy: row.busy[b].toFixed(1).replace(/\.0$/, ''),
    max: row.max || '—',
  })
  return row.failures[b] > 0 ? `${tip} · ${t('overview.clusterTipFailed', { n: row.failures[b] })}` : tip
}

/**
 * Relative time labels ("1m 52s ago … now") — easier to read than clock
 * times. Exact to the second: a minute-rounded "2 min ago" repeated on
 * every tick of a one-minute window says nothing.
 */
function Axis({ data, now }: { data: Utilization; now: number }) {
  const t = useT()
  const ticks = [0, 0.25, 0.5, 0.75, 1].map((f) => data.start + f * (data.end - data.start))
  return (
    <div className="relative mt-1.5 h-4">
      {ticks.map((tick, i) => (
        <span
          key={i}
          className={cn(
            'absolute top-0 whitespace-nowrap font-tabular text-[10px] text-(--color-ink-faint)',
            i === 0 ? 'translate-x-0' : i === ticks.length - 1 ? '-translate-x-full' : '-translate-x-1/2',
            i % 2 === 1 && 'hidden sm:block',
          )}
          style={{ left: `${(i / (ticks.length - 1)) * 100}%` }}
          title={formatClock(tick)}
        >
          {now - tick < 2000 ? t('overview.clusterNow') : t('overview.clusterAgo', { d: formatAgo(now - tick) })}
        </span>
      ))}
    </div>
  )
}

function Legend() {
  const t = useT()
  return (
    <span className="ml-auto flex flex-wrap items-center gap-3 text-[11px] text-(--color-ink-faint)">
      <span className="flex items-center gap-1.5">
        {t('overview.clusterLegendIdle')}
        <span
          className="relative h-2 w-16 overflow-hidden rounded-full bg-(--color-surface-inset)"
        >
          <span
            className="absolute inset-0 bg-(--color-status-running)"
            style={{ maskImage: 'linear-gradient(90deg, transparent, black)', WebkitMaskImage: 'linear-gradient(90deg, transparent, black)' }}
          />
        </span>
        {t('overview.clusterLegendFull')}
      </span>
      <span className="flex items-center gap-1.5">
        <span className="h-1 w-3 rounded-full bg-(--color-status-failed)" />
        {t('overview.clusterLegendFailed')}
      </span>
    </span>
  )
}

/** Whole seconds under a minute ("37s"), else minutes and seconds ("1m 52s"). */
function formatAgo(ms: number): string {
  const seconds = Math.round(ms / 1000)
  return seconds < 60 ? `${seconds}s` : formatDuration(ms)
}
