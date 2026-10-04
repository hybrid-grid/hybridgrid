import { useT } from '@/lib/i18n'
import { formatDuration } from '@/lib/utils'
import type { BuildInfo } from '@/lib/types'

interface BuildDurationsChartProps {
  builds: BuildInfo[]
}

/** Hand-rolled bar chart — no chart library, so bars stay themeable and tiny. */
export function BuildDurationsChart({ builds }: BuildDurationsChartProps) {
  const t = useT()
  const recent = builds.slice(0, 12).reverse()

  if (recent.length === 0) {
    return <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('builds.durationsEmpty')}</div>
  }

  const durations = recent.map((build) => Math.max(0, build.last_task_at_ms - build.first_task_at_ms))
  const max = Math.max(1, ...durations)

  return (
    <div className="flex h-40 items-end gap-2 p-4">
      {recent.map((build, i) => {
        const duration = durations[i]
        const heightPct = Math.max(4, (duration / max) * 100)
        const color = build.failed_tasks > 0 ? 'var(--color-status-failed)' : 'var(--color-status-success)'
        return (
          <div key={build.id} className="flex h-full flex-1 flex-col items-center justify-end gap-1">
            <div
              className="w-full min-w-[6px] rounded-t-(--radius-xs) transition-[filter] hover:brightness-125"
              style={{ height: `${heightPct}%`, backgroundColor: color }}
              title={`${build.id} · ${formatDuration(duration)}`}
            />
            <span className="w-full truncate text-center font-mono text-[9px] text-(--color-ink-faint)">
              {build.id.slice(-6)}
            </span>
          </div>
        )
      })}
    </div>
  )
}
