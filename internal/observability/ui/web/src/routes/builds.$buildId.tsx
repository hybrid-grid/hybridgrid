import { createFileRoute, Link } from '@tanstack/react-router'
import { ExternalLink } from 'lucide-react'
import { useState } from 'react'
import { BuildTimeline } from '@/components/BuildTimeline'
import { OutcomeDonut } from '@/components/OutcomeDonut'
import { Panel } from '@/components/Panel'
import { PageHeader } from '@/components/PageHeader'
import { StatusBadge } from '@/components/StatusBadge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { useT } from '@/lib/i18n'
import { useBuildDetail, useConsole } from '@/lib/queries'
import { cn, formatDuration } from '@/lib/utils'
import type { TaskInfo } from '@/lib/types'

export const Route = createFileRoute('/builds/$buildId')({
  component: BuildDetail,
})

function BuildDetail() {
  const { buildId } = Route.useParams()
  const t = useT()
  const { data, isLoading } = useBuildDetail(buildId)
  const [selectedTaskId, setSelectedTaskId] = useState<string | undefined>(undefined)

  const tasks = data?.tasks ?? []
  const selectedTask = tasks.find((task) => task.id === selectedTaskId) ?? tasks[0]
  const build = data?.build

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        eyebrow={
          <Link to="/builds" className="hover:text-(--color-ink-secondary)">
            {t('buildDetail.backLabel')}
          </Link>
        }
        title={<span className="font-mono">{buildId}</span>}
        actions={build && <StatusBadge status={build.status} />}
      />

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('buildDetail.loading')}</div>}

      {!isLoading && !data && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('buildDetail.notFound')}</div>
        </Panel>
      )}

      {data && build && (
        <>
          <Panel eyebrow={t('buildDetail.metaTitle')} padded={false}>
            <div className="grid grid-cols-1 gap-4 p-4 sm:grid-cols-[minmax(0,1fr)_auto]">
              <dl className="grid grid-cols-2 gap-3 text-[12px] sm:grid-cols-4">
                <MetaRow label={t('buildDetail.metaType')} value={build.build_type || 'native'} />
                <MetaRow label={t('buildDetail.metaTasks')} value={`${build.completed_tasks + build.failed_tasks}/${build.total_tasks}`} />
                <MetaRow label={t('buildDetail.metaFailed')} value={build.failed_tasks} />
                <MetaRow label={t('buildDetail.metaCache')} value={build.from_cache_count} />
                <MetaRow label={t('buildDetail.metaSpan')} value={formatDuration(build.last_task_at_ms - build.first_task_at_ms)} />
              </dl>
              <OutcomeDonut
                completed={build.completed_tasks}
                failed={build.failed_tasks}
                running={build.running_tasks}
                total={build.total_tasks}
              />
            </div>
          </Panel>

          <Panel eyebrow={t('buildDetail.timelineTitle')} padded={false}>
            <BuildTimeline tasks={tasks} onSelectTask={setSelectedTaskId} />
          </Panel>

          <div className="grid grid-cols-1 gap-4 xl:grid-cols-[minmax(0,1fr)_420px]">
            <Panel eyebrow={t('buildDetail.tasksTitle', { n: tasks.length })} padded={false}>
              <ScrollArea className="max-h-[520px]">
                <table className="w-full border-collapse text-left text-[12px]">
                  <thead className="sticky top-0 bg-(--color-surface)">
                    <tr className="border-b border-(--color-hairline) text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">
                      <th className="px-4 py-2 font-medium">{t('buildDetail.colTask')}</th>
                      <th className="px-4 py-2 font-medium">{t('buildDetail.colWorker')}</th>
                      <th className="px-4 py-2 font-medium">{t('buildDetail.colStatus')}</th>
                      <th className="px-4 py-2 font-medium">{t('buildDetail.colDuration')}</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-(--color-hairline)">
                    {tasks.map((task) => (
                      <TaskRow
                        key={task.id}
                        task={task}
                        selected={task.id === selectedTask?.id}
                        onSelect={() => setSelectedTaskId(task.id)}
                      />
                    ))}
                  </tbody>
                </table>
              </ScrollArea>
            </Panel>

            <Panel
              eyebrow={t('buildDetail.consoleTitle')}
              title={selectedTask ? selectedTask.id : t('buildDetail.consoleSelectTask')}
              action={
                selectedTask && (
                  <Button variant="ghost" size="sm" asChild>
                    <Link to="/tasks/$taskId/console" params={{ taskId: selectedTask.id }} className="gap-1.5">
                      <ExternalLink className="h-3 w-3" />
                      {t('buildDetail.consoleOpenFull')}
                    </Link>
                  </Button>
                )
              }
              padded={false}
            >
              {selectedTask ? (
                <TaskConsole taskId={selectedTask.id} />
              ) : (
                <div className="p-6 text-center text-[12px] text-(--color-ink-muted)">
                  {t('buildDetail.consoleSelectTask')}
                </div>
              )}
            </Panel>
          </div>
        </>
      )}
    </div>
  )
}

function MetaRow({ label, value }: { label: string; value: string | number }) {
  return (
    <div>
      <dt className="text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">{label}</dt>
      <dd className="font-tabular text-(--color-ink-secondary)">{value}</dd>
    </div>
  )
}

interface TaskRowProps {
  task: TaskInfo
  selected: boolean
  onSelect: () => void
}

function TaskRow({ task, selected, onSelect }: TaskRowProps) {
  const t = useT()
  return (
    <tr
      onClick={onSelect}
      className={cn(
        'cursor-pointer hover:bg-(--color-canvas-raised)',
        selected && 'bg-(--color-canvas-raised)',
      )}
    >
      <td className="px-4 py-2.5 font-mono text-(--color-ink)">{task.id}</td>
      <td className="px-4 py-2.5 text-(--color-ink-secondary)">{task.worker_id || '—'}</td>
      <td className="px-4 py-2.5">
        <StatusBadge status={task.status} />
      </td>
      <td className="px-4 py-2.5 font-tabular text-(--color-ink-secondary)">
        {task.duration_ms ? formatDuration(task.duration_ms) : '—'}
        {task.from_cache && <span className="ml-1 text-(--color-signal)">({t('common.cache')})</span>}
      </td>
    </tr>
  )
}

function TaskConsole({ taskId }: { taskId: string }) {
  const t = useT()
  const { data, isLoading } = useConsole(taskId)

  return (
    <Tabs defaultValue="stdout" className="flex flex-col">
      <div className="border-b border-(--color-hairline) px-4 py-2">
        <TabsList>
          <TabsTrigger value="stdout">{t('buildDetail.stdout')}</TabsTrigger>
          <TabsTrigger value="stderr">{t('buildDetail.stderr')}</TabsTrigger>
        </TabsList>
      </div>
      <TabsContent value="stdout">
        <ConsolePane text={data?.stdout} loading={isLoading} />
      </TabsContent>
      <TabsContent value="stderr">
        <ConsolePane text={data?.stderr} loading={isLoading} tone="failed" />
      </TabsContent>
    </Tabs>
  )
}

function ConsolePane({ text, loading, tone }: { text?: string; loading: boolean; tone?: 'failed' }) {
  const t = useT()
  return (
    <ScrollArea className="h-[380px] bg-(--color-surface-inset)">
      <pre
        className={cn(
          'whitespace-pre-wrap break-words p-4 font-mono text-[11px] leading-relaxed',
          tone === 'failed' ? 'text-(--color-status-failed)' : 'text-(--color-ink-secondary)',
        )}
      >
        {loading ? t('console.loading') : text || t('common.empty')}
      </pre>
    </ScrollArea>
  )
}
