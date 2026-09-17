import { useNavigate } from '@tanstack/react-router'
import { useMemo } from 'react'
import { TimelineTrack } from '@/components/TimelineTrack'
import { useT } from '@/lib/i18n'
import { layoutTimeline } from '@/lib/timeline'
import type { TaskInfo } from '@/lib/types'

interface ClusterActivityProps {
  tasks: TaskInfo[]
}

export function ClusterActivity({ tasks }: ClusterActivityProps) {
  const t = useT()
  const navigate = useNavigate()
  const layout = useMemo(() => layoutTimeline(tasks), [tasks])

  const lanes = useMemo(() => {
    const byWorker = new Map<string, typeof layout.rows>()
    for (const row of layout.rows) {
      const key = row.task.worker_id || 'unassigned'
      const existing = byWorker.get(key)
      if (existing) existing.push(row)
      else byWorker.set(key, [row])
    }
    return Array.from(byWorker.entries()).sort((a, b) => a[0].localeCompare(b[0]))
  }, [layout.rows])

  const peak = useMemo(() => {
    return lanes.reduce((max, [, rows]) => Math.max(max, rows.length), 0)
  }, [lanes])

  if (lanes.length === 0) {
    return (
      <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('overview.noClusterActivity')}</div>
    )
  }

  return (
    <div className="flex flex-col gap-2 p-4">
      <div className="mb-1 text-[11px] text-(--color-ink-faint)">
        {t('overview.clusterActivityNote', { peak, n: layout.rows.length })}
      </div>
      {lanes.map(([workerId, rows]) => (
        <div key={workerId} className="flex items-center gap-3">
          <div className="w-40 shrink-0">
            <div className="truncate font-mono text-[11px] text-(--color-ink-secondary)" title={workerId}>
              {workerId}
            </div>
            <div className="text-[10px] text-(--color-ink-faint)">{t('overview.taskCount', { n: rows.length })}</div>
          </div>
          <div className="flex-1">
            <TimelineTrack
              rows={rows}
              onSelect={(taskId) => navigate({ to: '/tasks/$taskId/console', params: { taskId } })}
            />
          </div>
        </div>
      ))}
    </div>
  )
}
