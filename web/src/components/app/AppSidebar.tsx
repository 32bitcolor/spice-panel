import * as React from 'react'
import { useAtom } from 'jotai'
import { Chip } from '../../ui'
import { useTranslation } from 'react-i18next'
import { Icon } from '../../dune-ui'
import { Brand } from '../Brand'
import { mobileNavOpenAtom } from '../../atoms/app'
import type { TabId } from '../../types'
import { BETA_TABS, DEFAULT_TAB, TAB_ICONS } from './nav'
import type { AppSidebarProps } from './interfaces'

export const AppSidebar: React.FC<AppSidebarProps> = ({ visibleNavGroups, pathname, navigate }) => {
  const { t } = useTranslation()
  const [mobileOpen, setMobileOpen] = useAtom(mobileNavOpenAtom)

  // Navigate, then close the off-canvas drawer (no-op on lg+ where it stays open).
  const go = (path: string): void => {
    navigate(path)
    setMobileOpen(false)
  }

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
        onClick={() => go(`/${key}`)}
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
    <React.Fragment>
      {renderBackdrop(mobileOpen, () => setMobileOpen(false))}
      <aside
        className={`fixed inset-y-0 left-0 z-50 flex w-60 shrink-0 flex-col border-r border-border bg-surface transition-transform duration-200 lg:static lg:z-auto lg:translate-x-0 ${mobileOpen ? 'translate-x-0' : '-translate-x-full'}`}
      >
        <div className="flex h-14 shrink-0 items-center border-b border-border px-2">
          <button
            type="button"
            className="flex h-full w-full items-center gap-0 px-2 hover:opacity-80"
            onClick={() => go(`/${DEFAULT_TAB}`)}
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
    </React.Fragment>
  )
}

const renderBackdrop = (open: boolean, onClose: () => void): React.ReactNode => {
  if (!open) return null
  return (
    <div
      className="fixed inset-0 z-40 bg-black/60 backdrop-blur-sm lg:hidden"
      onClick={onClose}
      aria-hidden="true"
    />
  )
}
