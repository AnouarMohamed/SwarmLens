package httpapi

import "net/http"

func (d *deps) registerAuditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/audit", d.authMiddleware(d.handleAuditList))
}
