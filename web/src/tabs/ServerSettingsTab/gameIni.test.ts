import { describe, it, expect } from 'vitest'
import type { ServerSetting } from '../../api/client'
import { buildGameIni, formatValue, isModified } from './gameIni'

const setting = (over: Partial<ServerSetting>): ServerSetting => ({
  section: '/Script/DuneSandbox.BuildingSettings',
  key: 'Field',
  type: 'float',
  default: '1',
  label: '',
  description: '',
  category: '',
  current: '1',
  is_overridden: false,
  source: '',
  layers: [],
  ...over,
})

describe('formatValue', () => {
  it('formats booleans as True/False', () => {
    expect(formatValue('true', 'bool')).toBe('True')
    expect(formatValue('1', 'bool')).toBe('True')
    expect(formatValue('false', 'bool')).toBe('False')
    expect(formatValue('0', 'bool')).toBe('False')
  })

  it('normalises floats', () => {
    expect(formatValue('2.500', 'float')).toBe('2.5')
    expect(formatValue('3', 'float')).toBe('3')
  })

  it('keeps ints as integers', () => {
    expect(formatValue('42', 'int')).toBe('42')
  })

  it('emits strings verbatim and falls back on unparsable numbers', () => {
    expect(formatValue('Hello', 'string')).toBe('Hello')
    expect(formatValue('notanumber', 'float')).toBe('notanumber')
  })
})

describe('isModified', () => {
  it('is false when current equals default', () => {
    expect(isModified(setting({ current: '1', default: '1' }))).toBe(false)
  })

  it('is true when current differs from default', () => {
    expect(isModified(setting({ current: '2', default: '1' }))).toBe(true)
  })

  it('is true when overridden even if values match', () => {
    expect(isModified(setting({ current: '1', default: '1', is_overridden: true }))).toBe(true)
  })
})

describe('buildGameIni', () => {
  it('returns empty string when nothing modified', () => {
    expect(buildGameIni([setting({}), setting({ current: '1', default: '1' })])).toBe('')
  })

  it('groups modified settings under their real [Section] headers', () => {
    const out = buildGameIni([
      setting({ section: '/Script/DuneSandbox.BuildingSettings', key: 'MaxHealth', type: 'int', default: '100', current: '200' }),
      setting({ section: '/Script/DuneSandbox.BuildingSettings', key: 'Decay', type: 'bool', default: 'true', current: 'false' }),
      setting({ section: '/Script/DuneSandbox.MiningSettings', key: 'Multiplier', type: 'float', default: '1.0', current: '2.500' }),
    ])
    expect(out).toBe(
      '[/Script/DuneSandbox.BuildingSettings]\n'
      + 'MaxHealth=200\n'
      + 'Decay=False\n\n'
      + '[/Script/DuneSandbox.MiningSettings]\n'
      + 'Multiplier=2.5\n',
    )
  })

  it('excludes unmodified settings from a section that has modified ones', () => {
    const out = buildGameIni([
      setting({ key: 'Changed', default: '1', current: '5' }),
      setting({ key: 'Unchanged', default: '1', current: '1' }),
    ])
    expect(out).toContain('Changed=5')
    expect(out).not.toContain('Unchanged')
  })
})
