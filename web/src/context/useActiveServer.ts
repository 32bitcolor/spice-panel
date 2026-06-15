import * as React from 'react'
import { ActiveServerContext } from './activeServerContext'
import type { ActiveServerContextValue } from './activeServerContext'

export function useActiveServer(): ActiveServerContextValue {
  return React.useContext(ActiveServerContext)
}
