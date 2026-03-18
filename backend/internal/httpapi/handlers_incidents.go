package httpapi

import (
	"encoding/json"
	"net/http"
)

func (d *deps) handleIncidentsList(w http.ResponseWriter, r *http.Request) {
	all := d.incidents.List()
	writeList(w, all, len(all))
}

func (d *deps) handleIncidentsCreate(w http.ResponseWriter, r *http.Request) {
	p := principalFrom(r.Context())
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
	inc := d.incidents.Create(body.Title, body.Description, body.Severity, p.Username, body.AffectedServices, body.DiagnosticRefs)
	writeJSON(w, http.StatusCreated, map[string]interface{}{"data": inc})
}

func (d *deps) handleIncidentsGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	inc, ok := d.incidents.Get(id)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}

func (d *deps) handleIncidentsUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeError(w, http.StatusBadRequest, "bad_request", "status is required")
		return
	}
	inc, ok := d.incidents.UpdateStatus(id, body.Status, p.Username, body.Note)
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}

func (d *deps) handleIncidentsResolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p := principalFrom(r.Context())
	inc, ok := d.incidents.UpdateStatus(id, "resolved", p.Username, "incident resolved")
	if !ok {
		writeError(w, http.StatusNotFound, "not_found", "incident not found")
		return
	}
	writeOK(w, inc)
}
