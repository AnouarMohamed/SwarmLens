#!/usr/bin/env bash
set -Eeuo pipefail
IFS=$'\n\t'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/_common.sh"

ROLLBACK_TO="${ROLLBACK_TO:-}"
HEALTHCHECK_TIMEOUT="${HEALTHCHECK_TIMEOUT:-150}"

preflight
git -C "$APP_DIR" fetch --tags origin

current="$(current_sha)"

if [[ -n "$ROLLBACK_TO" ]]; then
  target="$(resolve_ref "$ROLLBACK_TO")"
else
  target="$(previous_recorded_sha || true)"
fi

[[ -n "$target" ]] || die "No rollback target found. Set ROLLBACK_TO=<sha|tag|ref>."
if [[ "$target" == "$current" ]]; then
  die "Rollback target equals current commit ($current)."
fi

log "Rolling back from $current to $target"
git -C "$APP_DIR" reset --hard "$target"

log "Rendering compose config..."
render_compose

log "Re-deploying stack after rollback..."
deploy_stack

log "Waiting for backend health..."
wait_for_health "$HEALTHCHECK_TIMEOUT"

record_release "$target"
log "Rollback successful to $target"
