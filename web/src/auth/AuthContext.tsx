import * as React from 'react'
import { AUTH_EXPIRED_EVENT, authApi } from '../api/client'
import type { AuthStatus } from '../api/client'
import { AuthContext } from './context'
import type { AuthContextValue } from './context'

const disabledState: AuthStatus = {
  enabled: false,
  methods: { local: false, discord: false, guest: false },
  session: null,
}

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [status, setStatus] = React.useState<AuthStatus>(disabledState)
  const [loading, setLoading] = React.useState(true)
  const [error, setError] = React.useState(false)

  const refresh = async (): Promise<void> => {
    try {
      setStatus(await authApi.status())
      setError(false)
    }
    catch {
      // Backend unreachable — let the existing BackendUnreachable flow handle
      // it; treat auth as disabled so we don't block the error screen.
      setError(true)
    }
    finally {
      setLoading(false)
    }
  }

  React.useEffect(() => {
    void Promise.resolve().then(refresh)
  }, [])

  // Session expired mid-use (cookie TTL, kicked from guild): drop to login.
  React.useEffect(() => {
    const onExpired = () => {
      setStatus((s) => (s.enabled ? { ...s, session: null } : s))
    }
    window.addEventListener(AUTH_EXPIRED_EVENT, onExpired)
    return () => window.removeEventListener(AUTH_EXPIRED_EVENT, onExpired)
  }, [])

  const login = async (username: string, password: string): Promise<void> => {
    await authApi.login(username, password)
    await refresh()
  }

  const logout = async (): Promise<void> => {
    try {
      await authApi.logout()
    }
    finally {
      await refresh()
    }
  }

  const value: AuthContextValue = {
    enabled: status.enabled,
    methods: status.methods,
    session: status.session,
    loading,
    error,
    login,
    logout,
    refresh,
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
