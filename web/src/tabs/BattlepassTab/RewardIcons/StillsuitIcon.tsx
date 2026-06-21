import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Stillsuit — precious water recirculation survival suit ────────────────────
// Arrakis' most sacred technology. A teardrop silhouette (water = life)
// with internal tube loops etched inside — the stillsuit's capillary network.
export const StillsuitIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* outer droplet — slightly stylised, not perfectly round at the base */}
    <path
      d="M12 3 C8.5 9 6 13.5 6 16 A6 6 0 0 0 18 16 C18 13.5 15.5 9 12 3 Z"
      fill="currentColor"
      fillOpacity="0.18"
      strokeWidth="1.4"
    />
    {/* upper recirculation loop */}
    <path
      d="M9 12.5 C10.5 11.5 12 13 13.5 12"
      strokeWidth="1"
      strokeOpacity="0.58"
      fill="none"
    />
    {/* lower recirculation loop */}
    <path
      d="M9.5 15.5 C11 14.5 12.5 16 14 15"
      strokeWidth="1"
      strokeOpacity="0.58"
      fill="none"
    />
    {/* central micro-filter valve dot */}
    <circle cx="12" cy="17.5" r="1.2" fill="currentColor" fillOpacity="0.48" strokeWidth="0.6" />
  </Svg>
)
