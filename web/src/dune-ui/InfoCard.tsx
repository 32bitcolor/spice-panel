import * as React from 'react'
import { KPIGroup } from '@heroui-pro/react'
import type { CardProps, ItemProps } from './types'
import { InfoCardItem } from './InfoCardItem'

/**
 * Bordered, slightly-elevated label/value row card — the "Phase Reconciling
 * | Database Ready" health row pattern from BattlegroupTab.
 *
 * Backed by KPIGroup + KPI internally; the InfoCard / InfoCard.Item API is
 * preserved so existing call sites need no changes.
 */
export const InfoCard: React.FC<CardProps> & { Item: React.FC<ItemProps> } = ({ children, className = '' }): React.ReactElement => {
  return (
    <KPIGroup className={`flex-wrap ${className}`} orientation="horizontal">
      {children}
    </KPIGroup>
  )
}

// Namespace alias kept for callers using <InfoCard.Item>
InfoCard.Item = InfoCardItem
