package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/AnouarMohamed/swarmlens/backend/internal/auth"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func (d *deps) handleIncidentsList(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	all, err := d.store.ListIncidents(r.Context(), cluster.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "incident_list_failed", err.Error())
		return
	}
	writeList(w, all, len(all))
}

func (d *deps) handleIncidentsCreate(w http.ResponseWriter, r *http.Request) {
	p := principalFrom(r.Context())
	if err := auth.Require(p, model.RoleOperator); err != nil {
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	cluster := clusterFrom(r.Context())
	var body struct {
		Title            string   `json:"title"`
		Description      string   `json:"description"`
		Severity         string   `json:"severity"`
		AffectedServices []string `json:"affectedServices"`
		DiagnosticRefs   []string `json:"diagnosticRefs"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Title == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "title is required")
		return
	}
	if body.Severity == "" {
		body.Severity = "medium"
	}
	inc, err := d.store.CreateIncident(r.Context(), cluster.ID, body.Title, body.Description, body.Severity, p.Username, body.AffectedServices, body.DiagnosticRefs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "incident_create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": inc})
}

func (d *deps) handleIncidentsGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cluster := clusterFrom(r.Context())
	inc, err := d.store.GetIncident(r.Context(), cluster.ID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}

func (d *deps) handleIncidentsUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if err := auth.Require(p, model.RoleOperator); err != nil {
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	cluster := clusterFrom(r.Context())
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "status is required")
		return
	}
	inc, err := d.store.UpdateIncidentStatus(r.Context(), cluster.ID, id, body.Status, p.Username, body.Note)
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}

func (d *deps) handleIncidentsResolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	if err := auth.Require(p, model.RoleOperator); err != nil {
		writeError(w, http.StatusForbidden, "forbidden", err.Error())
		return
	}
	cluster := clusterFrom(r.Context())
	inc, err := d.store.UpdateIncidentStatus(r.Context(), cluster.ID, id, "resolved", p.Username, "incident resolved")
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}
