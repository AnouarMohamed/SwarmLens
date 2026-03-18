import { useClusterStore } from '../../store/clusterStore'
import { ResourceTable, type Column } from '../../components/ui/ResourceTable'
import { relativeTime } from '../../lib/utils'
import type { Secret, Config } from '../../types'

export function SecretsView() {
  const { secrets, configs, loading } = useClusterStore()

  const secretColumns: Column<Secret>[] = [
    { key: 'name',    header: 'Name',     render: s => <span className="mono">{s.name}</span> },
    { key: 'created', header: 'Created',  render: s => <span className="dim">{s.createdAt ? relativeTime(s.createdAt) : '—'}</span>, width: '100px' },
    { key: 'refs',    header: 'Used by',  render: s => <span className="dim">{s.serviceRefs?.join(', ') || '—'}</span> },
    { key: 'value',   header: 'Value',    render: () => <span className="dim italic">encrypted · never exposed</span> },
  ]

  const configColumns: Column<Config>[] = [
    { key: 'name',    header: 'Name',     render: c => <span className="mono">{c.name}</span> },
    { key: 'created', header: 'Created',  render: c => <span className="dim">{c.createdAt ? relativeTime(c.createdAt) : '—'}</span>, width: '100px' },
    { key: 'refs',    header: 'Used by',  render: c => <span className="dim">{c.serviceRefs?.join(', ') || '—'}</span> },
    { key: 'note',    header: 'Note',     render: () => <span className="dim italic">not encrypted at rest</span> },
  ]

  return (
    <div className="view">
      <section className="section">
        <h2 className="section-title">Secrets <span className="count-label">{secrets.length}</span></h2>
        <ResourceTable columns={secretColumns} rows={secrets} keyFn={s => s.id} loading={loading} empty="No secrets." />
      </section>
      <section className="section">
        <h2 className="section-title">Configs <span className="count-label">{configs.length}</span></h2>
        <ResourceTable columns={configColumns} rows={configs} keyFn={c => c.id} loading={loading} empty="No configs." />
      </section>
    </div>
  )
}
