import { describe, it, expect } from 'vitest'
import { retainSkippedStaged } from './giveItemsHelpers'

type Staged = { template: string, qty: number, quality: number }

describe('retainSkippedStaged', () => {
  it('returns empty when all items were given', () => {
    const staged: Staged[] = [
      { template: 'A', qty: 1, quality: 0 },
      { template: 'B', qty: 2, quality: 1 },
    ]
    const given: string[] = ['A', 'B']
    expect(retainSkippedStaged(staged, given)).toEqual([])
  })

  it('retains all staged when nothing was given', () => {
    const staged: Staged[] = [
      { template: 'A', qty: 1, quality: 0 },
      { template: 'B', qty: 2, quality: 0 },
    ]
    expect(retainSkippedStaged(staged, [])).toEqual(staged)
  })

  it('removes only given items, keeps skipped', () => {
    const staged: Staged[] = [
      { template: 'Ammo', qty: 500, quality: 0 },
      { template: 'Kindjal', qty: 1, quality: 0 },
      { template: 'HeavyArmor', qty: 1, quality: 0 },
    ]
    const given: string[] = ['Ammo', 'Kindjal']
    const result = retainSkippedStaged(staged, given)
    expect(result).toHaveLength(1)
    expect(result[0].template).toBe('HeavyArmor')
  })

  it('handles duplicate templates: removes one given, keeps skipped copy', () => {
    // Two staged rows with the same template — e.g. user added it twice.
    const staged: Staged[] = [
      { template: 'Ammo', qty: 100, quality: 0 },
      { template: 'Ammo', qty: 200, quality: 0 },
    ]
    // Backend gave one, skipped one.
    const result = retainSkippedStaged(staged, ['Ammo'])
    // Should retain exactly one Ammo row (the skipped one).
    expect(result).toHaveLength(1)
    expect(result[0].template).toBe('Ammo')
  })

  it('given wins over skipped when template appears in both (edge case)', () => {
    // If somehow the same template appears in both given and skipped,
    // we remove the given count and keep the skipped count.
    const staged: Staged[] = [
      { template: 'X', qty: 5, quality: 0 },
      { template: 'X', qty: 3, quality: 0 },
      { template: 'X', qty: 1, quality: 0 },
    ]
    const given: string[] = ['X', 'X'] // two given
    const result = retainSkippedStaged(staged, given)
    // 3 staged, 2 given → 1 retained
    expect(result).toHaveLength(1)
    expect(result[0].template).toBe('X')
  })

  it('preserves qty and quality of retained rows', () => {
    const staged: Staged[] = [
      { template: 'Sword', qty: 1, quality: 5 },
    ]
    const result = retainSkippedStaged(staged, [])
    expect(result[0]).toEqual({ template: 'Sword', qty: 1, quality: 5 })
  })

  it('empty staged input returns empty', () => {
    expect(retainSkippedStaged([], ['A'])).toEqual([])
  })
})
