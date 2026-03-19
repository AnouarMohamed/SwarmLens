package predictor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
)

func TestScoreUsesPostAndDecodesPredictorResponse(t *testing.T) {
	t.Helper()
	var gotMethod string
	var gotPath string
	var gotBody map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"score":      0.72,
			"confidence": 0.91,
			"factors":    []string{"replica drift"},
		})
	}))
	defer srv.Close()

	client := New(config.Config{
		PredictorBaseURL: srv.URL,
		PredictorSecret:  "secret",
	})
	score := client.Score(context.Background(), map[string]interface{}{"managers": 1})

	if gotMethod != http.MethodPost {
		t.Fatalf("expected POST, got %s", gotMethod)
	}
	if gotPath != "/score" {
		t.Fatalf("expected /score path, got %s", gotPath)
	}
	if gotBody["managers"] != float64(1) {
		t.Fatalf("expected payload managers=1, got %#v", gotBody["managers"])
	}
	if score.Source != "predictor" {
		t.Fatalf("expected source predictor, got %s", score.Source)
	}
	if score.Score <= 0 {
		t.Fatalf("expected score > 0")
	}
}

func TestScoreFallsBackOnServerError(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := New(config.Config{PredictorBaseURL: srv.URL})
	score := client.Score(context.Background(), map[string]interface{}{"managers": 3})

	if score.Source != "local" {
		t.Fatalf("expected local fallback, got %s", score.Source)
	}
	if len(score.Factors) == 0 {
		t.Fatalf("expected fallback factors")
	}
}
