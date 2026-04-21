package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func newClusterRoutingTestDeps(t *testing.T) (*deps, model.Cluster, model.Cluster) {
	t.Helper()

	d, defaultCluster, _ := newTestDeps(t, config.Config{
		AppMode:              "demo",
		DiagnosticsSchedule:  60,
		SnapshotStaleSeconds: 45,
	})

	secondary, err := d.store.SaveCluster(context.Background(), model.Cluster{
		Name:           "secondary",
		DockerHost:     "demo",
		ConnectionMode: model.ClusterConnectionDemo,
		Enabled:        true,
	})
	if err != nil {
		t.Fatalf("save cluster: %v", err)
	}

	return d, defaultCluster, secondary
}

func TestClusterMiddlewareUsesDefaultClusterWhenPathValueMissing(t *testing.T) {
	d, defaultCluster, _ := newClusterRoutingTestDeps(t)

	handler := d.clusterMiddleware(func(w http.ResponseWriter, r *http.Request) {
		cluster := clusterFrom(r.Context())
		runtime := runtimeFrom(r.Context())
		if cluster.ID != defaultCluster.ID {
			t.Fatalf("expected default cluster %q, got %q", defaultCluster.ID, cluster.ID)
		}
		if runtime == nil {
			t.Fatalf("expected runtime in context")
		}
		writeOK(w, map[string]string{"clusterID": cluster.ID})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/swarm", nil)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
}

func TestClusterMiddlewareUsesRequestedCluster(t *testing.T) {
	d, _, secondary := newClusterRoutingTestDeps(t)

	handler := d.clusterMiddleware(func(w http.ResponseWriter, r *http.Request) {
		cluster := clusterFrom(r.Context())
		runtime := runtimeFrom(r.Context())
		if cluster.ID != secondary.ID {
			t.Fatalf("expected cluster %q, got %q", secondary.ID, cluster.ID)
		}
		if runtime == nil {
			t.Fatalf("expected runtime in context")
		}
		writeOK(w, map[string]string{"clusterID": cluster.ID})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clusters/"+secondary.ID+"/swarm", nil)
	req.SetPathValue("clusterID", secondary.ID)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Data["clusterID"] != secondary.ID {
		t.Fatalf("expected response cluster id %q, got %q", secondary.ID, resp.Data["clusterID"])
	}
}

func TestClusterMiddlewareReturnsNotFoundForUnknownCluster(t *testing.T) {
	d, _, _ := newClusterRoutingTestDeps(t)

	handler := d.clusterMiddleware(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not run for missing cluster")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clusters/missing/swarm", nil)
	req.SetPathValue("clusterID", "missing")
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"] != "not_found" {
		t.Fatalf("expected not_found code, got %q", resp["code"])
	}
}

func TestClusterMiddlewareReturnsUnavailableForDisabledCluster(t *testing.T) {
	d, _, secondary := newClusterRoutingTestDeps(t)

	secondary.Enabled = false
	secondary, err := d.store.SaveCluster(context.Background(), secondary)
	if err != nil {
		t.Fatalf("disable cluster: %v", err)
	}

	handler := d.clusterMiddleware(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("next handler should not run for disabled cluster")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/clusters/"+secondary.ID+"/swarm", nil)
	req.SetPathValue("clusterID", secondary.ID)
	rr := httptest.NewRecorder()
	handler(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rr.Code)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["code"] != "cluster_unavailable" {
		t.Fatalf("expected cluster_unavailable code, got %q", resp["code"])
	}
}
