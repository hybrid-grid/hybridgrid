import { useT } from '@/lib/i18n'
import { useRealtimeStore } from '@/store/useRealtimeStore'
import { cn, formatClock } from '@/lib/utils'

const LABEL_KEY: Record<string, string> = {
  connecting: 'common.connecting',
  open: 'common.live',
  closed: 'common.reconnecting',
}

const COLOR: Record<string, string> = {
  connecting: 'var(--color-status-running)',
  open: 'var(--color-status-success)',
  closed: 'var(--color-status-failed)',
}

export function ConnectionStatus() {
  const status = useRealtimeStore((state) => state.status)
  const lastMessageAt = useRealtimeStore((state) => state.lastMessageAt)
  const t = useT()
  const color = COLOR[status]

  return (
    <div className="flex items-center gap-3">
      <span className="hidden font-mono text-[10px] text-(--color-ink-faint) sm:inline">
        {t('common.updated', { t: lastMessageAt ? formatClock(lastMessageAt) : t('common.never') })}
      </span>
      <div
        className="flex items-center gap-1.5 whitespace-nowrap rounded-full border border-(--color-hairline) px-2 py-1 text-[11px] font-medium text-(--color-ink-muted)"
        title={t(LABEL_KEY[status])}
      >
        <span
          className={cn('h-1.5 w-1.5 rounded-full', status === 'open' && 'animate-(--animate-heartbeat)')}
          style={{ backgroundColor: color }}
        />
        <span className="hidden sm:inline">{t(LABEL_KEY[status])}</span>
      </div>
    </div>
  )
}
