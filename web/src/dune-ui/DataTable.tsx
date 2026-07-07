import * as React from 'react'
import type { SortDescriptor } from 'react-aria-components'
import { Table, Pagination, Skeleton, cn } from '../ui'
import type { TableColumn } from '../ui'
import type { Column, DataTableProps } from './types'

export type { Column }

const coerce = (v: React.ReactNode): string | number => {
  if (typeof v === 'number') return v
  if (typeof v === 'string') return v
  return String(v ?? '')
}

const sortRows = <T, K extends string>(
  rows: readonly T[],
  sort: SortDescriptor | undefined,
  getVal: (row: T, key: K) => string | number,
): readonly T[] => {
  if (sort === undefined) return rows
  const key = sort.column as K
  const dir = sort.direction === 'descending' ? -1 : 1
  return [...rows].sort((a, b) => {
    const av = getVal(a, key)
    const bv = getVal(b, key)
    if (typeof av === 'number' && typeof bv === 'number') return (av - bv) * dir
    return String(av).localeCompare(String(bv), undefined, { numeric: true }) * dir
  })
}

export const DataTable = <T extends object, K extends string>({
  'aria-label': ariaLabel,
  columns,
  rows,
  rowId,
  renderCell,
  initialSort,
  sortValue,
  emptyState,
  loading = false,
  skeletonRows = 5,
  onRowAction,
  className,
  selectionMode,
  selectedKeys,
  onSelectionChange,
  pageSize,
}: DataTableProps<T, K>): React.ReactElement => {
  const [sort, setSort] = React.useState<SortDescriptor | undefined>(
    initialSort ? { column: initialSort.column, direction: initialSort.direction } : undefined,
  )
  const [page, setPage] = React.useState(1)
  // Reset to page 1 when the row set changes — React's "adjust state during
  // render" pattern (no effect needed; avoids a stale extra-page flash).
  const [prevRows, setPrevRows] = React.useState(rows)
  if (rows !== prevRows) {
    setPrevRows(rows)
    setPage(1)
  }

  const getSortVal = (row: T, key: K): string | number => {
    if (sortValue) return coerce(sortValue(row, key))
    return coerce(renderCell(row, key))
  }

  // React Compiler memoizes this pure derivation — no manual useMemo needed.
  const sortedRows = sortRows(rows, sort, getSortVal)

  const totalPages = pageSize ? Math.ceil(sortedRows.length / pageSize) : 1
  const pagedRows = pageSize ? sortedRows.slice((page - 1) * pageSize, page * pageSize) : sortedRows

  const tableColumns: TableColumn<K>[] = columns.map((col) => ({
    key: col.key,
    label: col.label,
    sortable: col.sortable !== false,
    isRowHeader: col.isRowHeader ?? false,
    align: col.align ?? 'start',
    ...(col.width !== undefined ? { width: col.width } : {}),
  }))

  if (loading) return renderSkeleton(columns, skeletonRows, className)

  const table = (
    <Table<T, K>
      aria-label={ariaLabel}
      columns={tableColumns}
      rows={pagedRows}
      rowId={rowId}
      renderCell={renderCell}
      sortDescriptor={sort ?? { column: '', direction: 'ascending' }}
      onSortChange={setSort}
      selectionMode={selectionMode ?? 'none'}
      {...(selectedKeys !== undefined ? { selectedKeys } : {})}
      {...(onSelectionChange !== undefined ? { onSelectionChange } : {})}
      {...(onRowAction !== undefined ? { onRowAction } : {})}
      {...(emptyState !== undefined ? { emptyState } : {})}
      {...(className !== undefined ? { className } : {})}
    />
  )

  if (!pageSize) return table

  return (
    <div className="flex h-full min-h-0 flex-col gap-2">
      <div className="min-h-0 flex-1 overflow-auto">{table}</div>
      {renderPager(page, totalPages, pageSize, sortedRows.length, setPage)}
    </div>
  )
}

const renderSkeleton = <K extends string>(
  columns: Column<K>[],
  skeletonRows: number,
  className: string | undefined,
): React.ReactElement => (
  <div className={cn('overflow-hidden ring-1 ring-inset ring-border/60 [border-radius:var(--radius)]', className)}>
    {Array.from({ length: skeletonRows }, (_, i) => (
      <div key={i} className="flex gap-3 border-b border-border/40 px-3 py-2.5 last:border-0">
        {columns.map((c) => (
          <Skeleton key={c.key} className="h-3.5 flex-1" />
        ))}
      </div>
    ))}
  </div>
)

const renderPager = (
  page: number,
  totalPages: number,
  pageSize: number,
  totalRows: number,
  setPage: (p: number) => void,
): React.ReactNode => {
  if (totalPages <= 1) return null
  return (
    <div className="flex shrink-0 items-center justify-between px-1 py-1">
      <span className="whitespace-nowrap text-xs tabular-nums text-muted">
        {(page - 1) * pageSize + 1}
        {' '}
        –
        {Math.min(page * pageSize, totalRows)}
        {' '}
        of
        {totalRows}
      </span>
      <Pagination page={page} total={totalPages} onChange={setPage} className="ml-auto" />
    </div>
  )
}
