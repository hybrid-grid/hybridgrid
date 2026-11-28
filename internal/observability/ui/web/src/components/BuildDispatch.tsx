import { Link } from '@tanstack/react-router'
import { GitBranch } from 'lucide-react'
import { laneNames, type DispatchLane, type DispatchView } from '@/lib/dispatch'
import { useT } from '@/lib/i18n'
import { powerTone } from '@/lib/machine'
import { normalizeStatus } from '@/lib/status'
import type { TaskInfo } from '@/lib/types'
import { cn, formatDuration, tailEllipsis, taskLabel } from '@/lib/utils'
import { MachineIllustration } from './MachineIllustration'

/** Lanes have a fixed height so the flow lines can be drawn without measuring the DOM. */
const LANE_H = 76
const LANE_GAP = 8
const FLOW_W = 96
const MAX_CHIPS = 4
/**
 * Chip labels keep the file name's tail; the full name is in the tooltip.
 * Sized so MAX_CHIPS chips (dot + label + clock) fit a 1440px lane
 * without the browser truncating the tail again.
 */
const CHIP_CHARS = 8

interface BuildDispatchProps {
  view: DispatchView
  now: number
  onSelectTask?: (task: TaskInfo) => void
}

/**
 * Where a build's compile units are going, as it happens: the build on
 * the left, one lane per worker on the right, and a line between them
 * that animates while the worker holds tasks of this build. Each lane
 * shows the files compiling there now (with a running clock) and the
 * ones it just finished.
 */
export function BuildDispatch({ view, now, onSelectTask }: BuildDispatchProps) {
  const t = useT()
  const live = view.running > 0
  const height = Math.max(LANE_H, view.lanes.length * (LANE_H + LANE_GAP) - LANE_GAP)
  const names = laneNames(view.lanes)

  return (
    <div className="grid grid-cols-1 gap-3 p-4 md:grid-cols-[180px_96px_minmax(0,1fr)] md:gap-0">
      <SourceNode view={view} height={height} />

      <svg
        className="hidden md:block"
        width={FLOW_W}
        height={height}
        viewBox={`0 0 ${FLOW_W} ${height}`}
        aria-hidden
      >
        {view.lanes.map((lane, i) => {
          const y = i * (LANE_H + LANE_GAP) + LANE_H / 2
          const d = `M0 ${height / 2} C ${FLOW_W * 0.55} ${height / 2}, ${FLOW_W * 0.45} ${y}, ${FLOW_W} ${y}`
          const busy = lane.running.length
          return (
            <g key={lane.workerId}>
              <path d={d} fill="none" stroke="var(--color-hairline-strong)" strokeWidth="1.25" />
              {busy > 0 && (
                <path
                  d={d}
                  fill="none"
                  stroke="var(--color-status-running)"
                  strokeWidth={Math.min(1.5 + busy * 0.5, 4)}
                  strokeLinecap="round"
                  className="dispatch-flow"
                />
              )}
            </g>
          )
        })}
      </svg>

      <div className="flex min-w-0 flex-col" style={{ gap: LANE_GAP }}>
        {view.lanes.map((lane) => (
          <Lane key={lane.workerId} lane={lane} name={names.get(lane.workerId) ?? lane.workerId} now={now} onSelectTask={onSelectTask} />
        ))}
        {view.lanes.length === 0 && (
          <div className="flex h-[76px] items-center text-[12px] text-(--color-ink-muted)">{t('overview.noWorkers')}</div>
        )}
      </div>

      <span className="sr-only">{live ? t('buildDetail.dispatchLive') : t('buildDetail.dispatchDone')}</span>
    </div>
  )
}

function SourceNode({ view, height }: { view: DispatchView; height: number }) {
  const t = useT()
  const ratio = view.total > 0 ? view.settled / view.total : 0
  return (
    <div className="flex md:items-center" style={{ minHeight: Math.min(height, 160) }}>
      <div className="w-full rounded-(--radius-lg) border border-(--color-hairline) bg-(--color-surface) p-3 md:self-center">
        <div className="flex items-center gap-2 text-[12px] font-medium text-(--color-ink-muted)">
          <GitBranch className="h-3.5 w-3.5" />
          {t('buildDetail.dispatchSource')}
        </div>
        <div
          className="mt-2 font-tabular text-[22px] font-semibold leading-none"
          style={{ color: view.running > 0 ? 'var(--color-status-running)' : 'var(--color-ink)' }}
        >
          {view.running}
        </div>
        <div className="mt-1 text-[12px] text-(--color-ink-muted)">
          {t('buildDetail.dispatchRunningLabel')}
        </div>
        <div className="mt-3 h-1 overflow-hidden rounded-full bg-(--color-surface-inset)">
          <div
            className="h-full rounded-full bg-(--color-status-success) transition-[width] duration-500 ease-out"
            style={{ width: `${Math.round(ratio * 100)}%` }}
          />
        </div>
        <div className="mt-1.5 font-tabular text-[11px] text-(--color-ink-faint)">
          {t('buildDetail.dispatchProgress', { done: view.settled, total: view.total })}
        </div>
        {view.unassigned > 0 && (
          <div className="mt-1 text-[11px] text-(--color-status-timeout)">
            {t('buildDetail.dispatchUnassigned', { n: view.unassigned })}
          </div>
        )}
      </div>
    </div>
  )
}

function Lane({
  lane,
  name,
  now,
  onSelectTask,
}: {
  lane: DispatchLane
  name: string
  now: number
  onSelectTask?: (task: TaskInfo) => void
}) {
  const t = useT()
  const worker = lane.worker
  const busy = lane.running.length > 0
  const idle = lane.total === 0
  const shown = lane.running.slice(0, MAX_CHIPS)
  const hidden = lane.running.length - shown.length

  return (
    <div
      className={cn(
        'grid min-w-0 items-center gap-3 rounded-(--radius-lg) border px-3 transition-colors sm:grid-cols-[minmax(0,180px)_minmax(0,1fr)_auto]',
        busy
          ? 'border-[color-mix(in_oklch,var(--color-status-running)_35%,transparent)] bg-[color-mix(in_oklch,var(--color-status-running)_4%,var(--color-surface))]'
          : 'border-(--color-hairline) bg-(--color-surface)',
        idle && 'opacity-55',
      )}
      style={{ height: LANE_H }}
    >
      <div className="flex min-w-0 items-center gap-2.5">
        <MachineIllustration
          os={worker?.os ?? ''}
          slots={worker?.max_parallel_tasks ?? lane.running.length}
          busy={worker?.active_tasks ?? lane.running.length}
          power={worker ? powerTone(worker) : 'failed'}
          caption=""
          label={name}
          className="w-14 shrink-0"
        />
        <div className="min-w-0">
          <Link
            to="/workers/$workerId"
            params={{ workerId: lane.workerId }}
            title={lane.workerId}
            className="block truncate font-mono text-[12px] font-medium text-(--color-ink) hover:underline"
          >
            {name}
          </Link>
          <div className="truncate font-tabular text-[11px] text-(--color-ink-muted)">
            {worker
              ? t('buildDetail.dispatchSlots', { active: worker.active_tasks, max: worker.max_parallel_tasks })
              : t('buildDetail.dispatchOffline')}
          </div>
        </div>
      </div>

      <div className="hidden min-w-0 items-center gap-1.5 overflow-hidden sm:flex">
        {shown.map((task) => (
          <TaskChip key={task.id} task={task} now={now} onSelect={onSelectTask} />
        ))}
        {hidden > 0 && (
          <span className="shrink-0 font-tabular text-[11px] text-(--color-status-running)">
            {t('buildDetail.dispatchMore', { n: hidden })}
          </span>
        )}
        {!busy && idle && <span className="truncate text-[12px] text-(--color-ink-faint)">{t('buildDetail.dispatchIdle')}</span>}
        {!busy && !idle && <RecentDots tasks={lane.recent} onSelect={onSelectTask} />}
      </div>

      <div className="hidden shrink-0 flex-col items-end gap-1 sm:flex">
        <span className="font-tabular text-[18px] font-semibold leading-none text-(--color-ink)">{lane.total}</span>
        <span className="font-tabular text-[11px] text-(--color-ink-muted)">
          {t('buildDetail.dispatchCounts', { ok: lane.succeeded, failed: lane.failed })}
        </span>
        {busy && lane.recent.length > 0 && <RecentDots tasks={lane.recent} onSelect={onSelectTask} compact />}
      </div>
    </div>
  )
}

function TaskChip({ task, now, onSelect }: { task: TaskInfo; now: number; onSelect?: (task: TaskInfo) => void }) {
  const elapsed = task.started_at_ms > 0 ? Math.max(0, now - task.started_at_ms) : undefined
  return (
    <button
      type="button"
      onClick={() => onSelect?.(task)}
      title={`${taskLabel(task)}\n${task.command || task.id}`}
      className="dispatch-enter flex shrink-0 items-center gap-1.5 rounded-full border border-[color-mix(in_oklch,var(--color-status-running)_40%,transparent)] bg-(--color-surface) px-2 py-1 font-mono text-[11px] text-(--color-ink) hover:border-(--color-status-running)"
    >
      <span className="h-1.5 w-1.5 shrink-0 rounded-full bg-(--color-status-running) animate-(--animate-heartbeat)" />
      <span className="truncate">{tailEllipsis(taskLabel(task), CHIP_CHARS)}</span>
      {elapsed !== undefined && (
        <span className="shrink-0 font-tabular text-(--color-status-running)">{formatDuration(elapsed)}</span>
      )}
    </button>
  )
}

/** The lane's latest finished files as small status dots (newest first). */
function RecentDots({ tasks, onSelect, compact = false }: { tasks: TaskInfo[]; onSelect?: (task: TaskInfo) => void; compact?: boolean }) {
  return (
    <span className={cn('flex min-w-0 items-center', compact ? 'gap-1' : 'gap-1.5')}>
      {tasks.map((task) => {
        const tone = normalizeStatus(task.status, task.exit_code)
        const failed = tone === 'failed' || tone === 'timeout'
        const color = failed ? 'var(--color-status-failed)' : 'var(--color-status-success)'
        return compact ? (
          <button
            key={task.id}
            type="button"
            onClick={() => onSelect?.(task)}
            title={taskLabel(task)}
            aria-label={taskLabel(task)}
            className="h-1.5 w-1.5 rounded-full"
            style={{ backgroundColor: color }}
          />
        ) : (
          <button
            key={task.id}
            type="button"
            onClick={() => onSelect?.(task)}
            title={`${taskLabel(task)}\n${task.command || task.id}`}
            className="flex shrink-0 items-center gap-1 rounded-full bg-(--color-surface-inset) px-2 py-0.5 font-mono text-[11px] text-(--color-ink-muted) hover:text-(--color-ink)"
          >
            <span className="h-1.5 w-1.5 shrink-0 rounded-full" style={{ backgroundColor: color }} />
            <span className="truncate">{tailEllipsis(taskLabel(task), CHIP_CHARS)}</span>
          </button>
        )
      })}
    </span>
  )
}
