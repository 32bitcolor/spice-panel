import * as React from 'react'
import { cn } from './lib/cn'

export interface PanelProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Remove the default inner padding (for tables / edge-to-edge content). */
  flush?: boolean
}

export const Panel: React.FC<PanelProps> = ({
  flush = false,
  className,
  children,
  ...props
}): React.ReactElement => (
  <div
    {...props}
    className={cn('hud-panel text-foreground', flush ? 'p-0' : 'p-4', className)}
  >
    {children}
  </div>
)
