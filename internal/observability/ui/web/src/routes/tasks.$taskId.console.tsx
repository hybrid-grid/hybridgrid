import { createFileRoute, Link, useRouter } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { ConsoleTabs } from '@/components/ConsoleTabs'
import { PageHeader } from '@/components/PageHeader'
import { Panel } from '@/components/Panel'
import { Button } from '@/components/ui/button'
import { useT } from '@/lib/i18n'
import { useConsole, useTasks } from '@/lib/queries'

interface ConsoleSearch {
  build?: string
}

// The console endpoint does not return the owning build, so fall back to
// the recent-task window when the link that got here did not carry it.
const TASK_LOOKUP_WINDOW = 200

export const Route = createFileRoute('/tasks/$taskId/console')({
  validateSearch: (search: Record<string, unknown>): ConsoleSearch => ({
    build: typeof search.build === 'string' && search.build ? search.build : undefined,
  }),
  component: TaskConsolePage,
})

function TaskConsolePage() {
  const { taskId } = Route.useParams()
  const router = useRouter()
  const t = useT()
  const { build: buildFromLink } = Route.useSearch()
  const { data, isLoading, isError } = useConsole(taskId)
  const { data: recent } = useTasks(TASK_LOOKUP_WINDOW)
  const buildId = buildFromLink ?? recent?.tasks.find((task) => task.id === taskId)?.build_id

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        eyebrow={<span className="font-mono normal-case tracking-normal">{taskId}</span>}
        title={t('console.title')}
        actions={
          buildId ? (
            <Button variant="outline" size="sm" asChild>
              <Link to="/builds/$buildId" params={{ buildId }} search={{ task: taskId }}>
                <ArrowLeft className="h-3.5 w-3.5" />
                {t('console.backToBuild')}
              </Link>
            </Button>
          ) : (
            <Button variant="outline" size="sm" onClick={() => router.history.back()}>
              <ArrowLeft className="h-3.5 w-3.5" />
              {t('console.back')}
            </Button>
          )
        }
      />

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('console.loading')}</div>}
      {isError && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('console.notAvailable')}</div>
        </Panel>
      )}

      {data && (
        <Panel padded={false} className="overflow-hidden">
          <ConsoleTabs output={data} heightClass="h-[64vh]" />
        </Panel>
      )}

      <p className="text-[11px] text-(--color-ink-faint)">{t('console.provenance')}</p>
    </div>
  )
}
