import * as React from 'react'
import { TextField as AriaTextField, Input } from 'react-aria-components'
import { cn } from '../ui'
import type { FieldInputProps } from './interfaces'

export const FieldInput: React.FC<FieldInputProps> = ({
  value,
  onChange,
  placeholder,
  type = 'text',
  className,
  ariaLabel,
  isDisabled,
}): React.ReactElement => (
  <AriaTextField
    value={value}
    onChange={onChange}
    {...(ariaLabel !== undefined ? { 'aria-label': ariaLabel } : {})}
    {...(isDisabled !== undefined ? { isDisabled } : {})}
  >
    <Input
      type={type}
      {...(placeholder !== undefined ? { placeholder } : {})}
      className={cn(
        'hud-field w-full bg-transparent px-3 py-2 font-mono text-[13px] text-foreground outline-none placeholder:text-muted/70 disabled:opacity-40',
        className,
      )}
    />
  </AriaTextField>
)
