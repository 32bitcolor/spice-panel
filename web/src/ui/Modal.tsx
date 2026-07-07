import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading as AriaHeading,
  Button as AriaButton,
} from 'react-aria-components'
import type { ModalOverlayProps } from 'react-aria-components'
import { tv } from 'tailwind-variants'
import type { VariantProps } from 'tailwind-variants'
import { cn } from './lib/cn'

const OVERLAY
  = 'fixed inset-0 z-[900] grid place-items-center overflow-y-auto bg-[linear-gradient(to_top,color-mix(in_srgb,var(--background)_88%,transparent),color-mix(in_srgb,var(--background)_45%,transparent))] p-4 backdrop-blur-sm transition-opacity duration-150 data-[exiting]:opacity-0'

const modalStyles = tv({
  base: 'hud-panel relative w-full outline-none',
  variants: {
    size: {
      'sm': 'max-w-sm',
      'md': 'max-w-lg',
      'lg': 'max-w-2xl',
      'xl': 'max-w-4xl',
      '2xl': 'max-w-5xl',
      'cover': 'max-w-6xl',
    },
  },
  defaultVariants: { size: 'md' },
})

export type ModalVariants = VariantProps<typeof modalStyles>

const CloseIcon: React.FC = (): React.ReactElement => (
  <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
    <path d="m4 4 8 8M12 4l-8 8" />
  </svg>
)

/* ── Compound API (HeroUI-compatible) ─────────────────────────────────────── */

export interface ModalBackdropProps extends Omit<ModalOverlayProps, 'className' | 'children'> {
  /** Accepted for compatibility; the HUD backdrop is always blurred. */
  variant?: string
  className?: string
  children?: React.ReactNode
}

const Backdrop: React.FC<ModalBackdropProps> = ({
  variant: _variant,
  className,
  children,
  ...props
}): React.ReactElement => (
  <AriaModalOverlay {...props} className={cn(OVERLAY, className)}>
    {children}
  </AriaModalOverlay>
)

export interface ModalContainerProps extends ModalVariants {
  /** Accepted for compatibility; overlay handles scrolling. */
  scroll?: string
  className?: string
  children?: React.ReactNode
}

const Container: React.FC<ModalContainerProps> = ({
  size,
  scroll: _scroll,
  className,
  children,
}): React.ReactElement => (
  <AriaModal className={cn(modalStyles({ size }), className)}>{children}</AriaModal>
)

const DialogSlot: React.FC<React.HTMLAttributes<HTMLElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <AriaDialog className={cn('outline-none', className)}>{children}</AriaDialog>
)

const CloseTrigger: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <AriaButton
    slot="close"
    aria-label="Close"
    className={cn(
      'absolute right-3 top-3 z-10 grid h-6 w-6 place-items-center text-muted outline-none transition hover:text-foreground data-[focus-visible]:hud-glow',
      className,
    )}
  >
    <CloseIcon />
  </AriaButton>
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
  <AriaHeading
    {...props}
    slot="title"
    className={cn('font-display text-sm uppercase tracking-[0.14em] text-foreground', className)}
  >
    {children}
  </AriaHeading>
)

const Body: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('p-5', className)}>
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

export interface ModalProps extends Omit<ModalOverlayProps, 'className' | 'children'>, ModalVariants {
  title?: string
  className?: string
  children?: React.ReactNode
  hideClose?: boolean
}

const ModalRoot: React.FC<ModalProps> = ({
  title,
  size,
  className,
  children,
  hideClose = false,
  ...props
}): React.ReactElement => (
  <Backdrop {...props}>
    <Container {...(size === undefined ? {} : { size })} {...(className === undefined ? {} : { className })}>
      <DialogSlot>
        {renderSimpleHeader(title, hideClose)}
        <Body>{children}</Body>
      </DialogSlot>
    </Container>
  </Backdrop>
)

const renderSimpleHeader = (title: string | undefined, hideClose: boolean): React.ReactNode => {
  if (title === undefined && hideClose) return null
  return (
    <HeaderSlot>
      <HeadingSlot>{title}</HeadingSlot>
      {hideClose ? null : <CloseTrigger className="static" />}
    </HeaderSlot>
  )
}

export const Modal = Object.assign(ModalRoot, {
  Backdrop,
  Container,
  Dialog: DialogSlot,
  CloseTrigger,
  Header: HeaderSlot,
  Heading: HeadingSlot,
  Body,
  Footer,
})
