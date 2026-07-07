import * as React from 'react'
import { Button, SearchField } from '../../ui'
import { useTranslation } from 'react-i18next'
import { Icon } from '../../dune-ui'
import type { MarketSidebarProps, Node } from './types'

const buildTree = (categories: string[]): { items: Node[], schematics: Node[] } => {
  const itemRoot: Node[] = []
  const schematicRoot: Node[] = []

  for (const cat of [...categories].sort()) {
    const isSchematic = cat.startsWith('schematics/')
    // Strip the top-level prefix before splitting so we don't create a spurious
    // "Schematics" parent node inside the schematics section (or "Items" inside items).
    const stripped = isSchematic
      ? cat.replace(/^schematics\//, '')
      : cat.replace(/^items\//, '')
    const parts = stripped.split('/')
    const root = isSchematic ? schematicRoot : itemRoot

    let current = root
    let displayPath = ''
    let filterPath = ''
    for (const part of parts) {
      displayPath = displayPath ? `${displayPath}/${part}` : part
      filterPath = isSchematic
        ? (filterPath ? `${filterPath}/${part}` : `schematics/${part}`)
        : (filterPath ? `${filterPath}/${part}` : `items/${part}`)

      let node = current.find((n) => n.label === part)
      if (!node) {
        node = { label: part, path: filterPath, displayPath, children: [] }
        current.push(node)
      }
      current = node.children
    }
  }

  return { items: itemRoot, schematics: schematicRoot }
}

const formatLabel = (label: string): string =>
  label
    .replace(/([a-z])([A-Z])/g, '$1 $2')
    .replace(/[-_]/g, ' ')
    .replace(/\b\w/g, (c) => c.toUpperCase())

const collectAncestorPaths = (categories: string[], selected: string): Set<string> => {
  const ancestors = new Set<string>()
  for (const cat of categories) {
    if (cat === selected || cat.startsWith(selected + '/') || selected.startsWith(cat + '/')) {
      const parts = cat.replace(/^items\//, '').replace(/^schematics\//, '').split('/')
      let cur = ''
      for (const p of parts) {
        cur = cur ? `${cur}/${p}` : p
        ancestors.add(cur)
      }
    }
  }
  return ancestors
}

// Prune a tree to nodes whose label (or a descendant's label) matches the query.
// A matching branch keeps all of its descendants so the user can drill in.
const filterNodes = (nodes: Node[], q: string): Node[] => {
  const out: Node[] = []
  for (const node of nodes) {
    const selfMatch = formatLabel(node.label).toLowerCase().includes(q) || node.label.toLowerCase().includes(q)
    if (selfMatch) {
      out.push(node)
      continue
    }
    const kids = filterNodes(node.children, q)
    if (kids.length) out.push({ ...node, children: kids })
  }
  return out
}

const flattenNodes = (nodes: Node[]): Node[] =>
  nodes.flatMap((n) => [n, ...flattenNodes(n.children)])

const flattenKeys = (nodes: Node[]): string[] =>
  nodes.flatMap((n) => [n.path, ...flattenKeys(n.children)])

// Seed the expand state on mount: open every top-level branch plus the ancestor
// chain of the currently selected node so the selection is visible.
const defaultExpanded = (categories: string[], selected: string): Set<string> => {
  const { items, schematics } = buildTree(categories)
  const set = new Set<string>()
  for (const node of [...items, ...schematics]) set.add(node.path)
  const ancestors = collectAncestorPaths(categories, selected)
  for (const n of flattenNodes([...items, ...schematics])) {
    if (ancestors.has(n.displayPath)) set.add(n.path)
  }
  return set
}

const ROW_BASE = 'flex items-center gap-1 rounded-[var(--radius)] py-1 pr-2 text-sm'
const ROW_ACTIVE = 'bg-accent text-accent-foreground'
const ROW_INACTIVE = 'text-foreground hover:bg-[color-mix(in_srgb,var(--accent)_10%,transparent)]'

export const MarketSidebar: React.FC<MarketSidebarProps> = ({ categories, selected, onSelect }: MarketSidebarProps) => {
  const { t } = useTranslation()
  const { items: allItems, schematics: allSchematics } = buildTree(categories)
  const [collapsed, setCollapsed] = React.useState(false)
  const [search, setSearch] = React.useState('')
  const [expanded, setExpanded] = React.useState<Set<string>>(() => defaultExpanded(categories, selected))

  const q = search.trim().toLowerCase()
  const items = q ? filterNodes(allItems, q) : allItems
  const schematics = q ? filterNodes(allSchematics, q) : allSchematics

  // While searching, force every surviving branch open so matches are visible.
  // Otherwise honour the user's own expand/collapse toggles.
  const effectiveExpanded = q
    ? new Set([...flattenKeys(items), ...flattenKeys(schematics)])
    : expanded

  const toggle = (path: string): void => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(path)) next.delete(path)
      else next.add(path)
      return next
    })
  }

  // Recursively render a category subtree. Branches get a chevron that toggles
  // expansion; every row's label selects that node (its full filter path).
  const renderNode = (node: Node, depth: number): React.ReactElement => {
    const isBranch = node.children.length > 0
    const isOpen = effectiveExpanded.has(node.path)
    const isActive = selected === node.path
    return (
      <li key={node.displayPath}>
        <div
          className={`${ROW_BASE} ${isActive ? ROW_ACTIVE : ROW_INACTIVE}`}
          style={{ paddingLeft: `${depth * 12 + 4}px` }}
        >
          {isBranch
            ? (
                <button
                  type="button"
                  aria-label={isOpen ? t('market.sidebar.collapseAriaLabel') : t('market.sidebar.expandAriaLabel')}
                  className="shrink-0 cursor-pointer text-muted hover:text-foreground"
                  onClick={() => toggle(node.path)}
                >
                  <Icon name={isOpen ? 'chevron-down' : 'chevron-right'} className="size-3.5" />
                </button>
              )
            : <span className="w-3.5 shrink-0" aria-hidden="true" />}
          <button
            type="button"
            className="min-w-0 flex-1 cursor-pointer truncate text-left"
            onClick={() => onSelect(node.path)}
          >
            {formatLabel(node.label)}
          </button>
        </div>
        {isBranch && isOpen
          ? (
              <ul className="flex flex-col">
                {node.children.map((child) => renderNode(child, depth + 1))}
              </ul>
            )
          : null}
      </li>
    )
  }

  const renderTree = (nodes: Node[]): React.ReactElement | null =>
    nodes.length === 0
      ? null
      : <ul className="flex flex-col">{nodes.map((n) => renderNode(n, 0))}</ul>

  const renderSchematics = (): React.ReactNode => {
    if (schematics.length === 0) return null
    return (
      <React.Fragment>
        <div className="my-2 border-t border-border/40" />
        <span className="text-[10px] font-semibold text-muted/60 uppercase tracking-wider px-1 mb-0.5 block">
          {t('market.sidebar.schematics')}
        </span>
        {renderTree(schematics)}
      </React.Fragment>
    )
  }

  if (collapsed) {
    return (
      <div className="flex flex-col items-center gap-1 shrink-0">
        <Button size="sm" variant="ghost" isIconOnly aria-label={t('market.sidebar.expandAriaLabel')} onPress={() => setCollapsed(false)}>
          <Icon name="chevron-right" />
        </Button>
      </div>
    )
  }

  return (
    <div className="w-56 shrink-0 flex flex-col gap-1 overflow-hidden pr-1">
      <div className="flex items-center justify-between">
        <span className="text-xs font-semibold text-muted uppercase tracking-wider">{t('market.sidebar.categories')}</span>
        <Button size="sm" variant="ghost" isIconOnly aria-label={t('market.sidebar.collapseAriaLabel')} onPress={() => setCollapsed(true)}>
          <Icon name="chevron-left" />
        </Button>
      </div>

      <SearchField aria-label={t('market.sidebar.categories')} value={search} onChange={setSearch}>
        <SearchField.Group>
          <SearchField.SearchIcon />
          <SearchField.Input placeholder={t('market.sidebar.categories')} />
          <SearchField.ClearButton />
        </SearchField.Group>
      </SearchField>

      <Button
        size="sm"
        variant="ghost"
        className={
          'w-full justify-start rounded-[var(--radius)] px-3 font-medium '
          + (selected === ''
            ? 'text-accent'
            : 'text-foreground hover:bg-default/60')
        }
        {...(selected === '' ? { style: { backgroundColor: 'color-mix(in srgb, var(--accent) 14%, var(--surface))' } } : {})}
        onPress={() => onSelect('')}
      >
        {t('market.sidebar.allItems')}
      </Button>

      <div className="flex-1 overflow-y-auto">
        {renderTree(items)}
        {renderSchematics()}
      </div>
    </div>
  )
}
