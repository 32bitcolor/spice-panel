import * as React from 'react'
import { Input as AriaInput } from 'react-aria-components'
import { cn } from './lib/cn'

export interface InputProps extends Omit<React.ComponentProps<typeof AriaInput>, 'className'> {
  className?: string
  fullWidth?: boolean
}

/**
 * Standalone HUD text input (styled RAC Input). Usable outside a TextField for
 * the many HeroUI-era `<Input value onChange />` call sites.
 */
export const Input: React.FC<InputProps> = ({
  className,
  fullWidth: _fullWidth,
  ...props
}): React.ReactElement => (
  <AriaInput
    {...props}
    className={cn(
      'hud-field w-full bg-transparent px-3 py-2 font-mono text-[13px] text-foreground outline-none placeholder:text-muted/70 disabled:opacity-40',
      className,
    )}
  />
)
