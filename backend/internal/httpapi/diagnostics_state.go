package httpapi

import (
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type diagnosticsState struct {
	mu       sync.RWMutex
	findings []model.Finding
	lastRun  time.Time
}

func (s *diagnosticsState) snapshot() ([]model.Finding, time.Time) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.findings) == 0 {
		return nil, s.lastRun
	}
	out := make([]model.Finding, len(s.findings))
	copy(out, s.findings)
	return out, s.lastRun
}

func (s *diagnosticsState) set(findings []model.Finding, ranAt time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(findings) == 0 {
		s.findings = nil
		s.lastRun = ranAt
		return
	}
	copied := make([]model.Finding, len(findings))
	copy(copied, findings)
	s.findings = copied
	s.lastRun = ranAt
}
