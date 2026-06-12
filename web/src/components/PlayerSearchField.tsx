import * as React from 'react'
import { useTranslation } from 'react-i18next'
import { SearchField, toast } from '@heroui/react'
import { api } from '../api/client'
import type { Player } from '../api/client'
import { useDebounce } from '../hooks/useDebounce'

export interface PlayerSearchFieldProps {
  /** Called with the full player row — consumers pick the ID they need
   *  (account_id, fls_id, or actor id). */
  onSelect: (player: Player) => void
  ariaLabel: string
  placeholder?: string
  className?: string
  /** Pre-loaded player list. When omitted the field lazily loads
   *  api.players.list() on first focus. */
  players?: Player[]
  /** Max suggestions rendered (default 10) — keeps the dropdown cheap even
   *  with thousands of players. */
  resultLimit?: number
  /** Exclude players from suggestions (e.g. the current player). */
  filter?: (player: Player) => boolean
  /** Clear the input after picking (default: show the picked name). */
  clearOnSelect?: boolean
  /** Called when the user empties the input (clear button or deleting the
   *  text) — lets consumers drop their current selection. */
  onClear?: () => void
}

/**
 * Debounced player search with a capped suggestion dropdown. The canonical
 * "select a player" control — renders at most `resultLimit` rows so the full
 * roster never hits the DOM.
 */
export const PlayerSearchField: React.FC<PlayerSearchFieldProps> = ({
  onSelect,
  ariaLabel,
  placeholder,
  className,
  players,
  resultLimit = 10,
  filter,
  clearOnSelect = false,
  onClear,
}) => {
  const { t } = useTranslation()
  const [query, setQuery] = React.useState('')
  const [open, setOpen] = React.useState(false)
  const [loaded, setLoaded] = React.useState<Player[] | null>(null)
  const [loading, setLoading] = React.useState(false)
  const debouncedQuery = useDebounce(query)

  const roster = React.useMemo(() => players ?? loaded ?? [], [players, loaded])

  const ensureLoaded = () => {
    if (players || loaded || loading) return
    setLoading(true)
    api.players
      .list()
      .then(setLoaded)
      .catch((e: unknown) => {
        toast.danger(t('playerSearch.loadFailed', { message: e instanceof Error ? e.message : String(e) }))
      })
      .finally(() => setLoading(false))
  }

  const matches = React.useMemo(() => {
    const base = filter ? roster.filter(filter) : roster
    const q = debouncedQuery.trim().toLowerCase()
    const hits = q
      ? base.filter((p) => p.name.toLowerCase().includes(q) || String(p.account_id).includes(q))
      : base
    return hits.slice(0, resultLimit)
  }, [roster, filter, debouncedQuery, resultLimit])

  const pick = (p: Player) => {
    setQuery(clearOnSelect ? '' : p.name)
    setOpen(false)
    onSelect(p)
  }

  return (
    <div
      className={`relative ${className ?? ''}`}
      onBlur={(e) => {
        if (!e.currentTarget.contains(e.relatedTarget as Node | null)) setOpen(false)
      }}
    >
      <SearchField
        value={query}
        onChange={(v) => {
          setQuery(v)
          setOpen(true)
          if (v === '') onClear?.()
        }}
        onFocus={() => {
          ensureLoaded()
          setOpen(true)
        }}
        className="w-full"
        aria-label={ariaLabel}
      >
        <SearchField.Group>
          <SearchField.SearchIcon />
          <SearchField.Input
            placeholder={loading ? t('playerSearch.loading') : (placeholder ?? t('playerSearch.placeholder'))}
            aria-label={ariaLabel}
            onKeyDown={(e) => {
              if (e.key === 'Escape') setOpen(false)
            }}
          />
          <SearchField.ClearButton />
        </SearchField.Group>
      </SearchField>
      {open && matches.length > 0 && (
        <div className="absolute z-50 w-full mt-1 rounded-[var(--radius)] border border-border bg-surface overflow-y-auto max-h-72 shadow-lg">
          {matches.map((p) => (
            <button
              key={p.account_id}
              type="button"
              className="w-full text-left px-3 py-1.5 text-xs cursor-pointer hover:bg-surface-hover flex items-center justify-between gap-2"
              onMouseDown={(e) => {
                e.preventDefault()
                pick(p)
              }}
            >
              <span className="font-medium">{p.name}</span>
              <span className="text-muted font-mono">
                #
                {p.account_id}
                {' · '}
                {p.online_status}
              </span>
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
