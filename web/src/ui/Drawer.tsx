import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading as AriaHeading,
} from 'react-aria-components'
import type { ModalOverlayProps } from 'react-aria-components'
import { cn } from './lib/cn'
import { CloseButton } from './CloseButton'

const OVERLAY
  = 'fixed inset-0 z-[900] bg-[color-mix(in_srgb,var(--background)_70%,transparent)] backdrop-blur-sm transition-opacity duration-150 data-[exiting]:opacity-0'

/* ── Compound API (HeroUI-compatible) ─────────────────────────────────────── */

export interface DrawerBackdropProps extends Omit<ModalOverlayProps, 'className' | 'children'> {
  variant?: string
  className?: string
  children?: React.ReactNode
}

const Backdrop: React.FC<DrawerBackdropProps> = ({
  variant: _variant,
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaModalOverlay {...props} className={cn(OVERLAY, className)}>
    {children}
  </AriaModalOverlay>
)

export interface DrawerContentProps {
  placement?: 'left' | 'right'
  className?: string
  children?: React.ReactNode
}

const Content: React.FC<DrawerContentProps> = ({
  placement = 'right',
  className,
  children,
}): React.ReactElement => (
  <AriaModal
    className={cn(
      'fixed inset-y-0 max-w-[92vw] bg-surface shadow-[inset_0_0_0_1px_var(--steel)] outline-none transition-transform duration-200',
      placement === 'right' ? 'right-0 data-[exiting]:translate-x-full' : 'left-0 data-[exiting]:-translate-x-full',
      className,
    )}
  >
    {children}
  </AriaModal>
)

const DialogSlot: React.FC<React.HTMLAttributes<HTMLElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <AriaDialog className={cn('flex h-full flex-col outline-none', className)}>{children}</AriaDialog>
)

const HeaderSlot: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex items-center justify-between gap-3 border-b border-border px-5 py-3.5', className)}>
    {children}
  </div>
)

const HeadingSlot: React.FC<React.HTMLAttributes<HTMLHeadingElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaHeading {...props} slot="title" className={cn('font-display text-sm uppercase tracking-[0.14em]', className)}>
    {children}
  </AriaHeading>
)

const CloseTrigger: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <CloseButton slot="close" {...(className === undefined ? {} : { className })} />
)

const Body: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex-1 overflow-y-auto p-5', className)}>
    {children}
  </div>
)

const Footer: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('flex items-center justify-end gap-2 border-t border-border px-5 py-3.5', className)}>
    {children}
  </div>
)

/* ── Simple all-in-one API ────────────────────────────────────────────────── */

export interface DrawerProps extends Omit<ModalOverlayProps, 'className' | 'children'> {
  title?: string
  side?: 'left' | 'right'
  width?: number
  className?: string
  children?: React.ReactNode
}

const DrawerRoot: React.FC<DrawerProps> = ({
  title,
  side = 'right',
  width = 400,
  className,
  children,
  ...props
}): React.ReactElement => (
  <Backdrop {...props}>
    <Content placement={side} className={className ?? ''}>
      <DialogSlot>
        <HeaderSlot>
          <HeadingSlot>{title}</HeadingSlot>
          <CloseTrigger />
        </HeaderSlot>
        <Body style={{ width }}>{children}</Body>
      </DialogSlot>
    </Content>
  </Backdrop>
)

export const Drawer = Object.assign(DrawerRoot, {
  Backdrop,
  Content,
  Dialog: DialogSlot,
  Header: HeaderSlot,
  Heading: HeadingSlot,
  CloseTrigger,
  Body,
  Footer,
})
