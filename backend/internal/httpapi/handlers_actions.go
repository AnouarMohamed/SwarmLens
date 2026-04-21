package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/auth"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type actionExecuteRequest struct {
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	ResourceID string                 `json:"resourceID"`
	Reason     string                 `json:"reason"`
	Params     map[string]interface{} `json:"params"`
}

type actionRequest struct {
	Action     string
	Resource   string
	ResourceID string
	Reason     string
	Params     map[string]interface{}
}

func (d *deps) handleActionRunsList(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	runs, err := d.store.ListActionRuns(r.Context(), cluster.ID, 100)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "action_list_failed", err.Error())
		return
	}
	writeList(w, runs, len(runs))
}

func (d *deps) handleApprovalsList(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	status := model.ApprovalStatus(strings.TrimSpace(r.URL.Query().Get("status")))
	approvals, err := d.store.ListApprovals(r.Context(), cluster.ID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "approval_list_failed", err.Error())
		return
	}
	writeList(w, approvals, len(approvals))
}

func (d *deps) handleApproveAction(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	p := principalFrom(r.Context())
	approval, err := d.store.GetApproval(r.Context(), r.PathValue("id"))
	if err != nil || approval.ClusterID != cluster.ID {
		writeError(w, http.StatusNotFound, "not_found", "approval request not found")
		return
	}
	run, err := d.store.GetActionRun(r.Context(), approval.ActionRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "action run not found")
		return
	}
	if approval.Status != model.ApprovalStatusPending {
		writeError(w, http.StatusConflict, "approval_closed", "approval request is already resolved")
		return
	}
	if _, err := d.store.ResolveApproval(r.Context(), approval.ID, model.ApprovalStatusApproved, p.Username, "approved"); err != nil {
		writeError(w, http.StatusInternalServerError, "approval_update_failed", err.Error())
		return
	}
	run.Status = model.ActionStatusDryRun
	run.ApprovalID = approval.ID
	outcome := d.fulfillActionRun(r.Context(), cluster, runtimeFrom(r.Context()), p, run)
	d.approvalCount.Add(1)
	writeOK(w, outcome)
}

func (d *deps) handleRejectAction(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	p := principalFrom(r.Context())
	approval, err := d.store.GetApproval(r.Context(), r.PathValue("id"))
	if err != nil || approval.ClusterID != cluster.ID {
		writeError(w, http.StatusNotFound, "not_found", "approval request not found")
		return
	}
	run, err := d.store.GetActionRun(r.Context(), approval.ActionRunID)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "action run not found")
		return
	}
	if approval.Status != model.ApprovalStatusPending {
		writeError(w, http.StatusConflict, "approval_closed", "approval request is already resolved")
		return
	}
	if _, err := d.store.ResolveApproval(r.Context(), approval.ID, model.ApprovalStatusRejected, p.Username, "rejected"); err != nil {
		writeError(w, http.StatusInternalServerError, "approval_update_failed", err.Error())
		return
	}
	run.Status = model.ActionStatusBlocked
	run.Mode = "blocked"
	run.Message = "Action was rejected during approval."
	run.BlockedReason = "approval_rejected"
	run.Executed = false
	run, _ = d.store.UpdateActionRun(r.Context(), run)
	auditEntry, _ := d.store.RecordAudit(r.Context(), model.AuditEntry{
		ClusterID:   cluster.ID,
		ActionRunID: run.ID,
		Actor:       p.Username,
		Role:        string(p.Role),
		Action:      run.Action,
		Resource:    run.Resource,
		ResourceID:  run.ResourceID,
		AfterSpec:   map[string]any{"params": run.Params},
		Result:      "rejected",
		Reason:      run.Message,
	})
	d.approvalCount.Add(1)
	writeOK(w, toActionOutcome(run, auditEntry.ID))
}

func (d *deps) handleActionExecute(w http.ResponseWriter, r *http.Request) {
	var req actionExecuteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Action == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "action is required")
		return
	}
	p := principalFrom(r.Context())
	if !isReadOnlyAction(req.Action) {
		if err := d.gate.Check(); err != nil {
			writeError(w, http.StatusForbidden, "writes_disabled", err.Error())
			return
		}
		if err := auth.Require(p, model.RoleOperator); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
	}
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), p, actionRequest{
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Reason:     req.Reason,
		Params:     req.Params,
	})
	writeOK(w, outcome)
}

func (d *deps) handleNodeDrain(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "node.drain",
		Resource:   "node",
		ResourceID: r.PathValue("id"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) handleNodeActivate(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "node.activate",
		Resource:   "node",
		ResourceID: r.PathValue("id"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) handleStackDeploy(w http.ResponseWriter, r *http.Request) {
	body := readJSONMap(r)
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "stack.deploy",
		Resource:   "stack",
		ResourceID: r.PathValue("name"),
		Reason:     stringValue(body["reason"]),
		Params:     body,
	})
	writeOK(w, outcome)
}

func (d *deps) handleStackRemove(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "stack.remove",
		Resource:   "stack",
		ResourceID: r.PathValue("name"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceScale(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Replicas int    `json:"replicas"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Replicas < 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "replicas must be a non-negative integer")
		return
	}
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "service.scale",
		Resource:   "service",
		ResourceID: r.PathValue("id"),
		Reason:     body.Reason,
		Params:     map[string]interface{}{"replicas": body.Replicas},
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "service.restart",
		Resource:   "service",
		ResourceID: r.PathValue("id"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceUpdate(w http.ResponseWriter, r *http.Request) {
	params := readJSONMap(r)
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "service.update",
		Resource:   "service",
		ResourceID: r.PathValue("id"),
		Reason:     stringValue(params["reason"]),
		Params:     params,
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceRollback(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "service.rollback",
		Resource:   "service",
		ResourceID: r.PathValue("id"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) handleTaskRestart(w http.ResponseWriter, r *http.Request) {
	outcome := d.executeAction(r.Context(), clusterFrom(r.Context()), runtimeFrom(r.Context()), principalFrom(r.Context()), actionRequest{
		Action:     "task.restart",
		Resource:   "task",
		ResourceID: r.PathValue("id"),
		Reason:     readActionReason(r),
	})
	writeOK(w, outcome)
}

func (d *deps) executeAction(ctx context.Context, cluster model.Cluster, runtime *clusterRuntime, p model.Principal, req actionRequest) model.ActionOutcome {
	if req.Resource == "" {
		req.Resource = resourceFromAction(req.Action)
	}
	req.Reason = strings.TrimSpace(req.Reason)

	run := model.ActionRun{
		ClusterID:     cluster.ID,
		Action:        req.Action,
		Resource:      req.Resource,
		ResourceID:    req.ResourceID,
		RequestedBy:   p.Username,
		RequestedRole: p.Role,
		Reason:        req.Reason,
		Status:        model.ActionStatusDryRun,
		Mode:          "dry_run",
		Params:        req.Params,
	}
	run, _ = d.store.CreateActionRun(ctx, run)
	d.actionCount.Add(1)
	if runtime != nil {
		_ = d.ensureSnapshotFresh(ctx, runtime, false)
	}

	if actionNeedsReason(req.Action) && req.Reason == "" {
		run.Status = model.ActionStatusBlocked
		run.Mode = "blocked"
		run.BlockedReason = "reason_required"
		run.Message = "A reason is required for this action."
		run, _ = d.store.UpdateActionRun(ctx, run)
		return toActionOutcome(run, "")
	}

	if isApprovalRequired(req, runtime, d.cfg.ActionSafeScaleDelta) {
		approval, _ := d.store.CreateApproval(ctx, model.ApprovalRequest{
			ClusterID:     cluster.ID,
			ActionRunID:   run.ID,
			Action:        req.Action,
			Resource:      req.Resource,
			ResourceID:    req.ResourceID,
			RequestedBy:   p.Username,
			RequestedRole: p.Role,
			Reason:        req.Reason,
			Status:        model.ApprovalStatusPending,
		})
		run.Status = model.ActionStatusPendingApproval
		run.Mode = "approval"
		run.ApprovalRequired = true
		run.ApprovalID = approval.ID
		run.Message = "Action is pending admin approval."
		run, _ = d.store.UpdateActionRun(ctx, run)
		auditEntry, _ := d.store.RecordAudit(ctx, model.AuditEntry{
			ClusterID:   cluster.ID,
			ActionRunID: run.ID,
			Actor:       p.Username,
			Role:        string(p.Role),
			Action:      req.Action,
			Resource:    req.Resource,
			ResourceID:  req.ResourceID,
			AfterSpec:   map[string]any{"params": req.Params},
			Result:      "pending_approval",
			Reason:      run.Message,
		})
		return toActionOutcome(run, auditEntry.ID)
	}

	return d.fulfillActionRun(ctx, cluster, runtime, p, run)
}

func (d *deps) fulfillActionRun(ctx context.Context, cluster model.Cluster, runtime *clusterRuntime, p model.Principal, run model.ActionRun) model.ActionOutcome {
	outcome := model.ActionOutcome{
		ID:               run.ID,
		ClusterID:        cluster.ID,
		Action:           run.Action,
		Resource:         run.Resource,
		ResourceID:       run.ResourceID,
		Reason:           run.Reason,
		Status:           model.ActionStatusDryRun,
		Mode:             "dry_run",
		Executed:         false,
		ApprovalID:       run.ApprovalID,
		ApprovalRequired: run.ApprovalRequired,
		Timestamp:        time.Now().UTC(),
	}

	if isReadOnlyAction(run.Action) {
		switch run.Action {
		case "diagnostics.run":
			findings := d.runDiagnostics(ctx, runtime)
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Message = fmt.Sprintf("Diagnostics completed with %d finding(s).", len(findings))
			outcome.Impact = "Cluster findings refreshed."
		case "telemetry.refresh":
			if err := d.ensureSnapshotFresh(ctx, runtime, true); err != nil {
				outcome.Mode = "live"
				outcome.Status = model.ActionStatusFailed
				outcome.Message = "Telemetry refresh failed."
				outcome.Impact = err.Error()
			} else {
				outcome.Mode = "live"
				outcome.Status = model.ActionStatusSuccess
				outcome.Executed = true
				outcome.Message = "Telemetry refresh completed."
				outcome.Impact = "Snapshot and events were refreshed."
			}
		case "incident.create":
			title := stringValue(run.Params["title"])
			if title == "" {
				title = "Operator-generated incident"
			}
			desc := stringValue(run.Params["description"])
			incident, err := d.store.CreateIncident(ctx, cluster.ID, title, desc, "high", p.Username, nil, nil)
			if err != nil {
				outcome.Mode = "live"
				outcome.Status = model.ActionStatusFailed
				outcome.Message = "Incident creation failed."
				outcome.Impact = err.Error()
			} else {
				outcome.Mode = "live"
				outcome.Status = model.ActionStatusSuccess
				outcome.Executed = true
				outcome.Resource = "incident"
				outcome.ResourceID = incident.ID
				outcome.Message = "Incident created successfully."
				outcome.Impact = fmt.Sprintf("Incident %s is now open.", incident.ID)
			}
		default:
			outcome.Mode = "blocked"
			outcome.Status = model.ActionStatusBlocked
			outcome.BlockedReason = "unknown_action"
			outcome.Message = "Action is not supported."
		}
		return d.finalizeActionRun(ctx, p, run, outcome)
	}

	if d.cfg.LiveActionPolicy == "demo_only" && runtime != nil && runtime.docker != nil && !runtime.docker.IsDemo() {
		outcome.Mode = "blocked"
		outcome.Status = model.ActionStatusBlocked
		outcome.BlockedReason = "policy_demo_only"
		outcome.Message = "Live mutation is blocked by policy."
		outcome.Impact = "Switch policy or use demo mode."
		outcome.Plan = dryRunPlan(requestFromRun(run))
		return d.finalizeActionRun(ctx, p, run, outcome)
	}

	if runtime != nil && runtime.docker != nil && runtime.docker.IsDemo() {
		outcome.Mode = "demo"
		outcome.Status = model.ActionStatusSuccess
		outcome.Executed = true
		outcome.Message = "Action simulated in demo mode."
		outcome.Impact = "No real Docker mutation executed."
		return d.finalizeActionRun(ctx, p, run, outcome)
	}

	switch run.Action {
	case "service.restart":
		if runtime == nil || runtime.docker == nil {
			outcome.Mode = "blocked"
			outcome.Status = model.ActionStatusBlocked
			outcome.BlockedReason = "runtime_unavailable"
			outcome.Message = "Cluster runtime is not available for live service actions."
			return d.finalizeActionRun(ctx, p, run, outcome)
		}
		if err := runtime.docker.ServiceRestart(ctx, run.ResourceID); err != nil {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusFailed
			outcome.Message = "Service restart failed."
			outcome.Impact = err.Error()
		} else {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Message = "Service restart requested."
			outcome.Impact = "Docker Swarm will reconcile the restarted service."
		}
	case "service.scale":
		if runtime == nil || runtime.docker == nil {
			outcome.Mode = "blocked"
			outcome.Status = model.ActionStatusBlocked
			outcome.BlockedReason = "runtime_unavailable"
			outcome.Message = "Cluster runtime is not available for live service actions."
			return d.finalizeActionRun(ctx, p, run, outcome)
		}
		replicas, _ := intValue(run.Params["replicas"])
		if err := runtime.docker.ServiceScale(ctx, run.ResourceID, uint64(replicas)); err != nil {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusFailed
			outcome.Message = "Service scale failed."
			outcome.Impact = err.Error()
		} else {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Message = fmt.Sprintf("Service scaled to %d replicas.", replicas)
			outcome.Impact = "Replica reconciliation has been requested from Docker Swarm."
		}
	case "service.rollback":
		if runtime == nil || runtime.docker == nil {
			outcome.Mode = "blocked"
			outcome.Status = model.ActionStatusBlocked
			outcome.BlockedReason = "runtime_unavailable"
			outcome.Message = "Cluster runtime is not available for live service actions."
			return d.finalizeActionRun(ctx, p, run, outcome)
		}
		if err := runtime.docker.ServiceRollback(ctx, run.ResourceID); err != nil {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusFailed
			outcome.Message = "Service rollback failed."
			outcome.Impact = err.Error()
		} else {
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Message = "Service rollback requested."
			outcome.Impact = "Docker Swarm will attempt to roll back to the previous service spec."
		}
	default:
		outcome.Mode = "dry_run"
		outcome.Status = model.ActionStatusDryRun
		outcome.Message = "Validated and generated an execution plan."
		outcome.Impact = "No live mutation executed."
		outcome.Plan = dryRunPlan(requestFromRun(run))
		if d.cfg.LiveActionPolicy == "allowlist_live" && liveAllowlist(run.Action) {
			outcome.BlockedReason = "action_not_implemented"
			outcome.Message = "Action is allowlisted but still running in dry-run until adapter is implemented."
		}
	}
	return d.finalizeActionRun(ctx, p, run, outcome)
}

func (d *deps) finalizeActionRun(ctx context.Context, p model.Principal, run model.ActionRun, outcome model.ActionOutcome) model.ActionOutcome {
	run.Status = outcome.Status
	run.Mode = outcome.Mode
	run.Executed = outcome.Executed
	run.Message = outcome.Message
	run.BlockedReason = outcome.BlockedReason
	run.Impact = outcome.Impact
	run.Plan = append([]string(nil), outcome.Plan...)
	run.ApprovalID = outcome.ApprovalID
	run.ApprovalRequired = outcome.ApprovalRequired
	updated, _ := d.store.UpdateActionRun(ctx, run)

	result := "success"
	switch outcome.Status {
	case model.ActionStatusDryRun, model.ActionStatusPendingApproval:
		result = "pending_approval"
	case model.ActionStatusBlocked:
		result = "rejected"
	case model.ActionStatusFailed:
		result = "failed"
	}
	auditEntry, _ := d.store.RecordAudit(ctx, model.AuditEntry{
		ClusterID:   run.ClusterID,
		ActionRunID: updated.ID,
		Actor:       p.Username,
		Role:        string(p.Role),
		Action:      run.Action,
		Resource:    run.Resource,
		ResourceID:  run.ResourceID,
		AfterSpec:   map[string]any{"params": run.Params},
		Result:      result,
		Reason:      outcome.Message,
	})
	updated.AuditID = auditEntry.ID
	updated, _ = d.store.UpdateActionRun(ctx, updated)
	return toActionOutcome(updated, auditEntry.ID)
}

func isReadOnlyAction(action string) bool {
	switch action {
	case "diagnostics.run", "telemetry.refresh", "incident.create":
		return true
	default:
		return false
	}
}

func actionNeedsReason(action string) bool {
	switch action {
	case "diagnostics.run", "telemetry.refresh", "incident.create", "service.restart", "service.scale", "service.rollback":
		return true
	default:
		return !strings.HasPrefix(action, "assistant.")
	}
}

func liveAllowlist(action string) bool {
	switch action {
	case "service.scale", "service.restart", "service.rollback":
		return true
	default:
		return false
	}
}

func resourceFromAction(action string) string {
	parts := strings.Split(action, ".")
	if len(parts) == 0 {
		return "cluster"
	}
	return parts[0]
}

func isApprovalRequired(req actionRequest, runtime *clusterRuntime, safeDelta int) bool {
	switch req.Action {
	case "service.rollback":
		return true
	case "service.scale":
		target, ok := intValue(req.Params["replicas"])
		if !ok {
			return false
		}
		current := currentServiceReplicas(runtime, req.ResourceID)
		return int(math.Abs(float64(target-current))) > safeDelta
	default:
		return false
	}
}

func currentServiceReplicas(runtime *clusterRuntime, serviceID string) int {
	if runtime == nil {
		return 0
	}
	snap, _ := runtime.cache.GetSnapshot()
	for _, service := range snap.Services {
		if service.ID == serviceID || service.Name == serviceID {
			return service.DesiredReplicas
		}
	}
	return 0
}

func dryRunPlan(req actionRequest) []string {
	switch req.Action {
	case "service.scale":
		return []string{
			"Read current service spec and desired replicas.",
			"Validate new replica target against policy constraints.",
			"Apply scale update and monitor reconciliation until stable.",
		}
	case "service.restart":
		return []string{
			"Inspect current service update status.",
			"Trigger a forced service update to recreate tasks.",
			"Watch the task rollout until the service stabilizes.",
		}
	case "service.rollback":
		return []string{
			"Read current and previous service specs.",
			"Request Docker Swarm rollback to the previous version.",
			"Validate recovery and monitor any follow-up failures.",
		}
	case "node.drain", "node.activate":
		return []string{
			"Inspect node role and current task allocation.",
			"Change availability state and monitor task rescheduling.",
			"Verify quorum and capacity after scheduling settles.",
		}
	case "task.restart":
		return []string{
			"Resolve task owner service and current rollout state.",
			"Trigger controlled rolling update on owning service.",
			"Validate task recovery and error-rate stabilization.",
		}
	default:
		return []string{
			"Validate request shape and authorization.",
			"Construct Docker mutation payload.",
			"Execute action and verify post-action health checks.",
		}
	}
}

func readActionReason(r *http.Request) string {
	body := readJSONMap(r)
	if reason := stringValue(body["reason"]); reason != "" {
		return reason
	}
	return strings.TrimSpace(r.URL.Query().Get("reason"))
}

func readJSONMap(r *http.Request) map[string]interface{} {
	if r.Body == nil {
		return map[string]interface{}{}
	}
	defer r.Body.Close()
	var params map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		return map[string]interface{}{}
	}
	if params == nil {
		return map[string]interface{}{}
	}
	return params
}

func stringValue(value interface{}) string {
	if raw, ok := value.(string); ok {
		return strings.TrimSpace(raw)
	}
	return ""
}

func intValue(value interface{}) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int32:
		return int(v), true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	case json.Number:
		n, err := v.Int64()
		return int(n), err == nil
	default:
		return 0, false
	}
}

func requestFromRun(run model.ActionRun) actionRequest {
	return actionRequest{
		Action:     run.Action,
		Resource:   run.Resource,
		ResourceID: run.ResourceID,
		Reason:     run.Reason,
		Params:     run.Params,
	}
}

func toActionOutcome(run model.ActionRun, auditID string) model.ActionOutcome {
	return model.ActionOutcome{
		ID:               run.ID,
		ClusterID:        run.ClusterID,
		Action:           run.Action,
		Resource:         run.Resource,
		ResourceID:       run.ResourceID,
		Reason:           run.Reason,
		Status:           run.Status,
		Mode:             run.Mode,
		Executed:         run.Executed,
		ApprovalID:       run.ApprovalID,
		ApprovalRequired: run.ApprovalRequired,
		Message:          run.Message,
		BlockedReason:    run.BlockedReason,
		Impact:           run.Impact,
		Plan:             append([]string(nil), run.Plan...),
		AuditID:          chooseAuditID(run.AuditID, auditID),
		Timestamp:        run.UpdatedAt,
	}
}

func chooseAuditID(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
