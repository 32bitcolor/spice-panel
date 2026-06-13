import * as React from 'react'
import { AuthContext } from '../auth/context'

export type UsePermissions = {
  /** Whether backend auth is enabled at all. */
  enabled: boolean
  /** True when the session has the capability (always true when auth is off). */
  can: (capability: string) => boolean
  /** Owners bypass the matrix and see the Permissions tab. */
  isOwner: boolean
}

// usePermissions gates UI elements on the session's capability set.
// With auth disabled (the default) every check passes, preserving the
// pre-auth dashboard behavior byte-for-byte.
export function usePermissions(): UsePermissions {
  const { enabled, session } = React.useContext(AuthContext)
  return React.useMemo(() => {
    if (!enabled) {
      return { enabled, can: () => true, isOwner: true }
    }
    const caps = new Set(session?.capabilities ?? [])
    const isOwner = session?.owner ?? false
    return {
      enabled,
      can: (capability: string) => isOwner || caps.has(capability),
      isOwner,
    }
  }, [enabled, session])
}
