package httpapi

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
)

func TestNewRouterRegistersCoreRoutes(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	h, err := NewRouter(config.Config{
		AppMode:              "demo",
		DiagnosticsSchedule:  60,
		SnapshotStaleSeconds: 45,
		LiveActionPolicy:     "read_only_dry_run",
	}, logger)
	if err != nil {
		t.Fatalf("expected router initialization, got error: %v", err)
	}

	tests := []struct {
		name   string
		method string
		path   string
		body   []byte
		want   int
	}{
		{name: "healthz", method: http.MethodGet, path: "/api/v1/healthz", want: http.StatusOK},
		{name: "swarm", method: http.MethodGet, path: "/api/v1/swarm", want: http.StatusOK},
		{name: "diagnostics", method: http.MethodGet, path: "/api/v1/diagnostics", want: http.StatusOK},
		{name: "audit", method: http.MethodGet, path: "/api/v1/audit", want: http.StatusOK},
		{
			name:   "actions execute",
			method: http.MethodPost,
			path:   "/api/v1/actions/execute",
			body:   []byte(`{"action":"diagnostics.run"}`),
			want:   http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader(tc.body))
			if tc.method == http.MethodPost {
				req.Header.Set("Content-Type", "application/json")
			}
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tc.want {
				t.Fatalf("%s %s expected %d, got %d", tc.method, tc.path, tc.want, rr.Code)
			}
		})
	}
}
