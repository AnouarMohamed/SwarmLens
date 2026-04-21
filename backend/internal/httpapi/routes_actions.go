package httpapi

import "net/http"

func (d *deps) registerActionRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/nodes/{id}/drain", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleNodeDrain)))))
	mux.HandleFunc("POST /api/v1/nodes/{id}/activate", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleNodeActivate)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/nodes/{id}/drain", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleNodeDrain)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/nodes/{id}/activate", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleNodeActivate)))))

	mux.HandleFunc("POST /api/v1/stacks/{name}/deploy", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleStackDeploy)))))
	mux.HandleFunc("DELETE /api/v1/stacks/{name}", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleStackRemove)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/stacks/{name}/deploy", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleStackDeploy)))))
	mux.HandleFunc("DELETE /api/v1/clusters/{clusterID}/stacks/{name}", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleStackRemove)))))

	mux.HandleFunc("POST /api/v1/services/{id}/scale", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceScale)))))
	mux.HandleFunc("POST /api/v1/services/{id}/restart", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceRestart)))))
	mux.HandleFunc("POST /api/v1/services/{id}/update", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceUpdate)))))
	mux.HandleFunc("POST /api/v1/services/{id}/rollback", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceRollback)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/services/{id}/scale", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceScale)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/services/{id}/restart", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceRestart)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/services/{id}/update", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceUpdate)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/services/{id}/rollback", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleServiceRollback)))))

	mux.HandleFunc("POST /api/v1/tasks/{id}/restart", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleTaskRestart)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/tasks/{id}/restart", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.writeMiddleware(d.handleTaskRestart)))))

	mux.HandleFunc("GET /api/v1/actions", d.authMiddleware(d.clusterMiddleware(d.handleActionRunsList)))
	mux.HandleFunc("POST /api/v1/actions/execute", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleActionExecute))))
	mux.HandleFunc("GET /api/v1/approvals", d.authMiddleware(d.clusterMiddleware(d.handleApprovalsList)))
	mux.HandleFunc("POST /api/v1/approvals/{id}/approve", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleApproveAction)))))
	mux.HandleFunc("POST /api/v1/approvals/{id}/reject", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleRejectAction)))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/actions", d.authMiddleware(d.clusterMiddleware(d.handleActionRunsList)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/actions/execute", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleActionExecute))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/approvals", d.authMiddleware(d.clusterMiddleware(d.handleApprovalsList)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/approvals/{id}/approve", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleApproveAction)))))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/approvals/{id}/reject", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.adminMiddleware(d.handleRejectAction)))))

	mux.HandleFunc("GET /api/v1/assistant/sessions", d.authMiddleware(d.clusterMiddleware(d.handleAssistantSessionsList)))
	mux.HandleFunc("POST /api/v1/assistant/sessions", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleAssistantSessionCreate))))
	mux.HandleFunc("GET /api/v1/assistant/sessions/{id}", d.authMiddleware(d.clusterMiddleware(d.handleAssistantSessionGet)))
	mux.HandleFunc("POST /api/v1/assistant/chat", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleAssistantChat))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/assistant/sessions", d.authMiddleware(d.clusterMiddleware(d.handleAssistantSessionsList)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/assistant/sessions", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleAssistantSessionCreate))))
	mux.HandleFunc("GET /api/v1/clusters/{clusterID}/assistant/sessions/{id}", d.authMiddleware(d.clusterMiddleware(d.handleAssistantSessionGet)))
	mux.HandleFunc("POST /api/v1/clusters/{clusterID}/assistant/chat", d.authMiddleware(d.clusterMiddleware(d.csrfMiddleware(d.handleAssistantChat))))
}
