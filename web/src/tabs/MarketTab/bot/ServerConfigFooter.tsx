import * as React from 'react'
import { Button, Spinner } from '../../../ui'
import { useTranslation } from 'react-i18next'
import type { ServerConfigFooterProps } from './interfaces'

export const ServerConfigFooter: React.FC<ServerConfigFooterProps> = ({ configRef }: ServerConfigFooterProps) => {
  const { t } = useTranslation()
  const [saving, setSaving] = React.useState(false)

  return (
    <div className="shrink-0 flex items-center justify-between gap-3 px-4 py-3">
      <p className="text-xs text-muted">{t('market.bot.serverConfig.changesNote')}</p>
      <Button
        size="sm"
        isDisabled={saving}
        onPress={() => {
          setSaving(true)
          configRef.current?.save()
            .catch(() => { /* toast shown inside save */ })
            .finally(() => setSaving(false))
        }}
      >
        {saving ? <Spinner size={16} /> : null}
        {t('common.save')}
      </Button>
    </div>
  )
}
