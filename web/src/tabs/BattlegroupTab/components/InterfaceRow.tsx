import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { Button, toast } from '../../../ui'
import { Icon } from '../../../dune-ui'
import { copyText } from '../../../utils/clipboard'
import type { InterfaceRowProps } from './types'

export const InterfaceRow: React.FC<InterfaceRowProps> = ({ item }) => {
  const { t } = useTranslation()
  // Proxied entries (proxyPort set) are reached through dune-admin's own host on
  // the assigned port — bypassing the game host the operator can't resolve/route.
  // The host comes from window.location; the scheme comes from the backend
  // (proxyScheme), NOT window.location.protocol: the proxy listeners are always
  // plain HTTP, so an HTTPS-served dashboard must still open them over http.
  // window.location.hostname brackets an IPv6 literal in Chrome/Safari but not in
  // Firefox/IE, so bracket it ourselves to keep `host:port` a valid URL.
  const host = window.location.hostname
  const hostForURL = host.includes(':') && !host.startsWith('[') ? `[${host}]` : host
  const openURL = item.proxyPort
    ? `${item.proxyScheme ?? 'http'}://${hostForURL}:${item.proxyPort}/`
    : item.url
  const copy = (): void => {
    copyText(openURL).then((ok) =>
      (ok ? toast.success(t('serverHealth.copied')) : toast.danger(t('serverHealth.copyFailed'))))
  }
  return (
    <div className="flex items-center gap-2">
      <Icon name="external-link" className="size-4 text-accent" />
      <div className="flex flex-col min-w-0 flex-1">
        <span className="text-sm font-semibold">{item.label}</span>
        <span className="text-xs text-muted font-mono truncate">{openURL}</span>
      </div>
      <Button size="sm" variant="ghost" isIconOnly aria-label={t('serverHealth.copy')} onPress={copy}>
        <Icon name="copy" />
      </Button>
      <Button size="sm" variant="ghost" onPress={() => window.open(openURL, '_blank', 'noopener,noreferrer')}>
        {t('serverHealth.open')}
      </Button>
    </div>
  )
}
