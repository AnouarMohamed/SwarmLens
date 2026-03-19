package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func (d *deps) handleDiagnosticsList(w http.ResponseWriter, r *http.Request) {
	findings, ranAt := d.diag.snapshot()

	// Auto-run if never run or stale (> schedule interval).
	if len(findings) == 0 || time.Since(ranAt) > time.Duration(d.cfg.DiagnosticsSchedule)*time.Second {
		findings = d.runDiagnostics()
	}

	filtered := filterFindingsBySeverity(findings, r.URL.Query().Get("severity"))
	writeList(w, filtered, len(filtered))
}

func (d *deps) handleDiagnosticsRun(w http.ResponseWriter, r *http.Request) {
	findings := d.runDiagnostics()
	writeList(w, findings, len(findings))
}

func (d *deps) handleDiagnosticsGet(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	findings, _ := d.diag.snapshot()
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

	critical, warning := summarizeFindings(findings)
	d.cache.SetFindingsSummary(critical, warning)
	d.cache.SetRisk(d.predictRisk(context.Background(), snap))
	d.diag.set(findings, time.Now())

	return findings
}

func filterFindingsBySeverity(findings []model.Finding, severity string) []model.Finding {
	if severity == "" {
		return findings
	}

	filtered := make([]model.Finding, 0, len(findings))
	for _, f := range findings {
		if string(f.Severity) == severity {
			filtered = append(filtered, f)
		}
	}
	return filtered
}

func summarizeFindings(findings []model.Finding) (critical int, warning int) {
	for _, f := range findings {
		switch strings.ToLower(string(f.Severity)) {
		case "critical":
			critical++
		case "high", "medium":
			warning++
		}
	}
	return critical, warning
}
