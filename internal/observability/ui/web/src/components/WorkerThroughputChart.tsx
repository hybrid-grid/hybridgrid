import { useMemo } from 'react'
import { useNow } from '@/hooks/useNow'
import { useT } from '@/lib/i18n'
import { niceCeil, workerThroughput } from '@/lib/metrics'
import type { TaskInfo, WorkerInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'

const SERIES_COUNT = 6
const LABEL_EVERY = 3

function seriesColor(index: number): string {
  return `var(--color-series-${(index % SERIES_COUNT) + 1})`
}

interface WorkerThroughputChartProps {
  tasks: TaskInfo[]
  workers: WorkerInfo[]
}

/**
 * Stacked bars, one column per time bucket, one segment per worker.
 * The column height answers "is the farm keeping up", the segment
 * split answers "is the scheduler spreading work" — stacking keeps
 * both readable where grouped bars would turn into slivers at 8+
 * workers and lines would hide each other when balanced.
 */
export function WorkerThroughputChart({ tasks, workers }: WorkerThroughputChartProps) {
  const t = useT()
  const now = useNow(5000)
  const data = useMemo(() => workerThroughput(tasks, now), [tasks, now])
  const hostById = useMemo(() => new Map(workers.map((w) => [w.id, w.host || w.id])), [workers])

  if (data.series.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('overview.throughputEmpty')}</div>
  }

  const axisMax = niceCeil(data.max)
  const finishedTotal = data.totals.reduce((sum, n) => sum + n, 0)

  return (
    <div className="flex flex-col gap-3 p-4">
      <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1">
        <span className="font-tabular text-[20px] font-semibold leading-none text-(--color-ink)">
          {t('overview.throughputTotal', { n: finishedTotal })}
        </span>
        <span className="text-[11px] text-(--color-ink-faint)">
          {t('overview.throughputNote', { bucket: formatDuration(data.bucketMs), n: tasks.length })}
        </span>
      </div>

      <div className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-2">
        <div className="flex h-36 flex-col justify-between text-right font-tabular text-[10px] text-(--color-ink-faint)">
          <span>{axisMax}</span>
          <span>{axisMax / 2}</span>
          <span>0</span>
        </div>
        <div className="relative h-36">
          <Gridlines />
          <div className="absolute inset-0 flex items-end gap-1">
            {data.bucketStarts.map((start, i) => (
              <Column
                key={start}
                start={start}
                end={start + data.bucketMs}
                total={data.totals[i]}
                heightPct={(data.totals[i] / axisMax) * 100}
                flip={i > data.bucketStarts.length / 2}
                segments={data.series.map((s, si) => ({
                  label: hostById.get(s.workerId) ?? s.workerId,
                  count: s.counts[i],
                  color: seriesColor(si),
                }))}
              />
            ))}
          </div>
        </div>
        <span />
        <div className="mt-1 flex gap-1">
          {data.bucketStarts.map((start, i) => (
            <span key={start} className="flex-1 truncate text-center font-mono text-[9px] text-(--color-ink-faint)">
              {i % LABEL_EVERY === 0 ? formatClock(start).slice(0, 5) : ''}
            </span>
          ))}
        </div>
      </div>

      <div className="flex flex-wrap gap-x-4 gap-y-1">
        {data.series.map((s, si) => (
          <span key={s.workerId} className="flex items-center gap-1.5 text-[11px] text-(--color-ink-muted)">
            <span className="h-2 w-2 rounded-[2px]" style={{ backgroundColor: seriesColor(si) }} />
            <span className="font-mono" title={s.workerId}>
              {hostById.get(s.workerId) ?? s.workerId}
            </span>
            <span className="font-tabular text-(--color-ink-faint)">{s.counts.reduce((a, b) => a + b, 0)}</span>
          </span>
        ))}
      </div>
    </div>
  )
}

function Gridlines() {
  return (
    <div className="pointer-events-none absolute inset-0 flex flex-col justify-between">
      {[0, 1, 2].map((i) => (
        <span key={i} className={cn('h-px w-full', i === 2 ? 'bg-(--color-hairline-strong)' : 'bg-(--color-hairline)')} />
      ))}
    </div>
  )
}

interface Segment {
  label: string
  count: number
  color: string
}

interface ColumnProps {
  start: number
  end: number
  total: number
  heightPct: number
  flip: boolean
  segments: Segment[]
}

function Column({ start, end, total, heightPct, flip, segments }: ColumnProps) {
  const t = useT()
  return (
    <div className="group relative flex h-full flex-1 flex-col justify-end">
      <div
        className="flex w-full flex-col-reverse overflow-hidden rounded-t-[3px] transition-[filter] group-hover:brightness-125"
        style={{ height: `${heightPct}%` }}
      >
        {segments
          .filter((s) => s.count > 0)
          .map((s) => (
            <span key={s.label} style={{ flexGrow: s.count, backgroundColor: s.color }} />
          ))}
      </div>
      {total > 0 && (
        <div
          role="tooltip"
          className={cn(
            'pointer-events-none invisible absolute bottom-full z-20 mb-2 w-44 rounded-(--radius-md) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) p-2.5 opacity-0 shadow-[0_8px_24px_-8px_oklch(0_0_0/50%)] transition-opacity duration-150 group-hover:visible group-hover:opacity-100',
            flip ? 'right-0' : 'left-0',
          )}
        >
          <div className="font-mono text-[10px] text-(--color-ink-faint)">
            {formatClock(start)} – {formatClock(end)}
          </div>
          <div className="mb-1.5 text-[12px] font-semibold text-(--color-ink)">{t('overview.throughputTotal', { n: total })}</div>
          {segments
            .filter((s) => s.count > 0)
            .map((s) => (
              <div key={s.label} className="flex items-center justify-between gap-2 text-[11px]">
                <span className="flex min-w-0 items-center gap-1.5 text-(--color-ink-muted)">
                  <span className="h-1.5 w-1.5 shrink-0 rounded-[2px]" style={{ backgroundColor: s.color }} />
                  <span className="truncate font-mono">{s.label}</span>
                </span>
                <span className="font-tabular text-(--color-ink-secondary)">{s.count}</span>
              </div>
            ))}
        </div>
      )}
    </div>
  )
}
