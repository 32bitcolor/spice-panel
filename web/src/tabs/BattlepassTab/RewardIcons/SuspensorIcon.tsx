import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Suspensor pack — anti-gravity levitation device ───────────────────────────
// A softly glowing device that punches upward thrust. A spherical core hovers
// above a wide levitation field ellipse; emission lines spike outward like heat shimmer.
export const SuspensorIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* levitation field — wide flat ellipse below the device */}
    <ellipse cx="12" cy="17.5" rx="9.5" ry="3" strokeWidth="1.4" />
    {/* secondary field ring — slightly inside, faded */}
    <ellipse cx="12" cy="16" rx="6" ry="1.8" strokeWidth="0.9" strokeOpacity="0.38" />
    {/* device body — the actual suspensor housing */}
    <circle cx="12" cy="10.5" r="4" fill="currentColor" fillOpacity="0.18" strokeWidth="1.4" />
    {/* glowing energy point at core */}
    <circle cx="12" cy="10.5" r="1.6" fill="currentColor" fillOpacity="0.55" strokeWidth="0" />
    {/* upward emission lines — three rays */}
    <line x1="12" y1="6.5" x2="12" y2="4.5" strokeWidth="1.2" strokeOpacity="0.42" />
    <line x1="15.5" y1="7.5" x2="17" y2="6" strokeWidth="1" strokeOpacity="0.35" />
    <line x1="8.5" y1="7.5" x2="7" y2="6" strokeWidth="1" strokeOpacity="0.35" />
  </Svg>
)
