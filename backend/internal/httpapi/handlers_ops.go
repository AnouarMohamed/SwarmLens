package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func (d *deps) handleOpsMetrics(w http.ResponseWriter, r *http.Request) {
	snap, freshness, lastUpdated, _ := d.snapshotForRequest(r)
	history := d.cache.GetHistory()
	if len(history) == 0 {
		d.cache.SetSnapshot(snap)
		history = d.cache.GetHistory()
	}

	writeOK(w, model.OpsMetrics{
		Freshness:   freshness,
		LastUpdated: lastUpdated,
		Series:      history,
		ServiceRisk: d.deriveServiceRisk(snap),
	})
}

func (d *deps) handleOpsInsights(w http.ResponseWriter, r *http.Request) {
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
	insight.SourceStrategy = "deterministic"
	insight.Provider = "none"

	if narrative, provider, ok := d.generateNarrative(r.Context(), snap, findings, risk, insight); ok {
		insight.Summary = narrative
		insight.SourceStrategy = "hybrid"
		insight.Provider = provider
	}

	writeOK(w, insight)
}

func (d *deps) latestFindings() []model.Finding {
	findings, _ := d.diag.snapshot()
	return findings
}

func (d *deps) deriveServiceRisk(snap model.Snapshot) []model.ServiceRisk {
	risks := make([]model.ServiceRisk, 0, len(snap.Services))
	for _, svc := range snap.Services {
		score := 0.05
		reasons := make([]string, 0, 4)
		if svc.DesiredReplicas > 0 && svc.RunningTasks < svc.DesiredReplicas {
			ratio := float64(svc.DesiredReplicas-svc.RunningTasks) / float64(svc.DesiredReplicas)
			score += ratio * 0.5
			reasons = append(reasons, fmt.Sprintf("replica drift %d/%d", svc.RunningTasks, svc.DesiredReplicas))
		}
		if svc.FailedTasks > 0 {
			score += 0.2
			reasons = append(reasons, fmt.Sprintf("%d failed tasks", svc.FailedTasks))
		}
		if strings.Contains(strings.ToLower(svc.UpdateState), "pause") || strings.Contains(strings.ToLower(svc.UpdateState), "rollback") {
			score += 0.15
			reasons = append(reasons, fmt.Sprintf("update state %s", svc.UpdateState))
		}
		if score < 0.1 {
			continue
		}
		if score > 1.0 {
			score = 1.0
		}
		actionability := "monitor"
		if score >= 0.7 {
			actionability = "immediate"
		} else if score >= 0.4 {
			actionability = "soon"
		}
		risks = append(risks, model.ServiceRisk{
			Service:       svc.Name,
			Score:         score,
			Reasons:       reasons,
			Actionability: actionability,
		})
	}
	sort.Slice(risks, func(i, j int) bool {
		return risks[i].Score > risks[j].Score
	})
	if len(risks) > 12 {
		risks = risks[:12]
	}
	return risks
}

func (d *deps) buildDeterministicInsights(
	snap model.Snapshot,
	findings []model.Finding,
	freshness model.FreshnessState,
	risk model.RiskAssessment,
) model.OpsInsights {
	critical := 0
	warning := 0
	for _, f := range findings {
		switch f.Severity {
		case model.SeverityCritical:
			critical++
		case model.SeverityHigh, model.SeverityMedium:
			warning++
		}
	}

	summary := ""
	switch freshness {
	case model.FreshnessDisconnected:
		summary = "SwarmLens is disconnected. Live telemetry and write controls are paused; showing last known posture."
	case model.FreshnessStale:
		summary = fmt.Sprintf("Telemetry is stale. Last known risk score %.2f with %d critical and %d warning findings.", risk.Score, critical, warning)
	default:
		if critical > 0 {
			summary = fmt.Sprintf("Cluster is degraded with %d critical and %d warning findings. Immediate triage is recommended.", critical, warning)
		} else if warning > 0 {
			summary = fmt.Sprintf("Cluster is stable but watch-listed with %d warning findings.", warning)
		} else {
			summary = "Cluster health is nominal. No active critical findings detected in the latest diagnostics run."
		}
	}

	hypotheses := make([]model.InsightHypothesis, 0, 3)
	for _, f := range findings {
		conf := 0.65
		if f.Severity == model.SeverityCritical {
			conf = 0.85
		}
		hypotheses = append(hypotheses, model.InsightHypothesis{
			Title:      f.Message,
			Why:        strings.Join(f.Evidence, "; "),
			Confidence: conf,
		})
		if len(hypotheses) == 3 {
			break
		}
	}
	if len(hypotheses) == 0 {
		hypotheses = append(hypotheses, model.InsightHypothesis{
			Title:      "No immediate root-cause pressure detected",
			Why:        "Current findings set is empty or informational only.",
			Confidence: 0.55,
		})
	}

	actions := []model.InsightAction{
		{
			Title:         "Run diagnostics",
			Description:   "Recompute cluster findings before applying any remediation.",
			EndpointHint:  "/api/v1/diagnostics/run",
			Priority:      1,
			Actionability: "immediate",
		},
		{
			Title:         "Inspect risky services",
			Description:   "Review services with replica drift or failed tasks.",
			EndpointHint:  "/api/v1/services",
			Priority:      2,
			Actionability: "soon",
		},
		{
			Title:         "Review incident timeline",
			Description:   "Open incidents and confirm mitigation ownership.",
			EndpointHint:  "/api/v1/incidents",
			Priority:      3,
			Actionability: "soon",
		},
	}
	if freshness == model.FreshnessDisconnected {
		actions = []model.InsightAction{
			{
				Title:         "Reconnect telemetry",
				Description:   "Restore manager API reachability to resume live updates.",
				EndpointHint:  "/api/v1/actions/execute?action=telemetry.refresh",
				Priority:      1,
				Actionability: "immediate",
			},
			{
				Title:         "Run diagnostics after reconnect",
				Description:   "Validate control plane, scheduling, and replica health on fresh state.",
				EndpointHint:  "/api/v1/diagnostics/run",
				Priority:      2,
				Actionability: "immediate",
			},
		}
	}

	return model.OpsInsights{
		Summary:        summary,
		Risk:           risk,
		Freshness:      freshness,
		Hypotheses:     actionsToHypothesisBias(hypotheses, snap),
		Actions:        actions,
		GeneratedAt:    time.Now().UTC(),
		Provider:       "none",
		SourceStrategy: "deterministic",
	}
}

func actionsToHypothesisBias(h []model.InsightHypothesis, snap model.Snapshot) []model.InsightHypothesis {
	if len(h) > 0 {
		return h
	}
	return []model.InsightHypothesis{
		{
			Title:      "No direct failure signature found",
			Why:        fmt.Sprintf("%d services and %d tasks scanned without high-risk evidence.", len(snap.Services), len(snap.Tasks)),
			Confidence: 0.5,
		},
	}
}

func (d *deps) generateNarrative(
	ctx context.Context,
	snap model.Snapshot,
	findings []model.Finding,
	risk model.RiskAssessment,
	base model.OpsInsights,
) (string, string, bool) {
	if strings.ToLower(d.cfg.AssistantProvider) != "openai" || d.cfg.AssistantAPIKey == "" {
		return "", "", false
	}

	modelName := d.cfg.AssistantModel
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}
	baseURL := strings.TrimSuffix(d.cfg.AssistantBaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}

	type chatMessage struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	payload := map[string]interface{}{
		"model": modelName,
		"messages": []chatMessage{
			{
				Role:    "system",
				Content: "You are an SRE copilot. Produce concise, operational, non-marketing summary text for infrastructure operators.",
			},
			{
				Role: "user",
				Content: fmt.Sprintf(
					"Mode=%s managers=%d workers=%d findings=%d risk=%.2f confidence=%.2f. Base summary: %s",
					d.cfg.AppMode, snap.Managers, snap.Workers, len(findings), risk.Score, risk.Confidence, base.Summary,
				),
			},
		},
		"temperature": 0.2,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", false
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", "", false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.cfg.AssistantAPIKey)

	client := &http.Client{Timeout: 4 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", false
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return "", "", false
	}

	var decoded struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", "", false
	}
	if len(decoded.Choices) == 0 {
		return "", "", false
	}
	out := strings.TrimSpace(decoded.Choices[0].Message.Content)
	if out == "" {
		return "", "", false
	}
	return out, "openai", true
}
