# Production Deployment Guide

SwarmLens v1.5 is designed to run as a small control plane stack: `backend`, `predictor`, and `postgres`, with persistent records stored in Postgres and schema migrations applied before the API comes up.

## Baseline

- Docker Swarm is the primary production target.
- Postgres is required in `prod`.
- OIDC session auth is the primary login path.
- Static tokens are break-glass only.
- Production rollouts should use versioned images and Swarm update rollback, not host-side `git reset --hard`.

## 1) Prepare secrets

Create the required Swarm secrets before deploy:

```bash
docker secret create swarmlens_postgres_password - <<< "<postgres-password>"
docker secret create swarmlens_database_url - <<< "postgres://swarmlens:<postgres-password>@postgres:5432/swarmlens?sslmode=disable"
docker secret create swarmlens_oidc_client_secret - <<< "<oidc-client-secret>"
docker secret create swarmlens_auth_tokens - <<< "breakglass-admin:admin:<strong-token>"
docker secret create swarmlens_predictor_secret - <<< "<predictor-secret>"
```

Optional if you enable the narrative assistant:

```bash
export ASSISTANT_API_KEY="<assistant-api-key>"
```

## 2) Prepare env

Start from the production template:

```bash
cp deploy/env/prod.env.example .env
chmod 600 .env
```

Set at least:

- `AUTH_OIDC_ISSUER_URL`
- `AUTH_OIDC_CLIENT_ID`
- `AUTH_OIDC_REDIRECT_URL`
- `AUTH_OIDC_VIEWER_GROUPS`
- `AUTH_OIDC_OPERATOR_GROUPS`
- `AUTH_OIDC_ADMIN_GROUPS`
- `CORS_ALLOW_ORIGINS`
- `ASSISTANT_PROVIDER` and `ASSISTANT_API_BASE_URL` if you want LLM-backed narrative output

## 3) Run migrations

The image supports a one-shot migration mode:

```bash
docker run --rm --env-file .env \
  -e RUN_MIGRATIONS_ONLY=true \
  ghcr.io/anouarmohamed/swarmlens:<tag>
```

The Compose and Swarm templates also include a `migrate` service using the same mode.

## 4) Deploy the stack

Use versioned images:

```bash
export SWARMLENS_IMAGE=ghcr.io/anouarmohamed/swarmlens:<tag>
export SWARMLENS_PREDICTOR_IMAGE=ghcr.io/anouarmohamed/swarmlens-predictor:<tag>
docker stack deploy -c deploy/overlays/prod/stack.yml swarmlens
```

Recommended rollout behavior already baked into the stack file:

- single replica backend on a manager
- `start-first` updates for the API
- automatic rollback on failed update
- persistent Postgres volume
- healthchecks on backend, predictor, and postgres

## 5) Verify the rollout

Check service state:

```bash
docker stack services swarmlens
docker service ps swarmlens_backend
docker service logs -f swarmlens_backend
```

Check control plane health:

```bash
curl -fsS http://localhost:8080/api/v1/healthz
curl -fsS http://localhost:8080/api/v1/readyz
curl -fsS http://localhost:8080/api/v1/runtime
```

Expected signals:

- `readyz` reports database readiness
- `runtime` shows persistent storage enabled
- the UI exposes cluster switching, approvals, and assistant sessions
- backend restarts do not lose incidents, audit entries, action runs, approvals, or assistant history

## 6) Roll forward and rollback

Roll forward by updating image tags and redeploying:

```bash
export SWARMLENS_IMAGE=ghcr.io/anouarmohamed/swarmlens:<new-tag>
export SWARMLENS_PREDICTOR_IMAGE=ghcr.io/anouarmohamed/swarmlens-predictor:<new-tag>
docker stack deploy -c deploy/overlays/prod/stack.yml swarmlens
```

If a rollout fails, Swarm uses the stack's rollback policy automatically. You can also force a previous image tag and redeploy.

## 7) Compose-based production-like testing

For a single-host validation environment, use `docker-compose.yml`:

```bash
cp .env.example .env
docker compose up --build
```

That stack now includes:

- `postgres`
- `migrate`
- `backend`
- `predictor`

## 8) Hardening checklist

- Put SwarmLens behind TLS.
- Set `SESSION_COOKIE_SECURE=true` in production.
- Restrict `CORS_ALLOW_ORIGINS` to exact operator origins.
- Keep `WRITE_APPROVAL_REQUIRED=true`.
- Keep `AUTH_TOKENS_FILE` populated with a break-glass admin token.
- Back up the Postgres volume and external secret material.
