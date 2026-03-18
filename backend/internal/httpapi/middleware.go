package httpapi

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
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
		p, err := d.authSvc.Extract(r)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized", err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), principalKey, p)
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

// ── CORS ──────────────────────────────────────────────────────────────────────

func middlewareCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Max-Age", "86400")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
	count    int
	resetAt  time.Time
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
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r)
			logger.Info("request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rw.status,
				"duration_ms", time.Since(start).Milliseconds(),
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
