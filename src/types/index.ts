import type { components } from './controlplane.generated'

type ContractSchema<Name extends keyof components['schemas']> = components['schemas'][Name]

// Control-plane types are generated from docs/openapi.yaml.
// Keep manual definitions only for slices that are not modeled in the contract yet.

export type Role = ContractSchema<'Role'>
export type AppMode = ContractSchema<'AppMode'>
export type Severity = ContractSchema<'Severity'>
export type FreshnessState = ContractSchema<'FreshnessState'>
export type ClusterConnectionMode = ContractSchema<'ClusterConnectionMode'>
export type ClusterHealth = ContractSchema<'ClusterHealth'>
export type Cluster = ContractSchema<'Cluster'>
export type ClusterCreateRequest = ContractSchema<'ClusterCreateRequest'>
export type ClusterUpdateRequest = ContractSchema<'ClusterUpdateRequest'>
export type AuthIdentity = ContractSchema<'AuthIdentity'>
export type SwarmInfo = ContractSchema<'SwarmInfo'>
export type RiskAssessment = ContractSchema<'RiskAssessment'>

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
export type Finding = ContractSchema<'Finding'>

// ── Incidents ─────────────────────────────────────────────────────────────────
export type IncidentStatus = ContractSchema<'IncidentStatus'>
export type RunbookStep = ContractSchema<'RunbookStep'>
export type TimelineEntry = ContractSchema<'TimelineEntry'>
export type Incident = ContractSchema<'Incident'>
export type IncidentCreateRequest = ContractSchema<'IncidentCreateRequest'>
export type IncidentUpdateRequest = ContractSchema<'IncidentUpdateRequest'>

// ── Audit ─────────────────────────────────────────────────────────────────────
export type AuditEntry = ContractSchema<'AuditEntry'>

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

export type OpsMetrics = ContractSchema<'OpsMetrics'>

export type InsightHypothesis = ContractSchema<'InsightHypothesis'>
export type InsightAction = ContractSchema<'InsightAction'>
export type OpsInsights = ContractSchema<'OpsInsights'>
export type ActionStatus = ContractSchema<'ActionStatus'>
export type ActionOutcome = ContractSchema<'ActionOutcome'>
export type ActionRun = ContractSchema<'ActionRun'>
export type ActionExecuteRequest = ContractSchema<'ActionExecuteRequest'>
export type ApprovalStatus = ContractSchema<'ApprovalStatus'>
export type ApprovalRequest = ContractSchema<'ApprovalRequest'>
export type AssistantCitation = ContractSchema<'AssistantCitation'>
export type AssistantActionProposal = ContractSchema<'AssistantActionProposal'>
export type AssistantMessage = ContractSchema<'AssistantMessage'>
export type AssistantSession = ContractSchema<'AssistantSession'>
export type AssistantSessionCreateRequest = ContractSchema<'AssistantSessionCreateRequest'>
export type AssistantChatRequest = ContractSchema<'AssistantChatRequest'>
