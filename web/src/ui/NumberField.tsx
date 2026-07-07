import * as React from 'react'
import {
  NumberField as AriaNumberField,
  Label,
  Group,
  Input,
  Button as AriaButton,
} from 'react-aria-components'
import type { NumberFieldProps as AriaNumberFieldProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface NumberFieldProps extends Omit<AriaNumberFieldProps, 'className' | 'children'> {
  label?: string
  className?: string
}

export const NumberField: React.FC<NumberFieldProps> = ({
  label,
  className,
  ...props
}): React.ReactElement => (
  <AriaNumberField {...props} className={cn('flex flex-col gap-1.5', className)}>
    {renderLabel(label)}
    <Group className="hud-field flex items-stretch data-[focus-within]:hud-glow">
      <Input className="w-full min-w-0 bg-transparent px-3 py-2 font-mono text-[13px] text-foreground outline-none" />
      <div className="flex flex-col border-l border-border">
        {renderStepper('increment', 'M4 10 8 6l4 4')}
        <span className="h-px bg-border" />
        {renderStepper('decrement', 'M4 6 8 10l4-4')}
      </div>
    </Group>
  </AriaNumberField>
)

const renderStepper = (slot: 'increment' | 'decrement', path: string): React.ReactElement => (
  <AriaButton
    slot={slot}
    className="grid flex-1 place-items-center px-2 text-muted outline-none transition hover:bg-accent/15 hover:text-focus"
  >
    <svg viewBox="0 0 16 16" className="h-3 w-3" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d={path} />
    </svg>
  </AriaButton>
)

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}
