package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	AppMode               string
	Port                  int
	DockerHost            string
	DockerTLSVerify       bool
	DockerCertPath        string
	AuthEnabled           bool
	AuthTokens            string
	AuthProvider          string
	OIDCIssuerURL         string
	OIDCClientID          string
	OIDCUsernameClaim     string
	OIDCRoleClaim         string
	TrustedProxyCIDRs     string
	CORSAllowOrigins      string
	WriteActionsEnabled   bool
	WriteApprovalRequired bool
	DiagnosticsSchedule   int
	PredictorBaseURL      string
	PredictorSecret       string
	LiveActionPolicy      string
	SnapshotStaleSeconds  int
	AssistantProvider     string
	AssistantBaseURL      string
	AssistantAPIKey       string
	AssistantModel        string
	AssistantRAGEnabled   bool
	RateLimitEnabled      bool
	RateLimitRequests     int
	RateLimitWindowSecs   int
	OTELEndpoint          string
	OTELServiceName       string
	SlackWebhookURL       string
	AlertmanagerURL       string
	PagerDutyKey          string
}

// Load reads config from environment, applies defaults, validates.
func Load() (Config, error) {
	cfg := Config{
		AppMode:               env("APP_MODE", "demo"),
		Port:                  envInt("PORT", 8080),
		DockerHost:            env("DOCKER_HOST", "unix:///var/run/docker.sock"),
		DockerTLSVerify:       envBool("DOCKER_TLS_VERIFY", false),
		DockerCertPath:        env("DOCKER_CERT_PATH", ""),
		AuthEnabled:           envBool("AUTH_ENABLED", false),
		AuthTokens:            envOrFile("AUTH_TOKENS", ""),
		AuthProvider:          env("AUTH_PROVIDER", ""),
		OIDCIssuerURL:         env("AUTH_OIDC_ISSUER_URL", ""),
		OIDCClientID:          env("AUTH_OIDC_CLIENT_ID", ""),
		OIDCUsernameClaim:     env("AUTH_OIDC_USERNAME_CLAIM", ""),
		OIDCRoleClaim:         env("AUTH_OIDC_ROLE_CLAIM", ""),
		TrustedProxyCIDRs:     env("AUTH_TRUSTED_PROXY_CIDRS", ""),
		CORSAllowOrigins:      env("CORS_ALLOW_ORIGINS", ""),
		WriteActionsEnabled:   envBool("WRITE_ACTIONS_ENABLED", false),
		WriteApprovalRequired: envBool("WRITE_APPROVAL_REQUIRED", true),
		DiagnosticsSchedule:   envInt("DIAGNOSTICS_SCHEDULE", 60),
		PredictorBaseURL:      env("PREDICTOR_BASE_URL", ""),
		PredictorSecret:       envOrFile("PREDICTOR_SHARED_SECRET", ""),
		LiveActionPolicy:      env("LIVE_ACTION_POLICY", "read_only_dry_run"),
		SnapshotStaleSeconds:  envInt("SNAPSHOT_STALE_SECONDS", 45),
		AssistantProvider:     env("ASSISTANT_PROVIDER", "none"),
		AssistantBaseURL:      env("ASSISTANT_API_BASE_URL", ""),
		AssistantAPIKey:       envOrFile("ASSISTANT_API_KEY", ""),
		AssistantModel:        env("ASSISTANT_MODEL", ""),
		AssistantRAGEnabled:   envBool("ASSISTANT_RAG_ENABLED", true),
		RateLimitEnabled:      envBool("RATE_LIMIT_ENABLED", true),
		RateLimitRequests:     envInt("RATE_LIMIT_REQUESTS", 300),
		RateLimitWindowSecs:   envInt("RATE_LIMIT_WINDOW_SECONDS", 60),
		OTELEndpoint:          env("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTELServiceName:       env("OTEL_SERVICE_NAME", "swarmlens-backend"),
		SlackWebhookURL:       env("SLACK_WEBHOOK_URL", ""),
		AlertmanagerURL:       env("ALERTMANAGER_WEBHOOK_URL", ""),
		PagerDutyKey:          env("PAGERDUTY_ROUTING_KEY", ""),
	}
	return cfg, cfg.validate()
}

func (c *Config) validate() error {
	valid := map[string]bool{"dev": true, "demo": true, "prod": true}
	if !valid[c.AppMode] {
		return errors.New("APP_MODE must be dev, demo, or prod")
	}
	if c.AppMode == "prod" && !c.AuthEnabled {
		return errors.New("prod mode requires AUTH_ENABLED=true")
	}
	if c.WriteActionsEnabled && !c.AuthEnabled && c.AppMode != "dev" {
		return errors.New("WRITE_ACTIONS_ENABLED=true requires AUTH_ENABLED=true outside dev mode")
	}
	validPolicies := map[string]bool{
		"read_only_dry_run": true,
		"allowlist_live":    true,
		"demo_only":         true,
	}
	if !validPolicies[c.LiveActionPolicy] {
		return errors.New("LIVE_ACTION_POLICY must be read_only_dry_run, allowlist_live, or demo_only")
	}
	if c.SnapshotStaleSeconds < 10 {
		return errors.New("SNAPSHOT_STALE_SECONDS must be >= 10")
	}
	return nil
}

func (c *Config) IsDemo() bool { return c.AppMode == "demo" }
func (c *Config) IsDev() bool  { return c.AppMode == "dev" }
func (c *Config) IsProd() bool { return c.AppMode == "prod" }

// ParseStaticTokens parses "user:role:token,..." into a map keyed by token.
func (c *Config) ParseStaticTokens() map[string]StaticToken {
	result := make(map[string]StaticToken)
	if c.AuthTokens == "" {
		return result
	}
	for _, entry := range strings.Split(c.AuthTokens, ",") {
		parts := strings.SplitN(strings.TrimSpace(entry), ":", 3)
		if len(parts) == 3 {
			result[parts[2]] = StaticToken{Username: parts[0], Role: parts[1]}
		}
	}
	return result
}

type StaticToken struct {
	Username string
	Role     string
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func envInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

func envOrFile(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	fileKey := key + "_FILE"
	path := strings.TrimSpace(os.Getenv(fileKey))
	if path == "" {
		return fallback
	}
	content, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return fallback
	}
	value := strings.TrimSpace(string(content))
	if value == "" {
		return fallback
	}
	return value
}
