import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Desert combat helmet ──────────────────────────────────────────────────────
// A hard angular dome, a horizontal visor slit that cuts across the face like
// a wound, cheek flares that deflect sandblast, and a chin strap curve.
export const HelmetIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* dome */}
    <path
      d="M4 14.5 Q4 4 12 4 Q20 4 20 14.5 Z"
      fill="currentColor"
      fillOpacity="0.18"
      strokeWidth="1.4"
    />
    {/* left cheek guard — angled flare */}
    <path
      d="M4 14.5 L4.5 20.5 L8.5 21 L9.5 14.5"
      fill="currentColor"
      fillOpacity="0.12"
      strokeWidth="1.2"
    />
    {/* right cheek guard */}
    <path
      d="M20 14.5 L19.5 20.5 L15.5 21 L14.5 14.5"
      fill="currentColor"
      fillOpacity="0.12"
      strokeWidth="1.2"
    />
    {/* visor slit — the distinctive horizontal gap, heavy and shadowed */}
    <line x1="7.5" y1="13.5" x2="16.5" y2="13.5" strokeWidth="3.2" strokeOpacity="0.48" strokeLinecap="butt" />
    {/* visor brow ridge above the slit */}
    <path d="M6.5 12 L17.5 12" strokeWidth="0.85" strokeOpacity="0.32" />
    {/* chin strap curve at base */}
    <path d="M4.5 20.5 Q12 23.5 19.5 20.5" strokeWidth="1.2" fill="none" />
  </Svg>
)
