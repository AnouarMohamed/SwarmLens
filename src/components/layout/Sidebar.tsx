import { NavLink } from 'react-router-dom'
import { useClusterStore } from '../../store/clusterStore'

const NAV = [
  { to: '/',           label: 'Overview',        icon: '◈' },
  { to: '/stacks',     label: 'Stacks',           icon: '⊞' },
  { to: '/services',   label: 'Services',         icon: '⬡' },
  { to: '/tasks',      label: 'Tasks',            icon: '◧' },
  { to: '/nodes',      label: 'Nodes',            icon: '⬢' },
  { to: '/networks',   label: 'Networks',         icon: '◎' },
  { to: '/volumes',    label: 'Volumes',          icon: '▣' },
  { to: '/secrets',    label: 'Secrets / Configs', icon: '⊕' },
  { to: '/diagnostics',label: 'Diagnostics',      icon: '⚑' },
  { to: '/incidents',  label: 'Incidents',        icon: '⚠' },
  { to: '/audit',      label: 'Audit Trail',      icon: '≡' },
]

export function Sidebar() {
  const swarm = useClusterStore(s => s.swarm)
  const mode = swarm?.mode ?? 'demo'

  return (
    <aside className="sidebar">
      <div className="brand">
        <span className="brand-icon">⬡</span>
        <div>
          <div className="brand-name">SwarmLens</div>
          <span className={`mode-badge mode-${mode}`}>{mode}</span>
        </div>
      </div>

      <nav className="nav">
        {NAV.map(({ to, label, icon }) => (
          <NavLink
            key={to}
            to={to}
            end={to === '/'}
            className={({ isActive }) => 'nav-item' + (isActive ? ' active' : '')}
          >
            <span className="nav-icon">{icon}</span>
            <span className="nav-label">{label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="sidebar-footer">
        <span className="pulse-dot" />
        <span className="cluster-stat">
          {swarm ? `${swarm.managers}m · ${swarm.workers}w` : 'connecting…'}
        </span>
        {swarm && !swarm.quorumHealthy && (
          <span className="quorum-warn" title="Quorum at risk">⚠</span>
        )}
      </div>
    </aside>
  )
}
