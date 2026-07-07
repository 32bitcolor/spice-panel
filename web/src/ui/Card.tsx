import * as React from 'react'
import { cn } from './lib/cn'

export interface CardProps extends React.HTMLAttributes<HTMLDivElement> {
  /** Optional eyebrow label rendered in the header. */
  title?: string
  /** Optional trailing header content (actions, chips). */
  action?: React.ReactNode
}

/**
 * A lighter surface than Panel: a raised, softly-cornered container for grouped
 * content. Panel is the primary chamfered HUD plate; Card is the inner grouping.
 */
export const Card: React.FC<CardProps> = ({
  title,
  action,
  className,
  children,
  ...props
}): React.ReactElement => (
  <div
    {...props}
    className={cn(
      'bg-surface-secondary ring-1 ring-inset ring-border [border-radius:var(--radius)]',
      className,
    )}
  >
    {renderHeader(title, action)}
    <div className="p-4">{children}</div>
  </div>
)

const renderHeader = (title: string | undefined, action: React.ReactNode): React.ReactNode => {
  if (title === undefined && action === undefined) return null
  return (
    <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-2.5">
      <span className="font-mono text-[11px] uppercase tracking-[0.18em] text-muted">{title}</span>
      {action}
    </div>
  )
}
