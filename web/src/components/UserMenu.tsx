import * as React from 'react'
import { Avatar, Button, Dropdown } from '@heroui/react'
import { useTranslation } from 'react-i18next'
import { AuthContext } from '../auth/context'
import { Icon } from '../dune-ui'

// UserMenu shows the signed-in account (Discord avatar when available) with
// a dropdown for sign-out. Rendered only when backend auth is enabled and a
// session exists.
export const UserMenu: React.FC = () => {
  const { enabled, session, logout } = React.useContext(AuthContext)
  const { t } = useTranslation()

  if (!enabled || !session) return null

  const initials = session.name
    .split(/\s+/)
    .map((w) => w[0])
    .join('')
    .slice(0, 2)
    .toUpperCase()

  return (
    <Dropdown>
      <Button
        isIconOnly
        variant="ghost"
        size="sm"
        aria-label={t('auth.accountMenu')}
        className="w-8 h-8 min-w-0 rounded-full p-0"
      >
        <Avatar size="sm" className="size-7">
          {session.avatar && <Avatar.Image src={session.avatar} alt={session.name} />}
          <Avatar.Fallback>{initials}</Avatar.Fallback>
        </Avatar>
      </Button>
      <Dropdown.Popover className="min-w-52">
        <div className="px-2 py-1.5 mb-1 flex items-center gap-2.5 border-b border-border">
          <Avatar size="sm" className="shrink-0">
            {session.avatar && <Avatar.Image src={session.avatar} alt={session.name} />}
            <Avatar.Fallback>{initials}</Avatar.Fallback>
          </Avatar>
          <div className="flex flex-col min-w-0 leading-tight">
            <span className="text-sm font-medium text-foreground truncate">{session.name}</span>
            <span className="text-xs text-muted truncate">
              {session.method === 'discord'
                ? 'Discord'
                : session.method === 'guest' ? t('auth.guestAccount') : t('auth.localAccount')}
              {session.owner && ` · ${t('auth.owner')}`}
            </span>
          </div>
        </div>
        <Dropdown.Menu
          aria-label={t('auth.accountMenu')}
          onAction={(key) => {
            if (String(key) === 'logout') void logout()
          }}
        >
          <Dropdown.Item id="logout" textValue={t('auth.logout')}>
            <Icon name="log-out" />
            {' '}
            {t('auth.logout')}
          </Dropdown.Item>
        </Dropdown.Menu>
      </Dropdown.Popover>
    </Dropdown>
  )
}
