import type { Container } from '../types'

export type { Container }

export type AddResult = { given: string[], skipped: { template: string, reason: string }[] } | null

export type AddStagedItem = { template: string, qty: number, quality: number, _key: string }
