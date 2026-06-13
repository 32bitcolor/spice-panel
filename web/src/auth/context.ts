import * as React from 'react'
import type { AuthStatus } from '../api/client'

export type AuthContextValue = {
  /** Backend auth feature flag. False → everything is allowed (legacy behavior). */
  enabled: boolean
  /** Which login methods the backend has configured. */
  methods: { local: boolean, discord: boolean, guest: boolean }
  /** Current session, or null when not logged in. */
  session: AuthStatus['session']
  /** True until the first /auth/status response lands. */
  loading: boolean
  /** True when /auth/status itself failed (backend unreachable). */
  error: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refresh: () => Promise<void>
}

export const AuthContext = React.createContext<AuthContextValue>({
  enabled: false,
  methods: { local: false, discord: false, guest: false },
  session: null,
  loading: false,
  error: false,
  login: async () => {},
  logout: async () => {},
  refresh: async () => {},
})
