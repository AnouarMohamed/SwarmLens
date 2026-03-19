// Mirrors backend/internal/model/types.go exactly.
// Keep in sync manually — field names and shapes must match.

export type Role = 'viewer' | 'operator' | 'admin'
export type AppMode = 'dev' | 'demo' | 'prod'
export type Severity = 'critical' | 'high' | 'medium' | 'low' | 'info'
export type FreshnessState = 'live' | 'stale' | 'disconnected'

// ── Swarm ─────────────────────────────────────────────────────────────────────
export interface SwarmInfo {
  clusterID: string
  createdAt: string
  updatedAt: string
  managers: number
  workers: number
  quorumHealthy: boolean
  raftState: 'healthy' | 'degraded' | 'single' | 'unknown'
  mode: AppMode
  writeEnabled: boolean
  freshness: FreshnessState
  lastSyncAt: string
  syncError?: string
  risk: RiskAssessment
}

export interface RiskAssessment {
  score: number
  confidence: number
  factors: string[]
  source: string
  updatedAt: string
}

// ── Nodes ─────────────────────────────────────────────────────────────────────
export interface Node {
  id: string
  hostname: string
  role: 'manager' | 'worker'
  availability: 'active' | 'pause' | 'drain'
  state: 'ready' | 'down' | 'disconnected' | 'unknown'
  labels: Record<string, string>
  cpuTotal: number
  cpuReserved: number
  memTotal: number
  memReserved: number
  runningTasks: number
  managerStatus?: { leader: boolean; reachability: 'reachable' | 'unreachable' }
  engineVersion: string
  addr: string
}

// ── Stacks ────────────────────────────────────────────────────────────────────
export interface Stack {
  name: string
  serviceCount: number
  runningServices: number
  totalReplicas: number
  runningReplicas: number
  healthScore: number
}

// ── Services ──────────────────────────────────────────────────────────────────
export type UpdateState =
  | 'updating' | 'paused' | 'completed'
  | 'rollback_started' | 'rollback_paused' | 'rollback_completed' | ''

export interface PublishedPort {
  publishedPort: number
  targetPort: number
  protocol: string
}

export interface Service {
  id: string
  name: string
  stack: string
  image: string
  mode: 'replicated' | 'global'
  desiredReplicas: number
  runningTasks: number
  failedTasks: number
  updateState: UpdateState
  updateParallelism: number
  updateDelay: string
  updateFailureAction: string
  rollbackParallelism: number
  rollbackDelay: string
  constraints: string[]
  preferences: string[]
  publishedPorts: PublishedPort[]
  secretRefs: string[]
  configRefs: string[]
  networkRefs: string[]
  createdAt: string
  updatedAt: string
}

// ── Tasks ─────────────────────────────────────────────────────────────────────
export type TaskState =
  | 'new' | 'pending' | 'assigned' | 'accepted' | 'preparing'
  | 'ready' | 'starting' | 'running' | 'complete' | 'shutdown'
  | 'failed' | 'rejected' | 'remove' | 'orphaned'

export interface Task {
  id: string
  serviceID: string
  serviceName: string
  nodeID: string
  nodeHostname: string
  desiredState: TaskState
  currentState: TaskState
  exitCode: number
  error: string
  image: string
  restartCount: number
  createdAt: string
  updatedAt: string
}

// ── Networks ──────────────────────────────────────────────────────────────────
export interface Network {
  id: string
  name: string
  driver: string
  scope: 'swarm' | 'local' | 'global'
  subnet: string
  attachable: boolean
  ingress: boolean
  serviceCount: number
}

// ── Volumes ───────────────────────────────────────────────────────────────────
export interface Volume {
  name: string
  driver: string
  scope: string
  mountpoint: string
  labels: Record<string, string>
}

// ── Secrets & Configs ─────────────────────────────────────────────────────────
export interface Secret {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  serviceRefs: string[]
}

export interface Config {
  id: string
  name: string
  createdAt: string
  updatedAt: string
  serviceRefs: string[]
}

// ── Events ────────────────────────────────────────────────────────────────────
export interface SwarmEvent {
  type: string
  action: string
  actor: string
  message: string
  timestamp: string
}

// ── Diagnostics ───────────────────────────────────────────────────────────────
export interface Finding {
  id: string
  severity: Severity
  resource: string
  scope: string
  message: string
  evidence: string[]
  recommendation: string
  source: string
  detectedAt: string
}

// ── Incidents ─────────────────────────────────────────────────────────────────
export type IncidentStatus = 'open' | 'investigating' | 'mitigating' | 'resolved'

export interface RunbookStep {
  id: string
  title: string
  description: string
  status: 'pending' | 'in_progress' | 'done' | 'skipped'
  completedBy?: string
  completedAt?: string
}

export interface TimelineEntry {
  id: string
  actor: string
  action: string
  note: string
  timestamp: string
}

export interface Incident {
  id: string
  title: string
  description: string
  severity: Severity
  status: IncidentStatus
  createdBy: string
  createdAt: string
  updatedAt: string
  resolvedAt?: string
  affectedServices: string[]
  diagnosticRefs: string[]
  runbookSteps: RunbookStep[]
  timeline: TimelineEntry[]
}

// ── Audit ─────────────────────────────────────────────────────────────────────
export interface AuditEntry {
  id: string
  actor: string
  role: Role
  action: string
  resource: string
  resourceID: string
  beforeSpec?: unknown
  afterSpec?: unknown
  result: 'success' | 'failed' | 'rejected' | 'pending_approval'
  reason?: string
  timestamp: string
}

// ── API envelopes ─────────────────────────────────────────────────────────────
export interface ListResponse<T> {
  data: T[]
  meta: { total: number }
}

export interface ItemResponse<T> {
  data: T
}

export interface APIError {
  error: string
  code: string
}

export interface OpsMetricPoint {
  timestamp: string
  healthyRatio: number
  managersOnline: number
  workersOnline: number
  runningTasks: number
  failedTasks: number
  restartCount: number
  critical: number
  warning: number
  riskScore: number
}

export interface ServiceRisk {
  service: string
  score: number
  reasons: string[]
  actionability: 'immediate' | 'soon' | 'monitor' | string
}

export interface OpsMetrics {
  freshness: FreshnessState
  lastUpdated: string
  series: OpsMetricPoint[]
  serviceRisk: ServiceRisk[]
}

export interface InsightHypothesis {
  title: string
  why: string
  confidence: number
}

export interface InsightAction {
  title: string
  description: string
  endpointHint: string
  priority: number
  actionability: 'immediate' | 'soon' | 'monitor' | string
}

export interface OpsInsights {
  summary: string
  risk: RiskAssessment
  freshness: FreshnessState
  hypotheses: InsightHypothesis[]
  actions: InsightAction[]
  generatedAt: string
  provider: string
  sourceStrategy: string
}

export type ActionStatus = 'success' | 'failed' | 'dry_run' | 'blocked'

export interface ActionOutcome {
  action: string
  resource: string
  resourceID: string
  status: ActionStatus
  mode: string
  executed: boolean
  message: string
  blockedReason?: string
  impact?: string
  plan?: string[]
  auditID?: string
  timestamp: string
}
