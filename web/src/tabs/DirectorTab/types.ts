import type { FieldKind } from '../types'

export type { FieldKind }

export interface DirectorEditorProps {
  kind: FieldKind
  value: string
  ariaLabel: string
  onChange: (v: string) => void
}
