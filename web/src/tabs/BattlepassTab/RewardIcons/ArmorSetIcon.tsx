import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Full armor set — layered heraldic shields ─────────────────────────────────
// Three shields slightly staggered in depth — communicates "multiple pieces,
// a complete protection system." The foremost shield has a central boss and
// an inlaid horizontal stripe.
export const ArmorSetIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* rearmost shield — offset up-right, lowest opacity */}
    <path
      d="M16 3 L22 6 L22 13.5 Q22 19 16 21.5"
      fill="currentColor"
      fillOpacity="0.08"
      strokeWidth="1.1"
      strokeOpacity="0.4"
    />
    {/* middle shield */}
    <path
      d="M12.5 4 L19 6.5 L19 14 Q19 19.5 12.5 22 Q6 19.5 6 14 L6 6.5 Z"
      fill="currentColor"
      fillOpacity="0.12"
      strokeWidth="1.2"
      strokeOpacity="0.55"
    />
    {/* front shield — main, boldest */}
    <path
      d="M4.5 5 L12 2 L19.5 5 L19.5 13 Q19.5 19.5 12 22.5 Q4.5 19.5 4.5 13 Z"
      fill="currentColor"
      fillOpacity="0.2"
      strokeWidth="1.5"
    />
    {/* horizontal stripe across the front shield (heraldic band) */}
    <path d="M6.5 12 L17.5 12" strokeWidth="0.9" strokeOpacity="0.38" />
    {/* central boss / umbo */}
    <circle cx="12" cy="10.5" r="2.5" fill="currentColor" fillOpacity="0.3" strokeWidth="0.9" />
  </Svg>
)
