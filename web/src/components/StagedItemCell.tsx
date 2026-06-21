import * as React from 'react'
import { Chip } from '@heroui/react'
import { ItemIcon } from './ItemIcon'
import type { StagedItemCellProps } from './interfaces'

// Sub-component exported so react-refresh treats it as a stable top-level component.
// Display-only (no picker semantics) — used in DataGrid template columns across all item surfaces.
export const StagedItemCell: React.FC<StagedItemCellProps> = ({ templateId, name, entry }) => {
  const rarity = entry?.rarity?.toLowerCase()
  return (
    <div className="flex items-center gap-2 py-0.5">
      <ItemIcon
        templateId={templateId}
        category={entry?.category}
        rarity={entry?.rarity}
        name={name || undefined}
      />
      <div className="flex-1 min-w-0">
        <div className="text-xs truncate text-foreground">{name || templateId}</div>
        {name && <div className="font-mono text-[10px] text-muted truncate">{templateId}</div>}
      </div>
      {!!entry?.tier && entry.tier > 0 && (
        <Chip size="sm" variant="soft" className="shrink-0">{`T${entry.tier}`}</Chip>
      )}
      {rarity && (
        <Chip size="sm" variant="soft" className="shrink-0 capitalize" style={{ color: `var(--rarity-${rarity})` }}>
          {rarity}
        </Chip>
      )}
    </div>
  )
}
