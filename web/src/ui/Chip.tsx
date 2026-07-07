import * as React from 'react'
import { tv } from 'tailwind-variants'
import type { VariantProps } from 'tailwind-variants'
import { cn } from './lib/cn'

export const chipStyles = tv({
  base: 'inline-flex items-center gap-1.5 font-mono text-[11px] uppercase tracking-[0.08em] [clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)]',
  variants: {
    color: {
      accent: 'bg-accent/15 text-accent',
      success: 'bg-success/15 text-success',
      warning: 'bg-warning/15 text-warning',
      danger: 'bg-danger/15 text-danger',
      muted: 'bg-muted/15 text-muted',
    },
    size: {
      sm: 'px-2 py-0.5',
      md: 'px-2.5 py-1',
    },
    dot: { true: '', false: '' },
  },
  compoundVariants: [
    {
      dot: true,
      class:
        'before:h-1.5 before:w-1.5 before:rounded-full before:bg-current before:shadow-[0_0_6px_currentColor]',
    },
  ],
  defaultVariants: { color: 'muted', size: 'sm', dot: false },
})

export type ChipVariants = VariantProps<typeof chipStyles>

export interface ChipProps extends Omit<React.HTMLAttributes<HTMLSpanElement>, 'color'>, ChipVariants {}

export const Chip: React.FC<ChipProps> = ({
  color,
  size,
  dot,
  className,
  children,
  ...props
}): React.ReactElement => (
  <span {...props} className={cn(chipStyles({ color, size, dot }), className)}>
    {children}
  </span>
)
