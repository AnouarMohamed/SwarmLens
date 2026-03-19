package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
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
	Params     map[string]interface{} `json:"params"`
}

type actionRequest struct {
	Action     string
	Resource   string
	ResourceID string
	Params     map[string]interface{}
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
	outcome := d.executeAction(r.Context(), p, actionRequest{
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Params:     req.Params,
	})
	writeOK(w, outcome)
}

func (d *deps) handleNodeDrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "node.drain",
		Resource:   "node",
		ResourceID: id,
	})
	writeOK(w, outcome)
}

func (d *deps) handleNodeActivate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "node.activate",
		Resource:   "node",
		ResourceID: id,
	})
	writeOK(w, outcome)
}

func (d *deps) handleStackDeploy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	params := map[string]interface{}{}
	_ = json.NewDecoder(r.Body).Decode(&params)
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "stack.deploy",
		Resource:   "stack",
		ResourceID: name,
		Params:     params,
	})
	writeOK(w, outcome)
}

func (d *deps) handleStackRemove(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "stack.remove",
		Resource:   "stack",
		ResourceID: name,
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceScale(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var body struct {
		Replicas int `json:"replicas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Replicas < 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "replicas must be a non-negative integer")
		return
	}
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "service.scale",
		Resource:   "service",
		ResourceID: id,
		Params:     map[string]interface{}{"replicas": body.Replicas},
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "service.restart",
		Resource:   "service",
		ResourceID: id,
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	params := map[string]interface{}{}
	_ = json.NewDecoder(r.Body).Decode(&params)
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "service.update",
		Resource:   "service",
		ResourceID: id,
		Params:     params,
	})
	writeOK(w, outcome)
}

func (d *deps) handleServiceRollback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "service.rollback",
		Resource:   "service",
		ResourceID: id,
	})
	writeOK(w, outcome)
}

func (d *deps) handleTaskRestart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	outcome := d.executeAction(r.Context(), principalFrom(r.Context()), actionRequest{
		Action:     "task.restart",
		Resource:   "task",
		ResourceID: id,
	})
	writeOK(w, outcome)
}

func (d *deps) executeAction(ctx context.Context, p model.Principal, req actionRequest) model.ActionOutcome {
	outcome := model.ActionOutcome{
		Action:     req.Action,
		Resource:   req.Resource,
		ResourceID: req.ResourceID,
		Mode:       "dry_run",
		Status:     model.ActionStatusDryRun,
		Executed:   false,
		Timestamp:  time.Now().UTC(),
	}

	if req.Resource == "" {
		req.Resource = resourceFromAction(req.Action)
		outcome.Resource = req.Resource
	}

	if isReadOnlyAction(req.Action) {
		switch req.Action {
		case "diagnostics.run":
			findings := d.runDiagnostics()
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Message = fmt.Sprintf("Diagnostics completed with %d finding(s).", len(findings))
			outcome.Impact = "Cluster findings refreshed."
		case "telemetry.refresh":
			if err := d.ensureSnapshotFresh(ctx, true); err != nil {
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
			title, _ := req.Params["title"].(string)
			desc, _ := req.Params["description"].(string)
			if title == "" {
				title = "Operator-generated incident"
			}
			inc := d.incidents.Create(title, desc, "high", p.Username, nil, nil)
			outcome.Mode = "live"
			outcome.Status = model.ActionStatusSuccess
			outcome.Executed = true
			outcome.Resource = "incident"
			outcome.ResourceID = inc.ID
			outcome.Message = "Incident created successfully."
			outcome.Impact = fmt.Sprintf("Incident %s is now open.", inc.ID)
		default:
			outcome.Mode = "blocked"
			outcome.Status = model.ActionStatusBlocked
			outcome.BlockedReason = "unknown_action"
			outcome.Message = "Action is not supported."
		}
		return d.auditAction(p, req, outcome)
	}

	isDemo := d.docker != nil && d.docker.IsDemo()
	if d.cfg.LiveActionPolicy == "demo_only" && !isDemo {
		outcome.Mode = "blocked"
		outcome.Status = model.ActionStatusBlocked
		outcome.BlockedReason = "policy_demo_only"
		outcome.Message = "Live mutation is blocked by policy."
		outcome.Impact = "Switch policy or use demo mode."
		outcome.Plan = dryRunPlan(req)
		return d.auditAction(p, req, outcome)
	}

	if isDemo {
		outcome.Mode = "demo"
		outcome.Status = model.ActionStatusSuccess
		outcome.Executed = true
		outcome.Message = "Action simulated in demo mode."
		outcome.Impact = "No real Docker mutation executed."
		return d.auditAction(p, req, outcome)
	}

	outcome.Mode = "dry_run"
	outcome.Status = model.ActionStatusDryRun
	outcome.Executed = false
	outcome.Message = "Validated and generated an execution plan."
	outcome.Impact = "No live mutation executed."
	outcome.Plan = dryRunPlan(req)
	if d.cfg.LiveActionPolicy == "allowlist_live" && liveAllowlist(req.Action) {
		outcome.BlockedReason = "action_not_implemented"
		outcome.Message = "Action is allowlisted but still running in dry-run until adapter is implemented."
	}
	return d.auditAction(p, req, outcome)
}

func (d *deps) auditAction(p model.Principal, req actionRequest, outcome model.ActionOutcome) model.ActionOutcome {
	result := "success"
	switch outcome.Status {
	case model.ActionStatusDryRun:
		result = "pending_approval"
	case model.ActionStatusBlocked:
		result = "rejected"
	case model.ActionStatusFailed:
		result = "failed"
	}

	entry := d.auditLog.Record(
		p.Username,
		string(p.Role),
		req.Action,
		req.Resource,
		req.ResourceID,
		nil,
		req.Params,
		result,
		outcome.Message,
	)
	outcome.AuditID = entry.ID
	return outcome
}

func isReadOnlyAction(action string) bool {
	switch action {
	case "diagnostics.run", "telemetry.refresh", "incident.create":
		return true
	default:
		return false
	}
}

func liveAllowlist(action string) bool {
	switch action {
	case "service.scale", "node.drain", "task.restart":
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

func dryRunPlan(req actionRequest) []string {
	switch req.Action {
	case "service.scale":
		return []string{
			"Read current service spec and desired replicas.",
			"Validate new replica target against policy constraints.",
			"Apply scale update and monitor reconciliation until stable.",
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
