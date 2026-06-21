import * as React from 'react'
import { Drawer } from '@heroui/react'
import { useAtomValue } from 'jotai'
import { itemDataSyncAtom } from '../data/store'
import { ItemDetailCard } from './ItemDetailCard'
import type { ItemDetailDrawerProps } from './interfaces'

// Right-slide item detail drawer. Reads itemDataSyncAtom internally so callers
// only need to pass templateId + an optional display name.
export const ItemDetailDrawer: React.FC<ItemDetailDrawerProps> = ({ templateId, name, onClose }) => {
  const itemData = useAtomValue(itemDataSyncAtom)
  return (
    <Drawer.Backdrop
      variant="opaque"
      isOpen={!!templateId}
      onOpenChange={(v) => { if (!v) onClose() }}
    >
      <Drawer.Content placement="right">
        <Drawer.Dialog className="w-[480px] max-w-[95vw] flex flex-col">
          <Drawer.Header>
            <div className="flex items-center gap-2 px-4 py-3 border-b border-border w-full">
              <Drawer.Heading className="font-semibold text-sm text-accent truncate flex-1">
                {templateId ? (name || templateId) : ''}
              </Drawer.Heading>
              <Drawer.CloseTrigger />
            </div>
          </Drawer.Header>
          <Drawer.Body className="flex flex-col gap-3 p-3 overflow-y-auto">
            {templateId && (
              <ItemDetailCard
                templateId={templateId}
                name={name}
                entry={itemData.items[templateId] ?? null}
              />
            )}
          </Drawer.Body>
        </Drawer.Dialog>
      </Drawer.Content>
    </Drawer.Backdrop>
  )
}
