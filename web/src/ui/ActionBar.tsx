import * as React from 'react'
import { cn } from './lib/cn'

export interface ActionBarProps {
  isOpen: boolean
  'aria-label'?: string
  className?: string
  children?: React.ReactNode
}

/**
 * Floating contextual action bar (shown on multi-select). Compound API matches
 * the previous HeroUI usage: <ActionBar isOpen><ActionBar.Prefix/> …
 * <ActionBar.Content/></ActionBar>.
 */
const ActionBarRoot: React.FC<ActionBarProps> = ({
  isOpen,
  'aria-label': ariaLabel,
  className,
  children,
}): React.ReactElement | null => {
  if (!isOpen) return null
  return (
    <div
      role="toolbar"
      aria-label={ariaLabel}
      className={cn(
        'hud-panel fixed bottom-6 left-1/2 z-[800] flex -translate-x-1/2 items-center gap-3 px-4 py-2.5 shadow-[0_8px_30px_-8px_rgba(0,0,0,0.6)]',
        className,
      )}
    >
      {children}
    </div>
  )
}

const Prefix: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex items-center gap-2', className)}>
    {children}
  </div>
)

const Content: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex items-center gap-2', className)}>
    {children}
  </div>
)

export const ActionBar = Object.assign(ActionBarRoot, { Prefix, Content, Suffix: Content })
