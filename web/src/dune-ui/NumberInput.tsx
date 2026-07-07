import * as React from 'react'
import { NumberField as AriaNumberField, Label, Group, Input, Button as AriaButton } from 'react-aria-components'
import { cn } from '../ui'
import type { NumberInputProps } from './interfaces'

const stepBtn =
  'grid flex-1 place-items-center px-2 text-muted outline-none transition hover:bg-accent/15 hover:text-focus'

export const NumberInput: React.FC<NumberInputProps> = ({
  value,
  onChange,
  min,
  max,
  step = 1,
  label,
  prefix,
  ariaLabel,
  isDisabled,
  className,
  showButtons = true,
  formatOptions,
}): React.ReactElement => {
  const field = (
    <AriaNumberField
      value={value}
      onChange={(v) => onChange(Number.isNaN(v) ? (min ?? 0) : v)}
      step={step}
      aria-label={ariaLabel ?? label ?? prefix ?? ''}
      className={cn('flex flex-col gap-1.5', prefix ? 'min-w-0 flex-1' : className)}
      {...(min !== undefined ? { minValue: min } : {})}
      {...(max !== undefined ? { maxValue: max } : {})}
      {...(isDisabled !== undefined ? { isDisabled } : {})}
      {...(formatOptions !== undefined ? { formatOptions } : {})}
    >
      {renderLabel(label)}
      <Group className="hud-field flex items-stretch data-[focus-within]:hud-glow">
        {renderStep(showButtons, 'decrement', 'M4 6 8 10l4-4')}
        <Input className="w-full min-w-0 bg-transparent px-3 py-2 font-mono text-[13px] text-foreground outline-none" />
        {renderStep(showButtons, 'increment', 'M4 10 8 6l4 4')}
      </Group>
    </AriaNumberField>
  )

  if (prefix === undefined) return field

  return (
    <div className={cn('flex items-stretch', className)}>
      <span className="flex shrink-0 items-center border border-r-0 border-border bg-surface-secondary px-2 text-xs text-muted [border-radius:var(--radius)_0_0_var(--radius)]">
        {prefix}
      </span>
      {field}
    </div>
  )
}

const renderStep = (
  show: boolean,
  slot: 'increment' | 'decrement',
  path: string,
): React.ReactNode => {
  if (!show) return null
  return (
    <AriaButton slot={slot} className={cn(stepBtn, slot === 'increment' ? 'border-l border-border' : 'border-r border-border')}>
      <svg viewBox="0 0 16 16" className="h-3 w-3" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
        <path d={path} />
      </svg>
    </AriaButton>
  )
}

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return <Label className="text-xs text-muted">{label}</Label>
}
