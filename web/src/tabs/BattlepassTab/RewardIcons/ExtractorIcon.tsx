import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Body fluid extractor — survival vial ─────────────────────────────────────
// A sealed glass vial with a stopper, three calibration marks, and a dark
// fluid meniscus — equal parts medical instrument and Arrakis survival kit.
export const ExtractorIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* vial body — rounded at the bottom where fluid collects */}
    <path
      d="M9 4 L9 17.5 Q9 22 12 22 Q15 22 15 17.5 L15 4 Z"
      fill="currentColor"
      fillOpacity="0.15"
      strokeWidth="1.4"
    />
    {/* stopper cap */}
    <rect x="7.5" y="2" width="9" height="3" rx="0.75" fill="currentColor" fillOpacity="0.22" strokeWidth="1.2" />
    {/* fluid meniscus line */}
    <path d="M9.5 15 Q12 13.5 14.5 15" strokeWidth="1" strokeOpacity="0.55" fill="none" />
    {/* calibration marks on left wall */}
    <line x1="9" y1="10.5" x2="11" y2="10.5" strokeWidth="0.8" strokeOpacity="0.45" />
    <line x1="9" y1="13" x2="11" y2="13" strokeWidth="0.8" strokeOpacity="0.45" />
    <line x1="9" y1="17.5" x2="11" y2="17.5" strokeWidth="0.8" strokeOpacity="0.45" />
    {/* fluid content — dark settled at the bottom */}
    <path
      d="M9.5 16.5 Q12 15.5 14.5 16.5 L14.5 18.5 Q12 21.5 9.5 18.5 Z"
      fill="currentColor"
      fillOpacity="0.38"
      strokeWidth="0"
    />
  </Svg>
)
