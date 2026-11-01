import { useT } from '@/lib/i18n'
import type { TimelineRow } from '@/lib/timeline'
import { TONE_COLOR, taskToneClass } from '@/lib/timeline'
import type { TaskInfo } from '@/lib/types'
import { cn, formatDuration } from '@/lib/utils'

interface TimelineTrackProps {
  rows: TimelineRow[]
  onSelect?: (task: TaskInfo) => void
  size?: 'md' | 'sm'
  /** Where the hover card opens; 'bottom' for rows with no headroom inside a scroller. */
  placement?: 'top' | 'bottom'
}

// Past this point a left-anchored card would run off the track's right edge.
const FLIP_AT_PCT = 60

/**
 * One shared row renderer for both the per-build Timeline (one row per
 * task) and the fleet-wide Cluster Activity view (one sub-row per
 * executor slot, several tasks in the same track). A muted "queue"
 * span shows coordinator dispatch wait; the colored "run" span shows
 * the compile, toned by outcome. Hover/focus reveals the timing
 * breakdown — pure CSS, so hundreds of segments cost no listeners.
 */
export function TimelineTrack({ rows, onSelect, size = 'md', placement = 'top' }: TimelineTrackProps) {
  return (
    <div
      className={cn(
        'relative w-full rounded-(--radius-xs) bg-(--color-surface-inset)',
        size === 'md' ? 'h-5' : 'h-3.5',
      )}
    >
      {rows.map((row) => {
        const tone = taskToneClass(row.task.status, row.task.exit_code)
        const color = TONE_COLOR[tone]
        const flip = row.runLeft > FLIP_AT_PCT
        return (
          <span key={row.task.id}>
            {row.queueWidth > 0 && (
              <span
                className="absolute top-1/2 h-1 -translate-y-1/2 rounded-full bg-(--color-ink-faint) opacity-50"
                style={{ left: `${row.queueLeft}%`, width: `${row.queueWidth}%` }}
              />
            )}
            <span
              className="group absolute inset-y-0 hover:z-20 focus-within:z-20"
              style={{ left: `${row.runLeft}%`, width: `${row.runWidth}%` }}
            >
              <button
                type="button"
                onClick={() => onSelect?.(row.task)}
                aria-label={`Task ${row.task.id} · ${tone}`}
                className={cn(
                  'absolute inset-x-0 top-1/2 min-w-[3px] -translate-y-1/2 cursor-pointer rounded-full transition-[filter] hover:brightness-125 focus-visible:outline-offset-1',
                  size === 'md' ? 'h-2.5' : 'h-2',
                  tone === 'running' && 'animate-(--animate-heartbeat)',
                )}
                style={{ backgroundColor: color }}
              />
              <TaskHoverCard
                task={row.task}
                tone={tone}
                className={cn(flip ? 'right-0' : 'left-0', placement === 'top' ? 'bottom-full mb-2' : 'top-full mt-2')}
              />
            </span>
          </span>
        )
      })}
    </div>
  )
}

interface TaskHoverCardProps {
  task: TaskInfo
  tone: ReturnType<typeof taskToneClass>
  className?: string
}

function TaskHoverCard({ task, tone, className }: TaskHoverCardProps) {
  const t = useT()
  const total = task.duration_ms ?? (task.completed_at_ms ? task.completed_at_ms - task.started_at_ms : undefined)

  return (
    <span
      role="tooltip"
      className={cn(
        'pointer-events-none invisible absolute w-60 translate-y-1 rounded-(--radius-md) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) p-2.5 text-left opacity-0 shadow-[0_8px_24px_-8px_oklch(0_0_0/50%)] transition-[opacity,transform] duration-150 ease-[cubic-bezier(0.23,1,0.32,1)] group-focus-within:visible group-focus-within:translate-y-0 group-focus-within:opacity-100 group-hover:visible group-hover:translate-y-0 group-hover:opacity-100',
        className,
      )}
    >
      <span className="flex items-center gap-1.5">
        <span className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: TONE_COLOR[tone] }} />
        <span className="truncate font-mono text-[11px] font-medium text-(--color-ink)">{task.id}</span>
      </span>
      <span className="mt-1 block truncate font-mono text-[10px] text-(--color-ink-faint)">
        {task.build_id || '—'} · {task.worker_id || t('overview.eventUnassigned')}
      </span>
      <span className="mt-2 grid grid-cols-3 gap-2 border-t border-(--color-hairline) pt-2">
        <Figure label={t('buildDetail.colQueue')} value={formatDuration(task.queue_ms)} />
        <Figure label={t('buildDetail.colCompile')} value={task.compile_ms ? formatDuration(task.compile_ms) : '—'} />
        <Figure
          label={t('buildDetail.colDuration')}
          value={total !== undefined ? formatDuration(total) : t(`status.${tone}`)}
        />
      </span>
      {task.error_message && (
        <span className="mt-2 line-clamp-2 block text-[10px] text-(--color-status-failed)">{task.error_message}</span>
      )}
    </span>
  )
}

function Figure({ label, value }: { label: string; value: string }) {
  return (
    <span className="flex flex-col">
      <span className="text-[9px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{label}</span>
      <span className="font-tabular text-[11px] text-(--color-ink-secondary)">{value}</span>
    </span>
  )
}
