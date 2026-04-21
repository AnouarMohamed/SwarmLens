package auth

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

type OIDCClaims struct {
	Subject           string
	Email             string
	PreferredUsername string
	Name              string
	ExplicitRole      string
	Groups            []string
}

type OIDCProvider struct {
	enabled        bool
	oauth2Config   oauth2.Config
	verifier       *oidc.IDTokenVerifier
	usernameClaim  string
	roleClaim      string
	groupsClaim    string
	viewerGroups   []string
	operatorGroups []string
	adminGroups    []string
}

func NewOIDC(ctx context.Context, cfg config.Config) (*OIDCProvider, error) {
	provider := &OIDCProvider{
		usernameClaim:  cfg.OIDCUsernameClaim,
		roleClaim:      cfg.OIDCRoleClaim,
		groupsClaim:    cfg.OIDCGroupsClaim,
		viewerGroups:   splitCSV(cfg.OIDCViewerGroups),
		operatorGroups: splitCSV(cfg.OIDCOperatorGroups),
		adminGroups:    splitCSV(cfg.OIDCAdminGroups),
	}
	if cfg.AuthProvider != "oidc" {
		return provider, nil
	}

	oidcProvider, err := oidc.NewProvider(ctx, cfg.OIDCIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("discover oidc provider: %w", err)
	}
	provider.enabled = true
	provider.verifier = oidcProvider.Verifier(&oidc.Config{ClientID: cfg.OIDCClientID})
	provider.oauth2Config = oauth2.Config{
		ClientID:     cfg.OIDCClientID,
		ClientSecret: cfg.OIDCClientSecret,
		Endpoint:     oidcProvider.Endpoint(),
		RedirectURL:  cfg.OIDCRedirectURL,
		Scopes:       splitCSV(cfg.OIDCScopes),
	}
	if len(provider.oauth2Config.Scopes) == 0 {
		provider.oauth2Config.Scopes = []string{oidc.ScopeOpenID, "profile", "email", "groups"}
	}
	return provider, nil
}

func (p *OIDCProvider) Enabled() bool {
	return p != nil && p.enabled
}

func (p *OIDCProvider) AuthCodeURL(state string) string {
	if !p.Enabled() {
		return ""
	}
	return p.oauth2Config.AuthCodeURL(state)
}

func (p *OIDCProvider) Exchange(ctx context.Context, code string) (OIDCClaims, error) {
	if !p.Enabled() {
		return OIDCClaims{}, ErrInvalidToken
	}
	token, err := p.oauth2Config.Exchange(ctx, code)
	if err != nil {
		return OIDCClaims{}, fmt.Errorf("exchange oidc code: %w", err)
	}
	rawIDToken, _ := token.Extra("id_token").(string)
	if rawIDToken == "" {
		return OIDCClaims{}, fmt.Errorf("oidc response missing id_token")
	}
	idToken, err := p.verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCClaims{}, fmt.Errorf("verify oidc token: %w", err)
	}
	var raw map[string]any
	if err := idToken.Claims(&raw); err != nil {
		return OIDCClaims{}, fmt.Errorf("decode oidc claims: %w", err)
	}
	return OIDCClaims{
		Subject:           stringClaim(raw, "sub"),
		Email:             stringClaim(raw, "email"),
		PreferredUsername: stringClaim(raw, "preferred_username"),
		Name:              stringClaim(raw, "name"),
		ExplicitRole:      stringClaim(raw, chooseString(p.roleClaim, "role")),
		Groups:            sliceClaim(raw, chooseString(p.groupsClaim, "groups")),
	}, nil
}

func (p *OIDCProvider) PrincipalFromClaims(claims OIDCClaims) model.Principal {
	username := claims.PreferredUsername
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = claims.Name
	}
	if username == "" {
		username = claims.Subject
	}

	role := model.RoleViewer
	switch strings.ToLower(strings.TrimSpace(claims.ExplicitRole)) {
	case "admin":
		role = model.RoleAdmin
	case "operator":
		role = model.RoleOperator
	case "viewer":
		role = model.RoleViewer
	default:
		switch {
		case overlaps(claims.Groups, p.adminGroups):
			role = model.RoleAdmin
		case overlaps(claims.Groups, p.operatorGroups):
			role = model.RoleOperator
		case overlaps(claims.Groups, p.viewerGroups):
			role = model.RoleViewer
		}
	}

	return model.Principal{
		Username: username,
		Role:     role,
		Provider: "oidc",
		Groups:   append([]string(nil), claims.Groups...),
	}
}

func stringClaim(raw map[string]any, key string) string {
	if key == "" {
		return ""
	}
	value, ok := raw[key]
	if !ok {
		return ""
	}
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}

func sliceClaim(raw map[string]any, key string) []string {
	if key == "" {
		return nil
	}
	value, ok := raw[key]
	if !ok {
		return nil
	}
	switch v := value.(type) {
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok && strings.TrimSpace(str) != "" {
				out = append(out, strings.TrimSpace(str))
			}
		}
		return out
	case []string:
		return append([]string(nil), v...)
	case string:
		if strings.TrimSpace(v) == "" {
			return nil
		}
		return splitCSV(v)
	default:
		return nil
	}
}

func splitCSV(raw string) []string {
	var values []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		values = append(values, item)
	}
	return values
}

func chooseString(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return fallback
}

func overlaps(actual, expected []string) bool {
	for _, candidate := range actual {
		if slices.Contains(expected, candidate) {
			return true
		}
	}
	return false
}
