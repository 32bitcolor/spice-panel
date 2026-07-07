import * as React from 'react'
import NumberFlow from '@number-flow/react'
import { cn } from './lib/cn'

type NumberFlowFormat = React.ComponentProps<typeof NumberFlow>['format']

export interface NumberValueProps extends Intl.NumberFormatOptions {
  value: number
  format?: NumberFlowFormat
  className?: string
}

/**
 * Animated, tabular numeric readout (replaces HeroUI's NumberValue). Accepts a
 * `format` object or inline Intl.NumberFormatOptions (maximumFractionDigits, …).
 */
export const NumberValue: React.FC<NumberValueProps> = ({
  value,
  format,
  className,
  ...opts
}): React.ReactElement => {
  const resolved = format ?? (Object.keys(opts).length > 0 ? (opts as NumberFlowFormat) : undefined)
  return (
    <NumberFlow
      value={value}
      {...(resolved === undefined ? {} : { format: resolved })}
      className={cn('font-mono tabular-nums', className)}
    />
  )
}
