import type { Status, WebInterface } from '../../../api/client'
import type { BGInfo, ServerRow } from '../types'

export type HealthProps = { bg?: BGInfo, servers: ServerRow[], status: Status | null }

export type InterfaceRowProps = {
  item: WebInterface
}

export type DirectorRowProps = {
  directorURL: string
}
