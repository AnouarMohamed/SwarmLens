import type { SVGProps } from 'react'
import { NavLink, useLocation, useNavigate } from 'react-router-dom'
import { useClusterStore } from '../../store/clusterStore'
import {
  AuditIcon,
  ChevronDownIcon,
  DiagnosticsIcon,
  IncidentIcon,
  InfoIcon,
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

type NavGroup = 'Cluster' | 'Workloads' | 'Infrastructure' | 'Security & Config' | 'Operations'

interface NavItem {
  to: string
  label: string
  group: NavGroup
  description: string
  icon: (props: SVGProps<SVGSVGElement>) => JSX.Element
}

const GROUP_ORDER: NavGroup[] = [
  'Cluster',
  'Workloads',
  'Infrastructure',
  'Security & Config',
  'Operations',
]

const NAV_ITEMS: NavItem[] = [
  {
    to: '/',
    label: 'Overview',
    group: 'Cluster',
    description: 'Health, replica status, and node connectivity in one view.',
    icon: OverviewIcon,
  },
  {
    to: '/diagnostics',
    label: 'Diagnostics',
    group: 'Operations',
    description: 'Run checks, inspect findings, and export current diagnostics.',
    icon: DiagnosticsIcon,
  },
  {
    to: '/incidents',
    label: 'Incidents',
    group: 'Operations',
    description: 'Triage active incidents and coordinate mitigation flow.',
    icon: IncidentIcon,
  },
  {
    to: '/audit',
    label: 'Audit Trail',
    group: 'Operations',
    description: 'Review write operations, actors, and reconciliation history.',
    icon: AuditIcon,
  },
  {
    to: '/assistant',
    label: 'Assistant',
    group: 'Operations',
    description: 'Stream AI-backed triage hypotheses and remediation actions.',
    icon: InfoIcon,
  },
  {
    to: '/stacks',
    label: 'Stacks',
    group: 'Workloads',
    description: 'Inspect stack health and deployment topology.',
    icon: StackIcon,
  },
  {
    to: '/services',
    label: 'Services',
    group: 'Workloads',
    description: 'Track replica drift, rollout state, and restart pressure.',
    icon: ServiceIcon,
  },
  {
    to: '/tasks',
    label: 'Tasks',
    group: 'Workloads',
    description: 'Debug task scheduling failures and container lifecycle signals.',
    icon: TaskIcon,
  },
  {
    to: '/nodes',
    label: 'Nodes',
    group: 'Infrastructure',
    description: 'Monitor manager quorum, worker readiness, and host pressure.',
    icon: NodeIcon,
  },
  {
    to: '/networks',
    label: 'Networks',
    group: 'Infrastructure',
    description: 'Validate overlay network integrity and service attachments.',
    icon: NetworkIcon,
  },
  {
    to: '/volumes',
    label: 'Volumes',
    group: 'Infrastructure',
    description: 'Review persistent volume footprint and mount allocations.',
    icon: VolumeIcon,
  },
  {
    to: '/secrets',
    label: 'Secrets / Configs',
    group: 'Security & Config',
    description: 'Audit secret and config references across workloads.',
    icon: SecretIcon,
  },
]

function cn(...parts: Array<string | undefined | false>) {
  return parts.filter(Boolean).join(' ')
}

export function Sidebar({ open, onClose }: SidebarProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const swarm = useClusterStore((s) => s.swarm)
  const stacks = useClusterStore((s) => s.stacks)
  const nodes = useClusterStore((s) => s.nodes)
  const tasks = useClusterStore((s) => s.tasks)
  const services = useClusterStore((s) => s.services)
  const connectionState = useClusterStore((s) => s.connectionState)
  const error = useClusterStore((s) => s.error)
  const fetchAll = useClusterStore((s) => s.fetchAll)

  const mode = (swarm?.mode ?? 'demo').toUpperCase()
  const clusterShort = swarm?.clusterID ? `cluster/${swarm.clusterID.slice(0, 20)}` : 'cluster/unset'
  const disconnected = connectionState === 'disconnected' || Boolean(error)

  const healthyStacks = stacks.filter((stack) => stack.runningServices >= stack.serviceCount).length
  const readyNodes = nodes.filter((node) => node.state === 'ready').length
  const runningTasks = tasks.filter((task) => task.currentState === 'running').length
  const healthyReplicas = services.reduce((sum, service) => sum + service.runningTasks, 0)
  const desiredReplicas = services.reduce((sum, service) => sum + service.desiredReplicas, 0)

  const ratioInputs = [
    stacks.length ? healthyStacks / stacks.length : 0,
    nodes.length ? readyNodes / nodes.length : 0,
    tasks.length ? runningTasks / tasks.length : 0,
    desiredReplicas > 0 ? healthyReplicas / desiredReplicas : 0,
  ]
  const availableSignals = ratioInputs.filter((value) => value > 0)
  const healthRatio = availableSignals.length
    ? Math.round((availableSignals.reduce((sum, value) => sum + value, 0) / availableSignals.length) * 100)
    : 0

  const activeItem = NAV_ITEMS.find((item) => {
    if (item.to === '/') return location.pathname === '/'
    return location.pathname.startsWith(item.to)
  })
  const operatorTip =
    activeItem?.description ??
    'Use keyboard-driven navigation to jump between diagnostics, incidents, and audit.'

  const groupedItems = GROUP_ORDER.map((group) => ({
    label: group,
    items: NAV_ITEMS.filter((item) => item.group === group),
  })).filter((group) => group.items.length > 0)

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
          'fixed inset-y-0 left-0 z-40 flex w-[256px] flex-col border-r border-white/10 bg-sidebar transition-transform duration-200 lg:translate-x-0',
          open ? 'translate-x-0' : '-translate-x-full',
        )}
      >
        <header className="border-b border-white/10 px-5 pb-5 pt-6">
          <p className="industrial-label text-text-secondary">Swarmlens</p>
          <button
            type="button"
            onClick={() => {
              navigate('/')
              onClose()
            }}
            className="mt-3 flex w-full items-center justify-between border-b border-white/10 pb-2 text-left transition-opacity duration-150 hover:opacity-100 focus-visible:outline-none"
            aria-label="Switch cluster context"
          >
            <span className="industrial-data text-[13px] text-text-primary">{clusterShort}</span>
            <ChevronDownIcon className="h-4 w-4 text-text-secondary" />
          </button>
          <p className="mt-2 industrial-label text-text-tertiary">{mode} environment</p>

          <div className="mt-4">
            <p className="industrial-label text-text-secondary">
              STACKS {healthyStacks}/{stacks.length || 0} | NODES {readyNodes}/{nodes.length || 0} | TASKS {runningTasks}/{tasks.length || 0}
            </p>
            <div className="mt-2 h-px w-full bg-white/10" aria-hidden="true">
              <div className="h-px bg-white/45 transition-[width] duration-150" style={{ width: `${healthRatio}%` }} />
            </div>
          </div>
        </header>

        <nav className="flex-1 overflow-y-auto px-3 py-4" aria-label="Primary navigation">
          {groupedItems.map((group) => (
            <section key={group.label} className="mt-2 first:mt-0" aria-label={group.label}>
              {group.items.length >= 4 ? (
                <p className="industrial-label px-2 pb-2 text-text-tertiary">{group.label}</p>
              ) : null}
              <ul className="space-y-1">
                {group.items.map((item) => {
                  const Icon = item.icon
                  return (
                    <li key={item.to}>
                      <NavLink
                        to={item.to}
                        end={item.to === '/'}
                        onClick={onClose}
                        className={({ isActive }) => cn('sl-nav-item block px-2 py-2.5', isActive && 'is-active')}
                      >
                        <div className="flex items-center gap-2.5">
                          <Icon className="h-4 w-4 shrink-0 text-text-secondary transition-opacity duration-120" />
                          <span className="truncate text-[14px] tracking-[0.03em] text-text-secondary transition-opacity duration-120">
                            {item.label}
                          </span>
                        </div>
                        <p className="sl-nav-description mt-1 pl-[26px] text-[11px] leading-[1.4] text-text-secondary">
                          {item.description}
                        </p>
                      </NavLink>
                    </li>
                  )
                })}
              </ul>
            </section>
          ))}
        </nav>

        <footer className="border-t border-white/10 px-5 py-5">
          {disconnected ? (
            <button
              type="button"
              onClick={() => {
                void fetchAll()
              }}
              className="industrial-action industrial-action-accent w-full justify-start text-left"
            >
              Reconnect Cluster
            </button>
          ) : (
            <div>
              <p className="industrial-label text-text-secondary">Operator Tip</p>
              <p className="mt-2 text-[12px] leading-relaxed text-text-secondary">{operatorTip}</p>
              <p className="mt-2 industrial-data text-[11px] text-text-tertiary">Shortcut: G then D</p>
            </div>
          )}
        </footer>
      </aside>
    </>
  )
}
