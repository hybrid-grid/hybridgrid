import { useNavigate } from '@tanstack/react-router'
import { useT } from '@/lib/i18n'
import { buildDurationMs, buildTone, niceCeil } from '@/lib/metrics'
import { STATUS_COLOR } from '@/lib/status'
import type { BuildInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'

const MAX_BARS = 20
// Keeps a sub-second build visible (and hoverable) next to minute-long ones.
const MIN_BAR_PCT = 2

interface BuildDurationsChartProps {
  builds: BuildInfo[]
}

/** Hand-rolled bar chart — no chart library, so bars inherit theme tokens directly. */
export function BuildDurationsChart({ builds }: BuildDurationsChartProps) {
  const t = useT()
  const navigate = useNavigate()
  const recent = builds
    .filter((build) => buildDurationMs(build) !== null)
    .slice(0, MAX_BARS)
    .reverse()

  if (recent.length === 0) {
    return <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('builds.durationsEmpty')}</div>
  }

  const durations = recent.map((build) => buildDurationMs(build) ?? 0)
  const axisMax = niceCeil(Math.max(...durations) / 1000) * 1000

  return (
    <div className="flex flex-col gap-2 p-4">
      <div className="text-[11px] text-(--color-ink-faint)">{t('builds.durationsNote', { n: recent.length })}</div>
      <div className="grid grid-cols-[auto_minmax(0,1fr)] gap-x-2">
        <div className="flex h-40 flex-col justify-between text-right font-tabular text-[10px] text-(--color-ink-faint)">
          <span>{formatDuration(axisMax)}</span>
          <span>{formatDuration(axisMax / 2)}</span>
          <span>0</span>
        </div>
        <div className="relative h-40">
          <div className="pointer-events-none absolute inset-0 flex flex-col justify-between">
            <span className="h-px bg-(--color-hairline)" />
            <span className="h-px bg-(--color-hairline)" />
            <span className="h-px bg-(--color-hairline-strong)" />
          </div>
          <div className="absolute inset-0 flex items-end gap-1.5">
            {recent.map((build, i) => {
              const tone = buildTone(build)
              const heightPct = Math.max(MIN_BAR_PCT, (durations[i] / axisMax) * 100)
              return (
                <button
                  key={build.id}
                  type="button"
                  onClick={() => navigate({ to: '/builds/$buildId', params: { buildId: build.id } })}
                  aria-label={`${build.id} · ${formatDuration(durations[i])}`}
                  className="group relative flex h-full min-w-[6px] max-w-14 flex-1 cursor-pointer flex-col justify-end"
                >
                  <span
                    className={cn(
                      'w-full rounded-t-[3px] transition-[filter] group-hover:brightness-125',
                      tone === 'running' && 'animate-(--animate-heartbeat)',
                    )}
                    style={{ height: `${heightPct}%`, backgroundColor: STATUS_COLOR[tone] }}
                  />
                  <DurationTooltip build={build} durationMs={durations[i]} flip={i > recent.length / 2} />
                </button>
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}

function DurationTooltip({ build, durationMs, flip }: { build: BuildInfo; durationMs: number; flip: boolean }) {
  const t = useT()
  const tone = buildTone(build)
  return (
    <span
      role="tooltip"
      className={cn(
        'pointer-events-none invisible absolute bottom-full z-20 mb-2 w-52 rounded-(--radius-md) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) p-2.5 text-left opacity-0 shadow-[0_8px_24px_-8px_oklch(0_0_0/50%)] transition-opacity duration-150 group-hover:visible group-hover:opacity-100 group-focus-visible:visible group-focus-visible:opacity-100',
        flip ? 'right-0' : 'left-0',
      )}
    >
      <span className="block truncate font-mono text-[11px] font-medium text-(--color-ink)">{build.id}</span>
      <span className="mt-0.5 block font-mono text-[10px] text-(--color-ink-faint)">{formatClock(build.first_task_at_ms)}</span>
      <span className="mt-2 flex items-baseline justify-between">
        <span className="font-tabular text-[16px] font-semibold text-(--color-ink)">{formatDuration(durationMs)}</span>
        <span className="text-[11px]" style={{ color: STATUS_COLOR[tone] }}>
          {t(`status.${tone}`)}
        </span>
      </span>
      <span className="mt-1 block font-tabular text-[11px] text-(--color-ink-muted)">
        {build.completed_tasks + build.failed_tasks}/{build.total_tasks} · {build.from_cache_count} {t('common.cache')}
      </span>
    </span>
  )
}
