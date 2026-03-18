import { useState } from 'react'
import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { StatusBadge } from '../../components/ui/Badge'
import type { Service } from '../../types'

export function ServicesView() {
  const { services, stacks, loading } = useClusterStore()
  const [stackFilter, setStackFilter] = useState('')

  const stackNames = ['', ...Array.from(new Set(services.map(s => s.stack).filter(Boolean)))]
  const filtered = stackFilter ? services.filter(s => s.stack === stackFilter) : services

  const columns: Column<Service>[] = [
    {
      key: 'name', header: 'Service',
      render: s => (
        <div>
          <span className="mono fw-medium">{s.name}</span>
          {s.stack && <span className="dim mono"> ({s.stack})</span>}
        </div>
      ),
    },
    {
      key: 'image', header: 'Image',
      render: s => <span className="mono dim">{s.image.split('/').pop()}</span>,
    },
    {
      key: 'replicas', header: 'Replicas',
      render: s => (
        <span className={s.runningTasks < s.desiredReplicas ? 'text-bad' : ''}>
          {s.runningTasks}/{s.desiredReplicas}
        </span>
      ),
      width: '90px',
    },
    {
      key: 'update', header: 'Update state',
      render: s => s.updateState
        ? <StatusBadge status={s.updateState} />
        : <span className="dim">—</span>,
      width: '140px',
    },
    {
      key: 'ports', header: 'Ports',
      render: s => s.publishedPorts.length > 0
        ? <span className="mono dim">{s.publishedPorts.map(p => p.publishedPort).join(', ')}</span>
        : <span className="dim">—</span>,
      width: '100px',
    },
    {
      key: 'mode', header: 'Mode',
      render: s => <span className="dim">{s.mode}</span>,
      width: '90px',
    },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{filtered.length} services</span>
        <div className="filter-row">
          <label className="filter-label">Stack</label>
          <select
            className="filter-select"
            value={stackFilter}
            onChange={e => setStackFilter(e.target.value)}
          >
            {stackNames.map(n => (
              <option key={n} value={n}>{n || 'All stacks'}</option>
            ))}
          </select>
        </div>
      </div>
      <ResourceTable
        columns={columns}
        rows={filtered}
        keyFn={s => s.id}
        loading={loading}
        empty="No services found."
      />
    </div>
  )
}
