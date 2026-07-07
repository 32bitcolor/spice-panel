import * as React from 'react'
import { Checkbox as AriaCheckbox } from 'react-aria-components'
import type { CheckboxProps as AriaCheckboxProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface CheckboxProps extends Omit<AriaCheckboxProps, 'className' | 'children'> {
  className?: string
  children?: React.ReactNode
}

export const Checkbox: React.FC<CheckboxProps> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaCheckbox
    {...props}
    className={cn(
      'group flex cursor-pointer items-center gap-2.5 text-[13px] text-foreground outline-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40',
      className,
    )}
  >
    <span className="grid h-[18px] w-[18px] place-items-center bg-[var(--void)] text-accent-foreground shadow-[inset_0_0_0_1px_var(--steel)] transition [clip-path:polygon(3px_0,100%_0,100%_calc(100%-3px),calc(100%-3px)_100%,0_100%,0_3px)] group-data-[selected]:bg-accent group-data-[selected]:shadow-[0_0_10px_-2px_var(--accent)] group-data-[indeterminate]:bg-accent group-data-[focus-visible]:hud-glow">
      <svg
        viewBox="0 0 16 16"
        className="h-3 w-3 opacity-0 transition group-data-[selected]:opacity-100"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.5"
        strokeLinecap="round"
        strokeLinejoin="round"
      >
        <path d="m3 8 3.5 3.5L13 4" />
      </svg>
    </span>
    {children}
  </AriaCheckbox>
)
