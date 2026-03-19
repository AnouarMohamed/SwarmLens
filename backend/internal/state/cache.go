// Package state provides a snapshot cache that holds the most recent
// Swarm cluster state. The inventory layer populates it; other packages read it.
package state

import (
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

const historyMaxPoints = 180

// Cache holds the most recently fetched Swarm snapshot.
type Cache struct {
	mu          sync.RWMutex
	snapshot    model.Snapshot
	events      []model.SwarmEvent
	lastUpdated time.Time
	lastError   string
	freshness   model.FreshnessState

	criticalFindings int
	warningFindings  int
	risk             model.RiskAssessment
	history          []model.OpsMetricPoint
}

// New returns an empty Cache.
func New() *Cache {
	return &Cache{freshness: model.FreshnessDisconnected}
}

// SetSnapshot replaces the cached snapshot atomically.
func (c *Cache) SetSnapshot(s model.Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = s
	c.lastUpdated = time.Now().UTC()
	c.lastError = ""
	c.freshness = model.FreshnessLive
	c.appendPointLocked(c.lastUpdated)
}

// GetSnapshot returns the current cached snapshot.
func (c *Cache) GetSnapshot() (model.Snapshot, time.Time) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.snapshot, c.lastUpdated
}

// SetEvents replaces the cached event list.
func (c *Cache) SetEvents(events []model.SwarmEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.events = events
}

// GetEvents returns the cached events list.
func (c *Cache) GetEvents() []model.SwarmEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]model.SwarmEvent, len(c.events))
	copy(result, c.events)
	return result
}

// SetFindingsSummary stores latest critical/warning counts.
func (c *Cache) SetFindingsSummary(critical, warning int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.criticalFindings = critical
	c.warningFindings = warning
	if !c.lastUpdated.IsZero() {
		c.appendPointLocked(time.Now().UTC())
	}
}

// GetFindingsSummary returns latest critical/warning findings counts.
func (c *Cache) GetFindingsSummary() (critical, warning int) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.criticalFindings, c.warningFindings
}

// SetRisk stores latest risk assessment.
func (c *Cache) SetRisk(r model.RiskAssessment) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.risk = r
	if !c.lastUpdated.IsZero() {
		c.appendPointLocked(time.Now().UTC())
	}
}

// GetRisk returns latest risk assessment.
func (c *Cache) GetRisk() model.RiskAssessment {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.risk
}

// SetRefreshError marks data as stale/disconnected depending on availability.
func (c *Cache) SetRefreshError(errMsg string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lastError = errMsg
	if c.lastUpdated.IsZero() {
		c.freshness = model.FreshnessDisconnected
		return
	}
	c.freshness = model.FreshnessStale
}

// Status returns freshness, timestamp, and the last refresh error.
func (c *Cache) Status(staleAfter time.Duration) (model.FreshnessState, time.Time, string) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	freshness := c.freshness
	if freshness == model.FreshnessLive && !c.lastUpdated.IsZero() && time.Since(c.lastUpdated) > staleAfter {
		freshness = model.FreshnessStale
	}
	if c.lastUpdated.IsZero() && c.lastError != "" {
		freshness = model.FreshnessDisconnected
	}
	return freshness, c.lastUpdated, c.lastError
}

// GetHistory returns recorded operational history.
func (c *Cache) GetHistory() []model.OpsMetricPoint {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]model.OpsMetricPoint, len(c.history))
	copy(result, c.history)
	return result
}

// LastUpdated returns when the snapshot was last refreshed.
func (c *Cache) LastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdated
}

func (c *Cache) appendPointLocked(ts time.Time) {
	managersOnline := 0
	workersOnline := 0
	healthyReplicas := 0
	desiredReplicas := 0
	failedTasks := 0
	runningTasks := 0
	restarts := 0

	for _, n := range c.snapshot.Nodes {
		ready := n.State == "ready"
		if n.Role == "manager" {
			if ready {
				managersOnline++
			}
			continue
		}
		if ready {
			workersOnline++
		}
	}
	for _, s := range c.snapshot.Services {
		healthyReplicas += s.RunningTasks
		desiredReplicas += s.DesiredReplicas
	}
	for _, t := range c.snapshot.Tasks {
		if t.CurrentState == "running" {
			runningTasks++
		}
		if t.CurrentState == "failed" || t.CurrentState == "rejected" {
			failedTasks++
		}
		restarts += t.RestartCount
	}

	healthyRatio := 1.0
	if desiredReplicas > 0 {
		healthyRatio = float64(healthyReplicas) / float64(desiredReplicas)
	}

	point := model.OpsMetricPoint{
		Timestamp:      ts,
		HealthyRatio:   healthyRatio,
		ManagersOnline: managersOnline,
		WorkersOnline:  workersOnline,
		RunningTasks:   runningTasks,
		FailedTasks:    failedTasks,
		RestartCount:   restarts,
		Critical:       c.criticalFindings,
		Warning:        c.warningFindings,
		RiskScore:      c.risk.Score,
	}
	c.history = append(c.history, point)
	if len(c.history) > historyMaxPoints {
		c.history = c.history[len(c.history)-historyMaxPoints:]
	}
}
