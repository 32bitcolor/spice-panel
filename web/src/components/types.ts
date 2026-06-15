import * as React from 'react'
import type { AppConfig } from '../api/client'

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

export interface SettingsConfigFormProps {
  saveRef?: React.MutableRefObject<(() => Promise<void>) | null>
  onSavingChange?: (saving: boolean) => void
  /** When set, overrides the internal tab state (wizard mode). */
  activeTab?: string
  /** When true, hides the Segment tab bar (wizard drives navigation). */
  hideTabBar?: boolean
  /** When true, skips loading existing config and starts with empty fields (add-server mode). */
  skipLoad?: boolean
  /**
   * Settings-modal only: invoked from the per-server Advanced "Danger Zone" to
   * request deletion of the active server. When omitted, the Danger Zone is
   * hidden (e.g. the wizard, or callers without server:control capability).
   */
  onRequestDeleteServer?: () => void
  /**
   * Add-server wizard: persist creates a NEW per-server entry via
   * POST /servers (not the flat config). Only per-server fields are sent;
   * global settings (auth, Discord, listen addr) are not part of a new server.
   */
  addMode?: boolean
  /** Add-server wizard: the name entered for the new server (drives id). */
  addServerName?: string
  /**
   * Add-server wizard: called whenever the live form config changes so the
   * wizard can read current values (control plane + SSH) to drive discovery.
   */
  onConfigChange?: (cfg: AppConfig) => void
  /**
   * Add-server wizard: discovered values to merge into the form config. Merged
   * whenever the object identity changes (set once per discovery run).
   */
  prefill?: Partial<AppConfig> | null
}
