import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Generic schematic scroll — fallback for unclassified rewards ──────────────
// A rolled blueprint — clearly "a crafting schematic" for anything that doesn't
// have a more specific icon.
export const SchematicIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* scroll body */}
    <rect x="5" y="5" width="14" height="15" rx="1" fill="currentColor" fillOpacity="0.15" strokeWidth="1.4" />
    {/* top rolled edge */}
    <path d="M5 6 Q5 3 8 5 L16 5 Q19 3 19 6" fill="currentColor" fillOpacity="0.1" strokeWidth="1.2" />
    {/* bottom rolled edge */}
    <path d="M5 19 Q5 22 8 20 L16 20 Q19 22 19 19" fill="currentColor" fillOpacity="0.1" strokeWidth="1.2" />
    {/* blueprint lines */}
    <line x1="8" y1="9.5" x2="16" y2="9.5" strokeWidth="0.9" strokeOpacity="0.5" />
    <line x1="8" y1="12.5" x2="16" y2="12.5" strokeWidth="0.9" strokeOpacity="0.5" />
    <line x1="8" y1="15.5" x2="13" y2="15.5" strokeWidth="0.9" strokeOpacity="0.5" />
  </Svg>
)
