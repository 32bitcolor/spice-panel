import * as React from 'react'
import { cn } from './lib/cn'

export interface LabelProps extends React.HTMLAttributes<HTMLSpanElement> {}

/** HUD section label: uppercase, mono, letter-tracked. Presentational only. */
export const Label: React.FC<LabelProps> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <span
    {...props}
    className={cn('font-mono text-[11px] uppercase tracking-[0.22em] text-muted', className)}
  >
    {children}
  </span>
)
