import { createFileRoute, useRouter } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Panel } from '@/components/Panel'
import { PageHeader } from '@/components/PageHeader'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useT } from '@/lib/i18n'
import { useConsole } from '@/lib/queries'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/tasks/$taskId/console')({
  component: TaskConsolePage,
})

function TaskConsolePage() {
  const { taskId } = Route.useParams()
  const router = useRouter()
  const t = useT()
  const { data, isLoading, isError } = useConsole(taskId)

  return (
    <div className="flex flex-col gap-5">
      <PageHeader
        eyebrow={<span className="font-mono">{taskId}</span>}
        title={t('console.title')}
        actions={
          <Button variant="outline" size="sm" onClick={() => router.history.back()} className="gap-1.5">
            <ArrowLeft className="h-3.5 w-3.5" />
            {t('console.back')}
          </Button>
        }
      />

      {isLoading && <div className="text-[12px] text-(--color-ink-muted)">{t('console.loading')}</div>}
      {isError && (
        <Panel>
          <div className="text-[12px] text-(--color-ink-muted)">{t('console.notAvailable')}</div>
        </Panel>
      )}

      {data && (
        <Panel padded={false}>
          <Tabs defaultValue="stdout">
            <div className="flex items-center justify-between border-b border-(--color-hairline) px-4 py-2">
              <TabsList>
                <TabsTrigger value="stdout">{t('buildDetail.stdout')}</TabsTrigger>
                <TabsTrigger value="stderr">{t('buildDetail.stderr')}</TabsTrigger>
              </TabsList>
              {data.truncated && (
                <span className="text-[10px] uppercase tracking-wide text-(--color-status-timeout)">
                  {t('console.truncatedNote')}
                </span>
              )}
            </div>
            <TabsContent value="stdout">
              <ConsolePane text={data.stdout} />
            </TabsContent>
            <TabsContent value="stderr">
              <ConsolePane text={data.stderr} tone="failed" />
            </TabsContent>
          </Tabs>
        </Panel>
      )}

      <p className="text-[11px] text-(--color-ink-faint)">{t('console.provenance')}</p>
    </div>
  )
}

function ConsolePane({ text, tone }: { text?: string; tone?: 'failed' }) {
  const t = useT()
  return (
    <ScrollArea className="h-[60vh] bg-(--color-surface-inset)">
      <pre
        className={cn(
          'whitespace-pre-wrap break-words p-4 font-mono text-[11px] leading-relaxed',
          tone === 'failed' ? 'text-(--color-status-failed)' : 'text-(--color-ink-secondary)',
        )}
      >
        {text || t('common.empty')}
      </pre>
    </ScrollArea>
  )
}
