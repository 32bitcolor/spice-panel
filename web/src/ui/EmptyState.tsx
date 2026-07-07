import * as React from 'react'
import { tv } from 'tailwind-variants'
import type { VariantProps } from 'tailwind-variants'
import { cn } from './lib/cn'

const rootStyles = tv({
  base: 'flex flex-col items-center justify-center text-center',
  variants: {
    size: {
      sm: 'gap-2 px-4 py-8',
      md: 'gap-3 px-6 py-12',
      lg: 'gap-4 px-8 py-16',
    },
  },
  defaultVariants: { size: 'md' },
})

export type EmptyStateVariants = VariantProps<typeof rootStyles>

export interface EmptyStateProps
  extends Omit<React.HTMLAttributes<HTMLDivElement>, 'title'>,
    EmptyStateVariants {}

const EmptyStateRoot: React.FC<EmptyStateProps> = ({
  size,
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn(rootStyles({ size }), className)}>
    {children}
  </div>
)

const Media: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div
    {...props}
    className={cn(
      'grid size-14 place-items-center bg-accent/8 text-accent shadow-[inset_0_0_0_1px_var(--steel)] [clip-path:polygon(9px_0,100%_0,100%_calc(100%-9px),calc(100%-9px)_100%,0_100%,0_9px)]',
      className,
    )}
  >
    {children}
  </div>
)

const Header: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex flex-col items-center gap-1', className)}>
    {children}
  </div>
)

const Title: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <p {...props} className={cn('text-sm font-medium tracking-[0.02em] text-foreground', className)}>
    {children}
  </p>
)

const Description: React.FC<React.HTMLAttributes<HTMLParagraphElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <p {...props} className={cn('max-w-xs text-[13px] text-muted', className)}>
    {children}
  </p>
)

const Actions: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('mt-1 flex items-center gap-2.5', className)}>
    {children}
  </div>
)

export const EmptyState = Object.assign(EmptyStateRoot, {
  Media,
  Header,
  Title,
  Description,
  Actions,
})
