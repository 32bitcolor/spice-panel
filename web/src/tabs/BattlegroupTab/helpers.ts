import type { ChipColor } from './types'

/**
 * Map a server / battlegroup phase string to a CSS color from our semantic
 * tokens. Used for the inline-text phase label in the InfoCard and the
 * Phase column of the servers table.
 */
export const phaseColor = (phase: string): string => {
  switch (phase?.toLowerCase()) {
    case 'running':
    case 'reconciling':
    case 'ready':
    case 'connected':
    case 'healthy': return 'var(--success)'
    case 'starting':
    case 'initializing': return 'var(--warning)'
    case 'stopping':
    case 'preshutdown':
    case 'terminating': return 'var(--danger)'
    case 'stopped':
    case 'terminated': return 'var(--muted)'
    default: return 'var(--muted)'
  }
}

export type { ChipColor }

/**
 * Map a phase string to a HeroUI Chip colour (the chip variant of [[phaseColor]]).
 * Used for the Server Health status chips and component-health rows.
 */
export const phaseChipColor = (phase: string): ChipColor => {
  switch (phase?.toLowerCase()) {
    case 'running':
    case 'reconciling':
    case 'ready':
    case 'connected':
    case 'healthy': return 'success'
    case 'starting':
    case 'initializing': return 'warning'
    case 'stopping':
    case 'preshutdown':
    case 'terminating':
    case 'disconnected': return 'danger'
    default: return 'default'
  }
}

/** BG uptime = the oldest running game process's age (0 when unknown). */
export const bgUptimeSeconds = (servers: { ageSeconds?: number }[]): number => {
  return servers.reduce((max, s) => Math.max(max, s.ageSeconds ?? 0), 0)
}

// Phases that mean the battlegroup is down or shutting down — readiness is false
// regardless of (possibly stale) per-server flags.
const DOWN_PHASES = new Set(['stopped', 'stopping', 'terminating', 'terminated', 'preshutdown'])

/**
 * Game is "ready" when every server reports ready and the battlegroup isn't in a
 * down phase. The per-server `ready` flag is authoritative; the battlegroup phase
 * is only used to exclude down states. The phase is NOT gated to "Running" — a
 * live battlegroup reports "Healthy", "Running" or "Reconciling" interchangeably,
 * and gating on "Running" wrongly showed "Not Ready" for the others (#200/#203).
 */
export const allServersReady = (phase: string | undefined, servers: { ready: boolean }[]): boolean => {
  const down = !!phase && DOWN_PHASES.has(phase.toLowerCase())
  return servers.length > 0 && !down && servers.every((s) => s.ready)
}
