import { twMerge } from 'tailwind-merge'

/**
 * Class-name combiner for the spice-panel UI library. Accepts the same loose
 * argument shapes as `clsx` (strings, arrays, conditional objects) and resolves
 * Tailwind conflicts via `tailwind-merge` so later utilities win.
 */
export type ClassValue =
  | string
  | number
  | null
  | undefined
  | false
  | ClassValue[]
  | Record<string, boolean | null | undefined>

const toString = (value: ClassValue): string => {
  if (value === null || value === undefined || value === false) return ''
  if (typeof value === 'string' || typeof value === 'number') return String(value)
  if (Array.isArray(value)) return value.map(toString).filter(Boolean).join(' ')
  return Object.keys(value)
    .filter((key) => Boolean(value[key]))
    .join(' ')
}

export const cn = (...inputs: ClassValue[]): string =>
  twMerge(inputs.map(toString).filter(Boolean).join(' '))
