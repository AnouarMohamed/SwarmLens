package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/docker"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence"
	"github.com/AnouarMohamed/swarmlens/backend/internal/intelligence/plugins"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
	"github.com/AnouarMohamed/swarmlens/backend/internal/state"
)

func newDiagnosticsTestDeps() *deps {
	d := &deps{
		cfg: config.Config{
			DiagnosticsSchedule:  60,
			SnapshotStaleSeconds: 45,
		},
		cache:  state.New(),
		engine: intelligence.New(plugins.Register()),
	}
	d.cache.SetSnapshot(docker.DemoSnapshot())
	return d
}

func TestHandleDiagnosticsListAutoRunsWhenCacheIsEmpty(t *testing.T) {
	d := newDiagnosticsTestDeps()

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics", nil)
	d.handleDiagnosticsList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Data []model.Finding `json:"data"`
		Meta struct {
			Total int `json:"total"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(resp.Data) == 0 {
		t.Fatalf("expected diagnostics findings from auto-run")
	}
	if resp.Meta.Total != len(resp.Data) {
		t.Fatalf("expected meta.total=%d, got %d", len(resp.Data), resp.Meta.Total)
	}

	cached, ranAt := d.diag.snapshot()
	if len(cached) == 0 {
		t.Fatalf("expected diagnostics cache to be populated")
	}
	if ranAt.IsZero() {
		t.Fatalf("expected non-zero diagnostics last run timestamp")
	}
}

func TestHandleDiagnosticsListAppliesSeverityFilter(t *testing.T) {
	d := newDiagnosticsTestDeps()
	all := d.runDiagnostics()
	if len(all) == 0 {
		t.Fatalf("expected at least one finding for filter test")
	}

	targetSeverity := string(all[0].Severity)
	d.diag.set(all, time.Now())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics?severity="+targetSeverity, nil)
	d.handleDiagnosticsList(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Data []model.Finding `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	for _, finding := range resp.Data {
		if string(finding.Severity) != targetSeverity {
			t.Fatalf("expected severity %q, got %q", targetSeverity, finding.Severity)
		}
	}
}

func TestHandleDiagnosticsGetReturnsNotFoundForUnknownID(t *testing.T) {
	d := newDiagnosticsTestDeps()
	d.diag.set(d.runDiagnostics(), time.Now())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics/missing", nil)
	req.SetPathValue("id", "missing")
	d.handleDiagnosticsGet(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
