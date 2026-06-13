import { describe, it, expect } from 'vitest'
import { phaseColor, phaseChipColor, allServersReady } from './helpers'

describe('phaseColor', () => {
  it('treats reconciling as online/healthy (success/green)', () => {
    expect(phaseColor('reconciling')).toBe('var(--success)')
    expect(phaseColor('Reconciling')).toBe('var(--success)')
  })

  it('keeps running as success/green', () => {
    expect(phaseColor('running')).toBe('var(--success)')
  })

  it('treats Healthy (the battlegroup-level online phase) as success/green', () => {
    expect(phaseColor('Healthy')).toBe('var(--success)')
    expect(phaseColor('healthy')).toBe('var(--success)')
  })

  it('keeps starting and initializing as warning', () => {
    expect(phaseColor('starting')).toBe('var(--warning)')
    expect(phaseColor('initializing')).toBe('var(--warning)')
  })
})

describe('phaseChipColor', () => {
  it('treats reconciling as online/healthy (success)', () => {
    expect(phaseChipColor('reconciling')).toBe('success')
    expect(phaseChipColor('Reconciling')).toBe('success')
  })

  it('keeps running as success', () => {
    expect(phaseChipColor('running')).toBe('success')
  })

  it('keeps starting and initializing as warning', () => {
    expect(phaseChipColor('starting')).toBe('warning')
    expect(phaseChipColor('initializing')).toBe('warning')
  })
})

describe('allServersReady', () => {
  const ready = [{ ready: true }]

  it('is ready for any non-down phase when servers report ready (#200/#203)', () => {
    // A live battlegroup reports these interchangeably — all must read as ready.
    expect(allServersReady('Running', ready)).toBe(true)
    expect(allServersReady('Reconciling', ready)).toBe(true)
    expect(allServersReady('Healthy', ready)).toBe(true)
  })

  it('is case-insensitive on the phase', () => {
    expect(allServersReady('healthy', ready)).toBe(true)
    expect(allServersReady('reconciling', ready)).toBe(true)
  })

  it('is ready even with an unknown/empty phase when servers are ready', () => {
    expect(allServersReady(undefined, ready)).toBe(true)
  })

  it('is not ready in a down phase', () => {
    expect(allServersReady('Stopped', ready)).toBe(false)
    expect(allServersReady('Terminating', ready)).toBe(false)
  })

  it('is not ready with no servers', () => {
    expect(allServersReady('Running', [])).toBe(false)
  })

  it('is not ready when a server is not ready', () => {
    expect(allServersReady('Running', [{ ready: true }, { ready: false }])).toBe(false)
  })
})
