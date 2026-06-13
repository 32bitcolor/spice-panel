import type { ServerSetting } from '../../api/client'

// Format a single value for emission into a Game.ini file according to its
// declared type. Booleans become True/False (UE INI convention), floats are
// normalised (trailing zeros trimmed), ints kept as integers; everything else
// is emitted verbatim. Falls back to the raw string when parsing fails.
const formatValue = (value: string, type: ServerSetting['type']): string => {
  const v = value.trim()
  switch (type) {
    case 'bool': {
      const lower = v.toLowerCase()
      if (lower === 'true' || lower === '1') return 'True'
      if (lower === 'false' || lower === '0') return 'False'
      return v
    }
    case 'float': {
      const n = Number(v)
      return Number.isFinite(n) ? String(n) : v
    }
    case 'int': {
      const n = Number(v)
      return Number.isInteger(n) ? String(n) : v
    }
    default:
      return v
  }
}

// A setting is considered modified when its current value differs from its
// default. is_overridden reflects whether a user layer wrote the value, so we
// treat either signal as "the operator changed it".
const isModified = (s: ServerSetting): boolean => {
  if (s.is_overridden) return true
  return s.current.trim() !== s.default.trim()
}

/**
 * Build the text of a client-side Game.ini containing only the operator's
 * modified (non-default) curated server settings, grouped under their real INI
 * [Section] headers. Returns an empty string when nothing has been modified.
 *
 * Pure function — no DOM access — so it is trivially unit-testable.
 */
const buildGameIni = (settings: ServerSetting[]): string => {
  const bySection = new Map<string, string[]>()
  // Preserve first-seen section order for stable, diff-friendly output.
  const order: string[] = []

  for (const s of settings) {
    if (!isModified(s)) continue
    if (!bySection.has(s.section)) {
      bySection.set(s.section, [])
      order.push(s.section)
    }
    bySection.get(s.section)!.push(`${s.key}=${formatValue(s.current, s.type)}`)
  }

  if (order.length === 0) return ''

  return order
    .map((section) => `[${section}]\n${bySection.get(section)!.join('\n')}`)
    .join('\n\n') + '\n'
}

export { buildGameIni, formatValue, isModified }
