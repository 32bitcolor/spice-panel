import * as React from 'react'
import { cn } from './lib/cn'

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {}

/**
 * A lighter surface than Panel: a raised, softly-cornered container for grouped
 * content. Compound API (Card.Header / Card.Title / Card.Content) matches the
 * previous HeroUI usage so call sites migrate unchanged.
 */
const CardRoot: React.FC<CardProps> = ({ className, children, ...props }): React.ReactElement => (
  <div
    {...props}
    className={cn(
      'bg-surface-secondary ring-1 ring-inset ring-border [border-radius:var(--radius)]',
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
  <div
    {...props}
    className={cn('flex items-center justify-between gap-3 border-b border-border px-4 py-2.5', className)}
  >
    {children}
  </div>
)

const Title: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <h3 {...props} className={cn('text-sm font-semibold text-foreground', className)}>
    {children}
  </h3>
)

const Content: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('p-4', className)}>
    {children}
  </div>
)

export const Card = Object.assign(CardRoot, { Header, Title, Content })
