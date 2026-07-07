import * as React from 'react'
import { Switch as AriaSwitch } from 'react-aria-components'
import type { SwitchProps as AriaSwitchProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface SwitchProps extends Omit<AriaSwitchProps, 'className' | 'children'> {
  className?: string
  children?: React.ReactNode
  /** HeroUI-era size — accepted for compatibility; styling is unified. */
  size?: 'sm' | 'md'
}

const SwitchRoot: React.FC<SwitchProps> = ({
  className,
  children,
  size: _size,
  ...props
}): React.ReactElement => (
  <AriaSwitch
    {...props}
    className={cn(
      'group flex cursor-pointer items-center gap-2.5 text-[13px] text-muted outline-none data-[disabled]:cursor-not-allowed data-[disabled]:opacity-40',
      className,
    )}
  >
    <span className="relative h-[22px] w-[42px] shrink-0 bg-[var(--void)] shadow-[inset_0_0_0_1px_var(--steel)] transition [clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)] group-data-[selected]:bg-accent/20 group-data-[selected]:shadow-[inset_0_0_0_1px_var(--accent),0_0_12px_-3px_var(--accent)] group-data-[focus-visible]:hud-glow">
      <span className="absolute left-[3px] top-[3px] h-4 w-4 bg-muted transition group-data-[selected]:left-[23px] group-data-[selected]:bg-focus group-data-[selected]:shadow-[0_0_8px_var(--accent)]" />
    </span>
    {children}
  </AriaSwitch>
)

// HeroUI-compatible compound slots. Our Switch draws its own track/thumb, so
// Control/Thumb render nothing; Content just wraps the label.
const Control: React.FC<React.HTMLAttributes<HTMLElement>> = (): null => null
const Thumb: React.FC = (): null => null
const Content: React.FC<React.HTMLAttributes<HTMLSpanElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <span {...props} className={cn(className)}>
    {children}
  </span>
)

export const Switch = Object.assign(SwitchRoot, { Control, Thumb, Content })
