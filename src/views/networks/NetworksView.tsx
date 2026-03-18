import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import type { Network } from '../../types'

export function NetworksView() {
  const { networks, loading } = useClusterStore()

  const columns: Column<Network>[] = [
    { key: 'name',    header: 'Name',    render: n => <span className="mono">{n.name}</span> },
    { key: 'driver',  header: 'Driver',  render: n => <span className="dim">{n.driver}</span>, width: '90px' },
    { key: 'scope',   header: 'Scope',   render: n => <span className="dim">{n.scope}</span>,  width: '80px' },
    { key: 'subnet',  header: 'Subnet',  render: n => <span className="mono dim">{n.subnet || '—'}</span> },
    { key: 'ingress', header: 'Ingress', render: n => n.ingress ? <span className="badge state-running">yes</span> : <span className="dim">—</span>, width: '80px' },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{networks.length} networks</span>
      </div>
      <ResourceTable columns={columns} rows={networks} keyFn={n => n.id} loading={loading} empty="No networks." />
    </div>
  )
}
