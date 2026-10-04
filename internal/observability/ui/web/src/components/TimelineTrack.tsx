import type { TimelineRow } from '@/lib/timeline'
import { TONE_COLOR, taskToneClass } from '@/lib/timeline'

interface TimelineTrackProps {
  rows: TimelineRow[]
  onSelect?: (taskId: string) => void
}

/**
 * One shared row renderer for both the per-build Timeline (one row per
 * task) and the fleet-wide Cluster Activity view (one row per worker,
 * several tasks' segments stacked in the same track). A muted "queue"
 * span shows coordinator dispatch wait; the colored "run" span shows
 * the actual compile, toned by outcome.
 */
export function TimelineTrack({ rows, onSelect }: TimelineTrackProps) {
  return (
    <div className="relative h-5 w-full rounded-(--radius-xs) bg-(--color-surface-inset)">
      {rows.map((row) => {
        const tone = taskToneClass(row.task.status)
        const color = TONE_COLOR[tone]
        return (
          <span key={row.task.id}>
            <span
              className="absolute top-1/2 h-1.5 -translate-y-1/2 rounded-full bg-(--color-ink-faint) opacity-40"
              style={{ left: `${row.queueLeft}%`, width: `${row.queueWidth}%` }}
            />
            <button
              type="button"
              onClick={() => onSelect?.(row.task.id)}
              title={`${row.task.id} · ${tone}`}
              className="absolute top-1/2 h-2.5 -translate-y-1/2 cursor-pointer rounded-full transition-[filter] hover:brightness-125"
              style={{ left: `${row.runLeft}%`, width: `${row.runWidth}%`, backgroundColor: color }}
            />
          </span>
        )
      })}
    </div>
  )
}
