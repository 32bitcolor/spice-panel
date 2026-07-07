import * as React from 'react'
import { cn } from '../ui'
import type { SideNavItem, SideNavProps, SideNavRenderSlot } from './types'

const renderSlot = (slot: SideNavRenderSlot | undefined, active: boolean): React.ReactNode => {
  if (slot === undefined || slot === null) return null
  return typeof slot === 'function' ? slot(active) : slot
}

const renderItem = <K extends string>(
  item: SideNavItem<K>,
  active: K | null,
  onSelect: (key: K) => void,
): React.ReactElement => {
  const isActive = item.key === active
  return (
    <button
      key={item.key}
      type="button"
      onClick={() => onSelect(item.key)}
      aria-current={isActive ? 'true' : undefined}
      className={cn(
        'flex w-full items-center gap-2.5 px-3 py-2 text-left outline-none transition',
        'border-l-2 border-transparent focus-visible:hud-glow',
        item.depth ? 'pl-6' : '',
        isActive
          ? 'border-accent bg-[linear-gradient(90deg,color-mix(in_srgb,var(--accent)_22%,transparent),color-mix(in_srgb,var(--accent)_8%,transparent))]'
          : 'hover:bg-[color-mix(in_srgb,var(--accent)_10%,transparent)]',
      )}
    >
      {renderIcon(item, isActive)}
      <div className="flex min-w-0 flex-1 flex-col">
        <span
          className={cn('truncate text-sm', isActive ? 'font-semibold text-focus' : 'text-foreground')}
        >
          {item.label}
        </span>
        {renderSublabel(item.sublabel)}
      </div>
      {renderHint(item, isActive)}
    </button>
  )
}

const renderIcon = <K extends string>(item: SideNavItem<K>, active: boolean): React.ReactNode => {
  if (item.icon === undefined || item.icon === null) return null
  return <div className="shrink-0">{renderSlot(item.icon, active)}</div>
}

const renderSublabel = (sublabel: React.ReactNode): React.ReactNode => {
  if (sublabel === undefined || sublabel === null) return null
  return <span className="truncate text-xs text-muted">{sublabel}</span>
}

const renderHint = <K extends string>(item: SideNavItem<K>, active: boolean): React.ReactNode => {
  if (item.hint === undefined || item.hint === null) return null
  return <div className="shrink-0 text-xs">{renderSlot(item.hint, active)}</div>
}

export const SideNav = <K extends string>({
  items,
  active,
  onSelect,
  title,
  titleAction,
  width,
  children,
  listHeader,
  emptyContent,
}: SideNavProps<K>): React.ReactElement => {
  const listRef = React.useRef<HTMLDivElement>(null)

  React.useEffect(() => {
    if (active == null || !listRef.current) return
    const el = listRef.current.querySelector<HTMLElement>('[aria-current="true"]')
    el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' })
  }, [active])

  return (
    <div
      className={cn(
        'flex shrink-0 flex-col overflow-hidden bg-surface ring-1 ring-inset ring-border/60 [border-radius:var(--radius)]',
        width ?? 'w-60',
      )}
    >
      {renderHeader(title, titleAction)}
      {renderChildren(children)}
      {renderListHeader(listHeader)}
      <div ref={listRef} className="min-h-0 flex-1 overflow-y-auto overflow-x-hidden">
        {items.length === 0 ? renderEmpty(emptyContent) : items.map((item) => renderItem(item, active, onSelect))}
      </div>
    </div>
  )
}

const renderHeader = (title: React.ReactNode, titleAction: React.ReactNode): React.ReactNode => {
  if (!title && !titleAction) return null
  return (
    <div className="flex shrink-0 items-center justify-between border-b border-border/60 p-3">
      {renderTitle(title)}
      {titleAction}
    </div>
  )
}

const renderTitle = (title: React.ReactNode): React.ReactNode => {
  if (!title) return null
  return (
    <span className="text-xs font-semibold uppercase tracking-widest text-accent">{title}</span>
  )
}

const renderChildren = (children: React.ReactNode): React.ReactNode => {
  if (!children) return null
  return <div className="flex shrink-0 flex-col gap-1 px-3 py-1.5">{children}</div>
}

const renderListHeader = (listHeader: React.ReactNode): React.ReactNode => {
  if (!listHeader) return null
  return <div className="shrink-0 px-3 pb-2">{listHeader}</div>
}

const renderEmpty = (emptyContent: React.ReactNode): React.ReactNode => {
  if (!emptyContent) return null
  return (
    <div className="flex h-full items-center justify-center px-4 py-6 text-center text-sm text-muted">
      {emptyContent}
    </div>
  )
}
