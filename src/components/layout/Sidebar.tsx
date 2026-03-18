import type { SVGProps } from 'react'
import { NavLink } from 'react-router-dom'
import { relativeTime } from '../../lib/utils'
import { useClusterStore } from '../../store/clusterStore'
import {
  AuditIcon,
  DiagnosticsIcon,
  IncidentIcon,
  NetworkIcon,
  NodeIcon,
  OverviewIcon,
  SecretIcon,
  ServiceIcon,
  StackIcon,
  TaskIcon,
  VolumeIcon,
} from '../ui/icons'

interface SidebarProps {
  open: boolean
  onClose: () => void
}

interface NavItem {
  to: string
  label: string
  icon: (props: SVGProps<SVGSVGElement>) => JSX.Element
}

interface NavGroup {
  label: string
  items: NavItem[]
}

const NAV_GROUPS: NavGroup[] = [
  {
    label: 'Cluster',
    items: [{ to: '/', label: 'Overview', icon: OverviewIcon }],
  },
  {
    label: 'Workloads',
    items: [
      { to: '/stacks', label: 'Stacks', icon: StackIcon },
      { to: '/services', label: 'Services', icon: ServiceIcon },
      { to: '/tasks', label: 'Tasks', icon: TaskIcon },
    ],
  },
  {
    label: 'Infrastructure',
    items: [
      { to: '/nodes', label: 'Nodes', icon: NodeIcon },
      { to: '/networks', label: 'Networks', icon: NetworkIcon },
      { to: '/volumes', label: 'Volumes', icon: VolumeIcon },
    ],
  },
  {
    label: 'Security & Config',
    items: [{ to: '/secrets', label: 'Secrets / Configs', icon: SecretIcon }],
  },
  {
    label: 'Operations',
    items: [
      { to: '/diagnostics', label: 'Diagnostics', icon: DiagnosticsIcon },
      { to: '/incidents', label: 'Incidents', icon: IncidentIcon },
      { to: '/audit', label: 'Audit Trail', icon: AuditIcon },
    ],
  },
]

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const swarm = useClusterStore((s) => s.swarm)
  const connectionState = useClusterStore((s) => s.connectionState)
  const error = useClusterStore((s) => s.error)
  const fetchAll = useClusterStore((s) => s.fetchAll)
  const lastRefresh = useClusterStore((s) => s.lastRefresh)

  const mode = (swarm?.mode ?? 'demo').toUpperCase()
  const clusterShort = swarm?.clusterID
    ? `cluster/${swarm.clusterID.slice(0, 16)}`
    : 'cluster/unset'
  const endpoint = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'
  const disconnected = connectionState === 'disconnected' || Boolean(error)
  const statusText = disconnected
    ? 'Disconnected'
    : connectionState === 'connecting'
      ? 'Connecting'
      : 'Connected'
  const freshness = lastRefresh ? relativeTime(new Date(lastRefresh).toISOString()) : 'never synced'

  return (
    <>
      <button
        type="button"
        aria-label="Close navigation drawer"
        onClick={onClose}
        className={cn(
          'fixed inset-0 z-30 bg-black/55 transition-opacity lg:hidden',
          open ? 'opacity-100' : 'pointer-events-none opacity-0',
        )}
      />

      <aside
        className={cn(
          'fixed inset-y-0 left-0 z-40 flex w-[272px] flex-col border-r border-border-muted bg-sidebar transition-transform duration-200 lg:translate-x-0',
          open ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <header className="border-b border-border-muted px-5 pb-5 pt-6">
          <p className="font-heading text-[1.7rem] uppercase leading-none tracking-[0.05em] text-text-primary">
            SwarmLens
          </p>
          <div className="mt-3 flex items-center gap-3">
            <span className="industrial-label text-text-secondary">{mode}</span>
            <span className="font-mono text-[11px] text-text-tertiary">{clusterShort}</span>
          </div>
        </header>

        <nav className="flex-1 space-y-6 overflow-y-auto px-4 py-6" aria-label="Primary navigation">
          {NAV_GROUPS.map((group) => (
            <section key={group.label} aria-labelledby={`nav-group-${group.label}`}>
              <h2 id={`nav-group-${group.label}`} className="industrial-label px-1">
                {group.label}
              </h2>
              <ul className="mt-2 space-y-0.5">
                {group.items.map((item) => {
                  const Icon = item.icon
                  return (
                    <li key={item.to}>
                      <NavLink
                        to={item.to}
                        end={item.to === '/'}
                        onClick={onClose}
                        className={({ isActive }) =>
                          cn(
                            'industrial-nav-link flex items-center gap-2.5 text-[14px] tracking-[0.03em] focus-visible:outline-none',
                            isActive && 'is-active',
                          )
                        }
                      >
                        <Icon className="h-4 w-4 shrink-0 text-text-tertiary" />
                        <span className="truncate">{item.label}</span>
                      </NavLink>
                    </li>
                  )
                })}
              </ul>
            </section>
          ))}
        </nav>

        <footer className="border-t border-border-muted px-5 py-4">
          <div className="flex items-center gap-2">
            <span
              aria-hidden="true"
              className={cn(
                'h-2 w-2 rounded-full',
                disconnected
                  ? 'bg-state-danger'
                  : connectionState === 'connecting'
                    ? 'bg-text-secondary'
                    : 'bg-text-primary',
              )}
            />
            <p className={cn('industrial-label', disconnected && 'text-state-danger')}>
              {statusText}
            </p>
          </div>
          <p className="mt-1 truncate font-mono text-[11px] text-text-tertiary">{clusterShort}</p>
          <p className="mt-2 truncate font-mono text-xs text-text-secondary">{endpoint}</p>
          <p className="mt-1 font-mono text-[11px] text-text-tertiary">Last sync {freshness}</p>

          {disconnected && (
            <button
              type="button"
              onClick={() => {
                void fetchAll()
              }}
              className="industrial-action industrial-action-accent mt-3"
            >
              Reconnect Cluster
            </button>
          )}
        </footer>
      </aside>
    </>
  )
}
