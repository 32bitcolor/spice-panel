import * as React from 'react'

export type ActiveServerContextValue = {
  servers: { id: string, name: string, active: boolean }[]
  activeID: string
  setActive: (id: string) => Promise<void>
  removeServer: (id: string) => Promise<void>
  refresh: () => Promise<void>
}

export const ActiveServerContext = React.createContext<ActiveServerContextValue>({
  servers: [],
  activeID: '',
  setActive: async () => {},
  removeServer: async () => {},
  refresh: async () => {},
})
