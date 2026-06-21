import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Static compactor — industrial compression machine ─────────────────────────
// Two heavy plates — one fixed, one driven — squeezing a block between them.
// Directional chevron arrows make the force visible.
export const CompactorIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* top driven plate */}
    <rect x="3" y="3.5" width="18" height="4.5" rx="1" fill="currentColor" fillOpacity="0.2" strokeWidth="1.4" />
    {/* bottom fixed plate */}
    <rect x="3" y="16" width="18" height="4.5" rx="1" fill="currentColor" fillOpacity="0.2" strokeWidth="1.4" />
    {/* material being compacted */}
    <rect x="6.5" y="10.5" width="11" height="3" rx="0.5" fill="currentColor" fillOpacity="0.32" strokeWidth="1.1" />
    {/* downward force arrow from upper plate */}
    <path d="M12 8 L12 10.5" strokeWidth="1.4" />
    <path d="M10.2 9.8 L12 11.5 L13.8 9.8" strokeWidth="1.3" fill="none" />
    {/* upward reaction arrow from lower plate */}
    <path d="M12 13.5 L12 16" strokeWidth="1.4" />
    <path d="M10.2 14.2 L12 12.5 L13.8 14.2" strokeWidth="1.3" fill="none" />
  </Svg>
)
