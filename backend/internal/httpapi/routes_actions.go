package httpapi

import "net/http"

func (d *deps) registerActionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/nodes/{id}/drain", d.authMiddleware(d.writeMiddleware(d.handleNodeDrain)))
	mux.HandleFunc("POST /api/v1/nodes/{id}/activate", d.authMiddleware(d.writeMiddleware(d.handleNodeActivate)))

	mux.HandleFunc("POST /api/v1/stacks/{name}/deploy", d.authMiddleware(d.writeMiddleware(d.handleStackDeploy)))
	mux.HandleFunc("DELETE /api/v1/stacks/{name}", d.authMiddleware(d.writeMiddleware(d.handleStackRemove)))

	mux.HandleFunc("POST /api/v1/services/{id}/scale", d.authMiddleware(d.writeMiddleware(d.handleServiceScale)))
	mux.HandleFunc("POST /api/v1/services/{id}/restart", d.authMiddleware(d.writeMiddleware(d.handleServiceRestart)))
	mux.HandleFunc("POST /api/v1/services/{id}/update", d.authMiddleware(d.writeMiddleware(d.handleServiceUpdate)))
	mux.HandleFunc("POST /api/v1/services/{id}/rollback", d.authMiddleware(d.writeMiddleware(d.handleServiceRollback)))

	mux.HandleFunc("POST /api/v1/tasks/{id}/restart", d.authMiddleware(d.writeMiddleware(d.handleTaskRestart)))

	mux.HandleFunc("POST /api/v1/actions/execute", d.authMiddleware(d.handleActionExecute))
	mux.HandleFunc("POST /api/v1/assistant/chat", d.authMiddleware(d.handleAssistantChat))
}
