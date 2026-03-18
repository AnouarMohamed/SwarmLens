import type { ReactNode } from 'react'

export interface Column<T> {
  key: string
  header: string
  render: (row: T) => ReactNode
  width?: string
}

interface Props<T> {
  columns: Column<T>[]
  rows: T[]
  keyFn: (row: T) => string
  onRowClick?: (row: T) => void
  empty?: string
  loading?: boolean
}

export function ResourceTable<T>({ columns, rows, keyFn, onRowClick, empty = 'No results.', loading }: Props<T>) {
  if (loading) {
    return <div className="table-loading">Loading…</div>
  }

  return (
    <div className="table-wrap">
      <table className="resource-table">
        <thead>
          <tr>
            {columns.map(col => (
              <th key={col.key} style={col.width ? { width: col.width } : undefined}>
                {col.header}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.length === 0 ? (
            <tr>
              <td colSpan={columns.length} className="table-empty">{empty}</td>
            </tr>
          ) : (
            rows.map(row => (
              <tr
                key={keyFn(row)}
                onClick={() => onRowClick?.(row)}
                className={onRowClick ? 'row-clickable' : undefined}
              >
                {columns.map(col => (
                  <td key={col.key}>{col.render(row)}</td>
                ))}
              </tr>
            ))
          )}
        </tbody>
      </table>
    </div>
  )
}
