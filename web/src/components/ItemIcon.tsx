import * as React from 'react'
import { iconUrl, categoryColor } from '../utils/icons'
import type { ItemIconProps } from './interfaces'

// Thumbnail box — CDN icon with category-colour background gradient and first-letter fallback.
export const ItemIcon: React.FC<ItemIconProps> = ({ templateId, category, rarity, name, sizeClassName }) => {
  const img = iconUrl(templateId, 'thumb')
  return (
    <div
      className={`${sizeClassName ?? 'w-6 h-6'} shrink-0 rounded flex items-center justify-center overflow-hidden`}
      style={{ background: categoryColor(category ?? '', rarity, templateId) }}
    >
      {img
        ? (
            <img
              src={img}
              alt=""
              className="w-full h-full object-contain"
              onError={(e) => { (e.currentTarget as HTMLImageElement).style.display = 'none' }}
            />
          )
        : (
            <span className="text-[10px] text-white/30 font-bold uppercase select-none">
              {(name || templateId).charAt(0)}
            </span>
          )}
    </div>
  )
}
