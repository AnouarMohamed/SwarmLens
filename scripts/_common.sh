#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="${APP_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"
ENV_FILE="${ENV_FILE:-$APP_DIR/.env}"
COMPOSE_FILE="${COMPOSE_FILE:-$APP_DIR/docker-compose.yml}"
PROJECT_NAME="${PROJECT_NAME:-swarmlens}"
HEALTHCHECK_URL="${HEALTHCHECK_URL:-http://localhost:8080/api/v1/healthz}"
RELEASE_DIR="${RELEASE_DIR:-$APP_DIR/.deploy-history}"
RELEASE_LOG="${RELEASE_LOG:-$RELEASE_DIR/releases.log}"

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || {
    echo "Missing required command: $1" >&2
    exit 1
  }
}

log() {
  printf '[%s] %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$*"
}

die() {
  log "ERROR: $*"
  exit 1
}

preflight() {
  require_cmd git
  require_cmd docker
  require_cmd curl
  [[ -f "$ENV_FILE" ]] || die "Missing env file: $ENV_FILE"
  [[ -f "$COMPOSE_FILE" ]] || die "Missing compose file: $COMPOSE_FILE"
  mkdir -p "$RELEASE_DIR"
  chmod 700 "$RELEASE_DIR"
  chmod 600 "$ENV_FILE" || true
}

compose() {
  docker compose \
    --project-name "$PROJECT_NAME" \
    --env-file "$ENV_FILE" \
    -f "$COMPOSE_FILE" \
    "$@"
}

render_compose() {
  compose config >/dev/null
}

deploy_stack() {
  compose up -d --build --remove-orphans
}

wait_for_health() {
  local timeout="${1:-120}"
  local start
  start="$(date +%s)"

  until curl --fail --silent "$HEALTHCHECK_URL" >/dev/null; do
    local now elapsed
    now="$(date +%s)"
    elapsed=$((now - start))
    if ((elapsed >= timeout)); then
      log "Health check timed out after ${timeout}s."
      compose ps || true
      compose logs --tail=120 || true
      return 1
    fi
    sleep 3
  done
  log "Health check passed: $HEALTHCHECK_URL"
}

current_sha() {
  git -C "$APP_DIR" rev-parse HEAD
}

resolve_ref() {
  local ref="$1"
  git -C "$APP_DIR" rev-parse "$ref"
}

record_release() {
  local sha="$1"
  printf '%s %s\n' "$(date -u '+%Y-%m-%dT%H:%M:%SZ')" "$sha" >>"$RELEASE_LOG"
}

last_recorded_sha() {
  [[ -f "$RELEASE_LOG" ]] || return 1
  awk 'NF >= 2 { print $2 }' "$RELEASE_LOG" | tail -n 1
}

previous_recorded_sha() {
  [[ -f "$RELEASE_LOG" ]] || return 1
  awk 'NF >= 2 { print $2 }' "$RELEASE_LOG" | tail -n 2 | head -n 1
}
