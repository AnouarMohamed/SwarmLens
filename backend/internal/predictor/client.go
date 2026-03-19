// Package predictor wraps the optional FastAPI predictor service.
// If the predictor is unavailable, all calls fall back to deterministic local scoring.
package predictor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
)

// RiskScore is the response from the predictor.
type RiskScore struct {
	Score      float64  `json:"score"`
	Confidence float64  `json:"confidence"`
	Factors    []string `json:"factors,omitempty"`
	Source     string   `json:"source"` // "predictor" or "local"
}

// Client calls the predictor service with a shared-secret header.
type Client struct {
	baseURL string
	secret  string
	http    *http.Client
}

// New creates a predictor Client. Returns nil if no base URL is configured.
func New(cfg config.Config) *Client {
	if cfg.PredictorBaseURL == "" {
		return nil
	}
	return &Client{
		baseURL: cfg.PredictorBaseURL,
		secret:  cfg.PredictorSecret,
		http:    &http.Client{Timeout: 3 * time.Second},
	}
}

// Score calls the predictor and returns a risk score.
// Falls back to local scoring on any error.
func (c *Client) Score(ctx context.Context, payload interface{}) RiskScore {
	if c == nil {
		return localScore(payload)
	}
	url := fmt.Sprintf("%s/score", c.baseURL)
	body, err := json.Marshal(payload)
	if err != nil {
		return localScore(payload)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return localScore(payload)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shared-Secret", c.secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return localScore(payload)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= http.StatusBadRequest {
		return localScore(payload)
	}
	var score RiskScore
	if err := json.NewDecoder(resp.Body).Decode(&score); err != nil {
		return localScore(payload)
	}
	score.Source = "predictor"
	return score
}

// localScore provides a basic deterministic fallback score.
func localScore(_ interface{}) RiskScore {
	return RiskScore{
		Score:      0.5,
		Confidence: 0.3,
		Factors:    []string{"predictor unavailable; using deterministic local baseline"},
		Source:     "local",
	}
}
