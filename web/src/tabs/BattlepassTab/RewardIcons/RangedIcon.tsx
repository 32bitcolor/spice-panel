import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Ranged weapon — plasma pistol / SMG / rifle silhouette ────────────────────
// Geometric sci-fi side profile: extended barrel, boxy receiver, angled grip, sight pin.
export const RangedIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* barrel */}
    <rect x="2" y="10.5" width="13" height="3" rx="1" fill="currentColor" fillOpacity="0.15" strokeWidth="1.3" />
    {/* receiver / body */}
    <path d="M14 8 L21 8 L21 15 L14 15 Z" fill="currentColor" fillOpacity="0.15" strokeWidth="1.4" />
    {/* angled grip */}
    <path
      d="M17 15 L15.5 21 L19.5 21 L21 15"
      fill="currentColor"
      fillOpacity="0.1"
      strokeWidth="1.3"
    />
    {/* sight pin on top of receiver */}
    <line x1="19" y1="8" x2="19" y2="6.5" strokeWidth="1.5" />
    {/* muzzle flash dot */}
    <circle cx="3" cy="12" r="0.9" fill="currentColor" strokeWidth="0" />
    {/* ejection port detail */}
    <line x1="16" y1="10" x2="19" y2="10" strokeWidth="0.75" strokeOpacity="0.45" />
  </Svg>
)
