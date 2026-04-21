package auth

import (
	"testing"

	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

func TestPrincipalFromClaimsPrefersExplicitRoleOverGroups(t *testing.T) {
	provider := &OIDCProvider{
		viewerGroups:   []string{"viewers"},
		operatorGroups: []string{"operators"},
		adminGroups:    []string{"admins"},
	}

	principal := provider.PrincipalFromClaims(OIDCClaims{
		Subject:           "user-123",
		PreferredUsername: "alice",
		ExplicitRole:      "operator",
		Groups:            []string{"admins"},
	})

	if principal.Role != model.RoleOperator {
		t.Fatalf("expected explicit operator role to win, got %s", principal.Role)
	}
	if principal.Username != "alice" {
		t.Fatalf("expected preferred username, got %q", principal.Username)
	}
	if principal.Provider != "oidc" {
		t.Fatalf("expected provider oidc, got %q", principal.Provider)
	}
}

func TestPrincipalFromClaimsMapsGroupsByPriority(t *testing.T) {
	provider := &OIDCProvider{
		viewerGroups:   []string{"viewers"},
		operatorGroups: []string{"operators"},
		adminGroups:    []string{"admins"},
	}

	tests := []struct {
		name     string
		groups   []string
		wantRole model.Role
		wantUser string
		claims   OIDCClaims
	}{
		{
			name:     "admin outranks operator",
			groups:   []string{"operators", "admins"},
			wantRole: model.RoleAdmin,
			wantUser: "admin@example.com",
			claims: OIDCClaims{
				Subject: "user-1",
				Email:   "admin@example.com",
			},
		},
		{
			name:     "operator outranks viewer",
			groups:   []string{"viewers", "operators"},
			wantRole: model.RoleOperator,
			wantUser: "operator@example.com",
			claims: OIDCClaims{
				Subject: "user-2",
				Email:   "operator@example.com",
			},
		},
		{
			name:     "viewer fallback",
			groups:   []string{"viewers"},
			wantRole: model.RoleViewer,
			wantUser: "viewer@example.com",
			claims: OIDCClaims{
				Subject: "user-3",
				Email:   "viewer@example.com",
			},
		},
		{
			name:     "unknown groups stay viewer",
			groups:   []string{"random"},
			wantRole: model.RoleViewer,
			wantUser: "user-4",
			claims: OIDCClaims{
				Subject: "user-4",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.claims.Groups = tc.groups
			principal := provider.PrincipalFromClaims(tc.claims)
			if principal.Role != tc.wantRole {
				t.Fatalf("expected role %s, got %s", tc.wantRole, principal.Role)
			}
			if principal.Username != tc.wantUser {
				t.Fatalf("expected username %q, got %q", tc.wantUser, principal.Username)
			}
		})
	}
}

func TestPrincipalFromClaimsUsernameFallbackOrder(t *testing.T) {
	provider := &OIDCProvider{}

	tests := []struct {
		name   string
		claims OIDCClaims
		want   string
	}{
		{
			name: "preferred username wins",
			claims: OIDCClaims{
				Subject:           "subject-id",
				Email:             "alice@example.com",
				Name:              "Alice Operator",
				PreferredUsername: "alice",
			},
			want: "alice",
		},
		{
			name: "email fallback",
			claims: OIDCClaims{
				Subject: "subject-id",
				Email:   "alice@example.com",
				Name:    "Alice Operator",
			},
			want: "alice@example.com",
		},
		{
			name: "name fallback",
			claims: OIDCClaims{
				Subject: "subject-id",
				Name:    "Alice Operator",
			},
			want: "Alice Operator",
		},
		{
			name: "subject fallback",
			claims: OIDCClaims{
				Subject: "subject-id",
			},
			want: "subject-id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			principal := provider.PrincipalFromClaims(tc.claims)
			if principal.Username != tc.want {
				t.Fatalf("expected username %q, got %q", tc.want, principal.Username)
			}
		})
	}
}

func TestSliceClaimSupportsArrayAndCSVShapes(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]any
		want []string
	}{
		{
			name: "string array",
			raw: map[string]any{
				"groups": []string{"admins", "operators"},
			},
			want: []string{"admins", "operators"},
		},
		{
			name: "generic array trims blanks",
			raw: map[string]any{
				"groups": []any{"admins", " operators ", "", 7},
			},
			want: []string{"admins", "operators"},
		},
		{
			name: "csv string",
			raw: map[string]any{
				"groups": "admins, operators, viewers",
			},
			want: []string{"admins", "operators", "viewers"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := sliceClaim(tc.raw, "groups")
			if len(got) != len(tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
			for idx := range tc.want {
				if got[idx] != tc.want[idx] {
					t.Fatalf("expected %v, got %v", tc.want, got)
				}
			}
		})
	}
}
