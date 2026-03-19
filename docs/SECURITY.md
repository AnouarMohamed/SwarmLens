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
| Auth toggle                  | Implemented                            | `AUTH_ENABLED=true` enforces bearer token auth.                             |
| Static role model            | Implemented                            | Roles: `viewer`, `operator`, `admin`.                                       |
| Write gate                   | Implemented                            | `WRITE_ACTIONS_ENABLED=false` blocks mutating handlers.                     |
| Action policy modes          | Implemented                            | `read_only_dry_run`, `allowlist_live`, `demo_only`.                         |
| Audit trail                  | Implemented                            | Action outcomes include `auditID`; audit records are append-only in memory. |
| Security headers             | Implemented                            | `X-Content-Type-Options`, `X-Frame-Options`, `Referrer-Policy`.             |
| CORS middleware              | Implemented                            | Reflects request origin with explicit methods/headers allowlist.            |
| Rate limiting                | Implemented                            | Per-client token-bucket style limiter via `RATE_LIMIT_*`.                   |
| Non-root containers          | Implemented                            | Backend and predictor images run as UID/GID 1001.                           |
| Docker socket mount mode     | Implemented in compose/stack templates | Mounted read-only (`:ro`).                                                  |
| Predictor shared-secret auth | Implemented                            | Backend sends `X-Shared-Secret`; predictor rejects mismatch with 401.       |
| File-based secret ingestion  | Implemented                            | `_FILE` support for backend and predictor secrets.                          |

## Current limitations

- Static tokens are currently the active auth mechanism; external OIDC fields are reserved and not fully wired.
- Audit store is in-memory; without external persistence it resets on restart.
- High-risk multi-party approval logic is policy-intent only today; no dedicated approval workflow endpoint is enforced yet.
- Incident create/update endpoints are authenticated but not role-gated beyond authenticated access.
- Event SSE endpoint uses auth middleware; ensure your deployment/auth strategy supports this path.
- CORS currently reflects incoming origin; production ingress should constrain reachable origins and networks.

## Production hardening checklist

- [ ] Set `APP_MODE=prod`
- [ ] Set `AUTH_ENABLED=true`
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
- `PREDICTOR_SHARED_SECRET_FILE`
- `ASSISTANT_API_KEY_FILE`

Use Swarm secrets or equivalent host secret stores. Avoid passing secrets inline in shell history.

## Vulnerability reporting

Report privately through GitHub Security Advisories or direct maintainer contact.
Do not open public issues with exploit details.
