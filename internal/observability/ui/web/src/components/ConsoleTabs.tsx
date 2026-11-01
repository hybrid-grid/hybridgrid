import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useT } from '@/lib/i18n'
import type { ConsoleOutput } from '@/lib/types'
import { cn } from '@/lib/utils'

interface ConsoleTabsProps {
  output: ConsoleOutput
  heightClass: string
}

function lineCount(text: string): number {
  return text ? text.split('\n').filter(Boolean).length : 0
}

/**
 * stdout/stderr viewer shared by the build-detail side panel and the
 * full console page. Opens on stderr whenever it has content — that is
 * where the compiler error the user came for lives.
 */
export function ConsoleTabs({ output, heightClass }: ConsoleTabsProps) {
  const t = useT()
  const stderrLines = lineCount(output.stderr)

  return (
    <Tabs defaultValue={stderrLines > 0 ? 'stderr' : 'stdout'} className="flex flex-col">
      <div className="flex items-center justify-between gap-3 border-b border-(--color-hairline) px-4 py-2">
        <TabsList>
          <TabsTrigger value="stderr">
            {t('buildDetail.stderr')}
            {stderrLines > 0 && <span className="ml-1.5 font-tabular text-(--color-status-failed)">{stderrLines}</span>}
          </TabsTrigger>
          <TabsTrigger value="stdout">
            {t('buildDetail.stdout')}
            <span className="ml-1.5 font-tabular text-(--color-ink-faint)">{lineCount(output.stdout)}</span>
          </TabsTrigger>
        </TabsList>
        {output.truncated && (
          <span className="text-[10px] uppercase tracking-[0.06em] text-(--color-status-timeout)">
            {t('console.truncatedNote')}
          </span>
        )}
      </div>
      <TabsContent value="stderr">
        <ConsolePane text={output.stderr} tone="failed" heightClass={heightClass} />
      </TabsContent>
      <TabsContent value="stdout">
        <ConsolePane text={output.stdout} heightClass={heightClass} />
      </TabsContent>
    </Tabs>
  )
}

function ConsolePane({ text, tone, heightClass }: { text: string; tone?: 'failed'; heightClass: string }) {
  const t = useT()
  return (
    <ScrollArea className={cn(heightClass, 'bg-(--color-surface-inset)')}>
      <pre
        className={cn(
          'whitespace-pre-wrap break-words p-4 font-mono text-[11px] leading-relaxed',
          !text && 'text-(--color-ink-faint)',
          text && (tone === 'failed' ? 'text-(--color-status-failed)' : 'text-(--color-ink-secondary)'),
        )}
      >
        {text || t('common.empty')}
      </pre>
    </ScrollArea>
  )
}
