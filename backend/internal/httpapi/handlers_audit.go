package httpapi

import (
	"fmt"
	"net/http"
	"strconv"
)

func (d *deps) handleAuditList(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	entries := d.auditLog.List(limit, offset)
	writeList(w, entries, d.auditLog.Total())
}

func (d *deps) handleAssistantChat(w http.ResponseWriter, r *http.Request) {
	if d.cfg.AssistantProvider == "none" || d.cfg.AssistantProvider == "" {
		writeError(w, http.StatusServiceUnavailable, "assistant_disabled", "assistant provider is not configured (ASSISTANT_PROVIDER=none)")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "sse_unsupported", "SSE not supported")
		return
	}

	snap, _ := d.cache.GetSnapshot()
	findings := d.engine.Run(snap)

	msg := fmt.Sprintf("event: message\ndata: {\"role\":\"assistant\",\"content\":\"Assistant context loaded: %d nodes, %d services, %d active findings. LLM provider: %s (stub response — wire ASSISTANT_API_KEY to enable).\"}\n\n",
		len(snap.Nodes), len(snap.Services), len(findings), d.cfg.AssistantProvider)
	_, _ = fmt.Fprint(w, msg)
	flusher.Flush()
	_, _ = fmt.Fprint(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}
