# Production Deployment Guide

This guide covers VPS deployment with Docker Compose and the included rollout scripts.

## 1) Host baseline

Recommended baseline:

- Linux host with Docker Engine + Compose v2
- Dedicated non-root deploy user in `docker` group
- Firewall open only for required ports (`22`, `80/443`, and app port if no reverse proxy)
- Repository checked out in a stable path such as `/opt/SwarmLens`

## 2) Prepare environment

Start from template:

```bash
cp deploy/env/prod.env.example .env
chmod 600 .env
```

Minimum production expectations:

- `APP_MODE=prod`
- `AUTH_ENABLED=true`
- `AUTH_TOKENS` or `AUTH_TOKENS_FILE` configured
- `WRITE_ACTIONS_ENABLED=false` by default
- `LIVE_ACTION_POLICY=read_only_dry_run`

## 3) Configure secrets

Prefer file-backed secrets. Supported by backend and predictor:

- `AUTH_TOKENS_FILE`
- `PREDICTOR_SHARED_SECRET_FILE`
- `ASSISTANT_API_KEY_FILE`

For Swarm stack overlay (`deploy/overlays/prod/stack.yml`), create required Swarm secrets first:

```bash
docker secret create swarmlens_auth_tokens - <<< "admin:admin:<strong-token>"
docker secret create swarmlens_predictor_secret - <<< "<predictor-secret>"
```

## 4) Preflight validation

Run the built-in checks before rollout:

```bash
./scripts/preflight-prod.sh
```

What it validates:

- Required env keys exist
- `APP_MODE=prod` implies `AUTH_ENABLED=true`
- Docker daemon is reachable
- Compose file renders successfully

## 5) Deploy

```bash
./scripts/deploy.sh
```

Useful options:

```bash
DEPLOY_REF=origin/main ./scripts/deploy.sh
HEALTHCHECK_TIMEOUT=240 ./scripts/deploy.sh
ENV_FILE=/opt/SwarmLens/.env ./scripts/deploy.sh
```

Deploy script behavior:

- fetches latest refs
- hard-resets to target ref
- renders compose config
- runs `docker compose up -d --build --remove-orphans`
- waits for health endpoint
- records release SHA in `.deploy-history/releases.log`

## 6) Rollback

Rollback to previous successful recorded release:

```bash
./scripts/rollback.sh
```

Rollback to explicit target:

```bash
ROLLBACK_TO=<sha-or-tag> ./scripts/rollback.sh
```

## 7) Post-deploy verification

```bash
curl -fsS http://localhost:8080/api/v1/healthz
curl -fsS http://localhost:8080/api/v1/runtime
curl -fsS http://localhost:8080/api/v1/swarm
```

Expected signals:

- health endpoint returns `{"status":"ok"}`
- runtime reports `mode=prod`
- swarm endpoint returns live managers/workers data
- UI reflects `live`, `stale`, or `disconnected` freshness correctly

## 8) Deployment models

### Compose on VPS

Use scripts in `scripts/` with `docker-compose.yml`.

### Docker Swarm stack

Use overlays in `deploy/overlays/`:

- `dev/stack.yml`
- `demo/stack.yml`
- `prod/stack.yml`

Example:

```bash
docker stack deploy -c deploy/overlays/prod/stack.yml swarmlens
```

## 9) Operational recommendations

- Put SwarmLens behind TLS reverse proxy and authentication boundary.
- Restrict API ingress to trusted networks.
- Centralize logs and metrics.
- Persist audit/history data externally if required for compliance.
- Add host-level backups for `.env`, secret metadata, and deploy history.
