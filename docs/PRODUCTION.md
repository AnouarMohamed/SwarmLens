# Production Deployment Guide

This guide assumes a Linux VPS running Docker and Docker Compose v2.

## 1) Host baseline

- Use a dedicated non-root deploy user with Docker group access.
- Open only required ports (`22`, `80/443`, and application port if no reverse proxy).
- Keep Docker and OS patched (`apt update && apt upgrade`).
- Store the repo in a fixed path, e.g. `/opt/SwarmLens`.

## 2) Production environment file

Start from the template:

```bash
cp deploy/env/prod.env.example .env
chmod 600 .env
```

Important requirements:

- `APP_MODE=prod`
- `AUTH_ENABLED=true`
- `AUTH_TOKENS` or `AUTH_TOKENS_FILE` must be set
- `WRITE_ACTIONS_ENABLED=false` unless explicitly approved
- `LIVE_ACTION_POLICY=read_only_dry_run` for safe default behavior

## 3) Secrets via files (recommended)

SwarmLens backend and predictor support `_FILE` environment variables:

- `AUTH_TOKENS_FILE`
- `PREDICTOR_SHARED_SECRET_FILE`
- `ASSISTANT_API_KEY_FILE`

This keeps secrets out of shell history and process lists.

## 4) Preflight + deploy

```bash
./scripts/preflight-prod.sh
./scripts/deploy.sh
```

Useful options:

```bash
DEPLOY_REF=origin/main ./scripts/deploy.sh
HEALTHCHECK_TIMEOUT=240 ./scripts/deploy.sh
```

## 5) Rollback

Rollback to previous successful release recorded by deploy history:

```bash
./scripts/rollback.sh
```

Rollback to a specific commit or tag:

```bash
ROLLBACK_TO=<sha-or-tag> ./scripts/rollback.sh
```

## 6) Validation checklist

- `curl -fsS http://localhost:8080/api/v1/healthz` returns `{"status":"ok"}`
- `curl -fsS http://localhost:8080/api/v1/runtime` shows `mode=prod`
- `curl -fsS http://localhost:8080/api/v1/swarm` shows real managers/workers
- UI shows live/stale/disconnected freshness correctly
- audit entries are written for action requests

## 7) Recommended next hardening

- Put SwarmLens behind Nginx/Caddy with TLS and auth boundary.
- Restrict API ingress by network ACL/security group.
- Enable centralized logs and alerting.
- Add off-host backup/retention for deployment history and env/secret material.
