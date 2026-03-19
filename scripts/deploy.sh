#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

DEPLOY_REF="${DEPLOY_REF:-origin/main}"
HEALTHCHECK_TIMEOUT="${HEALTHCHECK_TIMEOUT:-150}"

preflight

log "Fetching latest git refs..."
git -C "$APP_DIR" fetch --tags origin

from_sha="$(current_sha)"
target_sha="$(resolve_ref "$DEPLOY_REF")"

if [[ "$from_sha" != "$target_sha" ]]; then
  log "Updating working tree from $from_sha to $target_sha ($DEPLOY_REF)"
  git -C "$APP_DIR" reset --hard "$target_sha"
else
  log "Already on target commit $target_sha"
fi

log "Rendering compose config..."
render_compose

log "Deploying compose stack..."
deploy_stack

log "Waiting for backend health..."
wait_for_health "$HEALTHCHECK_TIMEOUT"

record_release "$target_sha"
log "Deployment successful at commit $target_sha"
