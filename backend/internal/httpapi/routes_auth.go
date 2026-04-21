package httpapi

import "net/http"

func (d *deps) registerAuthRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/auth/login", d.handleAuthLogin)
	mux.HandleFunc("GET /api/v1/auth/callback", d.handleAuthCallback)
	mux.HandleFunc("POST /api/v1/auth/logout", d.authMiddleware(d.csrfMiddleware(d.handleAuthLogout)))
	mux.HandleFunc("GET /api/v1/auth/me", d.handleAuthMe)
}
