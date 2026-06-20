export type Section = 'pending' | 'progress' | 'catalog' | 'track' | 'config'
export type TierKey = 'label' | 'category' | 'requirement' | 'intel' | 'rewards' | 'earned' | 'granted' | 'enabled' | 'actions'
export type PendingKey = 'name' | 'tier_label' | 'intel' | 'items' | 'actions'

export interface CardArtProps {
  folder: string
  file: string
}

export interface ThemeIconProps {
  folder: string
  name: string
  className?: string
}
