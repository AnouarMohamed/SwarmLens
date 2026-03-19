import { useMemo } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { relativeTime } from '../../lib/utils'
import {
  CheckCircleIcon,
  DisconnectIcon,
  MenuIcon,
  MoreIcon,
  RefreshIcon,
  WarningIcon,
} from '../ui/icons'

interface TopbarProps {
  onOpenSidebar: () => void
}

const TITLES: Record<string, string> = {
  '/': 'Overview',
  '/stacks': 'Stacks',
  '/services': 'Services',
  '/tasks': 'Tasks',
  '/nodes': 'Nodes',
  '/networks': 'Networks',
  '/volumes': 'Volumes',
  '/secrets': 'Secrets / Configs',
  '/diagnostics': 'Diagnostics',
  '/incidents': 'Incidents',
  '/audit': 'Audit Trail',
  '/assistant': 'Assistant',
}

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function computeHealth({
  disconnected,
  criticalCount,
  warningCount,
  unavailableNodes,
  unhealthyServices,
}: {
  disconnected: boolean
  criticalCount: number
  warningCount: number
  unavailableNodes: number
  unhealthyServices: number
}) {
  if (disconnected) return 'Unknown'
  if (criticalCount > 0 || unavailableNodes > 0 || unhealthyServices > 0) return 'Degraded'
  if (warningCount > 0) return 'Watch'
  return 'Healthy'
}

export function Topbar({ onOpenSidebar }: TopbarProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const title = TITLES[location.pathname] ?? 'SwarmLens'

  const { swarm, nodes, services, lastRefresh, loading, fetchAll, error, connectionState } =
    useClusterStore()
  const { run, running, findings } = useDiagnosticsStore()

  const disconnected =
    connectionState === 'disconnected' || Boolean(error) || swarm?.freshness === 'disconnected'
  const mode = swarm?.mode?.toUpperCase() ?? 'DEMO'
  const clusterName = swarm?.clusterID ? `cluster/${swarm.clusterID.slice(0, 20)}` : 'cluster/unset'

  const criticalCount = findings.filter((finding) => finding.severity === 'critical').length
  const warningCount = findings.filter(
    (finding) => finding.severity === 'high' || finding.severity === 'medium',
  ).length
  const unavailableNodes = nodes.filter((node) => node.state !== 'ready').length
  const unhealthyServices = services.filter(
    (service) => service.runningTasks < service.desiredReplicas || service.failedTasks > 0,
  ).length

  const health = computeHealth({
    disconnected,
    criticalCount,
    warningCount,
    unavailableNodes,
    unhealthyServices,
  })

  const freshness = useMemo(() => {
    if (swarm?.lastSyncAt) return `Last sync ${relativeTime(swarm.lastSyncAt)}`
    if (!lastRefresh) return 'No successful sync yet'
    return `Last sync ${relativeTime(new Date(lastRefresh).toISOString())}`
  }, [swarm?.lastSyncAt, lastRefresh])

  const connectionLabel = disconnected ? 'Disconnected' : connectionState === 'connecting' ? 'Connecting' : 'Connected'
  const disableDiagnostics = running || disconnected
  const operationalSummary = disconnected
    ? 'Live telemetry paused'
    : health === 'Degraded'
      ? `${criticalCount} critical findings · ${unavailableNodes} unavailable nodes`
      : health === 'Watch'
        ? `${warningCount} warnings in diagnostics`
        : 'No active critical findings'

  return (
    <header className="sticky top-0 z-20 border-b border-white/10 bg-app/95 backdrop-blur-sm">
      <div className="mx-auto flex w-full max-w-[1600px] items-start justify-between gap-6 px-4 py-4 sm:px-6 lg:px-8">
        <div className="flex min-w-0 items-start gap-3">
          <button
            type="button"
            onClick={onOpenSidebar}
            aria-label="Open navigation menu"
            className="mt-1 inline-flex h-9 w-9 shrink-0 items-center justify-center text-text-secondary transition-opacity duration-120 hover:opacity-100 focus-visible:opacity-100 lg:hidden"
          >
            <MenuIcon className="h-4 w-4" />
          </button>

          <div className="min-w-0">
            <h1 className="font-heading text-[2rem] uppercase leading-none tracking-[0.05em] text-text-primary">
              {title}
            </h1>
            <p className="mt-2 truncate font-mono text-sm text-text-secondary">
              {clusterName} · {mode} · {operationalSummary}
            </p>
            <div className="mt-3 flex flex-wrap items-center gap-x-5 gap-y-1">
              <span
                className={cn(
                  'industrial-label inline-flex items-center gap-1.5',
                  disconnected && 'text-state-danger',
                )}
              >
                {disconnected ? (
                  <DisconnectIcon className="h-3 w-3" />
                ) : (
                  <CheckCircleIcon className="h-3 w-3" />
                )}
                {connectionLabel}
              </span>
              <span className="industrial-label inline-flex items-center gap-1.5 text-text-secondary">
                Environment {mode}
              </span>
              <span
                className={cn(
                  'industrial-label inline-flex items-center gap-1.5',
                  health === 'Degraded' && 'text-state-danger',
                )}
              >
                {health === 'Degraded' ? <WarningIcon className="h-3 w-3" /> : null}
                Health {health}
              </span>
              <span className="industrial-label inline-flex items-center gap-1.5 text-text-secondary">
                {freshness}
              </span>
            </div>
          </div>
        </div>

        <div className="flex shrink-0 items-center gap-4 pt-1">
          {disconnected ? (
            <>
              <button
                type="button"
                onClick={() => {
                  void fetchAll()
                }}
                className="industrial-action industrial-action-accent"
              >
                Reconnect Cluster
              </button>
              <button type="button" disabled className="industrial-action cursor-not-allowed opacity-35">
                Run Diagnostics
              </button>
              <button
                type="button"
                onClick={() => navigate('/audit')}
                className="industrial-action opacity-35 hover:opacity-55 focus-visible:opacity-55"
              >
                Open Audit Trail
              </button>
            </>
          ) : (
            <>
              <button
                type="button"
                onClick={() => {
                  void run()
                }}
                disabled={disableDiagnostics}
                className={cn(
                  'industrial-action industrial-action-accent',
                  disableDiagnostics && 'cursor-not-allowed opacity-35',
                )}
              >
                {running ? 'Running Diagnostics' : 'Run Diagnostics'}
              </button>
              <button
                type="button"
                onClick={() => {
                  void fetchAll()
                }}
                disabled={loading}
                className={cn('industrial-action', loading && 'cursor-not-allowed opacity-35')}
              >
                <RefreshIcon className="h-3.5 w-3.5" />
                {loading ? 'Refreshing' : 'Refresh'}
              </button>
              <button
                type="button"
                onClick={() => navigate('/incidents')}
                className="industrial-action"
              >
                <MoreIcon className="h-3.5 w-3.5" />
                Open Incidents
              </button>
            </>
          )}
        </div>
      </div>
    </header>
  )
}
