import { Check, ChevronRight, Copy, Download, FileArchive, FileCode2 } from 'lucide-react'
import { useState } from 'react'
import { CodeBlock } from '@/components/CodeBlock'
import { Panel } from '@/components/Panel'
import { Button } from '@/components/ui/button'
import { ApiError, artifactURL } from '@/lib/api'
import { useRelativeTime, useT } from '@/lib/i18n'
import { useArtifacts } from '@/lib/queries'
import type { Artifact } from '@/lib/types'
import { cn, formatSize } from '@/lib/utils'

const ARCHIVE = /\.(tar|tgz|zst|gz|zip|7z|apk|aab|ipa|jar)$/i

/**
 * A build's stored outputs with one-click download. Empty and disabled
 * states say exactly what to run, since artifacts only exist when the
 * build command asks for them (C/C++ links on the client) or a remote
 * builder produced them (Flutter/Unity/plugins).
 */
export function ArtifactsPanel({ buildId }: { buildId: string }) {
  const t = useT()
  const { data, isLoading, error } = useArtifacts(buildId)
  const artifacts = data?.artifacts ?? []
  const disabled = error instanceof ApiError && error.status === 503

  return (
    <Panel
      title={t('buildDetail.artifactsTitle')}
      eyebrow={t('buildDetail.artifactsCaption')}
      action={
        artifacts.length > 0 && (
          <span className="font-tabular text-[12px] text-(--color-ink-muted)">{t('buildDetail.artifactsCount', { n: artifacts.length })}</span>
        )
      }
      padded={false}
    >
      {isLoading ? (
        <div className="p-4 text-[12px] text-(--color-ink-muted)">{t('buildDetail.loading')}</div>
      ) : disabled ? (
        <div className="p-4 text-[12px] text-(--color-ink-muted)">{t('buildDetail.artifactsDisabled')}</div>
      ) : error ? (
        <div className="p-4 text-[12px] text-(--color-status-failed)">{t('buildDetail.artifactsError')}</div>
      ) : artifacts.length === 0 ? (
        <EmptyState buildId={buildId} />
      ) : (
        <ul className="divide-y divide-(--color-hairline)">
          {artifacts.map((artifact) => (
            <ArtifactRow key={artifact.name} artifact={artifact} />
          ))}
        </ul>
      )}
    </Panel>
  )
}

function ArtifactRow({ artifact }: { artifact: Artifact }) {
  const t = useT()
  const relativeTime = useRelativeTime()
  const [open, setOpen] = useState(false)
  const Icon = ARCHIVE.test(artifact.name) ? FileArchive : FileCode2
  const hasFiles = artifact.files.length > 0

  return (
    <li className="px-4 py-3">
      <div className="flex items-center gap-3">
        <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-(--radius-md) bg-(--color-surface-inset) text-(--color-ink-secondary)">
          <Icon className="h-4 w-4" />
        </span>
        <div className="min-w-0 flex-1">
          <div className="truncate font-mono text-[13px] font-medium text-(--color-ink)" title={artifact.name}>
            {artifact.name}
          </div>
          <div className="mt-0.5 flex flex-wrap items-center gap-x-2 gap-y-0.5 text-[11px] text-(--color-ink-muted)">
            <span className="font-tabular text-(--color-ink-secondary)">{formatSize(artifact.size_bytes)}</span>
            <span aria-hidden>·</span>
            <span>{artifact.source === 'worker' ? t('buildDetail.artifactsFromWorker') : t('buildDetail.artifactsFromClient')}</span>
            <span aria-hidden>·</span>
            <span>{relativeTime(artifact.created_at_ms / 1000)}</span>
            <span aria-hidden>·</span>
            <ShaChip sha={artifact.sha256} />
          </div>
        </div>
        <Button variant="primary" size="sm" asChild>
          <a href={artifactURL(artifact.build_id, artifact.name)} download={artifact.name}>
            <Download className="h-3.5 w-3.5" />
            {t('buildDetail.artifactsDownload')}
          </a>
        </Button>
      </div>

      {hasFiles && (
        <div className="mt-2 pl-12">
          <button
            type="button"
            onClick={() => setOpen((v) => !v)}
            aria-expanded={open}
            className="flex items-center gap-1 text-[11px] text-(--color-ink-muted) hover:text-(--color-ink)"
          >
            <ChevronRight className={cn('h-3 w-3 transition-transform', open && 'rotate-90')} />
            {t('buildDetail.artifactsInside', { n: artifact.files.length })}
          </button>
          {open && (
            <ul className="mt-1.5 space-y-1 rounded-(--radius-md) bg-(--color-surface-inset) px-3 py-2">
              {artifact.files.map((file) => (
                <li key={file.path} className="flex items-center justify-between gap-3 font-mono text-[11px]">
                  <span className="truncate text-(--color-ink-secondary)" title={file.path}>
                    {file.path}
                  </span>
                  <span className="shrink-0 font-tabular text-(--color-ink-faint)">{formatSize(file.size_bytes)}</span>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </li>
  )
}

/** Short checksum with copy-to-clipboard, for verifying a download. */
function ShaChip({ sha }: { sha: string }) {
  const t = useT()
  const [copied, setCopied] = useState(false)
  if (!sha) return null
  async function copy() {
    try {
      await navigator.clipboard.writeText(sha)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard unavailable (insecure context): the full hash is in the tooltip.
    }
  }
  return (
    <button
      type="button"
      onClick={copy}
      title={`SHA-256 ${sha}`}
      aria-label={copied ? t('buildDetail.artifactsCopied') : t('buildDetail.artifactsCopySha')}
      className="inline-flex items-center gap-1 font-mono text-(--color-ink-faint) hover:text-(--color-ink)"
    >
      sha256:{sha.slice(0, 10)}
      {copied ? <Check className="h-3 w-3 text-(--color-status-success)" /> : <Copy className="h-3 w-3" />}
    </button>
  )
}

function EmptyState({ buildId }: { buildId: string }) {
  const t = useT()
  return (
    <div className="flex flex-col gap-2 p-4 text-[12px] text-(--color-ink-muted)">
      <div className="text-(--color-ink-secondary)">{t('buildDetail.artifactsEmpty')}</div>
      <div>{t('buildDetail.artifactsEmptyHint')}</div>
      <CodeBlock code="hgbuild make -j8 --artifact <file>" />
      <div className="mt-1">{t('buildDetail.artifactsEmptyHintLater')}</div>
      <CodeBlock code={`hgbuild artifact upload --build ${buildId} <file>`} />
    </div>
  )
}
