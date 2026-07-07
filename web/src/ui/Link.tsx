import * as React from 'react'
import { Link as AriaLink } from 'react-aria-components'
import type { LinkProps as AriaLinkProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface LinkProps extends Omit<AriaLinkProps, 'className'> {
  className?: string
}

export const Link: React.FC<LinkProps> = ({ className, ...props }): React.ReactElement => (
  <AriaLink
    {...props}
    className={cn(
      'cursor-pointer text-link underline-offset-2 outline-none transition hover:underline data-[focus-visible]:hud-glow',
      className,
    )}
  />
)
