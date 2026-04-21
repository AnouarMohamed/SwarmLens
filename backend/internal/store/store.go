package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

var ErrNotFound = errors.New("not found")

type Store struct {
	pool *pgxpool.Pool

	mu                sync.RWMutex
	clusters          map[string]model.Cluster
	sessions          map[string]model.AuthSession
	incidents         map[string]model.Incident
	auditEntries      []model.AuditEntry
	actionRuns        map[string]model.ActionRun
	approvals         map[string]model.ApprovalRequest
	assistantSessions map[string]model.AssistantSession
}

func New(ctx context.Context, cfg config.Config) (*Store, error) {
	s := &Store{
		clusters:          map[string]model.Cluster{},
		sessions:          map[string]model.AuthSession{},
		incidents:         map[string]model.Incident{},
		actionRuns:        map[string]model.ActionRun{},
		approvals:         map[string]model.ApprovalRequest{},
		assistantSessions: map[string]model.AssistantSession{},
	}
	if cfg.DatabaseURL == "" {
		return s, nil
	}

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	s.pool = pool
	if err := s.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) IsPersistent() bool {
	return s.pool != nil
}

func (s *Store) Ready(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	return s.pool.Ping(ctx)
}

func (s *Store) Migrate(ctx context.Context) error {
	if s.pool == nil {
		return nil
	}
	stmts := []string{
		`create table if not exists clusters (
			id text primary key,
			name text not null unique,
			docker_host text not null,
			connection_mode text not null,
			tls_enabled boolean not null default false,
			cert_ref text not null default '',
			enabled boolean not null default true,
			is_default boolean not null default false,
			created_at timestamptz not null,
			updated_at timestamptz not null
		)`,
		`create table if not exists auth_sessions (
			id text primary key,
			username text not null,
			role text not null,
			provider text not null,
			groups_json jsonb not null default '[]'::jsonb,
			csrf_token text not null,
			created_at timestamptz not null,
			expires_at timestamptz not null,
			last_seen_at timestamptz not null
		)`,
		`create table if not exists incidents (
			id text primary key,
			cluster_id text not null,
			title text not null,
			description text not null default '',
			severity text not null,
			status text not null,
			created_by text not null,
			created_at timestamptz not null,
			updated_at timestamptz not null,
			resolved_at timestamptz null,
			affected_services jsonb not null default '[]'::jsonb,
			diagnostic_refs jsonb not null default '[]'::jsonb,
			runbook_steps jsonb not null default '[]'::jsonb,
			timeline jsonb not null default '[]'::jsonb
		)`,
		`create table if not exists audit_entries (
			id text primary key,
			cluster_id text not null,
			action_run_id text not null default '',
			actor text not null,
			role text not null,
			action text not null,
			resource text not null,
			resource_id text not null,
			before_spec jsonb null,
			after_spec jsonb null,
			result text not null,
			reason text not null default '',
			timestamp timestamptz not null
		)`,
		`create table if not exists action_runs (
			id text primary key,
			cluster_id text not null,
			action text not null,
			resource text not null,
			resource_id text not null,
			requested_by text not null,
			requested_role text not null,
			reason text not null default '',
			status text not null,
			mode text not null,
			executed boolean not null default false,
			approval_required boolean not null default false,
			approval_id text not null default '',
			audit_id text not null default '',
			message text not null default '',
			blocked_reason text not null default '',
			impact text not null default '',
			plan_json jsonb not null default '[]'::jsonb,
			params_json jsonb not null default '{}'::jsonb,
			created_at timestamptz not null,
			updated_at timestamptz not null,
			completed_at timestamptz null
		)`,
		`create table if not exists approval_requests (
			id text primary key,
			cluster_id text not null,
			action_run_id text not null,
			action text not null,
			resource text not null,
			resource_id text not null,
			requested_by text not null,
			requested_role text not null,
			reason text not null,
			status text not null,
			resolution_reason text not null default '',
			resolved_by text not null default '',
			created_at timestamptz not null,
			resolved_at timestamptz null
		)`,
		`create table if not exists assistant_sessions (
			id text primary key,
			cluster_id text not null,
			incident_id text not null default '',
			title text not null,
			created_by text not null,
			last_summary text not null default '',
			created_at timestamptz not null,
			updated_at timestamptz not null
		)`,
		`create table if not exists assistant_messages (
			id text primary key,
			session_id text not null,
			role text not null,
			content text not null,
			citations_json jsonb not null default '[]'::jsonb,
			action_proposals_json jsonb not null default '[]'::jsonb,
			created_at timestamptz not null
		)`,
		`create index if not exists idx_incidents_cluster_updated on incidents(cluster_id, updated_at desc)`,
		`create index if not exists idx_audit_cluster_timestamp on audit_entries(cluster_id, timestamp desc)`,
		`create index if not exists idx_action_runs_cluster_created on action_runs(cluster_id, created_at desc)`,
		`create index if not exists idx_approvals_cluster_status on approval_requests(cluster_id, status, created_at desc)`,
		`create index if not exists idx_assistant_sessions_cluster_updated on assistant_sessions(cluster_id, updated_at desc)`,
		`create index if not exists idx_assistant_messages_session_created on assistant_messages(session_id, created_at asc)`,
	}
	for _, stmt := range stmts {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("migrate postgres: %w", err)
		}
	}
	return nil
}

func (s *Store) SeedDefaultCluster(ctx context.Context, cluster model.Cluster) (model.Cluster, error) {
	cluster.Default = true
	cluster.Enabled = true
	if cluster.ID == "" {
		cluster.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if cluster.CreatedAt.IsZero() {
		cluster.CreatedAt = now
	}
	cluster.UpdatedAt = now

	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		for id, existing := range s.clusters {
			if existing.Default {
				return cloneCluster(existing), nil
			}
			existing.Default = false
			s.clusters[id] = existing
		}
		s.clusters[cluster.ID] = cloneCluster(cluster)
		return cloneCluster(cluster), nil
	}

	var existingID string
	err := s.pool.QueryRow(ctx, `select id from clusters where is_default = true order by created_at asc limit 1`).Scan(&existingID)
	if err == nil && existingID != "" {
		return s.GetCluster(ctx, existingID)
	}
	_, err = s.pool.Exec(ctx, `
		insert into clusters(id, name, docker_host, connection_mode, tls_enabled, cert_ref, enabled, is_default, created_at, updated_at)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		on conflict (name) do update set
			docker_host = excluded.docker_host,
			connection_mode = excluded.connection_mode,
			tls_enabled = excluded.tls_enabled,
			cert_ref = excluded.cert_ref,
			enabled = excluded.enabled,
			is_default = excluded.is_default,
			updated_at = excluded.updated_at
	`, cluster.ID, cluster.Name, cluster.DockerHost, string(cluster.ConnectionMode), cluster.TLSEnabled, cluster.CertRef, cluster.Enabled, true, cluster.CreatedAt, cluster.UpdatedAt)
	if err != nil {
		return model.Cluster{}, fmt.Errorf("seed default cluster: %w", err)
	}
	return s.GetDefaultCluster(ctx)
}

func (s *Store) ListClusters(ctx context.Context) ([]model.Cluster, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]model.Cluster, 0, len(s.clusters))
		for _, cluster := range s.clusters {
			items = append(items, cloneCluster(cluster))
		}
		sort.Slice(items, func(i, j int) bool {
			if items[i].Default != items[j].Default {
				return items[i].Default
			}
			return items[i].Name < items[j].Name
		})
		return items, nil
	}

	rows, err := s.pool.Query(ctx, `
		select id, name, docker_host, connection_mode, tls_enabled, cert_ref, enabled, is_default, created_at, updated_at
		from clusters
		order by is_default desc, name asc
	`)
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	defer rows.Close()

	var items []model.Cluster
	for rows.Next() {
		cluster, err := scanCluster(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, cluster)
	}
	return items, rows.Err()
}

func (s *Store) GetCluster(ctx context.Context, id string) (model.Cluster, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		cluster, ok := s.clusters[id]
		if !ok {
			return model.Cluster{}, ErrNotFound
		}
		return cloneCluster(cluster), nil
	}
	row := s.pool.QueryRow(ctx, `
		select id, name, docker_host, connection_mode, tls_enabled, cert_ref, enabled, is_default, created_at, updated_at
		from clusters where id = $1
	`, id)
	cluster, err := scanCluster(row)
	if err != nil {
		return model.Cluster{}, err
	}
	return cluster, nil
}

func (s *Store) GetDefaultCluster(ctx context.Context) (model.Cluster, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		for _, cluster := range s.clusters {
			if cluster.Default {
				return cloneCluster(cluster), nil
			}
		}
		return model.Cluster{}, ErrNotFound
	}
	row := s.pool.QueryRow(ctx, `
		select id, name, docker_host, connection_mode, tls_enabled, cert_ref, enabled, is_default, created_at, updated_at
		from clusters where is_default = true order by created_at asc limit 1
	`)
	cluster, err := scanCluster(row)
	if err != nil {
		return model.Cluster{}, err
	}
	return cluster, nil
}

func (s *Store) SaveCluster(ctx context.Context, cluster model.Cluster) (model.Cluster, error) {
	now := time.Now().UTC()
	if cluster.ID == "" {
		cluster.ID = uuid.NewString()
	}
	if cluster.CreatedAt.IsZero() {
		cluster.CreatedAt = now
	}
	cluster.UpdatedAt = now

	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		if cluster.Default {
			for id, existing := range s.clusters {
				existing.Default = false
				s.clusters[id] = existing
			}
		}
		if !cluster.Enabled && cluster.Default {
			cluster.Enabled = true
		}
		s.clusters[cluster.ID] = cloneCluster(cluster)
		return cloneCluster(cluster), nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.Cluster{}, err
	}
	defer tx.Rollback(ctx)
	if cluster.Default {
		if _, err := tx.Exec(ctx, `update clusters set is_default = false where id <> $1`, cluster.ID); err != nil {
			return model.Cluster{}, fmt.Errorf("clear default cluster: %w", err)
		}
	}
	_, err = tx.Exec(ctx, `
		insert into clusters(id, name, docker_host, connection_mode, tls_enabled, cert_ref, enabled, is_default, created_at, updated_at)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		on conflict (id) do update set
			name = excluded.name,
			docker_host = excluded.docker_host,
			connection_mode = excluded.connection_mode,
			tls_enabled = excluded.tls_enabled,
			cert_ref = excluded.cert_ref,
			enabled = excluded.enabled,
			is_default = excluded.is_default,
			updated_at = excluded.updated_at
	`, cluster.ID, cluster.Name, cluster.DockerHost, string(cluster.ConnectionMode), cluster.TLSEnabled, cluster.CertRef, cluster.Enabled, cluster.Default, cluster.CreatedAt, cluster.UpdatedAt)
	if err != nil {
		return model.Cluster{}, fmt.Errorf("save cluster: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Cluster{}, err
	}
	return s.GetCluster(ctx, cluster.ID)
}

func (s *Store) CreateSession(ctx context.Context, session model.AuthSession) (model.AuthSession, error) {
	now := time.Now().UTC()
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.LastSeenAt.IsZero() {
		session.LastSeenAt = now
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.sessions[session.ID] = cloneSession(session)
		return cloneSession(session), nil
	}
	groupsJSON := mustJSON(session.Groups)
	_, err := s.pool.Exec(ctx, `
		insert into auth_sessions(id, username, role, provider, groups_json, csrf_token, created_at, expires_at, last_seen_at)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`, session.ID, session.Username, string(session.Role), session.Provider, groupsJSON, session.CSRFToken, session.CreatedAt, session.ExpiresAt, session.LastSeenAt)
	if err != nil {
		return model.AuthSession{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (s *Store) GetSession(ctx context.Context, id string) (model.AuthSession, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		session, ok := s.sessions[id]
		if !ok || time.Now().UTC().After(session.ExpiresAt) {
			return model.AuthSession{}, ErrNotFound
		}
		return cloneSession(session), nil
	}
	var session model.AuthSession
	var groupsRaw []byte
	err := s.pool.QueryRow(ctx, `
		select id, username, role, provider, groups_json, csrf_token, created_at, expires_at, last_seen_at
		from auth_sessions
		where id = $1 and expires_at > now()
	`, id).Scan(&session.ID, &session.Username, &session.Role, &session.Provider, &groupsRaw, &session.CSRFToken, &session.CreatedAt, &session.ExpiresAt, &session.LastSeenAt)
	if err != nil {
		return model.AuthSession{}, ErrNotFound
	}
	if len(groupsRaw) > 0 {
		_ = json.Unmarshal(groupsRaw, &session.Groups)
	}
	return session, nil
}

func (s *Store) TouchSession(ctx context.Context, id string) error {
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		session, ok := s.sessions[id]
		if !ok {
			return ErrNotFound
		}
		session.LastSeenAt = time.Now().UTC()
		s.sessions[id] = session
		return nil
	}
	if _, err := s.pool.Exec(ctx, `update auth_sessions set last_seen_at = now() where id = $1`, id); err != nil {
		return fmt.Errorf("touch session: %w", err)
	}
	return nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.sessions, id)
		return nil
	}
	if _, err := s.pool.Exec(ctx, `delete from auth_sessions where id = $1`, id); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func (s *Store) ListIncidents(ctx context.Context, clusterID string) ([]model.Incident, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]model.Incident, 0, len(s.incidents))
		for _, incident := range s.incidents {
			if clusterID != "" && incident.ClusterID != clusterID {
				continue
			}
			items = append(items, cloneIncident(incident))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
		return items, nil
	}
	rows, err := s.pool.Query(ctx, `
		select id, cluster_id, title, description, severity, status, created_by, created_at, updated_at, resolved_at,
			affected_services, diagnostic_refs, runbook_steps, timeline
		from incidents
		where cluster_id = $1
		order by updated_at desc
	`, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list incidents: %w", err)
	}
	defer rows.Close()
	var items []model.Incident
	for rows.Next() {
		incident, err := scanIncident(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, incident)
	}
	return items, rows.Err()
}

func (s *Store) GetIncident(ctx context.Context, clusterID, id string) (model.Incident, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		incident, ok := s.incidents[id]
		if !ok || (clusterID != "" && incident.ClusterID != clusterID) {
			return model.Incident{}, ErrNotFound
		}
		return cloneIncident(incident), nil
	}
	row := s.pool.QueryRow(ctx, `
		select id, cluster_id, title, description, severity, status, created_by, created_at, updated_at, resolved_at,
			affected_services, diagnostic_refs, runbook_steps, timeline
		from incidents
		where cluster_id = $1 and id = $2
	`, clusterID, id)
	return scanIncident(row)
}

func (s *Store) CreateIncident(ctx context.Context, clusterID, title, description, severity, createdBy string, affectedServices, diagRefs []string) (model.Incident, error) {
	now := time.Now().UTC()
	incident := model.Incident{
		ID:               uuid.NewString(),
		ClusterID:        clusterID,
		Title:            title,
		Description:      description,
		Severity:         severity,
		Status:           "open",
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
		AffectedServices: append([]string(nil), affectedServices...),
		DiagnosticRefs:   append([]string(nil), diagRefs...),
		RunbookSteps:     []model.RunbookStep{},
		Timeline: []model.TimelineEntry{
			{ID: uuid.NewString(), Actor: createdBy, Action: "created", Note: "Incident opened.", Timestamp: now},
		},
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.incidents[incident.ID] = cloneIncident(incident)
		return cloneIncident(incident), nil
	}
	_, err := s.pool.Exec(ctx, `
		insert into incidents(id, cluster_id, title, description, severity, status, created_by, created_at, updated_at, resolved_at,
			affected_services, diagnostic_refs, runbook_steps, timeline)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, incident.ID, incident.ClusterID, incident.Title, incident.Description, incident.Severity, incident.Status, incident.CreatedBy, incident.CreatedAt, incident.UpdatedAt, nil, mustJSON(incident.AffectedServices), mustJSON(incident.DiagnosticRefs), mustJSON(incident.RunbookSteps), mustJSON(incident.Timeline))
	if err != nil {
		return model.Incident{}, fmt.Errorf("create incident: %w", err)
	}
	return incident, nil
}

func (s *Store) UpdateIncidentStatus(ctx context.Context, clusterID, id, status, actor, note string) (model.Incident, error) {
	incident, err := s.GetIncident(ctx, clusterID, id)
	if err != nil {
		return model.Incident{}, err
	}
	now := time.Now().UTC()
	incident.Status = status
	incident.UpdatedAt = now
	if status == "resolved" {
		incident.ResolvedAt = &now
	}
	incident.Timeline = append(incident.Timeline, model.TimelineEntry{
		ID:        uuid.NewString(),
		Actor:     actor,
		Action:    "status_change",
		Note:      note,
		Timestamp: now,
	})
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.incidents[id] = cloneIncident(incident)
		return cloneIncident(incident), nil
	}
	_, err = s.pool.Exec(ctx, `
		update incidents
		set status = $3, updated_at = $4, resolved_at = $5, timeline = $6
		where cluster_id = $1 and id = $2
	`, clusterID, id, incident.Status, incident.UpdatedAt, incident.ResolvedAt, mustJSON(incident.Timeline))
	if err != nil {
		return model.Incident{}, fmt.Errorf("update incident: %w", err)
	}
	return incident, nil
}

func (s *Store) RecordAudit(ctx context.Context, entry model.AuditEntry) (model.AuditEntry, error) {
	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.auditEntries = append(s.auditEntries, cloneAudit(entry))
		return cloneAudit(entry), nil
	}
	_, err := s.pool.Exec(ctx, `
		insert into audit_entries(id, cluster_id, action_run_id, actor, role, action, resource, resource_id, before_spec, after_spec, result, reason, timestamp)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
	`, entry.ID, entry.ClusterID, entry.ActionRunID, entry.Actor, entry.Role, entry.Action, entry.Resource, entry.ResourceID, nilJSON(entry.BeforeSpec), nilJSON(entry.AfterSpec), entry.Result, entry.Reason, entry.Timestamp)
	if err != nil {
		return model.AuditEntry{}, fmt.Errorf("record audit: %w", err)
	}
	return entry, nil
}

func (s *Store) ListAudit(ctx context.Context, clusterID string, limit, offset int) ([]model.AuditEntry, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		var items []model.AuditEntry
		for _, entry := range s.auditEntries {
			if clusterID != "" && entry.ClusterID != clusterID {
				continue
			}
			items = append(items, cloneAudit(entry))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].Timestamp.After(items[j].Timestamp) })
		total := len(items)
		if offset >= total {
			return nil, total, nil
		}
		end := offset + limit
		if end > total {
			end = total
		}
		return items[offset:end], total, nil
	}
	var total int
	if err := s.pool.QueryRow(ctx, `select count(*) from audit_entries where cluster_id = $1`, clusterID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count audit: %w", err)
	}
	rows, err := s.pool.Query(ctx, `
		select id, cluster_id, action_run_id, actor, role, action, resource, resource_id, before_spec, after_spec, result, reason, timestamp
		from audit_entries where cluster_id = $1
		order by timestamp desc
		limit $2 offset $3
	`, clusterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list audit: %w", err)
	}
	defer rows.Close()
	var items []model.AuditEntry
	for rows.Next() {
		entry, err := scanAudit(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, entry)
	}
	return items, total, rows.Err()
}

func (s *Store) CreateActionRun(ctx context.Context, run model.ActionRun) (model.ActionRun, error) {
	now := time.Now().UTC()
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = now
	}
	if run.UpdatedAt.IsZero() {
		run.UpdatedAt = run.CreatedAt
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.actionRuns[run.ID] = cloneActionRun(run)
		return cloneActionRun(run), nil
	}
	_, err := s.pool.Exec(ctx, `
		insert into action_runs(id, cluster_id, action, resource, resource_id, requested_by, requested_role, reason, status, mode, executed, approval_required, approval_id, audit_id, message, blocked_reason, impact, plan_json, params_json, created_at, updated_at, completed_at)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22)
	`, run.ID, run.ClusterID, run.Action, run.Resource, run.ResourceID, run.RequestedBy, string(run.RequestedRole), run.Reason, string(run.Status), run.Mode, run.Executed, run.ApprovalRequired, run.ApprovalID, run.AuditID, run.Message, run.BlockedReason, run.Impact, mustJSON(run.Plan), mustJSON(mapOrEmpty(run.Params)), run.CreatedAt, run.UpdatedAt, run.CompletedAt)
	if err != nil {
		return model.ActionRun{}, fmt.Errorf("create action run: %w", err)
	}
	return run, nil
}

func (s *Store) UpdateActionRun(ctx context.Context, run model.ActionRun) (model.ActionRun, error) {
	run.UpdatedAt = time.Now().UTC()
	if run.Executed || run.Status == model.ActionStatusFailed || run.Status == model.ActionStatusBlocked || run.Status == model.ActionStatusSuccess {
		completed := run.UpdatedAt
		run.CompletedAt = &completed
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.actionRuns[run.ID] = cloneActionRun(run)
		return cloneActionRun(run), nil
	}
	_, err := s.pool.Exec(ctx, `
		update action_runs
		set status = $2, mode = $3, executed = $4, approval_required = $5, approval_id = $6, audit_id = $7,
			message = $8, blocked_reason = $9, impact = $10, plan_json = $11, params_json = $12, updated_at = $13, completed_at = $14
		where id = $1
	`, run.ID, string(run.Status), run.Mode, run.Executed, run.ApprovalRequired, run.ApprovalID, run.AuditID, run.Message, run.BlockedReason, run.Impact, mustJSON(run.Plan), mustJSON(mapOrEmpty(run.Params)), run.UpdatedAt, run.CompletedAt)
	if err != nil {
		return model.ActionRun{}, fmt.Errorf("update action run: %w", err)
	}
	return run, nil
}

func (s *Store) GetActionRun(ctx context.Context, id string) (model.ActionRun, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		run, ok := s.actionRuns[id]
		if !ok {
			return model.ActionRun{}, ErrNotFound
		}
		return cloneActionRun(run), nil
	}
	row := s.pool.QueryRow(ctx, `
		select id, cluster_id, action, resource, resource_id, requested_by, requested_role, reason, status, mode, executed, approval_required, approval_id, audit_id, message, blocked_reason, impact, plan_json, params_json, created_at, updated_at, completed_at
		from action_runs where id = $1
	`, id)
	return scanActionRun(row)
}

func (s *Store) ListActionRuns(ctx context.Context, clusterID string, limit int) ([]model.ActionRun, error) {
	if limit <= 0 {
		limit = 50
	}
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]model.ActionRun, 0, len(s.actionRuns))
		for _, run := range s.actionRuns {
			if clusterID != "" && run.ClusterID != clusterID {
				continue
			}
			items = append(items, cloneActionRun(run))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
		if len(items) > limit {
			items = items[:limit]
		}
		return items, nil
	}
	rows, err := s.pool.Query(ctx, `
		select id, cluster_id, action, resource, resource_id, requested_by, requested_role, reason, status, mode, executed, approval_required, approval_id, audit_id, message, blocked_reason, impact, plan_json, params_json, created_at, updated_at, completed_at
		from action_runs
		where cluster_id = $1
		order by created_at desc
		limit $2
	`, clusterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list action runs: %w", err)
	}
	defer rows.Close()
	var items []model.ActionRun
	for rows.Next() {
		run, err := scanActionRun(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, run)
	}
	return items, rows.Err()
}

func (s *Store) CreateApproval(ctx context.Context, approval model.ApprovalRequest) (model.ApprovalRequest, error) {
	if approval.ID == "" {
		approval.ID = uuid.NewString()
	}
	if approval.CreatedAt.IsZero() {
		approval.CreatedAt = time.Now().UTC()
	}
	if approval.Status == "" {
		approval.Status = model.ApprovalStatusPending
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.approvals[approval.ID] = cloneApproval(approval)
		return cloneApproval(approval), nil
	}
	_, err := s.pool.Exec(ctx, `
		insert into approval_requests(id, cluster_id, action_run_id, action, resource, resource_id, requested_by, requested_role, reason, status, resolution_reason, resolved_by, created_at, resolved_at)
		values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`, approval.ID, approval.ClusterID, approval.ActionRunID, approval.Action, approval.Resource, approval.ResourceID, approval.RequestedBy, string(approval.RequestedRole), approval.Reason, string(approval.Status), approval.ResolutionReason, approval.ResolvedBy, approval.CreatedAt, approval.ResolvedAt)
	if err != nil {
		return model.ApprovalRequest{}, fmt.Errorf("create approval: %w", err)
	}
	return approval, nil
}

func (s *Store) GetApproval(ctx context.Context, id string) (model.ApprovalRequest, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		approval, ok := s.approvals[id]
		if !ok {
			return model.ApprovalRequest{}, ErrNotFound
		}
		return cloneApproval(approval), nil
	}
	row := s.pool.QueryRow(ctx, `
		select id, cluster_id, action_run_id, action, resource, resource_id, requested_by, requested_role, reason, status, resolution_reason, resolved_by, created_at, resolved_at
		from approval_requests where id = $1
	`, id)
	return scanApproval(row)
}

func (s *Store) ListApprovals(ctx context.Context, clusterID string, status model.ApprovalStatus) ([]model.ApprovalRequest, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]model.ApprovalRequest, 0, len(s.approvals))
		for _, approval := range s.approvals {
			if clusterID != "" && approval.ClusterID != clusterID {
				continue
			}
			if status != "" && approval.Status != status {
				continue
			}
			items = append(items, cloneApproval(approval))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
		return items, nil
	}
	query := `
		select id, cluster_id, action_run_id, action, resource, resource_id, requested_by, requested_role, reason, status, resolution_reason, resolved_by, created_at, resolved_at
		from approval_requests
		where cluster_id = $1
	`
	args := []any{clusterID}
	if status != "" {
		query += ` and status = $2`
		args = append(args, string(status))
	}
	query += ` order by created_at desc`
	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list approvals: %w", err)
	}
	defer rows.Close()
	var items []model.ApprovalRequest
	for rows.Next() {
		approval, err := scanApproval(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, approval)
	}
	return items, rows.Err()
}

func (s *Store) ResolveApproval(ctx context.Context, id string, status model.ApprovalStatus, resolvedBy, reason string) (model.ApprovalRequest, error) {
	approval, err := s.GetApproval(ctx, id)
	if err != nil {
		return model.ApprovalRequest{}, err
	}
	now := time.Now().UTC()
	approval.Status = status
	approval.ResolvedBy = resolvedBy
	approval.ResolutionReason = reason
	approval.ResolvedAt = &now
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.approvals[id] = cloneApproval(approval)
		return cloneApproval(approval), nil
	}
	_, err = s.pool.Exec(ctx, `
		update approval_requests
		set status = $2, resolution_reason = $3, resolved_by = $4, resolved_at = $5
		where id = $1
	`, id, string(status), reason, resolvedBy, approval.ResolvedAt)
	if err != nil {
		return model.ApprovalRequest{}, fmt.Errorf("resolve approval: %w", err)
	}
	return approval, nil
}

func (s *Store) ListAssistantSessions(ctx context.Context, clusterID string) ([]model.AssistantSession, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		items := make([]model.AssistantSession, 0, len(s.assistantSessions))
		for _, session := range s.assistantSessions {
			if clusterID != "" && session.ClusterID != clusterID {
				continue
			}
			items = append(items, cloneAssistantSession(session))
		}
		sort.Slice(items, func(i, j int) bool { return items[i].UpdatedAt.After(items[j].UpdatedAt) })
		return items, nil
	}
	rows, err := s.pool.Query(ctx, `
		select id, cluster_id, incident_id, title, created_by, last_summary, created_at, updated_at
		from assistant_sessions
		where cluster_id = $1
		order by updated_at desc
	`, clusterID)
	if err != nil {
		return nil, fmt.Errorf("list assistant sessions: %w", err)
	}
	defer rows.Close()
	var items []model.AssistantSession
	for rows.Next() {
		session, err := scanAssistantSession(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, session)
	}
	return items, rows.Err()
}

func (s *Store) CreateAssistantSession(ctx context.Context, session model.AssistantSession) (model.AssistantSession, error) {
	now := time.Now().UTC()
	if session.ID == "" {
		session.ID = uuid.NewString()
	}
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.UpdatedAt.IsZero() {
		session.UpdatedAt = now
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.assistantSessions[session.ID] = cloneAssistantSession(session)
		return cloneAssistantSession(session), nil
	}
	_, err := s.pool.Exec(ctx, `
		insert into assistant_sessions(id, cluster_id, incident_id, title, created_by, last_summary, created_at, updated_at)
		values($1,$2,$3,$4,$5,$6,$7,$8)
	`, session.ID, session.ClusterID, session.IncidentID, session.Title, session.CreatedBy, session.LastSummary, session.CreatedAt, session.UpdatedAt)
	if err != nil {
		return model.AssistantSession{}, fmt.Errorf("create assistant session: %w", err)
	}
	return session, nil
}

func (s *Store) GetAssistantSession(ctx context.Context, clusterID, id string) (model.AssistantSession, error) {
	if s.pool == nil {
		s.mu.RLock()
		defer s.mu.RUnlock()
		session, ok := s.assistantSessions[id]
		if !ok || (clusterID != "" && session.ClusterID != clusterID) {
			return model.AssistantSession{}, ErrNotFound
		}
		return cloneAssistantSession(session), nil
	}
	row := s.pool.QueryRow(ctx, `
		select id, cluster_id, incident_id, title, created_by, last_summary, created_at, updated_at
		from assistant_sessions
		where cluster_id = $1 and id = $2
	`, clusterID, id)
	session, err := scanAssistantSession(row)
	if err != nil {
		return model.AssistantSession{}, err
	}
	msgRows, err := s.pool.Query(ctx, `
		select id, session_id, role, content, citations_json, action_proposals_json, created_at
		from assistant_messages
		where session_id = $1
		order by created_at asc
	`, id)
	if err != nil {
		return model.AssistantSession{}, fmt.Errorf("list assistant messages: %w", err)
	}
	defer msgRows.Close()
	for msgRows.Next() {
		msg, err := scanAssistantMessage(msgRows)
		if err != nil {
			return model.AssistantSession{}, err
		}
		session.Messages = append(session.Messages, msg)
	}
	return session, msgRows.Err()
}

func (s *Store) AppendAssistantMessage(ctx context.Context, sessionID string, message model.AssistantMessage, newSummary string) (model.AssistantMessage, error) {
	now := time.Now().UTC()
	if message.ID == "" {
		message.ID = uuid.NewString()
	}
	message.SessionID = sessionID
	if message.CreatedAt.IsZero() {
		message.CreatedAt = now
	}
	if s.pool == nil {
		s.mu.Lock()
		defer s.mu.Unlock()
		session, ok := s.assistantSessions[sessionID]
		if !ok {
			return model.AssistantMessage{}, ErrNotFound
		}
		session.Messages = append(session.Messages, cloneAssistantMessage(message))
		session.LastSummary = newSummary
		session.UpdatedAt = now
		s.assistantSessions[sessionID] = session
		return cloneAssistantMessage(message), nil
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return model.AssistantMessage{}, err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, `
		insert into assistant_messages(id, session_id, role, content, citations_json, action_proposals_json, created_at)
		values($1,$2,$3,$4,$5,$6,$7)
	`, message.ID, sessionID, message.Role, message.Content, mustJSON(message.Citations), mustJSON(message.ActionProposals), message.CreatedAt)
	if err != nil {
		return model.AssistantMessage{}, fmt.Errorf("insert assistant message: %w", err)
	}
	_, err = tx.Exec(ctx, `
		update assistant_sessions
		set last_summary = $2, updated_at = $3
		where id = $1
	`, sessionID, newSummary, now)
	if err != nil {
		return model.AssistantMessage{}, fmt.Errorf("update assistant session: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return model.AssistantMessage{}, err
	}
	return message, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanCluster(row scanner) (model.Cluster, error) {
	var cluster model.Cluster
	var mode string
	err := row.Scan(&cluster.ID, &cluster.Name, &cluster.DockerHost, &mode, &cluster.TLSEnabled, &cluster.CertRef, &cluster.Enabled, &cluster.Default, &cluster.CreatedAt, &cluster.UpdatedAt)
	if err != nil {
		return model.Cluster{}, ErrNotFound
	}
	cluster.ConnectionMode = model.ClusterConnectionMode(mode)
	return cluster, nil
}

func scanIncident(row scanner) (model.Incident, error) {
	var incident model.Incident
	var affected, refs, steps, timeline []byte
	err := row.Scan(&incident.ID, &incident.ClusterID, &incident.Title, &incident.Description, &incident.Severity, &incident.Status, &incident.CreatedBy, &incident.CreatedAt, &incident.UpdatedAt, &incident.ResolvedAt, &affected, &refs, &steps, &timeline)
	if err != nil {
		return model.Incident{}, ErrNotFound
	}
	_ = json.Unmarshal(affected, &incident.AffectedServices)
	_ = json.Unmarshal(refs, &incident.DiagnosticRefs)
	_ = json.Unmarshal(steps, &incident.RunbookSteps)
	_ = json.Unmarshal(timeline, &incident.Timeline)
	return incident, nil
}

func scanAudit(row scanner) (model.AuditEntry, error) {
	var entry model.AuditEntry
	var before, after []byte
	err := row.Scan(&entry.ID, &entry.ClusterID, &entry.ActionRunID, &entry.Actor, &entry.Role, &entry.Action, &entry.Resource, &entry.ResourceID, &before, &after, &entry.Result, &entry.Reason, &entry.Timestamp)
	if err != nil {
		return model.AuditEntry{}, ErrNotFound
	}
	if len(before) > 0 {
		_ = json.Unmarshal(before, &entry.BeforeSpec)
	}
	if len(after) > 0 {
		_ = json.Unmarshal(after, &entry.AfterSpec)
	}
	return entry, nil
}

func scanActionRun(row scanner) (model.ActionRun, error) {
	var run model.ActionRun
	var role string
	var planRaw, paramsRaw []byte
	err := row.Scan(&run.ID, &run.ClusterID, &run.Action, &run.Resource, &run.ResourceID, &run.RequestedBy, &role, &run.Reason, &run.Status, &run.Mode, &run.Executed, &run.ApprovalRequired, &run.ApprovalID, &run.AuditID, &run.Message, &run.BlockedReason, &run.Impact, &planRaw, &paramsRaw, &run.CreatedAt, &run.UpdatedAt, &run.CompletedAt)
	if err != nil {
		return model.ActionRun{}, ErrNotFound
	}
	run.RequestedRole = model.Role(role)
	_ = json.Unmarshal(planRaw, &run.Plan)
	_ = json.Unmarshal(paramsRaw, &run.Params)
	if run.Params == nil {
		run.Params = map[string]any{}
	}
	return run, nil
}

func scanApproval(row scanner) (model.ApprovalRequest, error) {
	var approval model.ApprovalRequest
	var role string
	err := row.Scan(&approval.ID, &approval.ClusterID, &approval.ActionRunID, &approval.Action, &approval.Resource, &approval.ResourceID, &approval.RequestedBy, &role, &approval.Reason, &approval.Status, &approval.ResolutionReason, &approval.ResolvedBy, &approval.CreatedAt, &approval.ResolvedAt)
	if err != nil {
		return model.ApprovalRequest{}, ErrNotFound
	}
	approval.RequestedRole = model.Role(role)
	return approval, nil
}

func scanAssistantSession(row scanner) (model.AssistantSession, error) {
	var session model.AssistantSession
	err := row.Scan(&session.ID, &session.ClusterID, &session.IncidentID, &session.Title, &session.CreatedBy, &session.LastSummary, &session.CreatedAt, &session.UpdatedAt)
	if err != nil {
		return model.AssistantSession{}, ErrNotFound
	}
	return session, nil
}

func scanAssistantMessage(row scanner) (model.AssistantMessage, error) {
	var msg model.AssistantMessage
	var citationsRaw, actionRaw []byte
	err := row.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &citationsRaw, &actionRaw, &msg.CreatedAt)
	if err != nil {
		return model.AssistantMessage{}, ErrNotFound
	}
	_ = json.Unmarshal(citationsRaw, &msg.Citations)
	_ = json.Unmarshal(actionRaw, &msg.ActionProposals)
	return msg, nil
}

func nilJSON(value any) any {
	if value == nil {
		return nil
	}
	return mustJSON(value)
}

func mustJSON(value any) []byte {
	blob, _ := json.Marshal(value)
	return blob
}

func mapOrEmpty(value map[string]any) map[string]any {
	if value == nil {
		return map[string]any{}
	}
	return value
}

func cloneCluster(cluster model.Cluster) model.Cluster {
	clone := cluster
	return clone
}

func cloneSession(session model.AuthSession) model.AuthSession {
	clone := session
	clone.Groups = append([]string(nil), session.Groups...)
	return clone
}

func cloneIncident(incident model.Incident) model.Incident {
	clone := incident
	clone.AffectedServices = append([]string(nil), incident.AffectedServices...)
	clone.DiagnosticRefs = append([]string(nil), incident.DiagnosticRefs...)
	clone.RunbookSteps = append([]model.RunbookStep(nil), incident.RunbookSteps...)
	clone.Timeline = append([]model.TimelineEntry(nil), incident.Timeline...)
	return clone
}

func cloneAudit(entry model.AuditEntry) model.AuditEntry {
	clone := entry
	return clone
}

func cloneActionRun(run model.ActionRun) model.ActionRun {
	clone := run
	clone.Plan = append([]string(nil), run.Plan...)
	clone.Params = mapOrEmpty(run.Params)
	if len(run.Params) > 0 {
		clone.Params = make(map[string]any, len(run.Params))
		for key, value := range run.Params {
			clone.Params[key] = value
		}
	}
	return clone
}

func cloneApproval(approval model.ApprovalRequest) model.ApprovalRequest {
	clone := approval
	return clone
}

func cloneAssistantMessage(message model.AssistantMessage) model.AssistantMessage {
	clone := message
	clone.Citations = append([]model.AssistantCitation(nil), message.Citations...)
	clone.ActionProposals = append([]model.AssistantActionProposal(nil), message.ActionProposals...)
	return clone
}

func cloneAssistantSession(session model.AssistantSession) model.AssistantSession {
	clone := session
	clone.Messages = make([]model.AssistantMessage, 0, len(session.Messages))
	for _, msg := range session.Messages {
		clone.Messages = append(clone.Messages, cloneAssistantMessage(msg))
	}
	return clone
}
