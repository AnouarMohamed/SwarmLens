package httpapi

import (
	"encoding/json"
	"net/http"
)

// All write handlers follow this pattern:
// 1. Extract principal (already verified by writeMiddleware)
// 2. Parse + validate request body
// 3. Execute action via Docker API (or demo no-op)
// 4. Write audit record
// 5. Return updated resource

// ── Nodes ─────────────────────────────────────────────────────────────────────

func (d *deps) handleNodeDrain(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "node.drain", "node", id, nil, map[string]string{"availability": "drain"}, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "node drain simulated (demo mode)"})
		return
	}
	// Phase 2: real Docker API call here
	writeError(w, http.StatusNotImplemented, "not_implemented", "node drain requires Phase 2 implementation")
}

func (d *deps) handleNodeActivate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "node.activate", "node", id, nil, map[string]string{"availability": "active"}, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "node activate simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "node activate requires Phase 2 implementation")
}

// ── Stacks ────────────────────────────────────────────────────────────────────

func (d *deps) handleStackDeploy(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p := principalFrom(r.Context())
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "stack.deploy", "stack", name, nil, body, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "stack deploy simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "stack deploy requires Phase 2 implementation")
}

func (d *deps) handleStackRemove(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	p := principalFrom(r.Context())
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "stack.remove", "stack", name, nil, nil, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "stack remove simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "stack remove requires Phase 2 implementation")
}

// ── Services ──────────────────────────────────────────────────────────────────

func (d *deps) handleServiceScale(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	var body struct {
		Replicas int `json:"replicas"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Replicas < 0 {
		writeError(w, http.StatusBadRequest, "bad_request", "replicas must be a non-negative integer")
		return
	}
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "service.scale", "service", id, nil, map[string]int{"replicas": body.Replicas}, "success", "demo mode")
		writeOK(w, map[string]interface{}{"status": "ok", "replicas": body.Replicas, "message": "scale simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "service scale requires Phase 2 implementation")
}

func (d *deps) handleServiceRestart(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "service.restart", "service", id, nil, nil, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "rolling restart simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "service restart requires Phase 2 implementation")
}

func (d *deps) handleServiceUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	var body map[string]interface{}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "service.update", "service", id, nil, body, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "service update simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "service update requires Phase 2 implementation")
}

func (d *deps) handleServiceRollback(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if d.docker.IsDemo() {
		d.auditLog.Record(p.Username, string(p.Role), "service.rollback", "service", id, nil, nil, "success", "demo mode")
		writeOK(w, map[string]string{"status": "ok", "message": "service rollback simulated (demo mode)"})
		return
	}
	writeError(w, http.StatusNotImplemented, "not_implemented", "service rollback requires Phase 2 implementation")
}
