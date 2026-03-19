package httpapi

import (
	"net/http"
	"runtime"
	"time"
)

var startTime = time.Now()

func (d *deps) handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (d *deps) handleReadyz(w http.ResponseWriter, r *http.Request) {
	checks := map[string]string{"api": "ok"}
	if !d.docker.IsDemo() {
		checks["docker"] = "ok"
	} else {
		checks["docker"] = "demo"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "checks": checks})
}

func (d *deps) handleMetrics(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	freshness, lastUpdated, _ := d.cache.Status(d.staleAfter())
	writeOK(w, map[string]interface{}{
		"uptime_seconds": time.Since(startTime).Seconds(),
		"goroutines":     runtime.NumGoroutine(),
		"heap_alloc_mb":  float64(mem.HeapAlloc) / 1024 / 1024,
		"mode":           d.cfg.AppMode,
		"freshness":      freshness,
		"last_sync":      lastUpdated,
	})
}

func (d *deps) handleMetricsPrometheus(w http.ResponseWriter, r *http.Request) {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("# HELP swarmlens_uptime_seconds Seconds since startup\n"))
	_, _ = w.Write([]byte("# TYPE swarmlens_uptime_seconds gauge\n"))
}

func (d *deps) handleRuntime(w http.ResponseWriter, r *http.Request) {
	freshness, lastUpdated, lastErr := d.cache.Status(d.staleAfter())
	writeOK(w, map[string]interface{}{
		"mode":               d.cfg.AppMode,
		"auth_enabled":       d.cfg.AuthEnabled,
		"writes_enabled":     d.cfg.WriteActionsEnabled,
		"live_action_policy": d.cfg.LiveActionPolicy,
		"demo":               d.docker.IsDemo(),
		"freshness":          freshness,
		"last_sync_at":       lastUpdated,
		"last_sync_error":    lastErr,
	})
}

func (d *deps) handleOpenAPI(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "docs/openapi.yaml")
}
