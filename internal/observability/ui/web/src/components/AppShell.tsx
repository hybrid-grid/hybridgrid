import type { ReactNode } from 'react'
import { useApplyTheme } from '@/hooks/useApplyTheme'
import { ConnectionStatus } from './ConnectionStatus'
import { LanguageToggle, ThemeToggle } from './SettingsToggles'
import { MobileNav, Sidebar } from './Sidebar'

interface AppShellProps {
  children: ReactNode
}

export function AppShell({ children }: AppShellProps) {
  useApplyTheme()

  return (
    <div className="flex h-screen w-screen overflow-hidden bg-(--color-canvas) text-(--color-ink)">
      <Sidebar />
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-11 shrink-0 items-center gap-2 border-b border-(--color-hairline) px-3 md:px-5">
          <MobileNav />
          <div className="flex-1" />
          <LanguageToggle />
          <ThemeToggle />
          <ConnectionStatus />
        </header>
        <main className="min-h-0 flex-1 overflow-auto p-4 md:p-6">{children}</main>
      </div>
    </div>
  )
}
