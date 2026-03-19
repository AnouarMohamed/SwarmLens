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
