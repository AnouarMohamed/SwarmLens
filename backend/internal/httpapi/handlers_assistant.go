package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func (d *deps) handleAssistantSessionsList(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	sessions, err := d.store.ListAssistantSessions(r.Context(), cluster.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "assistant_session_list_failed", err.Error())
		return
	}
	writeList(w, sessions, len(sessions))
}

func (d *deps) handleAssistantSessionCreate(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	p := principalFrom(r.Context())
	var body struct {
		Title      string `json:"title"`
		IncidentID string `json:"incidentID"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	title := strings.TrimSpace(body.Title)
	if title == "" {
		title = "Ops Copilot Session"
	}
	session, err := d.store.CreateAssistantSession(r.Context(), model.AssistantSession{
		ClusterID:  cluster.ID,
		IncidentID: strings.TrimSpace(body.IncidentID),
		Title:      title,
		CreatedBy:  p.Username,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "assistant_session_create_failed", err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"data": session})
}

func (d *deps) handleAssistantSessionGet(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	session, err := d.store.GetAssistantSession(r.Context(), cluster.ID, r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "not_found", "assistant session not found")
		return
	}
	writeOK(w, session)
}

func (d *deps) handleAssistantChat(w http.ResponseWriter, r *http.Request) {
	cluster := clusterFrom(r.Context())
	runtime := runtimeFrom(r.Context())
	p := principalFrom(r.Context())

	var body struct {
		Prompt     string `json:"prompt"`
		SessionID  string `json:"sessionID"`
		IncidentID string `json:"incidentID"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	prompt := strings.TrimSpace(body.Prompt)

	sessionID := strings.TrimSpace(body.SessionID)
	var session model.AssistantSession
	var err error
	if sessionID != "" {
		session, err = d.store.GetAssistantSession(r.Context(), cluster.ID, sessionID)
	}
	if session.ID == "" || err != nil {
		title := "Ops Copilot Session"
		if prompt != "" {
			title = truncate(prompt, 72)
		}
		session, err = d.store.CreateAssistantSession(r.Context(), model.AssistantSession{
			ClusterID:  cluster.ID,
			IncidentID: strings.TrimSpace(body.IncidentID),
			Title:      title,
			CreatedBy:  p.Username,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "assistant_session_create_failed", err.Error())
			return
		}
	}

	if prompt != "" {
		_, _ = d.store.AppendAssistantMessage(r.Context(), session.ID, model.AssistantMessage{
			Role:      "user",
			Content:   prompt,
			CreatedAt: time.Now().UTC(),
		}, session.LastSummary)
	}

	snap, freshness, _, _ := d.snapshotForRequest(r)
	findings := d.latestFindings(runtime)
	if len(findings) == 0 {
		findings = d.runDiagnostics(r.Context(), runtime)
	}

	risk := runtime.cache.GetRisk()
	if risk.UpdatedAt.IsZero() {
		risk = d.predictRisk(r.Context(), snap)
		runtime.cache.SetRisk(risk)
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

	incidents, _ := d.store.ListIncidents(r.Context(), cluster.ID)
	actions, _ := d.store.ListActionRuns(r.Context(), cluster.ID, 6)
	citations := d.buildAssistantCitations(findings, incidents, actions)
	proposals := d.buildAssistantActionProposals(snap)
	if prompt != "" {
		insight.Summary = fmt.Sprintf("%s\n\nPrompt focus: %s", insight.Summary, prompt)
	}

	assistantMessage := model.AssistantMessage{
		Role:            "assistant",
		Content:         insight.Summary,
		Citations:       citations,
		ActionProposals: proposals,
		CreatedAt:       time.Now().UTC(),
	}
	assistantMessage, err = d.store.AppendAssistantMessage(r.Context(), session.ID, assistantMessage, insight.Summary)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "assistant_message_store_failed", err.Error())
		return
	}
	d.assistantCount.Add(1)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "sse_unsupported", "SSE not supported")
		return
	}

	writeSSE(w, "session", session)
	writeSSE(w, "context", map[string]any{
		"clusterID":   cluster.ID,
		"clusterName": cluster.Name,
		"nodes":       len(snap.Nodes),
		"services":    len(snap.Services),
		"findings":    len(findings),
		"freshness":   freshness,
		"riskScore":   risk.Score,
		"provider":    insight.Provider,
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
	for _, citation := range citations {
		writeSSE(w, "citation", citation)
	}
	for _, proposal := range proposals {
		writeSSE(w, "action_proposal", proposal)
	}
	flusher.Flush()

	writeSSE(w, "message", assistantMessage)
	writeSSE(w, "done", map[string]bool{"ok": true})
	flusher.Flush()
}

func (d *deps) buildAssistantCitations(findings []model.Finding, incidents []model.Incident, actions []model.ActionRun) []model.AssistantCitation {
	var citations []model.AssistantCitation
	for _, finding := range findings {
		citations = append(citations, model.AssistantCitation{
			ID:      finding.ID,
			Kind:    "finding",
			Title:   finding.Message,
			Locator: "/diagnostics/" + finding.ID,
			Snippet: strings.Join(finding.Evidence, "; "),
		})
		if len(citations) == 3 {
			break
		}
	}
	for _, incident := range incidents {
		citations = append(citations, model.AssistantCitation{
			ID:      incident.ID,
			Kind:    "incident",
			Title:   incident.Title,
			Locator: "/incidents/" + incident.ID,
			Snippet: incident.Status,
		})
		if len(citations) == 5 {
			break
		}
	}
	for _, action := range actions {
		citations = append(citations, model.AssistantCitation{
			ID:      action.ID,
			Kind:    "action_run",
			Title:   action.Action,
			Locator: "/actions/" + action.ID,
			Snippet: string(action.Status) + " · " + action.Message,
		})
		if len(citations) == 7 {
			break
		}
	}
	return citations
}

func (d *deps) buildAssistantActionProposals(snap model.Snapshot) []model.AssistantActionProposal {
	proposals := []model.AssistantActionProposal{
		{
			Title:            "Refresh telemetry",
			Action:           "telemetry.refresh",
			Reason:           "Refresh cluster state before remediation decisions.",
			RequiresApproval: false,
		},
		{
			Title:            "Run diagnostics",
			Action:           "diagnostics.run",
			Reason:           "Regenerate findings and refresh the deterministic baseline.",
			RequiresApproval: false,
		},
	}
	for _, service := range snap.Services {
		if service.FailedTasks > 0 {
			proposals = append(proposals, model.AssistantActionProposal{
				Title:            "Restart degraded service",
				Action:           "service.restart",
				Resource:         "service",
				ResourceID:       service.ID,
				Reason:           fmt.Sprintf("%s has %d failed task(s).", service.Name, service.FailedTasks),
				RequiresApproval: false,
			})
			break
		}
		if strings.Contains(strings.ToLower(service.UpdateState), "rollback") {
			proposals = append(proposals, model.AssistantActionProposal{
				Title:            "Prepare rollback review",
				Action:           "service.rollback",
				Resource:         "service",
				ResourceID:       service.ID,
				Reason:           fmt.Sprintf("%s is in update state %s.", service.Name, service.UpdateState),
				RequiresApproval: true,
			})
			break
		}
	}
	if len(proposals) > 4 {
		proposals = proposals[:4]
	}
	return proposals
}

func truncate(value string, size int) string {
	value = strings.TrimSpace(value)
	if len(value) <= size {
		return value
	}
	return strings.TrimSpace(value[:size]) + "..."
}
