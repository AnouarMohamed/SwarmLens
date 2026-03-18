package plugins

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type PlacementFailure struct{}

func (p *PlacementFailure) Name() string { return "placement-failure" }

func (p *PlacementFailure) Analyze(snap model.Snapshot) []model.Finding {
	var findings []model.Finding
	activeNodes := 0
	for _, n := range snap.Nodes {
		if n.State == "ready" && n.Availability == "active" {
			activeNodes++
		}
	}
	for _, svc := range snap.Services {
		if len(svc.Constraints) == 0 || svc.RunningTasks >= svc.DesiredReplicas {
			continue
		}
		evidence := []string{
			fmt.Sprintf("desired replicas: %d, running: %d", svc.DesiredReplicas, svc.RunningTasks),
			fmt.Sprintf("active nodes in cluster: %d", activeNodes),
		}
		for _, c := range svc.Constraints {
			evidence = append(evidence, fmt.Sprintf("constraint: %s", c))
		}
		findings = append(findings, model.Finding{
			ID:             uuid.NewString(),
			Severity:       model.SeverityHigh,
			Resource:       resourceLabel(svc),
			Scope:          "service",
			Message:        fmt.Sprintf("Service %s has unsatisfied placement constraints.", svc.Name),
			Evidence:       evidence,
			Recommendation: "Verify node labels match the service constraints, or relax the placement constraints.",
			Source:         p.Name(),
			DetectedAt:     time.Now().UTC(),
		})
	}
	return findings
}
