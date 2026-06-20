import * as React from 'react'
import { KPI, KPIGroup } from '@heroui-pro/react'
import type { ItemProps } from './types'

export const InfoCardItem: React.FC<ItemProps> = ({ label, value, valueColor }): React.ReactElement => {
  return (
    <React.Fragment>
      <KPI>
        <KPI.Header>
          <KPI.Title>{label}</KPI.Title>
        </KPI.Header>
        <KPI.Content>
          <span
            className="text-2xl font-semibold"
            style={valueColor ? { color: valueColor } : undefined}
          >
            {value}
          </span>
        </KPI.Content>
      </KPI>
      <KPIGroup.Separator />
    </React.Fragment>
  )
}
