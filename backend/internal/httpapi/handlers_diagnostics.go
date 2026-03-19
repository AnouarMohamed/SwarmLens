package httpapi

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// findingsCache stores the most recent diagnostic run result.
var (
	findingsMu   sync.RWMutex
	lastFindings []model.Finding
	lastRun      time.Time
)

func (d *deps) handleDiagnosticsList(w http.ResponseWriter, r *http.Request) {
	findingsMu.RLock()
	findings := lastFindings
	ran := lastRun
	findingsMu.RUnlock()

	// Auto-run if never run or stale (> schedule interval)
	if findings == nil || time.Since(ran) > time.Duration(d.cfg.DiagnosticsSchedule)*time.Second {
		findings = d.runDiagnostics()
	}

	sev := r.URL.Query().Get("severity")
	if sev != "" {
		var filtered []model.Finding
		for _, f := range findings {
			if string(f.Severity) == sev {
				filtered = append(filtered, f)
			}
		}
		findings = filtered
	}

	writeList(w, findings, len(findings))
}

func (d *deps) handleDiagnosticsRun(w http.ResponseWriter, r *http.Request) {
	findings := d.runDiagnostics()
	writeList(w, findings, len(findings))
}

func (d *deps) handleDiagnosticsGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	findingsMu.RLock()
	findings := lastFindings
	findingsMu.RUnlock()
	for _, f := range findings {
		if f.ID == id {
			writeOK(w, f)
			return
		}
	}
	writeError(w, http.StatusNotFound, "not_found", "finding not found")
}

func (d *deps) runDiagnostics() []model.Finding {
	_ = d.ensureSnapshotFresh(context.Background(), false)
	snap, _ := d.cache.GetSnapshot()
	findings := d.engine.Run(snap)
	critical := 0
	warning := 0
	for _, f := range findings {
		switch strings.ToLower(string(f.Severity)) {
		case "critical":
			critical++
		case "high", "medium":
			warning++
		}
	}
	d.cache.SetFindingsSummary(critical, warning)
	d.cache.SetRisk(d.predictRisk(context.Background(), snap))
	findingsMu.Lock()
	lastFindings = findings
	lastRun = time.Now()
	findingsMu.Unlock()
	return findings
}
