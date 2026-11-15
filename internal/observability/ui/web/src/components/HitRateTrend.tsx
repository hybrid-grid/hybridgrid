import { useT } from '@/lib/i18n'
import { cn, formatPercent } from '@/lib/utils'
import type { StatsSample } from '@/store/useRealtimeStore'

const TREND_W = 300
const TREND_H = 56
const MIN_TREND_RANGE = 0.1

/**
 * Hit rate sampled from the 2s stats broadcast. The REST API keeps no
 * history, so the trace starts empty on load and fills while the page
 * is open — the label says "live" for that reason.
 */
export function HitRateTrend({ samples, tall = false }: { samples: StatsSample[]; tall?: boolean }) {
  const t = useT()

  if (samples.length < 2) {
    return (
      <div
        className={cn(
          'flex items-center justify-center rounded-(--radius-sm) border border-dashed border-(--color-hairline) text-[11px] text-(--color-ink-faint)',
          tall ? 'h-40' : 'h-14',
        )}
      >
        {t('overview.cacheTrendCollecting')}
      </div>
    )
  }

  // Auto-scaled domain: a session hovering around 75% would draw a flat
  // line on a fixed 0-100% axis. Keep at least a 10-point window so
  // noise is not exaggerated into drama.
  const rates = samples.map((s) => s.cacheHitRate)
  const mid = (Math.max(...rates) + Math.min(...rates)) / 2
  const half = Math.max(MIN_TREND_RANGE, Math.max(...rates) - Math.min(...rates)) / 2
  const lo = Math.max(0, Math.min(1 - half * 2, mid - half))
  const hi = Math.min(1, lo + half * 2)
  const step = TREND_W / (samples.length - 1)
  const points = rates.map((rate, i) => [i * step, TREND_H - ((rate - lo) / (hi - lo)) * (TREND_H - 4) - 2])
  const line = points.map(([x, y], i) => `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ')
  const area = `${line} L${TREND_W},${TREND_H} L0,${TREND_H} Z`

  return (
    <div className="relative">
      <span className="absolute right-0 top-0 font-tabular text-[9px] text-(--color-ink-faint)">{formatPercent(hi)}</span>
      <span className="absolute bottom-0 right-0 font-tabular text-[9px] text-(--color-ink-faint)">{formatPercent(lo)}</span>
      <svg viewBox={`0 0 ${TREND_W} ${TREND_H}`} preserveAspectRatio="none" className={cn('w-full pr-8', tall ? 'h-40' : 'h-14')} aria-hidden>
        <defs>
          <linearGradient id="hg-cache-fill" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor="var(--color-signal)" stopOpacity={0.22} />
            <stop offset="100%" stopColor="var(--color-signal)" stopOpacity={0} />
          </linearGradient>
        </defs>
        <line x1={0} x2={TREND_W} y1={TREND_H / 2} y2={TREND_H / 2} stroke="var(--color-hairline)" strokeDasharray="2 4" />
        <path d={area} fill="url(#hg-cache-fill)" />
        <path d={line} fill="none" stroke="var(--color-signal)" strokeWidth={1.5} vectorEffect="non-scaling-stroke" />
      </svg>
    </div>
  )
}
