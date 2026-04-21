package httpapi

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/docker"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
	"github.com/AnouarMohamed/swarmlens/backend/internal/state"
	"github.com/AnouarMohamed/swarmlens/backend/internal/store"
	"github.com/AnouarMohamed/swarmlens/backend/internal/stream"
)

const (
	clusterKey    contextKey = "cluster"
	runtimeKey    contextKey = "cluster_runtime"
	sessionKey    contextKey = "auth_session"
	authMethodKey contextKey = "auth_method"
)

type clusterRuntime struct {
	cluster   model.Cluster
	docker    *docker.Client
	cache     *state.Cache
	bus       *stream.Bus
	diag      diagnosticsState
	refreshMu sync.Mutex
}

func clusterFrom(ctx context.Context) model.Cluster {
	cluster, _ := ctx.Value(clusterKey).(model.Cluster)
	return cluster
}

func runtimeFrom(ctx context.Context) *clusterRuntime {
	runtime, _ := ctx.Value(runtimeKey).(*clusterRuntime)
	return runtime
}

func sessionFrom(ctx context.Context) (model.AuthSession, bool) {
	session, ok := ctx.Value(sessionKey).(model.AuthSession)
	return session, ok
}

func authMethodFrom(ctx context.Context) string {
	method, _ := ctx.Value(authMethodKey).(string)
	return method
}

func (d *deps) seedDefaultCluster(ctx context.Context) (model.Cluster, error) {
	cluster := model.Cluster{
		Name:           d.cfg.DefaultClusterName,
		DockerHost:     d.cfg.DockerHost,
		ConnectionMode: model.ClusterConnectionDirect,
		TLSEnabled:     d.cfg.DockerTLSVerify,
		CertRef:        d.cfg.DockerCertPath,
		Enabled:        true,
		Default:        true,
	}
	if d.cfg.IsDemo() {
		cluster.ConnectionMode = model.ClusterConnectionDemo
		cluster.DockerHost = "demo"
		cluster.TLSEnabled = false
		cluster.CertRef = ""
	}
	return d.store.SeedDefaultCluster(ctx, cluster)
}

func (d *deps) clusterMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clusterID := r.PathValue("clusterID")
		ctx := r.Context()
		var cluster model.Cluster
		var runtime *clusterRuntime
		var err error
		if strings.TrimSpace(clusterID) == "" {
			cluster, runtime, err = d.defaultRuntime(ctx)
		} else {
			cluster, runtime, err = d.runtimeForCluster(ctx, clusterID)
		}
		if err != nil {
			status := http.StatusInternalServerError
			code := "cluster_unavailable"
			if errors.Is(err, store.ErrNotFound) {
				status = http.StatusNotFound
				code = "not_found"
			}
			writeError(w, status, code, err.Error())
			return
		}
		ctx = context.WithValue(ctx, clusterKey, cluster)
		ctx = context.WithValue(ctx, runtimeKey, runtime)
		next(w, r.WithContext(ctx))
	}
}

func (d *deps) defaultRuntime(ctx context.Context) (model.Cluster, *clusterRuntime, error) {
	cluster, err := d.store.GetDefaultCluster(ctx)
	if err != nil {
		return model.Cluster{}, nil, fmt.Errorf("default cluster: %w", err)
	}
	return d.runtimeForCluster(ctx, cluster.ID)
}

func (d *deps) runtimeForCluster(ctx context.Context, clusterID string) (model.Cluster, *clusterRuntime, error) {
	cluster, err := d.store.GetCluster(ctx, clusterID)
	if err != nil {
		return model.Cluster{}, nil, err
	}
	if !cluster.Enabled {
		return model.Cluster{}, nil, fmt.Errorf("cluster %s is disabled", cluster.Name)
	}

	d.runtimesMu.Lock()
	defer d.runtimesMu.Unlock()

	if runtime, ok := d.runtimes[cluster.ID]; ok && runtime.cluster.UpdatedAt.Equal(cluster.UpdatedAt) {
		cluster.Health = d.clusterHealth(runtime)
		return cluster, runtime, nil
	}

	client, err := docker.NewForCluster(d.cfg, cluster)
	if err != nil {
		return model.Cluster{}, nil, err
	}
	runtime := &clusterRuntime{
		cluster: cluster,
		docker:  client,
		cache:   state.New(),
		bus:     stream.New(),
	}
	if cluster.ConnectionMode == model.ClusterConnectionDemo {
		runtime.cache.SetSnapshot(docker.DemoSnapshot())
		runtime.cache.SetEvents(docker.DemoEvents())
	}
	d.runtimes[cluster.ID] = runtime
	cluster.Health = d.clusterHealth(runtime)
	return cluster, runtime, nil
}

func (d *deps) clusterHealth(runtime *clusterRuntime) model.ClusterHealth {
	if runtime == nil {
		return model.ClusterHealth{Freshness: model.FreshnessDisconnected}
	}
	snap, lastSync := runtime.cache.GetSnapshot()
	freshness, _, errMsg := runtime.cache.Status(d.staleAfter())
	return model.ClusterHealth{
		Freshness:     freshness,
		LastSyncAt:    lastSync,
		LastSyncError: errMsg,
		Managers:      snap.Managers,
		Workers:       snap.Workers,
		Reachable:     freshness != model.FreshnessDisconnected,
	}
}

func (d *deps) clustersWithHealth(ctx context.Context) ([]model.Cluster, error) {
	clusters, err := d.store.ListClusters(ctx)
	if err != nil {
		return nil, err
	}
	for idx := range clusters {
		_, runtime, runtimeErr := d.runtimeForCluster(ctx, clusters[idx].ID)
		if runtimeErr != nil {
			clusters[idx].Health = model.ClusterHealth{
				Freshness:     model.FreshnessDisconnected,
				LastSyncError: runtimeErr.Error(),
			}
			continue
		}
		clusters[idx].Health = d.clusterHealth(runtime)
	}
	return clusters, nil
}

func (d *deps) staleAfter() time.Duration {
	return time.Duration(d.cfg.SnapshotStaleSeconds) * time.Second
}
