import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Flamethrower — Shaitan's Tongue ──────────────────────────────────────────
// A compact fuel tank on the hip, short nozzle, and a roiling mushroom of flame
// with a brighter inner core — fire is alive, not just a shape.
export const FlameIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* fuel tank — compact cylinder */}
    <rect x="2" y="14" width="5.5" height="7" rx="2" fill="currentColor" fillOpacity="0.18" strokeWidth="1.3" />
    {/* nozzle / barrel connecting tank to flame */}
    <line x1="7.5" y1="16" x2="14" y2="13.5" strokeWidth="2.2" />
    {/* outer flame billow — irregular organic shape */}
    <path
      d="M14 13.5 C15.5 8.5 20 9.5 19 4.5 C23 7 23.5 14.5 19.5 17 C18 18.5 15.5 18 14 13.5 Z"
      fill="currentColor"
      fillOpacity="0.18"
      strokeWidth="1.3"
    />
    {/* inner hot core — brighter, tighter */}
    <path
      d="M15.5 13.5 C16.5 10.5 19.5 11.5 19 8.5 C21.5 10.5 21.5 15 19 16.5 C17.5 17.5 15.5 16.5 15.5 13.5 Z"
      fill="currentColor"
      fillOpacity="0.42"
      strokeWidth="0.7"
    />
  </Svg>
)
