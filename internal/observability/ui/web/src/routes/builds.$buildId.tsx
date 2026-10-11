import { createFileRoute, Link } from '@tanstack/react-router'
import { AlertTriangle, ExternalLink } from 'lucide-react'
import { useMemo } from 'react'
import { BuildTimeline } from '@/components/BuildTimeline'
import { ConsoleTabs } from '@/components/ConsoleTabs'
import { OutcomeDonut } from '@/components/OutcomeDonut'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatusBadge } from '@/components/StatusBadge'
import { TaskTable } from '@/components/TaskTable'
import { Button } from '@/components/ui/button'
import { useT } from '@/lib/i18n'
import { buildDurationMs, buildTone } from '@/lib/metrics'
import { useBuildDetail, useConsole } from '@/lib/queries'
import { normalizeStatus } from '@/lib/status'
import type { BuildInfo, TaskInfo } from '@/lib/types'
import { cn, formatClock, formatDuration } from '@/lib/utils'

interface BuildSearch {
  task?: string
}

export const Route = createFileRoute('/builds/$buildId')({
  validateSearch: (search: Record<string, unknown>): BuildSearch => ({
    task: typeof search.task === 'string' && search.task ? search.task : undefined,
  }),
  component: BuildDetail,
})

function BuildDetail() {
  const { buildId } = Route.useParams()
  const { task: selectedTaskId } = Route.useSearch()
  const navigate = Route.useNavigate()
  const t = useT()
  const { data, isLoading } = useBuildDetail(buildId)

  const tasks = useMemo(() => data?.tasks ?? [], [data])
  const failedFirst = useMemo(() => tasks.find((task) => isFailure(task)), [tasks])
  const selectedTask = tasks.find((task) => task.id === selectedTaskId) ?? failedFirst ?? tasks[0]
  const build = data?.build

  function selectTask(task: TaskInfo): void {
    navigate({ search: { task: task.id }, replace: true, resetScroll: false })
  }

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow={
          <Link to="/builds" className="hover:text-(--color-ink-secondary)">
            ← {t('buildDetail.backLabel')}
          </Link>
        }
        title={<span className="font-mono">{buildId}</span>}
        actions={build && <StatusBadge status={build.status} tone={buildTone(build)} />}
      />

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('buildDetail.loading')}</div>}

      {!isLoading && !data && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('buildDetail.notFound')}</div>
        </Panel>
      )}

      {data && build && (
        <>
          {build.truncated && (
            <div className="flex items-center gap-2 rounded-(--radius-md) border border-[color-mix(in_oklch,var(--color-status-timeout)_35%,transparent)] bg-[color-mix(in_oklch,var(--color-status-timeout)_10%,transparent)] px-3 py-2 text-[12px] text-(--color-status-timeout)">
              <AlertTriangle className="h-3.5 w-3.5 shrink-0" />
              {t('buildDetail.truncatedNote')}
            </div>
          )}

          <Panel eyebrow={t('buildDetail.metaTitle')} padded={false}>
            <div className="grid grid-cols-1 gap-6 p-4 lg:grid-cols-[minmax(0,1fr)_auto]">
              <BuildMeta build={build} tasks={tasks} />
              <OutcomeDonut
                completed={build.completed_tasks}
                failed={build.failed_tasks}
                running={build.running_tasks}
                total={build.total_tasks}
              />
            </div>
          </Panel>

          <Panel eyebrow={t('buildDetail.timelineTitle')} padded={false}>
            <BuildTimeline tasks={tasks} selectedTaskId={selectedTask?.id} onSelectTask={selectTask} />
          </Panel>

          <div className="grid grid-cols-1 gap-4 2xl:grid-cols-[minmax(0,1fr)_480px]">
            <Panel eyebrow={t('buildDetail.tasksTitle', { n: tasks.length })} padded={false} className="overflow-hidden">
              <div className="max-h-[560px] overflow-auto">
                <TaskTable tasks={tasks} selectedId={selectedTask?.id} onSelect={selectTask} />
              </div>
            </Panel>

            <Panel
              eyebrow={t('buildDetail.consoleTitle')}
              title={<span className="font-mono text-[12px]">{selectedTask ? selectedTask.id : t('buildDetail.consoleSelectTask')}</span>}
              action={
                selectedTask && (
                  <Button variant="ghost" size="sm" asChild>
                    <Link to="/tasks/$taskId/console" params={{ taskId: selectedTask.id }} search={{ build: buildId }}>
                      <ExternalLink className="h-3 w-3" />
                      {t('buildDetail.consoleOpenFull')}
                    </Link>
                  </Button>
                )
              }
              padded={false}
            >
              {selectedTask ? (
                <TaskConsole key={selectedTask.id} task={selectedTask} />
              ) : (
                <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('buildDetail.consoleSelectTask')}</div>
              )}
            </Panel>
          </div>

          <p className="text-[11px] text-(--color-ink-faint)">{t('buildDetail.provenance')}</p>
        </>
      )}
    </div>
  )
}

function isFailure(task: TaskInfo): boolean {
  const tone = normalizeStatus(task.status, task.exit_code)
  return tone === 'failed' || tone === 'timeout'
}

function BuildMeta({ build, tasks }: { build: BuildInfo; tasks: TaskInfo[] }) {
  const t = useT()
  const span = buildDurationMs(build)
  const queued = tasks.filter((task) => task.started_at_ms > 0)
  const avgQueue = queued.length > 0 ? queued.reduce((sum, task) => sum + task.queue_ms, 0) / queued.length : undefined
  const slowest = tasks.reduce<TaskInfo | undefined>(
    (max, task) => ((task.duration_ms ?? 0) > (max?.duration_ms ?? 0) ? task : max),
    undefined,
  )
  const workerCount = new Set(tasks.map((task) => task.worker_id).filter(Boolean)).size

  return (
    <dl className="grid grid-cols-2 gap-x-6 gap-y-4 text-[12px] sm:grid-cols-4">
      <MetaRow label={t('buildDetail.metaSpan')} value={span !== null ? formatDuration(span) : '—'} lead />
      <MetaRow label={t('buildDetail.metaTasks')} value={`${build.completed_tasks + build.failed_tasks}/${build.total_tasks}`} lead />
      <MetaRow
        label={t('buildDetail.metaFailed')}
        value={build.failed_tasks}
        lead
        color={build.failed_tasks > 0 ? 'var(--color-status-failed)' : undefined}
      />
      <MetaRow label={t('buildDetail.metaCache')} value={build.from_cache_count} lead />
      <MetaRow label={t('buildDetail.metaType')} value={build.build_type || 'native'} />
      <MetaRow label={t('buildDetail.metaFirstTask')} value={formatClock(build.first_task_at_ms)} mono />
      <MetaRow label={t('buildDetail.metaLastActivity')} value={formatClock(build.last_task_at_ms)} mono />
      <MetaRow label={t('buildDetail.metaWorkers')} value={workerCount} />
      <MetaRow label={t('buildDetail.metaAvgQueue')} value={avgQueue !== undefined ? formatDuration(avgQueue) : '—'} />
      <MetaRow
        label={t('buildDetail.metaSlowest')}
        value={slowest?.duration_ms ? `${formatDuration(slowest.duration_ms)}` : '—'}
        hint={slowest?.duration_ms ? slowest.id : undefined}
      />
    </dl>
  )
}

interface MetaRowProps {
  label: string
  value: string | number
  hint?: string
  lead?: boolean
  mono?: boolean
  color?: string
}

function MetaRow({ label, value, hint, lead, mono, color }: MetaRowProps) {
  return (
    <div className="min-w-0">
      <dt className="text-[10px] font-medium uppercase tracking-[0.08em] text-(--color-ink-faint)">{label}</dt>
      <dd
        className={cn(
          'mt-0.5 truncate font-tabular',
          lead ? 'text-[20px] font-semibold leading-tight text-(--color-ink)' : 'text-[13px] text-(--color-ink-secondary)',
          mono && 'font-mono text-[12px]',
        )}
        style={color ? { color } : undefined}
      >
        {value}
      </dd>
      {hint && (
        <div className="truncate font-mono text-[10px] text-(--color-ink-faint)" title={hint}>
          {hint}
        </div>
      )}
    </div>
  )
}

function TaskConsole({ task }: { task: TaskInfo }) {
  const t = useT()
  const { data, isLoading, isError } = useConsole(task.id)

  if (isLoading) return <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('console.loading')}</div>
  if (isError || !data) {
    return <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">{t('console.notAvailable')}</div>
  }
  return <ConsoleTabs output={data} heightClass="h-[400px]" />
}
