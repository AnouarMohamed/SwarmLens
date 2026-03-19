package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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
	var body struct {
		Prompt string `json:"prompt"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	snap, freshness, _, _ := d.snapshotForRequest(r)
	findings := d.latestFindings()
	if len(findings) == 0 {
		findings = d.runDiagnostics()
	}

	risk := d.cache.GetRisk()
	if risk.UpdatedAt.IsZero() {
		risk = d.predictRisk(r.Context(), snap)
		d.cache.SetRisk(risk)
	}

	insight := d.buildDeterministicInsights(snap, findings, freshness, risk)
	if narrative, provider, ok := d.generateNarrative(r.Context(), snap, findings, risk, insight); ok {
		insight.Summary = narrative
		insight.SourceStrategy = "hybrid"
		insight.Provider = provider
	} else {
		insight.SourceStrategy = "deterministic"
		insight.Provider = "none"
	}

	if prompt := strings.TrimSpace(body.Prompt); prompt != "" {
		insight.Summary = fmt.Sprintf("%s | Prompt focus: %s", insight.Summary, prompt)
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "sse_unsupported", "SSE not supported")
		return
	}

	writeSSE(w, "context", map[string]interface{}{
		"nodes":      len(snap.Nodes),
		"services":   len(snap.Services),
		"findings":   len(findings),
		"freshness":  freshness,
		"risk_score": risk.Score,
		"provider":   insight.Provider,
	})
	flusher.Flush()

	writeSSE(w, "insight", insight)
	flusher.Flush()

	for _, h := range insight.Hypotheses {
		writeSSE(w, "hypothesis", h)
	}
	flusher.Flush()

	for _, action := range insight.Actions {
		writeSSE(w, "action", action)
	}
	flusher.Flush()

	writeSSE(w, "message", map[string]string{
		"role":    "assistant",
		"content": insight.Summary,
	})
	writeSSE(w, "done", map[string]bool{"ok": true})
	flusher.Flush()
}

func writeSSE(w http.ResponseWriter, event string, data interface{}) {
	payload, err := json.Marshal(data)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
}
