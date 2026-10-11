import { Link } from '@tanstack/react-router'
import { useT } from '@/lib/i18n'
import type { TaskInfo } from '@/lib/types'
import { cn, formatDuration } from '@/lib/utils'

const MAX_LISTED_TASKS = 4

interface ExecutorSlotsProps {
  active: number
  max: number
  /** Running tasks known for this worker, newest first. */
  running?: TaskInfo[]
  now?: number
}

/**
 * Slot grid for at-a-glance occupancy, plus the actual compile units
 * holding those slots — "3/8 busy" alone cannot tell you which file is
 * stuck, the task list underneath can.
 */
export function ExecutorSlots({ active, max, running = [], now }: ExecutorSlotsProps) {
  const t = useT()
  const slots = Array.from({ length: Math.max(max, active, 1) }, (_, i) => i < active)
  const listed = running.slice(0, MAX_LISTED_TASKS)

  return (
    <div className="flex flex-col gap-2">
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">
          {t('workers.slotsLabel')}
        </span>
        <span className="font-tabular text-[11px] text-(--color-ink-muted)">
          {t('workers.utilizationLabel', { active, max })}
        </span>
      </div>
      <div className="flex flex-wrap gap-1">
        {slots.map((busy, i) => (
          <span
            key={i}
            className={cn(
              'h-2.5 w-2.5 rounded-[3px] border',
              busy
                ? 'border-(--color-status-running) bg-(--color-status-running) animate-(--animate-heartbeat)'
                : 'border-(--color-hairline-strong) bg-transparent',
            )}
          />
        ))}
      </div>

      {listed.length > 0 ? (
        <ul className="mt-1 flex flex-col gap-1">
          {listed.map((task) => (
            <li key={task.id} className="flex items-center justify-between gap-2 text-[11px]">
              {task.build_id ? (
                <Link
                  to="/builds/$buildId"
                  params={{ buildId: task.build_id }}
                  search={{ task: task.id }}
                  className="min-w-0 truncate font-mono text-(--color-ink-secondary) hover:text-(--color-signal)"
                  title={task.id}
                >
                  {task.id}
                </Link>
              ) : (
                <span className="min-w-0 truncate font-mono text-(--color-ink-secondary)" title={task.id}>
                  {task.id}
                </span>
              )}
              {now !== undefined && task.started_at_ms > 0 && (
                <span className="shrink-0 font-tabular text-(--color-status-running)">
                  {formatDuration(now - task.started_at_ms)}
                </span>
              )}
            </li>
          ))}
          {running.length > listed.length && (
            <li className="text-[11px] text-(--color-ink-faint)">{t('workers.moreTasks', { n: running.length - listed.length })}</li>
          )}
        </ul>
      ) : (
        active === 0 && <span className="text-[11px] text-(--color-ink-faint)">{t('workers.allIdle')}</span>
      )}
    </div>
  )
}
