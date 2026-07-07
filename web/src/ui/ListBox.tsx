import * as React from 'react'
import { ListBox as AriaListBox, ListBoxItem as AriaListBoxItem } from 'react-aria-components'
import type {
  ListBoxProps as AriaListBoxProps,
  ListBoxItemProps as AriaListBoxItemProps,
} from 'react-aria-components'
import { cn } from './lib/cn'

export interface ListBoxProps<T extends object> extends Omit<AriaListBoxProps<T>, 'className'> {
  className?: string
}

const ListBoxRoot = <T extends object>({
  className,
  ...props
}: ListBoxProps<T>): React.ReactElement => (
  <AriaListBox {...props} className={cn('max-h-64 overflow-y-auto outline-none', className)} />
)

export interface ListBoxItemProps extends Omit<AriaListBoxItemProps, 'className'> {
  className?: string
}

const Item: React.FC<ListBoxItemProps> = ({ className, ...props }): React.ReactElement => (
  <AriaListBoxItem
    {...props}
    className={cn(
      'flex cursor-pointer items-center justify-between gap-2 px-3 py-1.5 font-mono text-[13px] text-foreground outline-none transition data-[focused]:bg-accent/15 data-[focused]:text-focus data-[selected]:text-accent',
      className,
    )}
  />
)

/** Check indicator shown on the selected item (HeroUI-compatible slot). */
const ItemIndicator: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <svg
    viewBox="0 0 16 16"
    className={cn('h-3.5 w-3.5 shrink-0 opacity-0 group-data-[selected]:opacity-100', className)}
    fill="none"
    stroke="currentColor"
    strokeWidth="2"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="m3 8 3.5 3.5L13 4" />
  </svg>
)

export const ListBox = Object.assign(ListBoxRoot, { Item, ItemIndicator })
