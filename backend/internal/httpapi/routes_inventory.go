package httpapi

import "net/http"

func (d *deps) registerInventoryRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/swarm", d.authMiddleware(d.clusterMiddleware(d.handleSwarm)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/swarm", d.authMiddleware(d.clusterMiddleware(d.handleSwarm)))

	mux.HandleFunc("GET /api/v1/nodes", d.authMiddleware(d.clusterMiddleware(d.handleNodesList)))
	mux.HandleFunc("GET /api/v1/nodes/{id}", d.authMiddleware(d.clusterMiddleware(d.handleNodesGet)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/nodes", d.authMiddleware(d.clusterMiddleware(d.handleNodesList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/nodes/{id}", d.authMiddleware(d.clusterMiddleware(d.handleNodesGet)))

	mux.HandleFunc("GET /api/v1/stacks", d.authMiddleware(d.clusterMiddleware(d.handleStacksList)))
	mux.HandleFunc("GET /api/v1/stacks/{name}", d.authMiddleware(d.clusterMiddleware(d.handleStacksGet)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/stacks", d.authMiddleware(d.clusterMiddleware(d.handleStacksList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/stacks/{name}", d.authMiddleware(d.clusterMiddleware(d.handleStacksGet)))

	mux.HandleFunc("GET /api/v1/services", d.authMiddleware(d.clusterMiddleware(d.handleServicesList)))
	mux.HandleFunc("GET /api/v1/services/{id}", d.authMiddleware(d.clusterMiddleware(d.handleServicesGet)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/services", d.authMiddleware(d.clusterMiddleware(d.handleServicesList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/services/{id}", d.authMiddleware(d.clusterMiddleware(d.handleServicesGet)))

	mux.HandleFunc("GET /api/v1/tasks", d.authMiddleware(d.clusterMiddleware(d.handleTasksList)))
	mux.HandleFunc("GET /api/v1/tasks/{id}", d.authMiddleware(d.clusterMiddleware(d.handleTasksGet)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/tasks", d.authMiddleware(d.clusterMiddleware(d.handleTasksList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/tasks/{id}", d.authMiddleware(d.clusterMiddleware(d.handleTasksGet)))

	mux.HandleFunc("GET /api/v1/networks", d.authMiddleware(d.clusterMiddleware(d.handleNetworksList)))
	mux.HandleFunc("GET /api/v1/volumes", d.authMiddleware(d.clusterMiddleware(d.handleVolumesList)))
	mux.HandleFunc("GET /api/v1/secrets", d.authMiddleware(d.clusterMiddleware(d.handleSecretsList)))
	mux.HandleFunc("GET /api/v1/configs", d.authMiddleware(d.clusterMiddleware(d.handleConfigsList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/networks", d.authMiddleware(d.clusterMiddleware(d.handleNetworksList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/volumes", d.authMiddleware(d.clusterMiddleware(d.handleVolumesList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/secrets", d.authMiddleware(d.clusterMiddleware(d.handleSecretsList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/configs", d.authMiddleware(d.clusterMiddleware(d.handleConfigsList)))

	mux.HandleFunc("GET /api/v1/events", d.authMiddleware(d.clusterMiddleware(d.handleEventsList)))
	mux.HandleFunc("GET /api/v1/stream/events", d.authMiddleware(d.clusterMiddleware(d.handleStreamEvents)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/events", d.authMiddleware(d.clusterMiddleware(d.handleEventsList)))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/stream/events", d.authMiddleware(d.clusterMiddleware(d.handleStreamEvents)))
}
