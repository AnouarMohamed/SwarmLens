import { useLocation } from 'react-router-dom'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useClusterStore } from '../../store/clusterStore'

const TITLES: Record<string, string> = {
  '/': 'Overview', '/stacks': 'Stacks', '/services': 'Services',
  '/tasks': 'Tasks', '/nodes': 'Nodes', '/networks': 'Networks',
  '/volumes': 'Volumes', '/secrets': 'Secrets / Configs',
  '/diagnostics': 'Diagnostics', '/incidents': 'Incidents', '/audit': 'Audit Trail',
}

export function Topbar() {
  const loc = useLocation()
  const title = TITLES[loc.pathname] ?? 'SwarmLens'
  const { run, running } = useDiagnosticsStore()
  const fetchAll = useClusterStore(s => s.fetchAll)
  const loading = useClusterStore(s => s.loading)

  return (
    <header className="topbar">
      <h1 className="page-title">{title}</h1>
      <div className="topbar-actions">
        <button className="btn-ghost" onClick={fetchAll} disabled={loading}>
          {loading ? '↻ refreshing…' : '↻ refresh'}
        </button>
        <button className="btn-ghost" onClick={run} disabled={running}>
          {running ? '⚑ running…' : '⚑ diagnostics'}
        </button>
      </div>
    </header>
  )
}
