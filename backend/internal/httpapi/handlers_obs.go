package httpapi

import (
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

var startTime = time.Now()

func (d *deps) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (d *deps) handleReadyz(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{"api": "ok"}
	if err := d.store.Ready(r.Context()); err != nil {
		checks["db"] = "error"
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"status": "degraded", "checks": checks, "error": err.Error()})
		return
	}
	if d.store.IsPersistent() {
		checks["db"] = "ok"
	} else {
		checks["db"] = "memory"
	}
	cluster, runtime, err := d.defaultRuntime(r.Context())
	if err != nil {
		checks["cluster"] = "error"
		writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{"status": "degraded", "checks": checks, "error": err.Error()})
		return
	}
	if runtime.docker.IsDemo() {
		checks["cluster"] = "demo"
	} else {
		checks["cluster"] = "ok"
	}
	checks["default_cluster"] = cluster.Name
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "checks": checks})
}

func (d *deps) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	cluster, runtimeState, err := d.defaultRuntime(r.Context())
	freshness := model.FreshnessDisconnected
	var lastUpdated time.Time
	if err == nil {
		freshness, lastUpdated, _ = runtimeState.cache.Status(d.staleAfter())
	}
	writeOK(w, map[string]interface{}{
		"uptime_seconds":    time.Since(startTime).Seconds(),
		"goroutines":        runtime.NumGoroutine(),
		"heap_alloc_mb":     float64(mem.HeapAlloc) / 1024 / 1024,
		"mode":              d.cfg.AppMode,
		"default_cluster":   cluster.ID,
		"freshness":         freshness,
		"last_sync":         lastUpdated,
		"db_persistent":     d.store.IsPersistent(),
		"cluster_refreshes": d.refreshCount.Load(),
		"actions_total":     d.actionCount.Load(),
		"approvals_total":   d.approvalCount.Load(),
		"assistant_total":   d.assistantCount.Load(),
	})
}

func (d *deps) handleMetricsPrometheus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("# HELP swarmlens_uptime_seconds Seconds since startup\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_uptime_seconds gauge\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("swarmlens_uptime_seconds %f\n", time.Since(startTime).Seconds())))
	_, _ = w.Write([]byte("# HELP swarmlens_cluster_refresh_total Snapshot refreshes across clusters\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_cluster_refresh_total counter\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("swarmlens_cluster_refresh_total %d\n", d.refreshCount.Load())))
	_, _ = w.Write([]byte("# HELP swarmlens_action_run_total Action pipeline executions\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_action_run_total counter\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("swarmlens_action_run_total %d\n", d.actionCount.Load())))
	_, _ = w.Write([]byte("# HELP swarmlens_approval_total Approval resolutions\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_approval_total counter\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("swarmlens_approval_total %d\n", d.approvalCount.Load())))
	_, _ = w.Write([]byte("# HELP swarmlens_assistant_runs_total Assistant requests served\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_assistant_runs_total counter\n"))
	_, _ = w.Write([]byte(fmt.Sprintf("swarmlens_assistant_runs_total %d\n", d.assistantCount.Load())))
}

func (d *deps) handleRuntime(w http.ResponseWriter, r *http.Request) {
	cluster, runtimeState, err := d.defaultRuntime(r.Context())
	freshness := model.FreshnessDisconnected
	var lastUpdated time.Time
	var lastErr string
	if err == nil {
		freshness, lastUpdated, lastErr = runtimeState.cache.Status(d.staleAfter())
	}
	writeOK(w, map[string]interface{}{
		"mode":               d.cfg.AppMode,
		"auth_enabled":       d.cfg.AuthEnabled,
		"auth_provider":      d.cfg.AuthProvider,
		"writes_enabled":     d.cfg.WriteActionsEnabled,
		"live_action_policy": d.cfg.LiveActionPolicy,
		"default_cluster":    cluster,
		"db_persistent":      d.store.IsPersistent(),
		"freshness":          freshness,
		"last_sync_at":       lastUpdated,
		"last_sync_error":    lastErr,
	})
}

func (d *deps) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "docs/openapi.yaml")
}
