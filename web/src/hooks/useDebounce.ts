import * as React from 'react'

// Shared debounce hook (canonical copy — the ActionsView-local one predates it).
export const useDebounce = <T>(value: T, delay = 300): T => {
  const [debounced, setDebounced] = React.useState(value)
  React.useEffect(() => {
    const t = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(t)
  }, [value, delay])
  return debounced
}
