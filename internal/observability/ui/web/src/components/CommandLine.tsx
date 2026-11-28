import { useT } from '@/lib/i18n'
import { cn } from '@/lib/utils'

/**
 * The compiler invocation the coordinator forwarded for a task, shown
 * above its console so the output has context. Renders nothing for
 * tasks without one (Flutter/Unity, older coordinators).
 */
export function CommandLine({ command, className }: { command?: string; className?: string }) {
  const t = useT()
  if (!command) return null
  return (
    <div className={cn('border-b border-(--color-hairline) px-4 py-2.5', className)}>
      <div className="text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">{t('console.command')}</div>
      <code className="mt-1 block break-all font-mono text-[11px] text-(--color-ink-secondary) select-all">{command}</code>
    </div>
  )
}
