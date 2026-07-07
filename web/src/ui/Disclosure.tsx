import * as React from 'react'
import {
  Disclosure as AriaDisclosure,
  DisclosurePanel,
  Button as AriaButton,
  Heading as AriaHeading,
} from 'react-aria-components'
import type { DisclosureProps as AriaDisclosureProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface DisclosureProps extends Omit<AriaDisclosureProps, 'className' | 'children'> {
  /** Simple API: header text. Omit when composing with Disclosure.Heading/.Content. */
  title?: React.ReactNode
  children: React.ReactNode
  className?: string
  defaultExpanded?: boolean
}

const Chevron: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <svg
    viewBox="0 0 16 16"
    className={cn('h-4 w-4 shrink-0 text-muted transition-transform group-data-[expanded]:rotate-180', className)}
    fill="none"
    stroke="currentColor"
    strokeWidth="1.8"
    strokeLinecap="round"
    strokeLinejoin="round"
  >
    <path d="m4 6 4 4 4-4" />
  </svg>
)

const DisclosureRoot: React.FC<DisclosureProps> = ({
  title,
  children,
  className,
  defaultExpanded,
  ...props
}): React.ReactElement => (
  <AriaDisclosure
    {...props}
    {...(defaultExpanded === undefined ? {} : { defaultExpanded })}
    className={cn('group border-b border-border', className)}
  >
    {title === undefined ? children : renderSimple(title, children)}
  </AriaDisclosure>
)

const renderSimple = (title: React.ReactNode, children: React.ReactNode): React.ReactNode => (
  <React.Fragment>
    <AriaHeading>
      <AriaButton
        slot="trigger"
        className="flex w-full items-center justify-between gap-2 py-3 text-left text-[13px] font-medium text-foreground outline-none data-[focus-visible]:hud-glow"
      >
        {title}
        <Chevron />
      </AriaButton>
    </AriaHeading>
    <DisclosurePanel className="pb-3 text-[13px] text-muted">{children}</DisclosurePanel>
  </React.Fragment>
)

/* ── Compound slots (HeroUI-compatible) ───────────────────────────────────── */

const HeadingSlot: React.FC<React.HTMLAttributes<HTMLButtonElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <AriaHeading>
    <AriaButton
      slot="trigger"
      className={cn(
        'flex w-full items-center justify-between gap-2 py-3 text-left text-[13px] font-medium text-foreground outline-none data-[focus-visible]:hud-glow',
        className,
      )}
    >
      {children}
    </AriaButton>
  </AriaHeading>
)

const Trigger: React.FC<React.PropsWithChildren> = ({ children }): React.ReactElement => (
  <React.Fragment>{children}</React.Fragment>
)

const Indicator: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <Chevron {...(className === undefined ? {} : { className })} />
)

const Content: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <DisclosurePanel className={cn('pb-3 text-[13px] text-muted', className)}>
    {children}
  </DisclosurePanel>
)

const Body: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn(className)}>
    {children}
  </div>
)

export const Disclosure = Object.assign(DisclosureRoot, {
  Heading: HeadingSlot,
  Trigger,
  Indicator,
  Content,
  Body,
})
