import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading,
} from 'react-aria-components'
import { cn } from './lib/cn'
import { Button } from './Button'
import { Modal } from './Modal'

export interface AlertDialogProps {
  isOpen: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: React.ReactNode
  confirmLabel?: string
  cancelLabel?: string
  /** Style the confirm action as destructive. */
  destructive?: boolean
  onConfirm: () => void
  className?: string
}

const AlertDialogRoot: React.FC<AlertDialogProps> = ({
  isOpen,
  onOpenChange,
  title,
  description,
  confirmLabel = 'Confirm',
  cancelLabel = 'Cancel',
  destructive = false,
  onConfirm,
  className,
}): React.ReactElement => (
  <AriaModalOverlay
    isOpen={isOpen}
    onOpenChange={onOpenChange}
    isDismissable
    className="fixed inset-0 z-[900] grid place-items-center bg-[color-mix(in_srgb,var(--background)_75%,transparent)] p-4 backdrop-blur-sm transition-opacity duration-150 data-[exiting]:opacity-0"
  >
    <AriaModal className={cn('hud-panel w-full max-w-md outline-none', className)}>
      <AriaDialog role="alertdialog" className="p-5 outline-none">
        {({ close }) => (
          <div className="flex flex-col gap-4">
            <div className="flex flex-col gap-1.5">
              <Heading slot="title" className="font-display text-sm uppercase tracking-[0.14em] text-foreground">
                {title}
              </Heading>
              {renderDescription(description)}
            </div>
            <div className="flex justify-end gap-2.5">
              <Button variant="ghost" size="sm" onPress={close}>
                {cancelLabel}
              </Button>
              <Button
                variant={destructive ? 'danger' : 'primary'}
                size="sm"
                onPress={() => {
                  onConfirm()
                  close()
                }}
              >
                {confirmLabel}
              </Button>
            </div>
          </div>
        )}
      </AriaDialog>
    </AriaModal>
  </AriaModalOverlay>
)

const renderDescription = (description: React.ReactNode): React.ReactNode => {
  if (description === undefined) return null
  return <p className="text-[13px] leading-relaxed text-muted">{description}</p>
}

/* ── Compound API (HeroUI-compatible) — reuses Modal's slots ───────────────── */

const STATUS_COLOR: Record<string, string> = {
  danger: 'text-danger',
  warning: 'text-warning',
  success: 'text-success',
  info: 'text-accent',
}

const AlertIcon: React.FC<{ status?: string; className?: string }> = ({
  status = 'danger',
  className,
}): React.ReactElement => (
  <span className={cn('shrink-0', STATUS_COLOR[status] ?? 'text-accent', className)}>
    <svg viewBox="0 0 24 24" className="h-6 w-6" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
      <path d="M12 9v4M12 17h.01M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0Z" />
    </svg>
  </span>
)

const AlertDialogSlot: React.FC<React.HTMLAttributes<HTMLElement>> = ({
  className,
  children,
}): React.ReactElement => (
  <AriaDialog role="alertdialog" className={cn('outline-none', className)}>
    {children}
  </AriaDialog>
)

export const AlertDialog = Object.assign(AlertDialogRoot, {
  Backdrop: Modal.Backdrop,
  Container: Modal.Container,
  Dialog: AlertDialogSlot,
  CloseTrigger: Modal.CloseTrigger,
  Header: Modal.Header,
  Heading: Modal.Heading,
  Body: Modal.Body,
  Footer: Modal.Footer,
  Icon: AlertIcon,
})
