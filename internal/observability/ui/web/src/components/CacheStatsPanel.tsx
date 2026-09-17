import { Database, Gamepad2, Smartphone } from 'lucide-react'
import type { ComponentType } from 'react'
import { useT } from '@/lib/i18n'
import type { Stats } from '@/lib/types'

interface CacheStatsPanelProps {
  stats: Stats
}

export function CacheStatsPanel({ stats }: CacheStatsPanelProps) {
  const t = useT()

  return (
    <div className="flex flex-col gap-4 p-4">
      <div>
        <div className="flex items-center gap-1.5 text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">
          <Database className="h-3 w-3" />
          {t('overview.cacheOverallLabel')}
        </div>
        <div className="font-tabular text-[26px] font-semibold text-(--color-signal)">
          {Math.round(stats.cache_hit_rate * 100)}%
        </div>
        <div className="text-[11px] text-(--color-ink-muted)">
          {t('overview.hintHitsMisses', { hits: stats.cache_hits, misses: stats.cache_misses })}
        </div>
      </div>

      <div className="h-px bg-(--color-hairline)" />

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
  return (
    <div className="flex items-center justify-between text-[12px]">
      <span className="flex items-center gap-1.5 text-(--color-ink-secondary)">
        <Icon className="h-3.5 w-3.5 text-(--color-ink-faint)" />
        {label}
      </span>
      <div className="flex items-center gap-2">
        <div className="h-1.5 w-20 overflow-hidden rounded-full bg-(--color-surface-inset)">
          <div className="h-full bg-(--color-signal)" style={{ width: `${Math.min(100, rate * 100)}%` }} />
        </div>
        <span className="w-20 text-right font-tabular text-(--color-ink-muted)">
          {rate.toFixed(0)}% ({hits}/{hits + misses})
        </span>
      </div>
    </div>
  )
}
