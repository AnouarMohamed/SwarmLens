import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import type { Volume } from '../../types'

export function VolumesView() {
  const { volumes, loading } = useClusterStore()

  const columns: Column<Volume>[] = [
    { key: 'name',   header: 'Name',       render: v => <span className="mono">{v.name}</span> },
    { key: 'driver', header: 'Driver',     render: v => <span className="dim">{v.driver}</span>, width: '90px' },
    { key: 'scope',  header: 'Scope',      render: v => <span className="dim">{v.scope}</span>,  width: '80px' },
    { key: 'mount',  header: 'Mountpoint', render: v => <span className="mono dim">{v.mountpoint || '—'}</span> },
  ]

  return (
    <div className="view">
      <div className="view-header">
        <span className="count-label">{volumes.length} volumes</span>
      </div>
      <ResourceTable columns={columns} rows={volumes} keyFn={v => v.name} loading={loading} empty="No volumes." />
    </div>
  )
}
