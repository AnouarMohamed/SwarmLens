package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

const refreshTTL = 12 * time.Second

func (d *deps) ensureSnapshotFresh(ctx context.Context, runtime *clusterRuntime, force bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if runtime == nil || runtime.docker == nil {
		return nil
	}
	freshness, last, _ := runtime.cache.Status(d.staleAfter())
	if !force && freshness == model.FreshnessLive && !last.IsZero() && time.Since(last) < refreshTTL {
		return nil
	}

	runtime.refreshMu.Lock()
	defer runtime.refreshMu.Unlock()

	// Re-check after waiting for lock.
	freshness, last, _ = runtime.cache.Status(d.staleAfter())
	if !force && freshness == model.FreshnessLive && !last.IsZero() && time.Since(last) < refreshTTL {
		return nil
	}

	refreshCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	snap, events, err := runtime.docker.Snapshot(refreshCtx)
	if err != nil {
		runtime.cache.SetRefreshError(err.Error())
		d.logger.Warn("snapshot refresh failed", "cluster_id", runtime.cluster.ID, "error", err)
		return err
	}
	runtime.cache.SetSnapshot(snap)
	d.refreshCount.Add(1)
	if len(events) > 0 {
		runtime.cache.SetEvents(events)
		latest := events[0]
		runtime.bus.Publish(latest)
	}
	return nil
}

func (d *deps) snapshotForRequest(r *http.Request) (model.Snapshot, model.FreshnessState, time.Time, string) {
	runtime := runtimeFrom(r.Context())
	_ = d.ensureSnapshotFresh(r.Context(), runtime, false)
	snap, updated := runtime.cache.GetSnapshot()
	freshness, _, errMsg := runtime.cache.Status(d.staleAfter())
	return snap, freshness, updated, errMsg
}

func (d *deps) eventsForRequest(r *http.Request) ([]model.SwarmEvent, model.FreshnessState, time.Time, string) {
	runtime := runtimeFrom(r.Context())
	_ = d.ensureSnapshotFresh(r.Context(), runtime, false)
	freshness, updated, errMsg := runtime.cache.Status(d.staleAfter())
	return runtime.cache.GetEvents(), freshness, updated, errMsg
}

func (d *deps) freshnessMessage(f model.FreshnessState, errMsg string) string {
	switch f {
	case model.FreshnessDisconnected:
		if errMsg != "" {
			return fmt.Sprintf("disconnected: %s", errMsg)
		}
		return "disconnected"
	case model.FreshnessStale:
		if errMsg != "" {
			return fmt.Sprintf("stale data: %s", errMsg)
		}
		return "stale data"
	default:
		return "live"
	}
}
