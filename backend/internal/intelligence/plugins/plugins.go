package plugins

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// ── Image Pull Failure ────────────────────────────────────────────────────────

type ImagePullFailure struct{}

func (p *ImagePullFailure) Name() string { return "image-pull-failure" }

func (p *ImagePullFailure) Analyze(snap model.Snapshot) []model.Finding {
	var findings []model.Finding
	for _, t := range snap.Tasks {
		if t.CurrentState != "failed" {
			continue
		}
		e := strings.ToLower(t.Error)
		if !strings.Contains(e, "pull") && !strings.Contains(e, "manifest") &&
			!strings.Contains(e, "unauthorized") && !strings.Contains(e, "not found") &&
			!strings.Contains(e, "no such image") {
			continue
		}
		findings = append(findings, model.Finding{
			ID:       uuid.NewString(),
			Severity: model.SeverityHigh,
			Resource: t.ServiceName,
			Scope:    "service",
			Message:  fmt.Sprintf("Task in service %s failed due to image pull error.", t.ServiceName),
			Evidence: []string{
				fmt.Sprintf("image: %s", t.Image),
				fmt.Sprintf("error: %s", t.Error),
			},
			Recommendation: "Verify the image tag exists in the registry. Check registry credentials (secrets). Ensure worker nodes can reach the registry.",
			Source:         p.Name(),
			DetectedAt:     time.Now().UTC(),
		})
	}
	return findings
}

// ── Port Conflict ─────────────────────────────────────────────────────────────

type PortConflict struct{}

func (p *PortConflict) Name() string { return "port-conflict" }

func (p *PortConflict) Analyze(snap model.Snapshot) []model.Finding {
	portMap := make(map[uint32][]string)
	for _, svc := range snap.Services {
		for _, pp := range svc.PublishedPorts {
			portMap[pp.PublishedPort] = append(portMap[pp.PublishedPort], resourceLabel(svc))
		}
	}
	var findings []model.Finding
	for port, svcs := range portMap {
		if len(svcs) < 2 {
			continue
		}
		findings = append(findings, model.Finding{
			ID:             uuid.NewString(),
			Severity:       model.SeverityHigh,
			Resource:       strings.Join(svcs, ", "),
			Scope:          "service",
			Message:        fmt.Sprintf("Published port %d is claimed by multiple services.", port),
			Evidence:       append([]string{fmt.Sprintf("port: %d", port)}, svcs...),
			Recommendation: fmt.Sprintf("Only one service can publish port %d on the routing mesh. Change the published port for one of the conflicting services.", port),
			Source:         p.Name(),
			DetectedAt:     time.Now().UTC(),
		})
	}
	return findings
}

// ── Secret / Config Reference ─────────────────────────────────────────────────

type SecretConfigRef struct{}

func (p *SecretConfigRef) Name() string { return "secret-config-ref" }

func (p *SecretConfigRef) Analyze(snap model.Snapshot) []model.Finding {
	secretNames := make(map[string]bool)
	for _, s := range snap.Secrets {
		secretNames[s.Name] = true
	}
	configNames := make(map[string]bool)
	for _, c := range snap.Configs {
		configNames[c.Name] = true
	}

	var findings []model.Finding
	for _, svc := range snap.Services {
		for _, ref := range svc.SecretRefs {
			if !secretNames[ref] {
				findings = append(findings, model.Finding{
					ID:             uuid.NewString(),
					Severity:       model.SeverityHigh,
					Resource:       resourceLabel(svc),
					Scope:          "secret",
					Message:        fmt.Sprintf("Service %s references missing secret '%s'.", svc.Name, ref),
					Evidence:       []string{fmt.Sprintf("referenced secret: %s", ref), "secret does not exist in swarm"},
					Recommendation: fmt.Sprintf("Create the secret: docker secret create %s <value>, or remove the reference.", ref),
					Source:         p.Name(),
					DetectedAt:     time.Now().UTC(),
				})
			}
		}
		for _, ref := range svc.ConfigRefs {
			if !configNames[ref] {
				findings = append(findings, model.Finding{
					ID:             uuid.NewString(),
					Severity:       model.SeverityHigh,
					Resource:       resourceLabel(svc),
					Scope:          "config",
					Message:        fmt.Sprintf("Service %s references missing config '%s'.", svc.Name, ref),
					Evidence:       []string{fmt.Sprintf("referenced config: %s", ref), "config does not exist in swarm"},
					Recommendation: fmt.Sprintf("Create the config: docker config create %s <value>, or remove the reference.", ref),
					Source:         p.Name(),
					DetectedAt:     time.Now().UTC(),
				})
			}
		}
	}
	return findings
}

// ── Quorum Risk ───────────────────────────────────────────────────────────────

type QuorumRisk struct{}

func (p *QuorumRisk) Name() string { return "quorum-risk" }

func (p *QuorumRisk) Analyze(snap model.Snapshot) []model.Finding {
	managers := snap.Managers
	if managers == 0 {
		return nil
	}
	faultTolerance := (managers - 1) / 2

	unreachable := 0
	for _, n := range snap.Nodes {
		if n.Role == "manager" && n.ManagerStatus != nil && n.ManagerStatus.Reachability == "unreachable" {
			unreachable++
		}
	}

	var findings []model.Finding

	if managers == 1 {
		findings = append(findings, model.Finding{
			ID: uuid.NewString(), Severity: model.SeverityHigh,
			Resource: "cluster", Scope: "cluster",
			Message:        "Swarm has only 1 manager — no fault tolerance.",
			Evidence:       []string{"managers: 1", "fault tolerance: 0"},
			Recommendation: "Add at least 2 more manager nodes for a 3-manager cluster.",
			Source:         p.Name(), DetectedAt: time.Now().UTC(),
		})
	} else if managers == 2 {
		findings = append(findings, model.Finding{
			ID: uuid.NewString(), Severity: model.SeverityMedium,
			Resource: "cluster", Scope: "cluster",
			Message:        "Swarm has 2 managers — loss of any one loses quorum.",
			Evidence:       []string{"managers: 2", "fault tolerance: 0 (split-brain risk)"},
			Recommendation: "Odd numbers of managers are strongly recommended. Add a third manager.",
			Source:         p.Name(), DetectedAt: time.Now().UTC(),
		})
	}

	if unreachable > 0 && unreachable >= faultTolerance {
		findings = append(findings, model.Finding{
			ID: uuid.NewString(), Severity: model.SeverityCritical,
			Resource: "cluster", Scope: "cluster",
			Message: fmt.Sprintf("%d manager(s) are unreachable — quorum is at risk.", unreachable),
			Evidence: []string{
				fmt.Sprintf("total managers: %d", managers),
				fmt.Sprintf("unreachable managers: %d", unreachable),
				fmt.Sprintf("fault tolerance: %d", faultTolerance),
			},
			Recommendation: "Immediately investigate unreachable manager nodes.",
			Source:         p.Name(), DetectedAt: time.Now().UTC(),
		})
	}
	return findings
}

// ── Update / Rollback State ───────────────────────────────────────────────────

type UpdateRollbackState struct{}

func (p *UpdateRollbackState) Name() string { return "update-rollback-state" }

func (p *UpdateRollbackState) Analyze(snap model.Snapshot) []model.Finding {
	var findings []model.Finding
	for _, svc := range snap.Services {
		switch svc.UpdateState {
		case "paused":
			findings = append(findings, model.Finding{
				ID: uuid.NewString(), Severity: model.SeverityHigh,
				Resource: resourceLabel(svc), Scope: "service",
				Message:        fmt.Sprintf("Service %s update is paused.", svc.Name),
				Evidence:       []string{"update_state: paused", "update may have hit failure threshold"},
				Recommendation: "Run `docker service update --force` to retry, or `docker service rollback` to revert.",
				Source:         p.Name(), DetectedAt: time.Now().UTC(),
			})
		case "rollback_started", "rollback_paused":
			findings = append(findings, model.Finding{
				ID: uuid.NewString(), Severity: model.SeverityHigh,
				Resource: resourceLabel(svc), Scope: "service",
				Message:        fmt.Sprintf("Service %s is in rollback state: %s.", svc.Name, svc.UpdateState),
				Evidence:       []string{fmt.Sprintf("update_state: %s", svc.UpdateState)},
				Recommendation: "Monitor rollback progress. If rollback is also failing, inspect task error logs.",
				Source:         p.Name(), DetectedAt: time.Now().UTC(),
			})
		}
	}
	return findings
}

// ── Node Pressure ─────────────────────────────────────────────────────────────

type NodePressure struct{}

func (p *NodePressure) Name() string { return "node-pressure" }

const pressureThreshold = 0.85

func (p *NodePressure) Analyze(snap model.Snapshot) []model.Finding {
	var findings []model.Finding
	for _, n := range snap.Nodes {
		if n.CPUTotal > 0 {
			ratio := float64(n.CPUReserved) / float64(n.CPUTotal)
			if ratio >= pressureThreshold {
				findings = append(findings, model.Finding{
					ID: uuid.NewString(), Severity: model.SeverityMedium,
					Resource: fmt.Sprintf("node/%s", n.Hostname), Scope: "node",
					Message: fmt.Sprintf("Node %s CPU reservation is at %.0f%%.", n.Hostname, ratio*100),
					Evidence: []string{
						fmt.Sprintf("cpu reserved: %d / %d nanoCPU", n.CPUReserved, n.CPUTotal),
						fmt.Sprintf("reservation ratio: %.1f%%", ratio*100),
					},
					Recommendation: "New tasks with CPU reservations may not schedule here. Consider adding workers.",
					Source:         p.Name(), DetectedAt: time.Now().UTC(),
				})
			}
		}
		if n.MemTotal > 0 {
			ratio := float64(n.MemReserved) / float64(n.MemTotal)
			if ratio >= pressureThreshold {
				findings = append(findings, model.Finding{
					ID: uuid.NewString(), Severity: model.SeverityMedium,
					Resource: fmt.Sprintf("node/%s", n.Hostname), Scope: "node",
					Message: fmt.Sprintf("Node %s memory reservation is at %.0f%%.", n.Hostname, ratio*100),
					Evidence: []string{
						fmt.Sprintf("memory reserved: %d / %d bytes", n.MemReserved, n.MemTotal),
						fmt.Sprintf("reservation ratio: %.1f%%", ratio*100),
					},
					Recommendation: "New tasks with memory reservations may not schedule here. Consider adding workers.",
					Source:         p.Name(), DetectedAt: time.Now().UTC(),
				})
			}
		}
	}
	return findings
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func resourceLabel(svc model.Service) string {
	if svc.Stack != "" {
		return svc.Stack + "/" + svc.Name
	}
	return svc.Name
}
