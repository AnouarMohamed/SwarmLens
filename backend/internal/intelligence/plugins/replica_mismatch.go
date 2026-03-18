package plugins

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type ReplicaMismatch struct{}

func (p *ReplicaMismatch) Name() string { return "replica-mismatch" }

func (p *ReplicaMismatch) Analyze(snap model.Snapshot) []model.Finding {
	var findings []model.Finding
	for _, svc := range snap.Services {
		if svc.DesiredReplicas == 0 {
			continue
		}
		if svc.RunningTasks < svc.DesiredReplicas {
			sev := model.SeverityHigh
			if svc.RunningTasks == 0 {
				sev = model.SeverityCritical
			}
			findings = append(findings, model.Finding{
				ID:       uuid.NewString(),
				Severity: sev,
				Resource: resourceLabel(svc),
				Scope:    "service",
				Message:  fmt.Sprintf("Service %s has %d/%d running tasks.", svc.Name, svc.RunningTasks, svc.DesiredReplicas),
				Evidence: []string{
					fmt.Sprintf("desired replicas: %d", svc.DesiredReplicas),
					fmt.Sprintf("running tasks: %d", svc.RunningTasks),
					fmt.Sprintf("shortfall: %d", svc.DesiredReplicas-svc.RunningTasks),
				},
				Recommendation: "Check task failure reasons in the Tasks view. Run diagnostics for placement or image pull failures.",
				Source:         p.Name(),
				DetectedAt:     time.Now().UTC(),
			})
		}
	}
	return findings
}
