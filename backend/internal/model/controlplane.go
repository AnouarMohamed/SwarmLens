package model

import "time"

type ClusterConnectionMode string

const (
	ClusterConnectionDirect ClusterConnectionMode = "direct"
	ClusterConnectionDemo   ClusterConnectionMode = "demo"
)

type Cluster struct {
	ID             string                `json:"id"`
	Name           string                `json:"name"`
	DockerHost     string                `json:"dockerHost"`
	ConnectionMode ClusterConnectionMode `json:"connectionMode"`
	TLSEnabled     bool                  `json:"tlsEnabled"`
	CertRef        string                `json:"certRef"`
	Enabled        bool                  `json:"enabled"`
	Default        bool                  `json:"default"`
	CreatedAt      time.Time             `json:"createdAt"`
	UpdatedAt      time.Time             `json:"updatedAt"`
	Health         ClusterHealth         `json:"health"`
}

type ClusterHealth struct {
	Freshness     FreshnessState `json:"freshness"`
	LastSyncAt    time.Time      `json:"lastSyncAt,omitempty"`
	LastSyncError string         `json:"lastSyncError,omitempty"`
	Managers      int            `json:"managers"`
	Workers       int            `json:"workers"`
	Reachable     bool           `json:"reachable"`
}

type AuthSession struct {
	ID         string    `json:"id"`
	Username   string    `json:"username"`
	Role       Role      `json:"role"`
	Provider   string    `json:"provider"`
	Groups     []string  `json:"groups"`
	CSRFToken  string    `json:"csrfToken,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
	ExpiresAt  time.Time `json:"expiresAt"`
	LastSeenAt time.Time `json:"lastSeenAt"`
}

type AuthIdentity struct {
	Authenticated bool       `json:"authenticated"`
	Username      string     `json:"username,omitempty"`
	Role          Role       `json:"role,omitempty"`
	Provider      string     `json:"provider,omitempty"`
	Groups        []string   `json:"groups,omitempty"`
	CSRFToken     string     `json:"csrfToken,omitempty"`
	ExpiresAt     *time.Time `json:"expiresAt,omitempty"`
}

type ApprovalStatus string

const (
	ApprovalStatusPending  ApprovalStatus = "pending"
	ApprovalStatusApproved ApprovalStatus = "approved"
	ApprovalStatusRejected ApprovalStatus = "rejected"
)

type ApprovalRequest struct {
	ID               string         `json:"id"`
	ClusterID        string         `json:"clusterID"`
	ActionRunID      string         `json:"actionRunID"`
	Action           string         `json:"action"`
	Resource         string         `json:"resource"`
	ResourceID       string         `json:"resourceID"`
	RequestedBy      string         `json:"requestedBy"`
	RequestedRole    Role           `json:"requestedRole"`
	Reason           string         `json:"reason"`
	Status           ApprovalStatus `json:"status"`
	ResolutionReason string         `json:"resolutionReason,omitempty"`
	ResolvedBy       string         `json:"resolvedBy,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	ResolvedAt       *time.Time     `json:"resolvedAt,omitempty"`
}

type ActionRun struct {
	ID               string         `json:"id"`
	ClusterID        string         `json:"clusterID"`
	Action           string         `json:"action"`
	Resource         string         `json:"resource"`
	ResourceID       string         `json:"resourceID"`
	RequestedBy      string         `json:"requestedBy"`
	RequestedRole    Role           `json:"requestedRole"`
	Reason           string         `json:"reason"`
	Status           ActionStatus   `json:"status"`
	Mode             string         `json:"mode"`
	Executed         bool           `json:"executed"`
	ApprovalRequired bool           `json:"approvalRequired"`
	ApprovalID       string         `json:"approvalID,omitempty"`
	AuditID          string         `json:"auditID,omitempty"`
	Message          string         `json:"message,omitempty"`
	BlockedReason    string         `json:"blockedReason,omitempty"`
	Impact           string         `json:"impact,omitempty"`
	Plan             []string       `json:"plan,omitempty"`
	Params           map[string]any `json:"params,omitempty"`
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	CompletedAt      *time.Time     `json:"completedAt,omitempty"`
}

type AssistantCitation struct {
	ID      string `json:"id"`
	Kind    string `json:"kind"`
	Title   string `json:"title"`
	Locator string `json:"locator"`
	Snippet string `json:"snippet,omitempty"`
}

type AssistantActionProposal struct {
	Title            string         `json:"title"`
	Action           string         `json:"action"`
	Resource         string         `json:"resource,omitempty"`
	ResourceID       string         `json:"resourceID,omitempty"`
	Reason           string         `json:"reason"`
	Params           map[string]any `json:"params,omitempty"`
	RequiresApproval bool           `json:"requiresApproval"`
}

type AssistantMessage struct {
	ID              string                    `json:"id"`
	SessionID       string                    `json:"sessionID"`
	Role            string                    `json:"role"`
	Content         string                    `json:"content"`
	Citations       []AssistantCitation       `json:"citations,omitempty"`
	ActionProposals []AssistantActionProposal `json:"actionProposals,omitempty"`
	CreatedAt       time.Time                 `json:"createdAt"`
}

type AssistantSession struct {
	ID          string             `json:"id"`
	ClusterID   string             `json:"clusterID"`
	IncidentID  string             `json:"incidentID,omitempty"`
	Title       string             `json:"title"`
	CreatedBy   string             `json:"createdBy"`
	LastSummary string             `json:"lastSummary,omitempty"`
	CreatedAt   time.Time          `json:"createdAt"`
	UpdatedAt   time.Time          `json:"updatedAt"`
	Messages    []AssistantMessage `json:"messages,omitempty"`
}
