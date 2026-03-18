// Package state provides a snapshot cache that holds the most recent
// Swarm cluster state. The inventory layer populates it; other packages read it.
package state

import (
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Cache holds the most recently fetched Swarm snapshot.
type Cache struct {
	mu          sync.RWMutex
	snapshot    model.Snapshot
	events      []model.SwarmEvent
	lastUpdated time.Time
}

// New returns an empty Cache.
func New() *Cache {
	return &Cache{}
}

// SetSnapshot replaces the cached snapshot atomically.
func (c *Cache) SetSnapshot(s model.Snapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.snapshot = s
	c.lastUpdated = time.Now()
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

// LastUpdated returns when the snapshot was last refreshed.
func (c *Cache) LastUpdated() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lastUpdated
}
