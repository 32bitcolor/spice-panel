import * as React from 'react'
import type { FieldProps } from '../types'

export const Field: React.FC<FieldProps> = ({ label, value }) => (
  <div>
    <div className="text-xs text-muted">{label}</div>
    <div className="text-foreground">{value}</div>
  </div>
)
