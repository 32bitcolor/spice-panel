import * as React from 'react'
import {
  SearchField as AriaSearchField,
  Input as AriaInput,
  Button as AriaButton,
  Label,
} from 'react-aria-components'
import type { SearchFieldProps as AriaSearchFieldProps } from 'react-aria-components'
import { cn } from './lib/cn'

export interface SearchFieldProps extends Omit<AriaSearchFieldProps, 'className' | 'children'> {
  label?: string
  placeholder?: string
  className?: string
  /** HeroUI-era visual variant — accepted for compatibility; styling is unified. */
  variant?: string
  /** Compound children (Group/SearchIcon/Input/ClearButton). Omit for the default layout. */
  children?: React.ReactNode
}

const SearchFieldRoot: React.FC<SearchFieldProps> = ({
  label,
  placeholder = 'Search…',
  className,
  variant: _variant,
  children,
  ...props
}): React.ReactElement => (
  <AriaSearchField {...props} className={cn('group flex flex-col gap-1.5', className)}>
    {renderLabel(label)}
    {children ?? (
      <Group>
        <SearchIcon />
        <SearchInput placeholder={placeholder} />
        <ClearButton />
      </Group>
    )}
  </AriaSearchField>
)

const renderLabel = (label: string | undefined): React.ReactNode => {
  if (label === undefined) return null
  return (
    <Label className="font-mono text-[11px] uppercase tracking-[0.22em] text-muted">{label}</Label>
  )
}

/* ── Compound slots (HeroUI-compatible) ───────────────────────────────────── */

const Group: React.FC<React.HTMLAttributes<HTMLDivElement>> = ({
  className,
  children,
  ...props
}): React.ReactElement => (
  <div {...props} className={cn('hud-field flex items-center gap-2 px-3 group-data-[focus-within]:hud-glow', className)}>
    {children}
  </div>
)

const SearchIcon: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <svg viewBox="0 0 16 16" className={cn('h-4 w-4 shrink-0 text-muted', className)} fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round">
    <circle cx="7" cy="7" r="4.5" />
    <path d="m13 13-2.6-2.6" />
  </svg>
)

const SearchInput: React.FC<React.ComponentProps<typeof AriaInput>> = ({
  className,
  ...props
}): React.ReactElement => (
  <AriaInput
    {...props}
    className={cn('w-full bg-transparent py-2 font-mono text-[13px] text-foreground outline-none placeholder:text-muted/70', className as string)}
  />
)

const ClearButton: React.FC<{ className?: string }> = ({ className }): React.ReactElement => (
  <AriaButton className={cn('grid h-4 w-4 shrink-0 place-items-center text-muted outline-none transition hover:text-foreground group-data-[empty]:hidden', className)}>
    <svg viewBox="0 0 16 16" className="h-3 w-3" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round">
      <path d="m4 4 8 8M12 4l-8 8" />
    </svg>
  </AriaButton>
)

export const SearchField = Object.assign(SearchFieldRoot, {
  Group,
  SearchIcon,
  Input: SearchInput,
  ClearButton,
})
