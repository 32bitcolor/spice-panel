import * as React from 'react'
import { cn } from '../ui'
import type { CardProps, ItemProps } from './types'
import { InfoCardItem } from './InfoCardItem'

/**
 * Bordered, slightly-elevated label/value row card — the "Phase Reconciling
 * | Database Ready" health row pattern. The InfoCard / InfoCard.Item API is
 * preserved so existing call sites need no changes.
 */
export const InfoCard: React.FC<CardProps> & { Item: React.FC<ItemProps> } = ({
  children,
  className = '',
}): React.ReactElement => (
  <div
    className={cn(
      'flex flex-wrap items-stretch gap-x-6 gap-y-3 bg-surface-secondary px-4 py-3 ring-1 ring-inset ring-border [border-radius:var(--radius)]',
      className,
    )}
  >
    {children}
  </div>
)

// Namespace alias kept for callers using <InfoCard.Item>
InfoCard.Item = InfoCardItem
