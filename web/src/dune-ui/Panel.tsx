import * as React from 'react'
import { cn } from '../ui'
import type { PanelProps } from './types'

export const Panel: React.FC<PanelProps> = ({
  children,
  className = '',
  contentClassName = '',
}): React.ReactElement => (
  <div className={cn('hud-panel dune-panel', className)}>
    <div className={cn('flex flex-col gap-2 p-8', contentClassName)}>{children}</div>
  </div>
)
