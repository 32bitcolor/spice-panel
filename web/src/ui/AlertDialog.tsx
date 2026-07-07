import * as React from 'react'
import {
  ModalOverlay as AriaModalOverlay,
  Modal as AriaModal,
  Dialog as AriaDialog,
  Heading,
} from 'react-aria-components'
import { cn } from './lib/cn'
import { Button } from './Button'

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

export const AlertDialog: React.FC<AlertDialogProps> = ({
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
