package httpapi

import (
	"context"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/docker"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence/plugins"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
	"github.com/AnouarMohamed/swarmlens/backend/internal/state"
	"github.com/AnouarMohamed/swarmlens/backend/internal/store"
	"github.com/AnouarMohamed/swarmlens/backend/internal/stream"
)

func newTestDeps(t *testing.T, cfg config.Config) (*deps, model.Cluster, *clusterRuntime) {
	t.Helper()
	dataStore, err := store.New(context.Background(), cfg)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}
	cluster, err := dataStore.SeedDefaultCluster(context.Background(), model.Cluster{
		Name:           "primary",
		DockerHost:     "demo",
		ConnectionMode: model.ClusterConnectionDemo,
		Enabled:        true,
		Default:        true,
	})
	if err != nil {
		t.Fatalf("seed cluster: %v", err)
	}
	demoClient, err := docker.New(config.Config{AppMode: "demo"})
	if err != nil {
		t.Fatalf("new demo docker client: %v", err)
	}
	runtime := &clusterRuntime{
		cluster: cluster,
		docker:  demoClient,
		cache:   state.New(),
		bus:     stream.New(),
	}
	runtime.cache.SetSnapshot(docker.DemoSnapshot())
	runtime.cache.SetEvents(docker.DemoEvents())
	d := &deps{
		cfg:       cfg,
		store:     dataStore,
		engine:    intelligence.New(plugins.Register()),
		runtimes:  map[string]*clusterRuntime{cluster.ID: runtime},
		predictor: nil,
	}
	return d, cluster, runtime
}

func TestExecuteActionReturnsPendingApprovalForLargeScale(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:              "prod",
		LiveActionPolicy:     "allowlist_live",
		ActionSafeScaleDelta: 2,
	})
	runtime.docker = nil
	runtime.cluster.ConnectionMode = model.ClusterConnectionDirect

	outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
		Username: "alice",
		Role:     model.RoleOperator,
	}, actionRequest{
		Action:     "service.scale",
		Resource:   "service",
		ResourceID: "svc-api-01",
		Reason:     "increase headroom for traffic spike",
		Params:     map[string]interface{}{"replicas": 7},
	})

	if outcome.Status != model.ActionStatusPendingApproval {
		t.Fatalf("expected pending approval, got %s", outcome.Status)
	}
	if outcome.ApprovalID == "" {
		t.Fatalf("expected approval id")
	}
}

func TestExecuteActionRunsDiagnosticsAsLiveReadAction(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:          "demo",
		LiveActionPolicy: "read_only_dry_run",
	})

	outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
		Username: "alice",
		Role:     model.RoleOperator,
	}, actionRequest{
		Action: "diagnostics.run",
		Reason: "refresh the cluster baseline",
	})

	if outcome.Status != model.ActionStatusSuccess {
		t.Fatalf("expected success, got %s", outcome.Status)
	}
	if !outcome.Executed {
		t.Fatalf("expected executed=true")
	}
}

func TestExecuteActionReadOnlyAndDryRunPaths(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:          "demo",
		LiveActionPolicy: "read_only_dry_run",
	})

	tests := []struct {
		name       string
		req        actionRequest
		wantStatus model.ActionStatus
		wantMode   string
		wantRun    bool
	}{
		{
			name:       "telemetry refresh",
			req:        actionRequest{Action: "telemetry.refresh", Reason: "refresh control plane data"},
			wantStatus: model.ActionStatusSuccess,
			wantMode:   "live",
			wantRun:    true,
		},
		{
			name: "incident create",
			req: actionRequest{
				Action: "incident.create",
				Reason: "capture operator escalation",
				Params: map[string]interface{}{"title": "Manual incident", "description": "Escalated by operator"},
			},
			wantStatus: model.ActionStatusSuccess,
			wantMode:   "live",
			wantRun:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
				Username: "alice",
				Role:     model.RoleOperator,
			}, tc.req)

			if outcome.Status != tc.wantStatus {
				t.Fatalf("expected status %s, got %s", tc.wantStatus, outcome.Status)
			}
			if outcome.Mode != tc.wantMode {
				t.Fatalf("expected mode %q, got %q", tc.wantMode, outcome.Mode)
			}
			if outcome.Executed != tc.wantRun {
				t.Fatalf("expected executed=%t, got %t", tc.wantRun, outcome.Executed)
			}
		})
	}
}

func TestExecuteActionLeavesNonImplementedMutationInDryRun(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:          "prod",
		LiveActionPolicy: "read_only_dry_run",
	})
	runtime.docker = nil
	runtime.cluster.ConnectionMode = model.ClusterConnectionDirect

	outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
		Username: "alice",
		Role:     model.RoleOperator,
	}, actionRequest{
		Action:   "service.update",
		Reason:   "review dry-run path",
		Params:   map[string]interface{}{"image": "acme/api:v2"},
		Resource: "service",
	})

	if outcome.Status != model.ActionStatusDryRun {
		t.Fatalf("expected dry_run, got %s", outcome.Status)
	}
	if outcome.Mode != "dry_run" {
		t.Fatalf("expected dry_run mode, got %q", outcome.Mode)
	}
}

func TestExecuteActionAllowsSafeScaleWithoutApproval(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:              "demo",
		LiveActionPolicy:     "allowlist_live",
		ActionSafeScaleDelta: 2,
	})

	outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
		Username: "alice",
		Role:     model.RoleOperator,
	}, actionRequest{
		Action:     "service.scale",
		Resource:   "service",
		ResourceID: "svc-api-01",
		Reason:     "add headroom before expected traffic",
		Params:     map[string]interface{}{"replicas": 5},
	})

	if outcome.Status != model.ActionStatusSuccess {
		t.Fatalf("expected success, got %s", outcome.Status)
	}
	if outcome.ApprovalRequired {
		t.Fatalf("expected approval to be skipped for safe delta")
	}
	if outcome.Mode != "demo" {
		t.Fatalf("expected demo mode execution, got %q", outcome.Mode)
	}
}

func TestExecuteActionBlocksMutationWhenReasonMissing(t *testing.T) {
	d, cluster, runtime := newTestDeps(t, config.Config{
		AppMode:          "demo",
		LiveActionPolicy: "allowlist_live",
	})

	outcome := d.executeAction(context.Background(), cluster, runtime, model.Principal{
		Username: "alice",
		Role:     model.RoleOperator,
	}, actionRequest{
		Action:     "service.restart",
		Resource:   "service",
		ResourceID: "svc-api-01",
	})

	if outcome.Status != model.ActionStatusBlocked {
		t.Fatalf("expected blocked, got %s", outcome.Status)
	}
	if outcome.BlockedReason != "reason_required" {
		t.Fatalf("expected reason_required block, got %q", outcome.BlockedReason)
	}
	if outcome.Mode != "blocked" {
		t.Fatalf("expected blocked mode, got %q", outcome.Mode)
	}
}
