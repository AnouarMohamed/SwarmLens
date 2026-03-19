package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
)

func TestMiddlewareCORSEnforcesConfiguredAllowlist(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := middlewareCORS(config.Config{CORSAllowOrigins: "https://ops.example.com,https://grafana.example.com"})(next)

	tests := []struct {
		name            string
		origin          string
		wantAllowOrigin string
	}{
		{name: "allowed origin", origin: "https://ops.example.com", wantAllowOrigin: "https://ops.example.com"},
		{name: "disallowed origin", origin: "https://evil.example.com", wantAllowOrigin: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Origin", tc.origin)
			rr := httptest.NewRecorder()

			h.ServeHTTP(rr, req)

			if got := rr.Header().Get("Access-Control-Allow-Origin"); got != tc.wantAllowOrigin {
				t.Fatalf("expected Allow-Origin %q, got %q", tc.wantAllowOrigin, got)
			}
		})
	}
}

func TestMiddlewareCORSDefaultMirrorsRequestOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	h := middlewareCORS(config.Config{})(next)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "https://app.example.com" {
		t.Fatalf("expected mirrored origin, got %q", got)
	}
}

func TestMiddlewareCORSHandlesPreflight(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	h := middlewareCORS(config.Config{})(next)
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "https://app.example.com")
	rr := httptest.NewRecorder()

	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for preflight, got %d", rr.Code)
	}
	if called {
		t.Fatalf("expected preflight request to skip downstream handler")
	}
}
