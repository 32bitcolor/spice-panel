import * as React from 'react'
import {
  Table as AriaTable,
  TableHeader,
  TableBody,
  Column as AriaColumn,
  Row as AriaRow,
  Cell as AriaCell,
} from 'react-aria-components'
import type { SortDescriptor, Selection } from 'react-aria-components'
import { cn } from './lib/cn'

/** RAC's ColumnSize accepts px numbers and `%`/`fr` template strings. */
export type ColumnWidth = React.ComponentProps<typeof AriaColumn>['width']

export interface TableColumn<K extends string> {
  key: K
  label: string
  sortable?: boolean
  isRowHeader?: boolean
  width?: ColumnWidth
  align?: 'start' | 'center' | 'end'
}

export interface TableProps<T, K extends string> {
  'aria-label': string
  columns: readonly TableColumn<K>[]
  rows: readonly T[]
  rowId: (row: T) => string
  renderCell: (row: T, key: K) => React.ReactNode
  sortDescriptor?: SortDescriptor
  onSortChange?: (descriptor: SortDescriptor) => void
  selectionMode?: 'none' | 'single' | 'multiple'
  selectedKeys?: Selection
  onSelectionChange?: (keys: Selection) => void
  onRowAction?: (row: T) => void
  emptyState?: React.ReactNode
  className?: string
}

const alignClass = (align: TableColumn<string>['align']): string =>
  align === 'end' ? 'text-right' : align === 'center' ? 'text-center' : 'text-left'

export const Table = <T, K extends string>({
  'aria-label': ariaLabel,
  columns,
  rows,
  rowId,
  renderCell,
  sortDescriptor,
  onSortChange,
  selectionMode = 'none',
  selectedKeys,
  onSelectionChange,
  onRowAction,
  emptyState,
  className,
}: TableProps<T, K>): React.ReactElement => (
  <div className={cn('w-full overflow-x-auto', className)}>
    <AriaTable
      aria-label={ariaLabel}
      selectionMode={selectionMode}
      className="w-full border-collapse"
      {...(sortDescriptor === undefined ? {} : { sortDescriptor })}
      {...(onSortChange === undefined ? {} : { onSortChange })}
      {...(selectedKeys === undefined ? {} : { selectedKeys })}
      {...(onSelectionChange === undefined ? {} : { onSelectionChange })}
    >
      <TableHeader className="border-b border-border bg-surface-tertiary">
        {columns.map((col) => (
          <AriaColumn
            key={col.key}
            id={col.key}
            isRowHeader={col.isRowHeader ?? false}
            allowsSorting={col.sortable !== false}
            {...(col.width === undefined ? {} : { width: col.width })}
            className={cn(
              'group cursor-default px-3.5 py-3 font-mono text-[10.5px] font-medium uppercase tracking-[0.18em] text-muted outline-none',
              alignClass(col.align),
            )}
          >
            <span className="inline-flex items-center gap-1">
              {col.label}
              <span className="text-accent opacity-0 group-aria-[sort=ascending]:opacity-100 group-aria-[sort=descending]:opacity-100 group-aria-[sort=descending]:rotate-180">
                ▲
              </span>
            </span>
          </AriaColumn>
        ))}
      </TableHeader>
      <TableBody
        items={rows.map((r) => ({ ...(r as object), __id: rowId(r), __row: r }) as { __id: string; __row: T })}
        {...(emptyState === undefined ? {} : { renderEmptyState: () => emptyState })}
      >
        {(item) => (
          <AriaRow
            id={item.__id}
            {...(onRowAction === undefined ? {} : { onAction: () => onRowAction(item.__row) })}
            className={cn(
              'border-b border-border/50 text-[13px] text-foreground outline-none transition',
              'data-[hovered]:bg-accent/6 data-[selected]:bg-accent/10',
              onRowAction === undefined ? '' : 'cursor-pointer',
            )}
          >
            {columns.map((col) => (
              <AriaCell key={col.key} className={cn('px-3.5 py-3 align-middle', alignClass(col.align))}>
                {renderCell(item.__row, col.key)}
              </AriaCell>
            ))}
          </AriaRow>
        )}
      </TableBody>
    </AriaTable>
  </div>
)
