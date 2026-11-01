import * as DropdownMenu from '@radix-ui/react-dropdown-menu'
import { Check, ChevronDown, Search, X } from 'lucide-react'
import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils'

const SEARCH_DEBOUNCE_MS = 200

interface SearchFieldProps {
  value: string
  onChange: (value: string) => void
  placeholder: string
  className?: string
}

/** Inset search input; debounced so typing does not rewrite the URL per keystroke. */
export function SearchField({ value, onChange, placeholder, className }: SearchFieldProps) {
  const [draft, setDraft] = useState(value)
  const [syncedValue, setSyncedValue] = useState(value)

  // External resets (e.g. "Reset filters") replace the draft; adjusting
  // during render avoids the extra pass an effect would cost.
  if (value !== syncedValue) {
    setSyncedValue(value)
    setDraft(value)
  }

  useEffect(() => {
    if (draft === value) return
    const id = setTimeout(() => onChange(draft), SEARCH_DEBOUNCE_MS)
    return () => clearTimeout(id)
  }, [draft, value, onChange])

  return (
    <label
      className={cn(
        'flex h-8 items-center gap-2 rounded-(--radius-sm) border border-(--color-hairline) bg-(--color-surface-inset) px-2.5 text-(--color-ink-muted) focus-within:border-(--color-focus-ring)',
        className,
      )}
    >
      <Search className="h-3.5 w-3.5 shrink-0" aria-hidden />
      <input
        type="search"
        value={draft}
        onChange={(event) => setDraft(event.target.value)}
        placeholder={placeholder}
        aria-label={placeholder}
        className="min-w-0 flex-1 bg-transparent text-[12px] text-(--color-ink) outline-none placeholder:text-(--color-ink-faint) [&::-webkit-search-cancel-button]:hidden"
      />
      {draft && (
        <button
          type="button"
          onClick={() => {
            setDraft('')
            onChange('')
          }}
          aria-label="Clear search"
          className="rounded-(--radius-xs) p-0.5 hover:text-(--color-ink)"
        >
          <X className="h-3 w-3" />
        </button>
      )}
    </label>
  )
}

interface Option<T extends string> {
  value: T
  label: string
}

interface SegmentedProps<T extends string> {
  value: T
  options: Option<T>[]
  onChange: (value: T) => void
  label: string
}

/** Mutually exclusive filter chips; a radiogroup so arrow-key users get the right semantics. */
export function Segmented<T extends string>({ value, options, onChange, label }: SegmentedProps<T>) {
  return (
    <div
      role="radiogroup"
      aria-label={label}
      className="inline-flex h-8 items-center gap-0.5 rounded-(--radius-sm) border border-(--color-hairline) bg-(--color-surface-inset) p-0.5"
    >
      {options.map((option) => {
        const active = option.value === value
        return (
          <button
            key={option.value}
            type="button"
            role="radio"
            aria-checked={active}
            onClick={() => onChange(option.value)}
            className={cn(
              'h-full rounded-[4px] px-2.5 text-[12px] font-medium transition-colors',
              active
                ? 'bg-(--color-canvas-overlay) text-(--color-ink) shadow-[0_0_0_1px_var(--color-hairline-strong)]'
                : 'text-(--color-ink-muted) hover:text-(--color-ink)',
            )}
          >
            {option.label}
          </button>
        )
      })}
    </div>
  )
}

interface MenuSelectProps<T extends string> {
  value: T
  options: Option<T>[]
  onChange: (value: T) => void
  label: string
}

/** Single-choice dropdown on the Radix menu primitive (keyboard, focus and collision handling included). */
export function MenuSelect<T extends string>({ value, options, onChange, label }: MenuSelectProps<T>) {
  const current = options.find((option) => option.value === value) ?? options[0]
  return (
    <DropdownMenu.Root>
      <DropdownMenu.Trigger
        aria-label={label}
        className="inline-flex h-8 items-center gap-2 rounded-(--radius-sm) border border-(--color-hairline) bg-(--color-surface-inset) px-2.5 text-[12px] font-medium text-(--color-ink-secondary) transition-colors hover:text-(--color-ink) data-[state=open]:border-(--color-hairline-strong)"
      >
        <span className="max-w-[160px] truncate">{current?.label}</span>
        <ChevronDown className="h-3 w-3 text-(--color-ink-faint)" />
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>
        <DropdownMenu.Content
          align="start"
          sideOffset={4}
          className="z-50 max-h-[320px] min-w-[180px] origin-(--radix-dropdown-menu-content-transform-origin) overflow-auto rounded-(--radius-md) border border-(--color-hairline-strong) bg-(--color-canvas-overlay) p-1 shadow-[0_8px_24px_-8px_oklch(0_0_0/50%)]"
        >
          <DropdownMenu.RadioGroup value={value} onValueChange={(next) => onChange(next as T)}>
            {options.map((option) => (
              <DropdownMenu.RadioItem
                key={option.value}
                value={option.value}
                className="flex cursor-pointer items-center gap-2 rounded-(--radius-xs) px-2 py-1.5 text-[12px] text-(--color-ink-secondary) outline-none data-[highlighted]:bg-(--color-canvas-raised) data-[highlighted]:text-(--color-ink)"
              >
                <span className="flex h-3 w-3 items-center justify-center">
                  <DropdownMenu.ItemIndicator>
                    <Check className="h-3 w-3 text-(--color-signal)" />
                  </DropdownMenu.ItemIndicator>
                </span>
                <span className="truncate font-mono">{option.label}</span>
              </DropdownMenu.RadioItem>
            ))}
          </DropdownMenu.RadioGroup>
        </DropdownMenu.Content>
      </DropdownMenu.Portal>
    </DropdownMenu.Root>
  )
}

interface PaginationProps {
  page: number
  pageCount: number
  onPage: (page: number) => void
  label: string
  prevLabel: string
  nextLabel: string
}

export function Pagination({ page, pageCount, onPage, label, prevLabel, nextLabel }: PaginationProps) {
  if (pageCount <= 1) return null
  const buttonClass =
    'h-7 rounded-(--radius-sm) border border-(--color-hairline) px-2.5 text-[12px] text-(--color-ink-secondary) transition-colors hover:bg-(--color-canvas-raised) hover:text-(--color-ink) disabled:pointer-events-none disabled:opacity-40'
  return (
    <div className="flex items-center justify-between gap-3 border-t border-(--color-hairline) px-4 py-2.5">
      <span className="font-tabular text-[11px] text-(--color-ink-faint)">{label}</span>
      <div className="flex gap-1.5">
        <button type="button" className={buttonClass} disabled={page <= 1} onClick={() => onPage(page - 1)}>
          {prevLabel}
        </button>
        <button type="button" className={buttonClass} disabled={page >= pageCount} onClick={() => onPage(page + 1)}>
          {nextLabel}
        </button>
      </div>
    </div>
  )
}
