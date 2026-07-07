import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { AlertDialog } from '../ui'
import type { ConfirmDialogProps } from './types'

export const ConfirmDialog: React.FC<ConfirmDialogProps> = ({
  open,
  title,
  description,
  confirmLabel,
  onConfirm,
  onCancel,
}): React.ReactElement => {
  const { t } = useTranslation()
  return (
    <AlertDialog
      isOpen={open}
      onOpenChange={(v) => {
        if (!v) onCancel()
      }}
      title={title}
      description={description}
      destructive
      cancelLabel={t('common.cancel')}
      confirmLabel={confirmLabel ?? t('common.confirm')}
      onConfirm={onConfirm}
    />
  )
}
