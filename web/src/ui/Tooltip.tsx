import * as React from 'react'
import {
  TooltipTrigger as AriaTooltipTrigger,
  Tooltip as AriaTooltip,
  Focusable,
  OverlayArrow,
} from 'react-aria-components'
import { cn } from './lib/cn'

export interface TooltipProps {
  /** The trigger element. Wrapped in Focusable so any element works. */
  children: React.ReactElement
  content: React.ReactNode
  placement?: 'top' | 'bottom' | 'left' | 'right'
  delay?: number
  className?: string
}

export const Tooltip: React.FC<TooltipProps> = ({
  children,
  content,
  placement = 'top',
  delay = 500,
  className,
}): React.ReactElement => (
  <AriaTooltipTrigger delay={delay}>
    <Focusable>
      {children as unknown as React.ComponentProps<typeof Focusable>['children']}
    </Focusable>
    <AriaTooltip
      placement={placement}
      offset={8}
      className={cn(
        'hud-panel z-[960] max-w-xs px-2.5 py-1.5 text-xs text-foreground data-[entering]:opacity-0 data-[exiting]:opacity-0',
        className,
      )}
    >
      <OverlayArrow>
        <svg width={8} height={8} viewBox="0 0 8 8" className="fill-[var(--surface)] stroke-[var(--steel)]">
          <path d="M0 0 L4 4 L8 0" />
        </svg>
      </OverlayArrow>
      {content}
    </AriaTooltip>
  </AriaTooltipTrigger>
)
