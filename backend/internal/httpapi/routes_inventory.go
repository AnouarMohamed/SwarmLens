package httpapi

import "net/http"

func (d *deps) registerInventoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/swarm", d.authMiddleware(d.handleSwarm))

	mux.HandleFunc("GET /api/v1/nodes", d.authMiddleware(d.handleNodesList))
	mux.HandleFunc("GET /api/v1/nodes/{id}", d.authMiddleware(d.handleNodesGet))

	mux.HandleFunc("GET /api/v1/stacks", d.authMiddleware(d.handleStacksList))
	mux.HandleFunc("GET /api/v1/stacks/{name}", d.authMiddleware(d.handleStacksGet))

	mux.HandleFunc("GET /api/v1/services", d.authMiddleware(d.handleServicesList))
	mux.HandleFunc("GET /api/v1/services/{id}", d.authMiddleware(d.handleServicesGet))

	mux.HandleFunc("GET /api/v1/tasks", d.authMiddleware(d.handleTasksList))
	mux.HandleFunc("GET /api/v1/tasks/{id}", d.authMiddleware(d.handleTasksGet))

	mux.HandleFunc("GET /api/v1/networks", d.authMiddleware(d.handleNetworksList))
	mux.HandleFunc("GET /api/v1/volumes", d.authMiddleware(d.handleVolumesList))
	mux.HandleFunc("GET /api/v1/secrets", d.authMiddleware(d.handleSecretsList))
	mux.HandleFunc("GET /api/v1/configs", d.authMiddleware(d.handleConfigsList))

	mux.HandleFunc("GET /api/v1/events", d.authMiddleware(d.handleEventsList))
	mux.HandleFunc("GET /api/v1/stream/events", d.authMiddleware(d.handleStreamEvents))
}
