import { Link } from '@tanstack/react-router'
import { StatusDot } from '@/components/StatusBadge'
import { useT } from '@/lib/i18n'
import { normalizeStatus } from '@/lib/status'
import type { TaskInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'

interface TaskTableProps {
  tasks: TaskInfo[]
  /** Build-detail mode: rows select in place instead of navigating. */
  selectedId?: string
  onSelect?: (task: TaskInfo) => void
  /** Fleet-wide listings add build + start-time columns; a single build's table does not need them. */
  showBuild?: boolean
  hideWorker?: boolean
}

/**
 * The compile-unit table used by build detail, the Tasks page and the
 * worker page. Errors render inline under the task id, so a failing
 * row explains itself without opening the console.
 */
export function TaskTable({ tasks, selectedId, onSelect, showBuild = false, hideWorker = false }: TaskTableProps) {
  const t = useT()
  return (
    <table className="w-full min-w-[760px] border-collapse text-left text-[12px]">
      <thead className="sticky top-0 z-10 bg-(--color-surface)">
        <tr className="border-b border-(--color-hairline) text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">
          <th className="w-8 py-2 pl-4" aria-label={t('buildDetail.colStatus')} />
          <th className="px-3 py-2 font-medium">{t('buildDetail.colTask')}</th>
          {showBuild && <th className="px-3 py-2 font-medium">{t('tasks.colBuild')}</th>}
          {!hideWorker && <th className="px-3 py-2 font-medium">{t('buildDetail.colWorker')}</th>}
          {showBuild && <th className="px-3 py-2 font-medium">{t('builds.colStarted')}</th>}
          <th className="px-3 py-2 text-right font-medium">{t('buildDetail.colQueue')}</th>
          <th className="px-3 py-2 text-right font-medium">{t('buildDetail.colCompile')}</th>
          <th className="px-3 py-2 text-right font-medium">{t('buildDetail.colDuration')}</th>
          <th className="px-3 py-2 text-right font-medium">{t('buildDetail.colExit')}</th>
          <th className="py-2 pl-3 pr-4 font-medium">{t('buildDetail.colCache')}</th>
        </tr>
      </thead>
      <tbody className="divide-y divide-(--color-hairline)">
        {tasks.map((task) => (
          <TaskRow
            key={task.id}
            task={task}
            selected={task.id === selectedId}
            onSelect={onSelect}
            showBuild={showBuild}
            hideWorker={hideWorker}
          />
        ))}
      </tbody>
    </table>
  )
}

interface TaskRowProps {
  task: TaskInfo
  selected: boolean
  onSelect?: (task: TaskInfo) => void
  showBuild: boolean
  hideWorker: boolean
}

function TaskRow({ task, selected, onSelect, showBuild, hideWorker }: TaskRowProps) {
  const t = useT()
  const tone = normalizeStatus(task.status, task.exit_code)
  const failed = tone === 'failed' || tone === 'timeout'

  return (
    <tr
      onClick={onSelect ? () => onSelect(task) : undefined}
      aria-selected={onSelect ? selected : undefined}
      className={cn(
        'align-top transition-colors hover:bg-(--color-canvas-raised)',
        onSelect && 'cursor-pointer',
        selected && 'bg-(--color-canvas-raised) shadow-[inset_2px_0_0_var(--color-signal)]',
      )}
    >
      <td className="py-2.5 pl-4">
        <StatusDot tone={tone} className="mt-1" />
      </td>
      <td className="max-w-[300px] px-3 py-2.5">
        <TaskIdCell task={task} selectable={Boolean(onSelect)} />
        {task.error_message && (
          <div className="mt-0.5 line-clamp-2 text-[11px] text-(--color-status-failed)" title={task.error_message}>
            {task.error_message}
          </div>
        )}
      </td>
      {showBuild && (
        <td className="max-w-[200px] px-3 py-2.5">
          {task.build_id ? (
            <Link
              to="/builds/$buildId"
              params={{ buildId: task.build_id }}
              className="block truncate font-mono text-[11px] text-(--color-ink-muted) hover:text-(--color-signal)"
              title={task.build_id}
            >
              {task.build_id}
            </Link>
          ) : (
            <span className="text-(--color-ink-faint)">—</span>
          )}
        </td>
      )}
      {!hideWorker && (
        <td className="px-3 py-2.5 font-mono text-[11px]">
          {task.worker_id ? (
            <Link
              to="/workers/$workerId"
              params={{ workerId: task.worker_id }}
              className="text-(--color-ink-muted) hover:text-(--color-signal)"
              onClick={(event) => event.stopPropagation()}
            >
              {task.worker_id}
            </Link>
          ) : (
            <span className="text-(--color-ink-faint)">—</span>
          )}
        </td>
      )}
      {showBuild && (
        <td className="px-3 py-2.5 font-mono text-[11px] text-(--color-ink-muted)">{formatClock(task.started_at_ms)}</td>
      )}
      <td className="px-3 py-2.5 text-right font-tabular text-(--color-ink-secondary)">{formatDuration(task.queue_ms)}</td>
      <td className="px-3 py-2.5 text-right font-tabular text-(--color-ink-secondary)">
        {task.compile_ms ? formatDuration(task.compile_ms) : '—'}
      </td>
      <td className="px-3 py-2.5 text-right font-tabular text-(--color-ink)">
        {task.duration_ms ? formatDuration(task.duration_ms) : tone === 'running' ? t('status.running') : '—'}
      </td>
      <td
        className={cn(
          'px-3 py-2.5 text-right font-mono text-[11px]',
          failed ? 'text-(--color-status-failed)' : 'text-(--color-ink-faint)',
        )}
      >
        {tone === 'running' ? '—' : (task.exit_code ?? 0)}
      </td>
      <td className="py-2.5 pl-3 pr-4">
        {task.from_cache && (
          <span className="rounded-(--radius-xs) bg-(--color-signal-dim) px-1.5 py-0.5 text-[10px] font-medium uppercase tracking-[0.06em] text-(--color-signal)">
            {t('common.cache')}
          </span>
        )}
      </td>
    </tr>
  )
}

function TaskIdCell({ task, selectable }: { task: TaskInfo; selectable: boolean }) {
  const className = 'block max-w-full truncate text-left font-mono text-(--color-ink)'
  if (selectable) {
    return (
      <button type="button" className={className} title={task.id}>
        {task.id}
      </button>
    )
  }
  if (task.build_id) {
    return (
      <Link
        to="/builds/$buildId"
        params={{ buildId: task.build_id }}
        search={{ task: task.id }}
        className={cn(className, 'hover:text-(--color-signal)')}
        title={task.id}
      >
        {task.id}
      </Link>
    )
  }
  return (
    <Link
      to="/tasks/$taskId/console"
      params={{ taskId: task.id }}
      className={cn(className, 'hover:text-(--color-signal)')}
      title={task.id}
    >
      {task.id}
    </Link>
  )
}
