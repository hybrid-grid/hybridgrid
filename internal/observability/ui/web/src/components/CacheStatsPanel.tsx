import { Gamepad2, Smartphone } from 'lucide-react'
import type { ComponentType } from 'react'
import { useT } from '@/lib/i18n'
import type { Stats } from '@/lib/types'
import { formatPercent } from '@/lib/utils'
import { useRealtimeStore } from '@/store/useRealtimeStore'
import { HitRateTrend } from './HitRateTrend'

interface CacheStatsPanelProps {
  stats: Stats
}

/** All *_hit_rate fields on the wire are 0..1 ratios. */
export function CacheStatsPanel({ stats }: CacheStatsPanelProps) {
  const t = useT()
  const history = useRealtimeStore((state) => state.statsHistory)

  return (
    <div className="flex h-full flex-col gap-4 p-4">
      <div className="flex items-end justify-between gap-4">
        <div>
          <div className="font-tabular text-[28px] font-semibold leading-none text-(--color-signal)">
            {formatPercent(stats.cache_hit_rate, 1)}
          </div>
          <div className="mt-1.5 font-tabular text-[11px] text-(--color-ink-muted)">
            {t('overview.hintHitsMisses', { hits: stats.cache_hits, misses: stats.cache_misses })}
          </div>
        </div>
        <div className="text-right text-[10px] uppercase tracking-[0.08em] text-(--color-ink-faint)">
          {t('overview.cacheTrendLabel')}
        </div>
      </div>

      <HitRateTrend samples={history} />

      <div className="flex flex-col gap-2.5 border-t border-(--color-hairline) pt-3">
        <CacheRow
          icon={Smartphone}
          label={t('overview.cacheFlutterLabel')}
          rate={stats.flutter_cache_hit_rate}
          hits={stats.flutter_cache_hits}
          misses={stats.flutter_cache_misses}
        />
        <CacheRow
          icon={Gamepad2}
          label={t('overview.cacheUnityLabel')}
          rate={stats.unity_cache_hit_rate}
          hits={stats.unity_cache_hits}
          misses={stats.unity_cache_misses}
        />
      </div>

      <p className="mt-auto text-[10px] leading-snug text-(--color-ink-faint)">{t('overview.cacheSourcesNote')}</p>
    </div>
  )
}

interface CacheRowProps {
  icon: ComponentType<{ className?: string }>
  label: string
  rate: number
  hits: number
  misses: number
}

function CacheRow({ icon: Icon, label, rate, hits, misses }: CacheRowProps) {
  const total = hits + misses
  return (
    <div className="flex items-center justify-between gap-3 text-[12px]">
      <span className="flex items-center gap-1.5 text-(--color-ink-secondary)">
        <Icon className="h-3.5 w-3.5 text-(--color-ink-faint)" />
        {label}
      </span>
      <div className="flex items-center gap-2">
        <div className="h-1 w-16 overflow-hidden rounded-full bg-(--color-surface-inset)">
          <div className="h-full bg-(--color-signal)" style={{ width: `${Math.min(100, rate * 100)}%` }} />
        </div>
        <span className="w-24 text-right font-tabular text-(--color-ink-muted)">
          {total > 0 ? formatPercent(rate) : '—'} <span className="text-(--color-ink-faint)">({hits}/{total})</span>
        </span>
      </div>
    </div>
  )
}
