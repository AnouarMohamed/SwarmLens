package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvOrFileReadsSecretFromFile(t *testing.T) {
	t.Setenv("AUTH_TOKENS", "")
	secretFile := filepath.Join(t.TempDir(), "auth_tokens")
	if err := os.WriteFile(secretFile, []byte("operator:operator:token-1\n"), 0o600); err != nil {
		t.Fatalf("failed writing secret file: %v", err)
	}
	t.Setenv("AUTH_TOKENS_FILE", secretFile)

	got := envOrFile("AUTH_TOKENS", "")
	if got != "operator:operator:token-1" {
		t.Fatalf("unexpected secret value: %q", got)
	}
}

func TestEnvOrFileEnvTakesPrecedence(t *testing.T) {
	t.Setenv("ASSISTANT_API_KEY", "from-env")
	t.Setenv("ASSISTANT_API_KEY_FILE", filepath.Join(t.TempDir(), "unused"))

	got := envOrFile("ASSISTANT_API_KEY", "")
	if got != "from-env" {
		t.Fatalf("expected env to win, got %q", got)
	}
}
