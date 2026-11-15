import { create } from 'zustand'
import { persist } from 'zustand/middleware'

export type Theme = 'dark' | 'light'
export type Lang = 'en' | 'vi'

interface SettingsState {
  theme: Theme
  lang: Lang
  setTheme: (theme: Theme) => void
  toggleTheme: () => void
  setLang: (lang: Lang) => void
}

export const useSettingsStore = create<SettingsState>()(
  persist(
    (set) => ({
      theme: 'dark',
      lang: 'en',
      setTheme: (theme) => set({ theme }),
      toggleTheme: () => set((state) => ({ theme: state.theme === 'dark' ? 'light' : 'dark' })),
      setLang: (lang) => set({ lang }),
    }),
    { name: 'hg-settings' },
  ),
)
