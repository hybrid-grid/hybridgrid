import { createFileRoute } from '@tanstack/react-router'
import { ArrowDown, ArrowUp } from 'lucide-react'
import { useCallback, useMemo } from 'react'
import { MenuSelect, Pagination, SearchField, Segmented } from '@/components/FilterControls'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { TaskTable } from '@/components/TaskTable'
import { Button } from '@/components/ui/button'
import { useT } from '@/lib/i18n'
import { taskWindowStats } from '@/lib/metrics'
import { TASK_HISTORY_LIMIT, useTaskHistory, useWorkers } from '@/lib/queries'
import { normalizeStatus } from '@/lib/status'
import {
  applyTaskQuery,
  paginate,
  type SortDir,
  type TaskCacheFilter,
  type TaskQuery,
  type TaskSortKey,
  type TaskStatusFilter,
} from '@/lib/taskQuery'
import type { TaskInfo } from '@/lib/types'
import { formatDuration } from '@/lib/utils'

const PAGE_SIZE = 50
const ALL_WORKERS = '__all__'

interface TasksSearch extends TaskQuery {
  page?: number
}

function pick<T extends string>(value: unknown, allowed: readonly T[]): T | undefined {
  return typeof value === 'string' && (allowed as readonly string[]).includes(value) ? (value as T) : undefined
}

export const Route = createFileRoute('/tasks/')({
  // Every filter lives in the URL: a filtered view is a link you can paste in chat.
  validateSearch: (search: Record<string, unknown>): TasksSearch => ({
    q: typeof search.q === 'string' && search.q ? search.q : undefined,
    status: pick(search.status, ['all', 'running', 'success', 'failed'] as const),
    worker: typeof search.worker === 'string' && search.worker ? search.worker : undefined,
    cache: pick(search.cache, ['all', 'hit', 'miss'] as const),
    sort: pick(search.sort, ['started', 'duration', 'queue', 'compile'] as const),
    dir: pick(search.dir, ['asc', 'desc'] as const),
    page: typeof search.page === 'number' && search.page > 1 ? Math.floor(search.page) : undefined,
  }),
  component: TasksPage,
})

function TasksPage() {
  const t = useT()
  const search = Route.useSearch()
  const navigate = Route.useNavigate()
  const { data, isLoading } = useTaskHistory()
  const { data: workersData } = useWorkers()
  const tasks = useMemo(() => data?.tasks ?? [], [data])

  const update = useCallback(
    (patch: Partial<TasksSearch>) =>
      navigate({ search: (prev) => ({ ...prev, page: undefined, ...patch }), replace: true, resetScroll: false }),
    [navigate],
  )
  const onSearch = useCallback((q: string) => update({ q: q || undefined }), [update])

  const filtered = useMemo(() => applyTaskQuery(tasks, search), [tasks, search])
  const { rows, page, pageCount } = paginate(filtered, search.page ?? 1, PAGE_SIZE)
  const workerOptions = useMemo(() => workerChoices(tasks, workersData?.workers.map((w) => w.id) ?? []), [tasks, workersData])
  const filtersActive = Boolean(search.q || search.worker || (search.status && search.status !== 'all') || (search.cache && search.cache !== 'all'))
  const sort = search.sort ?? 'started'
  const dir: SortDir = search.dir ?? 'desc'

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('tasks.eyebrow')} title={t('tasks.title')} />

      {tasks.length > 0 && <StatStrip cells={summaryCells(t, tasks)} />}

      <Panel padded={false} className="overflow-hidden">
        <div className="flex flex-col gap-3 border-b border-(--color-hairline) p-3 lg:flex-row lg:items-center">
          <SearchField
            value={search.q ?? ''}
            onChange={onSearch}
            placeholder={t('tasks.searchPlaceholder')}
            className="lg:w-80"
          />
          <div className="flex flex-wrap items-center gap-2">
            <Segmented<TaskStatusFilter>
              label={t('buildDetail.colStatus')}
              value={search.status ?? 'all'}
              onChange={(status) => update({ status: status === 'all' ? undefined : status })}
              options={[
                { value: 'all', label: t('tasks.statusAll') },
                { value: 'running', label: t('tasks.statusRunning') },
                { value: 'success', label: t('tasks.statusSuccess') },
                { value: 'failed', label: t('tasks.statusFailed') },
              ]}
            />
            <Segmented<TaskCacheFilter>
              label={t('buildDetail.colCache')}
              value={search.cache ?? 'all'}
              onChange={(cache) => update({ cache: cache === 'all' ? undefined : cache })}
              options={[
                { value: 'all', label: t('tasks.cacheAll') },
                { value: 'hit', label: t('tasks.cacheHit') },
                { value: 'miss', label: t('tasks.cacheMiss') },
              ]}
            />
            <MenuSelect
              label={t('tasks.workerLabel')}
              value={search.worker ?? ALL_WORKERS}
              onChange={(worker) => update({ worker: worker === ALL_WORKERS ? undefined : worker })}
              options={[{ value: ALL_WORKERS, label: t('tasks.workerAll') }, ...workerOptions.map((id) => ({ value: id, label: id }))]}
            />
          </div>
          <div className="flex items-center gap-2 lg:ml-auto">
            <MenuSelect<TaskSortKey>
              label="Sort"
              value={sort}
              onChange={(next) => update({ sort: next === 'started' ? undefined : next })}
              options={[
                { value: 'started', label: t('tasks.sortStarted') },
                { value: 'duration', label: t('tasks.sortDuration') },
                { value: 'queue', label: t('tasks.sortQueue') },
                { value: 'compile', label: t('tasks.sortCompile') },
              ]}
            />
            <Button
              variant="outline"
              size="icon"
              className="h-8 w-8"
              aria-label={dir === 'desc' ? 'Sort ascending' : 'Sort descending'}
              onClick={() => update({ dir: dir === 'desc' ? 'asc' : undefined })}
            >
              {dir === 'desc' ? <ArrowDown className="h-3.5 w-3.5" /> : <ArrowUp className="h-3.5 w-3.5" />}
            </Button>
          </div>
        </div>

        <div className="flex items-center justify-between gap-3 px-4 py-2 text-[11px] text-(--color-ink-faint)">
          <span className="font-tabular">{t('tasks.resultCount', { n: filtered.length, total: tasks.length })}</span>
          <span>{t('tasks.windowNote', { n: TASK_HISTORY_LIMIT })}</span>
        </div>

        {isLoading ? (
          <div className="px-4 py-10 text-center text-[12px] text-(--color-ink-muted)">{t('common.loading')}</div>
        ) : rows.length === 0 ? (
          <div className="flex flex-col items-center gap-3 px-4 py-10 text-[12px] text-(--color-ink-muted)">
            {t('tasks.empty')}
            {filtersActive && (
              <Button variant="outline" size="sm" onClick={() => navigate({ search: {}, replace: true })}>
                {t('tasks.reset')}
              </Button>
            )}
          </div>
        ) : (
          <div className="overflow-x-auto border-t border-(--color-hairline)">
            <TaskTable tasks={rows} showBuild />
          </div>
        )}

        <Pagination
          page={page}
          pageCount={pageCount}
          onPage={(next) => navigate({ search: (prev) => ({ ...prev, page: next > 1 ? next : undefined }), resetScroll: false })}
          label={t('tasks.page', { page, count: pageCount })}
          prevLabel={t('tasks.prev')}
          nextLabel={t('tasks.next')}
        />
      </Panel>
    </div>
  )
}

function workerChoices(tasks: TaskInfo[], registered: string[]): string[] {
  return [...new Set([...registered, ...tasks.map((task) => task.worker_id).filter(Boolean)])].sort()
}

function summaryCells(t: ReturnType<typeof useT>, tasks: TaskInfo[]): StatCell[] {
  const stats = taskWindowStats(tasks)
  const running = tasks.filter((task) => normalizeStatus(task.status) === 'running').length
  return [
    { label: t('tasks.summaryTasks'), value: stats.total },
    { label: t('tasks.summaryRunning'), value: running, color: running > 0 ? 'var(--color-status-running)' : undefined },
    { label: t('tasks.summaryFailed'), value: stats.failed, color: stats.failed > 0 ? 'var(--color-status-failed)' : undefined },
    { label: t('tasks.summaryAvgCompile'), value: stats.avgCompileMs !== undefined ? formatDuration(stats.avgCompileMs) : '—' },
    { label: t('tasks.summaryAvgQueue'), value: stats.avgQueueMs !== undefined ? formatDuration(stats.avgQueueMs) : '—' },
  ]
}
