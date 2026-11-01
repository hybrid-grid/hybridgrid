import { useEffect } from 'react'
import { useSettingsStore } from '@/store/useSettingsStore'

export function useApplyTheme(): void {
  const theme = useSettingsStore((state) => state.theme)

  useEffect(() => {
    document.documentElement.dataset.theme = theme
  }, [theme])
}
