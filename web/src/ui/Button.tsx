import * as React from 'react'
import { Button as AriaButton } from 'react-aria-components'
import type { ButtonProps as AriaButtonProps } from 'react-aria-components'
import { tv } from 'tailwind-variants'
import type { VariantProps } from 'tailwind-variants'
import { cn } from './lib/cn'

export const buttonStyles = tv({
  base: 'hud-plate-sm relative inline-flex cursor-pointer select-none items-center justify-center gap-2 font-semibold tracking-wide outline-none transition duration-150 data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[pressed]:translate-y-px data-[focus-visible]:hud-glow',
  variants: {
    variant: {
      primary:
        'bg-gradient-to-b from-focus to-accent font-bold text-accent-foreground data-[hovered]:hud-glow',
      solid: 'bg-surface-secondary text-foreground data-[hovered]:brightness-125',
      ghost:
        'bg-transparent text-foreground ring-1 ring-inset ring-border data-[hovered]:text-focus data-[hovered]:ring-accent',
      danger:
        'bg-danger/20 text-foreground ring-1 ring-inset ring-danger/50 data-[hovered]:bg-danger/30',
    },
    size: {
      sm: 'px-3 py-1.5 text-xs',
      md: 'px-[18px] py-[9px] text-[13px]',
    },
  },
  defaultVariants: { variant: 'solid', size: 'md' },
})

export type ButtonVariants = VariantProps<typeof buttonStyles>

export interface ButtonProps extends Omit<AriaButtonProps, 'className'>, ButtonVariants {
  className?: string
}

export const Button: React.FC<ButtonProps> = ({
  variant,
  size,
  className,
  ...props
}): React.ReactElement => (
  <AriaButton {...props} className={cn(buttonStyles({ variant, size }), className)} />
)
