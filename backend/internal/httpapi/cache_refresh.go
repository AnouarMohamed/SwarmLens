package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

const refreshTTL = 12 * time.Second

func (d *deps) ensureSnapshotFresh(ctx context.Context, force bool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if d.docker == nil {
		return nil
	}
	freshness, last, _ := d.cache.Status(d.staleAfter())
	if !force && freshness == model.FreshnessLive && !last.IsZero() && time.Since(last) < refreshTTL {
		return nil
	}

	d.refreshMu.Lock()
	defer d.refreshMu.Unlock()

	// Re-check after waiting for lock.
	freshness, last, _ = d.cache.Status(d.staleAfter())
	if !force && freshness == model.FreshnessLive && !last.IsZero() && time.Since(last) < refreshTTL {
		return nil
	}

	refreshCtx, cancel := context.WithTimeout(ctx, 6*time.Second)
	defer cancel()

	snap, events, err := d.docker.Snapshot(refreshCtx)
	if err != nil {
		d.cache.SetRefreshError(err.Error())
		d.logger.Warn("snapshot refresh failed", "error", err)
		return err
	}
	d.cache.SetSnapshot(snap)
	if len(events) > 0 {
		d.cache.SetEvents(events)
		latest := events[0]
		d.bus.Publish(latest)
	}
	return nil
}

func (d *deps) staleAfter() time.Duration {
	return time.Duration(d.cfg.SnapshotStaleSeconds) * time.Second
}

func (d *deps) snapshotForRequest(r *http.Request) (model.Snapshot, model.FreshnessState, time.Time, string) {
	_ = d.ensureSnapshotFresh(r.Context(), false)
	snap, updated := d.cache.GetSnapshot()
	freshness, _, errMsg := d.cache.Status(d.staleAfter())
	return snap, freshness, updated, errMsg
}

func (d *deps) eventsForRequest(r *http.Request) ([]model.SwarmEvent, model.FreshnessState, time.Time, string) {
	_ = d.ensureSnapshotFresh(r.Context(), false)
	freshness, updated, errMsg := d.cache.Status(d.staleAfter())
	return d.cache.GetEvents(), freshness, updated, errMsg
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
