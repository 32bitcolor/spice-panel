import * as React from 'react'
import type { MarketListing } from '../api/client'
import type { ItemEntry } from '../data/store'

export interface TimezoneSelectProps {
  value: string
  onChange: (v: string) => void
  className?: string
}

export interface BackendUnreachableProps {
  onRetry: () => void
}

export interface FieldProps {
  label: string
  hint?: string
  children: React.ReactNode
}

export interface TextInputProps {
  value: string | number
  onChange: (v: string) => void
  placeholder?: string
  type?: string
  autoComplete?: string
}

export interface CheckboxFieldProps {
  label: string
  checked: boolean
  onChange: (v: boolean) => void
  hint?: string
}

export interface GridRowProps {
  children: React.ReactNode
}

export interface DiscordRole { id: string, name: string }

export interface RolePickerProps {
  value: string
  onChange: (v: string) => void
  roles: DiscordRole[]
  label: string
  hint?: string
}

export interface ItemDetailCardRowProps {
  label: string
  value: string
  accent?: boolean
  wrap?: boolean
}

export interface MitigationBarProps {
  label: string
  value: number
}

export type MarketDetail = {
  lowestPrice: number
  totalStock: number
  botStock: number
  listingCount: number
  listings: MarketListing[]
  listingsLoading: boolean
}

export type ItemDetailCardProps = {
  templateId: string
  /** Display name override (e.g. from the templates list). Falls back to entry.name then templateId. */
  name?: string
  entry: ItemEntry | null
  market?: MarketDetail
}
