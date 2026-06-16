import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { Button, toast } from '@heroui/react'
import { Icon } from '../../../dune-ui'
import { copyText } from '../../../utils/clipboard'
import type { DirectorRowProps } from './types'

// DirectorRow is the automatic, read-only entry shown when director_url is set:
// the Director usually binds to loopback on the host, so "Open" goes through the
// same-origin /director/ reverse proxy. The configured target is shown for context.
export const DirectorRow: React.FC<DirectorRowProps> = ({ directorURL }) => {
  const { t } = useTranslation()
  const copy = () => {
    copyText(`${window.location.origin}/director/`).then((ok) =>
      (ok ? toast.success(t('serverHealth.copied')) : toast.danger(t('serverHealth.copyFailed'))))
  }
  return (
    <div className="flex items-center gap-2">
      <Icon name="external-link" className="size-4 text-accent" />
      <div className="flex flex-col min-w-0 flex-1">
        <span className="text-sm font-semibold">
          {t('serverHealth.director')}
          {' '}
          <span className="text-xs font-normal text-muted">{t('serverHealth.directorProxied')}</span>
        </span>
        <span className="text-xs text-muted font-mono truncate">{directorURL}</span>
      </div>
      <Button size="sm" variant="ghost" isIconOnly aria-label={t('serverHealth.copy')} onPress={copy}>
        <Icon name="copy" />
      </Button>
      <Button size="sm" variant="outline" onPress={() => window.open('/director/', '_blank', 'noopener,noreferrer')}>
        {t('serverHealth.open')}
      </Button>
    </div>
  )
}
