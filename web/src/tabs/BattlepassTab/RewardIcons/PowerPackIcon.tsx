import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Power pack — Old Sparky energy cell ──────────────────────────────────────
// The portable power unit that drives shields and kit. Classic battery silhouette
// with a bold lightning bolt — the kinetic promise inside.
export const PowerPackIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* battery body */}
    <rect x="5.5" y="5.5" width="13" height="16.5" rx="2" fill="currentColor" fillOpacity="0.15" strokeWidth="1.4" />
    {/* positive terminal nub */}
    <rect x="9" y="3.5" width="6" height="2.5" rx="1" fill="currentColor" fillOpacity="0.25" strokeWidth="1.2" />
    {/* lightning bolt — the energy inside */}
    <path
      d="M13.5 8.5 L10 14 L13 14 L10.5 19.5 L16 13 L12.5 13 Z"
      fill="currentColor"
      fillOpacity="0.62"
      strokeWidth="0.75"
    />
  </Svg>
)
