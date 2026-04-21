package httpapi

import "net/http"

func (d *deps) registerAuditRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/audit", d.authMiddleware(d.clusterMiddleware(d.handleAuditList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/audit", d.authMiddleware(d.clusterMiddleware(d.handleAuditList)))
}
