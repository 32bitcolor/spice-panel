import * as React from 'react'
import { Avatar, CloseButton, SearchField, Spinner } from '@heroui/react'
import { useTranslation } from 'react-i18next'
import { api } from '../api/client'
import type { DiscordMember } from '../api/client'

type DiscordMemberPickerProps = {
  /** Comma-separated Discord user IDs (the config field format). */
  value: string
  onChange: (v: string) => void
  ariaLabel: string
}

const splitIDs = (v: string) => v.split(',').map((s) => s.trim()).filter(Boolean)

// DiscordMemberPicker edits a comma-separated Discord user ID list by
// searching guild members (bot-token REST — works without the event bot).
// Selected members render as removable chips; names are remembered for IDs
// picked this session, raw IDs shown otherwise.
export const DiscordMemberPicker: React.FC<DiscordMemberPickerProps> = ({ value, onChange, ariaLabel }) => {
  const { t } = useTranslation()
  const [query, setQuery] = React.useState('')
  const [results, setResults] = React.useState<DiscordMember[]>([])
  const [searching, setSearching] = React.useState(false)
  const [open, setOpen] = React.useState(false)
  const [names, setNames] = React.useState<Record<string, string>>({})
  const debounceRef = React.useRef<ReturnType<typeof setTimeout> | null>(null)

  const ids = splitIDs(value)

  const search = (q: string) => {
    setQuery(q)
    setOpen(true)
    if (debounceRef.current) clearTimeout(debounceRef.current)
    if (!q.trim()) {
      setResults([])
      return
    }
    debounceRef.current = setTimeout(async () => {
      setSearching(true)
      try {
        setResults(await api.discord.membersSearch(q.trim()))
      }
      catch {
        setResults([])
      }
      finally {
        setSearching(false)
      }
    }, 300)
  }

  const add = (m: DiscordMember) => {
    if (!ids.includes(m.id)) {
      onChange([...ids, m.id].join(','))
    }
    setNames((n) => ({ ...n, [m.id]: m.name }))
    setQuery('')
    setResults([])
    setOpen(false)
  }

  const remove = (id: string) => {
    onChange(ids.filter((x) => x !== id).join(','))
  }

  return (
    <div className="flex flex-col gap-2">
      <div
        className="relative"
        onBlur={(e) => {
          if (!e.currentTarget.contains(e.relatedTarget as Node | null)) setOpen(false)
        }}
      >
        <SearchField
          value={query}
          onChange={search}
          onFocus={() => setOpen(true)}
          className="w-full"
          aria-label={ariaLabel}
        >
          <SearchField.Group>
            <SearchField.SearchIcon />
            <SearchField.Input
              placeholder={t('settings.auth.memberSearchPlaceholder')}
              aria-label={ariaLabel}
              onKeyDown={(e) => {
                if (e.key === 'Escape') setOpen(false)
              }}
            />
            {searching ? <Spinner size="sm" /> : <SearchField.ClearButton />}
          </SearchField.Group>
        </SearchField>
        {open && results.length > 0 && (
          <div className="absolute z-50 mt-1 w-full max-h-56 overflow-y-auto rounded-[var(--radius)] border border-border bg-surface shadow-lg">
            {results.map((m) => (
              <button
                key={m.id}
                type="button"
                className="w-full flex items-center gap-2 px-2 py-1.5 text-left hover:bg-surface-secondary"
                onMouseDown={(e) => {
                  e.preventDefault()
                  add(m)
                }}
              >
                <Avatar size="sm" className="size-6 shrink-0">
                  {m.avatar && <Avatar.Image src={m.avatar} alt={m.name} />}
                  <Avatar.Fallback>{m.name.slice(0, 2).toUpperCase()}</Avatar.Fallback>
                </Avatar>
                <span className="text-sm text-foreground">{m.name}</span>
                <span className="text-xs text-muted font-mono ml-auto">{m.id}</span>
              </button>
            ))}
          </div>
        )}
      </div>
      {ids.length > 0 && (
        <div className="flex flex-wrap gap-1">
          {ids.map((id) => (
            <span key={id} className="inline-flex items-center gap-1 rounded-full bg-accent/15 text-accent px-2 py-0.5 text-xs font-medium">
              {names[id] ?? id}
              <CloseButton
                aria-label={t('common.remove')}
                className="size-4 opacity-60 hover:opacity-100"
                onPress={() => remove(id)}
              />
            </span>
          ))}
        </div>
      )}
    </div>
  )
}
