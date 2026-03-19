package httpapi

import "net/http"

func (d *deps) registerIncidentsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/incidents", d.authMiddleware(d.handleIncidentsList))
	mux.HandleFunc("POST /api/v1/incidents", d.authMiddleware(d.handleIncidentsCreate))
	mux.HandleFunc("GET /api/v1/incidents/{id}", d.authMiddleware(d.handleIncidentsGet))
	mux.HandleFunc("PUT /api/v1/incidents/{id}", d.authMiddleware(d.handleIncidentsUpdate))
	mux.HandleFunc("POST /api/v1/incidents/{id}/resolve", d.authMiddleware(d.handleIncidentsResolve))
}
