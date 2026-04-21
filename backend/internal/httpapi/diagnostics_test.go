package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func newDiagnosticsTestDeps(t *testing.T) (*deps, model.Cluster, *clusterRuntime) {
	t.Helper()
	return newTestDeps(t, config.Config{
		AppMode:              "demo",
		DiagnosticsSchedule:  60,
		SnapshotStaleSeconds: 45,
	})
}

func TestHandleDiagnosticsListAutoRunsWhenCacheIsEmpty(t *testing.T) {
	d, cluster, runtime := newDiagnosticsTestDeps(t)
	runtime.diag.set(nil, time.Time{})

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics", nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), clusterKey, cluster), runtimeKey, runtime))
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

	cached, ranAt := runtime.diag.snapshot()
	if len(cached) == 0 {
		t.Fatalf("expected diagnostics cache to be populated")
	}
	if ranAt.IsZero() {
		t.Fatalf("expected non-zero diagnostics last run timestamp")
	}
}

func TestHandleDiagnosticsListAppliesSeverityFilter(t *testing.T) {
	d, cluster, runtime := newDiagnosticsTestDeps(t)
	all := d.runDiagnostics(context.Background(), runtime)
	if len(all) == 0 {
		t.Fatalf("expected at least one finding for filter test")
	}

	targetSeverity := string(all[0].Severity)
	runtime.diag.set(all, time.Now())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics?severity="+targetSeverity, nil)
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), clusterKey, cluster), runtimeKey, runtime))
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
	d, cluster, runtime := newDiagnosticsTestDeps(t)
	runtime.diag.set(d.runDiagnostics(context.Background(), runtime), time.Now())

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/diagnostics/missing", nil)
	req.SetPathValue("id", "missing")
	req = req.WithContext(context.WithValue(context.WithValue(req.Context(), clusterKey, cluster), runtimeKey, runtime))
	d.handleDiagnosticsGet(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rr.Code)
	}
}
