import * as React from 'react'
import { Button as AriaButton } from 'react-aria-components'
import { cn } from './lib/cn'

export interface PaginationProps {
  page: number
  total: number
  onChange: (page: number) => void
  /** Number of sibling page buttons around the current page. */
  siblings?: number
  className?: string
}

const range = (start: number, end: number): number[] =>
  Array.from({ length: end - start + 1 }, (_, i) => start + i)

const pages = (page: number, total: number, siblings: number): (number | 'gap')[] => {
  if (total <= siblings * 2 + 5) return range(1, total)
  const left = Math.max(page - siblings, 1)
  const right = Math.min(page + siblings, total)
  const items: (number | 'gap')[] = [1]
  if (left > 2) items.push('gap')
  items.push(...range(Math.max(left, 2), Math.min(right, total - 1)))
  if (right < total - 1) items.push('gap')
  if (total > 1) items.push(total)
  return items
}

const cell
  = 'grid h-8 min-w-8 place-items-center px-2 font-mono text-xs outline-none transition data-[focus-visible]:hud-glow [clip-path:polygon(4px_0,100%_0,100%_calc(100%-4px),calc(100%-4px)_100%,0_100%,0_4px)]'

export const Pagination: React.FC<PaginationProps> = ({
  page,
  total,
  onChange,
  siblings = 1,
  className,
}): React.ReactElement => {
  const go = (p: number): void => {
    if (p >= 1 && p <= total && p !== page) onChange(p)
  }

  return (
    <nav aria-label="Pagination" className={cn('flex items-center gap-1.5', className)}>
      <AriaButton
        aria-label="Previous page"
        isDisabled={page <= 1}
        onPress={() => go(page - 1)}
        className={cn(cell, 'bg-surface-secondary text-muted data-[hovered]:text-foreground data-[disabled]:opacity-30')}
      >
        ‹
      </AriaButton>
      {pages(page, total, siblings).map((p, i) =>
        p === 'gap'
          ? (
              <span key={`gap-${i}`} className="px-1 text-muted">
                …
              </span>
            )
          : (
              <AriaButton
                key={p}
                onPress={() => go(p)}
                {...(p === page ? { 'aria-current': 'page' as const } : {})}
                className={cn(
                  cell,
                  p === page
                    ? 'bg-accent font-bold text-accent-foreground'
                    : 'bg-surface-secondary text-muted data-[hovered]:text-foreground',
                )}
              >
                {p}
              </AriaButton>
            ),
      )}
      <AriaButton
        aria-label="Next page"
        isDisabled={page >= total}
        onPress={() => go(page + 1)}
        className={cn(cell, 'bg-surface-secondary text-muted data-[hovered]:text-foreground data-[disabled]:opacity-30')}
      >
        ›
      </AriaButton>
    </nav>
  )
}
