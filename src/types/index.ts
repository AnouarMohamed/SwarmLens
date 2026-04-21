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
export type ManagerStatus = ContractSchema<'ManagerStatus'>
export type NodeRole = ContractSchema<'NodeRole'>
export type NodeAvailability = ContractSchema<'NodeAvailability'>
export type NodeState = ContractSchema<'NodeState'>
export type Reachability = ContractSchema<'Reachability'>
export type Node = ContractSchema<'Node'>

// ── Stacks ────────────────────────────────────────────────────────────────────
export type Stack = ContractSchema<'Stack'>

// ── Services ──────────────────────────────────────────────────────────────────
export type ServiceMode = ContractSchema<'ServiceMode'>
export type UpdateState = ContractSchema<'UpdateState'>
export type PublishedPort = ContractSchema<'PublishedPort'>
export type Service = ContractSchema<'Service'>

// ── Tasks ─────────────────────────────────────────────────────────────────────
export type TaskState = ContractSchema<'TaskState'>
export type Task = ContractSchema<'Task'>

// ── Networks ──────────────────────────────────────────────────────────────────
export type NetworkScope = ContractSchema<'NetworkScope'>
export type Network = ContractSchema<'Network'>

// ── Volumes ───────────────────────────────────────────────────────────────────
export type Volume = ContractSchema<'Volume'>

// ── Secrets & Configs ─────────────────────────────────────────────────────────
export type Secret = ContractSchema<'Secret'>
export type Config = ContractSchema<'Config'>

// ── Events ────────────────────────────────────────────────────────────────────
export type SwarmEvent = ContractSchema<'SwarmEvent'>

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

export type OpsMetricPoint = ContractSchema<'OpsMetricPoint'>
export type ServiceRisk = ContractSchema<'ServiceRisk'>
export type OpsMetrics = ContractSchema<'OpsMetrics'>

export type InsightHypothesis = ContractSchema<'InsightHypothesis'>
export type InsightAction = ContractSchema<'InsightAction'>
export type OpsInsights = ContractSchema<'OpsInsights'>
export type ActionStatus = ContractSchema<'ActionStatus'>
export type ActionOutcome = ContractSchema<'ActionOutcome'>
export type ActionRun = ContractSchema<'ActionRun'>
export type ActionReasonRequest = ContractSchema<'ActionReasonRequest'>
export type ActionExecuteRequest = ContractSchema<'ActionExecuteRequest'>
export type StackDeployRequest = ContractSchema<'StackDeployRequest'>
export type ServiceScaleRequest = ContractSchema<'ServiceScaleRequest'>
export type ServiceUpdateRequest = ContractSchema<'ServiceUpdateRequest'>
export type ApprovalStatus = ContractSchema<'ApprovalStatus'>
export type ApprovalRequest = ContractSchema<'ApprovalRequest'>
export type AssistantCitation = ContractSchema<'AssistantCitation'>
export type AssistantActionProposal = ContractSchema<'AssistantActionProposal'>
export type AssistantMessage = ContractSchema<'AssistantMessage'>
export type AssistantSession = ContractSchema<'AssistantSession'>
export type AssistantSessionCreateRequest = ContractSchema<'AssistantSessionCreateRequest'>
export type AssistantChatRequest = ContractSchema<'AssistantChatRequest'>
