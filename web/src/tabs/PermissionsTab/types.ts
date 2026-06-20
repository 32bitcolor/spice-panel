import type { PermissionsData } from '../../api/client'

export type CapabilityGridProps = {
  capabilities: PermissionsData['capabilities']
  selected: string[]
  inherited?: string[]
  onToggle: (cap: string, on: boolean) => void
}
