import * as React from 'react'
import Svg from './Svg'
import type { IconProps } from './types'

// ── Fremen kindjal / crysknife ────────────────────────────────────────────────
// A slim curved blade tapering to a wicked point — the Fremen's signature weapon.
// The fuller groove and wrapped handle grip are hallmarks of the real prop design.
export const BladeIcon: React.FC<IconProps> = ({ className }): React.ReactElement => (
  <Svg className={className}>
    {/* blade body — diamond cross-section tapers from ricasso to tip */}
    <path
      d="M12 2 C12.6 5 13.2 9 13 14.5 L12 17.5 L11 14.5 C10.8 9 11.4 5 12 2 Z"
      fill="currentColor"
      fillOpacity="0.22"
      strokeWidth="1.3"
    />
    {/* right cutting edge */}
    <path d="M12 2 C12.6 5 13.2 9 13 14.5 L12 17.5" strokeWidth="1.3" />
    {/* fuller groove — dashed, runs 80% of blade length */}
    <line x1="11.4" y1="4.5" x2="11.4" y2="13.5" strokeWidth="0.65" strokeOpacity="0.45" strokeDasharray="2 1.5" />
    {/* crossguard — wider than the blade, slight angle */}
    <path d="M7.5 18 L16.5 18" strokeWidth="2.6" strokeLinecap="round" />
    {/* grip — wrapped handle with three bands */}
    <rect x="10.5" y="18.8" width="3" height="3.5" rx="0.6" fill="currentColor" fillOpacity="0.14" strokeWidth="1.15" />
    <line x1="10.5" y1="19.8" x2="13.5" y2="19.8" strokeWidth="0.65" strokeOpacity="0.5" />
    <line x1="10.5" y1="21.1" x2="13.5" y2="21.1" strokeWidth="0.65" strokeOpacity="0.5" />
  </Svg>
)
