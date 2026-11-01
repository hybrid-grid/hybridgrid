import { createFileRoute } from '@tanstack/react-router'
import { BuildDurationsChart } from '@/components/BuildDurationsChart'
import { BuildsTable } from '@/components/BuildsTable'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { StatStrip, type StatCell } from '@/components/StatStrip'
import { useT } from '@/lib/i18n'
import { averageFinishedBuildMs, buildTone } from '@/lib/metrics'
import { useBuilds } from '@/lib/queries'
import type { BuildInfo } from '@/lib/types'
import { formatDuration, formatPercent } from '@/lib/utils'

export const Route = createFileRoute('/builds/')({
  component: BuildsList,
})

function BuildsList() {
  const t = useT()
  const { data, isLoading } = useBuilds()
  const builds = data?.builds ?? []

  return (
    <div className="flex flex-col gap-4">
      <PageHeader eyebrow={t('builds.eyebrow')} title={t('builds.title')} />

      {builds.length > 0 && <StatStrip cells={summaryCells(t, builds)} />}

      <Panel eyebrow={t('builds.durationsEyebrow')} title={t('builds.durationsTitle')} padded={false}>
        <BuildDurationsChart builds={builds} />
      </Panel>

      <Panel padded={false} className="overflow-hidden">
        <BuildsTable builds={builds} loading={isLoading} />
      </Panel>
    </div>
  )
}

function summaryCells(t: ReturnType<typeof useT>, builds: BuildInfo[]): StatCell[] {
  const running = builds.filter((build) => buildTone(build) === 'running').length
  const failed = builds.filter((build) => build.failed_tasks > 0).length
  const avg = averageFinishedBuildMs(builds)
  const tasks = builds.reduce((sum, build) => sum + build.total_tasks, 0)
  const cached = builds.reduce((sum, build) => sum + build.from_cache_count, 0)

  return [
    { label: t('builds.summaryTotal'), value: builds.length },
    {
      label: t('builds.summaryRunning'),
      value: running,
      color: running > 0 ? 'var(--color-status-running)' : undefined,
    },
    {
      label: t('builds.summaryFailed'),
      value: failed,
      color: failed > 0 ? 'var(--color-status-failed)' : undefined,
    },
    { label: t('builds.summaryAvg'), value: avg !== undefined ? formatDuration(avg) : '—' },
    {
      label: t('builds.summaryCacheShare'),
      value: tasks > 0 ? formatPercent(cached / tasks) : '—',
      meter: tasks > 0 ? cached / tasks : undefined,
      hint: `${cached}/${tasks}`,
    },
  ]
}
