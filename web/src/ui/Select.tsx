import * as React from 'react'
import {
  Select as AriaSelect,
  SelectValue,
  Button as AriaButton,
  Popover,
  ListBox,
  ListBoxItem,
  Label,
} from 'react-aria-components'
import type { Key } from 'react-aria-components'
import { cn } from './lib/cn'

export interface SelectOption {
  value: string
  label: string
}

export interface SelectProps {
  label?: string
  value?: string
  defaultValue?: string
  onChange?: (value: string) => void
  options: readonly SelectOption[]
  placeholder?: string
  isDisabled?: boolean
  className?: string
  'aria-label'?: string
}

export const Select: React.FC<SelectProps> = ({
  label,
  value,
  defaultValue,
  onChange,
  options,
  placeholder = 'Select…',
  isDisabled,
  className,
  'aria-label': ariaLabel,
}): React.ReactElement => {
  const handleChange = (key: Key | null): void => {
    if (key !== null) onChange?.(String(key))
  }

  return (
    <AriaSelect
      className={cn('flex flex-col gap-1.5', className)}
      isDisabled={isDisabled ?? false}
      {...(value === undefined ? {} : { selectedKey: value })}
      {...(defaultValue === undefined ? {} : { defaultSelectedKey: defaultValue })}
      {...(ariaLabel === undefined ? {} : { 'aria-label': ariaLabel })}
      onSelectionChange={handleChange}
    >
      {renderLabel(label)}
      <AriaButton className="hud-field flex items-center justify-between gap-2 px-3 py-2 font-mono text-[13px] text-foreground outline-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[focus-visible]:hud-glow">
        <SelectValue className="truncate data-[placeholder]:text-muted">
          {({ isPlaceholder, selectedText }) => (isPlaceholder ? placeholder : selectedText)}
        </SelectValue>
        <svg viewBox="0 0 16 16" className="h-3.5 w-3.5 shrink-0 text-muted" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
          <path d="m4 6 4 4 4-4" />
        </svg>
      </AriaButton>
      <Popover className="hud-panel z-[950] w-[--trigger-width] p-1 outline-none data-[entering]:opacity-0 data-[exiting]:opacity-0">
        <ListBox className="max-h-64 overflow-y-auto outline-none">
          {options.map((opt) => (
            <ListBoxItem
              key={opt.value}
              id={opt.value}
              textValue={opt.label}
              className="cursor-pointer px-3 py-1.5 font-mono text-[13px] text-foreground outline-none transition data-[focused]:bg-accent/15 data-[focused]:text-focus data-[selected]:text-accent"
            >
              {opt.label}
            </ListBoxItem>
          ))}
        </ListBox>
      </Popover>
    </AriaSelect>
  )
}

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}
