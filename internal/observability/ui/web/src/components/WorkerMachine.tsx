import { useT } from '@/lib/i18n'
import { archLabel, osLabel, powerTone } from '@/lib/machine'
import type { WorkerInfo } from '@/lib/types'
import { cn } from '@/lib/utils'
import { MachineIllustration } from './MachineIllustration'

/**
 * A worker's machine drawing on a quiet stage, with an "OS · arch" tag.
 * Shared by the Workers grid and the worker detail page.
 */
export function WorkerMachine({ worker, className }: { worker: WorkerInfo; className?: string }) {
  const t = useT()
  const os = osLabel(worker.os)
  const arch = archLabel(worker.architecture)
  const platform = arch ? `${os} · ${arch}` : os
  const caption = t('workers.utilizationLabel', { active: worker.active_tasks, max: worker.max_parallel_tasks })
  const health = worker.healthy ? t('workers.healthy') : t('workers.unhealthy')

  return (
    <div className={cn('relative bg-(--color-machine-stage) px-6 pb-2 pt-7', className)}>
      <span className="absolute left-3 top-3 rounded-full border border-(--color-hairline) bg-(--color-surface) px-2 py-0.5 font-mono text-[10px] text-(--color-ink-muted)">
        {platform}
      </span>
      <MachineIllustration
        os={worker.os}
        slots={worker.max_parallel_tasks}
        busy={worker.active_tasks}
        power={powerTone(worker)}
        caption={caption}
        label={`${worker.host || worker.id}: ${platform}, ${caption}, ${health}`}
        className="mx-auto max-w-[260px]"
      />
    </div>
  )
}
