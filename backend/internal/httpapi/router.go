// Package httpapi wires all HTTP routes and middleware.
package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/AnouarMohamed/swarmlens/backend/internal/auth"
	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence/plugins"
	"github.com/AnouarMohamed/swarmlens/backend/internal/predictor"
	"github.com/AnouarMohamed/swarmlens/backend/internal/store"
)

// deps holds shared services injected into handlers.
type deps struct {
	cfg            config.Config
	logger         *slog.Logger
	predictor      *predictor.Client
	engine         *intelligence.Engine
	authSvc        *auth.Service
	oidcProvider   *auth.OIDCProvider
	gate           *auth.Gate
	store          *store.Store
	runtimesMu     sync.Mutex
	runtimes       map[string]*clusterRuntime
	refreshCount   atomic.Int64
	actionCount    atomic.Int64
	approvalCount  atomic.Int64
	assistantCount atomic.Int64
}

// NewRouter builds and returns the fully-wired HTTP router.
func NewRouter(cfg config.Config, logger *slog.Logger) (http.Handler, error) {
	ctx := context.Background()
	dataStore, err := store.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	oidcProvider, err := auth.NewOIDC(ctx, cfg)
	if err != nil {
		dataStore.Close()
		return nil, err
	}

	d := &deps{
		cfg:          cfg,
		logger:       logger,
		predictor:    predictor.New(cfg),
		engine:       intelligence.New(plugins.Register()),
		authSvc:      auth.New(cfg),
		oidcProvider: oidcProvider,
		gate:         auth.NewGate(cfg.WriteActionsEnabled),
		store:        dataStore,
		runtimes:     map[string]*clusterRuntime{},
	}

	defaultCluster, err := d.seedDefaultCluster(ctx)
	if err != nil {
		dataStore.Close()
		return nil, err
	}
	if _, _, err := d.runtimeForCluster(ctx, defaultCluster.ID); err != nil && cfg.IsDemo() {
		logger.Warn("default cluster runtime unavailable", "cluster_id", defaultCluster.ID, "error", err)
	}

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
	d.registerAuthRoutes(mux)
	d.registerClusterRoutes(mux)
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
