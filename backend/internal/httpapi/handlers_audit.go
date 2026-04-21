package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func (d *deps) handleAuditList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	cluster := clusterFrom(r.Context())
	entries, total, err := d.store.ListAudit(r.Context(), cluster.ID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "audit_list_failed", err.Error())
		return
	}
	writeList(w, entries, total)
}

func writeSSE(w http.ResponseWriter, event string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
}
