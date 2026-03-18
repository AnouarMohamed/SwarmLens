import { useEffect, useMemo, useRef } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { GrafanaEmbed } from '../../components/charts/GrafanaEmbed'
import { relativeTime } from '../../lib/utils'
import { grafanaConfig } from '../../lib/grafana'
import { buildOverviewTelemetry } from '../../lib/telemetry'
import { useClusterStore } from '../../store/clusterStore'
import { useDiagnosticsStore } from '../../store/diagnosticsStore'
import { useIncidentStore } from '../../store/incidentStore'
import {
  GrafanaAreaSeries,
  GrafanaBarSeries,
  GrafanaTimeSeries,
} from '../../components/charts/GrafanaCharts'
import type { Severity, SwarmEvent } from '../../types'
import {
  ArrowRightIcon,
  CheckCircleIcon,
  ClockIcon,
  DiagnosticsIcon,
  DisconnectIcon,
  RefreshIcon,
  WarningIcon,
} from '../../components/ui/icons'

type Tone = 'stable' | 'critical' | 'unknown'
type Scenario = 'healthy' | 'degraded' | 'disconnected'

type Model = {
  disconnected: boolean
  clusterName: string
  environment: string
  endpoint: string
  healthLabel: 'Healthy' | 'Degraded' | 'Unknown'
  healthTone: Tone
  summary: string
  freshness: string
  staleNote: string
  metrics: Array<{
    id: string
    label: string
    value: string
    detail: string
    note: string
    tone: Tone
    to: string
  }>
  checks: Array<{ id: string; label: string; detail: string; state: string; tone: Tone }>
  findings: Array<{
    id: string
    title: string
    object: string
    timestamp: string
    tone: Tone
    to: string
  }>
  events: Array<{
    id: string
    title: string
    source: string
    timestamp: string
    metadata?: string
    tone: Tone
  }>
  attention: Array<{ id: string; name: string; replicas: string; state: string; issue: string }>
  diagnostics: {
    status: string
    lastRun: number | null
    durationMs: number | null
    findingsCount: number
  }
  posture: Array<{ id: string; label: string; value: string; tone: Tone }>
  guidance: { title: string; description: string; steps: string[] }
}

const minutesAgo = (m: number) => new Date(Date.now() - m * 60000).toISOString()

function cn(...parts: Array<string | false | undefined>) {
  return parts.filter(Boolean).join(' ')
}

function toneClass(tone: Tone) {
  if (tone === 'critical') return 'text-state-danger'
  if (tone === 'unknown') return 'text-text-secondary'
  return 'text-text-primary'
}

function toneIcon(tone: Tone) {
  if (tone === 'critical') return <WarningIcon className="h-4 w-4 text-state-danger" />
  if (tone === 'unknown') return <DisconnectIcon className="h-4 w-4 text-text-secondary" />
  return <CheckCircleIcon className="h-4 w-4 text-text-secondary" />
}

function toneFromSeverity(severity: Severity): Tone {
  return severity === 'critical' || severity === 'high' || severity === 'medium'
    ? 'critical'
    : 'stable'
}

function toneFromEvent(evt: SwarmEvent): Tone {
  const t = `${evt.action} ${evt.message}`.toLowerCase()
  return /(failed|error|reject|down|unreachable|timeout|panic)/.test(t) ? 'critical' : 'stable'
}

function durationLabel(ms: number | null) {
  if (!ms) return 'n/a'
  return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(1)}s`
}

function clock(ts: string) {
  return new Date(ts).toLocaleTimeString([], {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  })
}

function delta(v: number) {
  return v > 0 ? `+${v}` : `${v}`
}

function leadingInt(value: string | undefined, fallback: number) {
  if (!value) return fallback
  const match = value.match(/\d+/)
  if (!match) return fallback
  const parsed = Number(match[0])
  return Number.isFinite(parsed) ? parsed : fallback
}

function scenarioModel(kind: Scenario, endpoint: string): Model {
  if (kind === 'disconnected') {
    return {
      disconnected: true,
      clusterName: 'cluster/swrm-prod-eu-01',
      environment: 'PROD',
      endpoint,
      healthLabel: 'Unknown',
      healthTone: 'unknown',
      summary: 'SwarmLens is not connected to the cluster. Live telemetry and controls are paused.',
      freshness: 'No successful sync yet',
      staleNote: 'Showing last known data from previous telemetry snapshot.',
      metrics: [
        {
          id: 'h',
          label: 'Cluster Health',
          value: 'Unknown',
          detail: 'Last known: 38/42 healthy',
          note: 'Telemetry stale',
          tone: 'unknown',
          to: '/diagnostics',
        },
        {
          id: 'n',
          label: 'Nodes',
          value: '8',
          detail: '3 managers / 5 workers',
          note: 'Last known snapshot',
          tone: 'unknown',
          to: '/nodes',
        },
        {
          id: 't',
          label: 'Running Tasks',
          value: '124',
          detail: 'Last known: 4 failed',
          note: 'No live updates',
          tone: 'unknown',
          to: '/tasks',
        },
        {
          id: 'f',
          label: 'Active Findings',
          value: '5',
          detail: '2 critical / 3 warning',
          note: 'Reconnect to update',
          tone: 'unknown',
          to: '/incidents',
        },
      ],
      checks: [
        {
          id: 'c1',
          label: 'Control plane reachable',
          detail: 'Connection lost to manager endpoint',
          state: 'Unknown',
          tone: 'unknown',
        },
        {
          id: 'c2',
          label: 'Node quorum',
          detail: 'Last known: degraded',
          state: 'Unknown',
          tone: 'unknown',
        },
      ],
      findings: [],
      events: [],
      attention: [],
      diagnostics: {
        status: 'Offline',
        lastRun: Date.now() - 900000,
        durationMs: 1800,
        findingsCount: 5,
      },
      posture: [
        { id: 'p1', label: 'Managers Online', value: 'unknown', tone: 'unknown' },
        { id: 'p2', label: 'Workers Online', value: 'unknown', tone: 'unknown' },
        { id: 'p3', label: 'Total Replicas', value: '42 (last known)', tone: 'unknown' },
      ],
      guidance: {
        title: 'Reconnect Required',
        description:
          'Restore connection to continue diagnostics, event stream, and write operations.',
        steps: [
          'Check manager reachability.',
          'Validate endpoint and credentials.',
          'Reconnect and refresh telemetry.',
        ],
      },
    }
  }

  if (kind === 'degraded') {
    return {
      disconnected: false,
      clusterName: 'cluster/swrm-prod-eu-01',
      environment: 'PROD',
      endpoint,
      healthLabel: 'Degraded',
      healthTone: 'critical',
      summary:
        'Cluster is partially degraded. Replica drift was detected on 2 services and 1 worker is unreachable.',
      freshness: 'Last sync 36s ago',
      staleNote: 'Telemetry is live.',
      metrics: [
        {
          id: 'h',
          label: 'Cluster Health',
          value: 'Degraded',
          detail: '38/42 replicas healthy',
          note: 'Checked 2m ago',
          tone: 'critical',
          to: '/diagnostics',
        },
        {
          id: 'n',
          label: 'Nodes',
          value: '8',
          detail: '3 managers / 5 workers',
          note: '1 unavailable',
          tone: 'critical',
          to: '/nodes',
        },
        {
          id: 't',
          label: 'Running Tasks',
          value: '124',
          detail: '120 healthy / 4 failed',
          note: '-3 since last refresh',
          tone: 'critical',
          to: '/tasks',
        },
        {
          id: 'f',
          label: 'Active Findings',
          value: '5',
          detail: '2 critical / 3 warning',
          note: 'Open incidents now',
          tone: 'critical',
          to: '/incidents',
        },
      ],
      checks: [
        {
          id: 'c1',
          label: 'Control plane reachable',
          detail: 'Manager API responding in degraded mode',
          state: 'Nominal',
          tone: 'stable',
        },
        {
          id: 'c2',
          label: 'Node quorum',
          detail: '2/3 managers reachable',
          state: 'Degraded',
          tone: 'critical',
        },
      ],
      findings: [
        {
          id: 'f1',
          title: 'Replica shortfall in payments-worker',
          object: 'service/payments-worker',
          timestamp: minutesAgo(6),
          tone: 'critical',
          to: '/incidents',
        },
        {
          id: 'f2',
          title: 'Worker node unreachable',
          object: 'node/worker-02',
          timestamp: minutesAgo(11),
          tone: 'critical',
          to: '/nodes',
        },
      ],
      events: [
        {
          id: 'e1',
          title: 'Task rejected due to placement constraint',
          source: 'payments-worker',
          timestamp: minutesAgo(4),
          metadata: 'node.labels.zone=eu-central',
          tone: 'critical',
        },
        {
          id: 'e2',
          title: 'Node heartbeat delayed',
          source: 'worker-02',
          timestamp: minutesAgo(8),
          metadata: 'reachability changed to unreachable',
          tone: 'critical',
        },
      ],
      attention: [
        {
          id: 'a1',
          name: 'payments-worker',
          replicas: '0 / 2',
          state: 'Replica Drift',
          issue: '3 placement failures in 10m',
        },
        {
          id: 'a2',
          name: 'api-gateway',
          replicas: '5 / 6',
          state: 'Task Failures',
          issue: '1 restart in rolling update',
        },
      ],
      diagnostics: {
        status: 'Degraded',
        lastRun: Date.now() - 120000,
        durationMs: 1700,
        findingsCount: 5,
      },
      posture: [
        { id: 'p1', label: 'Managers Online', value: '2 / 3', tone: 'critical' },
        { id: 'p2', label: 'Workers Online', value: '4 / 5', tone: 'critical' },
        { id: 'p3', label: 'Total Replicas', value: '42', tone: 'stable' },
      ],
      guidance: {
        title: 'Mitigation Sequence',
        description: 'Recover control-plane quorum first, then clear placement and replica drift.',
        steps: [
          'Restore manager reachability.',
          'Resolve placement constraints.',
          'Re-run diagnostics and validate findings.',
        ],
      },
    }
  }

  return {
    disconnected: false,
    clusterName: 'cluster/swrm-dev-lab-01',
    environment: 'DEMO',
    endpoint,
    healthLabel: 'Healthy',
    healthTone: 'stable',
    summary:
      'Cluster is healthy. No active critical findings. Scheduling and replicas are nominal.',
    freshness: 'Last sync 28s ago',
    staleNote: 'Telemetry is live.',
    metrics: [
      {
        id: 'h',
        label: 'Cluster Health',
        value: 'Healthy',
        detail: '24/24 replicas healthy',
        note: 'Checked 2m ago',
        tone: 'stable',
        to: '/diagnostics',
      },
      {
        id: 'n',
        label: 'Nodes',
        value: '5',
        detail: '3 managers / 2 workers',
        note: '0 unavailable',
        tone: 'stable',
        to: '/nodes',
      },
      {
        id: 't',
        label: 'Running Tasks',
        value: '72',
        detail: '72 healthy / 0 failed',
        note: '+1 since last refresh',
        tone: 'stable',
        to: '/tasks',
      },
      {
        id: 'f',
        label: 'Active Findings',
        value: '0',
        detail: 'No critical or warning findings',
        note: 'Diagnostics completed 8m ago',
        tone: 'stable',
        to: '/incidents',
      },
    ],
    checks: [
      {
        id: 'c1',
        label: 'Control plane reachable',
        detail: 'API and manager leadership responsive',
        state: 'Nominal',
        tone: 'stable',
      },
      {
        id: 'c2',
        label: 'Node quorum',
        detail: '3/3 managers reachable',
        state: 'Nominal',
        tone: 'stable',
      },
    ],
    findings: [],
    events: [
      {
        id: 'e1',
        title: 'Diagnostics completed with zero findings',
        source: 'diagnostics',
        timestamp: minutesAgo(8),
        tone: 'stable',
      },
    ],
    attention: [],
    diagnostics: {
      status: 'Clear',
      lastRun: Date.now() - 480000,
      durationMs: 900,
      findingsCount: 0,
    },
    posture: [
      { id: 'p1', label: 'Managers Online', value: '3 / 3', tone: 'stable' },
      { id: 'p2', label: 'Workers Online', value: '2 / 2', tone: 'stable' },
      { id: 'p3', label: 'Total Replicas', value: '24', tone: 'stable' },
    ],
    guidance: {
      title: 'No Active Workload Issues',
      description: 'Telemetry is stable and no mitigation workflow is required right now.',
      steps: [
        'Run diagnostics on regular cadence.',
        'Review audit trail for writes.',
        'Inspect services before rollouts.',
      ],
    },
  }
}

function OverviewSkeleton() {
  return (
    <div className="animate-pulse">
      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 lg:px-10">
          <div className="h-3 w-28 bg-white/10" />
          <div className="mt-5 h-20 w-full max-w-3xl bg-white/10" />
        </div>
      </section>
      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 lg:px-10">
          <div className="grid grid-cols-2 gap-8 xl:grid-cols-4">
            {Array.from({ length: 4 }).map((_, i) => (
              <div key={i} className="h-20 bg-white/10" />
            ))}
          </div>
        </div>
      </section>
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
    loading,
    error,
    connectionState,
    lastRefresh,
    fetchAll,
  } = useClusterStore()
  const {
    findings,
    fetch: fetchDiagnostics,
    run,
    running,
    lastRun,
    lastDurationMs,
    error: diagError,
  } = useDiagnosticsStore()
  const {
    incidents,
    fetch: fetchIncidents,
    loading: incidentsLoading,
    error: incidentsError,
  } = useIncidentStore()

  const prevRunningTasksRef = useRef<number | null>(null)
  const prevRefreshRef = useRef(0)

  useEffect(() => {
    void fetchDiagnostics()
    void fetchIncidents()
  }, [fetchDiagnostics, fetchIncidents])

  const runningTasks = tasks.filter((t) => t.currentState === 'running').length
  const failedTasks = tasks.filter(
    (t) => t.currentState === 'failed' || t.currentState === 'rejected',
  ).length

  useEffect(() => {
    if (lastRefresh && lastRefresh !== prevRefreshRef.current) {
      prevRefreshRef.current = lastRefresh
      prevRunningTasksRef.current = runningTasks
    }
  }, [lastRefresh, runningTasks])

  const taskDelta =
    prevRunningTasksRef.current === null ? 0 : runningTasks - prevRunningTasksRef.current
  const hasData =
    Boolean(swarm) ||
    nodes.length > 0 ||
    services.length > 0 ||
    tasks.length > 0 ||
    events.length > 0 ||
    findings.length > 0 ||
    incidents.length > 0
  const disconnected = connectionState === 'disconnected' || Boolean(error)
  const scenario = searchParams.get('scenario')
  const scenarioKind: Scenario | null =
    scenario === 'healthy' || scenario === 'degraded' || scenario === 'disconnected'
      ? scenario
      : !hasData && !loading
        ? disconnected
          ? 'disconnected'
          : 'healthy'
        : null
  const endpoint = (import.meta.env.VITE_API_BASE as string | undefined) ?? '/api/v1'

  const model = useMemo<Model>(() => {
    if (scenarioKind) return scenarioModel(scenarioKind, endpoint)

    const managers = nodes.filter((n) => n.role === 'manager')
    const workers = nodes.filter((n) => n.role === 'worker')
    const unavailableNodes = nodes.filter((n) => n.state !== 'ready').length
    const desiredReplicas = services.reduce((sum, s) => sum + s.desiredReplicas, 0)
    const runningReplicas = services.reduce((sum, s) => sum + s.runningTasks, 0)
    const riskyServices = services.filter(
      (s) => s.runningTasks < s.desiredReplicas || s.failedTasks > 0,
    )
    const criticalCount = findings.filter((f) => f.severity === 'critical').length
    const warningCount = findings.filter(
      (f) => f.severity === 'high' || f.severity === 'medium',
    ).length
    const degraded = criticalCount > 0 || unavailableNodes > 0 || riskyServices.length > 0
    const healthLabel: Model['healthLabel'] = disconnected
      ? 'Unknown'
      : degraded
        ? 'Degraded'
        : 'Healthy'
    const healthTone: Tone = disconnected ? 'unknown' : degraded ? 'critical' : 'stable'
    const freshness = lastRefresh
      ? `Last sync ${relativeTime(new Date(lastRefresh).toISOString())}`
      : 'No successful sync yet'
    const stale = !lastRefresh || Date.now() - lastRefresh > 120000

    return {
      disconnected,
      clusterName: swarm?.clusterID ? `cluster/${swarm.clusterID.slice(0, 16)}` : 'cluster/unset',
      environment: swarm?.mode?.toUpperCase() ?? 'DEMO',
      endpoint,
      healthLabel,
      healthTone,
      summary: disconnected
        ? 'SwarmLens is not connected to the cluster. Live telemetry and controls are paused.'
        : degraded
          ? `Cluster is partially degraded. ${riskyServices.length} services require attention and ${unavailableNodes} nodes are unavailable.`
          : 'Cluster is healthy. No active critical findings. Scheduling and replicas are nominal.',
      freshness,
      staleNote: disconnected
        ? lastRefresh
          ? 'Showing last known data.'
          : 'No previous telemetry snapshot available.'
        : stale
          ? 'Telemetry is stale.'
          : 'Telemetry is live.',
      metrics: [
        {
          id: 'h',
          label: 'Cluster Health',
          value: healthLabel,
          detail:
            desiredReplicas > 0
              ? `${runningReplicas}/${desiredReplicas} replicas healthy`
              : 'No scheduled replicas',
          note: lastRefresh
            ? `Checked ${relativeTime(new Date(lastRefresh).toISOString())}`
            : 'No check data',
          tone: healthTone,
          to: '/diagnostics',
        },
        {
          id: 'n',
          label: 'Nodes',
          value: `${nodes.length}`,
          detail: `${managers.length} managers / ${workers.length} workers`,
          note: `${unavailableNodes} unavailable`,
          tone: unavailableNodes > 0 ? 'critical' : disconnected ? 'unknown' : 'stable',
          to: '/nodes',
        },
        {
          id: 't',
          label: 'Running Tasks',
          value: `${runningTasks}`,
          detail: `${Math.max(runningTasks - failedTasks, 0)} healthy / ${failedTasks} failed`,
          note: `${delta(taskDelta)} since last refresh`,
          tone: failedTasks > 0 ? 'critical' : disconnected ? 'unknown' : 'stable',
          to: '/tasks',
        },
        {
          id: 'f',
          label: 'Active Findings',
          value: `${findings.length}`,
          detail: `${criticalCount} critical / ${warningCount} warning`,
          note: findings.length > 0 ? 'Review incidents and diagnostics' : 'No active findings',
          tone:
            criticalCount > 0 || warningCount > 0
              ? 'critical'
              : disconnected
                ? 'unknown'
                : 'stable',
          to: '/incidents',
        },
      ],
      checks: [
        {
          id: 'c1',
          label: 'Control plane reachable',
          detail: disconnected
            ? 'Connection lost to manager endpoint'
            : 'Manager API and leadership path responding',
          state: disconnected ? 'Unknown' : 'Nominal',
          tone: disconnected ? 'unknown' : 'stable',
        },
        {
          id: 'c2',
          label: 'Node quorum',
          detail: swarm?.quorumHealthy ? 'Manager quorum healthy' : 'Manager quorum at risk',
          state: disconnected ? 'Unknown' : swarm?.quorumHealthy ? 'Nominal' : 'Degraded',
          tone: disconnected ? 'unknown' : swarm?.quorumHealthy ? 'stable' : 'critical',
        },
      ],
      findings: [
        ...incidents
          .filter((i) => i.status !== 'resolved')
          .slice(0, 4)
          .map((i) => ({
            id: `inc-${i.id}`,
            title: i.title,
            object: i.affectedServices[0] ? `service/${i.affectedServices[0]}` : 'cluster',
            timestamp: i.updatedAt || i.createdAt,
            tone: toneFromSeverity(i.severity),
            to: '/incidents',
          })),
        ...findings.slice(0, 4).map((f) => ({
          id: `f-${f.id}`,
          title: f.message,
          object: f.resource,
          timestamp: f.detectedAt,
          tone: toneFromSeverity(f.severity),
          to: '/diagnostics',
        })),
      ].slice(0, 6),
      events: events.slice(0, 10).map((evt, idx) => ({
        id: `${evt.timestamp}-${idx}`,
        title: `${evt.action} ${evt.message}`.trim(),
        source: evt.actor || evt.type,
        timestamp: evt.timestamp,
        metadata: evt.type,
        tone: toneFromEvent(evt),
      })),
      attention: riskyServices.slice(0, 6).map((s) => ({
        id: s.id,
        name: s.name,
        replicas: `${s.runningTasks} / ${s.desiredReplicas}`,
        state: s.runningTasks < s.desiredReplicas ? 'Replica Drift' : 'Task Failures',
        issue:
          s.failedTasks > 0
            ? `${s.failedTasks} failed tasks`
            : `${s.desiredReplicas - s.runningTasks} replicas missing`,
      })),
      diagnostics: {
        status: disconnected
          ? 'Offline'
          : running
            ? 'Running'
            : criticalCount > 0 || warningCount > 0
              ? 'Degraded'
              : 'Clear',
        lastRun,
        durationMs: lastDurationMs,
        findingsCount: findings.length,
      },
      posture: [
        {
          id: 'p1',
          label: 'Managers Online',
          value: `${managers.filter((n) => n.state === 'ready').length} / ${managers.length}`,
          tone: disconnected ? 'unknown' : unavailableNodes > 0 ? 'critical' : 'stable',
        },
        {
          id: 'p2',
          label: 'Workers Online',
          value: `${workers.filter((n) => n.state === 'ready').length} / ${workers.length}`,
          tone: disconnected ? 'unknown' : unavailableNodes > 0 ? 'critical' : 'stable',
        },
        {
          id: 'p3',
          label: 'Total Replicas',
          value: `${desiredReplicas}`,
          tone: disconnected ? 'unknown' : 'stable',
        },
      ],
      guidance:
        services.length === 0
          ? {
              title: 'No Active Workloads Yet',
              description:
                'This cluster has no active workload telemetry. Provision a stack to start operational tracking.',
              steps: [
                'Create or deploy a stack.',
                'Verify managers and workers are connected.',
                'Run diagnostics to establish baseline findings.',
              ],
            }
          : disconnected
            ? {
                title: 'Connection Required',
                description:
                  'Reconnect to restore live telemetry, diagnostics execution, and operator actions.',
                steps: [
                  'Validate manager endpoint reachability.',
                  'Reconnect the cluster session.',
                  'Refresh and confirm live telemetry.',
                ],
              }
            : {
                title: 'Operational Guidance',
                description:
                  'Prioritize risk reduction in services with replica drift or failed tasks.',
                steps: [
                  'Inspect services with degraded replicas.',
                  'Run diagnostics to confirm remediation.',
                  'Check audit trail for recent changes.',
                ],
              },
    }
  }, [
    scenarioKind,
    endpoint,
    swarm,
    nodes,
    services,
    events,
    findings,
    incidents,
    disconnected,
    lastRefresh,
    runningTasks,
    failedTasks,
    taskDelta,
    running,
    lastRun,
    lastDurationMs,
  ])

  const criticalFindings = findings.filter((f) => f.severity === 'critical').length
  const warningFindings = findings.filter(
    (f) => f.severity === 'high' || f.severity === 'medium',
  ).length
  const managersOnline = nodes.filter((n) => n.role === 'manager' && n.state === 'ready').length
  const workersOnline = nodes.filter((n) => n.role === 'worker' && n.state === 'ready').length

  const telemetry = useMemo(
    () =>
      buildOverviewTelemetry({
        runningTasks:
          runningTasks > 0
            ? runningTasks
            : leadingInt(
                model.metrics.find((metric) => metric.id === 't')?.value,
                model.healthTone === 'critical' ? 124 : 72,
              ),
        failedTasks: failedTasks > 0 ? failedTasks : model.healthTone === 'critical' ? 4 : 0,
        managersOnline:
          managersOnline > 0 ? managersOnline : leadingInt(model.posture[0]?.value, 2),
        workersOnline: workersOnline > 0 ? workersOnline : leadingInt(model.posture[1]?.value, 4),
        criticalFindings:
          criticalFindings > 0 ? criticalFindings : model.healthTone === 'critical' ? 2 : 0,
        warningFindings:
          warningFindings > 0 ? warningFindings : model.healthTone === 'critical' ? 3 : 0,
        disconnected: model.disconnected,
        degraded: model.healthTone === 'critical',
      }),
    [
      runningTasks,
      failedTasks,
      managersOnline,
      workersOnline,
      criticalFindings,
      warningFindings,
      model.metrics,
      model.posture,
      model.disconnected,
      model.healthTone,
    ],
  )

  const findingDistribution = useMemo(
    () => [
      { bucket: 'Critical', findings: criticalFindings },
      { bucket: 'Warning', findings: warningFindings },
      {
        bucket: 'Info',
        findings: Math.max(findings.length - (criticalFindings + warningFindings), 0),
      },
    ],
    [criticalFindings, warningFindings, findings.length],
  )
  const grafana = grafanaConfig()

  const canRunDiagnostics = !model.disconnected && !running
  const findingsError = diagError || incidentsError
  const eventsError = error && events.length === 0 && !scenarioKind

  if (loading && !hasData && !scenarioKind) return <OverviewSkeleton />

  return (
    <div className="min-h-full bg-app text-text-primary">
      <section
        className={cn('industrial-section', model.disconnected && 'min-h-[40vh] flex items-center')}
      >
        <div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 lg:px-10">
          <p className="industrial-label">Cluster State</p>
          <h2
            className={cn(
              'industrial-data mt-5 leading-[0.92] tracking-[-0.03em]',
              model.disconnected
                ? 'industrial-pulse text-[clamp(2.8rem,8vw,7rem)] text-state-danger'
                : 'text-[clamp(2.6rem,7vw,5.8rem)] text-text-primary',
            )}
          >
            {model.clusterName}
          </h2>
          <p className={cn('mt-2 industrial-label', toneClass(model.healthTone))}>
            {model.healthLabel}
          </p>
          <p className="mt-4 max-w-3xl text-base leading-relaxed text-text-secondary">
            {model.summary}
          </p>
          <p className="mt-2 text-sm text-text-tertiary">{model.staleNote}</p>
          <div className="mt-8 flex flex-wrap items-center gap-8">
            <button
              type="button"
              onClick={() => {
                void run()
              }}
              disabled={!canRunDiagnostics}
              className={cn(
                'industrial-action industrial-action-accent',
                !canRunDiagnostics && 'cursor-not-allowed opacity-30',
              )}
            >
              <DiagnosticsIcon className="h-3.5 w-3.5" />
              {running ? 'Running Diagnostics' : 'Run Diagnostics'}
            </button>
            <button
              type="button"
              onClick={() => {
                void fetchAll()
              }}
              className="industrial-action"
            >
              <RefreshIcon className="h-3.5 w-3.5" />
              Refresh Telemetry
            </button>
            <button type="button" onClick={() => navigate('/audit')} className="industrial-action">
              <ClockIcon className="h-3.5 w-3.5" />
              Open Audit Trail
            </button>
          </div>
        </div>
      </section>
      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 lg:px-10">
          <p className="industrial-label">Operational Metrics</p>
          <div className="mt-6 grid grid-cols-2 gap-x-8 gap-y-8 xl:grid-cols-4">
            {model.metrics.map((metric) => (
              <button
                key={metric.id}
                type="button"
                onClick={() => navigate(metric.to)}
                className="industrial-row w-full text-left focus-visible:outline-none"
              >
                <p className={cn('industrial-metric', toneClass(metric.tone))}>{metric.value}</p>
                <p className="mt-3 industrial-label text-text-secondary">{metric.label}</p>
                <p className="mt-2 text-sm text-text-secondary">{metric.detail}</p>
                <p className={cn('mt-2 industrial-data text-xs', toneClass(metric.tone))}>
                  {metric.note}
                </p>
              </button>
            ))}
          </div>
        </div>
      </section>

      <section className="industrial-section">
        <div className="mx-auto w-full max-w-[1440px] px-4 sm:px-6 lg:px-10">
          <p className="industrial-label">Telemetry Graphs</p>
          {grafana.enabled ? (
            <>
              <div className="mt-6 grid grid-cols-1 gap-10 xl:grid-cols-2">
                <GrafanaEmbed
                  panelId={1}
                  title="Cluster Task Throughput"
                  subtitle="Live panel from Grafana"
                />
                <GrafanaEmbed
                  panelId={2}
                  title="Failure Pressure"
                  subtitle="Critical and warning findings"
                />
              </div>
              <div className="mt-10 grid grid-cols-1 gap-10 xl:grid-cols-2">
                <GrafanaEmbed
                  panelId={3}
                  title="Node Availability"
                  subtitle="Managers and workers online"
                />
                <GrafanaEmbed
                  panelId={4}
                  title="Severity Distribution"
                  subtitle="Current findings mix"
                />
              </div>
            </>
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
                  subtitle="Critical and warning signals over time"
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

      <section className="px-4 pb-16 pt-12 sm:px-6 lg:px-10">
        <div className="mx-auto grid w-full max-w-[1440px] grid-cols-1 gap-14 xl:grid-cols-12">
          <div className="space-y-16 xl:col-span-8">
            <section>
              <div className="flex items-end justify-between gap-6">
                <div>
                  <p className="industrial-label">Status Summary</p>
                  <h3 className="mt-2 font-heading text-[1.9rem] uppercase leading-none tracking-[0.04em]">
                    Cluster Status
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={() => navigate('/diagnostics')}
                  className="industrial-action"
                >
                  View Diagnostics
                  <ArrowRightIcon className="h-3.5 w-3.5" />
                </button>
              </div>
              <ul className="mt-6 divide-y divide-border-muted">
                {model.checks.map((check) => (
                  <li
                    key={check.id}
                    className="industrial-row flex items-start justify-between gap-6"
                  >
                    <div className="flex min-w-0 items-start gap-3">
                      <span className="mt-0.5 shrink-0">{toneIcon(check.tone)}</span>
                      <div>
                        <p className="text-sm text-text-primary">{check.label}</p>
                        <p className="mt-1 text-sm text-text-secondary">{check.detail}</p>
                      </div>
                    </div>
                    <span
                      className={cn('industrial-label whitespace-nowrap', toneClass(check.tone))}
                    >
                      {check.state}
                    </span>
                  </li>
                ))}
              </ul>
            </section>

            <section>
              <div className="flex items-end justify-between gap-6">
                <div>
                  <p className="industrial-label">Active Findings</p>
                  <h3 className="mt-2 font-heading text-[1.9rem] uppercase leading-none tracking-[0.04em]">
                    Incidents & Findings
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={() => navigate('/incidents')}
                  className="industrial-action"
                >
                  Open Incidents
                  <ArrowRightIcon className="h-3.5 w-3.5" />
                </button>
              </div>
              {findingsError && model.findings.length === 0 ? (
                <div className="mt-6 border-t border-state-danger/50 pt-6">
                  <p className="text-sm text-state-danger">Unable to load findings</p>
                  <p className="mt-1 text-sm text-text-secondary">{findingsError}</p>
                </div>
              ) : model.findings.length === 0 ? (
                <div className="mt-6 border-t border-border-muted pt-6">
                  <p className="text-sm text-text-primary">No active findings</p>
                  <p className="mt-1 text-sm text-text-secondary">
                    The latest diagnostics run did not detect warnings or critical issues.
                  </p>
                </div>
              ) : (
                <ul className="mt-6 divide-y divide-border-muted">
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
                              <p className={cn('truncate text-sm', toneClass(finding.tone))}>
                                {finding.title}
                              </p>
                            </div>
                            <p className="mt-1 font-mono text-xs text-text-secondary">
                              {finding.object} | {relativeTime(finding.timestamp)}
                            </p>
                          </div>
                          <span className="industrial-label whitespace-nowrap text-text-secondary">
                            Open
                          </span>
                        </div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section>
              <div className="flex items-end justify-between gap-6">
                <div>
                  <p className="industrial-label">Recent Activity</p>
                  <h3 className="mt-2 font-heading text-[1.9rem] uppercase leading-none tracking-[0.04em]">
                    Event Timeline
                  </h3>
                </div>
                <button
                  type="button"
                  onClick={() => navigate('/audit')}
                  className="industrial-action"
                >
                  Inspect Audit Trail
                  <ArrowRightIcon className="h-3.5 w-3.5" />
                </button>
              </div>
              {eventsError ? (
                <div className="mt-6 border-t border-state-danger/50 pt-6">
                  <p className="text-sm text-state-danger">Unable to load recent events</p>
                  <p className="mt-1 text-sm text-text-secondary">{error}</p>
                </div>
              ) : model.events.length === 0 ? (
                <div className="mt-6 border-t border-border-muted pt-6">
                  <p className="text-sm text-text-primary">No recent events</p>
                  <p className="mt-1 text-sm text-text-secondary">
                    No cluster activity has been recorded in this session yet.
                  </p>
                </div>
              ) : (
                <ul className="mt-6 divide-y divide-border-muted">
                  {model.events.map((evt) => (
                    <li key={evt.id}>
                      <div className="industrial-row">
                        <div className="flex flex-wrap items-start justify-between gap-4">
                          <div className="min-w-0">
                            <div className="flex items-center gap-2">
                              {toneIcon(evt.tone)}
                              <p className={cn('truncate text-sm', toneClass(evt.tone))}>
                                {evt.title}
                              </p>
                            </div>
                            <p className="mt-1 text-xs text-text-secondary">
                              <span className="font-mono">{evt.source}</span>
                              {evt.metadata ? (
                                <span className="font-mono text-text-tertiary">
                                  {' '}
                                  | {evt.metadata}
                                </span>
                              ) : null}
                            </p>
                          </div>
                          <span className="industrial-data text-xs text-text-secondary">
                            {clock(evt.timestamp)}
                          </span>
                        </div>
                      </div>
                    </li>
                  ))}
                </ul>
              )}
            </section>
          </div>

          <aside className="space-y-16 xl:col-span-4">
            <section>
              <p className="industrial-label">Services Requiring Attention</p>
              {model.attention.length === 0 ? (
                <div className="mt-6 border-t border-border-muted pt-6">
                  <p className="text-sm text-text-primary">No services require immediate action</p>
                  <p className="mt-1 text-sm text-text-secondary">
                    Replica targets are currently met and no elevated failure pressure was detected.
                  </p>
                </div>
              ) : (
                <ul className="mt-6 divide-y divide-border-muted">
                  {model.attention.map((service) => (
                    <li key={service.id}>
                      <button
                        type="button"
                        onClick={() => navigate('/services')}
                        className="industrial-row w-full text-left focus-visible:outline-none"
                      >
                        <div className="grid grid-cols-1 gap-3 sm:grid-cols-[1.2fr_0.6fr_0.8fr_1fr]">
                          <p className="truncate text-sm text-text-primary">{service.name}</p>
                          <p className="industrial-data text-xs text-text-secondary">
                            {service.replicas}
                          </p>
                          <p className="text-xs text-state-danger">{service.state}</p>
                          <p className="text-xs text-text-secondary">{service.issue}</p>
                        </div>
                      </button>
                    </li>
                  ))}
                </ul>
              )}
            </section>

            <section>
              <p className="industrial-label">Quick Actions</p>
              <ul className="mt-6 divide-y divide-border-muted">
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
                  },
                  {
                    id: 'q3',
                    label: 'Review Services',
                    detail: 'Inspect replica drift',
                    onClick: () => navigate('/services'),
                  },
                  {
                    id: 'q4',
                    label: 'Open Incidents',
                    detail: 'Triage active incidents',
                    onClick: () => navigate('/incidents'),
                  },
                  {
                    id: 'q5',
                    label: 'Audit Trail',
                    detail: 'Inspect operator writes',
                    onClick: () => navigate('/audit'),
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

            <section>
              <p className="industrial-label">Diagnostics</p>
              <dl className="mt-6 space-y-3 border-t border-border-muted pt-5">
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Status</dt>
                  <dd
                    className={cn(
                      'industrial-data text-sm',
                      model.diagnostics.status === 'Degraded'
                        ? 'text-state-danger'
                        : model.diagnostics.status === 'Offline'
                          ? 'text-text-secondary'
                          : 'text-text-primary',
                    )}
                  >
                    {model.diagnostics.status}
                  </dd>
                </div>
                <div className="flex items-center justify-between gap-4">
                  <dt className="industrial-label">Last Run</dt>
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
              </dl>
            </section>

            <section>
              <p className="industrial-label">Cluster Posture</p>
              <dl className="mt-6 space-y-3 border-t border-border-muted pt-5">
                {model.posture.map((item) => (
                  <div key={item.id} className="flex items-center justify-between gap-4">
                    <dt className="industrial-label">{item.label}</dt>
                    <dd className={cn('industrial-data text-sm', toneClass(item.tone))}>
                      {item.value}
                    </dd>
                  </div>
                ))}
              </dl>
            </section>

            <section>
              <p className="industrial-label">Guidance</p>
              <div className="mt-6 border-t border-border-muted pt-5">
                <p className={cn('text-sm', toneClass(model.healthTone))}>{model.guidance.title}</p>
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
