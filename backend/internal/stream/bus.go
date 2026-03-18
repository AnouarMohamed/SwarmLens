// Package stream implements an in-process event bus and SSE broadcaster.
package stream

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

// Bus holds a set of SSE subscriber channels.
type Bus struct {
	mu   sync.RWMutex
	subs map[chan model.SwarmEvent]struct{}
}

// New creates a Bus.
func New() *Bus {
	return &Bus{subs: make(map[chan model.SwarmEvent]struct{})}
}

// Publish sends an event to all subscribers.
func (b *Bus) Publish(evt model.SwarmEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- evt:
		default: // drop if subscriber is slow
		}
	}
}

// subscribe adds a subscriber and returns an unsubscribe func.
func (b *Bus) subscribe() (chan model.SwarmEvent, func()) {
	ch := make(chan model.SwarmEvent, 32)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.subs, ch)
		b.mu.Unlock()
		close(ch)
	}
}

// ServeSSE writes a Server-Sent Events stream to the response writer.
// Blocks until the client disconnects.
func (b *Bus) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch, unsub := b.subscribe()
	defer unsub()

	_, _ = fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
	flusher.Flush()

	for {
		select {
		case <-r.Context().Done():
			return
		case evt, ok := <-ch:
			if !ok {
				return
			}
			data, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: swarm\ndata: %s\n\n", data)
			flusher.Flush()
		}
	}
}
