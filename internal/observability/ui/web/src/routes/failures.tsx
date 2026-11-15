import { createFileRoute, Link } from '@tanstack/react-router'
import { Server, Terminal } from 'lucide-react'
import { useMemo } from 'react'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { Button } from '@/components/ui/button'
import { groupFailures, isWorkerSpecific, type FailureGroup } from '@/lib/failures'
import { useT, useRelativeTime } from '@/lib/i18n'
import { useTaskHistory, useWorkers } from '@/lib/queries'
import type { WorkerInfo } from '@/lib/types'

export const Route = createFileRoute('/failures')({
  component: FailuresPage,
})

const SHOWN_LOCATIONS = 4

function FailuresPage() {
  const t = useT()
  const { data, isLoading } = useTaskHistory()
  const { data: workersData } = useWorkers()
  const workers = useMemo(() => workersData?.workers ?? [], [workersData])
  const groups = useMemo(() => groupFailures(data?.tasks ?? []), [data])

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('failures.eyebrow')} title={t('failures.title')} />

      {groups.length > 0 && <StatStrip cells={summaryCells(t, groups, workers.length)} />}

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('common.loading')}</div>}

      {!isLoading && groups.length === 0 && (
        <Panel>
          <div className="py-8 text-center text-[12px] text-(--color-ink-muted)">{t('failures.empty')}</div>
        </Panel>
      )}

      <div className="flex flex-col gap-3">
        {groups.map((group) => (
          <FailureCard key={group.signature} group={group} workers={workers} />
        ))}
      </div>
    </div>
  )
}

function summaryCells(t: ReturnType<typeof useT>, groups: FailureGroup[], workerCount: number): StatCell[] {
  const failedTasks = groups.reduce((sum, group) => sum + group.count, 0)
  const builds = new Set(groups.flatMap((group) => group.buildIds)).size
  const workerSpecific = groups.filter((group) => isWorkerSpecific(group, workerCount)).length
  return [
    { label: t('failures.summaryGroups'), value: groups.length, color: 'var(--color-status-failed)', lead: true },
    { label: t('failures.summaryTasks'), value: failedTasks },
    { label: t('failures.summaryBuilds'), value: builds },
    {
      label: t('failures.summaryWorkerSpecific'),
      value: workerSpecific,
      color: workerSpecific > 0 ? 'var(--color-status-timeout)' : undefined,
    },
  ]
}

function FailureCard({ group, workers }: { group: FailureGroup; workers: WorkerInfo[] }) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const hostById = new Map(workers.map((w) => [w.id, w.host || w.id]))
  const workerSpecific = isWorkerSpecific(group, workers.length)
  const maxWorkerCount = group.workers[0]?.count ?? 1

  return (
    <Panel padded={false} className="overflow-hidden">
      <div className="flex flex-col gap-3 p-4 md:flex-row md:items-start">
        <div className="flex min-w-[88px] flex-row items-baseline gap-2 md:flex-col md:gap-0">
          <span className="font-tabular text-[28px] font-semibold leading-none text-(--color-status-failed)">{group.count}</span>
          <span className="text-[11px] text-(--color-ink-faint)">
            {t('failures.builds', { n: group.buildIds.length })} · {t('failures.lastSeen', { t: relativeTime(Math.round(group.lastAtMs / 1000)) })}
          </span>
        </div>

        <div className="min-w-0 flex-1">
          <p className="break-words font-mono text-[13px] leading-relaxed text-(--color-ink)">{group.signature}</p>

          {workerSpecific && (
            <div className="mt-2 inline-flex items-center gap-1.5 rounded-(--radius-xs) bg-[color-mix(in_oklch,var(--color-status-timeout)_12%,transparent)] px-2 py-1 text-[11px] text-(--color-status-timeout)">
              <Server className="h-3 w-3" />
              {t('failures.workerSpecific', { worker: hostById.get(group.workers[0].workerId) ?? group.workers[0].workerId })}
            </div>
          )}

          <div className="mt-3 grid grid-cols-1 gap-4 lg:grid-cols-2">
            {group.locations.length > 0 && (
              <div>
                <div className="mb-1 text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{t('failures.locations')}</div>
                <ul className="flex flex-col gap-0.5">
                  {group.locations.slice(0, SHOWN_LOCATIONS).map((loc) => (
                    <li key={`${loc.file}:${loc.line}`} className="truncate font-mono text-[11px] text-(--color-ink-secondary)" title={`${loc.file}:${loc.line}`}>
                      {loc.file}
                      <span className="text-(--color-ink-faint)">:{loc.line}</span>
                    </li>
                  ))}
                  {group.locations.length > SHOWN_LOCATIONS && (
                    <li className="text-[11px] text-(--color-ink-faint)">{t('failures.moreLocations', { n: group.locations.length - SHOWN_LOCATIONS })}</li>
                  )}
                </ul>
              </div>
            )}

            <div>
              <div className="mb-1 text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{t('failures.affectedWorkers')}</div>
              <ul className="flex flex-col gap-1">
                {group.workers.map(({ workerId, count }) => (
                  <li key={workerId || 'unassigned'} className="grid grid-cols-[minmax(0,140px)_minmax(0,1fr)_auto] items-center gap-2 text-[11px]">
                    {workerId ? (
                      <Link
                        to="/workers/$workerId"
                        params={{ workerId }}
                        className="truncate font-mono text-(--color-ink-secondary) hover:text-(--color-signal)"
                      >
                        {hostById.get(workerId) ?? workerId}
                      </Link>
                    ) : (
                      <span className="text-(--color-ink-faint)">{t('overview.eventUnassigned')}</span>
                    )}
                    <span className="h-1 overflow-hidden rounded-full bg-(--color-surface-inset)">
                      <span className="block h-full rounded-full bg-(--color-status-failed)" style={{ width: `${(count / maxWorkerCount) * 100}%` }} />
                    </span>
                    <span className="font-tabular text-(--color-ink-muted)">{count}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-between gap-2 border-t border-(--color-hairline) bg-(--color-canvas-raised) px-4 py-2">
        <span className="min-w-0 truncate text-[11px] text-(--color-ink-faint)">
          {t('failures.latestTask')}{' '}
          {group.sample.build_id ? (
            <Link
              to="/builds/$buildId"
              params={{ buildId: group.sample.build_id }}
              search={{ task: group.sample.id }}
              className="font-mono text-(--color-ink-secondary) hover:text-(--color-signal)"
            >
              {group.sample.id}
            </Link>
          ) : (
            <span className="font-mono text-(--color-ink-secondary)">{group.sample.id}</span>
          )}
        </span>
        <Button variant="ghost" size="sm" asChild>
          <Link
            to="/tasks/$taskId/console"
            params={{ taskId: group.sample.id }}
            search={{ build: group.sample.build_id || undefined }}
          >
            <Terminal className="h-3 w-3" />
            {t('failures.openConsole')}
          </Link>
        </Button>
      </div>
    </Panel>
  )
}
