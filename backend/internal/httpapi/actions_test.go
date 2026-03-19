package httpapi

import (
	"context"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/audit"
	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/incident"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence/plugins"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
	"github.com/AnouarMohamed/swarmlens/backend/internal/state"
	"github.com/AnouarMohamed/swarmlens/backend/internal/stream"
)

func TestExecuteActionReturnsDryRunForMutationInNonDemo(t *testing.T) {
	d := &deps{
		cfg: config.Config{
			AppMode:          "prod",
			LiveActionPolicy: "read_only_dry_run",
		},
		auditLog:  audit.New(100),
		incidents: incident.New(),
		cache:     state.New(),
		engine:    intelligence.New(plugins.Register()),
		bus:       stream.New(),
	}

	outcome := d.executeAction(context.Background(), model.Principal{Username: "alice", Role: model.RoleOperator}, actionRequest{
		Action:     "service.scale",
		Resource:   "service",
		ResourceID: "svc-1",
		Params:     map[string]interface{}{"replicas": 3},
	})

	if outcome.Status != model.ActionStatusDryRun {
		t.Fatalf("expected dry_run, got %s", outcome.Status)
	}
	if outcome.Executed {
		t.Fatalf("expected executed=false")
	}
	if outcome.AuditID == "" {
		t.Fatalf("expected audit record id")
	}
}

func TestExecuteActionRunsDiagnosticsAsLiveReadAction(t *testing.T) {
	d := &deps{
		cfg: config.Config{
			AppMode:          "demo",
			LiveActionPolicy: "read_only_dry_run",
		},
		auditLog:  audit.New(100),
		incidents: incident.New(),
		cache:     state.New(),
		engine:    intelligence.New(plugins.Register()),
		bus:       stream.New(),
	}

	outcome := d.executeAction(context.Background(), model.Principal{Username: "alice", Role: model.RoleOperator}, actionRequest{
		Action: "diagnostics.run",
	})

	if outcome.Status != model.ActionStatusSuccess {
		t.Fatalf("expected success, got %s", outcome.Status)
	}
	if !outcome.Executed {
		t.Fatalf("expected executed=true")
	}
}

func TestExecuteActionReadOnlyPaths(t *testing.T) {
	d := &deps{
		cfg: config.Config{
			AppMode:          "demo",
			LiveActionPolicy: "read_only_dry_run",
		},
		auditLog:  audit.New(100),
		incidents: incident.New(),
		cache:     state.New(),
		engine:    intelligence.New(plugins.Register()),
		bus:       stream.New(),
	}

	tests := []struct {
		name       string
		req        actionRequest
		wantStatus model.ActionStatus
		wantMode   string
		wantRun    bool
	}{
		{
			name:       "telemetry refresh",
			req:        actionRequest{Action: "telemetry.refresh"},
			wantStatus: model.ActionStatusSuccess,
			wantMode:   "live",
			wantRun:    true,
		},
		{
			name: "incident create",
			req: actionRequest{
				Action: "incident.create",
				Params: map[string]interface{}{"title": "Manual incident"},
			},
			wantStatus: model.ActionStatusSuccess,
			wantMode:   "live",
			wantRun:    true,
		},
		{
			name:       "unknown action",
			req:        actionRequest{Action: "unknown.action"},
			wantStatus: model.ActionStatusDryRun,
			wantMode:   "dry_run",
			wantRun:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			outcome := d.executeAction(context.Background(), model.Principal{
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
			if outcome.AuditID == "" {
				t.Fatalf("expected audit id to be populated")
			}
		})
	}
}
