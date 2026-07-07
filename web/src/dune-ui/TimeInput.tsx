import * as React from 'react'
import { Time } from '@internationalized/date'
import { TimeField as AriaTimeField, DateInput, DateSegment } from 'react-aria-components'
import type { TimeValue } from 'react-aria-components'
import { ToggleButton, ToggleButtonGroup, cn } from '../ui'
import type { TimeInputProps } from './types'

const parseHHMM = (s: string): Time | null => {
  const parts = s.split(':')
  const h = Number(parts[0])
  const m = Number(parts[1])
  if (Number.isNaN(h) || Number.isNaN(m)) return null
  return new Time(h, m)
}

const toHHMM = (t: Time): string =>
  `${String(t.hour).padStart(2, '0')}:${String(t.minute).padStart(2, '0')}`

export const TimeInput: React.FC<TimeInputProps> = ({
  value,
  onChange,
  ariaLabel,
  className,
  isDisabled,
}): React.ReactElement => {
  const timeValue = parseHHMM(value)
  const isAM = timeValue ? timeValue.hour < 12 : true

  const handleTimeChange = (t: TimeValue | null): void => {
    if (t) onChange(toHHMM(new Time(t.hour, t.minute)))
  }

  const handlePeriodChange = (keys: 'all' | Set<React.Key>): void => {
    if (!timeValue) return
    const period = keys === 'all' ? null : [...keys][0]
    if (!period) return
    let { hour } = timeValue
    const { minute } = timeValue
    if (period === 'pm' && hour < 12) hour += 12
    else if (period === 'am' && hour >= 12) hour -= 12
    onChange(toHHMM(new Time(hour, minute)))
  }

  return (
    <div className={cn('flex items-center gap-1', className)}>
      <AriaTimeField
        value={timeValue}
        onChange={handleTimeChange}
        hourCycle={12}
        granularity="minute"
        {...(ariaLabel !== undefined ? { 'aria-label': ariaLabel } : {})}
        {...(isDisabled !== undefined ? { isDisabled } : {})}
      >
        <DateInput className="hud-field flex items-center gap-0.5 px-3 py-2 font-mono text-[13px] text-foreground data-[focus-within]:hud-glow">
          {(segment) => (
            <DateSegment
              segment={segment}
              className={cn(
                'rounded-[2px] px-0.5 tabular-nums outline-none data-[focused]:bg-accent/25 data-[focused]:text-focus data-[placeholder]:text-muted',
                segment.type === 'dayPeriod' ? 'hidden' : '',
              )}
            />
          )}
        </DateInput>
      </AriaTimeField>
      <ToggleButtonGroup
        selectionMode="single"
        disallowEmptySelection
        selectedKeys={new Set([isAM ? 'am' : 'pm'])}
        onSelectionChange={handlePeriodChange}
        {...(isDisabled !== undefined ? { isDisabled } : {})}
      >
        <ToggleButton id="am" className="px-2.5 py-1">
          AM
        </ToggleButton>
        <ToggleButton id="pm" className="px-2.5 py-1">
          PM
        </ToggleButton>
      </ToggleButtonGroup>
    </div>
  )
}
