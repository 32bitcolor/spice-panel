import * as React from 'react'
import { Select } from '../ui'
import type { FieldSelectProps } from './interfaces'

// FieldSelect wraps the spice-panel Select for small, fixed option sets.
// For large lists (e.g. 400 IANA timezones), keep native <select> for type-to-search.
export const FieldSelect: React.FC<FieldSelectProps> = ({
  value,
  onChange,
  options,
  className,
  ariaLabel,
  isDisabled,
}): React.ReactElement => (
  <Select
    value={value}
    onChange={onChange}
    options={options.map((opt) => ({ value: opt, label: opt }))}
    {...(className !== undefined ? { className } : {})}
    {...(ariaLabel !== undefined ? { 'aria-label': ariaLabel } : {})}
    {...(isDisabled !== undefined ? { isDisabled } : {})}
  />
)
