package httpapi

import "net/http"

func (d *deps) registerObsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/healthz", d.handleHealthz)
	mux.HandleFunc("GET /api/v1/readyz", d.handleReadyz)
	mux.HandleFunc("GET /api/v1/metrics", d.handleMetrics)
	mux.HandleFunc("GET /api/v1/metrics/prometheus", d.handleMetricsPrometheus)
	mux.HandleFunc("GET /api/v1/runtime", d.handleRuntime)
	mux.HandleFunc("GET /api/v1/openapi.yaml", d.handleOpenAPI)
	mux.HandleFunc("GET /api/v1/ops/metrics", d.authMiddleware(d.handleOpsMetrics))
	mux.HandleFunc("GET /api/v1/ops/insights", d.authMiddleware(d.handleOpsInsights))
}
