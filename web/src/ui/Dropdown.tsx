import * as React from 'react'
import {
  MenuTrigger as AriaMenuTrigger,
  Menu as AriaMenu,
  MenuItem as AriaMenuItem,
  Popover,
} from 'react-aria-components'
import type { MenuProps as AriaMenuProps, MenuItemProps as AriaMenuItemProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface DropdownProps {
  children?: React.ReactNode
}

/** HeroUI-compatible Dropdown: <Dropdown><Button/><Dropdown.Popover><Dropdown.Menu>… */
const DropdownRoot: React.FC<DropdownProps> = ({ children }): React.ReactElement => (
  <AriaMenuTrigger>{children}</AriaMenuTrigger>
)

const DropdownPopover: React.FC<React.PropsWithChildren<{ className?: string }>> = ({
  className,
  children,
}): React.ReactElement => (
  <Popover className={cn('hud-panel z-[950] min-w-[10rem] p-1 outline-none data-[entering]:opacity-0 data-[exiting]:opacity-0', className)}>
    {children}
  </Popover>
)

const DropdownMenu = <T extends object>({
  className,
  ...props
}: AriaMenuProps<T> & { className?: string }): React.ReactElement => (
  <AriaMenu {...props} className={cn('outline-none', className)} />
)

export interface DropdownItemProps extends Omit<AriaMenuItemProps, 'className'> {
  className?: string
  danger?: boolean
}

const DropdownItem: React.FC<DropdownItemProps> = ({
  className,
  danger = false,
  ...props
}): React.ReactElement => (
  <AriaMenuItem
    {...props}
    className={cn(
      'flex cursor-pointer items-center gap-2 px-3 py-1.5 text-[13px] outline-none transition data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[focused]:bg-accent/15',
      danger ? 'text-danger' : 'text-foreground data-[focused]:text-focus',
      className,
    )}
  />
)

export const Dropdown = Object.assign(DropdownRoot, {
  Popover: DropdownPopover,
  Menu: DropdownMenu,
  Item: DropdownItem,
})
