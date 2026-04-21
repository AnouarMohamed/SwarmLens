package httpapi

import "net/http"

func (d *deps) registerDiagnosticsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/diagnostics", d.authMiddleware(d.clusterMiddleware(d.handleDiagnosticsList)))
	mux.HandleFunc("POST /api/v1/diagnostics/run", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleDiagnosticsRun))))
	mux.HandleFunc("GET /api/v1/diagnostics/{id}", d.authMiddleware(d.clusterMiddleware(d.handleDiagnosticsGet)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/diagnostics", d.authMiddleware(d.clusterMiddleware(d.handleDiagnosticsList)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/diagnostics/run", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleDiagnosticsRun))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/diagnostics/{id}", d.authMiddleware(d.clusterMiddleware(d.handleDiagnosticsGet)))
}
