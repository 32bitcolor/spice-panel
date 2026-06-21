import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { Button, toast } from '@heroui/react'
import { Icon } from '../../../dune-ui'
import { copyText } from '../../../utils/clipboard'
import type { InterfaceRowProps } from './types'

export const InterfaceRow: React.FC<InterfaceRowProps> = ({ item }) => {
  const { t } = useTranslation()
  const copy = () => {
    copyText(item.url).then((ok) =>
      (ok ? toast.success(t('serverHealth.copied')) : toast.danger(t('serverHealth.copyFailed'))))
  }
  return (
    <div className="flex items-center gap-2">
      <Icon name="external-link" className="size-4 text-accent" />
      <div className="flex flex-col min-w-0 flex-1">
        <span className="text-sm font-semibold">{item.label}</span>
        <span className="text-xs text-muted font-mono truncate">{item.url}</span>
      </div>
      <Button size="sm" variant="ghost" isIconOnly aria-label={t('serverHealth.copy')} onPress={copy}>
        <Icon name="copy" />
      </Button>
      <Button size="sm" variant="outline" onPress={() => window.open(item.url, '_blank', 'noopener')}>
        {t('serverHealth.open')}
      </Button>
    </div>
  )
}
