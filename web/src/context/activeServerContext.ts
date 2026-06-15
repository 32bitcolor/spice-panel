import * as React from 'react'

export type ActiveServerContextValue = {
  servers: { id: string, name: string, active: boolean }[]
  activeID: string
  /** True until the first server-list fetch resolves (avoids an empty-state flash). */
  loading: boolean
  setActive: (id: string) => Promise<void>
  removeServer: (id: string) => Promise<void>
  refresh: () => Promise<void>
}

export const ActiveServerContext = React.createContext<ActiveServerContextValue>({
  servers: [],
  activeID: '',
  loading: true,
  setActive: async () => {},
  removeServer: async () => {},
  refresh: async () => {},
})
