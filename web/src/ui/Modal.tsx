import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading,
  Button as AriaButton,
} from 'react-aria-components'
import type { ModalOverlayProps } from 'react-aria-components'
import { tv } from 'tailwind-variants'
import type { VariantProps } from 'tailwind-variants'
import { cn } from './lib/cn'

const modalStyles = tv({
  base: 'hud-panel relative w-full outline-none',
  variants: {
    size: {
      sm: 'max-w-sm',
      md: 'max-w-lg',
      lg: 'max-w-2xl',
      xl: 'max-w-4xl',
      cover: 'max-w-6xl',
    },
  },
  defaultVariants: { size: 'md' },
})

export type ModalVariants = VariantProps<typeof modalStyles>

export interface ModalProps extends Omit<ModalOverlayProps, 'className' | 'children'>, ModalVariants {
  /** Optional dialog title rendered in the header. */
  title?: string
  className?: string
  children?: React.ReactNode
  /** Hide the built-in close (✕) button. */
  hideClose?: boolean
}

export const Modal: React.FC<ModalProps> = ({
  title,
  size,
  className,
  children,
  hideClose = false,
  ...props
}): React.ReactElement => (
  <AriaModalOverlay
    {...props}
    className="fixed inset-0 z-[900] grid place-items-center overflow-y-auto bg-[linear-gradient(to_top,color-mix(in_srgb,var(--background)_88%,transparent),color-mix(in_srgb,var(--background)_45%,transparent))] p-4 backdrop-blur-sm transition-opacity duration-150 data-[exiting]:opacity-0"
  >
    <AriaModal className={cn(modalStyles({ size }), className)}>
      <AriaDialog className="outline-none">
        {renderHeader(title, hideClose)}
        <div className="p-5">{children}</div>
      </AriaDialog>
    </AriaModal>
  </AriaModalOverlay>
)

const renderHeader = (title: string | undefined, hideClose: boolean): React.ReactNode => {
  if (title === undefined && hideClose) return null
  return (
    <div className="flex items-center justify-between border-b border-border px-5 py-3.5">
      <Heading
        slot="title"
        className="font-display text-sm uppercase tracking-[0.14em] text-foreground"
      >
        {title}
      </Heading>
      {renderClose(hideClose)}
    </div>
  )
}

const renderClose = (hideClose: boolean): React.ReactNode => {
  if (hideClose) return null
  return (
    <AriaButton
      slot="close"
      aria-label="Close"
      className="grid h-6 w-6 place-items-center text-muted outline-none transition hover:text-foreground data-[focus-visible]:hud-glow"
    >
      <svg viewBox="0 0 16 16" className="h-4 w-4" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
        <path d="m4 4 8 8M12 4l-8 8" />
      </svg>
    </AriaButton>
  )
}
