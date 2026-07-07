import * as React from 'react'
import type { ItemProps } from './types'

export const InfoCardItem: React.FC<ItemProps> = ({
  label,
  value,
  valueColor,
}): React.ReactElement => (
  <div className="flex min-w-0 flex-col gap-0.5">
    <span className="font-mono text-[10.5px] uppercase tracking-[0.18em] text-muted">{label}</span>
    <span
      className="text-2xl font-semibold text-foreground"
      {...(valueColor !== undefined ? { style: { color: valueColor } } : {})}
    >
      {value}
    </span>
  </div>
)
