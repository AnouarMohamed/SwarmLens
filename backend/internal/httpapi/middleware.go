package httpapi

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/auth"
	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type contextKey string

const principalKey contextKey = "principal"

// principalFrom extracts the authenticated principal from context.
func principalFrom(ctx context.Context) model.Principal {
	p, _ := ctx.Value(principalKey).(model.Principal)
	return p
}

// ── Auth middleware ───────────────────────────────────────────────────────────

// authMiddleware extracts and validates the principal, then calls the handler.
func (d *deps) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !d.cfg.AuthEnabled {
			ctx := context.WithValue(r.Context(), principalKey, model.Principal{
				Username: "anonymous",
				Role:     model.RoleAdmin,
				Provider: "local",
			})
			next(w, r.WithContext(ctx))
			return
		}

		ctx := r.Context()
		if sessionID := readCookie(r, d.cfg.SessionCookieName); sessionID != "" {
			session, err := d.store.GetSession(ctx, sessionID)
			if err == nil {
				_ = d.store.TouchSession(ctx, session.ID)
				ctx = context.WithValue(ctx, principalKey, model.Principal{
					Username: session.Username,
					Role:     session.Role,
					Provider: session.Provider,
					Groups:   append([]string(nil), session.Groups...),
					Session:  session.ID,
				})
				ctx = context.WithValue(ctx, sessionKey, session)
				ctx = context.WithValue(ctx, authMethodKey, "session")
				next(w, r.WithContext(ctx))
				return
			}
		}

		p, err := d.authSvc.ExtractBearer(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", err.Error())
			return
		}
		ctx = context.WithValue(ctx, principalKey, p)
		ctx = context.WithValue(ctx, authMethodKey, "token")
		next(w, r.WithContext(ctx))
	}
}

// writeMiddleware checks the write gate and operator role before calling the handler.
func (d *deps) writeMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := d.gate.Check(); err != nil {
			writeError(w, http.StatusForbidden, "writes_disabled", err.Error())
			return
		}
		p := principalFrom(r.Context())
		if err := auth.Require(p, model.RoleOperator); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		next(w, r)
	}
}

func (d *deps) adminMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := principalFrom(r.Context())
		if err := auth.Require(p, model.RoleAdmin); err != nil {
			writeError(w, http.StatusForbidden, "forbidden", err.Error())
			return
		}
		next(w, r)
	}
}

func (d *deps) csrfMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
			next(w, r)
			return
		}
		if authMethodFrom(r.Context()) != "session" {
			next(w, r)
			return
		}
		session, ok := sessionFrom(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "unauthorized", "session not found")
			return
		}
		token := strings.TrimSpace(r.Header.Get("X-CSRF-Token"))
		if token == "" || token != session.CSRFToken {
			writeError(w, http.StatusForbidden, "csrf_invalid", "missing or invalid csrf token")
			return
		}
		next(w, r)
	}
}

// ── CORS ──────────────────────────────────────────────────────────────────────

func middlewareCORS(cfg config.Config) func(http.Handler) http.Handler {
	allowedOrigins := parseAllowedOrigins(cfg.CORSAllowOrigins)
	allowAll := len(allowedOrigins) == 0

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && (allowAll || allowedOrigins[origin]) {
				headers := w.Header()
				headers.Set("Access-Control-Allow-Origin", origin)
				headers.Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				headers.Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token, X-Request-Id")
				headers.Set("Access-Control-Allow-Credentials", "true")
				headers.Set("Access-Control-Max-Age", "86400")
				headers.Set("Vary", appendVary(headers.Get("Vary"), "Origin"))
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func readCookie(r *http.Request, name string) string {
	if name == "" {
		return ""
	}
	cookie, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	value, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		return strings.TrimSpace(cookie.Value)
	}
	return strings.TrimSpace(value)
}

// ── Security headers ──────────────────────────────────────────────────────────

func middlewareSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// ── Rate limiter (simple token bucket per IP) ─────────────────────────────────

type rateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*bucket
	limit   int
	window  time.Duration
}

type bucket struct {
	count   int
	resetAt time.Time
}

func middlewareRateLimit(cfg config.Config) func(http.Handler) http.Handler {
	if !cfg.RateLimitEnabled {
		return func(next http.Handler) http.Handler { return next }
	}
	rl := &rateLimiter{
		buckets: make(map[string]*bucket),
		limit:   cfg.RateLimitRequests,
		window:  time.Duration(cfg.RateLimitWindowSecs) * time.Second,
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			rl.mu.Lock()
			b, ok := rl.buckets[ip]
			now := time.Now()
			if !ok || now.After(b.resetAt) {
				b = &bucket{resetAt: now.Add(rl.window)}
				rl.buckets[ip] = b
			}
			b.count++
			over := b.count > rl.limit
			rl.mu.Unlock()
			if over {
				writeError(w, http.StatusTooManyRequests, "rate_limited", "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ── Logging ───────────────────────────────────────────────────────────────────

func middlewareLogging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = fmt.Sprintf("req-%d", start.UnixNano())
			}
			w.Header().Set("X-Request-Id", requestID)

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.Info("request",
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// ── Panic recovery ────────────────────────────────────────────────────────────

func middlewareRecover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.Error("panic recovered", "error", rec, "stack", string(debug.Stack()))
					writeError(w, http.StatusInternalServerError, "internal_error", "internal server error")
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// ── Response writer wrapper ───────────────────────────────────────────────────

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func parseAllowedOrigins(raw string) map[string]bool {
	allowed := map[string]bool{}
	for _, token := range strings.Split(raw, ",") {
		origin := strings.TrimSpace(token)
		if origin == "" {
			continue
		}
		allowed[origin] = true
	}
	return allowed
}

func appendVary(existing string, value string) string {
	if existing == "" {
		return value
	}

	for _, token := range strings.Split(existing, ",") {
		if strings.EqualFold(strings.TrimSpace(token), value) {
			return existing
		}
	}
	return existing + ", " + value
}
