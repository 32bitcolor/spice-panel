import { describe, it, expect } from 'vitest'
import { dowLabel } from './dowLabel'

// Jan 1 2023 00:00:00Z is a Sunday (d=0). Without timeZone:'UTC' the formatter
// interprets that UTC midnight as Dec 31 2022 in any timezone west of UTC, which
// shifts the label to 'Sat'. The tests below pin 'en-US' with TZ=America/New_York
// (set in vitest.config.ts) so the bug is caught even on UTC CI runners.
describe('dowLabel', () => {
  it('returns Sun for d=0 (UTC midnight anchor must not shift west of UTC)', () => {
    expect(dowLabel(0, 'en-US')).toBe('Sun')
  })

  it('returns all seven correct short labels in order', () => {
    const expected = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']
    expect([0, 1, 2, 3, 4, 5, 6].map((d) => dowLabel(d, 'en-US'))).toEqual(expected)
  })
})
