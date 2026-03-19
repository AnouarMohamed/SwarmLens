package httpapi

import "net/http"

func (d *deps) registerDiagnosticsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/diagnostics", d.authMiddleware(d.handleDiagnosticsList))
	mux.HandleFunc("POST /api/v1/diagnostics/run", d.authMiddleware(d.handleDiagnosticsRun))
	mux.HandleFunc("GET /api/v1/diagnostics/{id}", d.authMiddleware(d.handleDiagnosticsGet))
}
