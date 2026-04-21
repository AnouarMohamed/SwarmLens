# Security

SwarmLens is an operator console for Docker Swarm. It can read cluster state and orchestrate actions through policy gates.

This document lists implemented controls and current limitations.

## Trust boundaries

- Browser talks to SwarmLens backend API.
- Backend talks to Docker Engine API and optional predictor service.
- Browser never receives raw Docker socket or predictor secret.
- Secret/config payload values are not exposed via API list endpoints.

## Implemented controls

| Control                      | Status                                 | Notes                                                                       |
| ---------------------------- | -------------------------------------- | --------------------------------------------------------------------------- |
| OIDC and session auth        | Implemented                            | OIDC login with backend-managed session cookies is available.               |
| Static token auth            | Implemented                            | Supported for dev and break-glass access.                                   |
| Static role model            | Implemented                            | Roles: `viewer`, `operator`, `admin`.                                       |
| Write gate                   | Implemented                            | `WRITE_ACTIONS_ENABLED=false` blocks mutating handlers.                     |
| Action policy modes          | Implemented                            | `read_only_dry_run`, `allowlist_live`, `demo_only`.                         |
| Persistent control-plane DB  | Implemented                            | Postgres-backed sessions, incidents, audit, actions, approvals, assistant.  |
| Approval workflow            | Implemented                            | Risky actions can enter durable admin approval before execution.            |
| Security headers             | Implemented                            | `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`.             |
| CORS middleware              | Implemented                            | Reflects request origin with explicit methods/headers allowlist.            |
| CSRF protection              | Implemented                            | Mutating cookie-authenticated requests require `X-CSRF-Token`.              |
| Rate limiting                | Implemented                            | Per-client token-bucket style limiter via `RATE_LIMIT_*`.                   |
| Non-root containers          | Implemented                            | Backend and predictor images run as UID/GID 1001.                           |
| Docker socket mount mode     | Implemented in compose/stack templates | Mounted read-only (`:ro`).                                                  |
| Predictor shared-secret auth | Implemented                            | Backend sends `X-Shared-Secret`; predictor rejects mismatch with 401.       |
| File-based secret ingestion  | Implemented                            | `_FILE` support for backend and predictor secrets.                          |

## Current limitations

- Inventory and workload contracts are not fully generated yet; the OpenAPI-driven type flow currently covers the v1.5 control-plane slice first.
- Event SSE endpoints use auth middleware; production ingress needs to preserve cookies or bearer tokens for these paths.
- CORS still depends on deployment configuration; production should set strict `CORS_ALLOW_ORIGINS`.
- Some high-risk actions remain intentionally dry-run only in this milestone, including arbitrary service update, stack deploy/remove, and node drain/activate.

## Production hardening checklist

- [ ] Set `APP_MODE=prod`
- [ ] Set `AUTH_ENABLED=true`
- [ ] Set `AUTH_PROVIDER=oidc`
- [ ] Set `SESSION_COOKIE_SECURE=true`
- [ ] Set `CORS_ALLOW_ORIGINS` to exact trusted origins
- [ ] Use strong secrets via files (`*_FILE`) instead of plain env vars
- [ ] Keep `WRITE_ACTIONS_ENABLED=false` unless a controlled change window requires writes
- [ ] Keep `LIVE_ACTION_POLICY=read_only_dry_run` unless explicitly approved
- [ ] Keep Docker socket mount read-only
- [ ] Run behind TLS reverse proxy (Nginx/Caddy/Traefik)
- [ ] Restrict ingress by firewall/security group
- [ ] Enable centralized logs and alerting
- [ ] Back up `.env` and secret material with strict access controls

## Secret handling guidance

Prefer file-backed secrets:

- `AUTH_TOKENS_FILE`
- `DATABASE_URL_FILE`
- `AUTH_OIDC_CLIENT_SECRET_FILE`
- `PREDICTOR_SHARED_SECRET_FILE`
- `ASSISTANT_API_KEY_FILE`

Use Swarm secrets or equivalent host secret stores. Avoid passing secrets inline in shell history.

## Vulnerability reporting

Report privately through GitHub Security Advisories or direct maintainer contact.
Do not open public issues with exploit details.
