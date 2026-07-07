import * as React from 'react'
import {
  Disclosure as AriaDisclosure,
  DisclosurePanel,
  Button as AriaButton,
  Heading,
} from 'react-aria-components'
import type { DisclosureProps as AriaDisclosureProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface DisclosureProps extends Omit<AriaDisclosureProps, 'className' | 'children'> {
  title: React.ReactNode
  children: React.ReactNode
  className?: string
  defaultExpanded?: boolean
}

export const Disclosure: React.FC<DisclosureProps> = ({
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
    <Heading>
      <AriaButton
        slot="trigger"
        className="flex w-full items-center justify-between gap-2 py-3 text-left text-[13px] font-medium text-foreground outline-none data-[focus-visible]:hud-glow"
      >
        {title}
        <svg
          viewBox="0 0 16 16"
          className="h-4 w-4 shrink-0 text-muted transition-transform group-data-[expanded]:rotate-180"
          fill="none"
          stroke="currentColor"
          strokeWidth="1.8"
          strokeLinecap="round"
          strokeLinejoin="round"
        >
          <path d="m4 6 4 4 4-4" />
        </svg>
      </AriaButton>
    </Heading>
    <DisclosurePanel className="pb-3 text-[13px] text-muted">{children}</DisclosurePanel>
  </AriaDisclosure>
)
