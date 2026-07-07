import * as React from 'react'
import { TextField as AriaTextField, Label, TextArea as AriaTextArea } from 'react-aria-components'
import type { TextFieldProps as AriaTextFieldProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface TextAreaProps extends Omit<AriaTextFieldProps, 'className' | 'children'> {
  label?: string
  placeholder?: string
  rows?: number
  className?: string
  textareaClassName?: string
  /** HeroUI-era passthroughs. */
  maxLength?: number
  style?: React.CSSProperties
  disabled?: boolean
  fullWidth?: boolean
  /** Forwarded to the underlying <textarea>. */
  ref?: React.Ref<HTMLTextAreaElement>
}

export const TextArea: React.FC<TextAreaProps> = ({
  label,
  placeholder,
  rows = 4,
  className,
  textareaClassName,
  maxLength,
  style,
  disabled,
  fullWidth: _fullWidth,
  isDisabled,
  ref,
  ...props
}): React.ReactElement => (
  <AriaTextField
    {...props}
    isDisabled={isDisabled ?? disabled ?? false}
    className={cn('flex flex-col gap-1.5', className)}
  >
    {renderLabel(label)}
    <AriaTextArea
      rows={rows}
      {...(ref === undefined ? {} : { ref })}
      {...(placeholder === undefined ? {} : { placeholder })}
      {...(maxLength === undefined ? {} : { maxLength })}
      {...(style === undefined ? {} : { style })}
      className={cn(
        'hud-field w-full resize-y bg-transparent px-3 py-2 font-mono text-[13px] leading-relaxed text-foreground outline-none placeholder:text-muted/70',
        textareaClassName,
      )}
    />
  </AriaTextField>
)

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}
