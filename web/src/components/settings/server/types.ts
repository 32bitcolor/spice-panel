import type { AppConfig } from '../../../api/client'

export type ServerAdvancedVariant = 'add' | 'first-run' | 'manage'

export interface ServerAdvancedPanelProps {
  variant: ServerAdvancedVariant
  cfg: AppConfig
  set: (key: keyof AppConfig) => (v: string) => void
  setBool: (key: keyof AppConfig) => (v: boolean) => void
  backendUrl: string
  setBackendUrl: (v: string) => void
  activeName: string
  onRequestDeleteServer?: () => void
}

export interface MarketBotPanelProps {
  cfg: AppConfig
  setBool: (key: keyof AppConfig) => (v: boolean) => void
}

export interface PathsPanelProps {
  cfg: AppConfig
  set: (key: keyof AppConfig) => (v: string) => void
}
