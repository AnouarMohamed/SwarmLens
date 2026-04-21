package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

const (
	oidcStateCookie  = "swarmlens_oidc_state"
	oidcReturnCookie = "swarmlens_oidc_return"
)

func (d *deps) handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	if !d.oidcProvider.Enabled() {
		writeError(w, http.StatusNotImplemented, "auth_unavailable", "oidc login is not configured")
		return
	}
	state, err := randomToken(24)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "state_generation_failed", "failed to start oidc login")
		return
	}
	returnTo := strings.TrimSpace(r.URL.Query().Get("returnTo"))
	if returnTo == "" {
		returnTo = "/"
	}
	writeCookie(w, oidcStateCookie, state, 10*time.Minute, d.cfg.SessionCookieSecure, true)
	writeCookie(w, oidcReturnCookie, returnTo, 10*time.Minute, d.cfg.SessionCookieSecure, true)
	http.Redirect(w, r, d.oidcProvider.AuthCodeURL(state), http.StatusFound)
}

func (d *deps) handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !d.oidcProvider.Enabled() {
		writeError(w, http.StatusNotImplemented, "auth_unavailable", "oidc login is not configured")
		return
	}
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	expected := readCookie(r, oidcStateCookie)
	if state == "" || code == "" || expected == "" || state != expected {
		writeError(w, http.StatusBadRequest, "invalid_state", "oidc callback state mismatch")
		return
	}

	claims, err := d.oidcProvider.Exchange(r.Context(), code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "oidc_exchange_failed", err.Error())
		return
	}
	principal := d.oidcProvider.PrincipalFromClaims(claims)
	csrfToken, err := randomToken(24)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "csrf_generation_failed", "failed to create session")
		return
	}
	session := model.AuthSession{
		Username:   principal.Username,
		Role:       principal.Role,
		Provider:   principal.Provider,
		Groups:     append([]string(nil), principal.Groups...),
		CSRFToken:  csrfToken,
		CreatedAt:  time.Now().UTC(),
		LastSeenAt: time.Now().UTC(),
		ExpiresAt:  time.Now().UTC().Add(time.Duration(d.cfg.SessionTTLHours) * time.Hour),
	}
	saved, err := d.store.CreateSession(context.Background(), session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "session_create_failed", err.Error())
		return
	}

	writeCookie(w, d.cfg.SessionCookieName, saved.ID, time.Duration(d.cfg.SessionTTLHours)*time.Hour, d.cfg.SessionCookieSecure, true)
	clearCookie(w, oidcStateCookie, d.cfg.SessionCookieSecure)
	returnTo := readCookie(r, oidcReturnCookie)
	clearCookie(w, oidcReturnCookie, d.cfg.SessionCookieSecure)
	if returnTo == "" {
		returnTo = "/"
	}
	http.Redirect(w, r, returnTo, http.StatusFound)
}

func (d *deps) handleAuthLogout(w http.ResponseWriter, r *http.Request) {
	sessionID := readCookie(r, d.cfg.SessionCookieName)
	if sessionID != "" {
		_ = d.store.DeleteSession(r.Context(), sessionID)
	}
	clearCookie(w, d.cfg.SessionCookieName, d.cfg.SessionCookieSecure)
	writeJSON(w, http.StatusOK, map[string]any{"data": map[string]bool{"ok": true}})
}

func (d *deps) handleAuthMe(w http.ResponseWriter, r *http.Request) {
	if !d.cfg.AuthEnabled {
		writeOK(w, model.AuthIdentity{
			Authenticated: true,
			Username:      "anonymous",
			Role:          model.RoleAdmin,
			Provider:      "local",
		})
		return
	}

	if sessionID := readCookie(r, d.cfg.SessionCookieName); sessionID != "" {
		session, err := d.store.GetSession(r.Context(), sessionID)
		if err == nil {
			exp := session.ExpiresAt
			writeOK(w, model.AuthIdentity{
				Authenticated: true,
				Username:      session.Username,
				Role:          session.Role,
				Provider:      session.Provider,
				Groups:        append([]string(nil), session.Groups...),
				CSRFToken:     session.CSRFToken,
				ExpiresAt:     &exp,
			})
			return
		}
	}

	principal, err := d.authSvc.ExtractBearer(r)
	if err == nil {
		writeOK(w, model.AuthIdentity{
			Authenticated: true,
			Username:      principal.Username,
			Role:          principal.Role,
			Provider:      principal.Provider,
			Groups:        append([]string(nil), principal.Groups...),
		})
		return
	}

	writeOK(w, model.AuthIdentity{Authenticated: false})
}

func randomToken(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func writeCookie(w http.ResponseWriter, name, value string, ttl time.Duration, secure bool, httpOnly bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     "/",
		Expires:  time.Now().UTC().Add(ttl),
		MaxAge:   int(ttl.Seconds()),
		HttpOnly: httpOnly,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}

func clearCookie(w http.ResponseWriter, name string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	})
}
