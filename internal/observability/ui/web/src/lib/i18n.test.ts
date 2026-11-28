import { describe, expect, it } from 'vitest'
import { dictionaries, translate } from './i18n'

function keys(dict: Record<string, Record<string, string>>): string[] {
  return Object.entries(dict).flatMap(([section, entries]) => Object.keys(entries).map((key) => `${section}.${key}`))
}

describe('dictionaries', () => {
  it('keeps en and vi key sets identical', () => {
    expect(keys(dictionaries.vi).sort()).toEqual(keys(dictionaries.en).sort())
  })

  it('uses the same placeholders in vi as in en', () => {
    const placeholders = (text: string) => [...text.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort()
    for (const [section, entries] of Object.entries(dictionaries.en)) {
      for (const [key, text] of Object.entries(entries)) {
        expect(placeholders(dictionaries.vi[section][key]), `${section}.${key}`).toEqual(placeholders(text))
      }
    }
  })

  it('interpolates variables and falls back to the key when missing', () => {
    expect(translate('en', 'overview.hintPassed', { ok: 3, total: 4 })).toBe('3/4 tasks passed')
    expect(translate('en', 'nope.missing')).toBe('nope.missing')
  })
})
