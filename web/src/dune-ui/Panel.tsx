import type React from 'react'
import type { ReactNode } from 'react'

type PanelProps = {
  children: ReactNode
  className?: string
}

/**
 * Elevated bordered card. Use for content groups like the Progression Unlock
 * sub-panels in PlayerActionsModal.
 */
export const Panel: React.FC<PanelProps> = ({ children, className = '' }) => (
  <div
    className={
      'rounded-[var(--radius)] p-4 flex flex-col gap-2 '
      + 'bg-surface-secondary border border-border dune-lift '
      + className
    }
  >
    {children}
  </div>
)
