import { useMemo } from 'react'
import { useNow } from '@/hooks/useNow'
import { useT } from '@/lib/i18n'
import { layoutTimeline } from '@/lib/timeline'
import type { TaskInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'
import { TimelineTrack } from './TimelineTrack'

const SCROLL_AFTER_ROWS = 16
const BELOW_FIRST_ROWS = 3

interface BuildTimelineProps {
  tasks: TaskInfo[]
  selectedTaskId?: string
  onSelectTask?: (task: TaskInfo) => void
}

export function BuildTimeline({ tasks, selectedTaskId, onSelectTask }: BuildTimelineProps) {
  const t = useT()
  const now = useNow(1000)
  const layout = useMemo(() => layoutTimeline(tasks, now), [tasks, now])

  // Short builds render unclipped; long ones scroll, and inside a
  // scroller the first rows have no headroom, so their cards open below.
  const scrolls = layout.rows.length > SCROLL_AFTER_ROWS

  if (layout.rows.length === 0) {
    return <div className="px-4 py-8 text-center text-[12px] text-(--color-ink-muted)">{t('buildDetail.timelineEmpty')}</div>
  }

  return (
    <div className="flex flex-col gap-1.5 p-4">
      <div className="mb-1 grid grid-cols-[minmax(0,96px)_minmax(0,1fr)] sm:grid-cols-[minmax(0,180px)_minmax(0,1fr)] gap-x-3 text-[10px] text-(--color-ink-faint)">
        <span className="font-tabular">
          {t('buildDetail.spanLabel', { d: formatDuration(layout.windowEnd - layout.windowStart) })}
        </span>
        <span className="flex justify-between font-mono">
          <span>{formatClock(layout.windowStart)}</span>
          <span>{formatClock(layout.windowEnd)}</span>
        </span>
      </div>

      <div className={cn(scrolls && 'max-h-[420px] overflow-y-auto pr-1')}>
        <div className="flex flex-col gap-1.5">
          {layout.rows.map((row, index) => (
            <div key={row.task.id} className="grid grid-cols-[minmax(0,96px)_minmax(0,1fr)] sm:grid-cols-[minmax(0,180px)_minmax(0,1fr)] items-center gap-x-3">
              <button
                type="button"
                onClick={() => onSelectTask?.(row.task)}
                className={cn(
                  'truncate text-left font-mono text-[11px] transition-colors hover:text-(--color-signal)',
                  row.task.id === selectedTaskId ? 'text-(--color-signal)' : 'text-(--color-ink-muted)',
                )}
                title={row.task.id}
              >
                {row.task.id}
              </button>
              <TimelineTrack
                rows={[row]}
                onSelect={onSelectTask}
                placement={scrolls && index < BELOW_FIRST_ROWS ? 'bottom' : 'top'}
              />
            </div>
          ))}
        </div>
      </div>

      <div className="mt-2 flex flex-wrap items-center gap-3 text-[10px] text-(--color-ink-faint)">
        <Legend swatchClass="h-1 bg-(--color-ink-faint) opacity-50" label={t('buildDetail.legendQueue')} />
        <Legend swatchClass="h-1.5 bg-(--color-status-success)" label={t('buildDetail.legendSuccess')} />
        <Legend swatchClass="h-1.5 bg-(--color-status-running)" label={t('buildDetail.legendRunning')} />
        <Legend swatchClass="h-1.5 bg-(--color-status-failed)" label={t('buildDetail.legendFailed')} />
      </div>
    </div>
  )
}

function Legend({ swatchClass, label }: { swatchClass: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5">
      <span className={cn('w-3 rounded-full', swatchClass)} />
      {label}
    </span>
  )
}
