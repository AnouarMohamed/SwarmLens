// Package incident manages the incident lifecycle.
package incident

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Store holds incidents in memory (Phase 1). Phase 3 adds SQLite persistence.
type Store struct {
	mu        sync.RWMutex
	incidents map[string]model.Incident
}

// New creates a Store.
func New() *Store {
	return &Store{incidents: make(map[string]model.Incident)}
}

// Create adds a new incident.
func (s *Store) Create(title, description, severity, createdBy string, affectedServices, diagRefs []string) model.Incident {
	now := time.Now().UTC()
	inc := model.Incident{
		ID:               uuid.NewString(),
		Title:            title,
		Description:      description,
		Severity:         severity,
		Status:           "open",
		CreatedBy:        createdBy,
		CreatedAt:        now,
		UpdatedAt:        now,
		AffectedServices: affectedServices,
		DiagnosticRefs:   diagRefs,
		RunbookSteps:     []model.RunbookStep{},
		Timeline: []model.TimelineEntry{
			{ID: uuid.NewString(), Actor: createdBy, Action: "created", Note: "Incident opened.", Timestamp: now},
		},
	}
	s.mu.Lock()
	s.incidents[inc.ID] = inc
	s.mu.Unlock()
	return inc
}

// Get returns a single incident by ID.
func (s *Store) Get(id string) (model.Incident, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	inc, ok := s.incidents[id]
	return inc, ok
}

// List returns all incidents.
func (s *Store) List() []model.Incident {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]model.Incident, 0, len(s.incidents))
	for _, inc := range s.incidents {
		result = append(result, inc)
	}
	return result
}

// UpdateStatus changes an incident's status and appends a timeline entry.
func (s *Store) UpdateStatus(id, status, actor, note string) (model.Incident, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inc, ok := s.incidents[id]
	if !ok {
		return model.Incident{}, false
	}
	now := time.Now().UTC()
	inc.Status = status
	inc.UpdatedAt = now
	if status == "resolved" {
		inc.ResolvedAt = &now
	}
	inc.Timeline = append(inc.Timeline, model.TimelineEntry{
		ID: uuid.NewString(), Actor: actor, Action: "status_change",
		Note: note, Timestamp: now,
	})
	s.incidents[id] = inc
	return inc, true
}
