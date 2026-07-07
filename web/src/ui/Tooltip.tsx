import * as React from 'react'
import {
  TooltipTrigger as AriaTooltipTrigger,
  Tooltip as AriaTooltip,
  Focusable,
  OverlayArrow,
} from 'react-aria-components'
import { cn } from './lib/cn'

type Placement = 'top' | 'bottom' | 'left' | 'right'

export interface TooltipProps {
  children: React.ReactNode
  /** Simple API: tooltip body. Omit when using Tooltip.Trigger/Tooltip.Content. */
  content?: React.ReactNode
  placement?: Placement
  delay?: number
  className?: string
}

const TIP_CLS =
  'hud-panel z-[960] max-w-xs px-2.5 py-1.5 text-xs text-foreground data-[entering]:opacity-0 data-[exiting]:opacity-0'

const Arrow: React.FC = (): React.ReactElement => (
  <OverlayArrow>
    <svg width={8} height={8} viewBox="0 0 8 8" className="fill-[var(--surface)] stroke-[var(--steel)]">
      <path d="M0 0 L4 4 L8 0" />
    </svg>
  </OverlayArrow>
)

/* Marker slots for the compound API. */
const Trigger: React.FC<React.PropsWithChildren> = ({ children }): React.ReactElement => (
  <React.Fragment>{children}</React.Fragment>
)
const Content: React.FC<React.PropsWithChildren> = ({ children }): React.ReactElement => (
  <React.Fragment>{children}</React.Fragment>
)

const extractSlots = (
  children: React.ReactNode,
): { trigger: React.ReactNode; content: React.ReactNode } => {
  let trigger: React.ReactNode = null
  let content: React.ReactNode = null
  React.Children.forEach(children, (child) => {
    if (!React.isValidElement(child)) return
    if (child.type === Trigger) trigger = (child.props as React.PropsWithChildren).children
    else if (child.type === Content) content = (child.props as React.PropsWithChildren).children
  })
  return { trigger, content }
}

const TooltipRoot: React.FC<TooltipProps> = ({
  children,
  content,
  placement = 'top',
  delay = 500,
  className,
}): React.ReactElement => {
  const slots = content === undefined ? extractSlots(children) : { trigger: children, content }
  return (
    <AriaTooltipTrigger delay={delay}>
      <Focusable>
        {slots.trigger as unknown as React.ComponentProps<typeof Focusable>['children']}
      </Focusable>
      <AriaTooltip placement={placement} offset={8} className={cn(TIP_CLS, className)}>
        <Arrow />
        {slots.content}
      </AriaTooltip>
    </AriaTooltipTrigger>
  )
}

export const Tooltip = Object.assign(TooltipRoot, { Trigger, Content })
