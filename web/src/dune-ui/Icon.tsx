import type React from 'react'
import { Icon as IconifyIcon } from '@iconify/react'

type IconProps = {
  /** Lucide icon name (without the `lucide:` prefix), e.g. "refresh-cw". */
  name: string
  /** Optional size class — defaults to `size-4` (1rem square). */
  className?: string
}

/**
 * Thin wrapper around `@iconify/react` that defaults to the lucide icon set
 * and a sensible inline-text size. Use any lucide icon name from
 * https://lucide.dev/icons (kebab-case).
 */
export const Icon: React.FC<IconProps> = ({ name, className = 'size-4' }) => (
  <IconifyIcon icon={`lucide:${name}`} className={className} />
)
