#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

preflight

require_var() {
  local key="$1"
  if ! grep -Eq "^[[:space:]]*${key}=" "$ENV_FILE"; then
    die "Missing ${key} in ${ENV_FILE}"
  fi
}

log "Running production preflight checks..."
require_var APP_MODE
require_var AUTH_ENABLED
require_var DOCKER_HOST
require_var WRITE_ACTIONS_ENABLED
require_var LIVE_ACTION_POLICY

app_mode="$(grep -E '^[[:space:]]*APP_MODE=' "$ENV_FILE" | tail -n 1 | cut -d= -f2- | tr -d '"' | xargs)"
auth_enabled="$(grep -E '^[[:space:]]*AUTH_ENABLED=' "$ENV_FILE" | tail -n 1 | cut -d= -f2- | tr -d '"' | xargs)"

if [[ "$app_mode" == "prod" && "$auth_enabled" != "true" ]]; then
  die "APP_MODE=prod requires AUTH_ENABLED=true"
fi

render_compose
log "Compose config validated."

if ! docker info >/dev/null 2>&1; then
  die "Docker daemon is not reachable"
fi

log "Preflight complete."
