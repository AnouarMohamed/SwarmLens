package httpapi

import "net/http"

func (d *deps) registerIncidentsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/incidents", d.authMiddleware(d.clusterMiddleware(d.handleIncidentsList)))
	mux.HandleFunc("POST /api/v1/incidents", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsCreate))))
	mux.HandleFunc("GET /api/v1/incidents/{id}", d.authMiddleware(d.clusterMiddleware(d.handleIncidentsGet)))
	mux.HandleFunc("PUT /api/v1/incidents/{id}", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsUpdate))))
	mux.HandleFunc("POST /api/v1/incidents/{id}/resolve", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsResolve))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/incidents", d.authMiddleware(d.clusterMiddleware(d.handleIncidentsList)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/incidents", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsCreate))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/incidents/{id}", d.authMiddleware(d.clusterMiddleware(d.handleIncidentsGet)))
	mux.HandleFunc("PUT /api/v1/clusters/{clusterID}/incidents/{id}", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsUpdate))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/incidents/{id}/resolve", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleIncidentsResolve))))
}
