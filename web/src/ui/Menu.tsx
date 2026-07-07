import * as React from 'react'
import {
  MenuTrigger as AriaMenuTrigger,
  Menu as AriaMenu,
  MenuItem as AriaMenuItem,
  Popover,
} from 'react-aria-components'
import type { MenuItemProps as AriaMenuItemProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface MenuProps {
  /** The trigger element (a Button). */
  trigger: React.ReactElement
  children: React.ReactNode
  className?: string
}

/** Dropdown menu. Compose with <MenuItem>. Replaces HeroUI's Dropdown. */
export const Menu: React.FC<MenuProps> = ({ trigger, children, className }): React.ReactElement => (
  <AriaMenuTrigger>
    {trigger}
    <Popover className="hud-panel z-[950] min-w-[10rem] p-1 outline-none data-[entering]:opacity-0 data-[exiting]:opacity-0">
      <AriaMenu className={cn('outline-none', className)}>{children}</AriaMenu>
    </Popover>
  </AriaMenuTrigger>
)

export interface MenuItemProps extends Omit<AriaMenuItemProps, 'className'> {
  className?: string
  danger?: boolean
}

export const MenuItem: React.FC<MenuItemProps> = ({
  className,
  danger = false,
  ...props
}): React.ReactElement => (
  <AriaMenuItem
    {...props}
    className={cn(
      'flex cursor-pointer items-center gap-2 px-3 py-1.5 text-[13px] outline-none transition data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[focused]:bg-accent/15',
      danger ? 'text-danger data-[focused]:text-danger' : 'text-foreground data-[focused]:text-focus',
      className,
    )}
  />
)
