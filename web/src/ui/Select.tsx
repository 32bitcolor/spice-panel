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
  /** Simple API. */
  value?: string
  defaultValue?: string
  onChange?: (value: string) => void
  options?: readonly SelectOption[]
  /** HeroUI/RAC-style controlled API (used by compound call sites). */
  selectedKey?: string | null
  onSelectionChange?: (key: Key) => void
  placeholder?: string
  isDisabled?: boolean
  className?: string
  children?: React.ReactNode
  'aria-label'?: string
}

const TRIGGER_CLS =
  'hud-field flex items-center justify-between gap-2 px-3 py-2 font-mono text-[13px] text-foreground outline-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40 data-[focus-visible]:hud-glow'

const SelectRoot: React.FC<SelectProps> = ({
  label,
  value,
  defaultValue,
  onChange,
  options,
  selectedKey,
  onSelectionChange,
  placeholder = 'Select…',
  isDisabled,
  className,
  children,
  'aria-label': ariaLabel,
}): React.ReactElement => {
  const controlled = value ?? selectedKey ?? undefined
  const handleChange = (key: Key | null): void => {
    if (key === null) return
    onChange?.(String(key))
    onSelectionChange?.(key)
  }

  const common = {
    className: cn('flex flex-col gap-1.5', className),
    isDisabled: isDisabled ?? false,
    onSelectionChange: handleChange,
    ...(controlled === undefined ? {} : { selectedKey: controlled }),
    ...(defaultValue === undefined ? {} : { defaultSelectedKey: defaultValue }),
    ...(ariaLabel === undefined ? {} : { 'aria-label': ariaLabel }),
  }

  // Compound mode: caller supplies Trigger/Popover children.
  if (children !== undefined) {
    return (
      <AriaSelect {...common}>
        {renderLabel(label)}
        {children}
      </AriaSelect>
    )
  }

  // Simple mode: render trigger + popover from options.
  return (
    <AriaSelect {...common}>
      {renderLabel(label)}
      <AriaButton className={TRIGGER_CLS}>
        <SelectValue className="truncate data-[placeholder]:text-muted">
          {({ isPlaceholder, selectedText }) => (isPlaceholder ? placeholder : selectedText)}
        </SelectValue>
        <Indicator />
      </AriaButton>
      <SelectPopover>
        <ListBox className="max-h-64 overflow-y-auto outline-none">
          {(options ?? []).map((opt) => (
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
      </SelectPopover>
    </AriaSelect>
  )
}

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}

/* ── Compound slots (HeroUI-compatible) ───────────────────────────────────── */

const Trigger: React.FC<React.HTMLAttributes<HTMLButtonElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <AriaButton className={cn(TRIGGER_CLS, className)}>{children}</AriaButton>
)

const Value: React.FC<{ placeholder?: string; className?: string }> = ({
  placeholder,
  className,
}): React.ReactElement => (
  <SelectValue className={cn('truncate data-[placeholder]:text-muted', className)}>
    {({ isPlaceholder, selectedText }) =>
      isPlaceholder ? (placeholder ?? 'Select…') : selectedText
    }
  </SelectValue>
)

const Indicator: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <svg viewBox="0 0 16 16" className={cn('h-3.5 w-3.5 shrink-0 text-muted', className)} fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
    <path d="m4 6 4 4 4-4" />
  </svg>
)

const SelectPopover: React.FC<React.PropsWithChildren<{ className?: string }>> = ({
  className,
  children,
}): React.ReactElement => (
  <Popover className={cn('hud-panel z-[950] w-[--trigger-width] p-1 outline-none data-[entering]:opacity-0 data-[exiting]:opacity-0', className)}>
    {children}
  </Popover>
)

export const Select = Object.assign(SelectRoot, {
  Trigger,
  Value,
  Indicator,
  Popover: SelectPopover,
})
