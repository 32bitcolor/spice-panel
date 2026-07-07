import * as React from 'react'
import { cn } from './lib/cn'

export interface SeparatorProps {
  orientation?: 'horizontal' | 'vertical'
  className?: string
}

export const Separator: React.FC<SeparatorProps> = ({
  orientation = 'horizontal',
  className,
}): React.ReactElement => (
  <div
    role="separator"
    aria-orientation={orientation}
    className={cn(
      orientation === 'horizontal'
        ? 'h-px w-full bg-gradient-to-r from-transparent via-border to-transparent'
        : 'h-full w-px bg-gradient-to-b from-transparent via-border to-transparent',
      className,
    )}
  />
)
