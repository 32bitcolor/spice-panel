import * as React from 'react'
import type { QualityArcProps } from './types'

// 5-segment arc ring quality indicator.
// Segments run clockwise from 12 o'clock; filled = gold, empty = dim.
const SEG_COUNT = 5
const SEG_DEG = 58 // degrees per arc — 14° gaps between each
const GAP_DEG = (360 - SEG_COUNT * SEG_DEG) / SEG_COUNT

function arcPath(cx: number, cy: number, r: number, i: number): string {
  const start = -90 + GAP_DEG / 2 + i * (SEG_DEG + GAP_DEG)
  const end = start + SEG_DEG
  const s = (start * Math.PI) / 180
  const e = (end * Math.PI) / 180
  const x1 = (cx + r * Math.cos(s)).toFixed(3)
  const y1 = (cy + r * Math.sin(s)).toFixed(3)
  const x2 = (cx + r * Math.cos(e)).toFixed(3)
  const y2 = (cy + r * Math.sin(e)).toFixed(3)
  return `M ${x1} ${y1} A ${r} ${r} 0 0 1 ${x2} ${y2}`
}

export const QualityArc: React.FC<QualityArcProps> = ({ quality, size = 20 }) => {
  const cx = size / 2
  const cy = size / 2
  const r = size * 0.34
  return (
    <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} aria-hidden>
      {Array.from({ length: SEG_COUNT }, (_, i) => (
        <path
          key={i}
          d={arcPath(cx, cy, r, i)}
          fill="none"
          stroke={i < quality ? '#d4a843' : 'rgba(255,255,255,0.12)'}
          strokeWidth={size * 0.115}
          strokeLinecap="butt"
        />
      ))}
    </svg>
  )
}
