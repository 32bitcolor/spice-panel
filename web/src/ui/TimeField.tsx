import * as React from 'react'
import { TimeField as AriaTimeField, DateInput, DateSegment, Label } from 'react-aria-components'
import type { TimeValue } from 'react-aria-components'
import { parseTime } from '@internationalized/date'
import { cn } from './lib/cn'

export interface TimeFieldProps {
  'label'?: string
  /** 24-hour "HH:MM" string. */
  'value'?: string
  'onChange'?: (value: string) => void
  'className'?: string
  'aria-label'?: string
}

const pad = (n: number): string => String(n).padStart(2, '0')

const toValue = (value: string | undefined): TimeValue | null => {
  if (value === undefined || value === '') return null
  try {
    return parseTime(value)
  }
  catch {
    return null
  }
}

export const TimeField: React.FC<TimeFieldProps> = ({
  label,
  value,
  onChange,
  className,
  'aria-label': ariaLabel,
}): React.ReactElement => {
  const handleChange = (next: TimeValue | null): void => {
    onChange?.(next === null ? '' : `${pad(next.hour)}:${pad(next.minute)}`)
  }

  return (
    <AriaTimeField
      hourCycle={24}
      value={toValue(value)}
      onChange={handleChange}
      className={cn('flex flex-col gap-1.5', className)}
      {...(ariaLabel === undefined ? {} : { 'aria-label': ariaLabel })}
    >
      {renderLabel(label)}
      <DateInput className="hud-field flex items-center gap-0.5 px-3 py-2 font-mono text-[13px] text-foreground data-[focus-within]:hud-glow">
        {(segment) => (
          <DateSegment
            segment={segment}
            className="rounded-[2px] px-0.5 tabular-nums outline-none data-[focused]:bg-accent/25 data-[focused]:text-focus data-[placeholder]:text-muted"
          />
        )}
      </DateInput>
    </AriaTimeField>
  )
}

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}
