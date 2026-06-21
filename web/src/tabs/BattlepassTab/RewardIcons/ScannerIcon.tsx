import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Scanner — handheld life / survey scanner ──────────────────────────────────
// A radar display: outer housing ring, inner detection ring, a rotating sweep
// beam frozen mid-arc, and a blip dot where the beam last hit.
export const ScannerIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* outer housing ring */}
    <circle cx="12" cy="12" r="9.5" strokeWidth="1.4" />
    {/* inner detection radius ring */}
    <circle cx="12" cy="12" r="5.5" strokeWidth="0.9" strokeOpacity="0.42" />
    {/* sweep beam */}
    <line x1="12" y1="12" x2="20" y2="6.5" strokeWidth="1.4" />
    {/* trailing arc ghost of the sweep */}
    <path d="M 12 2.5 A 9.5 9.5 0 0 1 20 6.5" strokeWidth="1.1" strokeOpacity="0.48" fill="none" />
    {/* centre pivot dot */}
    <circle cx="12" cy="12" r="1.5" fill="currentColor" strokeWidth="0" />
    {/* contact blip on the inner ring */}
    <circle cx="17" cy="8" r="1.1" fill="currentColor" fillOpacity="0.52" strokeWidth="0" />
  </Svg>
)
