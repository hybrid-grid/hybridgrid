import { Link } from '@tanstack/react-router'
import { useT } from '@/lib/i18n'
import { normalizeStatus, STATUS_COLOR } from '@/lib/status'
import type { TaskInfo } from '@/lib/types'
import { formatDuration } from '@/lib/utils'
import type { RealtimeEvent } from '@/store/useRealtimeStore'

interface EventLogProps {
  events: RealtimeEvent[]
}

/**
 * Task lifecycle as it streams in: wall-clock time, outcome dot, task
 * id, and a one-line sentence (where it ran, how long it took). Stats
 * frames are deliberately absent — every 2s they would drown the log.
 */
export function EventLog({ events }: EventLogProps) {
  const t = useT()

  if (events.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('overview.waitingEvents')}</div>
  }

  return (
    <ol className="divide-y divide-(--color-hairline)">
      {events.map((event) => (
        <EventRow key={event.id} event={event} />
      ))}
    </ol>
  )
}

function EventRow({ event }: { event: RealtimeEvent }) {
  const t = useT()
  const task = event.message.data as TaskInfo | undefined
  if (!task) return null

  const started = event.message.type === 'task_started'
  const tone = started ? 'running' : normalizeStatus(task.status, task.exit_code)
  const worker = task.worker_id || t('overview.eventUnassigned')
  const duration = formatDuration(task.duration_ms ?? 0)
  const sentence = started
    ? t('overview.eventStarted', { worker })
    : tone === 'failed' || tone === 'timeout'
      ? t('overview.eventFailed', { d: duration })
      : t('overview.eventSucceeded', { d: duration })

  return (
    <li className="grid grid-cols-[64px_8px_minmax(0,1fr)] items-baseline gap-x-3 px-4 py-2 text-[12px]">
      <time className="font-mono text-[10px] text-(--color-ink-faint)">
        {new Date(event.message.timestamp * 1000).toLocaleTimeString([], { hour12: false })}
      </time>
      <span className="h-1.5 w-1.5 translate-y-[-1px] rounded-full" style={{ backgroundColor: STATUS_COLOR[tone] }} />
      <div className="min-w-0">
        <div className="flex min-w-0 items-baseline gap-2">
          {task.build_id ? (
            <Link
              to="/builds/$buildId"
              params={{ buildId: task.build_id }}
              search={{ task: task.id }}
              className="truncate font-mono text-(--color-ink-secondary) hover:text-(--color-signal)"
            >
              {task.id}
            </Link>
          ) : (
            <span className="truncate font-mono text-(--color-ink-secondary)">{task.id}</span>
          )}
          {task.from_cache && (
            <span className="shrink-0 text-[10px] uppercase tracking-[0.06em] text-(--color-signal)">
              {t('overview.eventFromCache')}
            </span>
          )}
        </div>
        <div className="truncate text-[11px] text-(--color-ink-muted)">
          {sentence}
          {task.error_message && <span className="text-(--color-status-failed)"> · {task.error_message}</span>}
        </div>
      </div>
    </li>
  )
}
