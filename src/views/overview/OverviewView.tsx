
import { useEffect, useMemo, useRef } from 'react'
import type { ComponentType, SVGProps } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { GrafanaEmbed } from '../../components/charts/GrafanaEmbed'
import {
  GrafanaAreaSeries,
  GrafanaBarSeries,
  GrafanaTimeSeries,
} from '../../components/charts/GrafanaCharts'
import { grafanaConfig } from '../../lib/grafana'
import { buildMockDiagnosticsFindings, buildMockSwarmEvents } from '../../lib/mockData'
import { buildOverviewTelemetry } from '../../lib/telemetry'
import { relativeTime } from '../../lib/utils'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useIncidentStore } from '../../store/incidentStore'
import type { Severity, SwarmEvent } from '../../types'
import {
  ActivityIcon,
  AlertIcon,
  ArrowRightIcon,
  CheckCircleIcon,
  DiagnosticsIcon,
  DisconnectIcon,
  RefreshIcon,
  ServerIcon,
  ShieldIcon,
  WarningIcon,
} from '../../components/ui/icons'

type Tone = 'primary' | 'attention' | 'muted'
type Scenario = 'healthy' | 'degraded' | 'disconnected'
type IconComponent = ComponentType<SVGProps<SVGSVGElement>>

interface Metric {
  id: string
  label: string
  value: string
  supporting: string
  note: string
  to: string
  tone: Tone
  icon: IconComponent
  unknown?: boolean
}

interface Check {
  id: string
  label: string
  detail: string
  status: string
  tone: Tone
}

interface FindingRow {
  id: string
  title: string
  object: string
  timestamp: string
  tone: Tone
  to: string
}

interface EventRow {
  id: string
  title: string
  source: string
  timestamp: string
  metadata?: string
  tone: Tone
}

interface AttentionRow {
  id: string
  name: string
  replicas: string
  state: string
  issue: string
}

interface PostureRow {
  id: string
  label: string
  value: string
  tone: Tone
}

interface OverviewModel {
  disconnected: boolean
  stale: boolean
  clusterName: string
  environment: string
  endpoint: string
  healthLabel: 'Healthy' | 'Degraded' | 'Unknown'
  summary: string
  freshness: string
  metrics: Metric[]
  checks: Check[]
  findings: FindingRow[]
  events: EventRow[]
  attention: AttentionRow[]
  diagnostics: {
    status: string
    lastRun: number | null
    durationMs: number | null
    findingsCount: number
  }
  posture: PostureRow[]
  guidance: {
    title: string
    description: string
    steps: string[]
  }
  telemetrySeed: {
    runningTasks: number
    failedTasks: number
    managersOnline: number
    workersOnline: number
    criticalFindings: number
    warningFindings: number
    degraded: boolean
  }
}

const PENDING_TASK_STATES = new Set(['new', 'pending', 'assigned', 'accepted', 'preparing', 'starting'])

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function toneClass(tone: Tone) {
  if (tone === 'attention') return 'text-state-danger'
  if (tone === 'muted') return 'text-text-secondary'
  return 'text-text-primary'
}

function toneIcon(tone: Tone) {
  if (tone === 'attention') return <WarningIcon className="h-4 w-4 text-state-danger" />
  if (tone === 'muted') return <DisconnectIcon className="h-4 w-4 text-text-secondary" />
  return <CheckCircleIcon className="h-4 w-4 text-text-secondary" />
}

function toneFromSeverity(severity: Severity): Tone {
  return severity === 'critical' || severity === 'high' || severity === 'medium' ? 'attention' : 'primary'
}

function toneFromEvent(evt: SwarmEvent): Tone {
  const payload = `${evt.action} ${evt.message}`.toLowerCase()
  return /(failed|error|reject|down|unreachable|timeout|panic)/.test(payload) ? 'attention' : 'primary'
}

function durationLabel(ms: number | null) {
  if (!ms) return 'n/a'
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`
}

function delta(v: number) {
  return v > 0 ? `+${v}` : `${v}`
}

function clock(ts: string) {
  return new Date(ts).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function scenarioModel(kind: Scenario, endpoint: string): OverviewModel {
  const baseChecks: Check[] = [
    {
      id: 'c1',
      label: 'Control plane reachable',
      detail: kind === 'disconnected' ? 'Manager endpoint unreachable.' : 'Manager API responding.',
      status: kind === 'disconnected' ? 'Unknown' : kind === 'degraded' ? 'Watch' : 'Nominal',
      tone: kind === 'disconnected' ? 'muted' : kind === 'degraded' ? 'attention' : 'primary',
    },
    {
      id: 'c2',
      label: 'Node quorum',
      detail:
        kind === 'degraded'
          ? '2/3 managers reachable.'
          : kind === 'disconnected'
            ? 'No live quorum data.'
            : '3/3 managers reachable.',
      status: kind === 'degraded' ? 'Degraded' : kind === 'disconnected' ? 'Unknown' : 'Nominal',
      tone: kind === 'degraded' ? 'attention' : kind === 'disconnected' ? 'muted' : 'primary',
    },
    {
      id: 'c3',
      label: 'Task scheduling',
      detail:
        kind === 'degraded'
          ? 'Placement failures detected.'
          : kind === 'disconnected'
            ? 'Scheduler stream paused.'
            : 'No pending failures.',
      status: kind === 'degraded' ? 'Degraded' : kind === 'disconnected' ? 'Unknown' : 'Nominal',
      tone: kind === 'degraded' ? 'attention' : kind === 'disconnected' ? 'muted' : 'primary',
    },
    {
      id: 'c4',
      label: 'Replica health',
      detail:
        kind === 'degraded'
          ? 'Replica drift in 2 services.'
          : kind === 'disconnected'
            ? 'Replica state unavailable.'
            : 'All replicas healthy.',
      status: kind === 'degraded' ? 'Degraded' : kind === 'disconnected' ? 'Unknown' : 'Nominal',
      tone: kind === 'degraded' ? 'attention' : kind === 'disconnected' ? 'muted' : 'primary',
    },
    {
      id: 'c5',
      label: 'Network/config health',
      detail:
        kind === 'disconnected'
          ? 'No live network telemetry.'
          : 'Overlay and config checks nominal.',
      status: kind === 'disconnected' ? 'Unknown' : 'Nominal',
      tone: kind === 'disconnected' ? 'muted' : 'primary',
    },
  ]

  if (kind === 'disconnected') {
    return {
      disconnected: true,
      stale: true,
      clusterName: 'cluster/swrm-prod-eu-01',
      environment: 'PROD',
      endpoint,
      healthLabel: 'Unknown',
      summary: 'SwarmLens is not connected. Live telemetry and controls are paused.',
      freshness: 'No successful sync yet',
      metrics: [
        {
          id: 'health',
          label: 'Cluster Health',
          value: 'Unknown',
          supporting: 'Last known: 38/42 replicas healthy',
          note: 'Showing last known telemetry',
          to: '/diagnostics',
          tone: 'attention',
          icon: ShieldIcon,
        },
        {
          id: 'nodes',
          label: 'Nodes',
          value: '—',
          supporting: 'UNKNOWN',
          note: 'Reconnect to read node state',
          to: '/nodes',
          tone: 'attention',
          icon: ServerIcon,
          unknown: true,
        },
        {
          id: 'tasks',
          label: 'Running Tasks',
          value: '—',
          supporting: 'UNKNOWN',
          note: 'Scheduler telemetry paused',
          to: '/tasks',
          tone: 'attention',
          icon: ActivityIcon,
          unknown: true,
        },
        {
          id: 'findings',
          label: 'Active Findings',
          value: '5',
          supporting: '2 critical / 3 warning (last known)',
          note: 'Reconnect to confirm current state',
          to: '/incidents',
          tone: 'attention',
          icon: AlertIcon,
        },
      ],
      checks: baseChecks,
      findings: [],
      events: [],
      attention: [],
      diagnostics: {
        status: 'Offline',
        lastRun: Date.now() - 14 * 60_000,
        durationMs: 1800,
        findingsCount: 5,
      },
      posture: [
        { id: 'p1', label: 'Managers online', value: 'unknown', tone: 'muted' },
        { id: 'p2', label: 'Workers online', value: 'unknown', tone: 'muted' },
        { id: 'p3', label: 'Total replicas', value: '42 (last known)', tone: 'muted' },
        { id: 'p4', label: 'Pending tasks', value: 'unknown', tone: 'muted' },
        { id: 'p5', label: 'Unhealthy services', value: 'unknown', tone: 'muted' },
        { id: 'p6', label: 'Swarm quorum', value: 'unknown', tone: 'muted' },
      ],
      guidance: {
        title: 'Reconnect cluster to resume operations',
        description: 'Live reads and write controls remain paused while disconnected.',
        steps: ['Validate manager reachability.', 'Check endpoint credentials.', 'Reconnect and refresh telemetry.'],
      },
      telemetrySeed: {
        runningTasks: 124,
        failedTasks: 4,
        managersOnline: 2,
        workersOnline: 4,
        criticalFindings: 2,
        warningFindings: 3,
        degraded: true,
      },
    }
  }

  if (kind === 'degraded') {
    return {
      disconnected: false,
      stale: false,
      clusterName: 'cluster/swrm-prod-eu-01',
      environment: 'PROD',
      endpoint,
      healthLabel: 'Degraded',
      summary: 'Cluster is partially degraded. 2 services have replica drift and 1 worker is unreachable.',
      freshness: 'Last sync 36s ago',
      metrics: [
        {
          id: 'health',
          label: 'Cluster Health',
          value: 'Degraded',
          supporting: '38/42 replicas healthy',
          note: 'Checked 2m ago',
          to: '/diagnostics',
          tone: 'attention',
          icon: ShieldIcon,
        },
        {
          id: 'nodes',
          label: 'Nodes',
          value: '8',
          supporting: '3 managers / 5 workers',
          note: '1 unavailable',
          to: '/nodes',
          tone: 'attention',
          icon: ServerIcon,
        },
        {
          id: 'tasks',
          label: 'Running Tasks',
          value: '124',
          supporting: '120 healthy / 4 failed',
          note: '-3 since last refresh',
          to: '/tasks',
          tone: 'attention',
          icon: ActivityIcon,
        },
        {
          id: 'findings',
          label: 'Active Findings',
          value: '5',
          supporting: '2 critical / 3 warning',
          note: 'Review incidents now',
          to: '/incidents',
          tone: 'attention',
          icon: AlertIcon,
        },
      ],
      checks: baseChecks,
      findings: [
        {
          id: 'f1',
          title: 'Replica shortfall in payments-worker',
          object: 'service/payments-worker',
          timestamp: new Date(Date.now() - 6 * 60_000).toISOString(),
          tone: 'attention',
          to: '/incidents',
        },
      ],
      events: [
        {
          id: 'e1',
          title: 'Task rejected due to placement constraint',
          source: 'payments-worker',
          timestamp: new Date(Date.now() - 4 * 60_000).toISOString(),
          metadata: 'node.labels.zone=eu-central',
          tone: 'attention',
        },
      ],
      attention: [
        {
          id: 'a1',
          name: 'payments-worker',
          replicas: '0 / 2',
          state: 'Replica drift',
          issue: '3 placement failures in 10m',
        },
      ],
      diagnostics: {
        status: 'Degraded',
        lastRun: Date.now() - 2 * 60_000,
        durationMs: 1700,
        findingsCount: 5,
      },
      posture: [
        { id: 'p1', label: 'Managers online', value: '2 / 3', tone: 'attention' },
        { id: 'p2', label: 'Workers online', value: '4 / 5', tone: 'attention' },
        { id: 'p3', label: 'Total replicas', value: '38 / 42', tone: 'attention' },
        { id: 'p4', label: 'Pending tasks', value: '6', tone: 'attention' },
        { id: 'p5', label: 'Unhealthy services', value: '2', tone: 'attention' },
        { id: 'p6', label: 'Swarm quorum', value: 'degraded', tone: 'attention' },
      ],
      guidance: {
        title: 'Prioritize quorum and replica recovery',
        description: 'Recover control-plane stability first, then clear placement failures.',
        steps: ['Restore manager reachability.', 'Resolve placement constraints.', 'Re-run diagnostics.'],
      },
      telemetrySeed: {
        runningTasks: 124,
        failedTasks: 4,
        managersOnline: 2,
        workersOnline: 4,
        criticalFindings: 2,
        warningFindings: 3,
        degraded: true,
      },
    }
  }

  return {
    disconnected: false,
    stale: false,
    clusterName: 'cluster/swrm-dev-lab-01',
    environment: 'DEMO',
    endpoint,
    healthLabel: 'Healthy',
    summary: 'Cluster is healthy. No active critical findings. Last diagnostics run 8 minutes ago.',
    freshness: 'Last sync 28s ago',
    metrics: [
      {
        id: 'health',
        label: 'Cluster Health',
        value: 'Healthy',
        supporting: '24/24 replicas healthy',
        note: 'Checked 2m ago',
        to: '/diagnostics',
        tone: 'primary',
        icon: ShieldIcon,
      },
      {
        id: 'nodes',
        label: 'Nodes',
        value: '5',
        supporting: '3 managers / 2 workers',
        note: '0 unavailable',
        to: '/nodes',
        tone: 'primary',
        icon: ServerIcon,
      },
      {
        id: 'tasks',
        label: 'Running Tasks',
        value: '72',
        supporting: '72 healthy / 0 failed',
        note: '+1 since last refresh',
        to: '/tasks',
        tone: 'primary',
        icon: ActivityIcon,
      },
      {
        id: 'findings',
        label: 'Active Findings',
        value: '0',
        supporting: 'No critical or warning findings',
        note: 'Diagnostics completed 8m ago',
        to: '/incidents',
        tone: 'primary',
        icon: AlertIcon,
      },
    ],
    checks: baseChecks,
    findings: [],
    events: [
      {
        id: 'e1',
        title: 'Diagnostics completed with zero findings',
        source: 'diagnostics',
        timestamp: new Date(Date.now() - 8 * 60_000).toISOString(),
        tone: 'primary',
      },
    ],
    attention: [],
    diagnostics: {
      status: 'Clear',
      lastRun: Date.now() - 8 * 60_000,
      durationMs: 900,
      findingsCount: 0,
    },
    posture: [
      { id: 'p1', label: 'Managers online', value: '3 / 3', tone: 'primary' },
      { id: 'p2', label: 'Workers online', value: '2 / 2', tone: 'primary' },
      { id: 'p3', label: 'Total replicas', value: '24 / 24', tone: 'primary' },
      { id: 'p4', label: 'Pending tasks', value: '0', tone: 'primary' },
      { id: 'p5', label: 'Unhealthy services', value: '0', tone: 'primary' },
      { id: 'p6', label: 'Swarm quorum', value: 'healthy', tone: 'primary' },
    ],
    guidance: {
      title: 'No active workload issues',
      description: 'Telemetry is stable and no mitigation workflow is required right now.',
      steps: ['Run diagnostics on cadence.', 'Review audit trail for writes.', 'Inspect services before rollouts.'],
    },
    telemetrySeed: {
      runningTasks: 72,
      failedTasks: 0,
      managersOnline: 3,
      workersOnline: 2,
      criticalFindings: 0,
      warningFindings: 0,
      degraded: false,
    },
  }
}

function OverviewSkeleton() {
  return (
    <div className="animate-pulse">
      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1600px] px-4 sm:px-6 lg:px-10">
          <div className="h-3 w-24 bg-white/10" />
          <div className="mt-5 h-20 w-full max-w-3xl bg-white/10" />
        </div>
      </section>
      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1600px] px-4 sm:px-6 lg:px-10">
          <div className="grid grid-cols-2 gap-8 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, idx) => (
              <div key={idx} className="h-20 bg-white/10" />
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}

interface SectionHeadingProps {
  id: string
  eyebrow: string
  title: string
  actionLabel?: string
  onAction?: () => void
  actionDisabled?: boolean
}

function SectionHeading({ id, eyebrow, title, actionLabel, onAction, actionDisabled }: SectionHeadingProps) {
  return (
    <div className="flex items-end justify-between gap-6">
      <div>
        <p className="industrial-label">{eyebrow}</p>
        <h3 id={id} className="mt-2 font-heading text-[1.85rem] uppercase leading-none tracking-[0.05em]">
          {title}
        </h3>
      </div>
      {actionLabel && onAction ? (
        <button
          type="button"
          onClick={onAction}
          disabled={actionDisabled}
          className={cn('industrial-action', actionDisabled && 'cursor-not-allowed opacity-35')}
        >
          {actionLabel}
          <ArrowRightIcon className="h-3.5 w-3.5" />
        </button>
      ) : null}
    </div>
  )
}

export function OverviewView() {
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const {
    swarm,
    nodes,
    services,
    tasks,
    events,
    networks,
    configs,
    loading,
    error,
    connectionState,
    lastRefresh,
    fetchAll,
  } = useClusterStore()
  const { findings, fetch: fetchDiagnostics, run, running, lastRun, lastDurationMs, error: findingsError } =
    useDiagnosticsStore()
  const { incidents, fetch: fetchIncidents, loading: incidentsLoading } = useIncidentStore()

  const prevRunningTasksRef = useRef<number | null>(null)
  const prevRefreshRef = useRef(0)

  useEffect(() => {
    void fetchDiagnostics()
    void fetchIncidents()
  }, [fetchDiagnostics, fetchIncidents])

  const liveRunningTasks = tasks.filter((task) => task.currentState === 'running').length
  const liveFailedTasks = tasks.filter(
    (task) => task.currentState === 'failed' || task.currentState === 'rejected',
  ).length

  useEffect(() => {
    if (lastRefresh && lastRefresh !== prevRefreshRef.current) {
      prevRefreshRef.current = lastRefresh
      prevRunningTasksRef.current = liveRunningTasks
    }
  }, [lastRefresh, liveRunningTasks])

  const taskDelta =
    prevRunningTasksRef.current === null ? 0 : liveRunningTasks - prevRunningTasksRef.current
  const disconnected = connectionState === 'disconnected' || Boolean(error)

  const hasData =
    Boolean(swarm) ||
    nodes.length > 0 ||
    services.length > 0 ||
    tasks.length > 0 ||
    events.length > 0 ||
    findings.length > 0 ||
    incidents.length > 0

  const endpoint = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'
  const forcedScenario = searchParams.get('scenario')
  const scenarioKind: Scenario | null =
    forcedScenario === 'healthy' || forcedScenario === 'degraded' || forcedScenario === 'disconnected'
      ? forcedScenario
      : !hasData && !loading
        ? disconnected
          ? 'disconnected'
          : 'healthy'
        : null
  const isDemoMode = (swarm?.mode ?? '').toLowerCase() === 'demo'
  const isDemoExperience = isDemoMode || Boolean(scenarioKind)
  const demoFindings = useMemo(() => buildMockDiagnosticsFindings(), [])
  const demoEvents = useMemo(() => buildMockSwarmEvents(), [])

  const model = useMemo<OverviewModel>(() => {
    if (scenarioKind) return scenarioModel(scenarioKind, endpoint)

    const managers = nodes.filter((node) => node.role === 'manager')
    const workers = nodes.filter((node) => node.role === 'worker')
    const managersOnline = managers.filter(
      (node) => node.state === 'ready' && node.managerStatus?.reachability !== 'unreachable',
    ).length
    const workersOnline = workers.filter((node) => node.state === 'ready').length
    const unavailableNodes = nodes.filter((node) => node.state !== 'ready').length
    const desiredReplicas = services.reduce((sum, service) => sum + service.desiredReplicas, 0)
    const runningReplicas = services.reduce((sum, service) => sum + service.runningTasks, 0)
    const unhealthyServices = services.filter(
      (service) => service.runningTasks < service.desiredReplicas || service.failedTasks > 0,
    )
    const pendingTasks = tasks.filter((task) => PENDING_TASK_STATES.has(task.currentState)).length
    const effectiveFindings = findings.length > 0 ? findings : isDemoMode ? demoFindings : []
    const effectiveEvents = events.length > 0 ? events : isDemoMode ? demoEvents : []
    const criticalFindings = effectiveFindings.filter(
      (finding) => finding.severity === 'critical' || finding.severity === 'high',
    ).length
    const warningFindings = effectiveFindings.filter((finding) => finding.severity === 'medium').length
    const findingsCount = effectiveFindings.length
    const stale = disconnected || !lastRefresh || Date.now() - lastRefresh > 120_000
    const unknown = disconnected && !lastRefresh && !hasData
    const degraded =
      disconnected || criticalFindings > 0 || unavailableNodes > 0 || unhealthyServices.length > 0
    const healthLabel: OverviewModel['healthLabel'] = disconnected
      ? 'Unknown'
      : degraded
        ? 'Degraded'
        : 'Healthy'

    return {
      disconnected,
      stale,
      clusterName: swarm?.clusterID ? `cluster/${swarm.clusterID.slice(0, 20)}` : 'cluster/unset',
      environment: swarm?.mode?.toUpperCase() ?? 'DEMO',
      endpoint,
      healthLabel,
      summary: disconnected
        ? 'SwarmLens is not connected. Live telemetry and controls are paused.'
        : degraded
          ? `Cluster is partially degraded. ${unhealthyServices.length} services require attention and ${unavailableNodes} nodes are unavailable.`
          : `Cluster is healthy. No active critical findings. Last diagnostics run ${lastRun ? relativeTime(new Date(lastRun).toISOString()) : 'not available'}.`,
      freshness: lastRefresh
        ? `Last sync ${relativeTime(new Date(lastRefresh).toISOString())}`
        : 'No successful sync yet',
      metrics: [
        {
          id: 'health',
          label: 'Cluster Health',
          value: unknown && desiredReplicas === 0 ? '—' : healthLabel,
          supporting:
            unknown && desiredReplicas === 0
              ? 'UNKNOWN'
              : `${runningReplicas}/${desiredReplicas || 0} replicas healthy`,
          note: lastRefresh
            ? `Checked ${relativeTime(new Date(lastRefresh).toISOString())}`
            : 'No telemetry snapshot',
          to: '/diagnostics',
          tone: degraded ? 'attention' : disconnected ? 'muted' : 'primary',
          icon: ShieldIcon,
          unknown: unknown && desiredReplicas === 0,
        },
        {
          id: 'nodes',
          label: 'Nodes',
          value: unknown && nodes.length === 0 ? '—' : `${nodes.length}`,
          supporting:
            unknown && nodes.length === 0
              ? 'UNKNOWN'
              : `${managers.length} managers / ${workers.length} workers`,
          note: `${unavailableNodes} unavailable`,
          to: '/nodes',
          tone: unavailableNodes > 0 ? 'attention' : disconnected ? 'muted' : 'primary',
          icon: ServerIcon,
          unknown: unknown && nodes.length === 0,
        },
        {
          id: 'tasks',
          label: 'Running Tasks',
          value: unknown && tasks.length === 0 ? '—' : `${liveRunningTasks}`,
          supporting:
            unknown && tasks.length === 0
              ? 'UNKNOWN'
              : `${Math.max(liveRunningTasks - liveFailedTasks, 0)} healthy / ${liveFailedTasks} failed`,
          note: `${delta(taskDelta)} since last refresh`,
          to: '/tasks',
          tone: liveFailedTasks > 0 ? 'attention' : disconnected ? 'muted' : 'primary',
          icon: ActivityIcon,
          unknown: unknown && tasks.length === 0,
        },
        {
          id: 'findings',
          label: 'Active Findings',
          value: unknown && findingsCount === 0 ? '—' : `${findingsCount}`,
          supporting:
            unknown && findingsCount === 0
              ? 'UNKNOWN'
              : `${criticalFindings} critical / ${warningFindings} warning`,
          note: findingsCount > 0 ? 'Review incidents and diagnostics' : 'No active findings',
          to: '/incidents',
          tone: findingsCount > 0 ? 'attention' : disconnected ? 'muted' : 'primary',
          icon: AlertIcon,
          unknown: unknown && findingsCount === 0,
        },
      ],
      checks: [
        {
          id: 'c1',
          label: 'Control plane reachable',
          detail: disconnected
            ? 'Manager endpoint currently unreachable.'
            : 'Manager API responding in expected range.',
          status: disconnected ? 'Unknown' : 'Nominal',
          tone: disconnected ? 'muted' : 'primary',
        },
        {
          id: 'c2',
          label: 'Node quorum',
          detail: `${managersOnline}/${managers.length || 0} managers reachable.`,
          status:
            disconnected
              ? 'Unknown'
              : managersOnline < Math.max(1, Math.ceil(managers.length / 2))
                ? 'Degraded'
                : 'Nominal',
          tone:
            disconnected
              ? 'muted'
              : managersOnline < Math.max(1, Math.ceil(managers.length / 2))
                ? 'attention'
                : 'primary',
        },
        {
          id: 'c3',
          label: 'Task scheduling',
          detail: `${pendingTasks} pending tasks, ${liveFailedTasks} failed tasks.`,
          status:
            disconnected ? 'Unknown' : liveFailedTasks > 0 ? 'Degraded' : pendingTasks > 0 ? 'Watch' : 'Nominal',
          tone:
            disconnected
              ? 'muted'
              : liveFailedTasks > 0 || pendingTasks > 0
                ? 'attention'
                : 'primary',
        },
        {
          id: 'c4',
          label: 'Replica health',
          detail: `${runningReplicas}/${desiredReplicas || 0} desired replicas running.`,
          status: disconnected ? 'Unknown' : runningReplicas < desiredReplicas ? 'Degraded' : 'Nominal',
          tone: disconnected ? 'muted' : runningReplicas < desiredReplicas ? 'attention' : 'primary',
        },
        {
          id: 'c5',
          label: 'Network/config health',
          detail: disconnected
            ? 'No live network/config telemetry.'
            : `${networks.length} networks / ${configs.length} configs in scope.`,
          status: disconnected ? 'Unknown' : 'Nominal',
          tone: disconnected ? 'muted' : 'primary',
        },
      ],

      findings: effectiveFindings.slice(0, 6).map((finding) => ({
        id: finding.id,
        title: finding.message,
        object: finding.resource,
        timestamp: finding.detectedAt,
        tone: toneFromSeverity(finding.severity),
        to: '/incidents',
      })),
      events: effectiveEvents.slice(0, 12).map((evt, idx) => ({
        id: `${evt.timestamp}-${idx}`,
        title: evt.message || `${evt.type} ${evt.action}`,
        source: evt.actor || evt.type,
        timestamp: evt.timestamp,
        metadata: evt.action ? `${evt.type} ${evt.action}` : undefined,
        tone: toneFromEvent(evt),
      })),
      attention: unhealthyServices.slice(0, 8).map((service) => ({
        id: service.id,
        name: service.name,
        replicas: `${service.runningTasks} / ${service.desiredReplicas}`,
        state: service.runningTasks < service.desiredReplicas ? 'Replica drift' : 'Task failures',
        issue:
          service.failedTasks > 0
            ? `${service.failedTasks} failed tasks`
            : `${service.desiredReplicas - service.runningTasks} replicas missing`,
      })),
      diagnostics: {
        status: disconnected ? 'Offline' : degraded ? 'Degraded' : 'Clear',
        lastRun,
        durationMs: lastDurationMs,
        findingsCount,
      },
      posture: [
        {
          id: 'p1',
          label: 'Managers online',
          value: `${managersOnline} / ${managers.length || 0}`,
          tone: disconnected ? 'muted' : managersOnline < managers.length ? 'attention' : 'primary',
        },
        {
          id: 'p2',
          label: 'Workers online',
          value: `${workersOnline} / ${workers.length || 0}`,
          tone: disconnected ? 'muted' : workersOnline < workers.length ? 'attention' : 'primary',
        },
        {
          id: 'p3',
          label: 'Total replicas',
          value: `${runningReplicas} / ${desiredReplicas || 0}`,
          tone: disconnected ? 'muted' : runningReplicas < desiredReplicas ? 'attention' : 'primary',
        },
        {
          id: 'p4',
          label: 'Pending tasks',
          value: `${pendingTasks}`,
          tone: disconnected ? 'muted' : pendingTasks > 0 ? 'attention' : 'primary',
        },
        {
          id: 'p5',
          label: 'Unhealthy services',
          value: `${unhealthyServices.length}`,
          tone: disconnected ? 'muted' : unhealthyServices.length > 0 ? 'attention' : 'primary',
        },
        {
          id: 'p6',
          label: 'Swarm quorum',
          value: disconnected ? 'unknown' : swarm?.raftState ?? 'unknown',
          tone: disconnected ? 'muted' : swarm?.quorumHealthy ? 'primary' : 'attention',
        },
      ],
      guidance: disconnected
        ? {
            title: 'Reconnect cluster to resume operations',
            description: 'Telemetry is paused while disconnected. Last known data is shown as stale.',
            steps: ['Check manager reachability.', 'Validate credentials and TLS.', 'Reconnect and refresh telemetry.'],
          }
        : services.length === 0
          ? {
              title: 'This cluster has no active workloads yet',
              description: 'Control plane is healthy, but no services are currently scheduled.',
              steps: ['Create your first stack.', 'Connect worker nodes.', 'Run diagnostics after deployment.'],
            }
          : {
              title: degraded ? 'Prioritize mitigation sequence' : 'No immediate action required',
              description: degraded
                ? 'Recover quorum first, then resolve placement and rollout pressure.'
                : 'Cluster posture is stable and no urgent intervention is needed.',
              steps: degraded
                ? ['Restore manager/worker reachability.', 'Clear placement failures.', 'Re-run diagnostics and validate closure.']
                : ['Run diagnostics on schedule.', 'Review audit trail after writes.', 'Inspect services before rollouts.'],
            },
      telemetrySeed: {
        runningTasks: liveRunningTasks,
        failedTasks: liveFailedTasks,
        managersOnline,
        workersOnline,
        criticalFindings,
        warningFindings,
        degraded,
      },
    }
  }, [
    scenarioKind,
    endpoint,
    isDemoMode,
    demoFindings,
    demoEvents,
    disconnected,
    nodes,
    services,
    tasks,
    events,
    findings,
    networks.length,
    configs.length,
    swarm,
    hasData,
    lastRefresh,
    lastRun,
    lastDurationMs,
    liveRunningTasks,
    liveFailedTasks,
    taskDelta,
  ])

  const telemetry = useMemo(
    () =>
      buildOverviewTelemetry({
        runningTasks: model.telemetrySeed.runningTasks,
        failedTasks: model.telemetrySeed.failedTasks,
        managersOnline: model.telemetrySeed.managersOnline,
        workersOnline: model.telemetrySeed.workersOnline,
        criticalFindings: model.telemetrySeed.criticalFindings,
        warningFindings: model.telemetrySeed.warningFindings,
        disconnected: model.disconnected,
        degraded: model.telemetrySeed.degraded,
      }),
    [model],
  )

  const findingDistribution = useMemo(() => {
    const rows = [
      { bucket: 'Critical', findings: model.telemetrySeed.criticalFindings },
      { bucket: 'Warning', findings: model.telemetrySeed.warningFindings },
    ]
    return rows.every((row) => row.findings === 0) ? [{ bucket: 'None', findings: 0 }] : rows
  }, [model.telemetrySeed.criticalFindings, model.telemetrySeed.warningFindings])

  const grafana = grafanaConfig()
  const canRunDiagnostics = !model.disconnected && !running
  const activeScenario: Scenario =
    scenarioKind ?? (model.disconnected ? 'disconnected' : model.healthLabel === 'Degraded' ? 'degraded' : 'healthy')

  function setScenario(next: Scenario) {
    const params = new URLSearchParams(searchParams)
    params.set('scenario', next)
    navigate({ pathname: '/', search: `?${params.toString()}` })
  }

  function clearScenario() {
    const params = new URLSearchParams(searchParams)
    params.delete('scenario')
    const query = params.toString()
    navigate({ pathname: '/', search: query ? `?${query}` : '' })
  }

  if (loading && !hasData && !scenarioKind) return <OverviewSkeleton />

  return (
    <div>
      <section
        className={cn('industrial-section pt-12', model.disconnected && 'flex min-h-[40vh] items-center')}
      >
        <div className="mx-auto w-full max-w-[1600px] px-4 sm:px-6 lg:px-10">
          <p className={cn('industrial-label', model.disconnected && 'text-state-danger')}>
            Cluster State
          </p>
          <h2
            className={cn(
              'industrial-data mt-5 leading-[0.92] tracking-[-0.03em]',
              model.disconnected
                ? 'industrial-pulse text-[clamp(2.8rem,8vw,7rem)] text-state-danger'
                : 'text-[clamp(2rem,4vw,5rem)] text-text-primary',
            )}
          >
            {model.clusterName}
          </h2>
          <p
            className={cn(
              'mt-4 max-w-3xl text-sm leading-relaxed',
              model.disconnected ? 'text-text-primary' : 'text-text-secondary',
            )}
          >
            {model.summary}
          </p>
          <p className={cn('mt-2 industrial-label', model.healthLabel === 'Degraded' && 'text-state-danger')}>
            {model.environment} · {model.healthLabel} · {model.freshness}
          </p>
          <p className="mt-1 industrial-data text-xs text-text-tertiary">{model.endpoint}</p>

          {model.disconnected ? (
            <div className="mt-8 space-y-4">
              <button
                type="button"
                onClick={() => {
                  void fetchAll()
                }}
                className="industrial-action industrial-action-accent"
              >
                Reconnect Cluster
              </button>
              <div className="flex flex-wrap items-center gap-6">
                <button type="button" disabled className="industrial-action cursor-not-allowed opacity-35">
                  Run Diagnostics
                </button>
                <button
                  type="button"
                  onClick={() => {
                    void fetchAll()
                  }}
                  className="industrial-action opacity-35 hover:opacity-55 focus-visible:opacity-55"
                >
                  Refresh Telemetry
                </button>
                <button
                  type="button"
                  onClick={() => navigate('/audit')}
                  className="industrial-action opacity-35 hover:opacity-55 focus-visible:opacity-55"
                >
                  Open Audit Trail
                </button>
              </div>
            </div>
          ) : (
            <div className="mt-8 flex flex-wrap items-center gap-8">
              <button
                type="button"
                onClick={() => {
                  void run()
                }}
                disabled={!canRunDiagnostics}
                className={cn(
                  'industrial-action industrial-action-accent',
                  !canRunDiagnostics && 'cursor-not-allowed opacity-35',
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
                {loading ? 'Refreshing' : 'Refresh Telemetry'}
              </button>
              <button type="button" onClick={() => navigate('/audit')} className="industrial-action">
                Open Audit Trail
              </button>
            </div>
          )}
        </div>
      </section>

      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1600px] px-4 sm:px-6 lg:px-10">
          <p className={cn('industrial-label', model.stale ? 'text-state-danger' : 'text-text-secondary')}>
            {model.stale ? 'STALE' : 'LIVE'}
          </p>
          <div className="mt-8 grid grid-cols-1 gap-x-8 gap-y-8 md:grid-cols-2 xl:grid-cols-4">
            {model.metrics.map((metric) => {
              const Icon = metric.icon
              return (
                <button
                  key={metric.id}
                  type="button"
                  onClick={() => navigate(metric.to)}
                  className={cn(
                    'industrial-row w-full text-left focus-visible:outline-none',
                    model.stale && 'opacity-40',
                  )}
                >
                  <div className="flex items-start justify-between gap-4">
                    <div className="min-w-0">
                      <p
                        className={cn(
                          'industrial-metric truncate',
                          metric.unknown ? 'text-state-danger' : toneClass(metric.tone),
                        )}
                      >
                        {metric.value}
                      </p>
                      <p
                        className={cn(
                          'mt-3 industrial-label',
                          metric.unknown ? 'text-state-danger' : 'text-text-secondary',
                        )}
                      >
                        {metric.label}
                      </p>
                      <p
                        className={cn(
                          'mt-2 text-sm',
                          metric.unknown ? 'text-state-danger' : 'text-text-secondary',
                        )}
                      >
                        {metric.supporting}
                      </p>
                      <p className={cn('mt-2 industrial-data text-xs', toneClass(metric.tone))}>
                        {metric.note}
                      </p>
                    </div>
                    <Icon
                      className={cn(
                        'mt-1 h-5 w-5 shrink-0',
                        metric.unknown ? 'text-state-danger' : toneClass(metric.tone),
                      )}
                    />
                  </div>
                </button>
              )
            })}
          </div>
        </div>
      </section>

      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1600px] px-4 sm:px-6 lg:px-10">
          <p className="industrial-label">Telemetry Graphs</p>
          {grafana.enabled ? (
            <div className="mt-6 grid grid-cols-1 gap-10 xl:grid-cols-2">
              <GrafanaEmbed panelId={1} title="Cluster Task Throughput" subtitle="Live panel from Grafana" />
              <GrafanaEmbed panelId={2} title="Failure Pressure" subtitle="Critical and warning findings" />
            </div>
          ) : (
            <>
              <div className="mt-6 grid grid-cols-1 gap-10 xl:grid-cols-2">
                <GrafanaTimeSeries
                  title="Task Throughput"
                  subtitle="Running vs failed tasks over the last 90 minutes"
                  data={telemetry.throughput}
                  xKey="time"
                  lines={[
                    { key: 'running', label: 'Running', color: 'rgba(255,255,255,0.85)' },
                    { key: 'failed', label: 'Failed', color: '#F5A623' },
                  ]}
                />
                <GrafanaAreaSeries
                  title="Findings Pressure"
                  subtitle="Critical and warning findings over time"
                  data={telemetry.throughput}
                  xKey="time"
                  areas={[
                    { key: 'critical', label: 'Critical', color: '#F5A623' },
                    { key: 'warning', label: 'Warning', color: 'rgba(255,255,255,0.7)' },
                  ]}
                />
              </div>
              <div className="mt-10 grid grid-cols-1 gap-10 xl:grid-cols-2">
                <GrafanaTimeSeries
                  title="Node Availability"
                  subtitle="Managers and workers online"
                  data={telemetry.nodeHealth}
                  xKey="time"
                  lines={[
                    { key: 'managers', label: 'Managers', color: '#F5A623' },
                    { key: 'workers', label: 'Workers', color: 'rgba(255,255,255,0.8)' },
                  ]}
                />
                <GrafanaBarSeries
                  title="Current Finding Mix"
                  subtitle="Severity split in the current window"
                  data={findingDistribution}
                  xKey="bucket"
                  bars={[{ key: 'findings', label: 'Findings', color: '#F5A623' }]}
                />
              </div>
            </>
          )}
        </div>
      </section>

      <section className="px-4 pb-16 pt-10 sm:px-6 lg:px-10">
        <div className="mx-auto grid w-full max-w-[1600px] grid-cols-1 gap-14 xl:grid-cols-12">
          <div className="space-y-16 xl:col-span-8">
            <section aria-labelledby="status-title">
              <SectionHeading
                id="status-title"
                eyebrow="Cluster Status Summary"
                title="Operational Narrative"
                actionLabel="View Diagnostics"
                onAction={() => navigate('/diagnostics')}
                actionDisabled={model.disconnected}
              />
              <p className="mt-5 text-sm leading-relaxed text-text-secondary">{model.summary}</p>
              <ul className="mt-6 divide-y divide-white/10 border-t border-white/10">
                {model.checks.map((check) => (
                  <li key={check.id} className="industrial-row flex items-start justify-between gap-6">
                    <div className="flex min-w-0 items-start gap-3">
                      <span className="mt-0.5 shrink-0">{toneIcon(check.tone)}</span>
                      <div>
                        <p className="text-sm text-text-primary">{check.label}</p>
                        <p className="mt-1 text-sm text-text-secondary">{check.detail}</p>
                      </div>
                    </div>
                    <span className={cn('industrial-label whitespace-nowrap', toneClass(check.tone))}>
                      {check.status}
                    </span>
                  </li>
                ))}
              </ul>
            </section>

            <section aria-labelledby="findings-title">
              <SectionHeading
                id="findings-title"
                eyebrow="Active Findings"
                title="Incidents & Findings"
                actionLabel="Open Incidents"
                onAction={() => navigate('/incidents')}
              />
              {findingsError && model.findings.length === 0 ? (
                <div className="mt-6 border-t border-white/10 pt-6">
                  <p className="text-sm text-state-danger">Unable to load findings</p>
                  <p className="mt-1 text-sm text-text-secondary">{findingsError}</p>
                </div>
              ) : model.findings.length === 0 ? (
                <div className="mt-6 border-t border-white/10 pt-6">
                  <p className="text-sm text-text-primary">No active findings</p>
                  <p className="mt-1 text-sm text-text-secondary">
                    Latest diagnostics did not detect warning or critical issues.
                  </p>
                  <button
                    type="button"
                    onClick={() => {
                      void run()
                    }}
                    disabled={!canRunDiagnostics}
                    className={cn('industrial-action mt-3', !canRunDiagnostics && 'cursor-not-allowed opacity-35')}
                  >
                    Run Diagnostics
                  </button>
                </div>
              ) : (
                <ul className="mt-6 divide-y divide-white/10 border-t border-white/10">
                  {model.findings.map((finding) => (
                    <li key={finding.id}>
                      <button
                        type="button"
                        onClick={() => navigate(finding.to)}
                        className="industrial-row w-full text-left focus-visible:outline-none"
                      >
                        <div className="flex items-start justify-between gap-6">
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              {toneIcon(finding.tone)}
                              <p className={cn('truncate text-sm', toneClass(finding.tone))}>{finding.title}</p>
                            </div>
                            <p className="mt-1 text-xs text-text-secondary">
                              <span className="industrial-data">{finding.object}</span> |{' '}
                              <span className="industrial-data">{relativeTime(finding.timestamp)}</span>
                            </p>
                          </div>
                          <span className="industrial-label whitespace-nowrap text-text-secondary">Open</span>
                        </div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section aria-labelledby="events-title">
              <SectionHeading
                id="events-title"
                eyebrow="Recent Events"
                title="Cluster Timeline"
                actionLabel="Inspect Audit Trail"
                onAction={() => navigate('/audit')}
              />
              {model.events.length === 0 ? (
                <div className="mt-6 border-t border-white/10 pt-6">
                  <p className="text-sm text-text-primary">No recent events</p>
                  <p className="mt-1 text-sm text-text-secondary">
                    No cluster activity has been recorded in this session yet.
                  </p>
                </div>
              ) : (
                <ul className="mt-6 divide-y divide-white/10 border-t border-white/10">
                  {model.events.map((evt) => (
                    <li key={evt.id}>
                      <div className="industrial-row">
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              {toneIcon(evt.tone)}
                              <p className={cn('truncate text-sm', toneClass(evt.tone))}>{evt.title}</p>
                            </div>
                            <p className="mt-1 text-xs text-text-secondary">
                              <span className="industrial-data">{evt.source}</span>
                              {evt.metadata ? (
                                <span className="industrial-data text-text-tertiary"> | {evt.metadata}</span>
                              ) : null}
                            </p>
                          </div>
                          <span className="industrial-data text-xs text-text-secondary">{clock(evt.timestamp)}</span>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>

          <aside className="space-y-16 xl:col-span-4">
            <section aria-labelledby="quick-actions-title">
              <SectionHeading id="quick-actions-title" eyebrow="Quick Actions" title="Operator Shortcuts" />
              <ul className="mt-6 divide-y divide-white/10 border-t border-white/10">
                {[
                  {
                    id: 'q1',
                    label: running ? 'Running Diagnostics' : 'Run Diagnostics',
                    detail: 'Execute health checks',
                    onClick: () => {
                      void run()
                    },
                    disabled: !canRunDiagnostics,
                  },
                  {
                    id: 'q2',
                    label: 'Inspect Nodes',
                    detail: 'Review node availability',
                    onClick: () => navigate('/nodes'),
                    disabled: false,
                  },
                  {
                    id: 'q3',
                    label: 'Review Services',
                    detail: 'Inspect replica drift',
                    onClick: () => navigate('/services'),
                    disabled: false,
                  },
                  {
                    id: 'q4',
                    label: 'Open Incidents',
                    detail: 'Triage active incidents',
                    onClick: () => navigate('/incidents'),
                    disabled: false,
                  },
                  {
                    id: 'q5',
                    label: 'Check Audit Trail',
                    detail: 'Inspect operator writes',
                    onClick: () => navigate('/audit'),
                    disabled: false,
                  },
                ].map((action) => (
                  <li key={action.id}>
                    <button
                      type="button"
                      onClick={action.onClick}
                      disabled={action.disabled}
                      className={cn(
                        'industrial-row w-full text-left focus-visible:outline-none',
                        action.disabled && 'cursor-not-allowed opacity-35',
                      )}
                    >
                      <p className="text-sm text-text-primary">{action.label}</p>
                      <p className="mt-1 text-xs text-text-secondary">{action.detail}</p>
                    </button>
                  </li>
                ))}
              </ul>
            </section>

            <section aria-labelledby="diag-title">
              <SectionHeading id="diag-title" eyebrow="Diagnostics Widget" title="Latest Run" />
              <dl className={cn('mt-6 space-y-4 border-t border-white/10 pt-5', model.stale && 'opacity-40')}>
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Status</dt>
                  <dd
                    className={cn(
                      'industrial-data text-sm',
                      model.diagnostics.status === 'Degraded' || model.diagnostics.status === 'Offline'
                        ? 'text-state-danger'
                        : 'text-text-primary',
                    )}
                  >
                    {model.diagnostics.status}
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Last run</dt>
                  <dd className="industrial-data text-sm text-text-secondary">
                    {model.diagnostics.lastRun
                      ? relativeTime(new Date(model.diagnostics.lastRun).toISOString())
                      : 'not available'}
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Duration</dt>
                  <dd className="industrial-data text-sm text-text-secondary">
                    {durationLabel(model.diagnostics.durationMs)}
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Findings</dt>
                  <dd className="industrial-data text-sm text-text-secondary">
                    {model.diagnostics.findingsCount}
                  </dd>
                </div>
                <button
                  type="button"
                  onClick={() => {
                    void run()
                  }}
                  disabled={!canRunDiagnostics}
                  className={cn('industrial-action pt-1', !canRunDiagnostics && 'cursor-not-allowed opacity-35')}
                >
                  <DiagnosticsIcon className="h-3.5 w-3.5" />
                  Rerun Diagnostics
                </button>
              </dl>
            </section>

            <section aria-labelledby="posture-title">
              <SectionHeading id="posture-title" eyebrow="Cluster Posture" title="At-a-Glance" />
              <dl className={cn('mt-6 space-y-4 border-t border-white/10 pt-5', model.stale && 'opacity-40')}>
                {model.posture.map((item) => (
                  <div key={item.id} className="flex items-center justify-between gap-4">
                    <dt className="industrial-label">{item.label}</dt>
                    <dd className={cn('industrial-data text-sm', toneClass(item.tone))}>{item.value}</dd>
                  </div>
                ))}
              </dl>
            </section>

            <section aria-labelledby="guidance-title">
              <SectionHeading id="guidance-title" eyebrow="Guidance" title="Next Steps" />
              <div className="mt-6 border-t border-white/10 pt-5">
                <p className={cn('text-sm', model.disconnected ? 'text-state-danger' : 'text-text-primary')}>
                  {model.guidance.title}
                </p>
                <p className="mt-2 text-sm leading-relaxed text-text-secondary">
                  {model.guidance.description}
                </p>
                <ol className="mt-4 space-y-2">
                  {model.guidance.steps.map((step, idx) => (
                    <li key={step} className="flex items-start gap-3 text-sm text-text-secondary">
                      <span className="industrial-data text-xs text-text-tertiary">
                        {String(idx + 1).padStart(2, '0')}
                      </span>
                      <span>{step}</span>
                    </li>
                  ))}
                </ol>
              </div>
            </section>

            {isDemoExperience ? (
              <section aria-labelledby="demo-mode-title">
                <SectionHeading id="demo-mode-title" eyebrow="Demo Mode" title="How It Works" />
                <div className="mt-6 border-t border-white/10 pt-5">
                  <p className="text-sm text-text-primary">
                    Demo mode uses deterministic baseline snapshots plus synthetic telemetry jitter so you can test workflows without a live swarm.
                  </p>
                  <ul className="mt-4 space-y-2 text-sm text-text-secondary">
                    <li className="flex items-start gap-3">
                      <span className="industrial-label text-text-tertiary">Baseline</span>
                      <span>Healthy, degraded, and disconnected scenarios are seeded from in-app model data.</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <span className="industrial-label text-text-tertiary">Telemetry</span>
                      <span>Charts are generated from synthetic 90-minute series and reflect the selected scenario profile.</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <span className="industrial-label text-text-tertiary">Diagnostics</span>
                      <span>When diagnostics API returns no findings in demo, synthetic findings are used for triage flows.</span>
                    </li>
                    <li className="flex items-start gap-3">
                      <span className="industrial-label text-text-tertiary">Actions</span>
                      <span>Buttons still call the same handlers as live mode, so interaction behavior matches production paths.</span>
                    </li>
                  </ul>

                  <div className="mt-5 border-t border-white/10 pt-4">
                    <p className="industrial-label text-text-secondary">Scenario Controls</p>
                    <p className="mt-1 industrial-data text-xs text-text-tertiary">
                      Active: {activeScenario.toUpperCase()}
                    </p>
                    <div className="mt-3 flex flex-wrap items-center gap-5">
                      <button
                        type="button"
                        onClick={() => setScenario('healthy')}
                        className={cn('industrial-action', activeScenario === 'healthy' && 'industrial-action-accent')}
                      >
                        Healthy
                      </button>
                      <button
                        type="button"
                        onClick={() => setScenario('degraded')}
                        className={cn('industrial-action', activeScenario === 'degraded' && 'industrial-action-accent')}
                      >
                        Degraded
                      </button>
                      <button
                        type="button"
                        onClick={() => setScenario('disconnected')}
                        className={cn('industrial-action', activeScenario === 'disconnected' && 'industrial-action-accent')}
                      >
                        Disconnected
                      </button>
                      {scenarioKind ? (
                        <button type="button" onClick={clearScenario} className="industrial-action">
                          Use Live Feed
                        </button>
                      ) : null}
                    </div>
                  </div>
                </div>
              </section>
            ) : null}
          </aside>
        </div>
      </section>

      {(incidentsLoading || running) && (
        <div className="fixed bottom-4 right-4 flex items-center gap-2 text-xs text-text-secondary">
          <span className="h-2 w-2 rounded-full bg-text-secondary" aria-hidden="true" />
          {running ? 'Diagnostics run in progress' : 'Syncing incidents'}
        </div>
      )}
    </div>
  )
}
