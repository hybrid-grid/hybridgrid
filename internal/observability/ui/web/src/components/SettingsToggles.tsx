import { Languages, Moon, Sun } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { useSettingsStore } from '@/store/useSettingsStore'

export function ThemeToggle() {
  const theme = useSettingsStore((state) => state.theme)
  const toggleTheme = useSettingsStore((state) => state.toggleTheme)

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={toggleTheme}
      aria-label="Toggle color theme"
      title={theme === 'dark' ? 'Switch to light theme' : 'Switch to dark theme'}
    >
      {theme === 'dark' ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
    </Button>
  )
}

export function LanguageToggle() {
  const lang = useSettingsStore((state) => state.lang)
  const setLang = useSettingsStore((state) => state.setLang)

  return (
    <Button
      variant="ghost"
      size="sm"
      onClick={() => setLang(lang === 'en' ? 'vi' : 'en')}
      aria-label="Toggle language"
      title="English / Tiếng Việt"
      className="gap-1.5 font-mono uppercase"
    >
      <Languages className="h-3.5 w-3.5" />
      {lang}
    </Button>
  )
}
