import * as React from 'react'
import {
  ToggleButton as AriaToggleButton,
  ToggleButtonGroup as AriaToggleButtonGroup,
} from 'react-aria-components'
import type {
  ToggleButtonProps as AriaToggleButtonProps,
  ToggleButtonGroupProps as AriaToggleButtonGroupProps,
} from 'react-aria-components'
import { cn } from './lib/cn'

export interface ToggleButtonProps extends Omit<AriaToggleButtonProps, 'className'> {
  className?: string
  isIconOnly?: boolean
}

export const ToggleButton: React.FC<ToggleButtonProps> = ({
  className,
  isIconOnly = false,
  ...props
}): React.ReactElement => (
  <AriaToggleButton
    {...props}
    className={cn(
      'hud-plate-sm inline-flex cursor-pointer select-none items-center justify-center gap-2 font-mono text-xs uppercase tracking-[0.06em] outline-none transition data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[focus-visible]:hud-glow',
      isIconOnly ? 'size-8 p-0' : 'px-3.5 py-1.5',
      'bg-surface-secondary text-muted data-[hovered]:text-foreground',
      'data-[selected]:bg-accent data-[selected]:font-bold data-[selected]:text-accent-foreground',
      className,
    )}
  />
)

export interface ToggleButtonGroupProps extends Omit<AriaToggleButtonGroupProps, 'className'> {
  className?: string
  /** Accepted for HeroUI compatibility; styling is unified. */
  size?: 'sm' | 'md'
}

export const ToggleButtonGroup: React.FC<ToggleButtonGroupProps> = ({
  className,
  size: _size,
  ...props
}): React.ReactElement => (
  <AriaToggleButtonGroup
    {...props}
    className={cn('inline-flex gap-1 bg-[var(--void)] p-1 shadow-[inset_0_0_0_1px_var(--steel)]', className)}
  />
)
