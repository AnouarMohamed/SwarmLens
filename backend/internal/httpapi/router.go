// Package httpapi wires all HTTP routes and middleware.
package httpapi

import (
	"log/slog"
	"net/http"
	"sync"

	"github.com/AnouarMohamed/swarmlens/backend/internal/audit"
	"github.com/AnouarMohamed/swarmlens/backend/internal/auth"
	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/docker"
	"github.com/AnouarMohamed/swarmlens/backend/internal/incident"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence/plugins"
	"github.com/AnouarMohamed/swarmlens/backend/internal/predictor"
	"github.com/AnouarMohamed/swarmlens/backend/internal/state"
	"github.com/AnouarMohamed/swarmlens/backend/internal/stream"
)

// deps holds all shared services injected into handlers.
type deps struct {
	cfg       config.Config
	logger    *slog.Logger
	docker    *docker.Client
	predictor *predictor.Client
	cache     *state.Cache
	engine    *intelligence.Engine
	authSvc   *auth.Service
	gate      *auth.Gate
	bus       *stream.Bus
	auditLog  *audit.Store
	incidents *incident.Store
	refreshMu sync.Mutex
}

// NewRouter builds and returns the fully-wired HTTP router.
func NewRouter(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	dockerClient, err := docker.New(cfg)
	if err != nil {
		return nil, err
	}

	d := &deps{
		cfg:       cfg,
		logger:    logger,
		docker:    dockerClient,
		predictor: predictor.New(cfg),
		cache:     state.New(),
		engine:    intelligence.New(plugins.Register()),
		authSvc:   auth.New(cfg),
		gate:      auth.NewGate(cfg.WriteActionsEnabled),
		bus:       stream.New(),
		auditLog:  audit.New(10_000),
		incidents: incident.New(),
	}

	// Seed demo data into cache if in demo mode
	if cfg.IsDemo() {
		snap := docker.DemoSnapshot()
		d.cache.SetSnapshot(snap)
		d.cache.SetEvents(docker.DemoEvents())
	}
	_ = d.ensureSnapshotFresh(nil, true)

	mux := http.NewServeMux()
	d.registerRoutes(mux)

	return chain(mux,
		middlewareRecover(logger),
		middlewareLogging(logger),
		middlewareCORS,
		middlewareSecurityHeaders,
		middlewareRateLimit(cfg),
	), nil
}

func (d *deps) registerRoutes(mux *http.ServeMux) {
	// ── Observability ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/healthz", d.handleHealthz)
	mux.HandleFunc("GET /api/v1/readyz", d.handleReadyz)
	mux.HandleFunc("GET /api/v1/metrics", d.handleMetrics)
	mux.HandleFunc("GET /api/v1/metrics/prometheus", d.handleMetricsPrometheus)
	mux.HandleFunc("GET /api/v1/runtime", d.handleRuntime)
	mux.HandleFunc("GET /api/v1/openapi.yaml", d.handleOpenAPI)
	mux.HandleFunc("GET /api/v1/ops/metrics", d.authMiddleware(d.handleOpsMetrics))
	mux.HandleFunc("GET /api/v1/ops/insights", d.authMiddleware(d.handleOpsInsights))

	// ── Swarm ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/swarm", d.authMiddleware(d.handleSwarm))

	// ── Nodes ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/nodes", d.authMiddleware(d.handleNodesList))
	mux.HandleFunc("GET /api/v1/nodes/{id}", d.authMiddleware(d.handleNodesGet))
	mux.HandleFunc("POST /api/v1/nodes/{id}/drain", d.authMiddleware(d.writeMiddleware(d.handleNodeDrain)))
	mux.HandleFunc("POST /api/v1/nodes/{id}/activate", d.authMiddleware(d.writeMiddleware(d.handleNodeActivate)))

	// ── Stacks ────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/stacks", d.authMiddleware(d.handleStacksList))
	mux.HandleFunc("GET /api/v1/stacks/{name}", d.authMiddleware(d.handleStacksGet))
	mux.HandleFunc("POST /api/v1/stacks/{name}/deploy", d.authMiddleware(d.writeMiddleware(d.handleStackDeploy)))
	mux.HandleFunc("DELETE /api/v1/stacks/{name}", d.authMiddleware(d.writeMiddleware(d.handleStackRemove)))

	// ── Services ──────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/services", d.authMiddleware(d.handleServicesList))
	mux.HandleFunc("GET /api/v1/services/{id}", d.authMiddleware(d.handleServicesGet))
	mux.HandleFunc("POST /api/v1/services/{id}/scale", d.authMiddleware(d.writeMiddleware(d.handleServiceScale)))
	mux.HandleFunc("POST /api/v1/services/{id}/restart", d.authMiddleware(d.writeMiddleware(d.handleServiceRestart)))
	mux.HandleFunc("POST /api/v1/services/{id}/update", d.authMiddleware(d.writeMiddleware(d.handleServiceUpdate)))
	mux.HandleFunc("POST /api/v1/services/{id}/rollback", d.authMiddleware(d.writeMiddleware(d.handleServiceRollback)))

	// ── Tasks ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/tasks", d.authMiddleware(d.handleTasksList))
	mux.HandleFunc("GET /api/v1/tasks/{id}", d.authMiddleware(d.handleTasksGet))
	mux.HandleFunc("POST /api/v1/tasks/{id}/restart", d.authMiddleware(d.writeMiddleware(d.handleTaskRestart)))

	// ── Networks ──────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/networks", d.authMiddleware(d.handleNetworksList))

	// ── Volumes ───────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/volumes", d.authMiddleware(d.handleVolumesList))

	// ── Secrets & Configs ─────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/secrets", d.authMiddleware(d.handleSecretsList))
	mux.HandleFunc("GET /api/v1/configs", d.authMiddleware(d.handleConfigsList))

	// ── Events & Stream ───────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/events", d.authMiddleware(d.handleEventsList))
	mux.HandleFunc("GET /api/v1/stream/events", d.authMiddleware(d.handleStreamEvents))

	// ── Diagnostics ───────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/diagnostics", d.authMiddleware(d.handleDiagnosticsList))
	mux.HandleFunc("POST /api/v1/diagnostics/run", d.authMiddleware(d.handleDiagnosticsRun))
	mux.HandleFunc("GET /api/v1/diagnostics/{id}", d.authMiddleware(d.handleDiagnosticsGet))

	// ── Incidents ─────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/incidents", d.authMiddleware(d.handleIncidentsList))
	mux.HandleFunc("POST /api/v1/incidents", d.authMiddleware(d.handleIncidentsCreate))
	mux.HandleFunc("GET /api/v1/incidents/{id}", d.authMiddleware(d.handleIncidentsGet))
	mux.HandleFunc("PUT /api/v1/incidents/{id}", d.authMiddleware(d.handleIncidentsUpdate))
	mux.HandleFunc("POST /api/v1/incidents/{id}/resolve", d.authMiddleware(d.handleIncidentsResolve))

	// ── Audit ─────────────────────────────────────────────────────────────────
	mux.HandleFunc("GET /api/v1/audit", d.authMiddleware(d.handleAuditList))

	// ── Assistant ─────────────────────────────────────────────────────────────
	mux.HandleFunc("POST /api/v1/assistant/chat", d.authMiddleware(d.handleAssistantChat))
	mux.HandleFunc("POST /api/v1/actions/execute", d.authMiddleware(d.handleActionExecute))
}

// chain applies middleware outermost-first.
func chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
