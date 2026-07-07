import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading,
} from 'react-aria-components'
import type { ModalOverlayProps } from 'react-aria-components'
import { cn } from './lib/cn'
import { CloseButton } from './CloseButton'

export interface DrawerProps extends Omit<ModalOverlayProps, 'className' | 'children'> {
  title?: string
  side?: 'left' | 'right'
  width?: number
  className?: string
  children?: React.ReactNode
}

export const Drawer: React.FC<DrawerProps> = ({
  title,
  side = 'right',
  width = 400,
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaModalOverlay
    {...props}
    className="fixed inset-0 z-[900] bg-[color-mix(in_srgb,var(--background)_70%,transparent)] backdrop-blur-sm transition-opacity duration-150 data-[exiting]:opacity-0"
  >
    <AriaModal
      style={{ width }}
      className={cn(
        'fixed inset-y-0 max-w-[92vw] bg-surface shadow-[inset_0_0_0_1px_var(--steel)] outline-none transition-transform duration-200',
        side === 'right'
          ? 'right-0 data-[exiting]:translate-x-full'
          : 'left-0 data-[exiting]:-translate-x-full',
        className,
      )}
    >
      <AriaDialog className="flex h-full flex-col outline-none">
        <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
          <Heading slot="title" className="font-display text-sm uppercase tracking-[0.14em]">
            {title}
          </Heading>
          <CloseButton slot="close" />
        </div>
        <div className="flex-1 overflow-y-auto p-5">{children}</div>
      </AriaDialog>
    </AriaModal>
  </AriaModalOverlay>
)
