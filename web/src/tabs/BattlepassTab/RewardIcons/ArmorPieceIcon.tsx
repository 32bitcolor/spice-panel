import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Single armor piece — pauldron / chest plate ───────────────────────────────
// For individual non-helmet armor pieces (chest, gloves, boots, legs).
// A wider shield profile without the stacked depth — one solid plate.
export const ArmorPieceIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    <path
      d="M4 6 L12 3 L20 6 L20 15 Q20 21 12 23 Q4 21 4 15 Z"
      fill="currentColor"
      fillOpacity="0.2"
      strokeWidth="1.5"
    />
    {/* mid-panel horizontal crease */}
    <path d="M6 13 L18 13" strokeWidth="0.85" strokeOpacity="0.4" />
    {/* central boss */}
    <circle cx="12" cy="9.5" r="2.2" fill="currentColor" fillOpacity="0.28" strokeWidth="0.9" />
  </Svg>
)
