// Package model defines the canonical data types for SwarmLens.
// All internal packages use these types. HTTP handlers serialise them to JSON.
// The frontend src/types/index.ts mirrors these shapes exactly.
package model

import "time"

// ── Roles ─────────────────────────────────────────────────────────────────────

type Role string

const (
	RoleViewer   Role = "viewer"
	RoleOperator Role = "operator"
	RoleAdmin    Role = "admin"
)

// ── Principal (authenticated actor) ──────────────────────────────────────────

type Principal struct {
	Username string
	Role     Role
}

// ── Swarm cluster ─────────────────────────────────────────────────────────────

type SwarmInfo struct {
	ClusterID     string         `json:"clusterID"`
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	Managers      int            `json:"managers"`
	Workers       int            `json:"workers"`
	QuorumHealthy bool           `json:"quorumHealthy"`
	RaftState     string         `json:"raftState"`
	Mode          string         `json:"mode"`
	WriteEnabled  bool           `json:"writeEnabled"`
	Freshness     FreshnessState `json:"freshness"`
	LastSyncAt    time.Time      `json:"lastSyncAt"`
	SyncError     string         `json:"syncError,omitempty"`
	Risk          RiskAssessment `json:"risk"`
}

// ── Nodes ─────────────────────────────────────────────────────────────────────

type Node struct {
	ID            string            `json:"id"`
	Hostname      string            `json:"hostname"`
	Role          string            `json:"role"`
	Availability  string            `json:"availability"`
	State         string            `json:"state"`
	Labels        map[string]string `json:"labels"`
	CPUTotal      int64             `json:"cpuTotal"`
	CPUReserved   int64             `json:"cpuReserved"`
	MemTotal      int64             `json:"memTotal"`
	MemReserved   int64             `json:"memReserved"`
	RunningTasks  int               `json:"runningTasks"`
	ManagerStatus *ManagerStatus    `json:"managerStatus,omitempty"`
	EngineVersion string            `json:"engineVersion"`
	Addr          string            `json:"addr"`
}

type ManagerStatus struct {
	Leader       bool   `json:"leader"`
	Reachability string `json:"reachability"`
}

// ── Stacks ────────────────────────────────────────────────────────────────────

type Stack struct {
	Name            string `json:"name"`
	ServiceCount    int    `json:"serviceCount"`
	RunningServices int    `json:"runningServices"`
	TotalReplicas   int    `json:"totalReplicas"`
	RunningReplicas int    `json:"runningReplicas"`
	HealthScore     int    `json:"healthScore"`
}

// ── Services ──────────────────────────────────────────────────────────────────

type Service struct {
	ID                  string          `json:"id"`
	Name                string          `json:"name"`
	Stack               string          `json:"stack"`
	Image               string          `json:"image"`
	Mode                string          `json:"mode"`
	DesiredReplicas     int             `json:"desiredReplicas"`
	RunningTasks        int             `json:"runningTasks"`
	FailedTasks         int             `json:"failedTasks"`
	UpdateState         string          `json:"updateState"`
	UpdateParallelism   uint64          `json:"updateParallelism"`
	UpdateDelay         string          `json:"updateDelay"`
	UpdateFailureAction string          `json:"updateFailureAction"`
	RollbackParallelism uint64          `json:"rollbackParallelism"`
	RollbackDelay       string          `json:"rollbackDelay"`
	Constraints         []string        `json:"constraints"`
	Preferences         []string        `json:"preferences"`
	PublishedPorts      []PublishedPort `json:"publishedPorts"`
	SecretRefs          []string        `json:"secretRefs"`
	ConfigRefs          []string        `json:"configRefs"`
	NetworkRefs         []string        `json:"networkRefs"`
	CreatedAt           time.Time       `json:"createdAt"`
	UpdatedAt           time.Time       `json:"updatedAt"`
}

type PublishedPort struct {
	PublishedPort uint32 `json:"publishedPort"`
	TargetPort    uint32 `json:"targetPort"`
	Protocol      string `json:"protocol"`
}

// ── Tasks ─────────────────────────────────────────────────────────────────────

type Task struct {
	ID           string    `json:"id"`
	ServiceID    string    `json:"serviceID"`
	ServiceName  string    `json:"serviceName"`
	NodeID       string    `json:"nodeID"`
	NodeHostname string    `json:"nodeHostname"`
	DesiredState string    `json:"desiredState"`
	CurrentState string    `json:"currentState"`
	ExitCode     int       `json:"exitCode"`
	Error        string    `json:"error"`
	Image        string    `json:"image"`
	RestartCount int       `json:"restartCount"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// ── Networks ──────────────────────────────────────────────────────────────────

type Network struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Driver       string `json:"driver"`
	Scope        string `json:"scope"`
	Subnet       string `json:"subnet"`
	Attachable   bool   `json:"attachable"`
	Ingress      bool   `json:"ingress"`
	ServiceCount int    `json:"serviceCount"`
}

// ── Volumes ───────────────────────────────────────────────────────────────────

type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
}

// ── Secrets & Configs ─────────────────────────────────────────────────────────

type Secret struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ServiceRefs []string  `json:"serviceRefs"`
}

type Config struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	ServiceRefs []string  `json:"serviceRefs"`
}

// ── Events ────────────────────────────────────────────────────────────────────

type SwarmEvent struct {
	Type      string    `json:"type"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// ── Diagnostics ───────────────────────────────────────────────────────────────

type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

type Finding struct {
	ID             string    `json:"id"`
	Severity       Severity  `json:"severity"`
	Resource       string    `json:"resource"`
	Scope          string    `json:"scope"`
	Message        string    `json:"message"`
	Evidence       []string  `json:"evidence"`
	Recommendation string    `json:"recommendation"`
	Source         string    `json:"source"`
	DetectedAt     time.Time `json:"detectedAt"`
}

// ── Incidents ─────────────────────────────────────────────────────────────────

type Incident struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	Description      string          `json:"description"`
	Severity         string          `json:"severity"`
	Status           string          `json:"status"`
	CreatedBy        string          `json:"createdBy"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
	ResolvedAt       *time.Time      `json:"resolvedAt,omitempty"`
	AffectedServices []string        `json:"affectedServices"`
	DiagnosticRefs   []string        `json:"diagnosticRefs"`
	RunbookSteps     []RunbookStep   `json:"runbookSteps"`
	Timeline         []TimelineEntry `json:"timeline"`
}

type RunbookStep struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CompletedBy string     `json:"completedBy,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

type TimelineEntry struct {
	ID        string    `json:"id"`
	Actor     string    `json:"actor"`
	Action    string    `json:"action"`
	Note      string    `json:"note"`
	Timestamp time.Time `json:"timestamp"`
}

// ── Audit ─────────────────────────────────────────────────────────────────────

type AuditEntry struct {
	ID         string      `json:"id"`
	Actor      string      `json:"actor"`
	Role       string      `json:"role"`
	Action     string      `json:"action"`
	Resource   string      `json:"resource"`
	ResourceID string      `json:"resourceID"`
	BeforeSpec interface{} `json:"beforeSpec,omitempty"`
	AfterSpec  interface{} `json:"afterSpec,omitempty"`
	Result     string      `json:"result"`
	Reason     string      `json:"reason,omitempty"`
	Timestamp  time.Time   `json:"timestamp"`
}

// ── Snapshot (used by intelligence engine) ───────────────────────────────────

type Snapshot struct {
	Nodes    []Node
	Services []Service
	Tasks    []Task
	Networks []Network
	Volumes  []Volume
	Secrets  []Secret
	Configs  []Config
	Managers int
	Workers  int
}

// —— Operational Telemetry / AI ————————————————————————————————————————————————————————

type FreshnessState string

const (
	FreshnessLive         FreshnessState = "live"
	FreshnessStale        FreshnessState = "stale"
	FreshnessDisconnected FreshnessState = "disconnected"
)

type RiskAssessment struct {
	Score      float64   `json:"score"`
	Confidence float64   `json:"confidence"`
	Factors    []string  `json:"factors"`
	Source     string    `json:"source"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

type OpsMetricPoint struct {
	Timestamp      time.Time `json:"timestamp"`
	HealthyRatio   float64   `json:"healthyRatio"`
	ManagersOnline int       `json:"managersOnline"`
	WorkersOnline  int       `json:"workersOnline"`
	RunningTasks   int       `json:"runningTasks"`
	FailedTasks    int       `json:"failedTasks"`
	RestartCount   int       `json:"restartCount"`
	Critical       int       `json:"critical"`
	Warning        int       `json:"warning"`
	RiskScore      float64   `json:"riskScore"`
}

type ServiceRisk struct {
	Service       string   `json:"service"`
	Score         float64  `json:"score"`
	Reasons       []string `json:"reasons"`
	Actionability string   `json:"actionability"`
}

type OpsMetrics struct {
	Freshness   FreshnessState   `json:"freshness"`
	LastUpdated time.Time        `json:"lastUpdated"`
	Series      []OpsMetricPoint `json:"series"`
	ServiceRisk []ServiceRisk    `json:"serviceRisk"`
}

type InsightAction struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	EndpointHint  string `json:"endpointHint"`
	Priority      int    `json:"priority"`
	Actionability string `json:"actionability"`
}

type InsightHypothesis struct {
	Title      string  `json:"title"`
	Why        string  `json:"why"`
	Confidence float64 `json:"confidence"`
}

type OpsInsights struct {
	Summary        string              `json:"summary"`
	Risk           RiskAssessment      `json:"risk"`
	Freshness      FreshnessState      `json:"freshness"`
	Hypotheses     []InsightHypothesis `json:"hypotheses"`
	Actions        []InsightAction     `json:"actions"`
	GeneratedAt    time.Time           `json:"generatedAt"`
	Provider       string              `json:"provider"`
	SourceStrategy string              `json:"sourceStrategy"`
}

// —— Write Actions ————————————————————————————————————————————————————————————————

type ActionStatus string

const (
	ActionStatusSuccess ActionStatus = "success"
	ActionStatusFailed  ActionStatus = "failed"
	ActionStatusDryRun  ActionStatus = "dry_run"
	ActionStatusBlocked ActionStatus = "blocked"
)

type ActionOutcome struct {
	Action        string       `json:"action"`
	Resource      string       `json:"resource"`
	ResourceID    string       `json:"resourceID"`
	Status        ActionStatus `json:"status"`
	Mode          string       `json:"mode"`
	Executed      bool         `json:"executed"`
	Message       string       `json:"message"`
	BlockedReason string       `json:"blockedReason,omitempty"`
	Impact        string       `json:"impact,omitempty"`
	Plan          []string     `json:"plan,omitempty"`
	AuditID       string       `json:"auditID,omitempty"`
	Timestamp     time.Time    `json:"timestamp"`
}
