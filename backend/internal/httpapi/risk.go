package httpapi

import (
	"context"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type predictorServiceSignal struct {
	Name            string `json:"name"`
	DesiredReplicas int    `json:"desired_replicas"`
	RunningTasks    int    `json:"running_tasks"`
	FailedTasks     int    `json:"failed_tasks"`
	RestartCountMax int    `json:"restart_count_max"`
	UpdateState     string `json:"update_state"`
}

type predictorNodeSignal struct {
	Hostname          string  `json:"hostname"`
	Role              string  `json:"role"`
	Availability      string  `json:"availability"`
	State             string  `json:"state"`
	CPUReservationPct float64 `json:"cpu_reservation_pct"`
	MemReservationPct float64 `json:"mem_reservation_pct"`
}

type predictorPayload struct {
	Managers int                      `json:"managers"`
	Workers  int                      `json:"workers"`
	Services []predictorServiceSignal `json:"services"`
	Nodes    []predictorNodeSignal    `json:"nodes"`
}

func (d *deps) predictRisk(ctx context.Context, snap model.Snapshot) model.RiskAssessment {
	payload := predictorPayload{
		Managers: snap.Managers,
		Workers:  snap.Workers,
		Services: make([]predictorServiceSignal, 0, len(snap.Services)),
		Nodes:    make([]predictorNodeSignal, 0, len(snap.Nodes)),
	}

	restartMaxByService := map[string]int{}
	for _, task := range snap.Tasks {
		if task.RestartCount > restartMaxByService[task.ServiceID] {
			restartMaxByService[task.ServiceID] = task.RestartCount
		}
	}

	for _, svc := range snap.Services {
		payload.Services = append(payload.Services, predictorServiceSignal{
			Name:            svc.Name,
			DesiredReplicas: svc.DesiredReplicas,
			RunningTasks:    svc.RunningTasks,
			FailedTasks:     svc.FailedTasks,
			RestartCountMax: restartMaxByService[svc.ID],
			UpdateState:     svc.UpdateState,
		})
	}
	for _, node := range snap.Nodes {
		cpuPct := 0.0
		memPct := 0.0
		if node.CPUTotal > 0 {
			cpuPct = float64(node.CPUReserved) / float64(node.CPUTotal)
		}
		if node.MemTotal > 0 {
			memPct = float64(node.MemReserved) / float64(node.MemTotal)
		}
		payload.Nodes = append(payload.Nodes, predictorNodeSignal{
			Hostname:          node.Hostname,
			Role:              node.Role,
			Availability:      node.Availability,
			State:             node.State,
			CPUReservationPct: cpuPct,
			MemReservationPct: memPct,
		})
	}

	score := d.predictor.Score(ctx, payload)
	return model.RiskAssessment{
		Score:      score.Score,
		Confidence: score.Confidence,
		Factors:    score.Factors,
		Source:     score.Source,
		UpdatedAt:  time.Now().UTC(),
	}
}
