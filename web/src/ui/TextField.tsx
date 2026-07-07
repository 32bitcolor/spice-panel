import * as React from 'react'
import { TextField as AriaTextField, Label, Input } from 'react-aria-components'
import type { TextFieldProps as AriaTextFieldProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface TextFieldProps extends Omit<AriaTextFieldProps, 'className' | 'children'> {
  label?: string
  placeholder?: string
  className?: string
  inputClassName?: string
  /** input type; TextField supports text-like types (text, search, url, tel, email, password). */
  type?: 'text' | 'search' | 'url' | 'tel' | 'email' | 'password'
  /** Compose your own field content (Label/Input/etc). Omit for the default input. */
  children?: React.ReactNode
}

export const TextField: React.FC<TextFieldProps> = ({
  label,
  placeholder,
  className,
  inputClassName,
  type = 'text',
  children,
  ...props
}): React.ReactElement => (
  <AriaTextField {...props} className={cn('flex flex-col gap-1.5', className)}>
    {children ?? (
      <React.Fragment>
        {renderLabel(label)}
        <Input
          type={type}
          {...(placeholder === undefined ? {} : { placeholder })}
          className={cn(
            'hud-field w-full bg-transparent px-3 py-2 font-mono text-[13px] text-foreground outline-none placeholder:text-muted/70',
            inputClassName,
          )}
        />
      </React.Fragment>
    )}
  </AriaTextField>
)

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">
      {label}
    </Label>
  )
}
