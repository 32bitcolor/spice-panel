import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Schematic fragment — T6 crafting blueprint shard ─────────────────────────
// A jagged hexagonal shard that reads as "a broken-off piece of something larger."
// Blueprint circuit traces inside hint at the technology locked within.
export const FragmentIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* irregular shard — hexagonal with one fractured edge at top-left */}
    <path
      d="M14.5 2 L21 7.5 L19 15.5 L13 18.5 L7 15 L6 7 Z"
      fill="currentColor"
      fillOpacity="0.15"
      strokeWidth="1.4"
    />
    {/* fracture break — dashed line suggesting it was cleaved from a larger piece */}
    <path d="M6 7 L14.5 2" strokeWidth="0.7" strokeOpacity="0.38" strokeDasharray="2 1.5" />
    {/* circuit trace — horizontal run with a node */}
    <path d="M9 10 L12.5 10 L12.5 13" strokeWidth="1" strokeOpacity="0.55" fill="none" />
    <path d="M12.5 13 L16 13" strokeWidth="1" strokeOpacity="0.55" fill="none" />
    {/* junction node */}
    <circle cx="12.5" cy="10" r="1" fill="currentColor" fillOpacity="0.5" strokeWidth="0" />
  </Svg>
)
