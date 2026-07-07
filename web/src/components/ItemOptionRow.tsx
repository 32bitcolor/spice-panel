import * as React from 'react'
import { Chip } from '../ui'
import { Icon } from '../dune-ui'
import { ItemIcon } from './ItemIcon'
import type { ItemEntry } from '../data/store'

export type ItemOptionRowProps = {
  id: string
  name: string
  entry: ItemEntry | null
  onPick: () => void
  /** If provided, renders an info button that opens the detail drawer. */
  onDetail?: () => void
}

export const ItemOptionRow: React.FC<ItemOptionRowProps> = ({ id, name, entry, onPick, onDetail }) => {
  const rarity = entry?.rarity?.toLowerCase()

  return (
    <div
      className="flex items-center gap-2 px-2 py-1.5 cursor-pointer hover:bg-surface-hover"
      onClick={onPick}
    >
      <ItemIcon
        templateId={id}
        category={entry?.category}
        rarity={entry?.rarity}
        name={name || undefined}
        sizeClassName="w-7 h-7"
      />

      {/* Name + id */}
      <div className="flex-1 min-w-0">
        <div className="text-xs truncate text-foreground">{name || id}</div>
        <div className="font-mono text-[10px] text-muted truncate">{id}</div>
      </div>

      {/* Chips */}
      {!!entry?.tier && entry.tier > 0 && (
        <Chip size="sm" variant="soft" className="shrink-0">
          {`T${entry.tier}`}
        </Chip>
      )}
      {rarity && (
        <Chip size="sm" variant="soft" className="shrink-0 capitalize" style={{ color: `var(--rarity-${rarity})` }}>
          {rarity}
        </Chip>
      )}

      {/* Detail info button — stopPropagation so it doesn't trigger onPick */}
      {onDetail && (
        <span
          className="shrink-0 text-muted hover:text-foreground p-0.5 rounded cursor-pointer"
          onClick={(e) => {
            e.stopPropagation()
            onDetail()
          }}
          role="button"
          aria-label="Item details"
          tabIndex={0}
          onKeyDown={(e) => {
            if (e.key === 'Enter' || e.key === ' ') {
              e.stopPropagation()
              onDetail()
            }
          }}
        >
          <Icon name="info" className="w-3.5 h-3.5" />
        </span>
      )}
    </div>
  )
}
