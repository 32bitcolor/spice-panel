import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Dew Reaper — water-harvesting scythe ─────────────────────────────────────
// The Dew Reaper sweeps moisture from the air at dawn. A long pole, a curved
// harvest blade echoing a crescent moon, and a pendant water drop at the tip.
export const DewReaperIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* handle pole — diagonal, grip at top right */}
    <line x1="18" y1="3" x2="11" y2="21" strokeWidth="1.5" />
    {/* curved harvest blade — sweeps from pole tip around to collection point */}
    <path
      d="M17.5 3.5 C21.5 7.5 21.5 17.5 11 21 C15.5 15 17 9.5 13 7.5 C14.5 4 17.5 3.5 17.5 3.5 Z"
      fill="currentColor"
      fillOpacity="0.16"
      strokeWidth="1.3"
    />
    {/* pendant water drop at collection tip — the whole point */}
    <path
      d="M11 19.5 C9 21 9.5 23 11.5 23 C13.5 23 14 21 12 19.5 C11.5 19 11 19.5 11 19.5 Z"
      fill="currentColor"
      fillOpacity="0.48"
      strokeWidth="0.9"
      strokeLinejoin="round"
    />
  </Svg>
)
