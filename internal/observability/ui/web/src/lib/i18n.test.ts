import { describe, expect, it } from 'vitest'
import { dictionaries, translate } from './i18n'

function keys(dict: Record<string, Record<string, string>>): string[] {
  return Object.entries(dict).flatMap(([section, entries]) => Object.keys(entries).map((key) => `${section}.${key}`))
}

describe('dictionaries', () => {
  it('keeps en and vi key sets identical', () => {
    expect(keys(dictionaries.vi).sort()).toEqual(keys(dictionaries.en).sort())
  })

  it('interpolates variables and falls back to the key when missing', () => {
    expect(translate('en', 'overview.hintPassed', { ok: 3, total: 4 })).toBe('3/4 tasks passed')
    expect(translate('en', 'nope.missing')).toBe('nope.missing')
  })
})
