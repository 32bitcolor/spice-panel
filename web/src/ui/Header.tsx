import * as React from 'react'
import { Header as AriaHeader } from 'react-aria-components'
import { cn } from './lib/cn'

export interface HeaderProps extends React.ComponentProps<typeof AriaHeader> {
  className?: string
}

/** Section header for grouped lists/menus (styled RAC Header). */
export const Header: React.FC<HeaderProps> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaHeader
    {...props}
    className={cn(
      'px-3 py-1.5 font-mono text-[10.5px] uppercase tracking-[0.18em] text-muted',
      className as string,
    )}
  >
    {children}
  </AriaHeader>
)
