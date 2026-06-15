import * as React from 'react'
import { api, getActiveServerID, setActiveServerID } from '../api/client'
import type { ServerInfo } from '../api/client'
import { ActiveServerContext } from './activeServerContext'
import type { ActiveServerContextValue } from './activeServerContext'

export const ActiveServerProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [servers, setServers] = React.useState<ServerInfo[]>([])
  const [activeID, setActiveID] = React.useState(getActiveServerID)

  const refresh = React.useCallback(async () => {
    const list = await api.servers.list().catch(() => [] as ServerInfo[])
    setServers(list)
    // Self-heal a stale persisted active server. localStorage is per-origin, so
    // the Vite dev server (:5173) and the embedded SPA (:8080) keep separate
    // copies; a deleted/recreated server leaves an id that no longer exists in
    // the registry. Without this, every request keeps sending a dead
    // X-Dune-Server header and the backend rejects it with 404.
    const current = getActiveServerID()
    if (current && !list.some((s) => s.id === current)) {
      const fallback = list.find((s) => s.active)?.id ?? ''
      setActiveServerID(fallback)
      setActiveID(fallback)
    }
  }, [])

  React.useEffect(() => {
    void Promise.resolve().then(refresh)
  }, [refresh])

  const setActive = React.useCallback(async (id: string) => {
    await api.servers.setActive(id)
    setActiveServerID(id)
    setActiveID(id)
    setServers((prev) => prev.map((s) => ({ ...s, active: s.id === id })))
  }, [])

  const removeServer = React.useCallback(async (id: string) => {
    await api.servers.remove(id)
    // Refetch the authoritative list and reconcile the active id. Deleting the
    // active server (backend reassigns active) or the last server (registry
    // empties → setup) both resolve here, so callers don't special-case them.
    await refresh()
  }, [refresh])

  const value = React.useMemo<ActiveServerContextValue>(
    () => ({ servers, activeID, setActive, removeServer, refresh }),
    [servers, activeID, setActive, removeServer, refresh],
  )

  return <ActiveServerContext.Provider value={value}>{children}</ActiveServerContext.Provider>
}
