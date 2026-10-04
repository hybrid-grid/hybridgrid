import { createFileRoute, Link } from '@tanstack/react-router'
import { Panel } from '@/components/Panel'
import { PageHeader } from '@/components/PageHeader'
import { StatusBadge } from '@/components/StatusBadge'
import { BuildDurationsChart } from '@/components/BuildDurationsChart'
import { useT, useRelativeTime } from '@/lib/i18n'
import { useBuilds } from '@/lib/queries'

export const Route = createFileRoute('/builds/')({
  component: BuildsList,
})

function BuildsList() {
  const t = useT()
  const relativeTime = useRelativeTime()
  const { data, isLoading } = useBuilds()
  const builds = data?.builds ?? []

  return (
    <div className="flex flex-col gap-5">
      <PageHeader eyebrow={t('builds.eyebrow')} title={t('builds.title')} />

      <Panel eyebrow={t('builds.durationsEyebrow')} title={t('builds.durationsTitle')} padded={false}>
        <BuildDurationsChart builds={builds} />
      </Panel>

      <Panel padded={false}>
        <table className="w-full border-collapse text-left text-[12px]">
          <thead>
            <tr className="border-b border-(--color-hairline) text-[10px] uppercase tracking-[0.06em] text-(--color-ink-faint)">
              <th className="px-4 py-2 font-medium">{t('builds.colBuild')}</th>
              <th className="px-4 py-2 font-medium">{t('builds.colType')}</th>
              <th className="px-4 py-2 font-medium">{t('builds.colStatus')}</th>
              <th className="px-4 py-2 font-medium">{t('builds.colTasks')}</th>
              <th className="px-4 py-2 font-medium">{t('builds.colCache')}</th>
              <th className="px-4 py-2 font-medium">{t('builds.colLast')}</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-(--color-hairline)">
            {isLoading && (
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-(--color-ink-muted)">
                  {t('builds.loading')}
                </td>
              </tr>
            )}
            {!isLoading && builds.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-6 text-center text-(--color-ink-muted)">
                  {t('builds.empty')}
                </td>
              </tr>
            )}
            {builds.map((build) => (
              <tr key={build.id} className="hover:bg-(--color-canvas-raised)">
                <td className="px-4 py-2.5">
                  <Link
                    to="/builds/$buildId"
                    params={{ buildId: build.id }}
                    className="font-mono text-(--color-signal) hover:underline"
                  >
                    {build.id}
                  </Link>
                </td>
                <td className="px-4 py-2.5 text-(--color-ink-secondary)">{build.build_type}</td>
                <td className="px-4 py-2.5">
                  <StatusBadge status={build.status} />
                </td>
                <td className="px-4 py-2.5 font-tabular text-(--color-ink-secondary)">
                  {build.completed_tasks}/{build.total_tasks}
                  {build.failed_tasks > 0 && (
                    <span className="ml-1 text-(--color-status-failed)">
                      {t('builds.failedSuffix', { n: build.failed_tasks })}
                    </span>
                  )}
                </td>
                <td className="px-4 py-2.5 font-tabular text-(--color-ink-secondary)">
                  {build.from_cache_count}
                </td>
                <td className="px-4 py-2.5 text-(--color-ink-muted)">
                  {relativeTime(Math.round(build.last_task_at_ms / 1000))}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </Panel>
    </div>
  )
}
