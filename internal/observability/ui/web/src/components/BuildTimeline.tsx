import { useMemo } from 'react'
import { TimelineTrack } from '@/components/TimelineTrack'
import { useT } from '@/lib/i18n'
import { layoutTimeline } from '@/lib/timeline'
import type { TaskInfo } from '@/lib/types'

interface BuildTimelineProps {
  tasks: TaskInfo[]
  onSelectTask?: (taskId: string) => void
}

export function BuildTimeline({ tasks, onSelectTask }: BuildTimelineProps) {
  const t = useT()
  const layout = useMemo(() => layoutTimeline(tasks), [tasks])

  if (layout.rows.length === 0) {
    return <div className="p-4 text-center text-[12px] text-(--color-ink-muted)">{t('buildDetail.timelineEmpty')}</div>
  }

  return (
    <div className="flex flex-col gap-2 p-4">
      {layout.rows.map((row) => (
        <div key={row.task.id} className="flex items-center gap-3">
          <button
            type="button"
            onClick={() => onSelectTask?.(row.task.id)}
            className="w-40 shrink-0 truncate text-left font-mono text-[11px] text-(--color-ink-secondary) hover:text-(--color-signal)"
            title={row.task.id}
          >
            {row.task.id}
          </button>
          <div className="flex-1">
            <TimelineTrack rows={[row]} onSelect={onSelectTask} />
          </div>
        </div>
      ))}

      <div className="mt-1 flex flex-wrap items-center gap-3 text-[10px] text-(--color-ink-faint)">
        <Legend swatchClass="bg-(--color-ink-faint) opacity-40" label={t('buildDetail.legendQueue')} />
        <Legend swatchClass="bg-(--color-status-success)" label={t('buildDetail.legendSuccess')} />
        <Legend swatchClass="bg-(--color-status-failed)" label={t('buildDetail.legendFailed')} />
        <Legend swatchClass="bg-(--color-status-running)" label={t('buildDetail.legendRunning')} />
      </div>
    </div>
  )
}

function Legend({ swatchClass, label }: { swatchClass: string; label: string }) {
  return (
    <span className="flex items-center gap-1.5">
      <span className={`h-1.5 w-3 rounded-full ${swatchClass}`} />
      {label}
    </span>
  )
}
