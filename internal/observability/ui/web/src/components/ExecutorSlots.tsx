import { useT } from '@/lib/i18n'
import { cn } from '@/lib/utils'

interface ExecutorSlotsProps {
  active: number
  max: number
}

export function ExecutorSlots({ active, max }: ExecutorSlotsProps) {
  const t = useT()
  const slots = Array.from({ length: Math.max(max, active, 1) }, (_, i) => i < active)
  const pct = max > 0 ? Math.min(100, (active / max) * 100) : 0

  return (
    <div className="flex flex-col gap-2">
      <div className="text-[10px] font-medium uppercase tracking-[0.06em] text-(--color-ink-faint)">
        {t('workers.slotsLabel')}
      </div>
      <div className="flex flex-wrap gap-1">
        {slots.map((busy, i) => (
          <span
            key={i}
            className={cn(
              'h-2.5 w-2.5 rounded-[3px] border',
              busy
                ? 'border-(--color-signal) bg-(--color-signal) animate-(--animate-heartbeat)'
                : 'border-(--color-hairline-strong) bg-transparent',
            )}
          />
        ))}
      </div>
      <div className="flex items-center gap-2">
        <div className="h-1.5 flex-1 overflow-hidden rounded-full bg-(--color-surface-inset)">
          <div className="h-full bg-(--color-signal)" style={{ width: `${pct}%` }} />
        </div>
        <span className="shrink-0 font-tabular text-[11px] text-(--color-ink-muted)">
          {t('workers.utilizationLabel', { active, max })}
        </span>
      </div>
    </div>
  )
}
