import { useT } from '@/lib/i18n'

interface OutcomeDonutProps {
  completed: number
  failed: number
  running: number
  total: number
}

const SIZE = 96
const STROKE = 10
const RADIUS = (SIZE - STROKE) / 2
const CIRCUMFERENCE = 2 * Math.PI * RADIUS

interface Segment {
  value: number
  color: string
}

export function OutcomeDonut({ completed, failed, running, total }: OutcomeDonutProps) {
  const t = useT()
  const queued = Math.max(0, total - completed - failed - running)
  const segments: Segment[] = [
    { value: completed, color: 'var(--color-status-success)' },
    { value: failed, color: 'var(--color-status-failed)' },
    { value: running, color: 'var(--color-status-running)' },
    { value: queued, color: 'var(--color-status-queued)' },
  ].filter((segment) => segment.value > 0)

  let offset = 0

  return (
    <div className="flex items-center gap-3">
      <svg width={SIZE} height={SIZE} viewBox={`0 0 ${SIZE} ${SIZE}`} className="-rotate-90">
        <circle
          cx={SIZE / 2}
          cy={SIZE / 2}
          r={RADIUS}
          fill="none"
          stroke="var(--color-hairline)"
          strokeWidth={STROKE}
        />
        {segments.map((segment, i) => {
          const length = total > 0 ? (segment.value / total) * CIRCUMFERENCE : 0
          const dashOffset = -offset
          offset += length
          return (
            <circle
              key={i}
              cx={SIZE / 2}
              cy={SIZE / 2}
              r={RADIUS}
              fill="none"
              stroke={segment.color}
              strokeWidth={STROKE}
              strokeDasharray={`${length} ${CIRCUMFERENCE - length}`}
              strokeDashoffset={dashOffset}
              strokeLinecap="butt"
            />
          )
        })}
      </svg>
      <div className="flex flex-col gap-1 text-[11px]">
        <LegendRow color="var(--color-status-success)" label={t('status.STATUS_COMPLETED')} value={completed} />
        <LegendRow color="var(--color-status-failed)" label={t('status.STATUS_FAILED')} value={failed} />
        <LegendRow color="var(--color-status-running)" label={t('status.STATUS_RUNNING')} value={running} />
        <LegendRow color="var(--color-status-queued)" label={t('status.STATUS_QUEUED')} value={queued} />
      </div>
    </div>
  )
}

function LegendRow({ color, label, value }: { color: string; label: string; value: number }) {
  return (
    <div className="flex items-center gap-1.5 text-(--color-ink-muted)">
      <span className="h-1.5 w-1.5 rounded-full" style={{ backgroundColor: color }} />
      {label} <span className="font-tabular text-(--color-ink-secondary)">{value}</span>
    </div>
  )
}
