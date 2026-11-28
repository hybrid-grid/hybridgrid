import { useT } from '@/lib/i18n'
import type { WorkerInfo } from '@/lib/types'

export function UnityCapabilityBadge({ worker }: { worker: WorkerInfo }) {
  const t = useT()
  if (!worker.unity_available) return null

  return (
    <span className="w-fit rounded-(--radius-xs) bg-(--color-surface-inset) px-1.5 py-0.5 font-mono text-[11px] text-(--color-ink-secondary)">
      Unity Editor{worker.unity_distributed_il2cpp && ` · ${t('workers.distributedIl2cpp')}`}
    </span>
  )
}
