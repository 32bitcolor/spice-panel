import * as React from 'react'
import { Chip } from '../../ui'
import { useTranslation } from 'react-i18next'
import { Icon } from '../../dune-ui'
import { Brand } from '../Brand'
import type { TabId } from '../../types'
import { BETA_TABS, DEFAULT_TAB, TAB_ICONS } from './nav'
import type { AppSidebarProps } from './interfaces'

export const AppSidebar: React.FC<AppSidebarProps> = ({ visibleNavGroups, pathname, navigate }) => {
  const { t } = useTranslation()

  // A single top-level menu item. Sub-sections (Database, Welcome Kits,
  // Battle Pass) live inside their tab via an in-header Segment, so every
  // sidebar item is a plain top-level entry. Active items get a left accent
  // border + focus text; the rest are muted with a hover surface.
  const menuItem = (key: TabId): React.ReactNode => {
    const label = visibleNavGroups.flatMap((g) => g.items).find((i) => i.key === key)?.label ?? key
    const isCurrent = pathname === `/${key}`
    const cls = isCurrent
      ? 'border-accent bg-accent/15 text-focus'
      : 'border-transparent text-muted hover:bg-surface-secondary hover:text-foreground'

    return (
      <button
        key={key}
        type="button"
        onClick={() => navigate(`/${key}`)}
        {...(isCurrent ? { 'aria-current': 'page' as const } : {})}
        className={`flex w-full items-center gap-2 border-l-2 px-3 py-2 text-left text-sm transition-colors ${cls}`}
      >
        <Icon name={TAB_ICONS[key]} />
        <span className="flex-1 truncate">{label}</span>
        {BETA_TABS.has(key) && (
          <Chip size="sm" color="accent" variant="soft" className="ml-1 text-[9px] h-4 px-1 min-w-0 shrink-0 self-center">{t('common.beta')}</Chip>
        )}
      </button>
    )
  }

  return (
    <aside className="flex w-60 shrink-0 flex-col border-r border-border bg-surface">
      <div className="flex h-14 shrink-0 items-center border-b border-border px-2">
        <button
          type="button"
          className="flex h-full w-full items-center gap-0 px-2 hover:opacity-80"
          onClick={() => navigate(`/${DEFAULT_TAB}`)}
          aria-label={t('app.goHome')}
        >
          <Brand />
        </button>
      </div>
      <nav aria-label={t('nav.menu')} className="min-h-0 flex-1 overflow-y-auto pb-2">
        {visibleNavGroups.map((group) => (
          <div key={group.title} className="py-1">
            <div className="px-3 py-1 text-xs font-semibold uppercase tracking-wide text-muted">{group.title}</div>
            {group.items.map((item) => menuItem(item.key))}
          </div>
        ))}
      </nav>
    </aside>
  )
}
