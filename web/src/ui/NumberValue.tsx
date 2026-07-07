import * as React from 'react'
import NumberFlow from '@number-flow/react'
import { cn } from './lib/cn'

type NumberFlowFormat = React.ComponentProps<typeof NumberFlow>['format']

export interface NumberValueProps {
  value: number
  format?: NumberFlowFormat
  className?: string
}

/** Animated, tabular numeric readout (replaces HeroUI's NumberValue). */
export const NumberValue: React.FC<NumberValueProps> = ({
  value,
  format,
  className,
}): React.ReactElement => (
  <NumberFlow
    value={value}
    {...(format === undefined ? {} : { format })}
    className={cn('font-mono tabular-nums', className)}
  />
)
