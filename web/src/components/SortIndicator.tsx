import type React from 'react'
import type { SortDir } from '../hooks/useTableSort'

interface SortIndicatorProps {
  active: boolean
  dir: SortDir
}

export const SortIndicator: React.FC<SortIndicatorProps> = ({ active, dir }) => {
  return (
    <span style={{ marginLeft: 4, opacity: active ? 1 : 0.25 }}>
      {active ? (dir === 'asc' ? '▲' : '▼') : '▲'}
    </span>
  )
}
