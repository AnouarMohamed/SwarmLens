// Package intelligence implements the SwarmLens deterministic diagnostic engine.
// It runs all registered plugins against a Snapshot and returns severity-ranked Findings.
package intelligence

import (
	"sort"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Plugin is the interface every diagnostic plugin must implement.
type Plugin interface {
	Name() string
	Analyze(snap model.Snapshot) []model.Finding
}

// Engine runs all registered plugins and returns sorted findings.
type Engine struct {
	plugins []Plugin
}

// New returns an Engine with all v1 plugins registered.
// Plugins are imported and registered in plugins/register.go.
func New(plugins []Plugin) *Engine {
	return &Engine{plugins: plugins}
}

// Run executes all plugins against the snapshot and returns
// all findings sorted: critical first, info last.
func (e *Engine) Run(snap model.Snapshot) []model.Finding {
	var all []model.Finding
	for _, p := range e.plugins {
		all = append(all, p.Analyze(snap)...)
	}
	sort.Slice(all, func(i, j int) bool {
		return severityRank(all[i].Severity) < severityRank(all[j].Severity)
	})
	return all
}

func severityRank(s model.Severity) int {
	switch s {
	case model.SeverityCritical:
		return 0
	case model.SeverityHigh:
		return 1
	case model.SeverityMedium:
		return 2
	case model.SeverityLow:
		return 3
	default:
		return 4
	}
}
