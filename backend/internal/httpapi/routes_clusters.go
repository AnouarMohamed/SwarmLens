package httpapi

import "net/http"

func (d *deps) registerClusterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/clusters", d.authMiddleware(d.handleClustersList))
	mux.HandleFunc("POST /api/v1/clusters", d.authMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleClustersCreate))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}", d.authMiddleware(d.handleClustersGet))
	mux.HandleFunc("PUT /api/v1/clusters/{clusterID}", d.authMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleClustersUpdate))))
}
