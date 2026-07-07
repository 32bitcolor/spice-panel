import * as React from 'react'
import { cn } from './lib/cn'

type SegmentSize = 'sm' | 'md'

interface SegmentContextValue {
  selectedKey: string | undefined
  onSelect: (key: string) => void
  size: SegmentSize
}

const SegmentContext = React.createContext<SegmentContextValue | null>(null)

const useSegment = (): SegmentContextValue => {
  const ctx = React.useContext(SegmentContext)
  if (ctx === null) throw new Error('Segment.Item must be used within a <Segment>')
  return ctx
}

export interface SegmentProps {
  'selectedKey'?: string
  'defaultSelectedKey'?: string
  'onSelectionChange'?: (key: string) => void
  'size'?: SegmentSize
  /** HeroUI-era visual variant — accepted for compatibility; styling is unified. */
  'variant'?: string
  'className'?: string
  'children'?: React.ReactNode
  'aria-label'?: string
}

const SegmentRoot: React.FC<SegmentProps> = ({
  selectedKey,
  defaultSelectedKey,
  onSelectionChange,
  size = 'md',
  variant: _variant,
  className,
  children,
  'aria-label': ariaLabel,
}): React.ReactElement => {
  const [internal, setInternal] = React.useState(defaultSelectedKey)
  const active = selectedKey ?? internal
  const value: SegmentContextValue = {
    selectedKey: active,
    size,
    onSelect: (key) => {
      if (selectedKey === undefined) setInternal(key)
      onSelectionChange?.(key)
    },
  }
  return (
    <SegmentContext.Provider value={value}>
      <div
        role="tablist"
        aria-label={ariaLabel}
        className={cn(
          'inline-flex bg-[var(--void)] p-0.5 shadow-[inset_0_0_0_1px_var(--steel)] [clip-path:polygon(5px_0,100%_0,100%_calc(100%-5px),calc(100%-5px)_100%,0_100%,0_5px)]',
          className,
        )}
      >
        {children}
      </div>
    </SegmentContext.Provider>
  )
}

export interface SegmentItemProps {
  id: string
  className?: string
  children?: React.ReactNode
}

const Item: React.FC<SegmentItemProps> = ({ id, className, children }): React.ReactElement => {
  const { selectedKey, onSelect, size } = useSegment()
  const selected = selectedKey === id
  return (
    <button
      type="button"
      role="tab"
      aria-selected={selected}
      onClick={() => onSelect(id)}
      className={cn(
        'inline-flex cursor-pointer items-center gap-1.5 font-mono uppercase tracking-[0.06em] outline-none transition [clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)] focus-visible:hud-glow',
        size === 'sm' ? 'px-3 py-1 text-[11px]' : 'px-4 py-1.5 text-xs',
        selected
          ? 'bg-accent font-bold text-accent-foreground'
          : 'text-muted hover:text-foreground',
        className,
      )}
    >
      {children}
    </button>
  )
}

/** Decorative divider slot (kept for API compatibility; renders nothing). */
const SegmentSeparator: React.FC = (): null => null

export const Segment = Object.assign(SegmentRoot, { Item, Separator: SegmentSeparator })
