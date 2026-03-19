package httpapi

import (
	"net/http"
	"strings"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// ── Swarm ─────────────────────────────────────────────────────────────────────

func (d *deps) handleSwarm(w http.ResponseWriter, r *http.Request) {
	snap, freshness, lastSync, syncErr := d.snapshotForRequest(r)
	quorum := snap.Managers >= 3 || snap.Managers == 1
	risk := d.cache.GetRisk()
	clusterID := "cluster-unset"
	if len(snap.Nodes) > 0 {
		clusterID = snap.Nodes[0].ID
	}
	writeOK(w, model.SwarmInfo{
		ClusterID:     clusterID,
		CreatedAt:     lastSync,
		UpdatedAt:     lastSync,
		Managers:      snap.Managers,
		Workers:       snap.Workers,
		QuorumHealthy: quorum,
		RaftState:     raftState(snap.Managers),
		Mode:          d.cfg.AppMode,
		WriteEnabled:  d.cfg.WriteActionsEnabled,
		Freshness:     freshness,
		LastSyncAt:    lastSync,
		SyncError:     syncErr,
		Risk:          risk,
	})
}

func raftState(managers int) string {
	switch {
	case managers >= 3:
		return "healthy"
	case managers == 2:
		return "degraded"
	case managers == 1:
		return "single"
	default:
		return "unknown"
	}
}

// ── Nodes ─────────────────────────────────────────────────────────────────────

func (d *deps) handleNodesList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	writeList(w, snap.Nodes, len(snap.Nodes))
}

func (d *deps) handleNodesGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	snap, _, _, _ := d.snapshotForRequest(r)
	for _, n := range snap.Nodes {
		if n.ID == id || n.Hostname == id {
			writeOK(w, n)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "node not found")
}

// ── Stacks ────────────────────────────────────────────────────────────────────

func (d *deps) handleStacksList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	stacks := deriveStacks(snap.Services)
	writeList(w, stacks, len(stacks))
}

func (d *deps) handleStacksGet(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	snap, _, _, _ := d.snapshotForRequest(r)
	stacks := deriveStacks(snap.Services)
	for _, s := range stacks {
		if s.Name == name {
			writeOK(w, s)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "stack not found")
}

func deriveStacks(services []model.Service) []model.Stack {
	type stat struct{ total, running, desired, runDesired int }
	m := make(map[string]*stat)
	for _, svc := range services {
		if svc.Stack == "" {
			continue
		}
		s, ok := m[svc.Stack]
		if !ok {
			s = &stat{}
			m[svc.Stack] = s
		}
		s.total++
		s.desired += svc.DesiredReplicas
		s.runDesired += svc.RunningTasks
		if svc.RunningTasks >= svc.DesiredReplicas && svc.DesiredReplicas > 0 {
			s.running++
		}
	}
	result := make([]model.Stack, 0, len(m))
	for name, s := range m {
		score := 100
		if s.desired > 0 {
			score = s.runDesired * 100 / s.desired
		}
		result = append(result, model.Stack{
			Name:            name,
			ServiceCount:    s.total,
			RunningServices: s.running,
			TotalReplicas:   s.desired,
			RunningReplicas: s.runDesired,
			HealthScore:     score,
		})
	}
	return result
}

// ── Services ──────────────────────────────────────────────────────────────────

func (d *deps) handleServicesList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	stack := r.URL.Query().Get("stack")
	if stack == "" {
		writeList(w, snap.Services, len(snap.Services))
		return
	}
	var filtered []model.Service
	for _, svc := range snap.Services {
		if svc.Stack == stack {
			filtered = append(filtered, svc)
		}
	}
	writeList(w, filtered, len(filtered))
}

func (d *deps) handleServicesGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	snap, _, _, _ := d.snapshotForRequest(r)
	for _, svc := range snap.Services {
		if svc.ID == id || svc.Name == id {
			writeOK(w, svc)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "service not found")
}

// ── Tasks ─────────────────────────────────────────────────────────────────────

func (d *deps) handleTasksList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	q := r.URL.Query()
	svcID := q.Get("service")
	nodeID := q.Get("node")
	stateFilter := q.Get("state")

	tasks := snap.Tasks
	if svcID != "" || nodeID != "" || stateFilter != "" {
		var filtered []model.Task
		for _, t := range tasks {
			if svcID != "" && t.ServiceID != svcID && t.ServiceName != svcID {
				continue
			}
			if nodeID != "" && t.NodeID != nodeID && t.NodeHostname != nodeID {
				continue
			}
			if stateFilter != "" && t.CurrentState != stateFilter {
				continue
			}
			filtered = append(filtered, t)
		}
		tasks = filtered
	}
	writeList(w, tasks, len(tasks))
}

func (d *deps) handleTasksGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	snap, _, _, _ := d.snapshotForRequest(r)
	for _, t := range snap.Tasks {
		if t.ID == id {
			writeOK(w, t)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "task not found")
}

// ── Networks ──────────────────────────────────────────────────────────────────

func (d *deps) handleNetworksList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	writeList(w, snap.Networks, len(snap.Networks))
}

// ── Volumes ───────────────────────────────────────────────────────────────────

func (d *deps) handleVolumesList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	writeList(w, snap.Volumes, len(snap.Volumes))
}

// ── Secrets & Configs ─────────────────────────────────────────────────────────

func (d *deps) handleSecretsList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	writeList(w, snap.Secrets, len(snap.Secrets))
}

func (d *deps) handleConfigsList(w http.ResponseWriter, r *http.Request) {
	snap, _, _, _ := d.snapshotForRequest(r)
	writeList(w, snap.Configs, len(snap.Configs))
}

// ── Events ────────────────────────────────────────────────────────────────────

func (d *deps) handleEventsList(w http.ResponseWriter, r *http.Request) {
	events, _, _, _ := d.eventsForRequest(r)
	typeFilter := r.URL.Query().Get("type")
	if typeFilter != "" {
		var filtered []model.SwarmEvent
		for _, e := range events {
			if strings.EqualFold(e.Type, typeFilter) {
				filtered = append(filtered, e)
			}
		}
		events = filtered
	}
	writeList(w, events, len(events))
}

func (d *deps) handleStreamEvents(w http.ResponseWriter, r *http.Request) {
	d.bus.ServeSSE(w, r)
}
