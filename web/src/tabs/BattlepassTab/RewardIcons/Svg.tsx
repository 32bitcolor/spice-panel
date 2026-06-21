import * as React from 'react'
import type { IconProps } from './types'

// All icons share this SVG wrapper: 24×24 viewBox, currentColor, Lucide-style defaults
const Svg: React.FC<IconProps & { children: React.ReactNode }> = ({ className, children }): React.ReactElement => (
  <svg
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    strokeWidth="1.5"
    strokeLinecap="round"
    strokeLinejoin="round"
    className={className ?? 'w-6 h-6'}
    aria-hidden="true"
  >
    {children}
  </svg>
)

export default Svg
