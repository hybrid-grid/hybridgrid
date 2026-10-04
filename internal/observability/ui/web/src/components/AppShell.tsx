import type { ReactNode } from 'react'
import { useApplyTheme } from '@/hooks/useApplyTheme'
import { ConnectionStatus } from './ConnectionStatus'
import { LanguageToggle, ThemeToggle } from './SettingsToggles'
import { Sidebar } from './Sidebar'

interface AppShellProps {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  useApplyTheme()

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-(--color-canvas) text-(--color-ink)">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-11 shrink-0 items-center justify-end gap-2 border-b border-(--color-hairline) px-5">
          <LanguageToggle />
          <ThemeToggle />
          <ConnectionStatus />
        </header>
        <main className="min-h-0 flex-1 overflow-auto p-6">{children}</main>
      </div>
    </div>
  )
}
