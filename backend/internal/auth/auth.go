// Package auth handles principal extraction and role-based checks.
package auth

import (
	"errors"
	"net/http"
	"strings"

	"github.com/AnouarMohamed/swarmlens/backend/internal/config"
	"github.com/AnouarMohamed/swarmlens/backend/internal/model"
)

var (
	ErrNoToken      = errors.New("no authentication token provided")
	ErrInvalidToken = errors.New("invalid or unknown token")
	ErrForbidden    = errors.New("insufficient role for this action")
)

// Service validates tokens and extracts principals.
type Service struct {
	enabled      bool
	staticTokens map[string]config.StaticToken
}

// New creates an auth Service from config.
func New(cfg config.Config) *Service {
	return &Service{
		enabled:      cfg.AuthEnabled,
		staticTokens: cfg.ParseStaticTokens(),
	}
}

// Extract returns the Principal for a request.
// If auth is disabled, returns a synthetic admin principal (dev/demo convenience).
func (s *Service) Extract(r *http.Request) (model.Principal, error) {
	if !s.enabled {
		return model.Principal{Username: "anonymous", Role: model.RoleAdmin}, nil
	}

	token := extractToken(r)
	if token == "" {
		return model.Principal{}, ErrNoToken
	}

	st, ok := s.staticTokens[token]
	if !ok {
		return model.Principal{}, ErrInvalidToken
	}

	return model.Principal{Username: st.Username, Role: model.Role(st.Role)}, nil
}

// Require returns an error if the principal does not have at least the required role.
func Require(p model.Principal, required model.Role) error {
	if roleRank(p.Role) < roleRank(required) {
		return ErrForbidden
	}
	return nil
}

func roleRank(r model.Role) int {
	switch r {
	case model.RoleAdmin:
		return 2
	case model.RoleOperator:
		return 1
	default:
		return 0
	}
}

func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	return ""
}
