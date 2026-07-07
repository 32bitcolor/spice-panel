import * as React from 'react'
import { Button as AriaButton } from 'react-aria-components'
import type { ButtonProps as AriaButtonProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface CloseButtonProps extends Omit<AriaButtonProps, 'className' | 'children'> {
  className?: string
  size?: number
}

export const CloseButton: React.FC<CloseButtonProps> = ({
  className,
  size = 16,
  ...props
}): React.ReactElement => (
  <AriaButton
    aria-label="Close"
    {...props}
    className={cn(
      'grid place-items-center text-muted outline-none transition hover:text-foreground data-[focus-visible]:hud-glow',
      className,
    )}
  >
    <svg width={size} height={size} viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <path d="m4 4 8 8M12 4l-8 8" />
    </svg>
  </AriaButton>
)
