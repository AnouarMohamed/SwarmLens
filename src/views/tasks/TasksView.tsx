import { useState } from 'react'
import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { StatusBadge } from '../../components/ui/Badge'
import { relativeTime } from '../../lib/utils'
import type { Task } from '../../types'

export function TasksView() {
  const { tasks, loading } = useClusterStore()
  const [stateFilter, setStateFilter] = useState('')

  const filtered = stateFilter ? tasks.filter(t => t.currentState === stateFilter) : tasks
  const failedCount = tasks.filter(t => t.currentState === 'failed').length

  const columns: Column<Task>[] = [
    {
      key: 'id', header: 'Task ID',
      render: t => <span className="mono dim">{t.id.slice(0, 12)}</span>,
      width: '120px',
    },
    {
      key: 'service', header: 'Service',
      render: t => <span className="mono">{t.serviceName}</span>,
    },
    {
      key: 'node', header: 'Node',
      render: t => <span className="mono dim">{t.nodeHostname || '—'}</span>,
      width: '120px',
    },
    {
      key: 'desired', header: 'Desired',
      render: t => <StatusBadge status={t.desiredState} size="sm" />,
      width: '90px',
    },
    {
      key: 'current', header: 'Current',
      render: t => <StatusBadge status={t.currentState} size="sm" />,
      width: '90px',
    },
    {
      key: 'restarts', header: 'Restarts',
      render: t => (
        <span className={t.restartCount >= 3 ? 'text-bad' : t.restartCount > 0 ? 'text-warn' : ''}>
          {t.restartCount}
        </span>
      ),
      width: '80px',
    },
    {
      key: 'error', header: 'Error',
      render: t => t.error
        ? <span className="mono dim truncate" title={t.error}>{t.error.slice(0, 60)}{t.error.length > 60 ? '…' : ''}</span>
        : <span className="dim">—</span>,
    },
    {
      key: 'updated', header: 'Updated',
      render: t => <span className="dim">{t.updatedAt ? relativeTime(t.updatedAt) : '—'}</span>,
      width: '90px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">
          {filtered.length} tasks
          {failedCount > 0 && <span className="count-bad"> · {failedCount} failed</span>}
        </span>
        <div className="filter-row">
          <label className="filter-label">State</label>
          <select
            className="filter-select"
            value={stateFilter}
            onChange={e => setStateFilter(e.target.value)}
          >
            <option value="">All states</option>
            {['running','failed','pending','shutdown','complete'].map(s => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
      </div>
      <ResourceTable
        columns={columns}
        rows={filtered}
        keyFn={t => t.id}
        loading={loading}
        empty="No tasks found."
      />
    </div>
  )
}
