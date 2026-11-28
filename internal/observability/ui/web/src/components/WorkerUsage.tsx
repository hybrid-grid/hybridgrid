import { useNow } from '@/hooks/useNow'
import { useRelativeTime, useT } from '@/lib/i18n'
import type { UsageSample, WorkerInfo } from '@/lib/types'
import { USAGE_HISTORY, isFresh, memoryPercent, pressureTone, type PressureTone } from '@/lib/usage'
import { cn, formatSize } from '@/lib/utils'
import { useUsageStore } from '@/store/useUsageStore'

const EMPTY: UsageSample[] = []
/** The chart spans the whole retained history: 150 samples × 2 s = 5 min. */
const WINDOW_MS = USAGE_HISTORY * 2000
const CPU_COLOR = 'var(--color-series-1)'
const RAM_COLOR = 'var(--color-series-4)'
const TONE_COLOR: Record<PressureTone, string | undefined> = {
  ok: undefined,
  warn: 'var(--color-status-timeout)',
  hot: 'var(--color-status-failed)',
}

/** A worker's live history, falling back to the latest reading on the worker record. */
export function useWorkerUsage(worker: Pick<WorkerInfo, 'id'> & Partial<WorkerInfo>) {
  const samples = useUsageStore((state) => state.history[worker.id] ?? EMPTY)
  const latest: UsageSample | undefined =
    samples.at(-1) ??
    (worker.usage_at_ms
      ? {
          at_ms: worker.usage_at_ms,
          cpu_percent: worker.cpu_percent ?? 0,
          memory_used_bytes: worker.memory_used_bytes ?? 0,
          memory_total_bytes: worker.memory_total_bytes ?? 0,
          load1: worker.load1 ?? 0,
        }
      : undefined)
  return { samples, latest }
}

interface WorkerUsageProps {
  worker: WorkerInfo
  /** card: meters + small chart; compact: same without the caption; detail: large chart; row: one line + sparkline. */
  variant?: 'card' | 'compact' | 'detail' | 'row'
  className?: string
}

/**
 * Live CPU and memory of the worker's machine (whole machine, not just
 * the worker process), from 2-second reports. A reading older than
 * USAGE_STALE_MS is shown as stale rather than as current.
 */
export function WorkerUsage({ worker, variant = 'card', className }: WorkerUsageProps) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const now = useNow(2000)
  const { samples, latest } = useWorkerUsage(worker)
  const fresh = latest ? isFresh(latest.at_ms, now) : false

  if (!latest) {
    return <div className={cn('text-[11px] text-(--color-ink-faint)', className)}>{t('workers.usageNoData')}</div>
  }

  const cpu = latest.cpu_percent
  const ram = memoryPercent(latest)

  if (variant === 'row') {
    return (
      <div className={cn('flex items-center gap-3', !fresh && 'opacity-50', className)}>
        <UsageChart samples={samples} now={now} className="h-6 w-28 shrink-0" />
        <span className="whitespace-nowrap font-tabular text-[11px] text-(--color-ink-secondary)">
          {t('workers.usageShort', { cpu: Math.round(cpu), ram: Math.round(ram) })}
        </span>
      </div>
    )
  }

  return (
    <div className={cn('flex flex-col gap-2', className)}>
      <div className={cn('flex flex-col gap-1.5', !fresh && 'opacity-50')}>
        <Meter label={t('workers.usageCpu')} percent={cpu} color={CPU_COLOR} value={`${Math.round(cpu)}%`} />
        <Meter
          label={t('workers.usageRam')}
          percent={ram}
          color={RAM_COLOR}
          value={`${formatSize(latest.memory_used_bytes)} / ${formatSize(latest.memory_total_bytes)}`}
        />
      </div>
      <UsageChart
        samples={samples}
        now={now}
        className={variant === 'detail' ? 'h-28' : variant === 'compact' ? 'h-6' : 'h-9'}
        grid={variant === 'detail'}
      />
      {variant !== 'compact' && (
        <div className="flex items-center justify-between gap-2 text-[10px] text-(--color-ink-faint)">
          <span>
            {fresh ? t('workers.usageChartCaption') : t('workers.usageStale', { t: relativeTime(latest.at_ms / 1000) })}
          </span>
          {latest.load1 > 0 && <span className="font-tabular">{t('workers.usageLoad', { v: latest.load1.toFixed(2) })}</span>}
        </div>
      )}
    </div>
  )
}

function Meter({ label, percent, color, value }: { label: string; percent: number; color: string; value: string }) {
  const tone = pressureTone(percent)
  return (
    <div className="grid grid-cols-[32px_minmax(0,1fr)_auto] items-center gap-2 text-[11px]">
      <span className="font-medium text-(--color-ink-muted)">{label}</span>
      <span className="h-1.5 overflow-hidden rounded-full bg-(--color-surface-inset)">
        <span
          className="block h-full rounded-full transition-[width] duration-700 ease-out"
          style={{ width: `${Math.max(1, Math.min(100, percent))}%`, backgroundColor: TONE_COLOR[tone] ?? color }}
        />
      </span>
      <span className="font-tabular text-(--color-ink-secondary)" style={{ color: TONE_COLOR[tone] }}>
        {value}
      </span>
    </div>
  )
}

/** CPU (blue) and RAM (violet) percent over the last five minutes. */
export function UsageChart({ samples, now, className, grid = false }: { samples: UsageSample[]; now: number; className?: string; grid?: boolean }) {
  const start = now - WINDOW_MS
  const x = (at: number) => Math.max(0, ((at - start) / WINDOW_MS) * 100)
  const line = (value: (s: UsageSample) => number) =>
    samples.filter((s) => s.at_ms >= start).map((s, i) => `${i === 0 ? 'M' : 'L'}${x(s.at_ms).toFixed(2)} ${(100 - value(s)).toFixed(2)}`).join(' ')
  const cpu = line((s) => s.cpu_percent)
  const ram = line(memoryPercent)
  const lastX = samples.length ? x(samples[samples.length - 1].at_ms) : 0
  const firstX = samples.length ? x(Math.max(start, samples[0].at_ms)) : 0

  return (
    <svg viewBox="0 0 100 100" preserveAspectRatio="none" className={cn('w-full overflow-visible', className)} aria-hidden>
      {grid && (
        <>
          <line x1="0" x2="100" y1="50" y2="50" stroke="var(--color-hairline)" vectorEffect="non-scaling-stroke" strokeDasharray="3 3" />
          <line x1="0" x2="100" y1="0" y2="0" stroke="var(--color-hairline)" vectorEffect="non-scaling-stroke" strokeDasharray="3 3" />
        </>
      )}
      <line x1="0" x2="100" y1="100" y2="100" stroke="var(--color-hairline)" vectorEffect="non-scaling-stroke" />
      {cpu && <path d={`${cpu} L${lastX} 100 L${firstX} 100 Z`} fill={CPU_COLOR} fillOpacity="0.12" />}
      {cpu && <path d={cpu} fill="none" stroke={CPU_COLOR} strokeWidth="1.5" vectorEffect="non-scaling-stroke" strokeLinejoin="round" />}
      {ram && <path d={ram} fill="none" stroke={RAM_COLOR} strokeWidth="1.5" vectorEffect="non-scaling-stroke" strokeLinejoin="round" />}
    </svg>
  )
}
