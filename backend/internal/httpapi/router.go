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

// deps holds shared services injected into handlers.
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
	diag      diagnosticsState
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

	// Seed demo data into cache if in demo mode.
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
		middlewareCORS(cfg),
		middlewareSecurityHeaders,
		middlewareRateLimit(cfg),
	), nil
}

func (d *deps) registerRoutes(mux *http.ServeMux) {
	d.registerObsRoutes(mux)
	d.registerInventoryRoutes(mux)
	d.registerDiagnosticsRoutes(mux)
	d.registerIncidentsRoutes(mux)
	d.registerAuditRoutes(mux)
	d.registerActionRoutes(mux)
}

// chain applies middleware outermost-first.
func chain(h http.Handler, mw ...func(http.Handler) http.Handler) http.Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h = mw[i](h)
	}
	return h
}
