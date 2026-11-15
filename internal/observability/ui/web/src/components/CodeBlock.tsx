import { Check, Copy } from 'lucide-react'
import { useState } from 'react'
import { cn } from '@/lib/utils'

interface CodeBlockProps {
  code: string
  className?: string
}

export function CodeBlock({ code, className }: CodeBlockProps) {
  const [copied, setCopied] = useState(false)

  async function handleCopy() {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    } catch {
      // Clipboard API unavailable (insecure context, permissions) — no-op.
    }
  }

  return (
    <div
      className={cn(
        'group relative rounded-(--radius-sm) border border-(--color-hairline) bg-(--color-surface-inset)',
        className,
      )}
    >
      <pre className="overflow-x-auto p-3 pr-10 font-mono text-[11.5px] leading-relaxed text-(--color-ink-secondary)">
        {code}
      </pre>
      <button
        type="button"
        onClick={handleCopy}
        aria-label="Copy to clipboard"
        className="absolute right-2 top-2 rounded-(--radius-xs) p-1.5 text-(--color-ink-faint) opacity-0 transition-opacity hover:bg-(--color-canvas-raised) hover:text-(--color-ink) group-hover:opacity-100 focus-visible:opacity-100"
      >
        {copied ? <Check className="h-3.5 w-3.5 text-(--color-signal)" /> : <Copy className="h-3.5 w-3.5" />}
      </button>
    </div>
  )
}
