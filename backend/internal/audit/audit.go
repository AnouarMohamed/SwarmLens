// Package audit records every mutating action with actor, before/after spec, and result.
package audit

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Store is an in-memory audit log (Phase 1). Phase 2 replaces with SQLite.
type Store struct {
	mu      sync.RWMutex
	entries []model.AuditEntry
	maxSize int
}

// New creates a Store with a bounded size.
func New(maxSize int) *Store {
	if maxSize <= 0 {
		maxSize = 10_000
	}
	return &Store{maxSize: maxSize}
}

// Record appends an audit entry.
func (s *Store) Record(actor, role, action, resource, resourceID string, before, after interface{}, result, reason string) model.AuditEntry {
	entry := model.AuditEntry{
		ID:         uuid.NewString(),
		Actor:      actor,
		Role:       role,
		Action:     action,
		Resource:   resource,
		ResourceID: resourceID,
		BeforeSpec: before,
		AfterSpec:  after,
		Result:     result,
		Reason:     reason,
		Timestamp:  time.Now().UTC(),
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, entry)
	if len(s.entries) > s.maxSize {
		s.entries = s.entries[len(s.entries)-s.maxSize:]
	}
	return entry
}

// List returns entries in reverse-chronological order, with optional pagination.
func (s *Store) List(limit, offset int) []model.AuditEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	reversed := make([]model.AuditEntry, len(s.entries))
	for i, e := range s.entries {
		reversed[len(s.entries)-1-i] = e
	}
	if offset >= len(reversed) {
		return nil
	}
	end := offset + limit
	if end > len(reversed) {
		end = len(reversed)
	}
	return reversed[offset:end]
}

// Total returns the total number of entries.
func (s *Store) Total() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.entries)
}
